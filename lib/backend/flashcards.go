package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/kidskoding/smart_cards/lib/backend/models"
)

func flashcardsHandler(w http.ResponseWriter, r *http.Request) {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("error loading .env file: ", err)
	}

	ctx := context.Background()
	rdb := connect()
	defer rdb.Close()

	flashcards, err := rdb.HGetAll(ctx, os.Getenv("REDIS_DB_NO")).Result()
	if err != nil {
		http.Error(w, "failed to fetch flashcards", http.StatusInternalServerError)
		log.Println("redis HGetAll error:", err)
		return
	}

	var result []models.SimpleFlashcard
	for term, definition := range flashcards {
		result = append(result, models.SimpleFlashcard{
			Term:       term,
			Definition: definition,
		})
	}

	if len(result) == 0 {
        result = []models.SimpleFlashcard{}
    }

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		log.Println("failed to encode JSON response:", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
