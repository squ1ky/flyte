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

type PaymentResultHandler interface {
	HandlePaymentResult(ctx context.Context, event domain.PaymentResultEvent) error
}

type Consumer struct {
	reader  *kafka.Reader
	handler PaymentResultHandler
	log     *slog.Logger
}

func NewConsumer(
	cfg config.KafkaConfig,
	handler PaymentResultHandler,
	log *slog.Logger,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.TopicResults,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:  reader,
		handler: handler,
		log:     log,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	c.log.InfoContext(ctx, "kafka payment consumer started",
		slog.String("topic", c.reader.Config().Topic),
	)

	for {
		select {
		case <-ctx.Done():
			c.log.InfoContext(ctx, "kafka payment consumer stopping")
			return ctx.Err()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			c.log.ErrorContext(ctx, "fetch message failed",
				slog.Any("error", err),
			)
			time.Sleep(time.Second)
			continue
		}

		if err := c.processMessage(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "process message failed",
				slog.Any("error", err),
				slog.Int64("offset", msg.Offset),
			)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.log.ErrorContext(ctx, "commit failed",
				slog.Any("error", err),
				slog.Int64("offset", msg.Offset),
			)
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) error {
	var event domain.PaymentResultEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("unmarshal payment result: %w", err)
	}

	c.log.InfoContext(ctx, "received payment result",
		slog.String("booking_id", event.BookingID),
		slog.String("status", string(event.Status)),
		slog.Int64("offset", msg.Offset),
	)

	return c.handler.HandlePaymentResult(ctx, event)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
