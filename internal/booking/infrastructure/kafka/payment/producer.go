package paymentmq

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"log/slog"
	"time"
)

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(cfg config.KafkaConfig, log *slog.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicRequests,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
		log:    log,
	}
}

func (p *Producer) SendPaymentRequest(ctx context.Context, event domain.PaymentRequestEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal payment request: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.BookingID),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write to kafka: %w", err)
	}

	p.log.InfoContext(ctx, "payment request sent",
		slog.String("booking_id", event.BookingID),
		slog.Int64("amount_cents", event.AmountCents),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
