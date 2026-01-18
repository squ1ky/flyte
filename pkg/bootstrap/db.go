package bootstrap

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/squ1ky/flyte/pkg/db"
	"log/slog"
)

func InitDB(cfg db.Config, migrationsPath string, log *slog.Logger) (*sqlx.DB, func(), error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode)

	if err := db.RunMigrations(dsn, migrationsPath, log); err != nil {
		return nil, nil, fmt.Errorf("migrations failed: %w", err)
	}

	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("db connection failed: %w", err)
	}

	cleanup := func() {
		log.Info("closing database connection")
		database.Close()
	}

	return database, cleanup, nil
}
