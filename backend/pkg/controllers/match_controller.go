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

var (
	pythonApiBaseUrl_mc string
	netClient_mc        *http.Client
)

func init() {
	pythonApiBaseUrl_mc = os.Getenv("PYTHON_API_URL")
	if pythonApiBaseUrl_mc == "" {
		pythonApiBaseUrl_mc = "http://localhost:8081"
		log.Println("PYTHON_API_URL not set for match_controller, using default:", pythonApiBaseUrl_mc)
	} else {
		log.Println("PYTHON_API_URL for match_controller:", pythonApiBaseUrl_mc)
	}
	netClient_mc = &http.Client{Timeout: time.Second * 10}
}

type MatchController struct {
	videoService services.VideoService
	// If pythonApiBaseUrl_mc and netClient_mc were to be controller-specific,
	// they would be fields here, initialized by NewMatchController.
	// For now, package-level variables are used as per existing pattern in other controllers.
}

func NewMatchController(vs services.VideoService) *MatchController {
	return &MatchController{
		videoService: vs,
	}
}

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

type PythonStatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

func (mc *MatchController) getAnalyticsStatus(matchID string, wg *sync.WaitGroup, statusChan chan<- struct{id string; status string; err error}) {
    if wg != nil {
        defer wg.Done()
    }

	statusUrl := fmt.Sprintf("%s/match/%s/status", pythonApiBaseUrl_mc, matchID)
	var analyticsStatus string
	var anError error

	resp, err := netClient_mc.Get(statusUrl)
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
			bodyBytes, _ := ioutil.ReadAll(resp.Body)
			log.Printf("Non-OK status (%s) fetching analytics status for match %s: %s", resp.Status, matchID, string(bodyBytes))
			analyticsStatus = fmt.Sprintf("error_status_%d", resp.StatusCode)
			anError = fmt.Errorf("status %d", resp.StatusCode)
		}
	}
	statusChan <- struct{id string; status string; err error}{matchID, analyticsStatus, anError}
}


func (mc *MatchController) ListMatches(w http.ResponseWriter, r *http.Request) {
	// For now, using default limit/offset and no filters.
	// These could be parsed from r.URL.Query() similar to video_controller.go/parsePaginationParams
	defaultLimit := 20 // Example default
	defaultOffset := 0
	videos, err := mc.videoService.ListVideos(defaultLimit, defaultOffset, make(map[string]string))
	if err != nil {
		log.Printf("Error listing videos: %v", err)
		http.Error(w, "Failed to retrieve match list", http.StatusInternalServerError)
		return
	}

	if videos == nil { // Ensure videos is not nil, even if empty
        videos = []*models.Video{}
    }

	matchListItems := make([]MatchListItem, len(videos))

    statusChan := make(chan struct{id string; status string; err error}, len(videos))
    var wg sync.WaitGroup

	if len(videos) > 0 {
		for _, video := range videos {
			wg.Add(1)
			// Launch goroutine to fetch status
			go mc.getAnalyticsStatus(video.ID, &wg, statusChan)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(statusChan)

		// Collect statuses
		statuses := make(map[string]string)
		for res := range statusChan {
			if res.err != nil {
                 // Log the error, status will contain an error indicator
                 log.Printf("Error detail for match %s status check: %v", res.id, res.err)
            }
			statuses[res.id] = res.status
		}

		// Populate matchListItems with statuses
		for i, video := range videos {
			matchListItems[i] = MatchListItem{
				ID:              video.ID,
				MatchName:       video.Title,       // Assuming Title is used for Match Name
				UploadDate:      video.CreatedAt,   // Assuming CreatedAt is the upload date
				AnalyticsStatus: statuses[video.ID], // Fetched status
				HomeTeam:        video.HomeTeam,
				AwayTeam:        video.AwayTeam,
				Competition:     video.Competition,
				Season:          video.Season,
			}
		}
	} else {
        // If there are no videos, statusChan would not be closed by wg.Wait(), so close it here.
        close(statusChan)
    }


	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matchListItems); err != nil {
		log.Printf("Error encoding match list response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
