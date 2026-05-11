package worker

import (
	"context"
	"github.com/squ1ky/flyte/internal/booking/domain"
)

// FlightClient is the subset of the gRPC flight client used by workers.
type FlightClient interface {
	CancelReservation(ctx context.Context, bookingID string) error
}

// PaymentProducer is the subset of the Kafka producer used by OutboxRelay.
type PaymentProducer interface {
	SendPaymentRequest(ctx context.Context, event domain.PaymentRequestEvent) error
}
