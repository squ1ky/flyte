package experiment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrNoPaymentAfter is returned when no payment row appears within the wait window.
var ErrNoPaymentAfter = errors.New("no payment created after injection")

// PaymentProbe reads payment-db to timestamp the moment the fault starts actually
// affecting the system: the first payment row created after the injection, i.e. the
// first payment the restarted (faulty) payment-service instance takes on. The payment
// row is inserted before the bank call, so its created_at is when the broken bank
// began handling the payment — for payment_outage this is the first refusal, for
// bank_timeout it is the first booking whose expiry timer started under the slow bank.
type PaymentProbe struct {
	db *pgx.Conn
}

func NewPaymentProbe(ctx context.Context, dbURL string) (*PaymentProbe, error) {
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("connect payment-db: %w", err)
	}
	return &PaymentProbe{db: conn}, nil
}

func (p *PaymentProbe) Close(ctx context.Context) error {
	return p.db.Close(ctx)
}

// ClockSkew returns payment-db clock minus host clock (round-trip midpoint estimate).
func (p *PaymentProbe) ClockSkew(ctx context.Context) (time.Duration, error) {
	before := time.Now()
	var dbNow time.Time
	if err := p.db.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&dbNow); err != nil {
		return 0, err
	}
	after := time.Now()
	midpoint := before.Add(after.Sub(before) / 2)
	return dbNow.Sub(midpoint), nil
}

// FirstPaymentAfter polls payment-db every interval until the first payment row with
// created_at > since (payment-db clock) appears, and returns its created_at
// (payment-db clock). Times out after timeout with ErrNoPaymentAfter.
func (p *PaymentProbe) FirstPaymentAfter(ctx context.Context, since time.Time, interval, timeout time.Duration) (time.Time, error) {
	deadline := time.Now().Add(timeout)
	for {
		var first *time.Time
		err := p.db.QueryRow(ctx,
			`SELECT min(created_at) FROM payments WHERE created_at > $1`, since).Scan(&first)
		if err != nil {
			return time.Time{}, fmt.Errorf("query first payment after injection: %w", err)
		}
		if first != nil {
			return *first, nil
		}
		if time.Now().After(deadline) {
			return time.Time{}, ErrNoPaymentAfter
		}
		if err := sleep(ctx, interval); err != nil {
			return time.Time{}, err
		}
	}
}
