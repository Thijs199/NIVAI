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
	"strings"
	"testing"

	"nivai/backend/pkg/controllers" // Adjust if necessary
	"nivai/backend/pkg/services"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"   // For mocking services
	"github.com/stretchr/testify/require"
)

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

// MockWriteCloser is a helper for mocking io.WriteCloser for storage.Create
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

// --- Mock Python API for /process-match ---
func mockPythonProcessMatchApi(t *testing.T, expectedMatchID string, expectedTrackingPath string, expectedEventPath string) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock Python /process-match API received request: %s %s", r.Method, r.URL.Path)
		assert.Equal(t, "/process-match", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		var body map[string]string
		err := json.NewDecoder(r.Body).Decode(&body)
		require.NoError(t, err)

		assert.Equal(t, expectedMatchID, body["match_id"])
		// Path comparison can be tricky if absolute paths vs relative are involved.
		// For now, direct string comparison.
		assert.Equal(t, expectedTrackingPath, body["tracking_data_path"])
		assert.Equal(t, expectedEventPath, body["event_file_path"])


		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted) // Python API might return 202
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Processing started by mock", "match_id": expectedMatchID,
		})
	}))
	return server
}


func TestUploadVideo(t *testing.T) {
	mockStorageSvc := new(MockStorageService)
	// VideoService is also used internally by VideoController, but its methods
	// like SaveVideoMetadata might not be directly called if the UploadVideo focuses on file ops
	// and then calls Python API. The current UploadVideo in controller calls SaveVideoMetadata.
	// So, we need MockVideoService as well.

	// The VideoController's NewVideoController creates its own VideoService.
	// To test VideoController with a mock VideoService, VideoController would need to accept VideoService as a param.
	// Current NewVideoController(storageService) means VideoService is not directly mockable unless StorageService is.
	// Let's assume we can test by mocking StorageService and verifying interactions.
	// If SaveVideoMetadata is called, we'd need a way to inject mockVideoSvc.
	// The current controller uses a videoService field initialized in NewVideoController.
	// For this test, we will re-initialize the controller with both mocks.
	// This requires changing NewVideoController or having a test-specific initializer.
	// Let's assume `NewVideoController(videoService, storageService)` for testability.
	// If not, we can only mock StorageService.
	// The provided controller code: NewVideoController(storage services.StorageService) *VideoController
	// It creates its own VideoService. This means we can't easily mock VideoService calls like SaveVideoMetadata.
	// We can only mock StorageService.
	// This is a limitation. I will proceed by mocking StorageService.
	// Calls to videoService.SaveVideoMetadata will be real calls to the actual service,
	// which might interact with the mock StorageService if designed that way.
	//
	// **Revised approach given controller structure:**
	// Mock StorageService. VideoService will use this mock.
	// We cannot directly mock VideoService.SaveVideoMetadata without altering NewVideoController.
	// So, we test the effects of SaveVideoMetadata (e.g. if it tries to access storage).
	// The current `videoService.SaveVideoMetadata` in `services/video_service.go` is a placeholder.
	// It doesn't interact with storage. So, we can't verify much about it via storage mock.
	// We will primarily test file saving and Python API call.

	videoController := controllers.NewVideoController(nil, mockStorageSvc) // Original constructor

	router := mux.NewRouter() // Needed if any part of the handler relies on mux features
	router.HandleFunc("/api/v1/videos", videoController.UploadVideo).Methods("POST")


	t.Run("Successful upload of all files", func(t *testing.T) {
		videoID := "" // Will be captured from storage path mock

		// Prepare multipart form data
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)

		// Add title
		writer.WriteField("title", "Test Match Title")

		// Add video_file (optional, but let's include it)
		videoPart, _ := writer.CreateFormFile("video_file", "test_video.mp4")
		videoPart.Write([]byte("dummy video content"))

		// Add tracking_file (required)
		trackingPart, _ := writer.CreateFormFile("tracking_file", "test_tracking.gzip")
		trackingPart.Write([]byte("dummy tracking content"))

		// Add event_file (required)
		eventPart, _ := writer.CreateFormFile("event_file", "test_events.gzip")
		eventPart.Write([]byte("dummy event content"))
		writer.Close()


		// --- Mock Expectations ---
		// 1. CreateDirectory
		mockStorageSvc.On("CreateDirectory", mock.AnythingOfType("string")).Return(nil).Once()

		// 2. Create for video_file, tracking_file, event_file
		//    We need to capture the generated videoID from the path argument.
		var capturedVideoPath, capturedTrackingPath, capturedEventPath string

		// Expect UploadFile for the video file
		mockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(path string) bool { return strings.Contains(path, ".mp4") })).Run(func(args mock.Arguments) {
			capturedVideoPath = args.String(1) // path is the second argument for UploadFile
			pathParts := strings.Split(filepath.ToSlash(capturedVideoPath), "/")
			videoID = pathParts[len(pathParts)-2]
		}).Return(&services.FileUploadInfo{Path: "dummy/path/video.mp4", Size: 12345}, nil).Once() // Use a fixed path for return if capturedVideoPath is not yet set

		// Expect UploadFile for the tracking file (controller uses saveUploadedFile -> UploadFile)
		mockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(path string) bool { return strings.HasSuffix(path, "_tracking.gzip") })).Run(func(args mock.Arguments) {
			capturedTrackingPath = args.String(1) // path is the second argument
		}).Return(&services.FileUploadInfo{Path: "dummy/path/tracking.gzip", Size: 123}, nil).Once()

		// Expect UploadFile for the event file (controller uses saveUploadedFile -> UploadFile)
		mockStorageSvc.On("UploadFile", mock.Anything, mock.MatchedBy(func(path string) bool { return strings.HasSuffix(path, "_events.gzip") })).Run(func(args mock.Arguments) {
			capturedEventPath = args.String(1) // path is the second argument
		}).Return(&services.FileUploadInfo{Path: "dummy/path/events.gzip", Size: 123}, nil).Once()

		// Mock Python API (will be called after files are "saved")
		// This relies on videoID being captured correctly.
		var mockApi *httptest.Server
		defer func() { if mockApi != nil { mockApi.Close() } }()


		// --- Make Request ---
		req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		// Need to wrap the execution to setup mock API after videoID is known
		// This is tricky because videoID is generated inside the handler.
		// One way: have the mock for Create (that captures videoID) also set up the Python API mock.
		// This couples mocks but might be necessary.

		// For now, let's assume we can predict videoID if it's based on something controllable,
		// or we test the Python API call part separately / with a fixed videoID for the mock.
		// The current controller generates a random UUID. So, we cannot predict it for the mock Python API setup easily.

		// **Strategy for Python API mock with dynamic videoID:**
		// The Python API mock needs to expect the `videoID` that's generated *during* the UploadVideo call.
		// We can't set up the mockPythonProcessMatchApi perfectly before the call.
		// Alternative: The mock Python API handler could be more lenient or capture the received videoID.

		// Let's make a generic Python API mock that just checks for /process-match
		// and captures the body for later assertion.
		var pythonApiCallDetails struct {
			Called bool
			Body map[string]string
		}
		pythonApiMockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pythonApiCallDetails.Called = true
			json.NewDecoder(r.Body).Decode(&pythonApiCallDetails.Body)
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{"message": "mocked processing"})
		}))
		defer pythonApiMockServer.Close()
		t.Setenv("PYTHON_API_URL", pythonApiMockServer.URL)
		// controllers.ReinitializeVideoControllerClient() // Hypothetical, if client is package-level in video_controller

		router.ServeHTTP(rr, req)

		// --- Assertions ---
		assert.Equal(t, http.StatusAccepted, rr.Code, "Response code should be 202 Accepted")

		var responseBody map[string]string
		err := json.NewDecoder(rr.Body).Decode(&responseBody)
		require.NoError(t, err)
		assert.Equal(t, "Upload received, processing initiated.", responseBody["message"])
		assert.NotEmpty(t, responseBody["video_id"], "Response should contain video_id")

		returnedVideoId := responseBody["video_id"]
		assert.Equal(t, videoID, returnedVideoId, "video_id in response should match captured/generated one")

		mockStorageSvc.AssertExpectations(t) // Verify all storage mocks were called

		// Verify Python API call
		assert.True(t, pythonApiCallDetails.Called, "Python API /process-match was not called")
		assert.Equal(t, videoID, pythonApiCallDetails.Body["match_id"])
		// Check if paths in pythonApiCallDetails.Body match captured paths (or derived from videoID)
		// This depends on whether absolute or relative paths are sent.
		// The controller sends `absTrackingPath` which is just `trackingDestPath` currently.
		// So, they should match `capturedTrackingPath` and `capturedEventPath`.
		assert.Equal(t, capturedTrackingPath, pythonApiCallDetails.Body["tracking_data_path"])
		assert.Equal(t, capturedEventPath, pythonApiCallDetails.Body["event_file_path"])
	})

	t.Run("Missing tracking file", func(t *testing.T) {
		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "Test Match Missing Tracking")
		// videoPart, _ := writer.CreateFormFile("video_file", "test_video.mp4") // Optional
		// videoPart.Write([]byte("dummy video content"))
		eventPart, _ := writer.CreateFormFile("event_file", "test_events.gzip")
		eventPart.Write([]byte("dummy event content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Tracking and event files are required")
	})

    t.Run("Storage service CreateDirectory fails", func(t *testing.T) {
        body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "Storage Fail Title")
		trackingPart, _ := writer.CreateFormFile("tracking_file", "track.gzip")
		trackingPart.Write([]byte("track"))
		eventPart, _ := writer.CreateFormFile("event_file", "event.gzip")
		eventPart.Write([]byte("event"))
		writer.Close()

        mockStorageSvc.On("CreateDirectory", mock.AnythingOfType("string")).Return(fmt.Errorf("disk full")).Once()

        req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusInternalServerError, rr.Code)
        assert.Contains(t, rr.Body.String(), "Failed to prepare storage directory")
        mockStorageSvc.AssertExpectations(t)
    })

    t.Run("Storage service Create (for file) fails", func(t *testing.T) {
        body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		writer.WriteField("title", "File Create Fail")
		trackingPart, _ := writer.CreateFormFile("tracking_file", "track.gzip")
		trackingPart.Write([]byte("track"))
		eventPart, _ := writer.CreateFormFile("event_file", "event.gzip") // This one will fail to be created
		eventPart.Write([]byte("event"))
		writer.Close()

        mockStorageSvc.On("CreateDirectory", mock.AnythingOfType("string")).Return(nil).Once()
        // Let tracking file save succeed
        mockStorageSvc.On("Create", mock.MatchedBy(func(p string) bool { return strings.HasSuffix(p, "_tracking.gzip")})).Return(&MockWriteCloser{Writer: io.Discard}, nil).Once()
        // Let event file save fail
        mockStorageSvc.On("Create", mock.MatchedBy(func(p string) bool { return strings.HasSuffix(p, "_events.gzip")})).Return(nil, fmt.Errorf("cannot create event file")).Once()
        // Expect a call to Delete for the successfully saved tracking file during cleanup
        mockStorageSvc.On("Delete", mock.MatchedBy(func(p string) bool { return strings.HasSuffix(p, "_tracking.gzip")})).Return(nil).Once()


        req := httptest.NewRequest("POST", "/api/v1/videos", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusInternalServerError, rr.Code)
        assert.Contains(t, rr.Body.String(), "cannot create event file")
        mockStorageSvc.AssertExpectations(t)
    })

	// TODO: Add tests for GetVideo, ListVideos, DeleteVideo
	// These will primarily mock VideoService methods.
	// For DeleteVideo, also mock StorageService.Delete.
}

