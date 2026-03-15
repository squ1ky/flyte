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

type AircraftRepo struct {
	db *sqlx.DB
}

func NewAircraftRepo(db *sqlx.DB) *AircraftRepo {
	return &AircraftRepo{db: db}
}

func (r *AircraftRepo) Create(ctx context.Context, aircraft *domain.Aircraft) (int64, error) {
	query := `
		INSERT INTO aircrafts (model, total_seats)
		VALUES ($1, $2)
		RETURNING id
	`
	var id int64
	if err := r.db.QueryRowContext(ctx, query, aircraft.Model, aircraft.TotalSeats).Scan(&id); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrAircraftAlreadyExists
		}
		return 0, fmt.Errorf("create aircraft: %w", err)
	}
	return id, nil
}

func (r *AircraftRepo) GetByID(ctx context.Context, aircraftID int64) (*domain.Aircraft, error) {
	query := "SELECT * FROM aircrafts WHERE id = $1"
	var a domain.Aircraft
	if err := r.db.GetContext(ctx, &a, query, aircraftID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAircraftNotFound
		}
		return nil, fmt.Errorf("get aircraft: %w", err)
	}
	return &a, nil
}

func (r *AircraftRepo) AddSeats(ctx context.Context, aircraftID int64, seats []domain.AircraftSeat) error {
	for i := range seats {
		seats[i].AircraftID = aircraftID
	}

	query := `
		INSERT INTO aircraft_seats (aircraft_id, seat_number, seat_class, row_number, col_number)
		VALUES (:aircraft_id, :seat_number, :seat_class, :row_number, :col_number)
		ON CONFLICT (aircraft_id, seat_number) DO UPDATE
		SET
			seat_class = EXCLUDED.seat_class,
			row_number = EXCLUDED.row_number,
			col_number = EXCLUDED.col_number
	`
	if _, err := r.db.NamedExecContext(ctx, query, seats); err != nil {
		return fmt.Errorf("batch insert seats: %w", err)
	}
	return nil
}

func (r *AircraftRepo) GetSeatsByAircraftID(ctx context.Context, aircraftID int64) ([]domain.AircraftSeat, error) {
	query := `
		SELECT * FROM aircraft_seats
		WHERE aircraft_id = $1
		ORDER BY seat_number
	`
	var seats []domain.AircraftSeat
	if err := r.db.SelectContext(ctx, &seats, query, aircraftID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.AircraftSeat{}, nil
		}
		return nil, fmt.Errorf("get aircraft seats: %w", err)
	}
	return seats, nil
}
