package main

import (
	"fmt"
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

	userConn, err := grpc.NewClient(cfg.Clients.User.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("failed to connect to user service", slog.Any("error", err))
		os.Exit(1)
	}
	defer userConn.Close()

	userClient := userv1.NewUserServiceClient(userConn)
	log.Info("connected to user service", slog.String("addr", cfg.Clients.User.Addr))

	userHandler := handler.NewUserHandler(userClient)

	gatewayHandler := handler.NewGatewayHandler(userHandler)

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
