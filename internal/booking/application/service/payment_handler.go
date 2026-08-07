package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

type PaymentHandler struct {
	bookings domain.BookingRepository
	tickets  domain.TicketRepository
	flights  FlightClient
	outbox   outbox.Repository
	logger   *slog.Logger
}

func NewPaymentHandler(
	bookings domain.BookingRepository,
	tickets domain.TicketRepository,
	flights FlightClient,
	outbox outbox.Repository,
	logger *slog.Logger,
) *PaymentHandler {
	return &PaymentHandler{
		bookings: bookings,
		tickets:  tickets,
		flights:  flights,
		outbox:   outbox,
		logger:   logger,
	}
}

func (h *PaymentHandler) HandlePaymentResult(ctx context.Context, event domain.PaymentResultEvent) error {
	log := h.logger.With(
		slog.String("booking_id", event.BookingID),
		slog.String("payment_status", string(event.Status)),
	)

	booking, err := h.bookings.GetByID(ctx, event.BookingID)
	if err != nil {
		return fmt.Errorf("get booking: %w", err)
	}

	if booking.Status.IsTerminal() {
		log.WarnContext(ctx, "booking already in terminal status, skipping",
			slog.String("current_status", string(booking.Status)),
		)
		return nil
	}

	switch event.Status {
	case domain.PaymentStatusSuccess:
		return h.onSuccess(ctx, booking, log)
	case domain.PaymentStatusFailed:
		return h.onFailure(ctx, booking, log)
	default:
		return fmt.Errorf("unknown payment status: %s", string(event.Status))
	}
}

func (h *PaymentHandler) onSuccess(ctx context.Context, booking *domain.Booking, log *slog.Logger) error {
	log.InfoContext(ctx, "payment succeeded, confirm booking")

	tickets, err := h.tickets.ListByBookingID(ctx, booking.ID)
	if err != nil {
		return fmt.Errorf("list tickets: %w", err)
	}

	seatNumbers := make([]string, len(tickets))
	for i, ticket := range tickets {
		seatNumbers[i] = ticket.SeatNumber
	}

	if err := h.flights.ConfirmReservation(ctx, booking.FlightID, seatNumbers, booking.ID); err != nil {
		log.ErrorContext(ctx, "failed to confirm reservation in flight service",
			slog.Any("error", err),
		)
		return fmt.Errorf("confirm reservation: %w", err)
	}

	ticketNumbers := make(map[string]string, len(tickets))
	for _, ticket := range tickets {
		ticketNumbers[ticket.ID] = generateTicketNumber()
	}

	if err := h.tickets.IssueTickets(ctx, booking.ID, ticketNumbers); err != nil {
		log.ErrorContext(ctx, "issue tickets failed",
			slog.Any("error", err),
		)
		return fmt.Errorf("issue tickets: %w", err)
	}

	if err := h.bookings.UpdateStatus(ctx, booking.ID, domain.BookingStatusPaid, nil); err != nil {
		return fmt.Errorf("update status to paid: %w", err)
	}

	if err := h.publishBookingPaid(ctx, booking.ID); err != nil {
		log.ErrorContext(ctx, "publish booking paid event failed",
			slog.Any("error", err),
		)
	}

	log.InfoContext(ctx, "booking confirmed, tickets issued",
		slog.Int("ticket_count", len(tickets)),
	)
	return nil
}

func (h *PaymentHandler) onFailure(ctx context.Context, booking *domain.Booking, log *slog.Logger) error {
	log.WarnContext(ctx, "payment failed, cancelling booking")

	reason := domain.CancelReasonPaymentFailed
	if err := h.bookings.UpdateStatus(ctx, booking.ID, domain.BookingStatusCancelled, &reason); err != nil {
		return fmt.Errorf("update status to cancelled: %w", err)
	}

	if err := h.publishBookingCancelled(ctx, booking.ID, reason); err != nil {
		log.ErrorContext(ctx, "publish booking cancelled event failed",
			slog.Any("error", err),
		)
	}

	if err := h.flights.CancelReservation(ctx, booking.ID); err != nil {
		log.ErrorContext(ctx, "compensation failed: cancel reservation",
			slog.Any("error", err),
		)
	}

	log.InfoContext(ctx, "booking cancelled after payment failure")
	return nil
}

func (h *PaymentHandler) publishBookingPaid(ctx context.Context, bookingID string) error {
	event := domain.BookingPaidEvent{
		BookingID: bookingID,
		PaidAt:    time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal booking paid event: %w", err)
	}

	return h.outbox.Insert(ctx, outbox.Event{
		ID:        uuid.New().String(),
		EventType: outbox.EventBookingPaid,
		Payload:   payload,
		Status:    outbox.StatusPending,
		CreatedAt: time.Now(),
	})
}

func (h *PaymentHandler) publishBookingCancelled(ctx context.Context, bookingID string, reason domain.CancelReason) error {
	event := domain.BookingCancelledEvent{
		BookingID:   bookingID,
		Reason:      reason,
		CancelledAt: time.Now(),
	}

	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal booking cancelled event: %w", err)
	}

	return h.outbox.Insert(ctx, outbox.Event{
		ID:        uuid.New().String(),
		EventType: outbox.EventBookingCancelled,
		Payload:   payload,
		Status:    outbox.StatusPending,
		CreatedAt: time.Now(),
	})
}

func generateTicketNumber() string {
	return fmt.Sprintf("%010d", rand.Int63n(10_000_000_000))
}
