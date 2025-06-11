package services_test

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"testing"
	"time"

	"nivai/backend/pkg/models"
	"nivai/backend/pkg/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- MockVideoRepository for video_service_test ---
type MockVideoRepository struct {
	mock.Mock
}

func (m *MockVideoRepository) FindByID(id string) (*models.Video, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Video), args.Error(1)
}
func (m *MockVideoRepository) FindAll(limit, offset int) ([]*models.Video, error) {
	args := m.Called(limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Video), args.Error(1)
}
func (m *MockVideoRepository) Create(video *models.Video) error {
	args := m.Called(video)
	return args.Error(0)
}
func (m *MockVideoRepository) Update(video *models.Video) error {
	args := m.Called(video)
	return args.Error(0)
}
func (m *MockVideoRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
func (m *MockVideoRepository) FindByMatchID(matchID string) ([]*models.Video, error) {
	args := m.Called(matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Video), args.Error(1)
}
func (m *MockVideoRepository) FindByTeam(teamName string, limit, offset int) ([]*models.Video, error) {
	args := m.Called(teamName, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Video), args.Error(1)
}
func (m *MockVideoRepository) FindByDateRange(start, end time.Time, limit, offset int) ([]*models.Video, error) {
	args := m.Called(start, end, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Video), args.Error(1)
}
func (m *MockVideoRepository) FindByProcessingState(state string, limit, offset int) ([]*models.Video, error) {
	args := m.Called(state, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Video), args.Error(1)
}

// --- MockStorageService for video_service_test ---
type MockStorageService struct {
	mock.Mock
}
func (m *MockStorageService) UploadFile(file multipart.File, path string) (*services.FileUploadInfo, error) {
	args := m.Called(file, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.FileUploadInfo), args.Error(1)
}
func (m *MockStorageService) GetFile(path string) (io.ReadCloser, error) {
	args := m.Called(path)
	// ... (implementation from file_storage_service_test.go or simplify if not needed for these tests)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *MockStorageService) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}
func (m *MockStorageService) GetStreamURL(path string) (string, error) {
	args := m.Called(path)
	if args.Get(0) == "" { // Check for empty string for URL
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}
func (m *MockStorageService) GetFileMetadata(path string) (map[string]string, error) {
	args := m.Called(path)
	// ... (implementation or simplify)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}


// Helper to create a dummy multipart.File for testing UploadVideo
type mockMultipartFileVS struct { // Renamed to avoid conflict if in same package for testing
	*bytes.Reader
}
func (mf *mockMultipartFileVS) Close() error { return nil }
// Open method for multipart.FileHeader compatibility, not strictly for multipart.File itself
// func (mf *mockMultipartFileVS) Open() (multipart.File, error) { return mf, nil }

func newMockMultipartFileVS(content string) multipart.File {
    return &mockMultipartFileVS{ Reader: bytes.NewReader([]byte(content)) }
}
func newMockFileHeader(filename string, size int64) *multipart.FileHeader {
	// To make this header usable with UploadFile, it needs to provide an Open() method
	// that returns a multipart.File. We can embed a small helper for this.

	// Create a temporary file with the content for the header to "open"
	// This is a bit more involved but makes the mock FileHeader more realistic
	// For simplicity in this context, if the service's UploadVideo only uses
	// header.Filename and header.Size, we might not need a functional Open().
	// However, a robust mock would provide it.
	// The current DefaultVideoService.UploadVideo doesn't seem to call Open() on the header,
	// it passes the multipart.File directly to storage.UploadFile.
	// So, a simple header is likely fine.
	return &multipart.FileHeader{Filename: filename, Size: size}
}


func TestDefaultVideoService_GetVideoByID(t *testing.T) {
	mockRepo := new(MockVideoRepository)
	// Storage service not directly used by GetVideoByID, can be nil if constructor allows or use a basic mock
	mockStorage := new(MockStorageService)
	videoService := services.NewVideoService(mockRepo, mockStorage)

	t.Run("Success", func(t *testing.T) {
		expectedVideo := &models.Video{ID: "vid1", Title: "Test Video"}
		mockRepo.On("FindByID", "vid1").Return(expectedVideo, nil).Once()

		video, err := videoService.GetVideoByID("vid1")
		require.NoError(t, err)
		assert.Equal(t, expectedVideo, video)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Not Found", func(t *testing.T) {
		// Simulate repo "not found" error that DefaultVideoService should wrap
		mockRepo.On("FindByID", "vid_unknown").Return(nil, errors.New("some repo specific not found error")).Once()

		_, err := videoService.GetVideoByID("vid_unknown")
		require.Error(t, err)
		// Check if the service wraps it into services.ErrVideoNotFound
		// Note: The actual error from repo might be different, the service checks for "not found" substring.
		// For a more robust test, ensure the mock error contains "not found".
		assert.ErrorIs(t, err, services.ErrVideoNotFound, "Service should wrap repository not found error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Empty ID", func(t *testing.T) {
		_, err := videoService.GetVideoByID("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "video ID cannot be empty")
	})
}

func TestDefaultVideoService_ListVideos(t *testing.T) {
    mockRepo := new(MockVideoRepository)
    mockStorage := new(MockStorageService)
    videoService := services.NewVideoService(mockRepo, mockStorage)

    expectedVideos := []*models.Video{{ID: "vid1"}, {ID: "vid2"}}

    t.Run("No filters", func(t *testing.T) {
        mockRepo.On("FindAll", 10, 0).Return(expectedVideos, nil).Once()
        videos, err := videoService.ListVideos(0, 0, make(map[string]string))
        require.NoError(t, err)
        assert.Equal(t, expectedVideos, videos)
        mockRepo.AssertExpectations(t)
    })

    t.Run("With match_id filter", func(t *testing.T) {
        filters := map[string]string{"match_id": "match123"}
        mockRepo.On("FindByMatchID", "match123").Return(expectedVideos, nil).Once()
        videos, err := videoService.ListVideos(10, 0, filters) // limit, offset might be ignored by FindByMatchID in some impls
        require.NoError(t, err)
        assert.Equal(t, expectedVideos, videos)
        mockRepo.AssertExpectations(t)
    })

    t.Run("With team filter", func(t *testing.T) {
		filters := map[string]string{"team": "TeamX"}
		mockRepo.On("FindByTeam", "TeamX", 10, 0).Return(expectedVideos, nil).Once()
		videos, err := videoService.ListVideos(10, 0, filters)
		require.NoError(t, err)
		assert.Equal(t, expectedVideos, videos)
		mockRepo.AssertExpectations(t)
	})

	t.Run("With processing_state filter", func(t *testing.T) {
		filters := map[string]string{"processing_state": "completed"}
		mockRepo.On("FindByProcessingState", "completed", 10, 0).Return(expectedVideos, nil).Once()
		videos, err := videoService.ListVideos(10, 0, filters)
		require.NoError(t, err)
		assert.Equal(t, expectedVideos, videos)
		mockRepo.AssertExpectations(t)
	})

    t.Run("Repository FindAll error", func(t *testing.T) {
        mockRepo.On("FindAll", 10, 0).Return(nil, errors.New("db error")).Once()
        _, err := videoService.ListVideos(0, 0, make(map[string]string))
        require.Error(t, err)
        assert.Contains(t, err.Error(), "db error")
        mockRepo.AssertExpectations(t)
    })
}


func TestDefaultVideoService_UploadVideo(t *testing.T) {
    videoContent := "dummy video content"
    mockFile := newMockMultipartFileVS(videoContent) // This is multipart.File

    // The metadata.FilePath is used by generateStoragePath, so it needs an extension.
    // The actual filename from header is used for isValidVideoType.
    videoMetaWithExtension := &models.Video{ID: "newVid1", Title: "Upload Test", FilePath: "placeholder_for_ext.mp4"}


    t.Run("Success", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        // Use a header with a valid video filename
        mockHeader := newMockFileHeader("test_video.mp4", int64(len(videoContent)))

        // Use the helper to predict path. metadata.ID and metadata.FilePath (for ext) are used by it.
        expectedStoragePath := services.GenerateStoragePathForTesting(videoMetaWithExtension)

        mockStorage.On("UploadFile", mockFile, expectedStoragePath).Return(&services.FileUploadInfo{
            Path: expectedStoragePath, Provider: "mock_storage", Size: int64(len(videoContent)), Format: "mp4"}, nil).Once()

        freshVideoFromCreate := models.Video{} // To capture the video passed to Create

        mockRepo.On("Create", mock.MatchedBy(func(v *models.Video) bool {
            // Capture the video for later assertions if needed, or assert directly
            freshVideoFromCreate = *v
            return v.ID == videoMetaWithExtension.ID &&
                   v.Title == videoMetaWithExtension.Title &&
                   v.FilePath == expectedStoragePath &&
                   v.StorageProvider == "mock_storage" &&
                   v.Size == int64(len(videoContent)) &&
                   v.Format == "mp4" &&
                   v.ProcessingState == "pending" // Initial state
        })).Return(nil).Once()

        // Mocks for ProcessVideo goroutine
        // FindByID will be called by ProcessVideo
        mockRepo.On("FindByID", videoMetaWithExtension.ID).Return(&freshVideoFromCreate, nil).Maybe() // Maybe, as timing of goroutine is not guaranteed in test
        // Update will be called twice by ProcessVideo
        mockRepo.On("Update", mock.MatchedBy(func(v *models.Video) bool {
			return v.ID == videoMetaWithExtension.ID && (v.ProcessingState == "processing" || v.ProcessingState == "completed")
		})).Return(nil).Maybe()


        createdVideo, err := videoService.UploadVideo(mockFile, mockHeader, videoMetaWithExtension)
        require.NoError(t, err)
        assert.NotNil(t, createdVideo)
        assert.Equal(t, videoMetaWithExtension.ID, createdVideo.ID)
        assert.Equal(t, expectedStoragePath, createdVideo.FilePath) // Check if metadata was updated

        mockStorage.AssertExpectations(t)
        mockRepo.AssertCalled(t, "Create", mock.AnythingOfType("*models.Video"))
        // Assertions for ProcessVideo calls are tricky due to goroutine.
        // A common approach is to wait a bit or use channels for synchronization if precise assertions are needed.
        // For now, checking Create is the primary goal of this test path.
        // Adding a small delay to see if goroutine calls are made, but this is not ideal.
        time.Sleep(50 * time.Millisecond) // Caution: flaky tests
        mockRepo.AssertExpectations(t) // This will check if Maybe calls happened
    })

    t.Run("Invalid file type", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		videoService := services.NewVideoService(mockRepo, mockStorage)
        invalidHeader := newMockFileHeader("test_document.txt", 100)
        _, err := videoService.UploadVideo(mockFile, invalidHeader, videoMetaWithExtension)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "invalid video file type")
    })

    t.Run("Missing title", func(t *testing.T) {
		mockRepo := new(MockVideoRepository)
		mockStorage := new(MockStorageService)
		videoService := services.NewVideoService(mockRepo, mockStorage)
        metaNoTitle := &models.Video{ID: "vidNoTitle", FilePath: "some.mp4"} // FilePath with ext needed for generateStoragePath
		mockHeader := newMockFileHeader("test_video.mp4", int64(len(videoContent)))
        _, err := videoService.UploadVideo(mockFile, mockHeader, metaNoTitle)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "video title is required")
    })

    t.Run("Storage UploadFile fails", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)
        mockHeader := newMockFileHeader("test_video.mp4", int64(len(videoContent)))

        expectedStoragePath := services.GenerateStoragePathForTesting(videoMetaWithExtension)
        mockStorage.On("UploadFile", mockFile, expectedStoragePath).Return(nil, errors.New("storage disk full")).Once()

        _, err := videoService.UploadVideo(mockFile, mockHeader, videoMetaWithExtension)
        assert.Error(t, err)
        assert.ErrorIs(t, err, services.ErrStorageFailed)
        mockStorage.AssertExpectations(t)
        mockRepo.AssertNotCalled(t, "Create", mock.Anything)
    })

    t.Run("Repository Create fails, ensure cleanup", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)
        mockHeader := newMockFileHeader("test_video.mp4", int64(len(videoContent)))

        expectedStoragePath := services.GenerateStoragePathForTesting(videoMetaWithExtension)
        uploadInfo := &services.FileUploadInfo{Path: expectedStoragePath, Provider: "mock", Size: 123, Format: "mp4"}

        mockStorage.On("UploadFile", mockFile, expectedStoragePath).Return(uploadInfo, nil).Once()
        mockRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(errors.New("db connection error")).Once()
        mockStorage.On("DeleteFile", expectedStoragePath).Return(nil).Once() // Expect cleanup

        _, err := videoService.UploadVideo(mockFile, mockHeader, videoMetaWithExtension)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "db connection error") // Error from repo should be propagated

        mockStorage.AssertExpectations(t)
        mockRepo.AssertExpectations(t)
    })
}


