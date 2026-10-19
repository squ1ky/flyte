// Command experiment runs the stream-vs-batch detection latency experiment
// against a running docker compose stack: it generates booking traffic through the gateway,
// injects a fault by recreating a service with overridden env, and records when every
// anomaly detector fires into a CSV file.
//
// Usage (from the repository root, with the whole stack up):
//
//	go run ./cmd/experiment -scenario payment_outage -runs 10
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/squ1ky/flyte/internal/experiment"
)

func main() {
	if err := run(); err != nil {
		slog.Error("experiment failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		envFile     = flag.String("env-file", ".env", "project .env (ports, credentials, RESERVATION_TTL)")
		composeFile = flag.String("compose-file", "docker-compose.yaml", "base docker compose file")
		scenarioArg = flag.String("scenario", experiment.ScenarioPaymentOutage, "payment_outage | bank_timeout")
		runs        = flag.Int("runs", 10, "number of inject/rollback cycles")
		warmup      = flag.Duration("warmup", 2*time.Minute, "traffic before the first run")
		baseline    = flag.Duration("baseline", 3*time.Minute, "no-fault period before every injection")
		timeout     = flag.Duration("timeout", 0, "max wait for detections after injection (0 = scenario default)")
		cooldown    = flag.Duration("cooldown", 0, "recovery after rollback (0 = scenario default)")
		bankDelay   = flag.Duration("bank-delay", 0, "bank_timeout: injected bank delay (0 = 2 x RESERVATION_TTL)")
		rate        = flag.Float64("rate", 0.5, "bookings per second")
		cancelProb  = flag.Float64("cancel-prob", 0.05, "share of bookings cancelled by the user")
		cancelDelay = flag.Duration("cancel-delay", time.Second, "delay before a user cancellation")
		poll        = flag.Duration("poll", time.Second, "how often to check analytics-db for detections")
		out         = flag.String("out", "", "results CSV (default experiment-results/<scenario>-<timestamp>.csv)")
	)
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	env, err := experiment.LoadEnvFile(*envFile)
	if err != nil {
		return err
	}

	var timing experiment.SystemTiming
	if timing.ReservationTTL, err = env.Duration("RESERVATION_TTL"); err != nil {
		return err
	}
	if timing.CleanerInterval, err = env.Duration("BOOKING_CLEANER_INTERVAL"); err != nil {
		return err
	}
	if timing.OutboxInterval, err = env.Duration("BOOKING_OUTBOX_INTERVAL"); err != nil {
		timing.OutboxInterval = 5 * time.Second // service default when not set in .env
	}

	scenario, err := experiment.BuildScenario(*scenarioArg, timing, *bankDelay)
	if err != nil {
		return err
	}
	if *timeout > 0 {
		scenario.Timeout = *timeout
	}
	if *cooldown > 0 {
		scenario.Cooldown = *cooldown
	}

	gatewayPort, err := env.Get("GATEWAY_HTTP_PORT")
	if err != nil {
		return err
	}
	analyticsPort, err := env.Get("ANALYTICS_HTTP_PORT")
	if err != nil {
		return err
	}
	userDB, err := env.PostgresURL("USER_DB")
	if err != nil {
		return err
	}
	analyticsDB, err := env.PostgresURL("ANALYTICS_DB")
	if err != nil {
		return err
	}

	experimentID := fmt.Sprintf("%s-%s", scenario.Name, time.Now().Format("20060102-150405"))
	if *out == "" {
		*out = filepath.Join("experiment-results", experimentID+".csv")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	gw := experiment.NewGatewayClient(fmt.Sprintf("http://localhost:%s/api/v1", gatewayPort))
	if err := experiment.PrepareAdmin(ctx, gw, userDB); err != nil {
		return err
	}
	pool, err := experiment.NewSeatPool(ctx, gw, logger)
	if err != nil {
		return err
	}

	analytics, err := experiment.NewAnalyticsClient(ctx, fmt.Sprintf("http://localhost:%s", analyticsPort), analyticsDB)
	if err != nil {
		return err
	}
	defer analytics.Close(context.Background())

	injector, err := experiment.NewComposeInjector(*composeFile)
	if err != nil {
		return err
	}
	defer injector.Close()

	results, err := experiment.NewResultWriter(*out)
	if err != nil {
		return err
	}
	defer results.Close()

	traffic := experiment.NewTrafficGenerator(experiment.TrafficConfig{
		Rate:              *rate,
		CancelProbability: *cancelProb,
		CancelDelay:       *cancelDelay,
	}, gw, pool, logger)

	logger.Info("starting experiment",
		"id", experimentID, "scenario", scenario.Name, "runs", *runs, "results", *out,
		"timeout", scenario.Timeout, "cooldown", scenario.Cooldown, "reservation_ttl", timing.ReservationTTL)

	runner := experiment.NewRunner(experiment.RunnerConfig{
		ExperimentID: experimentID,
		Runs:         *runs,
		Warmup:       *warmup,
		Baseline:     *baseline,
		Timeout:      scenario.Timeout,
		Cooldown:     scenario.Cooldown,
		PollInterval: *poll,
	}, scenario, analytics, injector, traffic, results, logger)

	if err := runner.Run(ctx); err != nil {
		return err
	}
	logger.Info("experiment finished", "results", *out, "traffic", traffic.Stats())
	return nil
}
