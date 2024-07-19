package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func initTables(db *sql.DB) error {
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS users (
		userid INTEGER PRIMARY KEY,
		ads TEXT,
		username TEXT
	);

    CREATE TABLE IF NOT EXISTS ads (
        id SERIAL PRIMARY KEY,
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
        is_posted BOOLEAN DEFAULT FALSE,
        chat_message_id INTEGER,
        FOREIGN KEY(user_id) REFERENCES users(userid)
    );
    `
	_, err := db.Exec(sqlStmt)
	return err
}

func InitializeDatabase() *sql.DB {
	dbHost := os.Getenv("POSTGRES_HOST")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbName := os.Getenv("POSTGRES_DB")
	dbPort := os.Getenv("POSTGRES_PORT")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := initTables(db); err != nil {
		log.Fatalf("Failed to initialize database tables: %v", err)
	}

	return db
}
