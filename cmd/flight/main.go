package main

import (
	"context"
	"fmt"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	"github.com/squ1ky/flyte/internal/flight/config"
	flightgrpc "github.com/squ1ky/flyte/internal/flight/handler/grpc"
	"github.com/squ1ky/flyte/internal/flight/repository/elastic"
	"github.com/squ1ky/flyte/internal/flight/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/flight/service"
	"github.com/squ1ky/flyte/internal/flight/service/worker"
	"github.com/squ1ky/flyte/pkg/bootstrap"
	"github.com/squ1ky/flyte/pkg/db"
	"github.com/squ1ky/flyte/pkg/logger"
	"github.com/squ1ky/flyte/pkg/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"os"
)

const migrationsPath = "migrations/flight"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.SetupLogger(cfg.Env)
	log.Info("starting flight service", slog.String("env", cfg.Env))

	dbCfg := db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	}
	database, dbClose, err := bootstrap.InitDB(dbCfg, migrationsPath, log)
	if err != nil {
		log.Error("init db failed", "error", err)
		os.Exit(1)
	}
	defer dbClose()

	esRepo, err := elastic.NewFlightSearchRepo(cfg.Elastic.URL)
	if err != nil {
		log.Error("failed to connect to elasticsearch", "error", err)
		os.Exit(1)
	}
	log.Info("connected to elasticsearch", slog.String("url", cfg.Elastic.URL))

	flightRepo := pgrepo.NewFlightRepo(database)
	aircraftRepo := pgrepo.NewAircraftRepo(database)

	flightService := service.NewFlightService(flightRepo, esRepo, log)
	aircraftService := service.NewAircraftService(aircraftRepo, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	esSyncWorker := worker.NewElasticSyncWorker(database, flightRepo, esRepo, log)
	seatCleaner := worker.NewSeatCleaner(database, log, cfg.Cleaner.Interval, cfg.Cleaner.ReservationTTL)
	go esSyncWorker.Start(ctx)
	go seatCleaner.Start(ctx)

	grpcServerImpl := flightgrpc.NewServer(flightService, aircraftService)

	grpcServer := grpc.NewServer()
	flightv1.RegisterFlightServiceServer(grpcServer, grpcServerImpl)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen", "port", cfg.GRPC.Port, "error", err)
		os.Exit(1)
	}

	go func() {
		log.Info("grpc server started", slog.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	shutdown.Graceful(log, cancel, grpcServer)
}
