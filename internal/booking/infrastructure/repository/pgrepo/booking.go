package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"time"
)

type BookingRepo struct {
	db *sqlx.DB
}

func NewBookingRepo(db *sqlx.DB) *BookingRepo {
	return &BookingRepo{db: db}
}

func (r *BookingRepo) Create(ctx context.Context, booking *domain.Booking, tickets []domain.Ticket) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const bookingQuery = `
		INSERT INTO bookings (id, user_id, flight_id, ref_code, status,
		                      total_price_cents, currency, contact_email, expires_at, created_at)
		VALUES (:id, :user_id, :flight_id, :ref_code, :status,
		        :total_price_cents, :currency, :contact_email, :expires_at, :created_at)
	`

	if _, err := tx.NamedExecContext(ctx, bookingQuery, booking); err != nil {
		return fmt.Errorf("insert booking: %w", err)
	}

	const ticketQuery = `
		INSERT INTO tickets (id, booking_id, passenger_first_name, passenger_last_name,
		                     passenger_doc_num, seat_number, price_cents, status)
		VALUES (:id, :booking_id, :passenger_first_name, :passenger_last_name,
		        :passenger_doc_num, :seat_number, :price_cents, :status)
	`

	for i := range tickets {
		if _, err := tx.NamedExecContext(ctx, ticketQuery, &tickets[i]); err != nil {
			return fmt.Errorf("insert ticket %s: %w", tickets[i].SeatNumber, err)
		}
	}

	return tx.Commit()
}

func (r *BookingRepo) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	var booking domain.Booking

	const query = `
		SELECT id, user_id, flight_id, ref_code, status,
			   total_price_cents, currency, contact_email, expires_at, created_at
		FROM bookings
		WHERE id = $1
	`

	if err := r.db.GetContext(ctx, &booking, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking by id: %w", err)
	}
	return &booking, nil
}

func (r *BookingRepo) GetByRefCode(ctx context.Context, refCode string) (*domain.Booking, error) {
	var booking domain.Booking

	const query = `
		SELECT id, user_id, flight_id, ref_code, status,
			   total_price_cents, currency, contact_email, expires_at, created_at
		FROM bookings
		WHERE ref_code = $1
	`

	if err := r.db.GetContext(ctx, &booking, query, refCode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrBookingNotFound
		}
		return nil, fmt.Errorf("get booking by ref_code: %w", err)
	}
	return &booking, nil
}

func (r *BookingRepo) ListByUserID(ctx context.Context, userID int64, limit, offset int) ([]domain.Booking, error) {
	var bookings []domain.Booking

	const query = `
		SELECT id, user_id, flight_id, ref_code, status,
			   total_price_cents, currency, contact_email, expires_at, created_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	if err := r.db.SelectContext(ctx, &bookings, query, userID, limit, offset); err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	return bookings, nil
}

func (r *BookingRepo) UpdateStatus(ctx context.Context, id string, status domain.BookingStatus) error {
	const query = `
		UPDATE bookings
		SET status = $1
		WHERE id = $2
	`

	res, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrBookingNotFound
	}
	return nil
}

func (r *BookingRepo) FindExpired(ctx context.Context, limit int) ([]domain.Booking, error) {
	var bookings []domain.Booking

	const query = `
		SELECT id, user_id, flight_id, ref_code, status,
			   total_price_cents, currency, contact_email, expires_at, created_at
		FROM bookings
		WHERE status = 'PENDING' AND expires_at < $1
		LIMIT $2
	`

	if err := r.db.SelectContext(ctx, &bookings, query, time.Now(), limit); err != nil {
		return nil, fmt.Errorf("find expired bookings: %w", err)
	}
	return bookings, nil
}
