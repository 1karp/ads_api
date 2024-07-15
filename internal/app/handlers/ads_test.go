package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCreateAd(t *testing.T) {
	// Create a new mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	// Set our mock database as the handler's database
	db = mockDB

	// Create a new ad
	ad := Ad{
		UserID:   1,
		Username: "testuser",
		Photos:   "photo1.jpg,photo2.jpg",
		Rooms:    "2",
		Price:    1000,
		Type:     "apartment",
		Area:     50,
		Building: "modern",
		District: "downtown",
		Text:     "Nice apartment",
		IsPosted: 1,
	}

	// Expect the insert query
	mock.ExpectExec("INSERT INTO ads").WithArgs(
		ad.UserID, ad.Username, ad.Photos, ad.Rooms, ad.Price, ad.Type, ad.Area,
		ad.Building, ad.District, ad.Text, ad.IsPosted, ad.ChatMessageId,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	// Create a request body
	body, _ := json.Marshal(ad)
	req, err := http.NewRequest("POST", "/ads", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(CreateAd)
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	// Check the response body
	var createdAd Ad
	err = json.Unmarshal(rr.Body.Bytes(), &createdAd)
	if err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if createdAd.ID != 1 {
		t.Errorf("Expected created ad ID to be 1, got %d", createdAd.ID)
	}

	// We make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Test case for invalid JSON
	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/ads", strings.NewReader(`{"invalid json"`))
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateAd)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	// Test case for database error
	t.Run("Database Error", func(t *testing.T) {
		// Mock the database to return an error
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		mock.ExpectExec("INSERT INTO ads").WillReturnError(fmt.Errorf("database error"))

		ad := Ad{UserID: 1, Username: "testuser", Price: 1000}
		body, _ := json.Marshal(ad)
		req, _ := http.NewRequest("POST", "/ads", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(CreateAd)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestGetAdByID(t *testing.T) {
	// Create a new mock database
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	// Set our mock database as the handler's database
	db = mockDB

	// Create a sample ad
	ad := Ad{
		ID:       1,
		UserID:   1,
		Username: "testuser",
		Photos:   "photo1.jpg,photo2.jpg",
		Rooms:    "2",
		Price:    1000,
		Type:     "apartment",
		Area:     50,
		Building: "modern",
		District: "downtown",
		Text:     "Nice apartment",
		IsPosted: 1,
	}

	// Expect the select query
	rows := sqlmock.NewRows([]string{"id", "user_id", "username", "photos", "rooms", "price", "type", "area", "building", "district", "text", "created_at", "is_posted", "chat_message_id"}).
		AddRow(ad.ID, ad.UserID, ad.Username, ad.Photos, ad.Rooms, ad.Price, ad.Type, ad.Area, ad.Building, ad.District, ad.Text, "2023-05-01", ad.IsPosted, ad.ChatMessageId)

	mock.ExpectQuery("SELECT (.+) FROM ads WHERE id = ?").WithArgs("1").WillReturnRows(rows)

	// Create a request
	req, err := http.NewRequest("GET", "/ads/1", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set up the mux router
	router := mux.NewRouter()
	router.HandleFunc("/ads/{id}", GetAdByID)

	// Create a response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	router.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Check the response body
	var retrievedAd Ad
	err = json.Unmarshal(rr.Body.Bytes(), &retrievedAd)
	if err != nil {
		t.Errorf("Could not unmarshal response: %v", err)
	}

	if retrievedAd.ID != ad.ID {
		t.Errorf("Expected ad ID to be %d, got %d", ad.ID, retrievedAd.ID)
	}

	// We make sure that all expectations were met
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Test case for non-existent ad
	t.Run("Non-existent Ad", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		mock.ExpectQuery("SELECT (.+) FROM ads WHERE id = ?").WithArgs("999").WillReturnError(sql.ErrNoRows)

		req, _ := http.NewRequest("GET", "/ads/999", nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/ads/{id}", GetAdByID)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	// Test case for database error
	t.Run("Database Error", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		mock.ExpectQuery("SELECT (.+) FROM ads WHERE id = ?").WithArgs("1").WillReturnError(fmt.Errorf("database error"))

		req, _ := http.NewRequest("GET", "/ads/1", nil)
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/ads/{id}", GetAdByID)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}

func TestGetAds(t *testing.T) {
	// ... existing test ...

	// Test case for database error
	t.Run("Database Error", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		mock.ExpectQuery("SELECT (.+) FROM ads").WillReturnError(fmt.Errorf("database error"))

		req, _ := http.NewRequest("GET", "/ads", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetAds)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	// Test case for empty result
	t.Run("Empty Result", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		rows := sqlmock.NewRows([]string{"id", "user_id", "username", "photos", "rooms", "price", "type", "area", "building", "district", "text", "created_at", "is_posted", "chat_message_id"})
		mock.ExpectQuery("SELECT (.+) FROM ads").WillReturnRows(rows)

		req, _ := http.NewRequest("GET", "/ads", nil)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(GetAds)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "null", strings.TrimSpace(rr.Body.String()))
	})
}

func TestUpdateAd(t *testing.T) {
	// ... existing test ...

	// Test case for non-existent ad
	t.Run("Non-existent Ad", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		mock.ExpectQuery("SELECT id FROM ads WHERE id = ?").WithArgs("999").WillReturnError(sql.ErrNoRows)

		ad := Ad{UserID: 1, Username: "testuser", Price: 1000}
		body, _ := json.Marshal(ad)
		req, _ := http.NewRequest("PUT", "/ads/999", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/ads/{id}", UpdateAd)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	// Test case for database error during update
	t.Run("Database Error During Update", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		defer db.Close()
		InitDB(db)

		mock.ExpectQuery("SELECT id FROM ads WHERE id = ?").WithArgs("1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec("UPDATE ads SET").WillReturnError(fmt.Errorf("database error"))

		ad := Ad{UserID: 1, Username: "testuser", Price: 1000}
		body, _ := json.Marshal(ad)
		req, _ := http.NewRequest("PUT", "/ads/1", bytes.NewBuffer(body))
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/ads/{id}", UpdateAd)
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
