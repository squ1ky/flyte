package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/squ1ky/flyte/internal/booking/domain"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makePaymentEvent(bookingID string) []byte {
	payload, err := json.Marshal(domain.PaymentRequestEvent{
		BookingID:   bookingID,
		UserID:      1,
		AmountCents: 5000,
		Currency:    "RUB",
	})
	if err != nil {
		panic(err)
	}
	return payload
}

func newRelay(ob *mockOutboxRepo, p *mockPaymentProducer) *OutboxRelay {
	return &OutboxRelay{
		outbox:   ob,
		producer: p,
		logger:   discardLogger,
	}
}

func TestOutboxRelay_Poll(t *testing.T) {
	events := []outbox.Event{
		{ID: "e-1", EventType: outbox.EventPaymentRequest, Payload: makePaymentEvent("bk-1")},
		{ID: "e-2", EventType: outbox.EventPaymentRequest, Payload: makePaymentEvent("bk-2")},
	}

	cases := []struct {
		name           string
		fetchResult    []outbox.Event
		fetchErr       error
		sendErr        error
		markSentErr    error
		wantSentCount  int
		wantMarkedCount int
	}{
		{
			name:            "happy path: all events processed",
			fetchResult:     events,
			wantSentCount:   2,
			wantMarkedCount: 2,
		},
		{
			name:            "fetch error: nothing processed",
			fetchErr:        errors.New("db error"),
			wantSentCount:   0,
			wantMarkedCount: 0,
		},
		{
			name:            "send error: skips mark_sent",
			fetchResult:     events,
			sendErr:         errors.New("kafka error"),
			wantSentCount:   0,
			wantMarkedCount: 0,
		},
		{
			name:            "mark_sent error: logged but does not stop processing",
			fetchResult:     events,
			markSentErr:     errors.New("db error"),
			wantSentCount:   2,
			wantMarkedCount: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sentCount := 0
			markedCount := 0

			ob := &mockOutboxRepo{
				fetchPendingFn: func(_ context.Context, _ int) ([]outbox.Event, error) {
					return tc.fetchResult, tc.fetchErr
				},
				markSentFn: func(_ context.Context, _ string) error {
					markedCount++
					return tc.markSentErr
				},
			}
			p := &mockPaymentProducer{
				sendFn: func(_ context.Context, _ domain.PaymentRequestEvent) error {
					if tc.sendErr == nil {
						sentCount++
					}
					return tc.sendErr
				},
			}

			relay := newRelay(ob, p)
			relay.poll(context.Background())

			assert.Equal(t, tc.wantSentCount, sentCount)
			assert.Equal(t, tc.wantMarkedCount, markedCount)
		})
	}
}

func TestOutboxRelay_ProcessEvent(t *testing.T) {
	t.Run("valid payload sent to producer", func(t *testing.T) {
		var received domain.PaymentRequestEvent
		p := &mockPaymentProducer{
			sendFn: func(_ context.Context, e domain.PaymentRequestEvent) error {
				received = e
				return nil
			},
		}
		relay := newRelay(&mockOutboxRepo{}, p)

		payload := makePaymentEvent("bk-42")
		err := relay.processEvent(context.Background(), outbox.Event{
			ID:        "e-1",
			EventType: outbox.EventPaymentRequest,
			Payload:   payload,
		})

		require.NoError(t, err)
		assert.Equal(t, "bk-42", received.BookingID)
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		relay := newRelay(&mockOutboxRepo{}, &mockPaymentProducer{})
		err := relay.processEvent(context.Background(), outbox.Event{
			Payload: []byte("not json"),
		})
		assert.Error(t, err)
	})
}
