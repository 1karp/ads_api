package router

import (
	"github.com/1karp/ads_api/internal/app/handlers"
	"github.com/gorilla/mux"
)

func SetupRoutes() *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/ads", handlers.CreateAd).Methods("POST")
	router.HandleFunc("/ads", handlers.GetAds).Methods("GET")
	router.HandleFunc("/ads/{id}", handlers.GetAdByID).Methods("GET")
	router.HandleFunc("/ads/{id}", handlers.UpdateAd).Methods("PUT")
	router.HandleFunc("/ads/{id}/post", handlers.PostAd).Methods("POST")
	router.HandleFunc("/ads/{id}/edit-post", handlers.EditAdInTelegram).Methods("POST")

	router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	router.HandleFunc("/users/{userid}", handlers.GetUserByID).Methods("GET")
	router.HandleFunc("/users/{userid}", handlers.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{userid}/ads", handlers.GetAdsByUserID).Methods("GET")

	return router
}
