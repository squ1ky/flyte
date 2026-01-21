package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/domain/events"
	"github.com/squ1ky/flyte/internal/booking/kafka"
	"github.com/squ1ky/flyte/internal/booking/repository"
	"log/slog"
	"time"
)

type OutboxProcessor struct {
	repo     repository.BookingRepository
	producer *kafka.PaymentEventProducer
	log      *slog.Logger
	interval time.Duration
}

func NewOutboxProcessor(
	repo repository.BookingRepository,
	producer *kafka.PaymentEventProducer,
	log *slog.Logger,
	interval time.Duration,
) *OutboxProcessor {
	return &OutboxProcessor{
		repo:     repo,
		producer: producer,
		log:      log,
		interval: interval,
	}
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	p.log.Info("starting outbox processor", "interval", p.interval)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.log.Info("stopping outbox processor")
			return
		case <-ticker.C:
			if err := p.processBatch(ctx); err != nil {
				p.log.Error("failed to process batch", "error", err)
			}
		}
	}
}

func (p *OutboxProcessor) processBatch(ctx context.Context) error {
	batchSize := 10
	pendingEvents, err := p.repo.GetPendingOutboxEvents(ctx, batchSize)
	if err != nil {
		return fmt.Errorf("fetch events: %w", err)
	}

	if len(pendingEvents) == 0 {
		return nil
	}

	for _, event := range pendingEvents {
		log := p.log.With("outbox_id", event.ID, "type", event.EventType)

		var paymentEvent events.PaymentRequestEvent

		if err := json.Unmarshal(event.Payload, &paymentEvent); err != nil {
			log.Error("invalid payload json", "error", err)
			reason := fmt.Sprintf("invalid json: %v", err)
			if markErr := p.repo.MarkOutboxEventFailed(ctx, event.ID, reason); markErr != nil {
				log.Error("failed to mark invalid event as failed", "error", markErr)
			}
			continue
		}

		if err := p.producer.SendPaymentRequest(ctx, paymentEvent); err != nil {
			log.Error("failed to publish to kafka", "error", err)
			continue
		}

		if err := p.repo.MarkOutboxEventProcessed(ctx, event.ID); err != nil {
			log.Error("failed to mark processed", "error", err)
		}
	}

	return nil
}
