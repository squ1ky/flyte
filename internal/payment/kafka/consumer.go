package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/payment/config"
	"log/slog"
	"time"
)

type PaymentRequestDTO struct {
	BookingID string  `json:"booking_id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
}

type PaymentConsumer struct {
	reader  *kafka.Reader
	handler MessageHandler
	log     *slog.Logger
}

func NewPaymentConsumer(
	cfg config.KafkaConfig,
	handler MessageHandler,
	log *slog.Logger,
) *PaymentConsumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.TopicRequests,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &PaymentConsumer{
		reader:  reader,
		handler: handler,
		log:     log,
	}
}

func (c *PaymentConsumer) Start(ctx context.Context) error {
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
			c.log.Error("failed to process message", "error", err, "offset", m.Offset)
			// Dead Letter Queue or Retry Policy
			continue
		}

		if err := c.reader.CommitMessages(ctx, m); err != nil {
			c.log.Error("failed to commit message", "error", err, "offset", m.Offset)
		}
	}
}

func (c *PaymentConsumer) processMessage(ctx context.Context, m kafka.Message) error {
	var req PaymentRequestDTO
	if err := json.Unmarshal(m.Value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	c.log.Info("received payment request",
		"booking_id", req.BookingID,
		"amount", req.Amount,
		"offset", m.Offset)

	return c.handler.HandlePaymentRequest(ctx, req)
}

func (c *PaymentConsumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
