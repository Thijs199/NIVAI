package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"nivai/backend/pkg/models"
	"nivai/backend/pkg/services"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

/**
 * VideoController manages HTTP requests related to video resources.
 * Handles CRUD operations and specialized video-related endpoints.
 */
type VideoController struct {
	videoService  services.VideoService
	storageService services.StorageService
}

/**
 * NewVideoController creates a new controller for video-related endpoints.
 *
 * @param storageService The service for file storage operations
 * @return A new video controller instance
 */
func NewVideoController(storageService services.StorageService) *VideoController {
	// Create VideoService with the storage service
	videoService := services.NewVideoService(storageService)

	return &VideoController{
		videoService: videoService,
		storageService: storageService,
	}
}

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
 * UploadVideo handles the video upload process.
 * Accepts multipart form data and stores the video file using the storage service.
 * Handles the POST /api/v1/videos endpoint.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
func (c *VideoController) UploadVideo(w http.ResponseWriter, r *http.Request) {
	// Limit the request body size (e.g., 100MB)
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	// Parse multipart form
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "File too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	// Get the file from the request
	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Missing or invalid video file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate a unique ID for the video
	videoID := uuid.New().String()

	// Create video metadata object
	videoMetadata := &models.Video{
		ID:              videoID,
		Title:           r.FormValue("title"),
		Description:     r.FormValue("description"),
		ProcessingState: "pending",
		UploadedAt:      time.Now(),
		FileSize:        header.Size,
		ContentType:     header.Header.Get("Content-Type"),
		Filename:        header.Filename,
	}

	// Get match metadata if provided
	if matchID := r.FormValue("match_id"); matchID != "" {
		videoMetadata.MatchID = matchID
		videoMetadata.HomeTeam = r.FormValue("home_team")
		videoMetadata.AwayTeam = r.FormValue("away_team")
		videoMetadata.Competition = r.FormValue("competition")
		videoMetadata.Season = r.FormValue("season")
	}

	// Determine storage path and filename
	fileExt := filepath.Ext(header.Filename)
	storageFilename := videoID + fileExt
	storagePath := filepath.Join("videos", videoID[0:2], videoID[2:4], videoID)

	// Ensure the directory exists
	if err := c.storageService.CreateDirectory(storagePath); err != nil {
		http.Error(w, "Failed to prepare storage directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Create the destination file path
	destPath := filepath.Join(storagePath, storageFilename)

	// Create a writer to the destination on storage service
	writer, err := c.storageService.Create(destPath)
	if err != nil {
		http.Error(w, "Failed to create destination file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer writer.Close()

	// Copy the file data to the destination
	if _, err := io.Copy(writer, file); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the file path in the video metadata
	videoMetadata.FilePath = destPath

	// Save the video metadata
	savedVideo, err := c.videoService.SaveVideoMetadata(videoMetadata)
	if err != nil {
		// Try to clean up the file if metadata storage fails
		c.storageService.Delete(destPath)
		http.Error(w, "Failed to save video metadata: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return created video metadata
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(savedVideo); err != nil {
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

	// Delete the actual file first
	if video.FilePath != "" {
		if err := c.storageService.Delete(video.FilePath); err != nil && !os.IsNotExist(err) {
			// Log the error but continue with metadata deletion
			// TODO: Use a proper logger instead of printing
			println("Warning: Failed to delete video file:", err.Error())
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