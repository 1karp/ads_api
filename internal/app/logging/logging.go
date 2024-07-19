package logging

import (
	"log/slog"
	"os"

	"github.com/1karp/ads_api/internal/app/config"
)

func SetupLogging(cfg *config.Config) *slog.Logger {
	var logHandler slog.Handler

	if cfg.Environment == "production" {
		logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		logHandler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}

	return slog.New(logHandler)
}
