package service

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/payment/domain"
	"github.com/squ1ky/flyte/internal/payment/infrastructure/bank"
	"log/slog"
)

type BankGateway interface {
	Charge(ctx context.Context, req bank.ChargeRequest) (*bank.ChargeResult, error)
}

type PaymentService struct {
	repo domain.PaymentRepository
	bank BankGateway
	log  *slog.Logger
}

func NewPaymentService(
	repo domain.PaymentRepository,
	bank BankGateway,
	log *slog.Logger,
) *PaymentService {
	return &PaymentService{
		repo: repo,
		bank: bank,
		log:  log,
	}
}

func (s *PaymentService) ProcessPayment(
	ctx context.Context,
	bookingID string,
	userID, amountCents int64,
	currency string,
) (*domain.Payment, error) {
	payment := &domain.Payment{
		BookingID:   bookingID,
		UserID:      userID,
		AmountCents: amountCents,
		Currency:    currency,
		Status:      domain.PaymentStatusPending,
	}

	result, err := s.repo.CreateOrGet(ctx, payment)
	if err != nil {
		s.log.ErrorContext(ctx, "failed to create or get payment",
			slog.Any("error", err),
			slog.String("booking_id", bookingID),
		)
		return nil, fmt.Errorf("create or get payment: %w", err)
	}

	currentPayment := result.Payment

	if !result.IsNew && currentPayment.Status != domain.PaymentStatusPending {
		s.log.InfoContext(ctx, "payment already processed, returning existing",
			slog.String("booking_id", bookingID),
			slog.String("payment_id", currentPayment.ID),
			slog.String("status", string(currentPayment.Status)),
		)
		return currentPayment, nil
	}

	if !result.IsNew {
		s.log.WarnContext(ctx, "found pending payment, resuming processing",
			slog.String("booking_id", bookingID),
			slog.String("payment_id", currentPayment.ID),
		)
	}

	chargeResult, err := s.bank.Charge(ctx, bank.ChargeRequest{
		PaymentID:   currentPayment.ID,
		UserID:      currentPayment.UserID,
		AmountCents: currentPayment.AmountCents,
		Currency:    currentPayment.Currency,
	})
	if err != nil {
		s.log.ErrorContext(ctx, "bank charge failed with transport error",
			slog.Any("error", err),
			slog.String("payment_id", currentPayment.ID),
		)
		return nil, fmt.Errorf("bank charge: %w", err)
	}

	newStatus, providerTxnID, errorMsg := mapChargeResult(chargeResult)

	if err := s.repo.UpdateStatus(ctx, currentPayment.ID, newStatus, providerTxnID, errorMsg); err != nil {
		s.log.ErrorContext(ctx, "failed to update payment status",
			slog.Any("error", err),
			slog.String("payment_id", currentPayment.ID),
			slog.String("status", string(newStatus)),
		)
		return nil, fmt.Errorf("update payment status: %w", err)
	}

	currentPayment.Status = newStatus
	currentPayment.ProviderTxnID = providerTxnID
	currentPayment.ErrorMessage = errorMsg

	s.log.InfoContext(ctx, "payment processed",
		slog.String("booking_id", bookingID),
		slog.String("payment_id", currentPayment.ID),
		slog.String("status", string(newStatus)),
	)

	return currentPayment, nil
}

func mapChargeResult(result *bank.ChargeResult) (
	status domain.PaymentStatus,
	providerTxnID *string,
	errorMsg *string,
) {
	if result.Success {
		txnID := result.ProviderTxnID
		return domain.PaymentStatusSuccess, &txnID, nil
	}

	msg := result.ErrorMessage
	if msg == "" {
		msg = "payment declined"
	}
	return domain.PaymentStatusFailed, nil, &msg
}
