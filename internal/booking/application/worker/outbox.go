package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

const outboxBatchSize = 50

type OutboxRelay struct {
	outbox                outbox.Repository
	paymentProducer       PaymentProducer
	bookingEventsProducer BookingEventsProducer
	interval              time.Duration
	logger                *slog.Logger
}

func NewOutboxRelay(
	outbox outbox.Repository,
	paymentProducer PaymentProducer,
	bookingEventsProducer BookingEventsProducer,
	cfg config.OutboxConfig,
	log *slog.Logger,
) *OutboxRelay {
	return &OutboxRelay{
		outbox:                outbox,
		paymentProducer:       paymentProducer,
		bookingEventsProducer: bookingEventsProducer,
		interval:              cfg.Interval,
		logger:                log,
	}
}

func (r *OutboxRelay) Run(ctx context.Context) error {
	r.logger.InfoContext(ctx, "outbox relay started",
		slog.String("interval", r.interval.String()),
	)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			r.logger.InfoContext(ctx, "outbox relay stopping")
			return ctx.Err()
		case <-ticker.C:
			r.poll(ctx)
		}
	}
}

func (r *OutboxRelay) poll(ctx context.Context) {
	events, err := r.outbox.FetchPending(ctx, outboxBatchSize)
	if err != nil {
		r.logger.ErrorContext(ctx, "fetch pending outbox events failed",
			slog.Any("error", err),
		)
		return
	}

	for _, event := range events {
		log := r.logger.With(slog.String("event_id", event.ID))

		if err := r.processEvent(ctx, event); err != nil {
			log.ErrorContext(ctx, "process outbox event failed",
				slog.Any("error", err),
			)
			continue
		}

		if err := r.outbox.MarkSent(ctx, event.ID); err != nil {
			log.ErrorContext(ctx, "mark outbox event sent failed",
				slog.Any("error", err),
			)
		}
	}
}

func (r *OutboxRelay) processEvent(ctx context.Context, event outbox.Event) error {
	switch event.EventType {
	case outbox.EventPaymentRequest:
		var paymentReq domain.PaymentRequestEvent
		if err := json.Unmarshal(event.Payload, &paymentReq); err != nil {
			return err
		}
		return r.paymentProducer.SendPaymentRequest(ctx, paymentReq)

	case outbox.EventBookingCreated, outbox.EventBookingPaid, outbox.EventBookingCancelled:
		bookingID, err := extractBookingID(event.Payload)
		if err != nil {
			return fmt.Errorf("extract booking_id: %w", err)
		}
		return r.bookingEventsProducer.SendBookingEvent(ctx, bookingID, event.EventType, event.Payload)

	default:
		return fmt.Errorf("unknown outbox event type: %s", event.EventType)
	}
}

func extractBookingID(payload []byte) (string, error) {
	var partial struct {
		BookingID string `json:"booking_id"`
	}
	if err := json.Unmarshal(payload, &partial); err != nil {
		return "", err
	}
	if partial.BookingID == "" {
		return "", fmt.Errorf("booking_id is empty")
	}
	return partial.BookingID, nil
}
