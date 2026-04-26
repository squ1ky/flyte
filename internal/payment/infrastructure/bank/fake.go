package bank

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/squ1ky/flyte/internal/payment/config"
	"log/slog"
	"math/rand/v2"
	"time"
)

const percentScale = 100
const msgInsufficientFunds = "insufficient funds or bank error"

type FakeBank struct {
	successChancePercent int
	minDelay             time.Duration
	maxDelay             time.Duration
	log                  *slog.Logger
}

func NewFakeBank(cfg config.FakeBankConfig, log *slog.Logger) *FakeBank {
	return &FakeBank{
		successChancePercent: cfg.SuccessChancePercent,
		minDelay:             cfg.MinDelay,
		maxDelay:             cfg.MaxDelay,
		log:                  log,
	}
}

type ChargeRequest struct {
	PaymentID   string
	UserID      int64
	AmountCents int64
	Currency    string
}

type ChargeResult struct {
	Success       bool
	ProviderTxnID string
	ErrorMessage  string
}

func (b *FakeBank) Charge(ctx context.Context, req ChargeRequest) (*ChargeResult, error) {
	if err := b.simulateLatency(ctx); err != nil {
		return nil, fmt.Errorf("bank charge: %w", err)
	}

	if b.isSuccessful() {
		b.log.InfoContext(ctx, "fake bank accepted charge",
			slog.String("payment_id", req.PaymentID),
			slog.Int64("amount_cents", req.AmountCents),
		)
		return &ChargeResult{
			Success:       true,
			ProviderTxnID: "fake_txn_" + uuid.NewString(),
		}, nil
	}

	b.log.WarnContext(ctx, "fake bank rejected charge",
		slog.String("payment_id", req.PaymentID),
	)
	return &ChargeResult{
		Success:      false,
		ErrorMessage: msgInsufficientFunds,
	}, nil
}

func (b *FakeBank) simulateLatency(ctx context.Context) error {
	delay := b.minDelay
	if delta := b.maxDelay - b.minDelay; delta > 0 {
		delay += time.Duration(rand.Int64N(int64(delta)))
	}

	select {
	case <-time.After(delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *FakeBank) isSuccessful() bool {
	return rand.IntN(percentScale) < b.successChancePercent
}
