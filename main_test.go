package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func setupTestDB() *sql.DB {
	// Connect to the test database
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	// Create the users and ads tables
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		email TEXT NOT NULL
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
		house_name TEXT,
		district TEXT,
		text TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);
	`)
	if err != nil {
		panic(err)
	}

	return db
}

func TestCreateAd(t *testing.T) {
	// Setup the test database
	db = setupTestDB()

	// Create a new user in the database for the ad
	_, err := db.Exec("INSERT INTO users (username, email) VALUES (?, ?)", "ikarp", "ikarp@example.com")
	if err != nil {
		t.Fatalf("could not create test user: %v", err)
	}

	// Define the test ad data
	adData := map[string]interface{}{
		"user_id":    1,
		"username":   "ikarp",
		"photos":     "AgACAgIAAxkBAAIOYGaPofYmywRStaolZQL1ruSbKzLWAAKXyzEbNrZoSUAtq3dyXXWAAQADAgADeAADNQQ,AgACAgIAAxkBAAIOYWaPofapfPl66Y7pH-vKU8_A64ZlAAKYyzEbNrZoSbqZCMBWanj0AQADAgADeAADNQQ,AgACAgIAAxkBAAIOYmaPofYiqJxM9sMdMuj5_ubvKDsbAAKZyzEbNrZoSR9VaD1FQcJ0AQADAgADeAADNQQ,AgACAgIAAxkBAAIOY2aPofay_mJ-hl1WkO96iCgaUcaKAAKayzEbNrZoSVZenY4DrPbBAQADAgADeAADNQQ,AgACAgIAAxkBAAIOZGaPofZiq2K6LlQ5erANKYD5h4cnAAKbyzEbNrZoSZ1QwjfhHkjbAQADAgADeAADNQQ,AgACAgIAAxkBAAIOZWaPofamA1JqGo_Thwont_OYyko7AAKcyzEbNrZoSeu-OcZW8pWcAQADAgADeAADNQQ,AgACAgIAAxkBAAIOZmaPofbaDRttobjNsJfdnT3WzvwtAAKdyzEbNrZoSbQXMbPcmG5sAQADAgADeAADNQQ",
		"rooms":      "1",
		"price":      1,
		"type":       "1",
		"area":       1,
		"house_name": "1",
		"district":   "1",
		"text":       "1",
	}

	// Convert the ad data to JSON
	adDataJSON, err := json.Marshal(adData)
	if err != nil {
		t.Fatalf("could not marshal test ad data: %v", err)
	}

	// Create a new HTTP request with the test ad data
	req, err := http.NewRequest("POST", "/ads", bytes.NewBuffer(adDataJSON))
	if err != nil {
		t.Fatalf("could not create test request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder to capture the response
	rr := httptest.NewRecorder()

	// Create a new HTTP handler
	handler := http.HandlerFunc(createAd)

	// Serve the HTTP request
	handler.ServeHTTP(rr, req)

	// Print the response body for debugging
	fmt.Println(rr.Body.String())

	// Check the status code
	assert.Equal(t, http.StatusCreated, rr.Code)

	// Check the response body
	var createdAd Ad
	if err := json.NewDecoder(rr.Body).Decode(&createdAd); err != nil {
		t.Fatalf("could not decode response body: %v", err)
	}

	assert.Equal(t, adData["user_id"], createdAd.UserID)
	assert.Equal(t, adData["username"], createdAd.Username)
	assert.Equal(t, adData["photos"], createdAd.Photos)
	assert.Equal(t, adData["rooms"], createdAd.Rooms)
	assert.Equal(t, adData["price"], createdAd.Price)
	assert.Equal(t, adData["type"], createdAd.Type)
	assert.Equal(t, adData["area"], createdAd.Area)
	assert.Equal(t, adData["house_name"], createdAd.HouseName)
	assert.Equal(t, adData["district"], createdAd.District)
	assert.Equal(t, adData["text"], createdAd.Text)
}
