package pgrepo_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/squ1ky/flyte/internal/booking/infrastructure/outbox"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/repository/pgrepo"
)

func newOutboxEvent(eventType outbox.EventType) outbox.Event {
	payload, _ := json.Marshal(map[string]string{"booking_id": uuid.NewString()})
	return outbox.Event{
		ID:        uuid.NewString(),
		EventType: eventType,
		Payload:   payload,
		Status:    outbox.StatusPending,
		CreatedAt: time.Now().UTC().Truncate(time.Microsecond),
	}
}

func TestOutboxRepo_Insert(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewOutboxRepo(testDB)

	ev := newOutboxEvent(outbox.EventPaymentRequest)
	err := repo.Insert(ctx, ev)
	require.NoError(t, err)

	pending, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 1)
	assert.Equal(t, ev.ID, pending[0].ID)
	assert.Equal(t, outbox.EventPaymentRequest, pending[0].EventType)
}

func TestOutboxRepo_FetchPending_OrderedByCreatedAt(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewOutboxRepo(testDB)

	ev1 := newOutboxEvent(outbox.EventPaymentRequest)
	ev1.CreatedAt = time.Now().UTC().Add(-2 * time.Second).Truncate(time.Microsecond)
	ev2 := newOutboxEvent(outbox.EventPaymentRequest)
	ev2.CreatedAt = time.Now().UTC().Add(-1 * time.Second).Truncate(time.Microsecond)

	require.NoError(t, repo.Insert(ctx, ev2))
	require.NoError(t, repo.Insert(ctx, ev1))

	pending, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)
	require.Len(t, pending, 2)
	// oldest first
	assert.Equal(t, ev1.ID, pending[0].ID)
	assert.Equal(t, ev2.ID, pending[1].ID)
}

func TestOutboxRepo_FetchPending_RespectsLimit(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewOutboxRepo(testDB)

	for range 5 {
		require.NoError(t, repo.Insert(ctx, newOutboxEvent(outbox.EventPaymentRequest)))
	}

	pending, err := repo.FetchPending(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, pending, 3)
}

func TestOutboxRepo_FetchPending_SkipsSent(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewOutboxRepo(testDB)

	ev := newOutboxEvent(outbox.EventPaymentRequest)
	require.NoError(t, repo.Insert(ctx, ev))
	require.NoError(t, repo.MarkSent(ctx, ev.ID))

	pending, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)
	assert.Empty(t, pending)
}

func TestOutboxRepo_MarkSent(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewOutboxRepo(testDB)

	ev := newOutboxEvent(outbox.EventPaymentRequest)
	require.NoError(t, repo.Insert(ctx, ev))
	require.NoError(t, repo.MarkSent(ctx, ev.ID))

	// already sent — must no longer appear in pending
	pending, err := repo.FetchPending(ctx, 10)
	require.NoError(t, err)
	assert.Empty(t, pending)
}

func TestOutboxRepo_MarkSent_NotFound(t *testing.T) {
	cleanTables(t)
	ctx := context.Background()
	repo := pgrepo.NewOutboxRepo(testDB)

	err := repo.MarkSent(ctx, uuid.NewString())
	assert.Error(t, err)
}
