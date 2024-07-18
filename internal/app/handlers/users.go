package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type User struct {
	UserID   int    `json:"userid"`
	Ads      string `json:"ads"`
	Username string `json:"username"`
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingUser User
	err := db.QueryRow("SELECT userid FROM users WHERE username = ?", user.Username).Scan(&existingUser.UserID)
	if err == nil {
		log.Printf("User already exists: %v", existingUser)
		w.WriteHeader(http.StatusOK)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("Error checking for existing user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := db.Exec("INSERT INTO users (userid, username) VALUES (?, ?)", user.UserID, user.Username)
	if err != nil {
		log.Printf("Error inserting user into database: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error getting last insert ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.UserID = int(id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("User created: %v", user)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT userid, ads, username FROM users")
	if err != nil {
		log.Printf("Error querying users: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.UserID, &user.Ads, &user.Username); err != nil {
			log.Printf("Error scanning user: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("Users retrieved: %v", users)
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userid"]

	var user User
	err := db.QueryRow("SELECT userid, ads, username FROM users WHERE userid = ?", userID).Scan(
		&user.UserID, &user.Ads, &user.Username,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			log.Printf("Error querying user by ID: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("User retrieved: %v", user)
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userid"]

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingUser User
	err := db.QueryRow("SELECT userid FROM users WHERE userid = ?", userID).Scan(&existingUser.UserID)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Error checking existing user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stmt, err := db.Prepare("UPDATE users SET ads = ?, username = ? WHERE userid = ?")
	if err != nil {
		log.Printf("Error preparing statement: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Ads, user.Username, userID)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = db.QueryRow("SELECT userid, ads, username FROM users WHERE userid = ?", userID).Scan(&user.UserID, &user.Ads, &user.Username)
	if err != nil {
		log.Printf("Error fetching updated user: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
	log.Printf("User updated: %v", user)
}

func GetAdsByUserID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userid"]
	rows, err := db.Query("SELECT ads FROM users WHERE userid = ?", userID)
	if err != nil {
		log.Printf("Error querying ads by user ID: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	ads := []string{}
	for rows.Next() {
		var ad string
		if err := rows.Scan(&ad); err != nil {
			log.Printf("Error scanning ad: %v", err)
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
