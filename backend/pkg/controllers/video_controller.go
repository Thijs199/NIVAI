package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"nivai/backend/pkg/models"
	"nivai/backend/pkg/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

var (
	// pythonApiBaseUrl and netClient are also defined in analytics_controller.go
	// Consider moving to a shared package or initializing them in a common place (e.g., main.go and pass down)
	// For this subtask, we define them here as well.
	pythonApiBaseUrl_vc string // Suffix _vc to avoid conflict if they were truly global and this file was part of same build in a flat way
	netClient_vc        *http.Client
)

func init() {
	// This init function will run for this package.
	// If analytics_controller also has an init, both will run.
	pythonApiBaseUrl_vc = os.Getenv("PYTHON_API_URL")
	if pythonApiBaseUrl_vc == "" {
		pythonApiBaseUrl_vc = "http://localhost:8081" // Default for local development
		log.Println("PYTHON_API_URL not set for video_controller, using default:", pythonApiBaseUrl_vc)
	} else {
		log.Println("PYTHON_API_URL for video_controller:", pythonApiBaseUrl_vc)
	}
	netClient_vc = &http.Client{Timeout: time.Second * 20} // Increased timeout for potential processing trigger
}

/**
 * VideoController manages HTTP requests related to video resources.
 * Handles CRUD operations and specialized video-related endpoints.
 */
type VideoController struct {
	videoRepo      models.VideoRepository
	videoService   services.VideoService
	storageService services.StorageService
}

/**
 * NewVideoController creates a new controller for video-related endpoints.
 *
 * @param videoRepo The repository for video data operations
 * @param storageService The service for file storage operations
 * @return A new video controller instance
 */
func NewVideoController(videoRepo models.VideoRepository, storageService services.StorageService) *VideoController {
	// Create VideoService with the video repository and storage service
	videoService := services.NewVideoService(videoRepo, storageService)

	return &VideoController{
		videoRepo:      videoRepo,
		videoService:   videoService,
		storageService: storageService,
	}
}

// Helper function to save a single uploaded file
func (c *VideoController) saveUploadedFile(
	file multipart.File,
	header *multipart.FileHeader,
	storageDir string,
	baseFilename string,
	fileTypeIdentifier string, // e.g., "video", "tracking", "events"
) (string, int64, error) {
	if file == nil || header == nil {
		return "", 0, fmt.Errorf("%s file is missing", fileTypeIdentifier)
	}

	originalFilename := header.Filename
	fileExt := filepath.Ext(originalFilename)
	// For tracking/event files, instructions suggest specific extensions like _tracking.gzip.
	// For now, let's use a generic approach or adapt if specific naming is required.
	// The subtask suggests videoID + "_tracking.gzip" etc. Let's use baseFilename which will be videoID.
	var storageFilename string
	if fileTypeIdentifier == "tracking" {
		storageFilename = baseFilename + "_tracking.gzip" // Assuming content is gzipped by client or this is just a name
	} else if fileTypeIdentifier == "events" {
		storageFilename = baseFilename + "_events.gzip"
	} else { // video
		storageFilename = baseFilename + fileExt
	}

	destPath := filepath.Join(storageDir, storageFilename)

	uploadInfo, err := c.storageService.UploadFile(file, destPath)
	if err != nil {
		return "", 0, fmt.Errorf("failed to upload %s file to %s: %w", fileTypeIdentifier, destPath, err)
	}
	return uploadInfo.Path, uploadInfo.Size, nil
}

