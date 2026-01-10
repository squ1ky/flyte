package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/segmentio/kafka-go"
	"github.com/squ1ky/flyte/internal/payment/config"
	"github.com/squ1ky/flyte/internal/payment/domain"
	"log/slog"
	"time"
)

type PaymentResultDTO struct {
	BookingID    string    `json:"booking_id"`
	PaymentID    string    `json:"payment_id"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	ProcessedAt  time.Time `json:"processed_at"`
}

type PaymentProducer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewPaymentProducer(cfg config.KafkaConfig, log *slog.Logger) *PaymentProducer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicResults,
		Balancer: &kafka.LeastBytes{},
	}

	return &PaymentProducer{
		writer: writer,
		log:    log,
	}
}

func (p *PaymentProducer) SendPaymentResult(ctx context.Context, payment *domain.Payment) error {
	resp := PaymentResultDTO{
		BookingID:   payment.BookingID,
		PaymentID:   payment.ID,
		Status:      string(payment.Status),
		ProcessedAt: time.Now(),
	}

	if payment.ErrorMessage != nil {
		resp.ErrorMessage = *payment.ErrorMessage
	}

	respBytes, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(payment.BookingID),
		Value: respBytes,
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write response to kafka: %w", err)
	}

	p.log.Info("payment result send",
		"booking_id", payment.BookingID,
		"payment_id", payment.ID,
		"status", resp.Status)

	return nil
}

func (p *PaymentProducer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
