package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

type AircraftRepo struct {
	db *sqlx.DB
}

func NewAircraftRepo(db *sqlx.DB) *AircraftRepo {
	return &AircraftRepo{db: db}
}

func (r *AircraftRepo) CreateAircraft(ctx context.Context, model string, totalSeats int) (int64, error) {
	query := `
		INSERT INTO aircrafts (model, total_seats) VALUES ($1, $2)
		RETURNING id;
	`
	var id int64
	if err := r.db.QueryRowContext(ctx, query, model, totalSeats).Scan(&id); err != nil {
		return 0, fmt.Errorf("create aircraft: %w", err)
	}
	return id, nil
}

func (r *AircraftRepo) GetAircraftByID(ctx context.Context, id int64) (*domain.Aircraft, error) {
	query := "SELECT * FROM aircrafts WHERE id = $1"
	var a domain.Aircraft
	if err := r.db.GetContext(ctx, &a, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAircraftNotFound
		}
		return nil, fmt.Errorf("get aircraft: %w", err)
	}
	return &a, nil
}

func (r *AircraftRepo) GetAircrafts(ctx context.Context) ([]domain.Aircraft, error) {
	query := "SELECT * FROM aircrafts ORDER BY model"
	var list []domain.Aircraft
	if err := r.db.SelectContext(ctx, &list, query); err != nil {
		return nil, fmt.Errorf("get aircrafts: %w", err)
	}
	return list, nil
}

func (r *AircraftRepo) AddAircraftSeats(ctx context.Context, aircraftID int64, seats []domain.AircraftSeat) error {
	for i := range seats {
		seats[i].AircraftID = aircraftID
	}

	query := `
		INSERT INTO aircraft_seats (aircraft_id, seat_number, seat_class, price_multiplier)
		VALUES (:aircraft_id, :seat_number, :seat_class, :price_multiplier)
		ON CONFLICT (aircraft_id, seat_number) DO UPDATE
		SET seat_class = EXCLUDED.seat_class, price_multiplier = EXCLUDED.price_multiplier                                        
	`
	if _, err := r.db.NamedExecContext(ctx, query, seats); err != nil {
		return fmt.Errorf("batch insert seats: %w", err)
	}
	return nil
}

func (r *AircraftRepo) GetAircraftSeats(ctx context.Context, aircraftID int64) ([]domain.AircraftSeat, error) {
	query := `
		SELECT * FROM aircraft_seats
		WHERE aircraft_id = $1
		ORDER BY seat_number
	`
	var seats []domain.AircraftSeat
	if err := r.db.SelectContext(ctx, &seats, query, aircraftID); err != nil {
		return nil, fmt.Errorf("get aircraft seats: %w", err)
	}
	return seats, nil
}
