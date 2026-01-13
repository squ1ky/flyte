package kafka

import (
	"context"
	"fmt"
	"log/slog"
)

type PaymentResultProcessor interface {
	ProcessResult(ctx context.Context, bookingID string, status PaymentStatus) error
}

type MessageHandler interface {
	HandlePaymentResult(ctx context.Context, res PaymentResultDTO) error
}

type BookingMessageHandler struct {
	service PaymentResultProcessor
	log     *slog.Logger
}

func NewBookingMessageHandler(service PaymentResultProcessor, log *slog.Logger) *BookingMessageHandler {
	return &BookingMessageHandler{
		service: service,
		log:     log,
	}
}

func (h *BookingMessageHandler) HandlePaymentResult(ctx context.Context, res PaymentResultDTO) error {
	err := h.service.ProcessResult(ctx, res.BookingID, res.Status)
	if err != nil {
		return fmt.Errorf("failed to process payment result: %w", err)
	}

	return nil
}
