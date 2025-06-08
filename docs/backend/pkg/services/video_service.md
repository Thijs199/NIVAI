# Video Service Documentation

> This document describes the video service that manages video uploads, processing, and streaming capabilities in the NIVAI application.

## Architecture

```mermaid
classDiagram
    class VideoService {
        <<interface>>
        +GetVideoByID(id) Video
        +ListVideos(limit, offset, filters) Video[]
        +UploadVideo(file, header, metadata) Video
        +DeleteVideo(id) error
        +GetVideoStreamURL(id) string
        +ProcessVideo(id) error
    }

    class DefaultVideoService {
        -VideoRepository videoRepo
        -StorageService storageService
        +NewVideoService(repo, storage) VideoService
    }

    class StorageService {
        <<interface>>
    }

    class VideoRepository {
        <<interface>>
    }

    class Video {
        +String ID
        +String Title
        +String FilePath
        +String StorageProvider
        +Int64 Size
        +String Format
        +String ProcessingState
        +Time CreatedAt
        +Time UpdatedAt
    }

    VideoService <|.. DefaultVideoService
    DefaultVideoService --> StorageService : uses
    DefaultVideoService --> VideoRepository : uses
    DefaultVideoService ..> Video : manages

    class ProcessingStates {
        <<enumeration>>
        pending
        processing
        completed
        failed
    }

    Video --> ProcessingStates : has state
```

## Components

### Video Service Interface

Provides high-level video operations:

- Video retrieval and listing
- Upload management
- Streaming URL generation
- Processing coordination

### Default Implementation

Features:

- File validation
- Metadata management
- Storage coordination
- Processing state tracking

## Video Processing Flow

```mermaid
sequenceDiagram
    participant Client
    participant VideoService
    participant Storage
    participant Repository
    participant Processor

    Client->>VideoService: Upload Video
    VideoService->>VideoService: Validate File
    VideoService->>Storage: Store File
    VideoService->>Repository: Save Metadata
    VideoService->>Processor: Queue Processing
    Processor->>Repository: Update State (processing)
    Processor-->>Repository: Update Metadata
    Processor->>Repository: Update State (completed)
```

## File Organization

### Storage Path Structure

```
videos/
├── matches/
│   └── {match_id}/
│       └── {YYYY/MM/DD}/
│           └── {video_id}.{ext}
└── uploads/
    └── {YYYY/MM/DD}/
        └── {video_id}.{ext}
```

## Supported Formats

Video formats supported:

- MP4 (.mp4)
- QuickTime (.mov)
- AVI (.avi)
- Matroska (.mkv)
- WebM (.webm)

## Error Handling

### Common Errors

- `ErrVideoNotFound`: Video not in repository
- `ErrInvalidVideo`: Invalid video data/format
- `ErrStorageFailed`: Storage operation failure

## Configuration

### Processing Settings

- Default page size: 10 videos
- Processing states: pending, processing, completed
- File organization: date-based hierarchy

## Usage Examples

```go
// Initialize service
videoService := NewVideoService(videoRepo, storageService)

// Upload video
video, err := videoService.UploadVideo(file, header, metadata)

// Get stream URL
url, err := videoService.GetVideoStreamURL(video.ID)

// List videos with filters
filters := map[string]string{
    "match_id": "match123",
    "processing_state": "completed"
}
videos, err := videoService.ListVideos(10, 0, filters)
```

## Related Files

- `models/video.go`: Video data model
- `storage_service.go`: Storage interface
- `controllers/video_controller.go`: HTTP handlers
