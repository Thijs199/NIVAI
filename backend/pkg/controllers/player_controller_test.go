package controllers_test // Use _test package to test as a blackbox (mostly)

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"nivai/backend/pkg/controllers" // Adjust import path if your module structure is different

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchPlayerImage(t *testing.T) {
	playerController := controllers.NewPlayerController()

	t.Run("Successful placeholder generation", func(t *testing.T) {
		playerName := "Test Player"
		req := httptest.NewRequest("GET", "/api/v1/analytics/players/image_search?name="+url.QueryEscape(playerName), nil)
		rr := httptest.NewRecorder()

		// In a real scenario with mux, you'd setup a router and serve via it.
		// For a direct handler test, this is fine.
		// If the handler relied on mux.Vars, then a router setup would be needed.
		// This handler uses r.URL.Query().Get(), so direct call is okay.
		http.HandlerFunc(playerController.SearchPlayerImage).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		expectedImageUrlPart := "https://via.placeholder.com/150/808080/FFFFFF?Text=Player+" + url.QueryEscape(playerName)
		assert.Equal(t, expectedImageUrlPart, response["image_url"])
	})

	t.Run("Missing name query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/analytics/players/image_search", nil) // No name query
		rr := httptest.NewRecorder()

		http.HandlerFunc(playerController.SearchPlayerImage).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)

		bodyString := rr.Body.String()
		assert.Contains(t, bodyString, "Query parameter 'name' (player name) is required.")
	})

	t.Run("Empty name query parameter", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/analytics/players/image_search?name=", nil) // Empty name
		rr := httptest.NewRecorder()

		http.HandlerFunc(playerController.SearchPlayerImage).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		bodyString := rr.Body.String()
		assert.Contains(t, bodyString, "Query parameter 'name' (player name) is required.")
	})

	t.Run("Player name with spaces and special characters", func(t *testing.T) {
		playerName := "Player Name & Son"
		escapedName := url.QueryEscape(playerName)
		req := httptest.NewRequest("GET", "/api/v1/analytics/players/image_search?name="+escapedName, nil)
		rr := httptest.NewRecorder()

		http.HandlerFunc(playerController.SearchPlayerImage).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		expectedImageUrl := "https://via.placeholder.com/150/808080/FFFFFF?Text=Player+" + escapedName
		assert.Equal(t, expectedImageUrl, response["image_url"])
		// Check that the placeholder URL itself is well-formed (the part after Text= is what was escaped)
		assert.True(t, strings.HasSuffix(response["image_url"], url.QueryEscape("Player "+playerName)))
	})
}
