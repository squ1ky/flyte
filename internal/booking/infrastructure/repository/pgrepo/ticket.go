package pgrepo

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/booking/domain"
)

type TicketRepo struct {
	db *sqlx.DB
}

func NewTicketRepo(db *sqlx.DB) *TicketRepo {
	return &TicketRepo{db: db}
}

func (r *TicketRepo) ListByBookingID(ctx context.Context, bookingID string) ([]domain.Ticket, error) {
	var tickets []domain.Ticket

	const query = `
		SELECT id, booking_id, passenger_first_name, passenger_last_name,
			   passenger_doc_num, seat_number, price_cents, ticket_number, status
		FROM tickets
		WHERE booking_id = $1
	`
	if err := r.db.SelectContext(ctx, &tickets, query, bookingID); err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	return tickets, nil
}

func (r *TicketRepo) IssueTickets(ctx context.Context, bookingID string, ticketNumbers map[string]string) error {
	if len(ticketNumbers) == 0 {
		return nil
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	const query = `
		UPDATE tickets
		SET ticket_number = $1, status = 'ISSUED'
		WHERE booking_id = $2 AND id = $3
	`

	for ticketID, ticketNum := range ticketNumbers {
		res, err := tx.ExecContext(ctx, query, ticketNum, bookingID, ticketID)
		if err != nil {
			return fmt.Errorf("issue ticket %s: %w", ticketID, err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("ticket %s not found in booking %s", ticketID, bookingID)
		}
	}

	return tx.Commit()
}
