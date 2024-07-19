package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/1karp/ads_api/internal/app/config"
	"github.com/1karp/ads_api/internal/app/database"
	"github.com/1karp/ads_api/internal/app/handlers"
	"github.com/1karp/ads_api/internal/app/logging"
	"github.com/1karp/ads_api/internal/app/router"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		slog.Error("Error loading .env file", "error", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Setup logging
	logger := logging.SetupLogging(cfg)
	slog.SetDefault(logger)

	// Initialize database
	db := database.InitializeDatabase()
	defer db.Close()

	// Initialize handlers
	handlers.InitDB(db)

	// Setup router
	r := router.SetupRoutes()

	// Start server
	slog.Info("Server starting", "port", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		slog.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}
