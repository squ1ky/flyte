package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/squ1ky/flyte/internal/flight/domain"
	"time"
)

type ReservationRepo struct {
	db      *sqlx.DB
	lockTTL time.Duration
}

func NewReservationRepo(db *sqlx.DB, lockTTL time.Duration) *ReservationRepo {
	return &ReservationRepo{
		db:      db,
		lockTTL: lockTTL,
	}
}

func (r *ReservationRepo) ReserveSeats(ctx context.Context, params domain.SeatReservationParams) error {
	if len(params.SeatNumbers) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	expiresAt := time.Now().Add(r.lockTTL)

	reservations := make([]domain.SeatReservation, len(params.SeatNumbers))
	for i, seatNum := range params.SeatNumbers {
		reservations[i] = domain.SeatReservation{
			FlightID:   params.FlightID,
			SeatNumber: seatNum,
			BookingID:  params.BookingID,
			Status:     domain.ReservationStatusLocked,
			ExpiresAt:  &expiresAt,
		}
	}

	query := `
		INSERT INTO seat_reservations (flight_id, seat_number, booking_id, status, expires_at)
		VALUES (:flight_id, :seat_number, :booking_id, :status, :expires_at)
	`

	if _, err := tx.NamedExecContext(ctx, query, reservations); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrSeatUnavailable
		}
		return fmt.Errorf("insert seat reservations: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *ReservationRepo) ConfirmReservation(ctx context.Context, params domain.SeatReservationParams) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var confirmedCount int
	queryCheckIdempotency := `
		SELECT COUNT(*)
		FROM seat_reservations
		WHERE booking_id = $1
		  AND flight_id = $2
		  AND seat_number = ANY($3)
		  AND status = $4
	`
	err = tx.GetContext(ctx, &confirmedCount, queryCheckIdempotency,
		params.BookingID, params.FlightID, pq.Array(params.SeatNumbers),
		domain.ReservationStatusConfirmed)
	if err != nil {
		return fmt.Errorf("check idempotency: %w", err)
	}
	if confirmedCount == len(params.SeatNumbers) {
		return nil
	}

	queryUpdate := `
		UPDATE seat_reservations
		SET status = $1, expires_at = NULL
		WHERE booking_id = $2
		  AND flight_id = $3
		  AND seat_number = ANY($4)
		  AND status = $5
		  AND expires_at > NOW()
	`

	res, err := tx.ExecContext(ctx, queryUpdate,
		domain.ReservationStatusConfirmed,
		params.BookingID,
		params.FlightID,
		pq.Array(params.SeatNumbers),
		domain.ReservationStatusLocked,
	)
	if err != nil {
		return fmt.Errorf("update reservation status: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}

	if int(rowsAffected) != len(params.SeatNumbers) {
		return domain.ErrReservationExpired
	}

	queryUpdateFares := `
		UPDATE flight_fares ff
		SET seats_booked = seats_booked + sub.count
		FROM (
				SELECT s.seat_class, COUNT(*) as count
				FROM aircraft_seats s
				JOIN flights f ON f.aircraft_id = s.aircraft_id
				WHERE f.id = $1 AND s.seat_number = ANY($2)
				GROUP BY s.seat_class
		) sub
		WHERE ff.flight_id = $1 AND ff.seat_class = sub.seat_class
	`
	if _, err := tx.ExecContext(ctx, queryUpdateFares, params.FlightID, pq.Array(params.SeatNumbers)); err != nil {
		return fmt.Errorf("update fare counters: %w", err)
	}

	eventPayload := domain.SeatsStatusPayload{
		FlightID:    params.FlightID,
		BookingID:   params.BookingID,
		SeatNumbers: params.SeatNumbers,
	}
	if err := insertOutboxEvent(ctx, tx, domain.EventTypeSeatsConfirmed, eventPayload); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func (r *ReservationRepo) CancelReservation(ctx context.Context, bookingID string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	deleteQuery := `
		DELETE FROM seat_reservations
		WHERE booking_id = $1
		RETURNING flight_id, seat_number, status
	`

	var cancelled []struct {
		FlightID   int64  `db:"flight_id"`
		SeatNumber string `db:"seat_number"`
		Status     string `db:"status"`
	}
	if err := tx.SelectContext(ctx, &cancelled, deleteQuery, bookingID); err != nil {
		return fmt.Errorf("delete reservations: %w", err)
	}

	if len(cancelled) == 0 {
		return domain.ErrReservationNotFound
	}

	decrementSeatsBooked := `
		UPDATE flight_fares ff
        SET seats_booked = seats_booked - sub.count
		FROM (
			SELECT s.seat_class, COUNT(*) as count
        	FROM aircraft_seats s
			JOIN flights f on f.aircraft_id = s.aircraft_id
			WHERE f.id = $1 AND s.seat_number = ANY($2)
			GROUP BY s.seat_class
		) sub
		WHERE ff.flight_id = $1 AND ff.seat_class = sub.seat_class
	`

	var confirmedSeats []string
	var flightID int64
	for _, c := range cancelled {
		flightID = c.FlightID
		if c.Status == string(domain.ReservationStatusConfirmed) {
			confirmedSeats = append(confirmedSeats, c.SeatNumber)
		}
	}

	if len(confirmedSeats) > 0 {
		if _, err := tx.ExecContext(ctx, decrementSeatsBooked, flightID, pq.Array(confirmedSeats)); err != nil {
			return fmt.Errorf("update fare counters: %w", err)
		}
	}

	seatNumbers := make([]string, len(cancelled))
	for i, c := range cancelled {
		seatNumbers[i] = c.SeatNumber
	}

	payload := domain.SeatsStatusPayload{
		FlightID:    flightID,
		BookingID:   bookingID,
		SeatNumbers: seatNumbers,
	}
	if err := insertOutboxEvent(ctx, tx, domain.EventTypeSeatsReleased, payload); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *ReservationRepo) GetReservedSeats(ctx context.Context, flightID int64) ([]domain.SeatReservation, error) {
	query := `
		SELECT id, flight_id, seat_number, booking_id, status, expires_at
		FROM seat_reservations
		WHERE flight_id = $1
	`

	var seats []domain.SeatReservation
	if err := r.db.SelectContext(ctx, &seats, query, flightID); err != nil {
		return nil, fmt.Errorf("get reserved seats: %w", err)
	}

	return seats, nil
}

func (r *ReservationRepo) CleanExpiredReservations(ctx context.Context) (int64, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	queryDelete := `
		DELETE FROM seat_reservations
		WHERE status = $1 AND expires_at < NOW()
		RETURNING flight_id, seat_number
	`

	var deleted []struct {
		FlightID   int64  `db:"flight_id"`
		SeatNumber string `db:"seat_number"`
	}

	if err := tx.SelectContext(ctx, &deleted, queryDelete, domain.ReservationStatusLocked); err != nil {
		return 0, fmt.Errorf("delete expired locks: %w", err)
	}

	if len(deleted) == 0 {
		return 0, nil
	}

	seatsByFlight := make(map[int64][]string)
	for _, item := range deleted {
		seatsByFlight[item.FlightID] = append(seatsByFlight[item.FlightID], item.SeatNumber)
	}

	for flightID, seats := range seatsByFlight {
		payload := domain.SeatsStatusPayload{
			FlightID:    flightID,
			SeatNumbers: seats,
		}

		if err := insertOutboxEvent(ctx, tx, domain.EventTypeSeatsReleased, payload); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("commit tx: %w", err)
	}

	return int64(len(seatsByFlight)), nil
}
