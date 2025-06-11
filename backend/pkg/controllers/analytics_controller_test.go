package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"nivai/backend/pkg/controllers" // Adjust import path
	// Assuming the actual analytics_controller.go initializes its own pythonApiBaseUrl and netClient
	// If not, and they are package level, this test might interfere or need to use those.
	// The current analytics_controller.go uses an init() for its client, so tests will use that.
	// For more control, pythonApiBaseUrl should be configurable in the controller instance.

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPythonApi serves as a mock Python API for analytics endpoints
func mockPythonApi(t *testing.T, expectedPathPrefix string, responseBody map[string]interface{}, statusCode int) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("Mock Python API received request: %s %s", r.Method, r.URL.Path)
		if !strings.HasPrefix(r.URL.Path, expectedPathPrefix) {
			t.Errorf("Mock Python API expected path prefix %s, got %s", expectedPathPrefix, r.URL.Path)
			http.Error(w, "Unexpected path", http.StatusInternalServerError)
			return
		}
		// Optional: check query params if needed for player/team analytics specific tests

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		err := json.NewEncoder(w).Encode(responseBody)
		require.NoError(t, err, "Failed to encode mock python api response")
	}))
	return server
}

func TestGetMatchAnalytics(t *testing.T) {
	t.Run("Successful data relay", func(t *testing.T) {
		matchID := "testmatch123"
		expectedResponse := map[string]interface{}{"data": "match_summary_data", "id": matchID}
		mockApi := mockPythonApi(t, fmt.Sprintf("/match/%s/stats/summary", matchID), expectedResponse, http.StatusOK)
		defer mockApi.Close()

		ac := controllers.NewAnalyticsController(mockApi.URL, mockApi.Client())
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/analytics/matches/{id}", ac.GetMatchAnalytics).Methods("GET")

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/analytics/matches/%s", matchID), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var actualResponse map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&actualResponse)
		require.NoError(t, err)
		assert.Equal(t, expectedResponse, actualResponse)
	})

	t.Run("Python API returns 404", func(t *testing.T) {
		matchID := "notfoundmatch"
		errorResponse := map[string]interface{}{"detail": "match not found in python api"}
		mockApi := mockPythonApi(t, fmt.Sprintf("/match/%s/stats/summary", matchID), errorResponse, http.StatusNotFound)
		defer mockApi.Close()

		ac := controllers.NewAnalyticsController(mockApi.URL, mockApi.Client())
		router := mux.NewRouter()
		router.HandleFunc("/api/v1/analytics/matches/{id}", ac.GetMatchAnalytics).Methods("GET")

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/analytics/matches/%s", matchID), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code) // Should relay the 404
		var actualResponse map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&actualResponse)
		require.NoError(t, err)
		assert.Equal(t, errorResponse, actualResponse)
	})

	t.Run("Python API unavailable", func(t *testing.T) {
		matchID := "api_down_match"
		// Mock server is started but immediately closed to simulate unavailability
		mockApi := mockPythonApi(t, "", nil, http.StatusOK)
		mockApi.Close() // Simulate server down

		ac := controllers.NewAnalyticsController(mockApi.URL, nil) // Use nil client, it should default
		// For this specific test, we can use a local router or call the method directly if no mux vars are needed by the handler itself
		// Given GetMatchAnalytics uses mux.Vars, a router is needed.
		localRouter := mux.NewRouter()
		localRouter.HandleFunc("/api/v1/analytics/matches/{id}", ac.GetMatchAnalytics).Methods("GET")

		req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/analytics/matches/%s", matchID), nil)
		rr := httptest.NewRecorder()
		localRouter.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadGateway, rr.Code)
		// Check for plain text error message
		responseBody := rr.Body.String()
		assert.Contains(t, responseBody, "Error connecting to analytics service")
	})

	t.Run("Missing match_id in path", func(t *testing.T){
		// This test primarily tests mux routing.
		// We need an AnalyticsController instance to register its methods.
		ac := controllers.NewAnalyticsController("", nil) // URL/client don't matter as it shouldn't be called
		testRouter := mux.NewRouter()
		testRouter.HandleFunc("/api/v1/analytics/matches/{id}", ac.GetMatchAnalytics).Methods("GET")

		nonMatchingReq := httptest.NewRequest("GET", "/api/v1/analytics/matches/", nil) // No ID
		nonMatchingRr := httptest.NewRecorder()
		testRouter.ServeHTTP(nonMatchingRr, nonMatchingReq)
		assert.Equal(t, http.StatusNotFound, nonMatchingRr.Code) // Mux should 404 this
	})
}


