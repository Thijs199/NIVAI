package models

import (
	"database/sql"
	"errors"
	"time"
)

/**
 * Video represents a stored video file with metadata.
 * Contains information about the video file, its storage location,
 * associated metadata, and processing status.
 */
type Video struct {
	ID              string       `json:"id"`
	Title           string       `json:"title"`
	Description     string        `json:"description"`
	FilePath        string       `json:"file_path"`
	StorageProvider string       `json:"storage_provider"` // "azure_blob", "local", etc.
	Duration        float64      `json:"duration"`         // Duration in seconds
	Resolution      string       `json:"resolution"`       // e.g., "1920x1080"
	Format          string       `json:"format"`           // e.g., "mp4", "mov"
	Size            int64        `json:"size"`             // Size in bytes
	ProcessingState string       `json:"processing_state"` // "pending", "processing", "completed", "failed"
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
	DeletedAt       sql.NullTime `json:"deleted_at,omitempty"`

	// Metadata related to the match/event
	MatchID      string     `json:"match_id,omitempty"`
	MatchDate    time.Time  `json:"match_date,omitempty"`
	HomeTeam     string     `json:"home_team,omitempty"`
	AwayTeam     string     `json:"away_team,omitempty"`
	Competition  string     `json:"competition,omitempty"`
	Season       string     `json:"season,omitempty"`

	// Tracking data information
	HasTrackingData bool       `json:"has_tracking_data"`
	TrackingPath    string     `json:"tracking_path,omitempty"`
	EventFilePath   string     `json:"event_file_path,omitempty"`
}

/**
 * VideoRepository defines the interface for video data access operations.
 * Follows the repository pattern to abstract database operations.
 */
type VideoRepository interface {
	// Basic CRUD operations
	FindByID(id string) (*Video, error)
	FindAll(limit, offset int) ([]*Video, error)
	Create(video *Video) error
	Update(video *Video) error
	Delete(id string) error

	// Additional query methods
	FindByMatchID(matchID string) ([]*Video, error)
	FindByTeam(teamName string, limit, offset int) ([]*Video, error)
	FindByDateRange(start, end time.Time, limit, offset int) ([]*Video, error)
	FindByProcessingState(state string, limit, offset int) ([]*Video, error)
}

/**
 * PostgresVideoRepository implements VideoRepository using PostgreSQL.
 * Handles database operations for video data.
 */
type PostgresVideoRepository struct {
	db *sql.DB
}

/**
 * NewPostgresVideoRepository creates a new PostgreSQL-backed video repository.
 * Initializes the repository with a database connection.
 *
 * @param db Database connection
 * @return A new video repository
 */
func NewPostgresVideoRepository(db *sql.DB) VideoRepository {
	return &PostgresVideoRepository{db: db}
}

/**
 * FindByID retrieves a video by its unique identifier.
 *
 * @param id The unique ID of the video
 * @return The found video or an error
 */
func (r *PostgresVideoRepository) FindByID(id string) (*Video, error) {
	if id == "" {
		return nil, errors.New("id cannot be empty")
	}

	query := `
		SELECT id, title, description, file_path, storage_provider,
			   duration, resolution, format, size, processing_state,
			   created_at, updated_at, deleted_at,
			   match_id, match_date, home_team, away_team, competition, season,
			   has_tracking_data, tracking_path
		FROM videos
		WHERE id = $1 AND deleted_at IS NULL
	`

	var video Video
	err := r.db.QueryRow(query, id).Scan(
		&video.ID, &video.Title, &video.Description, &video.FilePath, &video.StorageProvider,
		&video.Duration, &video.Resolution, &video.Format, &video.Size, &video.ProcessingState,
		&video.CreatedAt, &video.UpdatedAt, &video.DeletedAt,
		&video.MatchID, &video.MatchDate, &video.HomeTeam, &video.AwayTeam, &video.Competition, &video.Season,
		&video.HasTrackingData, &video.TrackingPath,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("video not found")
		}
		return nil, err
	}

	return &video, nil
}

/**
 * FindAll retrieves a paginated list of videos.
 *
 * @param limit Maximum number of videos to return
 * @param offset Number of videos to skip
 * @return A slice of videos or an error
 */
func (r *PostgresVideoRepository) FindAll(limit, offset int) ([]*Video, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT id, title, description, file_path, storage_provider,
			   duration, resolution, format, size, processing_state,
			   created_at, updated_at, deleted_at,
			   match_id, match_date, home_team, away_team, competition, season,
			   has_tracking_data, tracking_path
		FROM videos
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*Video
	for rows.Next() {
		var video Video
		err := rows.Scan(
			&video.ID, &video.Title, &video.Description, &video.FilePath, &video.StorageProvider,
			&video.Duration, &video.Resolution, &video.Format, &video.Size, &video.ProcessingState,
			&video.CreatedAt, &video.UpdatedAt, &video.DeletedAt,
			&video.MatchID, &video.MatchDate, &video.HomeTeam, &video.AwayTeam, &video.Competition, &video.Season,
			&video.HasTrackingData, &video.TrackingPath,
		)

		if err != nil {
			return nil, err
		}

		videos = append(videos, &video)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return videos, nil
}

// Additional repository methods would be implemented here
// Create, Update, Delete, FindByMatchID, FindByTeam, FindByDateRange, FindByProcessingState