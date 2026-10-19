package experiment

import (
	"fmt"
	"time"
)

// Scenario is one controlled fault: which compose service gets which env override,
// and which metrics are expected to detect it.
type Scenario struct {
	Name     string
	Service  string
	FaultEnv map[string]string
	// Detectors of these metrics (every algorithm, every approach) are awaited after injection.
	// Detectors of other metrics are still recorded if they fire, flagged as not expected.
	ExpectedMetrics []string
	// Defaults, overridable from the command line.
	Timeout  time.Duration
	Cooldown time.Duration
}

type SystemTiming struct {
	ReservationTTL  time.Duration
	CleanerInterval time.Duration
	OutboxInterval  time.Duration
}

const (
	ScenarioPaymentOutage = "payment_outage"
	ScenarioBankTimeout   = "bank_timeout"

	MetricPaymentSuccessRate     = "payment_success_rate"
	MetricConversionRate         = "conversion_rate"
	MetricAbandonmentRateExpired = "abandonment_rate_expired"
)

// BuildScenario returns the named scenario. bankDelay is used only by bank_timeout (0 = 2 x TTL).
func BuildScenario(name string, timing SystemTiming, bankDelay time.Duration) (Scenario, error) {
	switch name {
	case ScenarioPaymentOutage:
		// The bank declines everything: payment_success_rate drops immediately (payment_results),
		// conversion_rate drops as payment_failed cancellations arrive (booking_events).
		return Scenario{
			Name:            name,
			Service:         "payment-service",
			FaultEnv:        map[string]string{"FAKE_BANK_SUCCESS_CHANCE": "0"},
			ExpectedMetrics: []string{MetricPaymentSuccessRate, MetricConversionRate},
			Timeout:         10 * time.Minute,
			// Failed payments resolve within seconds; this mostly lets windows of the fault age out
			// of the detectors' view before the next baseline.
			Cooldown: 3 * time.Minute,
		}, nil

	case ScenarioBankTimeout:
		// The bank answers slower than the reservation lives: bookings stay PENDING until the
		// cleaner expires them, abandonment_rate_expired grows. Detection is inherently bounded
		// below by RESERVATION_TTL + cleaner tick, equally for stream and batch.
		if bankDelay == 0 {
			bankDelay = 2 * timing.ReservationTTL
		}
		if bankDelay <= timing.ReservationTTL {
			return Scenario{}, fmt.Errorf("bank delay %s must exceed RESERVATION_TTL %s", bankDelay, timing.ReservationTTL)
		}
		expiryLag := timing.ReservationTTL + timing.CleanerInterval + timing.OutboxInterval
		return Scenario{
			Name:    name,
			Service: "payment-service",
			FaultEnv: map[string]string{
				"FAKE_BANK_MIN_DELAY": bankDelay.String(),
				"FAKE_BANK_MAX_DELAY": bankDelay.String(),
			},
			ExpectedMetrics: []string{MetricAbandonmentRateExpired},
			Timeout:         expiryLag + 10*time.Minute,
			// Bookings created during the fault keep expiring for up to expiryLag after rollback,
			// plus the payment backlog accumulated during the fault has to drain.
			Cooldown: expiryLag + 5*time.Minute,
		}, nil

	default:
		return Scenario{}, fmt.Errorf("unknown scenario %q (known: %s, %s)", name, ScenarioPaymentOutage, ScenarioBankTimeout)
	}
}

func (s Scenario) Expects(metric string) bool {
	for _, m := range s.ExpectedMetrics {
		if m == metric {
			return true
		}
	}
	return false
}
