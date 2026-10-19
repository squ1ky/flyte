package experiment

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type RunnerConfig struct {
	ExperimentID string
	Runs         int
	// Warmup runs once, before the first run: windows fill up and the batch poller seeds its state.
	Warmup time.Duration
	// Baseline is the no-fault period before every injection; any detection in it is a false positive.
	Baseline     time.Duration
	Timeout      time.Duration
	Cooldown     time.Duration
	PollInterval time.Duration
}

// Runner executes the protocol from CLAUDE.md 6.10 against the live system:
// baseline -> inject (t0) -> wait for every expected detector (t1) or timeout -> rollback -> cooldown.
// Stream and batch detectors observe the same traffic in the same run, so t1 - t0 is comparable.
type Runner struct {
	cfg       RunnerConfig
	scenario  Scenario
	analytics *AnalyticsClient
	injector  *ComposeInjector
	traffic   *TrafficGenerator
	results   *ResultWriter
	logger    *slog.Logger
}

func NewRunner(
	cfg RunnerConfig,
	scenario Scenario,
	analytics *AnalyticsClient,
	injector *ComposeInjector,
	traffic *TrafficGenerator,
	results *ResultWriter,
	logger *slog.Logger,
) *Runner {
	return &Runner{cfg: cfg, scenario: scenario, analytics: analytics, injector: injector,
		traffic: traffic, results: results, logger: logger}
}

func (r *Runner) Run(ctx context.Context) error {
	// A previous runner may have died mid-fault: always start from the base configuration.
	if err := r.injector.Rollback(ctx, r.scenario.Service); err != nil {
		return fmt.Errorf("initial rollback: %w", err)
	}
	// Never leave the system broken, including on Ctrl+C (ctx is already cancelled then).
	defer func() {
		rollbackCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := r.injector.Rollback(rollbackCtx, r.scenario.Service); err != nil {
			r.logger.Error("final rollback failed, restore the service manually", "service", r.scenario.Service, "error", err)
		}
	}()

	trafficCtx, stopTraffic := context.WithCancel(ctx)
	trafficDone := make(chan struct{})
	go func() {
		r.traffic.Run(trafficCtx)
		close(trafficDone)
	}()
	defer func() {
		stopTraffic()
		<-trafficDone
	}()

	r.logger.Info("warmup", "duration", r.cfg.Warmup)
	if err := sleep(ctx, r.cfg.Warmup); err != nil {
		return err
	}

	for run := 1; run <= r.cfg.Runs; run++ {
		if err := r.runOnce(ctx, run); err != nil {
			return fmt.Errorf("run %d: %w", run, err)
		}
	}
	return nil
}

func (r *Runner) runOnce(ctx context.Context, run int) error {
	log := r.logger.With("run", run, "scenario", r.scenario.Name)

	monitors, err := r.analytics.Monitors(ctx)
	if err != nil {
		return fmt.Errorf("list monitors: %w", err)
	}
	pending := map[Monitor]bool{}
	for _, m := range monitors {
		if r.scenario.Expects(m.MetricName) {
			pending[m] = true
		}
	}
	if len(pending) == 0 {
		return fmt.Errorf("analytics-service has no detectors for metrics %v", r.scenario.ExpectedMetrics)
	}

	// Baseline.
	if err := r.analytics.ResetDetectors(ctx); err != nil {
		return fmt.Errorf("reset detectors: %w", err)
	}
	skew, err := r.analytics.ClockSkew(ctx)
	if err != nil {
		return fmt.Errorf("measure clock skew: %w", err)
	}
	baselineStart := time.Now()
	log.Info("baseline", "duration", r.cfg.Baseline, "awaited_detectors", len(pending), "traffic", r.traffic.Stats())
	if err := sleep(ctx, r.cfg.Baseline); err != nil {
		return err
	}

	// Injection. Reset again so that a baseline false positive does not keep a detector latched.
	if err := r.analytics.ResetDetectors(ctx); err != nil {
		return fmt.Errorf("reset detectors: %w", err)
	}
	if skew, err = r.analytics.ClockSkew(ctx); err != nil {
		return fmt.Errorf("measure clock skew: %w", err)
	}
	log.Info("injecting fault", "service", r.scenario.Service, "env", r.scenario.FaultEnv)
	if err := r.injector.Inject(ctx, r.scenario.Service, r.scenario.FaultEnv); err != nil {
		return fmt.Errorf("inject: %w", err)
	}
	// t0 = the faulty container is up. Injection start would also count compose's recreate time.
	t0 := time.Now()
	log.Info("fault active", "t0", t0.Format(time.RFC3339Nano), "clock_skew", skew)

	row := func(phase, status string, m Monitor, start time.Time) ResultRow {
		return ResultRow{
			ExperimentID: r.cfg.ExperimentID, Scenario: r.scenario.Name, Run: run, Phase: phase, Status: status,
			Monitor: m, Expected: r.scenario.Expects(m.MetricName), T0: start, ClockSkew: skew,
		}
	}

	type detectionKey struct {
		Monitor
		detectedAt int64
	}
	written := map[detectionKey]bool{}
	deadline := t0.Add(r.cfg.Timeout)
	for {
		detections, err := r.analytics.DetectionsSince(ctx, baselineStart.Add(skew))
		if err != nil {
			return fmt.Errorf("read detections: %w", err)
		}

		for _, d := range detections {
			key := detectionKey{d.Monitor, d.DetectedAt.UnixNano()}
			if written[key] {
				continue
			}
			written[key] = true

			t1 := d.DetectedAt.Add(-skew)
			var res ResultRow
			if t1.Before(t0) {
				res = row(PhaseBaseline, StatusFalsePositive, d.Monitor, baselineStart)
				log.Warn("false positive before injection", "metric", d.MetricName, "approach", d.Approach, "algorithm", d.Algorithm)
			} else {
				res = row(PhaseFault, StatusDetected, d.Monitor, t0)
				delete(pending, d.Monitor)
				log.Info("detected", "metric", d.MetricName, "approach", d.Approach, "algorithm", d.Algorithm,
					"delta", t1.Sub(t0).Round(time.Millisecond), "remaining", len(pending))
			}
			res.T1, res.WindowStart, res.Score = t1, d.WindowStart, d.Score
			if err := r.results.Write(res); err != nil {
				return err
			}
		}

		if len(pending) == 0 || time.Now().After(deadline) {
			break
		}
		if err := sleep(ctx, r.cfg.PollInterval); err != nil {
			return err
		}
	}

	for m := range pending {
		log.Warn("not detected before timeout", "metric", m.MetricName, "approach", m.Approach, "algorithm", m.Algorithm)
		if err := r.results.Write(row(PhaseFault, StatusTimeout, m, t0)); err != nil {
			return err
		}
	}

	log.Info("rolling back fault")
	if err := r.injector.Rollback(ctx, r.scenario.Service); err != nil {
		return fmt.Errorf("rollback: %w", err)
	}
	log.Info("cooldown", "duration", r.cfg.Cooldown)
	return sleep(ctx, r.cfg.Cooldown)
}

func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