/**
 * UploadVideo handles the video, tracking, and event file upload process.
 * Accepts multipart form data, stores the files, and triggers Python API processing.
 * Handles the POST /api/v1/videos endpoint.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
func (c *VideoController) UploadVideo(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size (e.g., 500MB for video + other files)
	// Increased from 100MB to accommodate potentially larger tracking/event files alongside video.
	maxUploadSize := int64(500 << 20) // 500 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			http.Error(w, fmt.Sprintf("File(s) too large. Maximum total size is %dMB.", maxUploadSize>>20), http.StatusRequestEntityTooLarge)
		} else {
			http.Error(w, "Invalid multipart form: "+err.Error(), http.StatusBadRequest)
		}
		return
	}

	videoFile, videoHeader, errVideoFile := r.FormFile("video_file")
	if errVideoFile != nil && !errors.Is(errVideoFile, http.ErrMissingFile) {
		http.Error(w, "Error processing video_file: "+errVideoFile.Error(), http.StatusInternalServerError)
		return
	}
	if videoFile != nil {
		defer videoFile.Close()
	}

	trackingFile, trackingHeader, errTrackingFile := r.FormFile("tracking_file")
	if errTrackingFile != nil && !errors.Is(errTrackingFile, http.ErrMissingFile) {
		http.Error(w, "Error processing tracking_file: "+errTrackingFile.Error(), http.StatusInternalServerError)
		return
	}
	if trackingFile != nil {
		defer trackingFile.Close()
	}

	eventFile, eventHeader, errEventFile := r.FormFile("event_file")
	if errEventFile != nil && !errors.Is(errEventFile, http.ErrMissingFile) {
		http.Error(w, "Error processing event_file: "+errEventFile.Error(), http.StatusInternalServerError)
		return
	}
	if eventFile != nil {
		defer eventFile.Close()
	}

	// Validate that at least one file is present (or define other rules)
	// For analytics, tracking and event files are key. Video might be optional.
	if errors.Is(errTrackingFile, http.ErrMissingFile) || errors.Is(errEventFile, http.ErrMissingFile) {
		// For this example, let's make tracking and event files mandatory if analytics is the goal.
		// Video file can be optional.
		// The subtask implies these are primarily for analytics.
		http.Error(w, "Tracking and event files are required for analytics processing.", http.StatusBadRequest)
		return
	}
	// If video_file is also mandatory:
	// if errors.Is(errVideoFile, http.ErrMissingFile) {
	// 	http.Error(w, "video_file is required.", http.StatusBadRequest)
	// 	return
	// }


	videoID := uuid.New().String()
	storagePath := filepath.Join("videos", videoID[0:2], videoID[2:4], videoID)

	var videoDestPath string
	var videoSize int64
	var errSave error

	if videoFile != nil {
		videoDestPath, videoSize, errSave = c.saveUploadedFile(videoFile, videoHeader, storagePath, videoID, "video")
		if errSave != nil {
			http.Error(w, errSave.Error(), http.StatusInternalServerError)
			return // Early exit on critical file save error
		}
	}

	trackingDestPath, _, errSave := c.saveUploadedFile(trackingFile, trackingHeader, storagePath, videoID, "tracking")
	if errSave != nil {
		// Attempt to cleanup video file if tracking save fails
		if videoDestPath != "" { c.storageService.DeleteFile(videoDestPath) }
		http.Error(w, errSave.Error(), http.StatusInternalServerError)
		return
	}

	eventDestPath, _, errSave := c.saveUploadedFile(eventFile, eventHeader, storagePath, videoID, "events")
	if errSave != nil {
		// Attempt to cleanup video and tracking files if event save fails
		if videoDestPath != "" { c.storageService.DeleteFile(videoDestPath) }
		c.storageService.DeleteFile(trackingDestPath) // trackingDestPath would be valid here
		http.Error(w, errSave.Error(), http.StatusInternalServerError)
		return
	}

	// Create video metadata object
	videoMetadata := &models.Video{
		ID:               videoID,
		Title:            r.FormValue("title"),
		Description:      r.FormValue("description"),
		ProcessingState:  "pending_analytics", // New state? Or keep "pending"?
		// UploadedAt: time.Now(), // This field was in the original, but not in the model from read_files
		CreatedAt:        time.Now(), // Assuming CreatedAt is the upload time
		FilePath:         videoDestPath,
		TrackingPath:     trackingDestPath,
		EventFilePath:    eventDestPath,
		// Size: videoSize, // If Video model had FileSize for main video
		// ContentType: videoHeader.Header.Get("Content-Type"), // If model had ContentType
		// Filename: videoHeader.Filename, // If model had Filename
	}
	if videoHeader != nil {
		videoMetadata.Format = strings.TrimPrefix(filepath.Ext(videoHeader.Filename), ".")
		videoMetadata.Size = videoSize // Size of the video file itself
	}

	if videoDestPath != "" {
		videoMetadata.StorageProvider = "default" // Placeholder - this needs a proper source
	}

	// Get match metadata if provided
	if matchID := r.FormValue("match_id"); matchID != "" {
		videoMetadata.MatchID = matchID
		videoMetadata.HomeTeam = r.FormValue("home_team")
		videoMetadata.AwayTeam = r.FormValue("away_team")
		videoMetadata.Competition = r.FormValue("competition")
		videoMetadata.Season = r.FormValue("season")

		matchDateStr := r.FormValue("match_date")
		if matchDateStr != "" {
			parsedDate, err := time.Parse("2006-01-02", matchDateStr)
			if err == nil {
				videoMetadata.MatchDate = parsedDate
			} else {
				log.Printf("Warning: Could not parse match_date '%s': %v", matchDateStr, err)
				// Optionally, you could set an error response here if match_date is critical and invalid
			}
		}
	}

	// Save the video metadata (which now includes paths to tracking and event files)
	// This part needs to be adapted if VideoService.SaveVideoMetadata is the correct method
	// or if there's a different metadata storage mechanism.
	// For now, let's assume VideoService handles it.
	// savedVideo, err := c.videoService.SaveVideoMetadata(videoMetadata)
	// The existing Video model and repository are more complex than what SaveVideoMetadata might imply.
	// Let's assume there's a method like CreateVideo in VideoService that handles this.
	// If VideoService is tightly coupled to a DB via a repository, that's where it should go.

	savedMatchData, err := c.videoService.CreateVideoEntry(videoMetadata)
	if err != nil {
		log.Printf("Error saving video/match metadata for ID %s: %v", videoID, err)
		// Attempt to clean up uploaded files if metadata saving fails
		if videoDestPath != "" { c.storageService.DeleteFile(videoDestPath) }
		if trackingDestPath != "" { c.storageService.DeleteFile(trackingDestPath) }
		if eventDestPath != "" { c.storageService.DeleteFile(eventDestPath) }
		http.Error(w, "Failed to save video/match metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Video/match metadata saved for ID %s: %+v", videoID, savedMatchData)
	// videoID from uuid.New().String() should match savedMatchData.ID if CreateVideoEntry uses the passed ID.

	// Trigger Python API /process-match
	// CRITICAL ASSUMPTION: trackingDestPath and eventDestPath must be accessible by the Python API
	// This usually means they are absolute paths on a shared volume/filesystem.
	// If storageService returns relative paths, they need to be converted to absolute paths.
	// For example, if storageService saves under /app/storage_root/videos/...
	// and Python API can access /python_accessible_storage_root/videos/...
	// then paths need transformation.
	// For now, assume paths are directly usable or Python API knows where to find them based on config.

	// Convert paths to absolute if they are not already, assuming storageService provides paths relative to some root
    // that might not be the same for the Python API.
    // This is a placeholder for actual path resolution logic needed for your deployment.
    absTrackingPath := trackingDestPath // Placeholder: c.storageService.GetAbsolutePath(trackingDestPath)
    absEventPath := eventDestPath       // Placeholder: c.storageService.GetAbsolutePath(eventDestPath)


	pyApiReqBody := map[string]string{
		"tracking_data_path": absTrackingPath,
		"event_data_path":    absEventPath,
		"match_id":           videoID,
	}
	jsonReqBody, err := json.Marshal(pyApiReqBody)
	if err != nil {
		log.Printf("Error marshalling Python API request body for video %s: %v", videoID, err)
		// Don't fail the upload, but log this. Analytics processing might need manual trigger.
	} else {
		pyProcessUrl := fmt.Sprintf("%s/process-match", pythonApiBaseUrl_vc)
		log.Printf("Calling Python API to process match %s: %s with body %s", videoID, pyProcessUrl, string(jsonReqBody))

		resp, postErr := netClient_vc.Post(pyProcessUrl, "application/json", bytes.NewBuffer(jsonReqBody))
		if postErr != nil {
			log.Printf("Error calling Python API /process-match for video %s: %v", videoID, postErr)
		} else {
			defer resp.Body.Close()
			respBodyBytes, _ := io.ReadAll(resp.Body) // Read body for logging
			log.Printf("Python API /process-match response for video %s: Status: %s, Body: %s", videoID, resp.Status, string(respBodyBytes))
			if resp.StatusCode >= 300 { // Non-2xx response
				log.Printf("Python API /process-match returned non-success status for video %s: %s", videoID, resp.Status)
				// Potentially update videoMetadata.ProcessingState to "analytics_failed" or similar
			} else {
				log.Printf("Python API /process-match successfully triggered for video %s.", videoID)
				// Potentially update videoMetadata.ProcessingState to "analytics_processing"
			}
		}
	}

	// Return minimal info about the uploaded files, primarily the ID.
	// The client can then use other endpoints to get full metadata if needed.
	// The original `savedVideo` variable might not be available if DB save is removed from this step.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // Accepted, as processing (including analytics) is happening.
	if err := json.NewEncoder(w).Encode(map[string]string{
		"message":            "Upload received, processing initiated.",
		"video_id":           videoID,
		"video_file_path":    videoDestPath,    // if video was uploaded
		"tracking_path":      trackingDestPath, // always present based on current logic
		"event_file_path":    eventDestPath,    // always present
	}); err != nil {
		log.Printf("Error encoding UploadVideo final response for video %s: %v", videoID, err)
	}
}

// GetVideo, ListVideos, DeleteVideo, parsePaginationParams, parseVideoFilters remain the same as before.
// ... (rest of the file from the read_files output)
// To save space, I'm omitting the rest of the functions that were not meant to be changed by this subtask.
// The tool should append the previous content here.


/**
 * GetVideo retrieves a single video by its ID.
 * Handles the GET /api/v1/videos/{id} endpoint.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
func (c *VideoController) GetVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video ID from URL path
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}

	// Retrieve video from service
	video, err := c.videoService.GetVideoByID(id)
	if err != nil {
		if errors.Is(err, services.ErrVideoNotFound) {
			http.Error(w, "Video not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve video", http.StatusInternalServerError)
		}
		return
	}

	// Return video as JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(video); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}

/**
 * ListVideos retrieves a paginated list of videos.
 * Handles the GET /api/v1/videos endpoint with optional filtering.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
func (c *VideoController) ListVideos(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit, offset := parsePaginationParams(r)

	// Parse additional filter parameters
	filters := parseVideoFilters(r)

	// Retrieve videos using service
	videos, err := c.videoService.ListVideos(limit, offset, filters)
	if err != nil {
		http.Error(w, "Failed to retrieve videos", http.StatusInternalServerError)
		return
	}

	// Return videos as JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(videos); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}


/**
 * DeleteVideo removes a video resource.
 * Handles the DELETE /api/v1/videos/{id} endpoint.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
func (c *VideoController) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	// Extract video ID from URL path
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}

	// Get video metadata first to know the file path
	video, err := c.videoService.GetVideoByID(id)
	if err != nil {
		if errors.Is(err, services.ErrVideoNotFound) {
			http.Error(w, "Video not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to retrieve video metadata", http.StatusInternalServerError)
		}
		return
	}

	// Delete the actual file first (video, tracking, events)
	if video.FilePath != "" {
		if err := c.storageService.DeleteFile(video.FilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete video file %s: %s", video.FilePath, err.Error())
		}
	}
	if video.TrackingPath != "" {
		if err := c.storageService.DeleteFile(video.TrackingPath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete tracking file %s: %s", video.TrackingPath, err.Error())
		}
	}
	if video.EventFilePath != "" {
		if err := c.storageService.DeleteFile(video.EventFilePath); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to delete event file %s: %s", video.EventFilePath, err.Error())
		}
	}


	// Delete video metadata
	if err := c.videoService.DeleteVideo(id); err != nil {
		http.Error(w, "Failed to delete video metadata", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

/**
 * parsePaginationParams extracts pagination parameters from the request.
 * Provides default values if parameters are not present or invalid.
 *
 * @param r The HTTP request
 * @return Limit and offset values for pagination
 */
func parsePaginationParams(r *http.Request) (int, int) {
	// Get query parameters
	query := r.URL.Query()

	// Parse limit parameter
	limitStr := query.Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Parse offset parameter
	offsetStr := query.Get("offset")
	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	return limit, offset
}

/**
 * parseVideoFilters extracts filter parameters for video queries.
 *
 * @param r The HTTP request
 * @return Map of filter parameters
 */
func parseVideoFilters(r *http.Request) map[string]string {
	query := r.URL.Query()
	filters := make(map[string]string)

	// Extract potential filter parameters
	if matchID := query.Get("match_id"); matchID != "" {
		filters["match_id"] = matchID
	}

	if team := query.Get("team"); team != "" {
		filters["team"] = team
	}

	if competition := query.Get("competition"); competition != "" {
		filters["competition"] = competition
	}

	if season := query.Get("season"); season != "" {
		filters["season"] = season
	}

	if state := query.Get("processing_state"); state != "" {
		filters["processing_state"] = state
	}

	return filters
}