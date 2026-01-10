package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/squ1ky/flyte/internal/payment/config"
	"github.com/squ1ky/flyte/internal/payment/kafka"
	"github.com/squ1ky/flyte/internal/payment/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/payment/service"
	"github.com/squ1ky/flyte/pkg/db"
	"github.com/squ1ky/flyte/pkg/logger"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const migrationsPath = "migrations/payment"

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.SetupLogger(cfg.Env)
	log.Info("starting payment service", slog.String("env", cfg.Env))

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
	defer func() {
		log.Info("closing database connection")
		if err := database.Close(); err != nil {
			log.Error("failed to close database connection", slog.Any("error", err))
		}
	}()
	log.Info("db connection established")

	repo := pgrepo.NewPaymentRepo(database)
	paymentService := service.NewPaymentService(repo, log)

	producer := kafka.NewPaymentProducer(cfg.Kafka, log)
	handler := kafka.NewPaymentMessageHandler(paymentService, producer, log)
	consumer := kafka.NewPaymentConsumer(cfg.Kafka, handler, log)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Info("kafka consumer started",
			slog.String("topic", cfg.Kafka.TopicRequests),
			slog.String("group_id", cfg.Kafka.GroupID),
		)

		if err := consumer.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Error("consumer stopped with error", "error", err)
			}
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("shutting down...", slog.String("signal", sign.String()))

	cancel()

	if err := consumer.Close(); err != nil {
		log.Error("failed to close consumer", "error", err)
	}

	if err := producer.Close(); err != nil {
		log.Error("failed to close producer", "error", err)
	}

	log.Info("server stopped")
}
