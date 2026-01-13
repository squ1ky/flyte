package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/booking/config"
	"log/slog"
	"time"
)

type PaymentStatus string

const (
	PaymentStatusSuccess PaymentStatus = "SUCCESS"
	PaymentStatusFailed  PaymentStatus = "FAILED"
)

type PaymentResultDTO struct {
	BookingID    string        `json:"booking_id"`
	PaymentID    string        `json:"payment_id"`
	Status       PaymentStatus `json:"status"`
	ErrorMessage string        `json:"error_message,omitempty"`
	ProcessedAt  string        `json:"processed_at"`
}

type BookingConsumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	log     *slog.Logger
}

func NewBookingConsumer(
	cfg config.KafkaConfig,
	handler MessageHandler,
	log *slog.Logger,
) *BookingConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.TopicResults,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &BookingConsumer{
		reader:  reader,
		handler: handler,
		log:     log,
	}
}

func (c *BookingConsumer) Start(ctx context.Context) error {
	c.log.Info("starting kafka consumer", "topic", c.reader.Config().Topic)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.log.Error("failed to fetch message", "error", err)
			time.Sleep(time.Second)
			continue
		}

		if err := c.processMessage(ctx, m); err != nil {
			c.log.Error("failed to process message",
				"error", err,
				"offset", m.Offset)
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("failed to commit message",
				"error", err,
				"offset", m.Offset)
		}
	}
}

func (c *BookingConsumer) processMessage(ctx context.Context, m kafka.Message) error {
	var res PaymentResultDTO
	if err := json.Unmarshal(m.Value, &res); err != nil {
		return fmt.Errorf("failed to unmarshal result: %w", err)
	}

	c.log.Info("received payment result",
		"booking_id", res.BookingID,
		"status", res.Status,
		"offset", m.Offset)

	return c.handler.HandlePaymentResult(ctx, res)
}

func (c *BookingConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
