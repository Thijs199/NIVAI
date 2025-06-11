package services

import (
	"errors"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"nivai/backend/pkg/models"
)

// Common service errors
var (
	ErrVideoNotFound = errors.New("video not found")
	ErrInvalidVideo  = errors.New("invalid video data")
	ErrStorageFailed = errors.New("storage operation failed")
)

/**
 * VideoService defines the interface for video-related business logic.
 * Abstracts operations related to video management and processing.
 */
type VideoService interface {
	GetVideoByID(id string) (*models.Video, error)
	ListVideos(limit, offset int, filters map[string]string) ([]*models.Video, error)
	UploadVideo(file multipart.File, header *multipart.FileHeader, metadata *models.Video) (*models.Video, error)
	DeleteVideo(id string) error
	GetVideoStreamURL(id string) (string, error)
	ProcessVideo(id string) error
	CreateVideoEntry(metadata *models.Video) (*models.Video, error)
}

/**
 * DefaultVideoService implements the VideoService interface.
 * Provides concrete implementations of video-related operations.
 */
type DefaultVideoService struct {
	videoRepo      models.VideoRepository
	storageService StorageService
	// Add more dependencies as needed (e.g., queue service, notification service)
}

/**
 * NewVideoService creates a new video service instance.
 *
 * @param videoRepo Repository for video data access
 * @param storageService Service for file storage operations
 * @return A new video service implementation
 */
func NewVideoService(videoRepo models.VideoRepository, storageService StorageService) VideoService {
	return &DefaultVideoService{
		videoRepo:      videoRepo,
		storageService: storageService,
	}
}

/**
 * GetVideoByID retrieves a video by its unique identifier.
 * Validates the ID and delegates to the repository for data access.
 *
 * @param id The unique ID of the video to retrieve
 * @return The video if found, or an error
 */
func (s *DefaultVideoService) GetVideoByID(id string) (*models.Video, error) {
	if id == "" {
		return nil, errors.New("video ID cannot be empty")
	}

	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		// Check if it's a "not found" error from the repository
		if strings.Contains(err.Error(), "not found") {
			return nil, ErrVideoNotFound
		}
		return nil, err
	}

	return video, nil
}

/**
 * ListVideos retrieves a filtered, paginated list of videos.
 * Processes filters and delegates to the repository for data access.
 *
 * @param limit Maximum number of videos to return
 * @param offset Number of videos to skip for pagination
 * @param filters Map of filter criteria
 * @return A slice of videos matching the criteria, or an error
 */
func (s *DefaultVideoService) ListVideos(limit, offset int, filters map[string]string) ([]*models.Video, error) {
	// Apply default pagination if needed
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Process filters
	if matchID, ok := filters["match_id"]; ok && matchID != "" {
		// Return videos for a specific match
		return s.videoRepo.FindByMatchID(matchID)
	}

	if team, ok := filters["team"]; ok && team != "" {
		// Return videos for a specific team
		return s.videoRepo.FindByTeam(team, limit, offset)
	}

	if state, ok := filters["processing_state"]; ok && state != "" {
		// Return videos with a specific processing state
		return s.videoRepo.FindByProcessingState(state, limit, offset)
	}

	// If no specific filters are applied, return all videos with pagination
	return s.videoRepo.FindAll(limit, offset)
}

/**
 * UploadVideo handles the file upload and storage process.
 * Validates the file, stores it, and creates metadata in the database.
 *
 * @param file The multipart file from the HTTP request
 * @param header The file header with metadata
 * @param metadata The video metadata provided by the client
 * @return The created video record, or an error
 */
func (s *DefaultVideoService) UploadVideo(file multipart.File, header *multipart.FileHeader, metadata *models.Video) (*models.Video, error) {
	// Validate file type
	if !isValidVideoType(header.Filename) {
		return nil, errors.New("invalid video file type")
	}

	// Validate metadata
	if metadata.Title == "" {
		return nil, errors.New("video title is required")
	}

	// Generate storage path
	storagePath := generateStoragePath(metadata)

	// Upload file to storage
	uploadInfo, err := s.storageService.UploadFile(file, storagePath)
	if err != nil {
		return nil, ErrStorageFailed
	}

	// Update metadata with storage information
	metadata.FilePath = uploadInfo.Path
	metadata.StorageProvider = uploadInfo.Provider
	metadata.Size = uploadInfo.Size
	metadata.Format = uploadInfo.Format
	metadata.ProcessingState = "pending"
	metadata.CreatedAt = time.Now()
	metadata.UpdatedAt = time.Now()

	// Save metadata to database
	if err := s.videoRepo.Create(metadata); err != nil {
		// If database save fails, try to clean up the uploaded file
		_ = s.storageService.DeleteFile(uploadInfo.Path)
		return nil, err
	}

	// Queue video for processing (extraction of duration, resolution, etc.)
	go s.ProcessVideo(metadata.ID)

	return metadata, nil
}

