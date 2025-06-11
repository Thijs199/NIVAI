package controllers

import (
	"encoding/json"
	"net/http"
	"time"
)

/**
 * HealthCheck provides a simple health check endpoint for the API.
 * It returns a 200 OK response with a timestamp and status.
 * This endpoint can be used by load balancers and monitoring systems.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Create response data
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "AIFAA API",
	}

	// Set content type and status code
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write JSON response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Log error but don't expose to client
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
