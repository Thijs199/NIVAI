package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"nivai/backend/pkg/controllers" // Adjust import path as necessary

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogin(t *testing.T) {
	t.Run("Successful login with mock credentials", func(t *testing.T) {
		credentials := map[string]string{
			"username": "testuser",
			"password": "password",
		}
		body, _ := json.Marshal(credentials)
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		http.HandlerFunc(controllers.Login).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "mock_access_token", response["access_token"])
		assert.Equal(t, "mock_refresh_token", response["refresh_token"])
		assert.Equal(t, float64(3600), response["expires_in"]) // JSON numbers are float64
		assert.Equal(t, "Bearer", response["token_type"])
	})

	t.Run("Invalid request payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		http.HandlerFunc(controllers.Login).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
	})

	t.Run("Empty request payload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/login", nil) // No body
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		http.HandlerFunc(controllers.Login).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload") // Due to EOF error in JSON decoding
	})
}

func TestRefreshToken(t *testing.T) {
	t.Run("Successful token refresh with mock token", func(t *testing.T) {
		requestBody := map[string]string{
			"refresh_token": "some_refresh_token",
		}
		body, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		http.HandlerFunc(controllers.RefreshToken).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]interface{}
		err := json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, "new_mock_access_token", response["access_token"])
		assert.Equal(t, float64(3600), response["expires_in"])
		assert.Equal(t, "Bearer", response["token_type"])
	})

	t.Run("Invalid request payload for refresh token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/refresh", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		http.HandlerFunc(controllers.RefreshToken).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
	})

	t.Run("Empty request payload for refresh token", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/refresh", nil) // No body
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		http.HandlerFunc(controllers.RefreshToken).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid request payload")
	})
}
