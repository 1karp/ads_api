package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type Ad struct {
	ID            int    `json:"id"`
	UserID        int    `json:"user_id"`
	Username      string `json:"username"`
	Photos        string `json:"photos"`
	Rooms         string `json:"rooms"`
	Price         int    `json:"price"`
	Type          string `json:"type"`
	Area          int    `json:"area"`
	Building      string `json:"building"`
	District      string `json:"district"`
	Text          string `json:"text"`
	CreatedAt     string `json:"created_at"`
	IsPosted      int    `json:"is_posted"`
	ChatMessageId int    `json:"chat_message_id"`
}

var db *sql.DB

func main() {
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

	router := mux.NewRouter()

	router.HandleFunc("/ads", createAd).Methods("POST")
	router.HandleFunc("/ads", getAds).Methods("GET")
	router.HandleFunc("/ads/{id}", getAdByID).Methods("GET")
	router.HandleFunc("/ads/{id}", updateAd).Methods("PUT")

	log.Println("Server running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func createAd(w http.ResponseWriter, r *http.Request) {
	var ad Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := db.Exec(
		"INSERT INTO ads (user_id, username, photos, rooms, price, type, area, building, district, text, is_posted, chat_message_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		ad.UserID, ad.Username, ad.Photos, ad.Rooms, ad.Price, ad.Type, ad.Area, ad.Building, ad.District, ad.Text, ad.IsPosted, ad.ChatMessageId,
	)
	if err != nil {
		log.Printf("Error inserting ad into database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ad.ID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Ad created: %v", ad)
}

func getAds(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username != "" {
		getAdsByUsername(w, r)
		return
	}

	rows, err := db.Query("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads")
	if err != nil {
		log.Printf("Error querying ads from database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ads []Ad
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.CreatedAt, &ad.IsPosted, &ad.ChatMessageId); err != nil {
			log.Printf("Error scanning ad row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ads = append(ads, ad)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Ads retrieved: %v", ads)
}

func updateAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.Exec(
		"UPDATE ads SET user_id = ?, username = ?, photos = ?, rooms = ?, price = ?, type = ?, area = ?, building = ?, district = ?, text = ?, is_posted = ?, chat_message_id = ? WHERE id = ?",
		ad.UserID, ad.Username, ad.Photos, ad.Rooms, ad.Price, ad.Type, ad.Area, ad.Building, ad.District, ad.Text, ad.IsPosted, ad.ChatMessageId, id,
	)
	if err != nil {
		log.Printf("Error updating ad in database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Ad updated: %v", ad)
}

func getAdByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	err := db.QueryRow("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads WHERE id = ?", id).Scan(
		&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.CreatedAt, &ad.IsPosted, &ad.ChatMessageId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Ad not found", http.StatusNotFound)
		} else {
			log.Printf("Error querying ad by ID: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Ad retrieved by ID: %v", ad)
}

func getAdsByUsername(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	rows, err := db.Query("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads WHERE username = ?", username)
	if err != nil {
		log.Printf("Error querying ads by username from database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ads []Ad
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.CreatedAt, &ad.IsPosted, &ad.ChatMessageId); err != nil {
			log.Printf("Error scanning ad row: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ads = append(ads, ad)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Ads retrieved by username: %v", ads)
}