// Note on testing UploadVideo's call to Python API:
// The videoID is generated inside UploadVideo. To make the mock Python API server
// expect a call with the correct videoID, the mock server's handler needs to be
// either very generic (just check if /process-match was called) or the videoID generation
// needs to be predictable in tests (e.g., mock uuid.NewString).
// The current "Successful upload" test uses a more generic Python API mock that captures the call details.
//
// Testability of VideoController and its VideoService:
// As noted, NewVideoController creates its own VideoService. This makes it hard to inject a
// mock VideoService for testing VideoController's interaction with VideoService methods like SaveVideoMetadata.
// The tests above primarily focus on StorageService interactions and the overall flow of UploadVideo.
// If SaveVideoMetadata (or CreateVideo) had critical logic that needed mocking (e.g., database interactions),
// the VideoController or VideoService would need refactoring for better testability (dependency injection).
// The current SaveVideoMetadata in services/video_service.go is a placeholder, so this is less critical now.
// The test for `UploadVideo` does not explicitly assert `SaveVideoMetadata` calls due to this.
// It asserts the final HTTP response and interactions with StorageService and Python API.
//
// The `videoID` capture in the "Successful upload" test is a bit fragile, relying on path structure.
// A more robust way would be to mock `uuid.New().String()` if precise ID matching is needed for mocks
// set up *before* the handler call. The current dynamic capture is okay for asserting the Python API call body.
//
// The `videoController := controllers.NewVideoController(mockStorageSvc)` line is correct
// based on the actual constructor of `VideoController`.
// The `videoService` field within `videoController` will be the *real* `VideoService`,
// but it will be initialized with the `mockStorageSvc`. So, any calls from the real `VideoService`
// to `StorageService` methods will go to `mockStorageSvc`. This is a valid way to test.
// The `SaveVideoMetadata` method in `video_service.go` is currently a TODO placeholder,
// so it doesn't do much that would need complex mocking via storage service for now.
// If it did, for example, try to read the saved file to get metadata, then the mock storage Open would be hit.
//
// The current mock `videoService.SaveVideoMetadata` is not called by `UploadVideo` because
// `UploadVideo` in `video_controller.go` does not call `c.videoService.SaveVideoMetadata(videoMetadata)`.
// It logs "Video metadata prepared..." and then calls Python API.
// This was a change made in a previous step to simplify `UploadVideo` by removing DB interaction.
// So, no need to mock `SaveVideoMetadata` for `UploadVideo` test. If this changes in `video_controller.go`,
// then the test setup for `VideoService` would become more relevant.
// The test `TestUploadVideo` has been written according to the current `UploadVideo` implementation
// which does not call `c.videoService.SaveVideoMetadata`.
//
// The `MockVideoService` defined earlier is not used in `TestUploadVideo` because `VideoController` creates its own.
// It would be used for testing other methods like `GetVideo`, `ListVideos`, `DeleteVideo`.
// I will add those tests now.

