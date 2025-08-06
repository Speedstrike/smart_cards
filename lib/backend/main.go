package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.Use(enableCORS)
	connect()

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]string{
			"status":  "OK",
			"message": "flashcard backend is running",
			"time":    time.Now().Format(time.RFC3339),
		}
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	router.HandleFunc("/upload", uploadFilesHandler).Methods("POST")
	router.HandleFunc("/flashcards", flashcardsHandler).Methods("GET")

	log.Println("server running on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}
