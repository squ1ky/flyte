package main

import (
	"context"
	"fmt"
	"github.com/squ1ky/flyte/internal/booking/clients/grpc/flight"
	"github.com/squ1ky/flyte/internal/booking/config"
	bookinggrpc "github.com/squ1ky/flyte/internal/booking/handler/grpc"
	"github.com/squ1ky/flyte/internal/booking/kafka"
	"github.com/squ1ky/flyte/internal/booking/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/booking/service"
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

const migrationsPath = "migrations/booking"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.SetupLogger(cfg.Env)
	log.Info("starting booking service", slog.String("env", cfg.Env))

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

	flightClient, err := flight.NewClient(cfg.FlightService.Address, cfg.GRPC.Timeout)
	if err != nil {
		log.Error("failed to create flight service client", "error", err)
	}

	producer := kafka.NewProducer(cfg.Kafka, log)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error("failed to close kafka producer", "error", err)
		}
	}()

	bookingRepo := pgrepo.NewBookingRepo(database)
	bookingService := service.NewBookingService(bookingRepo, producer, flightClient, log)
	paymentProcessor := service.NewPaymentProcessor(bookingRepo, flightClient, log)

	kafkaHandler := kafka.NewBookingMessageHandler(paymentProcessor, log)
	consumer := kafka.NewBookingConsumer(cfg.Kafka, kafkaHandler, log)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error("failed to close kafka consumer", "error", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Info("kafka consumer started")
		if err := consumer.Start(ctx); err != nil {
			log.Error("kafka consumer stopped with error", "error", err)
		}
	}()

	grpcServerImpl := bookinggrpc.NewServer(bookingService, cfg.GRPC.Timeout)
	grpcServer := grpc.NewServer()
	grpcServerImpl.Register(grpcServer)
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