func TestGetVideo(t *testing.T) {
    mockStorageSvc := new(MockStorageService) // Not directly used by GetVideo if VideoService handles all
    videoController := controllers.NewVideoController(nil, mockStorageSvc)
    // To properly test GetVideo, VideoService needs to be mockable.
    // Assuming VideoController's videoService field could be replaced for testing, or NewVideoController took VideoService.
    // For now, this test will be limited as videoService is internal.
    // This highlights the need for dependency injection for services into controllers.
    // If VideoService.GetVideoByID is a simple pass-through or has no external calls, it might be okay.
    // But if it hits a DB, this test is not a unit test.
    // Let's assume for a moment we *could* inject a mock VideoService for other methods.
    // However, sticking to the current structure of NewVideoController:
    // We can't mock videoService.GetVideoByID directly.
    // This test is therefore more of an integration test for GetVideo with the real VideoService
    // (which itself might be minimal if it's just a placeholder).
    // The current VideoService.GetVideoByID is a placeholder returning ErrVideoNotFound.

    router := mux.NewRouter()
    router.HandleFunc("/videos/{id}", videoController.GetVideo)

    t.Run("GetVideo not found", func(t *testing.T) {
        // The real VideoService.GetVideoByID is a placeholder and returns ErrVideoNotFound.
        req := httptest.NewRequest("GET", "/videos/nonexistent", nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusNotFound, rr.Code)
        assert.Contains(t, rr.Body.String(), "Video not found")
    })
    // To test a "found" case, data would need to be seeded if using a real DB,
    // or VideoService would need to be properly mocked and injected.
}

