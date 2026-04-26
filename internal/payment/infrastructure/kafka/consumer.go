package paymentmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/payment/config"
	"github.com/squ1ky/flyte/internal/payment/domain"
	"log/slog"
	"time"
)

type PaymentRequestDTO struct {
	BookingID   string `json:"booking_id"`
	UserID      int64  `json:"user_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type PaymentProcessor interface {
	ProcessPayment(ctx context.Context, bookingID string, userID, amountCents int64, currency string) (*domain.Payment, error)
}

type ResultSender interface {
	SendPaymentResult(ctx context.Context, payment *domain.Payment) error
}

type Consumer struct {
	reader    *kafka.Reader
	processor PaymentProcessor
	producer  ResultSender
	log       *slog.Logger
}

func NewConsumer(
	cfg config.KafkaConfig,
	processor PaymentProcessor,
	producer ResultSender,
	log *slog.Logger,
) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.TopicRequests,
		GroupID:  cfg.GroupID,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	return &Consumer{
		reader:    reader,
		processor: processor,
		producer:  producer,
		log:       log,
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
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			c.log.ErrorContext(ctx, "failed to fetch message",
				slog.Any("error", err),
			)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
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
	var req PaymentRequestDTO
	if err := json.Unmarshal(msg.Value, &req); err != nil {
		return fmt.Errorf("failed to unmarshal request: %w", err)
	}

	c.log.InfoContext(ctx, "received payment request",
		slog.String("booking_id", req.BookingID),
		slog.Int64("amount_cents", req.AmountCents),
		slog.Int64("offset", msg.Offset),
	)

	payment, err := c.processor.ProcessPayment(ctx,
		req.BookingID,
		req.UserID,
		req.AmountCents,
		req.Currency,
	)
	if err != nil {
		return fmt.Errorf("handle payment request: %w", err)
	}

	if err := c.producer.SendPaymentResult(ctx, payment); err != nil {
		return fmt.Errorf("send payment result: %w", err)
	}

	return nil
}

func (c *Consumer) Close() error {
	if c.reader != nil {
		return c.reader.Close()
	}
	return nil
}
