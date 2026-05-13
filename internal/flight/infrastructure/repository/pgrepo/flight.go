package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

type FlightRepo struct {
	db *sqlx.DB
}

func NewFlightRepo(db *sqlx.DB) *FlightRepo {
	return &FlightRepo{db: db}
}

func (r *FlightRepo) Create(ctx context.Context, flight *domain.Flight, fares []domain.FlightFare) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	queryFlight := `
		INSERT INTO flights (flight_number, airline_id, aircraft_id, departure_airport, arrival_airport,
		                     departure_time, arrival_time, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var flightID int64
	err = tx.QueryRowContext(ctx, queryFlight,
		flight.FlightNumber, flight.AirlineID, flight.AircraftID,
		flight.DepartureAirportCode, flight.ArrivalAirportCode,
		flight.DepartureTime, flight.ArrivalTime, flight.Status,
	).Scan(&flightID)

	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrFlightAlreadyExists
		}
		return 0, fmt.Errorf("insert flight: %w", err)
	}

	if len(fares) > 0 {
		var seatCounts []struct {
			SeatClass string `db:"seat_class"`
			Count     int    `db:"count"`
		}
		if err := tx.SelectContext(ctx, &seatCounts,
			`SELECT seat_class, COUNT(*) as count FROM aircraft_seats WHERE aircraft_id = $1 GROUP BY seat_class`,
			flight.AircraftID,
		); err != nil {
			return 0, fmt.Errorf("query seat counts: %w", err)
		}
		countByClass := make(map[string]int, len(seatCounts))
		for _, sc := range seatCounts {
			countByClass[sc.SeatClass] = sc.Count
		}

		for i := range fares {
			fares[i].FlightID = flightID
			fares[i].SeatsTotal = countByClass[string(fares[i].SeatClass)]
		}
		queryFares := `
			INSERT INTO flight_fares (flight_id, seat_class, price_cents, seats_total, seats_booked)
			VALUES (:flight_id, :seat_class, :price_cents, :seats_total, 0)
		`
		if _, err := tx.NamedExecContext(ctx, queryFares, fares); err != nil {
			return 0, fmt.Errorf("insert fares: %w", err)
		}
	}

	eventPayload := domain.FlightCreatedPayload{
		FlightID: flightID,
	}
	if err := insertOutboxEvent(ctx, tx, domain.EventTypeFlightCreated, eventPayload); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return flightID, nil
}

func (r *FlightRepo) UpdateStatus(ctx context.Context, flightID int64, status domain.FlightStatus) error {
	query := `
		UPDATE flights
		SET status = $1
		WHERE id = $2
	`
	res, err := r.db.ExecContext(ctx, query, status, flightID)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrFlightNotFound
	}
	return nil
}

type flightDetailsRow struct {
	FlightID           int64               `db:"flight_id"`
	FlightNumber       string              `db:"flight_number"`
	DepartureTime      time.Time           `db:"departure_time"`
	ArrivalTime        time.Time           `db:"arrival_time"`
	Status             domain.FlightStatus `db:"status"`
	MinPriceCents      int64               `db:"min_price_cents"`
	AvailableSeats     int                 `db:"available_seats"`
	AirlineID          int64               `db:"airline_id"`
	AirlineIataCode    string              `db:"airline_iata_code"`
	AirlineName        string              `db:"airline_name"`
	AirlineCountry     string              `db:"airline_country"`
	AirlineLogoURL     string              `db:"airline_logo_url"`
	AircraftID         int64               `db:"aircraft_id"`
	AircraftModel      string              `db:"aircraft_model"`
	AircraftTotalSeats int                 `db:"aircraft_total_seats"`
	DepCode            string              `db:"dep_code"`
	DepName            string              `db:"dep_name"`
	DepCity            string              `db:"dep_city"`
	DepCountry         string              `db:"dep_country"`
	DepTimezone        string              `db:"dep_timezone"`
	ArrCode            string              `db:"arr_code"`
	ArrName            string              `db:"arr_name"`
	ArrCity            string              `db:"arr_city"`
	ArrCountry         string              `db:"arr_country"`
	ArrTimezone        string              `db:"arr_timezone"`
}

func (r *FlightRepo) GetByID(ctx context.Context, flightID int64) (*domain.FlightDetails, error) {
	query := `
		SELECT
		    f.id               AS flight_id,
		    f.flight_number,
		    f.departure_time,
		    f.arrival_time,
		    f.status,
		    COALESCE(MIN(ff.price_cents), 0)                   AS min_price_cents,
		    COALESCE(SUM(ff.seats_total - ff.seats_booked), 0) AS available_seats,
		    al.id              AS airline_id,
		    al.iata_code       AS airline_iata_code,
		    al.name            AS airline_name,
		    al.country         AS airline_country,
		    al.logo_url        AS airline_logo_url,
		    ac.id              AS aircraft_id,
		    ac.model           AS aircraft_model,
		    ac.total_seats     AS aircraft_total_seats,
		    dep.code           AS dep_code,
		    dep.name           AS dep_name,
		    dep.city           AS dep_city,
		    dep.country        AS dep_country,
		    dep.timezone       AS dep_timezone,
		    arr.code           AS arr_code,
		    arr.name           AS arr_name,
		    arr.city           AS arr_city,
		    arr.country        AS arr_country,
		    arr.timezone       AS arr_timezone
		FROM flights f
		JOIN airlines al   ON al.id    = f.airline_id
		JOIN aircrafts ac  ON ac.id    = f.aircraft_id
		JOIN airports dep  ON dep.code = f.departure_airport
		JOIN airports arr  ON arr.code = f.arrival_airport
		LEFT JOIN flight_fares ff ON ff.flight_id = f.id
		WHERE f.id = $1
		GROUP BY f.id, al.id, ac.id, dep.code, arr.code
	`

	var row flightDetailsRow
	if err := r.db.GetContext(ctx, &row, query, flightID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrFlightNotFound
		}
		return nil, fmt.Errorf("get flight by id: %w", err)
	}

	return &domain.FlightDetails{
		ID:             row.FlightID,
		FlightNumber:   row.FlightNumber,
		DepartureTime:  row.DepartureTime,
		ArrivalTime:    row.ArrivalTime,
		Status:         row.Status,
		MinPriceCents:  row.MinPriceCents,
		AvailableSeats: row.AvailableSeats,
		Airline: domain.Airline{
			ID:       row.AirlineID,
			IATACode: row.AirlineIataCode,
			Name:     row.AirlineName,
			Country:  row.AirlineCountry,
			LogoURL:  row.AirlineLogoURL,
		},
		Aircraft: domain.Aircraft{
			ID:         row.AircraftID,
			Model:      row.AircraftModel,
			TotalSeats: row.AircraftTotalSeats,
		},
		Departure: domain.Airport{
			Code:     row.DepCode,
			Name:     row.DepName,
			City:     row.DepCity,
			Country:  row.DepCountry,
			Timezone: row.DepTimezone,
		},
		Arrival: domain.Airport{
			Code:     row.ArrCode,
			Name:     row.ArrName,
			City:     row.ArrCity,
			Country:  row.ArrCountry,
			Timezone: row.ArrTimezone,
		},
	}, nil
}

func (r *FlightRepo) GetFaresByFlightID(ctx context.Context, flightID int64) ([]domain.FlightFare, error) {
	query := `SELECT * FROM flight_fares WHERE flight_id = $1`
	var fares []domain.FlightFare
	if err := r.db.SelectContext(ctx, &fares, query, flightID); err != nil {
		return nil, fmt.Errorf("get fares: %w", err)
	}
	return fares, nil
}

type flightListRow struct {
	FlightID        int64               `db:"flight_id"`
	FlightNumber    string              `db:"flight_number"`
	DepartureTime   time.Time           `db:"departure_time"`
	ArrivalTime     time.Time           `db:"arrival_time"`
	Status          domain.FlightStatus `db:"status"`
	MinPriceCents   int64               `db:"min_price_cents"`
	AvailableSeats  int                 `db:"available_seats"`
	AirlineID       int64               `db:"airline_id"`
	AirlineIataCode string              `db:"airline_iata_code"`
	AirlineName     string              `db:"airline_name"`
	AirlineCountry  string              `db:"airline_country"`
	AirlineLogoURL  string              `db:"airline_logo_url"`
	DepCode         string              `db:"dep_code"`
	DepName         string              `db:"dep_name"`
	DepCity         string              `db:"dep_city"`
	DepCountry      string              `db:"dep_country"`
	DepTimezone     string              `db:"dep_timezone"`
	ArrCode         string              `db:"arr_code"`
	ArrName         string              `db:"arr_name"`
	ArrCity         string              `db:"arr_city"`
	ArrCountry      string              `db:"arr_country"`
	ArrTimezone     string              `db:"arr_timezone"`
}

func (r *FlightRepo) List(ctx context.Context, limit, offset int) ([]domain.FlightDocument, error) {
	query := `
		SELECT
		    f.id               AS flight_id,
		    f.flight_number,
		    f.departure_time,
		    f.arrival_time,
		    f.status,
		    COALESCE(MIN(ff.price_cents), 0)                       AS min_price_cents,
		    COALESCE(SUM(ff.seats_total - ff.seats_booked), 0)     AS available_seats,
		    al.id              AS airline_id,
		    al.iata_code       AS airline_iata_code,
		    al.name            AS airline_name,
		    al.country         AS airline_country,
		    al.logo_url        AS airline_logo_url,
		    dep.code           AS dep_code,
		    dep.name           AS dep_name,
		    dep.city           AS dep_city,
		    dep.country        AS dep_country,
		    dep.timezone       AS dep_timezone,
		    arr.code           AS arr_code,
		    arr.name           AS arr_name,
		    arr.city           AS arr_city,
		    arr.country        AS arr_country,
		    arr.timezone       AS arr_timezone
		FROM flights f
		JOIN airlines al  ON al.id    = f.airline_id
		JOIN airports dep ON dep.code = f.departure_airport
		JOIN airports arr ON arr.code = f.arrival_airport
		LEFT JOIN flight_fares ff ON ff.flight_id = f.id
		GROUP BY f.id, al.id, dep.code, arr.code
		ORDER BY f.departure_time DESC
		LIMIT $1 OFFSET $2
	`

	var rows []flightListRow
	if err := r.db.SelectContext(ctx, &rows, query, limit, offset); err != nil {
		return nil, fmt.Errorf("list flights: %w", err)
	}

	docs := make([]domain.FlightDocument, len(rows))
	for i, row := range rows {
		docs[i] = domain.FlightDocument{
			ID:             row.FlightID,
			FlightNumber:   row.FlightNumber,
			DepartureTime:  row.DepartureTime,
			ArrivalTime:    row.ArrivalTime,
			Status:         row.Status,
			MinPriceCents:  row.MinPriceCents,
			AvailableSeats: row.AvailableSeats,
			Airline: domain.Airline{
				ID:       row.AirlineID,
				IATACode: row.AirlineIataCode,
				Name:     row.AirlineName,
				Country:  row.AirlineCountry,
				LogoURL:  row.AirlineLogoURL,
			},
			Departure: domain.Airport{
				Code:     row.DepCode,
				Name:     row.DepName,
				City:     row.DepCity,
				Country:  row.DepCountry,
				Timezone: row.DepTimezone,
			},
			Arrival: domain.Airport{
				Code:     row.ArrCode,
				Name:     row.ArrName,
				City:     row.ArrCity,
				Country:  row.ArrCountry,
				Timezone: row.ArrTimezone,
			},
		}
	}
	return docs, nil
}
