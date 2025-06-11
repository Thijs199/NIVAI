package controllers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// AnalyticsController handles requests for analytics data.
type AnalyticsController struct {
	PythonApiBaseUrl string
	HttpClient       *http.Client
}

// NewAnalyticsController creates a new AnalyticsController.
// If pythonApiBaseUrl is empty, it tries to get it from PYTHON_API_URL env var,
// then defaults to "http://localhost:8081".
// If client is nil, a default client with a 10-second timeout is used.
func NewAnalyticsController(pythonApiBaseUrl string, client *http.Client) *AnalyticsController {
	if pythonApiBaseUrl == "" {
		envURL := os.Getenv("PYTHON_API_URL")
		if envURL != "" {
			pythonApiBaseUrl = envURL
		} else {
			pythonApiBaseUrl = "http://localhost:8081" // Default
		}
		log.Println("AnalyticsController: Using Python API URL:", pythonApiBaseUrl)
	}
	if client == nil {
		client = &http.Client{Timeout: time.Second * 10}
	}
	return &AnalyticsController{
		PythonApiBaseUrl: pythonApiBaseUrl,
		HttpClient:       client,
	}
}

// relayRequest is a helper method to relay requests to the Python API.
func (ac *AnalyticsController) relayRequest(w http.ResponseWriter, r *http.Request, targetUrl string, handlerName string) {
	log.Printf("[%s] Relaying request to: %s", handlerName, targetUrl)

	resp, err := ac.HttpClient.Get(targetUrl)
	if err != nil {
		log.Printf("[%s] Error making GET request to Python API (%s): %v", handlerName, targetUrl, err)
		http.Error(w, fmt.Sprintf("Error connecting to analytics service: %v", err), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
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
	}
}

// GetMatchAnalytics handles requests for match analytics.
// Path: /analytics/match/{id}
func (ac *AnalyticsController) GetMatchAnalytics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	matchID, ok := vars["id"]
	if !ok {
		log.Println("[GetMatchAnalytics] Error: match_id not found in path variables")
		http.Error(w, "Match ID is required in path", http.StatusBadRequest)
		return
	}

	targetUrl := fmt.Sprintf("%s/match/%s/stats/summary", ac.PythonApiBaseUrl, matchID)
	ac.relayRequest(w, r, targetUrl, "GetMatchAnalytics")
}

// GetPlayerAnalytics handles requests for player analytics.
// Path: /analytics/player/{id}?match_id=<match_id_value>
func (ac *AnalyticsController) GetPlayerAnalytics(w http.ResponseWriter, r *http.Request) {
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

	targetUrl := fmt.Sprintf("%s/match/%s/player/%s/details", ac.PythonApiBaseUrl, matchID, playerID)
	ac.relayRequest(w, r, targetUrl, "GetPlayerAnalytics")
}

// GetTeamAnalytics handles requests for team analytics.
// Path: /analytics/team/{id}?match_id=<match_id_value>
func (ac *AnalyticsController) GetTeamAnalytics(w http.ResponseWriter, r *http.Request) {
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

	targetUrl := fmt.Sprintf("%s/match/%s/team/%s/summary-over-time", ac.PythonApiBaseUrl, matchID, teamID)
	ac.relayRequest(w, r, targetUrl, "GetTeamAnalytics")
}
