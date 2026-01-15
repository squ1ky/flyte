package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

type FlightRepo struct {
	db *sqlx.DB
}

func NewFlightRepo(db *sqlx.DB) *FlightRepo {
	return &FlightRepo{db: db}
}

func (r *FlightRepo) CreateFlight(ctx context.Context, f *domain.Flight) (int64, error) {
	query := `
		INSERT INTO flights (flight_number, departure_airport, arrival_airport,
		                     departure_time, arrival_time, price_cents, total_seats, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query, f.FlightNumber, f.DepartureAirport, f.ArrivalAirport,
		f.DepartureTime, f.ArrivalTime, f.PriceCents, f.TotalSeats, f.Status).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, err
}

func (r *FlightRepo) GetByID(ctx context.Context, id int64) (*domain.Flight, error) {
	query := `
		SELECT f.*, COUNT(s.id) FILTER (WHERE s.is_booked = FALSE) as available_seats
		FROM flights f
		LEFT JOIN seats s ON f.id = s.flight_id
		WHERE f.id = $1
		GROUP BY f.id
	`

	var flight domain.Flight
	if err := r.db.GetContext(ctx, &flight, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrFlightNotFound
		}
		return nil, err
	}

	return &flight, nil
}

func (r *FlightRepo) GetSeatsByFlightID(ctx context.Context, flightID int64) ([]domain.Seat, error) {
	var exists bool
	err := r.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM flights WHERE id = $1)", flightID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrFlightNotFound
	}

	query := `SELECT * FROM seats WHERE flight_id = $1 ORDER BY seat_number`
	var seats []domain.Seat
	if err := r.db.SelectContext(ctx, &seats, query, flightID); err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *FlightRepo) GetAirports(ctx context.Context) ([]domain.Airport, error) {
	query := `SELECT * FROM airports ORDER BY city`
	var airports []domain.Airport
	if err := r.db.SelectContext(ctx, &airports, query); err != nil {
		return nil, err
	}
	return airports, nil
}

func (r *FlightRepo) DeleteFlight(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM flights WHERE id = $1", id)
	return err
}

func (r *FlightRepo) BookSeat(ctx context.Context, flightID int64, seatNumber string) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	query := `
		SELECT id, is_booked
		FROM seats
		WHERE flight_id = $1 AND seat_number = $2
		FOR UPDATE
	`

	var seat struct {
		ID       int64 `db:"id"`
		IsBooked bool  `db:"is_booked"`
	}

	if err := tx.GetContext(ctx, &seat, query, flightID, seatNumber); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrSeatNotFound
		}
		return 0, err
	}

	if seat.IsBooked {
		return 0, domain.ErrSeatAlreadyBooked
	}

	_, err = tx.ExecContext(ctx, "UPDATE seats SET is_booked = TRUE WHERE id = $1", seat.ID)
	if err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return seat.ID, nil
}

func (r *FlightRepo) ReleaseSeat(ctx context.Context, flightID int64, seatNumber string) error {
	query := `UPDATE seats SET is_booked = FALSE WHERE flight_id = $1 AND seat_number = $2`
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
