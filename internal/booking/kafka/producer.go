package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/domain/events"
	"log/slog"
	"time"
)

type PaymentEventProducer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewPaymentEventProducer(cfg config.KafkaConfig, log *slog.Logger) *PaymentEventProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicRequests,
		Balancer: &kafka.LeastBytes{},
	}

	return &PaymentEventProducer{
		writer: writer,
		log:    log,
	}
}

func (p *PaymentEventProducer) SendPaymentRequest(ctx context.Context, event events.PaymentRequestEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal payment request event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(event.BookingID),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	return nil
}

func (p *PaymentEventProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
