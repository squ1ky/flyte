package main

import (
	"fmt"
	bookingv1 "github.com/squ1ky/flyte/gen/go/booking"
	flightv1 "github.com/squ1ky/flyte/gen/go/flight"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/gateway/config"
	"github.com/squ1ky/flyte/internal/gateway/handler"
	"github.com/squ1ky/flyte/internal/gateway/router"
	"github.com/squ1ky/flyte/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
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
	defer userConn.Close()
	userClient := userv1.NewUserServiceClient(userConn)
	log.Info("connected to user service", slog.String("addr", cfg.Clients.UserAddr))

	// Flight Service
	flightConn, err := grpc.NewClient(cfg.Clients.FlightAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to flight service", slog.Any("error", err))
		os.Exit(1)
	}
	defer flightConn.Close()
	flightClient := flightv1.NewFlightServiceClient(flightConn)
	log.Info("connected to flight service", slog.String("addr", cfg.Clients.FlightAddr))

	// Booking Service
	bookingConn, err := grpc.NewClient(cfg.Clients.BookingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to booking service", slog.Any("error", err))
		os.Exit(1)
	}
	defer bookingConn.Close()
	bookingClient := bookingv1.NewBookingServiceClient(bookingConn)
	log.Info("connected to booking service", slog.String("addr", cfg.Clients.BookingAddr))

	// Handlers
	userHandler := handler.NewUserHandler(userClient)
	flightHandler := handler.NewFlightHandler(flightClient)
	bookingHandler := handler.NewBookingHandler(bookingClient)

	gatewayHandler := handler.NewGatewayHandler(userHandler, flightHandler, bookingHandler)

	r := router.NewRouter(gatewayHandler, userClient)

	go func() {
		addr := fmt.Sprintf(":%d", cfg.HTTP.Port)
		log.Info("gateway server listening", slog.String("addr", addr))

		if err := r.Run(addr); err != nil {
			log.Error("failed to run server", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("shutting down...", slog.String("signal", sign.String()))
	log.Info("gateway stopped")
}
