package config

import (
	"os"
)

type Config struct {
	Environment string
	Port        string
	LogLevel    string
	// Add other configuration fields as needed
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port
	}

	return &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        port,
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}, nil
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