// Similar tests for GetPlayerAnalytics and GetTeamAnalytics
// Need to handle query parameters in these tests and in the mockPythonApi if necessary

func TestGetPlayerAnalytics(t *testing.T) {
    t.Run("Successful player data relay", func(t *testing.T) {
        playerID := "player1"
        matchID := "match1"
        expectedPath := fmt.Sprintf("/match/%s/player/%s/details", matchID, playerID)
        expectedResponse := map[string]interface{}{"data": "player_details", "player_id": playerID}

        mockApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            assert.Equal(t, expectedPath, r.URL.Path)
            // Removed: assert.Equal(t, matchID, r.URL.Query().Get("match_id"))
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(expectedResponse)
        }))
        defer mockApi.Close()

        ac := controllers.NewAnalyticsController(mockApi.URL, mockApi.Client())
		router := mux.NewRouter()
	// The actual route is /api/v1/analytics/players/{id} but mux expects path variables in handler registration
	router.HandleFunc("/analytics/players/{id}", ac.GetPlayerAnalytics).Methods("GET")

        reqPath := fmt.Sprintf("/analytics/players/%s?match_id=%s", playerID, matchID)
        req := httptest.NewRequest("GET", reqPath, nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        var actualResponse map[string]interface{}
        err := json.NewDecoder(rr.Body).Decode(&actualResponse)
        require.NoError(t, err)
        assert.Equal(t, expectedResponse, actualResponse)
    })

    t.Run("Missing match_id query for player", func(t *testing.T) {
        playerID := "player1"
        // No mock API needed as it should fail before calling it.
        ac := controllers.NewAnalyticsController("", nil) // URL/client don't matter
		router := mux.NewRouter()
	router.HandleFunc("/analytics/players/{id}", ac.GetPlayerAnalytics).Methods("GET")

        reqPath := fmt.Sprintf("/analytics/players/%s", playerID) // Missing match_id query
        req := httptest.NewRequest("GET", reqPath, nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusBadRequest, rr.Code)
        assert.Contains(t, rr.Body.String(), "match_id query parameter is required")
    })
}

func TestGetTeamAnalytics(t *testing.T) {
    t.Run("Successful team data relay", func(t *testing.T) {
        teamID := "teamA"
        matchID := "match1"
        expectedPath := fmt.Sprintf("/match/%s/team/%s/summary-over-time", matchID, teamID)
        expectedResponse := map[string]interface{}{"data": "team_summary_over_time", "team_id": teamID}

        mockApi := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            assert.Equal(t, expectedPath, r.URL.Path)
            // Removed: assert.Equal(t, matchID, r.URL.Query().Get("match_id"))
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusOK)
            json.NewEncoder(w).Encode(expectedResponse)
        }))
        defer mockApi.Close()

        ac := controllers.NewAnalyticsController(mockApi.URL, mockApi.Client())
		router := mux.NewRouter()
	router.HandleFunc("/analytics/teams/{id}", ac.GetTeamAnalytics).Methods("GET")

        reqPath := fmt.Sprintf("/analytics/teams/%s?match_id=%s", teamID, matchID)
        req := httptest.NewRequest("GET", reqPath, nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusOK, rr.Code)
        var actualResponse map[string]interface{}
        err := json.NewDecoder(rr.Body).Decode(&actualResponse)
        require.NoError(t, err)
        assert.Equal(t, expectedResponse, actualResponse)
    })

    t.Run("Missing match_id query for team", func(t *testing.T) {
        teamID := "teamA"
        ac := controllers.NewAnalyticsController("", nil) // URL/client don't matter
		router := mux.NewRouter()
	router.HandleFunc("/analytics/teams/{id}", ac.GetTeamAnalytics).Methods("GET")

        reqPath := fmt.Sprintf("/analytics/teams/%s", teamID) // Missing match_id
        req := httptest.NewRequest("GET", reqPath, nil)
        rr := httptest.NewRecorder()
        router.ServeHTTP(rr, req)

        assert.Equal(t, http.StatusBadRequest, rr.Code)
        assert.Contains(t, rr.Body.String(), "match_id query parameter is required")
    })
}

// Note: The refactoring to AnalyticsController with constructor injection
// makes the tests much cleaner and removes the dependency on t.Setenv
// or global variable manipulation for setting the Python API URL and HTTP client.
// The long comment block below discussing those older strategies can now be removed.
