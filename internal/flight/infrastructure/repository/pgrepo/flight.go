package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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
		for i := range fares {
			fares[i].FlightID = flightID
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

func (r *FlightRepo) GetByID(ctx context.Context, flightID int64) (*domain.Flight, error) {
	query := `SELECT * FROM flights WHERE id = $1`

	var f domain.Flight
	if err := r.db.GetContext(ctx, &f, query, flightID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrFlightNotFound
		}
		return nil, fmt.Errorf("get flight by id: %w", err)
	}

	return &f, nil
}

func (r *FlightRepo) GetFaresByFlightID(ctx context.Context, flightID int64) ([]domain.FlightFare, error) {
	query := `SELECT * FROM flight_fares WHERE flight_id = $1`
	var fares []domain.FlightFare
	if err := r.db.SelectContext(ctx, &fares, query, flightID); err != nil {
		return nil, fmt.Errorf("get fares: %w", err)
	}
	return fares, nil
}
