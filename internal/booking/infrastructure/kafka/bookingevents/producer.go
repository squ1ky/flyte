package bookingeventsmq

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

type eventEnvelope struct {
	BookingID string          `json:"booking_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
}

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(cfg config.KafkaConfig, log *slog.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.TopicBookingEvents,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}

	return &Producer{
		writer: writer,
		log:    log,
	}
}

func (p *Producer) SendBookingEvent(ctx context.Context, bookingID string, eventType outbox.EventType, payload []byte) error {
	envelope := eventEnvelope{
		BookingID: bookingID,
		EventType: string(eventType),
		Payload:   payload,
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("marshal booking event envelope: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(bookingID),
		Value: data,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("write booking event to kafka: %w", err)
	}

	p.log.InfoContext(ctx, "booking event sent",
		slog.String("booking_id", bookingID),
		slog.String("event_type", string(eventType)),
	)

	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