// Add similar placeholder tests for ListVideos and DeleteVideo,
// acknowledging the same limitations about mocking the internal VideoService.
// For DeleteVideo, StorageService.Delete will be called by the real VideoService.

func TestDeleteVideo(t *testing.T) {
    mockStorageSvc := new(MockStorageService)
    videoController := controllers.NewVideoController(nil, mockStorageSvc)
    router := mux.NewRouter()
    router.HandleFunc("/videos/{id}", videoController.DeleteVideo)

    // To test DeleteVideo, GetVideoByID is called first.
    // Current GetVideoByID is a placeholder returning ErrVideoNotFound.
    // So, DeleteVideo will likely return 404 unless VideoService is developed.

    t.Run("DeleteVideo not found due to GetVideoByID placeholder", func(t *testing.T) {
        mockStorageSvc.On("Delete", mock.AnythingOfType("string")).Return(nil).Maybe() // May or may not be called if GetVideoByID fails

        req := httptest.NewRequest("DELETE", "/videos/anyid", nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusNotFound, rr.Code) // Because GetVideoByID will say not found
        // mockStorageSvc.AssertNotCalled(t, "Delete", mock.AnythingOfType("string")) // GetVideoByID fails first
    })

    // A more complete test for DeleteVideo would require:
    // 1. VideoService.GetVideoByID to return an actual video object (needs mocking or data seeding).
    // 2. VideoService.DeleteVideo to be mockable or testable.
    // 3. StorageService.Delete mock to be verified.
}
