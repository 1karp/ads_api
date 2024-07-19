package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
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
		slog.Error("Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := db.QueryRow(
		"INSERT INTO ads (user_id, username, photos, rooms, price, type, area, building, district, text, is_posted, chat_message_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id",
		ad.UserID, ad.Username, ad.Photos, ad.Rooms, ad.Price, ad.Type, ad.Area, ad.Building, ad.District, ad.Text, ad.IsPosted, ad.ChatMessageId,
	).Scan(&ad.ID)
	if err != nil {
		slog.Error("Error inserting ad into database", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE users SET ads = COALESCE(ads, '') || CASE WHEN ads = '' THEN $1::text ELSE ',' || $1::text END WHERE userid = $2", ad.ID, ad.UserID)
	if err != nil {
		slog.Error("Error updating user's ads field", "error", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
	slog.Info("Ad created successfully", "ad_id", ad.ID)
}

func GetAds(w http.ResponseWriter, r *http.Request) {
	userid := r.URL.Query().Get("userid")
	if userid != "" {
		GetAdsByUserId(w, r)
		return
	}

	rows, err := db.Query("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads")
	if err != nil {
		slog.Error("Error querying ads from database", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ads []Ad
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.CreatedAt, &ad.IsPosted, &ad.ChatMessageId); err != nil {
			slog.Error("Error scanning ad row", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ads = append(ads, ad)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
	slog.Info("Ads retrieved successfully", "count", len(ads))
}

func GetAdByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	err := db.QueryRow("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads WHERE id = $1", id).Scan(
		&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.CreatedAt, &ad.IsPosted, &ad.ChatMessageId,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Ad not found", http.StatusNotFound)
		} else {
			slog.Error("Error querying ad by ID", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
	slog.Info("Ad retrieved successfully", "ad_id", ad.ID)
}

func GetAdsByUserId(w http.ResponseWriter, r *http.Request) {
	userid := r.URL.Query().Get("userid")
	if userid == "" {
		http.Error(w, "Userid is required", http.StatusBadRequest)
		return
	}

	var adIDs string
	err := db.QueryRow("SELECT ads FROM users WHERE userid = $1", userid).Scan(&adIDs)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			slog.Error("Error querying user ads", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	adIDSlice := strings.Split(adIDs, ",")

	query := fmt.Sprintf("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, created_at, is_posted, chat_message_id FROM ads WHERE id = ANY($1::int[])")

	rows, err := db.Query(query, pq.Array(adIDSlice))
	if err != nil {
		slog.Error("Error querying ads by IDs", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ads []Ad
	for rows.Next() {
		var ad Ad
		if err := rows.Scan(&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.CreatedAt, &ad.IsPosted, &ad.ChatMessageId); err != nil {
			slog.Error("Error scanning ad row", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ads = append(ads, ad)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ads); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
	slog.Info("Ads retrieved successfully by user ID", "user_id", userid, "count", len(ads))
}

func UpdateAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	if err := json.NewDecoder(r.Body).Decode(&ad); err != nil {
		slog.Error("Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingAd Ad
	err := db.QueryRow("SELECT id FROM ads WHERE id = $1", id).Scan(&existingAd.ID)
	if err == sql.ErrNoRows {
		http.Error(w, "Ad not found", http.StatusNotFound)
		return
	} else if err != nil {
		slog.Error("Error checking existing ad", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = db.Exec(
		"UPDATE ads SET user_id = $1, username = $2, photos = $3, rooms = $4, price = $5, type = $6, area = $7, building = $8, district = $9, text = $10, is_posted = $11, chat_message_id = $12 WHERE id = $13",
		ad.UserID, ad.Username, ad.Photos, ad.Rooms, ad.Price, ad.Type, ad.Area, ad.Building, ad.District, ad.Text, ad.IsPosted, ad.ChatMessageId, id,
	)
	if err != nil {
		slog.Error("Error updating ad in database", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ad); err != nil {
		slog.Error("Error encoding response", "error", err)
	}
	slog.Info("Ad updated successfully", "ad_id", id)
}

func PostAd(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	err := db.QueryRow("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, is_posted, chat_message_id FROM ads WHERE id = $1", id).Scan(
		&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.IsPosted, &ad.ChatMessageId)
	if err != nil {
		slog.Error("Error fetching ad details", "error", err)
		http.Error(w, "Error fetching ad details", http.StatusInternalServerError)
		return
	}

	if ad.IsPosted == 1 {
		http.Error(w, "Ad already posted", http.StatusBadRequest)
		return
	}

	err = postToTelegramChannel(ad)
	if err != nil {
		slog.Error("Error posting to Telegram", "error", err)
		http.Error(w, "Error posting to Telegram", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("UPDATE ads SET is_posted = 1 WHERE id = $1", id)
	if err != nil {
		slog.Error("Error updating ad status", "error", err)
		http.Error(w, "Error updating ad status", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ad successfully posted to Telegram channel")
	slog.Info("Ad posted successfully to Telegram", "ad_id", id)
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
		_, err = db.Exec("UPDATE ads SET chat_message_id = $1 WHERE id = $2", result.Result[0].MessageID, ad.ID)
		if err != nil {
			return fmt.Errorf("error updating chat_message_id: %v", err)
		}
	}

	slog.Info("Ad successfully posted to Telegram", "ad_id", ad.ID)
	return nil
}

func EditAdInTelegram(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var ad Ad
	err := db.QueryRow("SELECT id, user_id, username, photos, rooms, price, type, area, building, district, text, is_posted, chat_message_id FROM ads WHERE id = $1", id).Scan(
		&ad.ID, &ad.UserID, &ad.Username, &ad.Photos, &ad.Rooms, &ad.Price, &ad.Type, &ad.Area, &ad.Building, &ad.District, &ad.Text, &ad.IsPosted, &ad.ChatMessageId)
	if err != nil {
		slog.Error("Error fetching ad details", "error", err)
		http.Error(w, "Error fetching ad details", http.StatusInternalServerError)
		return
	}

	if ad.IsPosted != 1 || ad.ChatMessageId == 0 {
		http.Error(w, "Ad not posted or message ID not available", http.StatusBadRequest)
		return
	}

	err = editTelegramMessage(ad)
	if err != nil {
		slog.Error("Error editing Telegram message", "error", err)
		http.Error(w, "Error editing Telegram message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Ad successfully edited in Telegram channel")
	slog.Info("Ad successfully edited in Telegram channel", "ad_id", id)
}

func editTelegramMessage(ad Ad) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	channelID := os.Getenv("TELEGRAM_CHANNEL_ID")
	if botToken == "" || channelID == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN or TELEGRAM_CHANNEL_ID not set")
	}

	districtHash := strings.ReplaceAll(ad.District, " ", "_")
	priceHash := fmt.Sprintf("%d", calculatePriceHash(ad.Price))
	newText := generateAdText(ad, districtHash, priceHash)

	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageCaption", botToken)

	params := map[string]interface{}{
		"chat_id":    channelID,
		"message_id": ad.ChatMessageId,
		"caption":    newText,
		"parse_mode": "HTML",
	}

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return fmt.Errorf("error marshaling params: %v", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonParams))
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Ok          bool   `json:"ok"`
		ErrorCode   int    `json:"error_code,omitempty"`
		Description string `json:"description,omitempty"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !result.Ok {
		return fmt.Errorf("failed to edit message: %s (code: %d)", result.Description, result.ErrorCode)
	}

	slog.Info("Telegram message successfully edited", "ad_id", ad.ID)
	return nil
}
