package application

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	pb "github.com/squ1ky/flyte/gen/proto/booking"
	"github.com/squ1ky/flyte/internal/booking/application/service"
	"github.com/squ1ky/flyte/internal/booking/application/worker"
	"github.com/squ1ky/flyte/internal/booking/config"
	flightclient "github.com/squ1ky/flyte/internal/booking/infrastructure/clients/grpc/flight"
	paymentmq "github.com/squ1ky/flyte/internal/booking/infrastructure/kafka/payment"
	"github.com/squ1ky/flyte/internal/booking/infrastructure/repository/pgrepo"
	grpcserver "github.com/squ1ky/flyte/internal/booking/infrastructure/transport/grpc"
	"github.com/squ1ky/flyte/pkg/db"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const migrationsPath = "migrations/booking"

type Worker interface {
	Run(ctx context.Context) error
}

type App struct {
	cfg             *config.Config
	logger          *slog.Logger
	db              *sqlx.DB
	grpcServer      *grpc.Server
	paymentConsumer *paymentmq.Consumer
	paymentProducer *paymentmq.Producer
	flights         *flightclient.Client
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
	bookingRepo := pgrepo.NewBookingRepo(database)
	ticketRepo := pgrepo.NewTicketRepo(database)
	outboxRepo := pgrepo.NewOutboxRepo(database)

	// Flight gRPC client
	flights, err := flightclient.New(cfg.FlightService.Address)
	if err != nil {
		return nil, fmt.Errorf("flight client: %w", err)
	}

	// Services
	saga := service.NewBookingSaga(bookingRepo, ticketRepo, outboxRepo, flights, cfg.Cleaner, logger)
	queryService := service.NewBookingQueryService(bookingRepo, ticketRepo, logger)
	cancelService := service.NewCancelService(bookingRepo, flights, logger)
	paymentHandler := service.NewPaymentHandler(bookingRepo, ticketRepo, flights, logger)

	// Kafka
	paymentProducer := paymentmq.NewProducer(cfg.Kafka, logger)
	paymentConsumer := paymentmq.NewConsumer(cfg.Kafka, paymentHandler, logger)

	// Workers
	outboxRelay := worker.NewOutboxRelay(outboxRepo, paymentProducer, cfg.Outbox, logger)
	expireWorker := worker.NewExpiredBookingCleaner(bookingRepo, flights, cfg.Cleaner, logger)

	// gRPC server
	srv := grpc.NewServer()
	handler := grpcserver.NewHandler(saga, queryService, cancelService, logger)
	pb.RegisterBookingServiceServer(srv, handler)

	return &App{
		cfg:             cfg,
		logger:          logger,
		db:              database,
		grpcServer:      srv,
		paymentConsumer: paymentConsumer,
		paymentProducer: paymentProducer,
		flights:         flights,
		workers:         []Worker{outboxRelay, expireWorker},
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.InfoContext(ctx, "starting grpc server",
			slog.Int("port", a.cfg.GRPC.Port),
		)
		if err := a.grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.InfoContext(ctx, "shutdown signal received")
	case err := <-errCh:
		return err
	}

	a.shutdown(ctx, &wg)
	return nil
}

func (a *App) shutdown(ctx context.Context, wg *sync.WaitGroup) {
	a.logger.InfoContext(ctx, "stopping grpc server")
	a.grpcServer.GracefulStop()

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
	if err := a.flights.Close(); err != nil {
		a.logger.ErrorContext(ctx, "close flight client",
			slog.Any("error", err),
		)
	}

	a.db.Close()
}
