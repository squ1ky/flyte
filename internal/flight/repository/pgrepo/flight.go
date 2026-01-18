package pgrepo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

const (
	pgErrUniqueViolation = "23505"
)

type FlightRepo struct {
	db *sqlx.DB
}

func NewFlightRepo(db *sqlx.DB) *FlightRepo {
	return &FlightRepo{db: db}
}

func (r *FlightRepo) CreateFlight(ctx context.Context, f *domain.Flight) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	queryFlight := `
		INSERT INTO flights (flight_number, aircraft_id, departure_airport, arrival_airport,
		                     departure_time, arrival_time, base_price_cents, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var flightID int64
	err = tx.QueryRowContext(ctx, queryFlight, f.FlightNumber, f.AircraftID, f.DepartureAirport, f.ArrivalAirport,
		f.DepartureTime, f.ArrivalTime, f.BasePriceCents, f.Status).Scan(&flightID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation {
			return 0, domain.ErrFlightAlreadyExists
		}
		return 0, fmt.Errorf("insert flight: %w", err)
	}
	f.ID = flightID

	queryCopySeats := `
		INSERT INTO seats (flight_id, seat_number, seat_class, price_multiplier, is_booked)
		SELECT $1, seat_number, seat_class, price_multiplier, FALSE
		FROM aircraft_seats
		WHERE aircraft_id = $2
	`
	res, err := tx.ExecContext(ctx, queryCopySeats, flightID, f.AircraftID)
	if err != nil {
		return 0, fmt.Errorf("copy seats: %w", err)
	}

	rowsCopied, _ := res.RowsAffected()
	if rowsCopied == 0 {
		return 0, fmt.Errorf("no seats template found for aircraft_id %d", f.AircraftID)
	}
	f.AvailableSeats = int(rowsCopied)

	if err := r.insertOutboxEvent(ctx, tx, domain.EventFlightCreated, f); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return flightID, nil
}

func (r *FlightRepo) GetByID(ctx context.Context, id int64) (*domain.Flight, error) {
	query := `
		SELECT f.*,
		       (SELECT COUNT(*)
		        FROM seats s
		        WHERE s.flight_id = f.id AND s.is_booked = FALSE) as available_seats
		FROM flights f
		WHERE f.id = $1
	`

	var flight domain.Flight
	if err := r.db.GetContext(ctx, &flight, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrFlightNotFound
		}
		return nil, fmt.Errorf("get flight by id: %w", err)
	}

	return &flight, nil
}

func (r *FlightRepo) DeleteFlight(ctx context.Context, id int64) error {
	query := `DELETE FROM flights WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete flight: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrFlightNotFound
	}
	return nil
}

func (r *FlightRepo) GetSeatsByFlightID(ctx context.Context, flightID int64) ([]domain.Seat, error) {
	queryCheck := "SELECT EXISTS(SELECT 1 FROM flights WHERE id = $1)"
	var exists bool
	if err := r.db.GetContext(ctx, &exists, queryCheck, flightID); err != nil {
		return nil, fmt.Errorf("check flight exists: %w", err)
	}
	if !exists {
		return nil, domain.ErrFlightNotFound
	}

	querySelect := "SELECT * FROM seats WHERE flight_id = $1 ORDER BY seat_number"
	var seats []domain.Seat
	if err := r.db.SelectContext(ctx, &seats, querySelect, flightID); err != nil {
		return nil, fmt.Errorf("select seats: %w", err)
	}
	return seats, nil
}

func (r *FlightRepo) BookSeat(ctx context.Context, flightID int64, seatNumber string) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queryLock := `
		SELECT id, is_booked
		FROM seats
		WHERE flight_id = $1 AND seat_number = $2
		FOR UPDATE
	`

	var seat struct {
		ID       int64 `db:"id"`
		IsBooked bool  `db:"is_booked"`
	}

	if err := tx.GetContext(ctx, &seat, queryLock, flightID, seatNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrSeatNotFound
		}
		return 0, fmt.Errorf("fetch seat: %w", err)
	}
	if seat.IsBooked {
		return 0, domain.ErrSeatAlreadyBooked
	}

	queryUpdate := `
		UPDATE seats
		SET is_booked = TRUE, reserved_at = NOW()
		WHERE id = $1
	`
	if _, err := tx.ExecContext(ctx, queryUpdate, seat.ID); err != nil {
		return 0, fmt.Errorf("update seat: %w", err)
	}

	outboxPayload := map[string]int64{
		"flight_id": flightID,
	}
	if err := r.insertOutboxEvent(ctx, tx, domain.EventSeatsChanged, outboxPayload); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return seat.ID, nil
}

func (r *FlightRepo) ReleaseSeat(ctx context.Context, flightID int64, seatNumber string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	queryUpdate := `
		UPDATE seats
		SET is_booked = FALSE, reserved_at = NULL
		WHERE flight_id = $1 AND seat_number = $2
	`
	res, err := r.db.ExecContext(ctx, queryUpdate, flightID, seatNumber)
	if err != nil {
		return fmt.Errorf("update seat: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrSeatNotFound
	}

	payload := map[string]int64{
		"flight_id": flightID,
	}
	if err := r.insertOutboxEvent(ctx, tx, domain.EventSeatsChanged, payload); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *FlightRepo) ConfirmSeat(ctx context.Context, flightID int64, seatNumber string) error {
	query := `
		UPDATE seats
		SET reserved_at = NULL
		WHERE flight_id = $1 AND seat_number = $2 AND is_booked = TRUE
	`

	res, err := r.db.ExecContext(ctx, query, flightID, seatNumber)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return domain.ErrSeatNotFound
	}

	return nil
}

func (r *FlightRepo) GetAirports(ctx context.Context) ([]domain.Airport, error) {
	query := `SELECT * FROM airports ORDER BY city`
	var airports []domain.Airport
	if err := r.db.SelectContext(ctx, &airports, query); err != nil {
		return nil, err
	}
	return airports, nil
}

func (r *FlightRepo) insertOutboxEvent(ctx context.Context, tx *sqlx.Tx, eventType domain.EventType, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal outbox payload: %w", err)
	}

	query := `
		INSERT INTO flight_outbox (event_type, payload)
		VALUES ($1, $2)
	`

	if _, err := tx.ExecContext(ctx, query, eventType, data); err != nil {
		return fmt.Errorf("insert outbox event: %w", err)
	}
	return nil
}
