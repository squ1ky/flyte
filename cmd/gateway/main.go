package main

import (
	"context"
	"fmt"
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/config"
	"github.com/squ1ky/flyte/internal/gateway/handler"
	"github.com/squ1ky/flyte/internal/gateway/router"
	"github.com/squ1ky/flyte/pkg/httpserver"
	"github.com/squ1ky/flyte/pkg/logger"
	"github.com/squ1ky/flyte/pkg/shutdown"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.SetupLogger(cfg.Env)
	log.Info("starting gateway service", slog.String("env", cfg.Env))

	// User Service
	userConn, err := grpc.NewClient(cfg.Clients.UserAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to user service", slog.Any("error", err))
		os.Exit(1)
	}

	userClient := userv1.NewUserServiceClient(userConn)
	log.Info("connected to user service", slog.String("addr", cfg.Clients.UserAddr))

	// Flight Service
	flightConn, err := grpc.NewClient(cfg.Clients.FlightAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to flight service", slog.Any("error", err))
		os.Exit(1)
	}

	flightClient := flightv1.NewFlightServiceClient(flightConn)
	log.Info("connected to flight service", slog.String("addr", cfg.Clients.FlightAddr))

	// Booking Service
	bookingConn, err := grpc.NewClient(cfg.Clients.BookingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to booking service", slog.Any("error", err))
		os.Exit(1)
	}

	bookingClient := bookingv1.NewBookingServiceClient(bookingConn)
	log.Info("connected to booking service", slog.String("addr", cfg.Clients.BookingAddr))

	defer func() {
		if err := bookingConn.Close(); err != nil {
			log.Error("error closing booking conn", "error", err)
		}
		if err := flightConn.Close(); err != nil {
			log.Error("error closing flight conn", "error", err)
		}
		if err := userConn.Close(); err != nil {
			log.Error("error closing user conn", "error", err)
		}
	}()

	// Handlers
	userHandler := handler.NewUserHandler(userClient)
	flightHandler := handler.NewFlightHandler(flightClient)
	bookingHandler := handler.NewBookingHandler(bookingClient)

	gatewayHandler := handler.NewGatewayHandler(userHandler, flightHandler, bookingHandler)

	r := router.InitRoutes(gatewayHandler, userClient)
	srv := httpserver.New(r, cfg.HTTP.Port)

	go func() {
		log.Info("gateway server listening", "port", cfg.HTTP.Port)

		if err := srv.Start(); err != nil {
			log.Error("failed to run server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdown.Graceful(log, cancel, srv)
}
