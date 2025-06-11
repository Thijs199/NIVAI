package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"nivai/backend/pkg/models"
	"nivai/backend/pkg/services"
	// "github.com/gorilla/mux" // Not strictly needed if not extracting path vars here
)

// MatchController handles requests related to matches.
type MatchController struct {
	videoService     services.VideoService
	PythonApiBaseUrl string
	HttpClient       *http.Client
}

// NewMatchController creates a new MatchController.
// If pythonApiBaseUrl is empty, it tries to get it from PYTHON_API_URL env var,
// then defaults to "http://localhost:8081".
// If client is nil, a default client with a 10-second timeout is used.
func NewMatchController(vs services.VideoService, pythonApiBaseUrl string, client *http.Client) *MatchController {
	if pythonApiBaseUrl == "" {
		envURL := os.Getenv("PYTHON_API_URL")
		if envURL != "" {
			pythonApiBaseUrl = envURL
		} else {
			pythonApiBaseUrl = "http://localhost:8081" // Default
		}
		log.Println("Using Python API URL for MatchController:", pythonApiBaseUrl)
	}
	if client == nil {
		client = &http.Client{Timeout: time.Second * 10}
	}
	return &MatchController{
		videoService:     vs,
		PythonApiBaseUrl: pythonApiBaseUrl,
		HttpClient:       client,
	}
}

// MatchListItem represents a single item in the list of matches.
type MatchListItem struct {
	ID              string    `json:"id"`
	MatchName       string    `json:"match_name"` // This is video.Title
	UploadDate      time.Time `json:"upload_date"` // This is video.CreatedAt
	AnalyticsStatus string    `json:"analytics_status"`
	HomeTeam        string    `json:"home_team,omitempty"`
	AwayTeam        string    `json:"away_team,omitempty"`
	Competition     string    `json:"competition,omitempty"`
	Season          string    `json:"season,omitempty"`
	// Potentially other fields like video thumbnail, duration etc.
}

// PythonStatusResponse is used to decode the status from the Python API.
// Note: This struct might be duplicated in tests if not exported or shared.
// For now, keeping it unexported as it's specific to this controller's interaction.
type PythonStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// getAnalyticsStatus fetches the analytics status for a given match ID.
// It's a method of MatchController now.
func (mc *MatchController) getAnalyticsStatus(matchID string, wg *sync.WaitGroup, statusChan chan<- struct {
	id     string
	status string
	err    error
}) {
	if wg != nil {
		defer wg.Done()
	}

	statusUrl := fmt.Sprintf("%s/match/%s/status", mc.PythonApiBaseUrl, matchID)
	var analyticsStatus string
	var anError error

	resp, err := mc.HttpClient.Get(statusUrl)
	if err != nil {
		log.Printf("Error fetching analytics status for match %s: %v", matchID, err)
		analyticsStatus = "error_fetching_status"
		anError = err
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			var statusResp PythonStatusResponse
			if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
				log.Printf("Error decoding analytics status for match %s: %v", matchID, err)
				analyticsStatus = "error_decoding_status"
				anError = err
			} else {
				analyticsStatus = statusResp.Status
			}
		} else {
			bodyBytes, _ := ioutil.ReadAll(resp.Body) // Read body for more context on error
			log.Printf("Non-OK status (%s) fetching analytics status for match %s: %s", resp.Status, matchID, string(bodyBytes))
			analyticsStatus = fmt.Sprintf("error_status_%d", resp.StatusCode)
			anError = fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
		}
	}
	statusChan <- struct {
		id     string
		status string
		err    error
	}{matchID, analyticsStatus, anError}
}

// ListMatches handles requests to list all matches.
func (mc *MatchController) ListMatches(w http.ResponseWriter, r *http.Request) {
	defaultLimit := 20
	defaultOffset := 0
	videos, err := mc.videoService.ListVideos(defaultLimit, defaultOffset, make(map[string]string))
	if err != nil {
		log.Printf("Error listing videos: %v", err)
		http.Error(w, "Failed to retrieve match list", http.StatusInternalServerError)
		return
	}

	if videos == nil {
		videos = []*models.Video{}
	}

	matchListItems := make([]MatchListItem, len(videos))
	statusChan := make(chan struct {
		id     string
		status string
		err    error
	}, len(videos))
	var wg sync.WaitGroup

	if len(videos) > 0 {
		for _, video := range videos {
			wg.Add(1)
			go mc.getAnalyticsStatus(video.ID, &wg, statusChan)
		}

		wg.Wait()
		close(statusChan)

		statuses := make(map[string]string)
		for res := range statusChan {
			if res.err != nil {
				log.Printf("Error detail for match %s status check: %v", res.id, res.err)
			}
			statuses[res.id] = res.status
		}

		for i, video := range videos {
			matchListItems[i] = MatchListItem{
				ID:              video.ID,
				MatchName:       video.Title,
				UploadDate:      video.CreatedAt,
				AnalyticsStatus: statuses[video.ID],
				HomeTeam:        video.HomeTeam,
				AwayTeam:        video.AwayTeam,
				Competition:     video.Competition,
				Season:          video.Season,
			}
		}
	} else {
		close(statusChan) // Ensure channel is closed even if no videos
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matchListItems); err != nil {
		log.Printf("Error encoding match list response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
