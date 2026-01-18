package shutdown

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Graceful(log *slog.Logger, cancel context.CancelFunc, servers ...interface{ GracefulStop() }) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("shutting down...", slog.String("signal", sign.String()))

	cancel()

	for _, s := range servers {
		s.GracefulStop()
	}

	log.Info("server stopped")
}
