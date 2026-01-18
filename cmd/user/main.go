package main

import (
	"context"
	"fmt"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/user/config"
	grpchandler "github.com/squ1ky/flyte/internal/user/handler/grpc"
	"github.com/squ1ky/flyte/internal/user/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/user/service"
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

const (
	migrationsPath = "migrations/user"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.SetupLogger(cfg.Env)
	log.Info("starting user service", slog.String("env", cfg.Env))

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

	userRepo := pgrepo.NewUserRepo(database)
	passRepo := pgrepo.NewPassengerRepo(database)

	authService := service.NewAuthService(userRepo, cfg.JWT)
	passService := service.NewPassengerService(passRepo)

	userHandler := grpchandler.NewServer(authService, passService)
	grpcServer := grpc.NewServer()
	userv1.RegisterUserServiceServer(grpcServer, userHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPC.Port))
	if err != nil {
		log.Error("failed to listen", slog.Any("error", err))
		os.Exit(1)
	}

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Info("grpc server started", slog.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("failed to serve", slog.Any("error", err))
			cancel()
		}
	}()

	shutdown.Graceful(log, cancel, grpcServer)
}
