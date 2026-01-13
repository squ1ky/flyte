package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/booking/domain"
)

type BookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) Create(ctx context.Context, b *domain.Booking) (string, error) {
	query := `
		INSERT INTO bookings (
			user_id, flight_id, seat_number,
		    passenger_name, passenger_passport,
		    price, currency, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(ctx, query,
		b.UserID, b.FlightID, b.SeatNumber,
		b.PassengerName, b.PassengerPassport,
		b.Price, b.Currency, b.Status,
	).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to create booking: %w", err)
	}

	return id, nil
}

func (r *BookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	var booking domain.Booking
	query := `SELECT * FROM bookings WHERE id = $1`

	if err := r.db.GetContext(ctx, &booking, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("booking not found")
		}
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	return &booking, nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	query := `
		UPDATE bookings
		SET status = $1,
			updated_at = NOW()
		WHERE id = $2
		AND status NOT IN ('PAID', 'CANCELLED', 'FAILED')
	`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to execute update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("booking with id %s not found", id)
	}

	return nil
}

func (r *BookingRepo) ListByUserID(ctx context.Context, userID int64) ([]domain.Booking, error) {
	var bookings []domain.Booking
	query := `SELECT * FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`

	err := r.db.SelectContext(ctx, &bookings, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings: %w", err)
	}

	return bookings, nil
}
