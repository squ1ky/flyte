package pgrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/flight/domain"
)

type AirportRepo struct {
	db *sqlx.DB
}

func NewAirportRepo(db *sqlx.DB) *AirportRepo {
	return &AirportRepo{db: db}
}

func (r *AirportRepo) GetByCode(ctx context.Context, code string) (*domain.Airport, error) {
	query := `SELECT * FROM airports WHERE code = $1;`

	var a domain.Airport
	if err := r.db.GetContext(ctx, &a, query, code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrAirportNotFound
		}
		return nil, fmt.Errorf("get airport: %w", err)
	}

	return &a, nil
}

func (r *AirportRepo) SearchAirports(ctx context.Context, searchQuery string) ([]domain.Airport, error) {
	query := `
		SELECT * FROM airports
		WHERE
		    code ILIKE $1 OR
		    name ILIKE $1 OR
		    city ILIKE $1
		ORDER BY
		    CASE
				WHEN code ILIKE $1 THEN 1
				ELSE 2
			END,
			name
		LIMIT 20
	`

	param := "%" + searchQuery + "%"

	var airports []domain.Airport
	if err := r.db.SelectContext(ctx, &airports, query, param); err != nil {
		return nil, fmt.Errorf("search airports: %w", err)
	}
	return airports, nil
}