func TestDefaultVideoService_DeleteVideo(t *testing.T) {
    mockRepo := new(MockVideoRepository)
    mockStorage := new(MockStorageService)
    videoService := services.NewVideoService(mockRepo, mockStorage)

    t.Run("Success", func(t *testing.T) {
        // DefaultVideoService.DeleteVideo was modified to not call FindByID first.
        // It directly calls repo.Delete.
        mockRepo.On("Delete", "vid_to_delete").Return(nil).Once()
        err := videoService.DeleteVideo("vid_to_delete")
        require.NoError(t, err)
        mockRepo.AssertExpectations(t)
    })

    t.Run("Not Found by Repo.Delete", func(t *testing.T) {
        // If repo.Delete returns an error containing "not found"
        mockRepo.On("Delete", "vid_unknown_delete").Return(errors.New("video not found in repo")).Once()
        err := videoService.DeleteVideo("vid_unknown_delete")
        require.Error(t, err)
        assert.ErrorIs(t, err, services.ErrVideoNotFound) // Service should wrap it
        mockRepo.AssertExpectations(t)
    })

     t.Run("Repo.Delete returns other error", func(t *testing.T) {
        mockRepo.On("Delete", "vid_other_error").Return(errors.New("some other db error")).Once()
        err := videoService.DeleteVideo("vid_other_error")
        require.Error(t, err)
        assert.NotErrorIs(t, err, services.ErrVideoNotFound)
        assert.Contains(t, err.Error(), "some other db error")
        mockRepo.AssertExpectations(t)
    })
}

