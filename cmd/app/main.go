package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/1karp/ads_api/internal/app/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func initDB() {
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
		userid INTEGER PRIMARY KEY,
		ads TEXT,
		username TEXT
	);

    CREATE TABLE IF NOT EXISTS ads (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id INTEGER NOT NULL,
        username TEXT,
        photos TEXT NOT NULL,
        rooms TEXT,
        price INTEGER,
        type TEXT,
        area INTEGER,
        building TEXT,
        district TEXT,
        text TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        is_posted INTEGER DEFAULT 0,
        chat_message_id INTEGER,
        FOREIGN KEY(user_id) REFERENCES users(userid)
    );
    `
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	logFile, err := os.OpenFile("ads_api.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	db, err = sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	initDB()

	handlers.InitDB(db)

	router := mux.NewRouter()

	router.HandleFunc("/ads", handlers.CreateAd).Methods("POST")
	router.HandleFunc("/ads", handlers.GetAds).Methods("GET")
	router.HandleFunc("/ads/{id}", handlers.GetAdByID).Methods("GET")
	router.HandleFunc("/ads/{id}", handlers.UpdateAd).Methods("PUT")
	router.HandleFunc("/ads/{id}/post", handlers.PostAd).Methods("POST")

	router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	router.HandleFunc("/users/{userid}", handlers.GetUserByID).Methods("GET")
	router.HandleFunc("/users/{userid}", handlers.UpdateUser).Methods("PUT")

	log.Println("Server running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
