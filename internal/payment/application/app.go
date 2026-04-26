package application

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/internal/payment/application/service"
	"github.com/squ1ky/flyte/internal/payment/config"
	"github.com/squ1ky/flyte/internal/payment/infrastructure/bank"
	paymentmq "github.com/squ1ky/flyte/internal/payment/infrastructure/kafka"
	"github.com/squ1ky/flyte/internal/payment/infrastructure/repository/pgrepo"
	"github.com/squ1ky/flyte/pkg/db"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const migrationsPath = "migrations/payment"

type Worker interface {
	Run(ctx context.Context) error
}

type App struct {
	cfg             *config.Config
	logger          *slog.Logger
	db              *sqlx.DB
	paymentConsumer *paymentmq.Consumer
	paymentProducer *paymentmq.Producer
	workers         []Worker
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	dbCfg := db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	}

	if err := db.RunMigrations(dbCfg.DSN(), migrationsPath, logger); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	database, err := db.NewPostgresDB(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	// Repository
	paymentRepo := pgrepo.NewPaymentRepo(database)

	// Bank gateway (fake)
	fakeBank := bank.NewFakeBank(cfg.FakeBank, logger)

	// Services
	paymentService := service.NewPaymentService(paymentRepo, fakeBank, logger)

	// Kafka
	paymentProducer := paymentmq.NewProducer(cfg.Kafka, logger)
	paymentConsumer := paymentmq.NewConsumer(cfg.Kafka, paymentService, paymentProducer, logger)

	return &App{
		cfg:             cfg,
		logger:          logger,
		db:              database,
		paymentConsumer: paymentConsumer,
		paymentProducer: paymentProducer,
		workers:         []Worker{},
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	workers := append(a.workers, a.paymentConsumer)

	for _, w := range workers {
		wg.Add(1)
		go func(w Worker) {
			defer wg.Done()
			if err := w.Run(ctx); err != nil && ctx.Err() == nil {
				a.logger.ErrorContext(ctx, "worker failed",
					slog.Any("error", err),
				)
			}
		}(w)
	}

	<-ctx.Done()
	a.logger.InfoContext(ctx, "shutdown signal received")

	a.shutdown(ctx, &wg)
	return nil
}

func (a *App) shutdown(ctx context.Context, wg *sync.WaitGroup) {
	a.logger.InfoContext(ctx, "waiting for workers to finish")
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.InfoContext(ctx, "all workers stopped")
	case <-time.After(10 * time.Second):
		a.logger.WarnContext(ctx, "workers shutdown timed out")
	}

	a.logger.InfoContext(ctx, "closing resources")

	if err := a.paymentConsumer.Close(); err != nil {
		a.logger.ErrorContext(ctx, "close payment consumer",
			slog.Any("error", err),
		)
	}
	if err := a.paymentProducer.Close(); err != nil {
		a.logger.ErrorContext(ctx, "close payment producer",
			slog.Any("error", err),
		)
	}

	if err := a.db.Close(); err != nil {
		a.logger.ErrorContext(ctx, "close db",
			slog.Any("error", err),
		)
	}
}
