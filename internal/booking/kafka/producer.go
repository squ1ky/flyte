package kafka

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

type PaymentRequestDTO struct {
	BookingID   string `json:"booking_id"`
	UserID      int64  `json:"user_id"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
}

type BookingProducer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(cfg config.KafkaConfig, log *slog.Logger) *BookingProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicRequests,
		Balancer: &kafka.LeastBytes{},
	}

	return &BookingProducer{
		writer: writer,
		log:    log,
	}
}

func (p *BookingProducer) SendPaymentRequest(ctx context.Context, booking *domain.Booking) error {
	req := PaymentRequestDTO{
		BookingID:   booking.ID,
		UserID:      booking.UserID,
		AmountCents: booking.PriceCents,
		Currency:    booking.Currency,
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(booking.ID),
		Value: reqBytes,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write request to kafka: %w", err)
	}

	p.log.Info("payment request sent",
		"booking_id", booking.ID,
		"amount", booking.PriceCents)

	return nil
}

func (p *BookingProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