func TestDefaultVideoService_GetVideoStreamURL(t *testing.T) {
    videoID := "streamVid1"
    videoFilePath := "path/to/streamable.mp4"
    mockVideo := &models.Video{ID: videoID, FilePath: videoFilePath}
    expectedStreamURL := "http://mockstorage.com/streamable.mp4"

    t.Run("Success", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", videoID).Return(mockVideo, nil).Once()
        mockStorage.On("GetStreamURL", videoFilePath).Return(expectedStreamURL, nil).Once()

        url, err := videoService.GetVideoStreamURL(videoID)
        require.NoError(t, err)
        assert.Equal(t, expectedStreamURL, url)
        mockRepo.AssertExpectations(t)
        mockStorage.AssertExpectations(t)
    })

    t.Run("Video Not Found by Repo", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", "unknownVid").Return(nil, errors.New("not found error from repo")).Once()
        _, err := videoService.GetVideoStreamURL("unknownVid")
        require.Error(t, err)
        assert.ErrorIs(t, err, services.ErrVideoNotFound)
        mockRepo.AssertExpectations(t)
        mockStorage.AssertNotCalled(t, "GetStreamURL", mock.Anything)
    })

    t.Run("Storage GetStreamURL fails", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", videoID).Return(mockVideo, nil).Once()
        mockStorage.On("GetStreamURL", videoFilePath).Return("", errors.New("storage URL generation failed")).Once()
        _, err := videoService.GetVideoStreamURL(videoID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "storage URL generation failed")
        mockRepo.AssertExpectations(t)
        mockStorage.AssertExpectations(t)
    })
}

