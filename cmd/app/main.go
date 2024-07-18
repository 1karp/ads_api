package main

import (
	"log"
	"net/http"
	"os"

	"github.com/1karp/ads_api/internal/app/database"
	"github.com/1karp/ads_api/internal/app/handlers"
	"github.com/1karp/ads_api/internal/app/router"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

func setupLogging() {
	logFile, err := os.OpenFile("ads_api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	log.SetOutput(logFile)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Setup logging
	setupLogging()

	// Initialize database
	db := database.InitializeDatabase()
	defer db.Close()

	// Initialize handlers
	handlers.InitDB(db)

	// Setup router
	r := router.SetupRoutes()

	// Start server
	log.Println("Server running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
