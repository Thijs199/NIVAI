package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url" // For url.QueryEscape
)

// PlayerController handles requests related to player data, like image searches.
type PlayerController struct {
	// Placeholder for future dependencies e.g., an image search service client
}

// NewPlayerController creates a new instance of PlayerController.
func NewPlayerController() *PlayerController {
	return &PlayerController{}
}

// SearchPlayerImage handles requests to search for a player's image.
// For now, it returns a placeholder image URL.
// Query Parameters:
// - name: The name of the player to search for.
func (pc *PlayerController) SearchPlayerImage(w http.ResponseWriter, r *http.Request) {
	playerName := r.URL.Query().Get("name")

	if playerName == "" {
		http.Error(w, "Query parameter 'name' (player name) is required.", http.StatusBadRequest)
		return
	}

	log.Printf("Received request for SearchPlayerImage for player name: %s", playerName)

	// Placeholder logic: Return a fixed placeholder image URL using via.placeholder.com
	// URL encode the player name to handle spaces or special characters in the text parameter.
	encodedPlayerName := url.QueryEscape(playerName)
	placeholderImageUrl := "https://via.placeholder.com/150/808080/FFFFFF?Text=Player+" + encodedPlayerName

	response := map[string]string{"image_url": placeholderImageUrl}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Explicitly set StatusOK

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding JSON response for SearchPlayerImage: %v", err)
		// If headers are already sent, this might not effectively change the response,
		// but it's good practice to log the error.
		// http.Error might fail if headers already written.
	}
}
