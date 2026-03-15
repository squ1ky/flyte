package application

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	pb "github.com/squ1ky/flyte/gen/proto/user"
	"github.com/squ1ky/flyte/internal/user/application/service"
	"github.com/squ1ky/flyte/internal/user/config"
	"github.com/squ1ky/flyte/internal/user/infrastructure/repository/pgrepo"
	"github.com/squ1ky/flyte/internal/user/infrastructure/transport/grpc"
	"github.com/squ1ky/flyte/pkg/db"
	"google.golang.org/grpc"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const migrationsPath = "migrations/user"

type App struct {
	cfg        *config.Config
	logger     *slog.Logger
	db         *sqlx.DB
	grpcServer *grpc.Server
}

func New(cfg *config.Config, logger *slog.Logger) (*App, error) {
	dbCfg := db.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		Name:     cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	}

	if err := db.RunMigrations(dbCfg.DSN(), migrationsPath, logger); err != nil {
		return nil, fmt.Errorf("migrations: %w", err)
	}

	database, err := db.NewPostgresDB(dbCfg)
	if err != nil {
		return nil, fmt.Errorf("db connect: %w", err)
	}

	// Repository
	userRepo := pgrepo.NewUserRepo(database)
	passengerRepo := pgrepo.NewPassengerRepo(database)

	// Service
	authService := service.NewAuthService(userRepo, cfg.JWT, logger)
	passengerService := service.NewPassengerService(passengerRepo, logger)

	// gRPC
	grpcSrv := grpc.NewServer()
	userService := grpcserver.NewServer(authService, passengerService)
	pb.RegisterUserServiceServer(grpcSrv, userService)

	return &App{
		cfg:        cfg,
		logger:     logger,
		db:         database,
		grpcServer: grpcSrv,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// gRPC
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", a.cfg.GRPC.Port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	errCh := make(chan error, 1)
	go func() {
		a.logger.InfoContext(ctx, "starting grpc server",
			slog.Int("port", a.cfg.GRPC.Port),
		)
		if err := a.grpcServer.Serve(lis); err != nil {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		a.logger.InfoContext(ctx, "shutdown signal received")
	case err := <-errCh:
		return err
	}

	a.shutdown(ctx)
	return nil
}

func (a *App) shutdown(ctx context.Context) {
	a.logger.InfoContext(ctx, "stopping grpc server")
	a.grpcServer.GracefulStop()

	a.logger.InfoContext(ctx, "closing database connection")
	a.db.Close()
}
