package paymentmq

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

type Producer struct {
	writer *kafka.Writer
	log    *slog.Logger
}

func NewProducer(cfg config.KafkaConfig, log *slog.Logger) *Producer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(cfg.Brokers...),
		Topic:    cfg.TopicResults,
		Balancer: &kafka.LeastBytes{},
	}

	return &Producer{
		writer: writer,
		log:    log,
	}
}

func (p *Producer) SendPaymentResult(ctx context.Context, payment *domain.Payment) error {
	resp := PaymentResultDTO{
		BookingID: payment.BookingID,
		PaymentID: payment.ID,
		Status:    string(payment.Status),
	}

	if payment.ErrorMessage != nil {
		resp.ErrorMessage = *payment.ErrorMessage
	}

	if payment.ProcessedAt != nil {
		resp.ProcessedAt = *payment.ProcessedAt
	} else {
		resp.ProcessedAt = time.Now()
	}

	payload, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal payment result: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(payment.BookingID),
		Value: payload,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		return fmt.Errorf("failed to write response to kafka: %w", err)
	}

	p.log.InfoContext(ctx, "payment result sent",
		slog.String("booking_id", payment.BookingID),
		slog.String("payment_id", payment.ID),
		slog.String("status", resp.Status),
	)

	return nil
}

func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}
