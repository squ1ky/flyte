package kafka

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/payment/service"
	"log/slog"
)

type MessageHandler interface {
	HandlePaymentRequest(ctx context.Context, req PaymentRequestDTO) error
}

type PaymentMessageHandler struct {
	service  *service.PaymentService
	producer *PaymentProducer
	log      *slog.Logger
}

func NewPaymentMessageHandler(
	service *service.PaymentService,
	producer *PaymentProducer,
	log *slog.Logger,
) *PaymentMessageHandler {
	return &PaymentMessageHandler{
		service:  service,
		producer: producer,
		log:      log,
	}
}

func (h *PaymentMessageHandler) HandlePaymentRequest(ctx context.Context, req PaymentRequestDTO) error {
	payment, err := h.service.ProcessPayment(ctx, req.BookingID, req.UserID, req.AmountCents, req.Currency)
	if err != nil {
		return fmt.Errorf("server processing error: %w", err)
	}

	if err := h.producer.SendPaymentResult(ctx, payment); err != nil {
		return fmt.Errorf("failed to send result: %w", err)
	}

	return nil
}
