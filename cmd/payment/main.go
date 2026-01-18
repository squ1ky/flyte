package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/squ1ky/flyte/internal/payment/config"
	"github.com/squ1ky/flyte/internal/payment/kafka"
	"github.com/squ1ky/flyte/internal/payment/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/payment/service"
	"github.com/squ1ky/flyte/pkg/bootstrap"
	"github.com/squ1ky/flyte/pkg/db"
	"github.com/squ1ky/flyte/pkg/logger"
	"github.com/squ1ky/flyte/pkg/shutdown"
	"log/slog"
	"os"
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

	producer := kafka.NewPaymentProducer(cfg.Kafka, log)
	defer func() {
		if err := producer.Close(); err != nil {
			log.Error("failed to close producer", "error", err)
		}
	}()

	repo := pgrepo.NewPaymentRepo(database)
	paymentService := service.NewPaymentService(repo, log)

	handler := kafka.NewPaymentMessageHandler(paymentService, producer, log)
	consumer := kafka.NewPaymentConsumer(cfg.Kafka, handler, log)
	defer func() {
		if err := consumer.Close(); err != nil {
			log.Error("failed to close consumer", "error", err)
		}
	}()

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

	shutdown.Graceful(log, cancel)
}
