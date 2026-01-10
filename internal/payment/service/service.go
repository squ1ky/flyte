package service

import (
	"context"
	"crypto/rand"
	"github.com/squ1ky/flyte/internal/payment/domain"
	"github.com/squ1ky/flyte/internal/payment/repository"
	"log/slog"
	"math/big"
	"time"
)

const (
	ProbabilityBase   = 100
	BankSuccessChance = 80
	BankMinDelay      = 500 * time.Millisecond
	BankMaxDelay      = 2000 * time.Millisecond
)

type PaymentService struct {
	repo repository.PaymentRepository
	log  *slog.Logger
}

func NewPaymentService(repo repository.PaymentRepository, log *slog.Logger) *PaymentService {
	return &PaymentService{
		repo: repo,
		log:  log,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, bookingID, userID string, amount float64, currency string) (*domain.Payment, error) {
	payment := &domain.Payment{
		BookingID: bookingID,
		UserID:    userID,
		Amount:    amount,
		Currency:  currency,
		Status:    domain.PaymentStatusPending,
	}

	result, err := s.repo.CreateOrGet(ctx, payment)
	if err != nil {
		s.log.Error("failed to create or get payment",
			"error", err,
			"booking_id", bookingID)
		return nil, err
	}

	currentPayment := result.Payment

	if !result.IsNew {
		s.log.Info("payment request duplicate, returning existing status",
			"booking_id", bookingID,
			"status", currentPayment.Status,
		)
		return currentPayment, nil
	}

	s.simulateBankLatency()

	var newStatus domain.PaymentStatus
	var errorMsgPtr *string

	if s.isBankSuccessful() {
		newStatus = domain.PaymentStatusSuccess
		s.log.Info("bank accepted payment", "booking_id", bookingID)
	} else {
		newStatus = domain.PaymentStatusFailed
		msg := "insufficient funds or bank error"
		errorMsgPtr = &msg
		s.log.Warn("bank rejected payment",
			"booking_id", bookingID,
			"reason", msg)
	}

	if err := s.repo.UpdateStatus(ctx, currentPayment.ID, newStatus, errorMsgPtr); err != nil {
		s.log.Error("failed to update payment status",
			"error", err,
			"booking_id", bookingID,
			"status", newStatus)
		return nil, err
	}

	currentPayment.Status = newStatus
	currentPayment.ErrorMessage = errorMsgPtr
	now := time.Now()
	currentPayment.ProcessedAt = &now

	return currentPayment, nil
}

func (s *PaymentService) simulateBankLatency() {
	delta := int64(BankMaxDelay - BankMinDelay)
	if delta <= 0 {
		time.Sleep(BankMinDelay)
		return
	}

	randomDuration := s.mustCryptoRandInt64(delta)
	time.Sleep(BankMinDelay + time.Duration(randomDuration))
}

func (s *PaymentService) isBankSuccessful() bool {
	val := s.mustCryptoRandInt64(ProbabilityBase)
	return val < int64(BankSuccessChance)
}

func (s *PaymentService) mustCryptoRandInt64(max int64) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		s.log.Error("crypto/rand failed", "error", err)
		return 0
	}
	return n.Int64()
}
