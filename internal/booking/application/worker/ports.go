package worker

import (
	"context"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
)

// FlightClient is the subset of the gRPC flight client used by workers.
type FlightClient interface {
	CancelReservation(ctx context.Context, bookingID string) error
}

// PaymentProducer is the subset of the Kafka producer used by OutboxRelay.
type PaymentProducer interface {
	SendPaymentRequest(ctx context.Context, event domain.PaymentRequestEvent) error
}

type BookingEventsProducer interface {
	SendBookingEvent(ctx context.Context, bookingID string, eventType outbox.EventType, payload []byte) error
}
