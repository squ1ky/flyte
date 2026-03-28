package main

import (
	"github.com/squ1ky/flyte/internal/booking/application"
	"github.com/squ1ky/flyte/internal/booking/config"
	"github.com/squ1ky/flyte/pkg/logger"
	"log/slog"
	"os"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	rootLogger := logger.SetupLogger(cfg.LogLevel)

	app, err := application.New(cfg, rootLogger)
	if err != nil {
		rootLogger.Error("failed to initialize app",
			slog.Any("error", err),
		)
		os.Exit(1)
	}

	if err := app.Run(); err != nil {
		rootLogger.Error("app exited with error",
			slog.Any("error", err),
		)
		os.Exit(1)
	}
}