/**
 * DeleteVideo removes a video and its associated resources.
 * Performs a soft delete in the database and optionally in storage.
 *
 * @param id The unique ID of the video to delete
 * @return Error if deletion fails
 */
func (s *DefaultVideoService) DeleteVideo(id string) error {
	// The controller calls GetVideoByID first. This service method should focus on deletion.
	// The FindByID call here is redundant and causes issues with mocking call counts.
	// video, err := s.videoRepo.FindByID(id) // This was the redundant call
	// if err != nil {
	// 	if strings.Contains(err.Error(), "not found") {
	// 		return ErrVideoNotFound
	// 	}
	// 	return err
	// }

	// Soft delete in database
	if err := s.videoRepo.Delete(id); err != nil {
		if strings.Contains(err.Error(), "not found") { // Or use errors.Is if a specific error var exists
			return ErrVideoNotFound
		}
		return err
	}
	return nil
}

/**
 * GetVideoStreamURL generates a URL for streaming the video.
 * May create temporary authenticated URLs for cloud storage.
 *
 * @param id The unique ID of the video
 * @return Streaming URL, or an error
 */
func (s *DefaultVideoService) GetVideoStreamURL(id string) (string, error) {
	// Get video metadata
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return "", ErrVideoNotFound
		}
		return "", err
	}

	// Generate streaming URL based on storage provider
	streamURL, err := s.storageService.GetStreamURL(video.FilePath)
	if err != nil {
		return "", err
	}

	return streamURL, nil
}

/**
 * ProcessVideo initiates or handles video processing.
 * May extract metadata, generate thumbnails, or prepare for analysis.
 *
 * @param id The unique ID of the video to process
 * @return Error if processing fails
 */
func (s *DefaultVideoService) ProcessVideo(id string) error {
	// Get video metadata
	video, err := s.videoRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Update processing state
	video.ProcessingState = "processing"
	video.UpdatedAt = time.Now()
	if err := s.videoRepo.Update(video); err != nil {
		return err
	}

	// TODO: Implement actual processing
	// This would typically be handled by a separate service or worker
	// For now, we'll just simulate processing by updating some fields

	// Simulate extraction of video properties
	video.Duration = 120.5 // Example: 2 minutes and 30 seconds
	video.Resolution = "1920x1080"

	// Update processing state to completed
	video.ProcessingState = "completed"
	video.UpdatedAt = time.Now()

	return s.videoRepo.Update(video)
}

/**
 * isValidVideoType checks if the file extension is an allowed video format.
 *
 * @param filename The name of the file to validate
 * @return Whether the file type is valid
 */
func isValidVideoType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".mp4":  true,
		".mov":  true,
		".avi":  true,
		".mkv":  true,
		".webm": true,
	}

	return validExtensions[ext]
}

/**
 * generateStoragePath creates a unique path for storing the video.
 * Typically organizes files by date, type, etc. for easy management.
 *
 * @param metadata The video metadata
 * @return Storage path for the video
 */
func generateStoragePath(metadata *models.Video) string {
	// Generate path based on date and ID
	datePrefix := time.Now().Format("2006/01/02")
	fileName := metadata.ID + filepath.Ext(metadata.FilePath)

	// If this is associated with a match, include match info in path
	if metadata.MatchID != "" {
		return filepath.Join("videos", "matches", metadata.MatchID, datePrefix, fileName)
	}

	return filepath.Join("videos", "uploads", datePrefix, fileName)
}

func (s *DefaultVideoService) CreateVideoEntry(metadata *models.Video) (*models.Video, error) {
	if metadata.ID == "" {
		// Or generate UUID here if not already set by controller
		return nil, errors.New("metadata ID is required")
	}
	metadata.UpdatedAt = time.Now()
	// CreatedAt should already be set by the controller
	// StorageProvider might also be set by controller after uploading main video file

	if err := s.videoRepo.Create(metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

// GenerateStoragePathForTesting provides access to the unexported generateStoragePath for testing purposes.
func GenerateStoragePathForTesting(metadata *models.Video) string {
	return generateStoragePath(metadata)
}
