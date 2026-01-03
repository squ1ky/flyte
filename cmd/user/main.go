package main

import (
	"fmt"
	userv1 "github.com/squ1ky/flyte/gen/go/user"
	"github.com/squ1ky/flyte/internal/user/config"
	grpchandler "github.com/squ1ky/flyte/internal/user/handler/grpc"
	"github.com/squ1ky/flyte/internal/user/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/user/service"
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

	// Graceful Shutdown
	go func() {
		log.Info("grpc server started", slog.Int("port", cfg.GRPC.Port))
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("failed to serve", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("shutting down...", slog.String("signal", sign.String()))

	grpcServer.GracefulStop()
	log.Info("server stopped")
}
