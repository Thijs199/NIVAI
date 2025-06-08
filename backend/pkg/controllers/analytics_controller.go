package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var (
	pythonApiBaseUrl string
	netClient        *http.Client
)

func init() {
	pythonApiBaseUrl = os.Getenv("PYTHON_API_URL")
	if pythonApiBaseUrl == "" {
		pythonApiBaseUrl = "http://localhost:8081" // Default for local development
		log.Println("PYTHON_API_URL not set, using default:", pythonApiBaseUrl)
	} else {
		log.Println("PYTHON_API_URL:", pythonApiBaseUrl)
	}
	netClient = &http.Client{Timeout: time.Second * 10}
}

// Helper function to relay requests to the Python API
func relayRequest(w http.ResponseWriter, r *http.Request, targetUrl string, handlerName string) {
	log.Printf("[%s] Relaying request to: %s", handlerName, targetUrl)

	resp, err := netClient.Get(targetUrl)
	if err != nil {
		log.Printf("[%s] Error making GET request to Python API (%s): %v", handlerName, targetUrl, err)
		http.Error(w, fmt.Sprintf("Error connecting to analytics service: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[%s] Error reading response body from Python API (%s): %v", handlerName, targetUrl, err)
		http.Error(w, "Error reading response from analytics service", http.StatusInternalServerError)
		return
	}

	// Relay headers, status code, and body
	w.Header().Set("Content-Type", "application/json") // Assuming Python API always returns JSON
	// Potentially copy more headers from resp.Header if needed
	w.WriteHeader(resp.StatusCode)
	_, writeErr := w.Write(bodyBytes)
	if writeErr != nil {
		log.Printf("[%s] Error writing response to client: %v", handlerName, writeErr)
		// Cannot send http.Error here as headers/body might have been partially written
	}
}

// GetMatchAnalytics handles requests for match analytics.
// Path: /analytics/match/{id}
func GetMatchAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, ok := vars["id"]
	if !ok {
		log.Println("[GetMatchAnalytics] Error: match_id not found in path variables")
		http.Error(w, "Match ID is required in path", http.StatusBadRequest)
		return
	}

	targetUrl := fmt.Sprintf("%s/match/%s/stats/summary", pythonApiBaseUrl, matchID)
	relayRequest(w, r, targetUrl, "GetMatchAnalytics")
}

// GetPlayerAnalytics handles requests for player analytics.
// Path: /analytics/player/{id}?match_id=<match_id_value>
func GetPlayerAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	playerID, ok := vars["id"]
	if !ok {
		log.Println("[GetPlayerAnalytics] Error: player_id not found in path variables")
		http.Error(w, "Player ID is required in path", http.StatusBadRequest)
		return
	}

	matchID := r.URL.Query().Get("match_id")
	if matchID == "" {
		log.Println("[GetPlayerAnalytics] Error: match_id query parameter is required")
		http.Error(w, "match_id query parameter is required", http.StatusBadRequest)
		return
	}

	targetUrl := fmt.Sprintf("%s/match/%s/player/%s/details", pythonApiBaseUrl, matchID, playerID)
	relayRequest(w, r, targetUrl, "GetPlayerAnalytics")
}

// GetTeamAnalytics handles requests for team analytics.
// Path: /analytics/team/{id}?match_id=<match_id_value>
func GetTeamAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	teamID, ok := vars["id"]
	if !ok {
		log.Println("[GetTeamAnalytics] Error: team_id not found in path variables")
		http.Error(w, "Team ID is required in path", http.StatusBadRequest)
		return
	}

	matchID := r.URL.Query().Get("match_id")
	if matchID == "" {
		log.Println("[GetTeamAnalytics] Error: match_id query parameter is required")
		http.Error(w, "match_id query parameter is required", http.StatusBadRequest)
		return
	}

	targetUrl := fmt.Sprintf("%s/match/%s/team/%s/summary-over-time", pythonApiBaseUrl, matchID, teamID)
	relayRequest(w, r, targetUrl, "GetTeamAnalytics")
}
