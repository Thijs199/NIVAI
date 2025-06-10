package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"errors"
	"strings"
	"testing"
	"time"

	"nivai/backend/pkg/controllers"
	"nivai/backend/pkg/models"
	"nivai/backend/pkg/services"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// --- MockVideoRepository ---
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

// --- Mock StorageService ---
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) CreateDirectory(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockStorageService) Create(path string) (io.WriteCloser, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.WriteCloser), args.Error(1)
}

func (m *MockStorageService) Open(path string) (io.ReadCloser, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorageService) Delete(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockStorageService) DeleteFile(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *MockStorageService) GetFile(path string) (io.ReadCloser, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorageService) GetFileMetadata(path string) (map[string]string, error) {
	args := m.Called(path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockStorageService) GetStreamURL(path string) (string, error) {
	args := m.Called(path)
	return args.String(0), args.Error(1)
}

func (m *MockStorageService) UploadFile(file multipart.File, path string) (*services.FileUploadInfo, error) {
	args := m.Called(file, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.FileUploadInfo), args.Error(1)
}

// MockWriteCloser
type MockWriteCloser struct {
	io.Writer
	closeFunc func() error
}

func (mwc *MockWriteCloser) Close() error {
	if mwc.closeFunc != nil {
		return mwc.closeFunc()
	}
	return nil
}

// mockPythonProcessMatchApi
func mockPythonProcessMatchApi(t *testing.T, expectedMatchID string, expectedTrackingPath string, expectedEventPath string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/process-match", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)
		assert.Equal(t, expectedMatchID, body["match_id"])
		assert.Equal(t, expectedTrackingPath, body["tracking_data_path"])
		assert.Equal(t, expectedEventPath, body["event_data_path"]) // Corrected key
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Processing started by mock", "match_id": expectedMatchID})
	}))
	return server
}

func TestUploadVideo(t *testing.T) {
	t.Run("Successful upload of all files", func(t *testing.T) {
		mockVideoRepo := new(MockVideoRepository)
		mockStorageSvc := new(MockStorageService)
		videoService := services.NewVideoService(mockVideoRepo, mockStorageSvc)

		var videoID string

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "Test Match Title")
		videoFilePart, _ := writer.CreateFormFile("video_file", "test_video.mp4")
		videoFilePart.Write([]byte("dummy video content"))
		trackingFilePart, _ := writer.CreateFormFile("tracking_file", "test_tracking.gzip")
		trackingFilePart.Write([]byte("dummy tracking content"))
		eventFilePart, _ := writer.CreateFormFile("event_file", "test_events.gzip")
		eventFilePart.Write([]byte("dummy event content"))
		writer.Close()

		expectedVideoPath := "videos/vid123/video.mp4"
		expectedTrackingPath := "videos/vid123/tracking.gzip"
		expectedEventPath := "videos/vid123/events.gzip"

		mockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(path string) bool { return strings.Contains(path, ".mp4") })).Run(func(args mock.Arguments) {
			p := args.String(1)
			pathParts := strings.Split(filepath.ToSlash(p), "/")
			if len(pathParts) >= 2 { videoID = pathParts[len(pathParts)-2] }
		}).Return(&services.FileUploadInfo{Path: expectedVideoPath, Size: 12345}, nil).Once()

		mockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(path string) bool { return strings.HasSuffix(path, "_tracking.gzip") })).Return(&services.FileUploadInfo{Path: expectedTrackingPath, Size: 123}, nil).Once()

		mockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(path string) bool { return strings.HasSuffix(path, "_events.gzip") })).Return(&services.FileUploadInfo{Path: expectedEventPath, Size: 123}, nil).Once()

		mockVideoRepo.On("Create", mock.AnythingOfType("*models.Video")).Return(nil).Once()

		var pythonApiCallDetails struct { Called bool; Body map[string]string }
		pythonApiMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bodyBytes, _ := io.ReadAll(r.Body)
			r.Body.Close() // Important: close body after reading
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Replace body for decoder
			t.Logf("Mock Python API received body: %s", string(bodyBytes))

			pythonApiCallDetails.Body = make(map[string]string) // Clear map
			if err := json.NewDecoder(bytes.NewBuffer(bodyBytes)).Decode(&pythonApiCallDetails.Body); err != nil {
				t.Logf("Mock Python API: Error decoding request body: %v", err)
				http.Error(w, "bad request body", http.StatusBadRequest)
				return
			}
			pythonApiCallDetails.Called = true
			t.Logf("Mock Python API decoded body: %+v", pythonApiCallDetails.Body)
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{"message": "mocked processing"})
		}))
		defer pythonApiMockServer.Close()

		videoController := controllers.NewVideoController(videoService, mockStorageSvc, pythonApiMockServer.URL, pythonApiMockServer.Client())

		req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		testRouter := mux.NewRouter()
		testRouter.HandleFunc("/api/v1/videos", videoController.UploadVideo).Methods("POST")
		testRouter.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusAccepted, rr.Code)
		var responseBody map[string]string
		errJSON := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, errJSON)
		assert.Equal(t, "Upload received, processing initiated.", responseBody["message"])
		assert.NotEmpty(t, responseBody["video_id"])

		if videoID != "" {
		    assert.Equal(t, videoID, responseBody["video_id"])
		}

		mockStorageSvc.AssertExpectations(t)
		mockVideoRepo.AssertExpectations(t)

		assert.True(t, pythonApiCallDetails.Called, "Python API /process-match was not called")
		if videoID != "" {
			assert.Equal(t, videoID, pythonApiCallDetails.Body["match_id"])
		} else {
			assert.NotEmpty(t, pythonApiCallDetails.Body["match_id"], "match_id in Python API call should not be empty")
		}
		assert.Equal(t, expectedTrackingPath, pythonApiCallDetails.Body["tracking_data_path"])
		assert.Equal(t, expectedEventPath, pythonApiCallDetails.Body["event_data_path"])
	})

	t.Run("Missing tracking file", func(t *testing.T) {
		localMockVideoRepo := new(MockVideoRepository)
		localMockStorageSvc := new(MockStorageService)
		localVideoService := services.NewVideoService(localMockVideoRepo, localMockStorageSvc)
		localVideoController := controllers.NewVideoController(localVideoService, localMockStorageSvc, "", nil)
		localRouter := mux.NewRouter()
		localRouter.HandleFunc("/api/v1/videos", localVideoController.UploadVideo).Methods("POST")

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "Test Match Missing Tracking")
		eventPart, _ := writer.CreateFormFile("event_file", "test_events.gzip")
		eventPart.Write([]byte("dummy event content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		localRouter.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Tracking and event files are required")
	})

    t.Run("Storage service Create (for file) fails", func(t *testing.T) {
		localMockVideoRepo := new(MockVideoRepository)
		localMockStorageSvc := new(MockStorageService)
		localVideoService := services.NewVideoService(localMockVideoRepo, localMockStorageSvc)
		localVideoController := controllers.NewVideoController(localVideoService, localMockStorageSvc, "", nil)
		localRouter := mux.NewRouter()
		localRouter.HandleFunc("/api/v1/videos", localVideoController.UploadVideo).Methods("POST")

        body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "File Create Fail")
		trackingPart, _ := writer.CreateFormFile("tracking_file", "track.gzip")
		trackingPart.Write([]byte("track"))
		eventPart, _ := writer.CreateFormFile("event_file", "event.gzip")
		eventPart.Write([]byte("event"))
		videoFilePart, _ := writer.CreateFormFile("video_file", "video.mp4")
		videoFilePart.Write([]byte("dummy video"))
		writer.Close()

        localMockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(p string) bool { return strings.Contains(p, ".mp4")})).Return(&services.FileUploadInfo{Path: "path/to/video.mp4"}, nil).Once()
        localMockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(p string) bool { return strings.HasSuffix(p, "_tracking.gzip")})).Return(&services.FileUploadInfo{Path: "path/to/tracking.gzip"}, nil).Once()
        localMockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(p string) bool { return strings.HasSuffix(p, "_events.gzip")})).Return(nil, fmt.Errorf("cannot create event file")).Once()

        localMockStorageSvc.On("DeleteFile", "path/to/video.mp4").Return(nil).Once()
        localMockStorageSvc.On("DeleteFile", "path/to/tracking.gzip").Return(nil).Once()


        req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		localRouter.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusInternalServerError, rr.Code)
        assert.Contains(t, rr.Body.String(), "cannot create event file")
        localMockStorageSvc.AssertExpectations(t)
    })
}

