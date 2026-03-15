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

type AirlineRepo struct {
	db *sqlx.DB
}

func NewAirlineRepo(db *sqlx.DB) *AirlineRepo {
	return &AirlineRepo{db: db}
}

func (r *AirlineRepo) Create(ctx context.Context, airline *domain.Airline) (int64, error) {
	query := `
		INSERT INTO airlines (iata_code, name, country, logo_url)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, query,
		airline.IATACode,
		airline.Name,
		airline.Country,
		airline.LogoURL,
	).Scan(&id); err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, domain.ErrAirlineAlreadyExists
		}
		return 0, fmt.Errorf("create airline: %w", err)
	}

	return id, nil
}

func (r *AirlineRepo) GetByID(ctx context.Context, airlineID int64) (*domain.Airline, error) {
	query := `SELECT * FROM airlines WHERE id = $1`

	var a domain.Airline
	if err := r.db.GetContext(ctx, &a, query, airlineID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAirlineNotFound
		}
		return nil, fmt.Errorf("get airline: %w", err)
	}
	return &a, nil
}

func (r *AirlineRepo) ListAirlines(ctx context.Context, limit, offset int) ([]domain.Airline, error) {
	query := `
		SELECT * FROM airlines
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`

	airlines := make([]domain.Airline, 0)
	if err := r.db.SelectContext(ctx, &airlines, query, limit, offset); err != nil {
		return nil, fmt.Errorf("list airlines: %w", err)
	}

	return airlines, nil
}
