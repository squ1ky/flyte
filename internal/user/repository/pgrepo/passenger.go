package pgrepo

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/user/domain"
)

type PassengerRepo struct {
	db *sqlx.DB
}

func NewPassengerRepo(db *sqlx.DB) *PassengerRepo {
	return &PassengerRepo{db: db}
}

func (r *PassengerRepo) Create(ctx context.Context, p *domain.Passenger) (int64, error) {
	query := `
		INSERT INTO passengers (
		    user_id, first_name, last_name, middle_name,
		    birth_date, gender, document_number, document_type, citizenship
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, query,
		p.UserID, p.FirstName, p.LastName, p.MiddleName,
		p.BirthDate, p.Gender, p.DocumentNumber, p.DocumentType, p.Citizenship,
	).Scan(&id); err != nil {
		return 0, fmt.Errorf("failed to add passenger: %w", err)
	}

	return id, nil
}

func (r *PassengerRepo) GetByUserID(ctx context.Context, userID int64) ([]domain.Passenger, error) {
	var passengers []domain.Passenger

	query := `SELECT * FROM passengers WHERE user_id = $1 ORDER BY created_at DESC`
	if err := r.db.SelectContext(ctx, &passengers, query, userID); err != nil {
		return nil, fmt.Errorf("failed to get passengers by user id %d: %w", userID, err)
	}

	return passengers, nil
}
