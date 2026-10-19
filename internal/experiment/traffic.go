package experiment

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

type TrafficConfig struct {
	// Bookings created per second, at a constant rate.
	Rate float64
	// Fraction of bookings the "user" cancels shortly after creating them (user_cancelled background noise).
	CancelProbability float64
	CancelDelay       time.Duration
}

// TrafficGenerator keeps a steady stream of bookings going through the real gateway for the
// whole experiment, so that the system is in a steady state before, during and after each fault.
// Fault injection only changes how the system reacts to this traffic, never the traffic itself.
type TrafficGenerator struct {
	cfg    TrafficConfig
	gw     *GatewayClient
	pool   *SeatPool
	logger *slog.Logger

	seq       atomic.Int64
	created   atomic.Int64
	failed    atomic.Int64
	cancelled atomic.Int64
	wg        sync.WaitGroup
}

func NewTrafficGenerator(cfg TrafficConfig, gw *GatewayClient, pool *SeatPool, logger *slog.Logger) *TrafficGenerator {
	return &TrafficGenerator{cfg: cfg, gw: gw, pool: pool, logger: logger}
}

// Run blocks until ctx is cancelled, then waits for in-flight requests.
func (t *TrafficGenerator) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(float64(time.Second) / t.cfg.Rate))
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.wg.Wait()
			return
		case <-ticker.C:
			t.wg.Add(1)
			// One goroutine per booking: a slow gateway response must not lower the offered rate.
			go func() {
				defer t.wg.Done()
				t.createOne(ctx)
			}()
		}
	}
}

func (t *TrafficGenerator) createOne(ctx context.Context) {
	n := t.seq.Add(1)

	flightID, seat, err := t.pool.Next(ctx)
	if err != nil {
		t.fail(ctx, "seat pool", err)
		return
	}

	bookingID, err := t.gw.CreateBooking(ctx, flightID, seat, n)
	if err != nil {
		t.fail(ctx, "create booking", err)
		return
	}
	t.created.Add(1)

	if rand.Float64() >= t.cfg.CancelProbability {
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(t.cfg.CancelDelay):
	}
	// Losing the race with payment (booking already PAID) is expected and not an error worth counting.
	if err := t.gw.CancelBooking(ctx, bookingID); err == nil {
		t.cancelled.Add(1)
	}
}

func (t *TrafficGenerator) fail(ctx context.Context, op string, err error) {
	if ctx.Err() != nil {
		return
	}
	// Log only the first few and then every 50th, so an outage of the gateway does not flood the output.
	if n := t.failed.Add(1); n <= 5 || n%50 == 0 {
		t.logger.Warn("traffic request failed", "op", op, "error", err, "failed_total", n)
	}
}

type TrafficStats struct {
	Created, Failed, Cancelled int64
}

func (t *TrafficGenerator) Stats() TrafficStats {
	return TrafficStats{Created: t.created.Load(), Failed: t.failed.Load(), Cancelled: t.cancelled.Load()}
}