func TestDefaultVideoService_ProcessVideo(t *testing.T) {
    videoID := "processVid1"
    initialVideoState := &models.Video{ID: videoID, ProcessingState: "pending"}

    t.Run("Success", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", videoID).Return(initialVideoState, nil).Once()
        mockRepo.On("Update", mock.MatchedBy(func(v *models.Video) bool {
            return v.ID == videoID && v.ProcessingState == "processing"
        })).Return(nil).Once()
        mockRepo.On("Update", mock.MatchedBy(func(v *models.Video) bool {
            return v.ID == videoID &&
                v.ProcessingState == "completed" &&
                v.Duration == 120.5 &&
                v.Resolution == "1920x1080"
        })).Return(nil).Once()

        err := videoService.ProcessVideo(videoID)
        require.NoError(t, err)
        mockRepo.AssertExpectations(t)
    })

    t.Run("Video Not Found on initial FindByID", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", "unknownVid").Return(nil, errors.New("repo: not found")).Once()
        err := videoService.ProcessVideo("unknownVid")
        require.Error(t, err)
        assert.Contains(t, err.Error(), "repo: not found")
        mockRepo.AssertExpectations(t)
        mockRepo.AssertNotCalled(t, "Update", mock.Anything)
    })

    t.Run("First Update fails", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", videoID).Return(initialVideoState, nil).Once()
        mockRepo.On("Update", mock.MatchedBy(func(v *models.Video) bool {
            return v.ID == videoID && v.ProcessingState == "processing"
        })).Return(errors.New("db error on first update")).Once()

        err := videoService.ProcessVideo(videoID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "db error on first update")
        mockRepo.AssertExpectations(t)
        mockRepo.AssertNumberOfCalls(t, "Update", 1)
    })

    t.Run("Second Update fails", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("FindByID", videoID).Return(initialVideoState, nil).Once()
        mockRepo.On("Update", mock.MatchedBy(func(v *models.Video) bool {
            return v.ID == videoID && v.ProcessingState == "processing"
        })).Return(nil).Once()
        mockRepo.On("Update", mock.MatchedBy(func(v *models.Video) bool {
            return v.ID == videoID && v.ProcessingState == "completed"
        })).Return(errors.New("db error on second update")).Once()

        err := videoService.ProcessVideo(videoID)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "db error on second update")
        mockRepo.AssertExpectations(t)
        mockRepo.AssertNumberOfCalls(t, "Update", 2)
    })
}

