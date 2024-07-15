package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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

func InitDB(database *sql.DB) {
	db = database
}

func CreateAd(w http.ResponseWriter, r *http.Request) {
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

	// Update the user's ads field with the new ad_id
	_, err = db.Exec("UPDATE users SET ads = CASE WHEN ads IS NULL THEN ? ELSE ads || ',' || ? END WHERE userid = ?", ad.ID, ad.ID, ad.UserID)
	if err != nil {
		log.Printf("Error updating user's ads field: %v", err)
		// Note: We're not returning here as the ad was successfully created
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Ad created: %v", ad)
}

func GetAds(w http.ResponseWriter, r *http.Request) {
	userid := r.URL.Query().Get("userid")
	if userid != "" {
		GetAdsByUserId(w, r)
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

func GetAdByID(w http.ResponseWriter, r *http.Request) {
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

func GetAdsByUserId(w http.ResponseWriter, r *http.Request) {
	userid := r.URL.Query().Get("userid")
	if userid == "" {
		http.Error(w, "Userid is required", http.StatusBadRequest)
		return
	}

	// First, get the string of ad IDs from the users table
	var adIDs string
	err := db.QueryRow("SELECT ads FROM users WHERE userid = ?", userid).Scan(&adIDs)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Error querying user ads: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Split the adIDs string into a slice
	adIDSlice := strings.Split(adIDs, ",")

	// Prepare the query to get ads by their IDs
	query := "SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads WHERE id IN (?" + strings.Repeat(",?", len(adIDSlice)-1) + ")"

	// Convert adIDSlice to []interface{} for the query
	args := make([]interface{}, len(adIDSlice))
	for i, id := range adIDSlice {
		args[i] = id
	}

	// Execute the query
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error querying ads by IDs: %v", err)
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

func UpdateAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the ad exists before updating
	var existingAd Ad
	err := db.QueryRow("SELECT id FROM ads WHERE id = ?", id).Scan(&existingAd.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "Ad not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking existing ad: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(
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

func PostAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	err := db.QueryRow("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, is_posted, chat_message_id FROM ads WHERE id = ?", id).Scan(
		&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.IsPosted, &ad.ChatMessageId)
	if err != nil {
		log.Printf("Error fetching ad details: %v", err)
		http.Error(w, "Error fetching ad details", http.StatusInternalServerError)
		return
	}

	if ad.IsPosted == 1 {
		http.Error(w, "Ad already posted", http.StatusBadRequest)
		return
	}

	err = postToTelegramChannel(ad)
	if err != nil {
		log.Printf("Error posting to Telegram: %v", err)
		http.Error(w, "Error posting to Telegram", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE ads SET is_posted = 1 WHERE id = ?", id)
	if err != nil {
		log.Printf("Error updating ad status: %v", err)
		http.Error(w, "Error updating ad status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ad successfully posted to Telegram channel")
}

func calculatePriceHash(price int) int {
	return ((price-1)/10000 + 1) * 10000
}

func generateAdText(ad Ad, districtHash, priceHash string) string {
	return fmt.Sprintf(
		"#%s, #under_%s\n\n"+
			"Rooms: %s\n"+
			"Price: %d AED/Year\n"+
			"Type: %s\n"+
			"Area: %d sqm\n"+
			"Building: %s\n"+
			"District: %s\n\n"+
			"%s\n\n"+
			"Contact: @%s",
		districtHash, priceHash,
		ad.Rooms, ad.Price, ad.Type, ad.Area,
		ad.Building, ad.District, ad.Text, ad.Username)
}

func postToTelegramChannel(ad Ad) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	channelID := os.Getenv("TELEGRAM_CHANNEL_ID")
	if botToken == "" || channelID == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN or TELEGRAM_CHANNEL_ID not set")
	}

	districtHash := strings.ReplaceAll(ad.District, " ", "_")
	priceHash := fmt.Sprintf("%d", calculatePriceHash(ad.Price))
	text := generateAdText(ad, districtHash, priceHash)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMediaGroup", botToken)
	photos := strings.Split(ad.Photos, ",")
	media := make([]map[string]string, len(photos))

	for i, photo := range photos {
		mediaItem := map[string]string{
			"type":  "photo",
			"media": photo,
		}
		if i == 0 {
			mediaItem["caption"] = text
			mediaItem["parse_mode"] = "HTML"
		}
		media[i] = mediaItem
	}

	payload := map[string]interface{}{
		"chat_id": channelID,
		"media":   media,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ok     bool `json:"ok"`
		Result []struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	if len(result.Result) > 0 {
		_, err = db.Exec("UPDATE ads SET chat_message_id = ? WHERE id = ?", result.Result[0].MessageID, ad.ID)
		if err != nil {
			return fmt.Errorf("error updating chat_message_id: %v", err)
		}
	}

	return nil
}
