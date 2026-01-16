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
	"github.com/squ1ky/flyte/pkg/db"
	"github.com/squ1ky/flyte/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.Name, cfg.DB.SSLMode)

	if err := db.RunMigrations(dsn, migrationsPath, log); err != nil {
		log.Error("failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	database, err := db.NewPostgresDB(db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		log.Error("failed to connect to db", slog.Any("error", err))
		os.Exit(1)
	}
	defer database.Close()
	log.Info("db connection established")

	esRepo, err := elastic.NewFlightSearchRepo(cfg.Elastic.URL)
	if err != nil {
		log.Error("failed to connect to elasticsearch", "error", err)
		os.Exit(1)
	}
	log.Info("connected to elasticsearch", slog.String("url", cfg.Elastic.URL))

	flightRepo := pgrepo.NewFlightRepo(database)

	flightService := service.NewFlightService(flightRepo, esRepo, log)

	outboxProcessor := service.NewElasticOutboxProcessor(database, esRepo, log)
	seatCleaner := service.NewSeatCleaner(database, flightRepo, esRepo, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go outboxProcessor.Start(ctx)
	go seatCleaner.Start(ctx)

	grpcServerImpl := flightgrpc.NewServer(flightService)

	grpcServer := grpc.NewServer()
	flightv1.RegisterFlightServiceServer(grpcServer, grpcServerImpl)
	reflection.Register(grpcServer)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen", "port", cfg.GRPC.Port, "error", err)
		os.Exit(1)
	}

	// Graceful Shutdown
	go func() {
		log.Info("grpc server started", slog.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("shutting down...", slog.String("signal", sign.String()))

	cancel()

	grpcServer.GracefulStop()
	log.Info("server stopped")
}
