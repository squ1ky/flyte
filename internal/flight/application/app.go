package application

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	pb "github.com/squ1ky/flyte/gen/proto/flight"
	"github.com/squ1ky/flyte/internal/flight/application/service"
	"github.com/squ1ky/flyte/internal/flight/application/worker"
	"github.com/squ1ky/flyte/internal/flight/config"
	"github.com/squ1ky/flyte/internal/flight/infrastructure/repository/elastic"
	"github.com/squ1ky/flyte/internal/flight/infrastructure/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/flight/infrastructure/transport/grpc"
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

const migrationsPath = "migrations/flight"

type Worker interface {
	Start(ctx context.Context)
}

type App struct {
	cfg        *config.Config
	logger     *slog.Logger
	db         *sqlx.DB
	grpcServer *grpc.Server
	workers    []Worker
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
	flightRepo := pgrepo.NewFlightRepo(database)
	airlineRepo := pgrepo.NewAirlineRepo(database)
	airportRepo := pgrepo.NewAirportRepo(database)
	aircraftRepo := pgrepo.NewAircraftRepo(database)
	reservationRepo := pgrepo.NewReservationRepo(database, cfg.App.ReservationTTL)
	outboxRepo := pgrepo.NewOutboxRepo(database)

	// Elasticsearch
	searchRepo, err := elastic.NewFlightSearchRepo(cfg.Elastic.URL)
	if err != nil {
		return nil, fmt.Errorf("elastic: %w", err)
	}

	// Service
	flightService := service.NewFlightService(flightRepo, searchRepo, logger)
	airlineService := service.NewAirlineService(airlineRepo, logger)
	airportService := service.NewAirportService(airportRepo, logger)
	aircraftService := service.NewAircraftService(aircraftRepo, logger)
	reservationService := service.NewReservationService(reservationRepo, logger)

	// gRPC
	grpcSrv := grpc.NewServer()
	flightServer := grpcserver.NewServer(
		aircraftService,
		airlineService,
		airportService,
		flightService,
		reservationService,
	)
	pb.RegisterFlightServiceServer(grpcSrv, flightServer)

	// Workers
	elasticSync := worker.NewElasticSync(
		outboxRepo,
		flightRepo,
		searchRepo,
		logger,
	)
	seatCleaner := worker.NewSeatCleaner(
		reservationRepo,
		logger,
		cfg.Workers.SeatLockCleanupInterval,
	)

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         database,
		grpcServer: grpcSrv,
		workers:    []Worker{elasticSync, seatCleaner},
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Workers
	var wg sync.WaitGroup
	for _, w := range a.workers {
		wg.Add(1)
		go func(w Worker) {
			defer wg.Done()
			w.Start(ctx)
		}(w)
	}

	// gRPC
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

	a.logger.InfoContext(ctx, "closing database connection")
	a.db.Close()
}