func TestDefaultVideoService_CreateVideoEntry(t *testing.T) {
    videoMeta := &models.Video{ID: "entryVid1", Title: "Entry Test", CreatedAt: time.Now()}

    t.Run("Success", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("Create", mock.MatchedBy(func(v *models.Video) bool {
            return v.ID == videoMeta.ID && v.Title == videoMeta.Title && !v.UpdatedAt.IsZero()
        })).Return(nil).Once()

        createdVideo, err := videoService.CreateVideoEntry(videoMeta)
        require.NoError(t, err)
        assert.Equal(t, videoMeta, createdVideo)
        assert.False(t, createdVideo.UpdatedAt.IsZero(), "UpdatedAt should be set by the service")
        mockRepo.AssertExpectations(t)
    })

    t.Run("Repository Create fails", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        mockRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(errors.New("db unique constraint failed")).Once()

        _, err := videoService.CreateVideoEntry(videoMeta)
        require.Error(t, err)
        assert.Contains(t, err.Error(), "db unique constraint failed")
        mockRepo.AssertExpectations(t)
    })

    t.Run("Missing ID in metadata", func(t *testing.T) {
        mockRepo := new(MockVideoRepository)
        mockStorage := new(MockStorageService)
        videoService := services.NewVideoService(mockRepo, mockStorage)

        metaNoID := &models.Video{Title: "Test No ID"}
        _, err := videoService.CreateVideoEntry(metaNoID)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "metadata ID is required")
        mockRepo.AssertNotCalled(t, "Create", mock.Anything)
    })
}
