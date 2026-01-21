package kafka

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/domain/events"
	"log/slog"
)

type PaymentResultProcessor interface {
	ProcessPaymentResult(ctx context.Context, bookingID string, status events.PaymentStatus) error
}

type MessageHandler interface {
	HandlePaymentResult(ctx context.Context, res events.PaymentResultEvent) error
}

type PaymentResultHandler struct {
	service PaymentResultProcessor
	log     *slog.Logger
}

func NewPaymentResultHandler(service PaymentResultProcessor, log *slog.Logger) *PaymentResultHandler {
	return &PaymentResultHandler{
		service: service,
		log:     log,
	}
}

func (h *PaymentResultHandler) HandlePaymentResult(ctx context.Context, res events.PaymentResultEvent) error {
	err := h.service.ProcessPaymentResult(ctx, res.BookingID, res.Status)
	if err != nil {
		return fmt.Errorf("failed to process payment result: %w", err)
	}

	return nil
}