func TestGetVideo(t *testing.T) {
    mockVideoRepo := new(MockVideoRepository)
    mockStorageSvc := new(MockStorageService)
    videoService := services.NewVideoService(mockVideoRepo, mockStorageSvc)
    videoController := controllers.NewVideoController(videoService, mockStorageSvc, "", nil)

    router := mux.NewRouter()
    router.HandleFunc("/videos/{id}", videoController.GetVideo)

    t.Run("GetVideo not found", func(t *testing.T) {
		mockVideoRepo.On("FindByID", "nonexistent").Return(nil, errors.New("video not found")).Once()

        req := httptest.NewRequest("GET", "/videos/nonexistent", nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusNotFound, rr.Code)
        assert.Contains(t, rr.Body.String(), "Video not found")
		mockVideoRepo.AssertExpectations(t)
    })
}

func TestDeleteVideo(t *testing.T) {
    mockVideoRepo := new(MockVideoRepository)
    mockStorageSvc := new(MockStorageService)
    videoService := services.NewVideoService(mockVideoRepo, mockStorageSvc)
    videoController := controllers.NewVideoController(videoService, mockStorageSvc, "", nil)

    router := mux.NewRouter()
    router.HandleFunc("/videos/{id}", videoController.DeleteVideo)

    t.Run("DeleteVideo not found", func(t *testing.T) {
        mockVideoRepo.On("FindByID", "anyid").Return(nil, errors.New("video not found")).Once()

        req := httptest.NewRequest("DELETE", "/videos/anyid", nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusNotFound, rr.Code)
		mockVideoRepo.AssertExpectations(t)
    })

	t.Run("DeleteVideo successful", func(t *testing.T) {
		videoID := "existingID"
		mockVideo := &models.Video{ID: videoID, FilePath: "videos/some/path/video.mp4", TrackingPath: "videos/some/path/tracking.gzip", EventFilePath: "videos/some/path/events.gzip"}

		mockVideoRepo.On("FindByID", videoID).Return(mockVideo, nil).Twice() // Expect FindByID to be called twice
		mockVideoRepo.On("Delete", videoID).Return(nil).Once()

		mockStorageSvc.On("DeleteFile", mockVideo.FilePath).Return(nil).Once()
		mockStorageSvc.On("DeleteFile", mockVideo.TrackingPath).Return(nil).Once()
		mockStorageSvc.On("DeleteFile", mockVideo.EventFilePath).Return(nil).Once()


		req := httptest.NewRequest("DELETE", "/videos/"+videoID, nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code) // Corrected expected status code
		mockVideoRepo.AssertExpectations(t)
		mockStorageSvc.AssertExpectations(t)
	})
}
// End of video_controller_test.go
