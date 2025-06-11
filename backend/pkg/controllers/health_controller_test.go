package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"nivai/backend/pkg/controllers" // Adjust import path as necessary

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthCheck(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	rr := httptest.NewRecorder()

	http.HandlerFunc(controllers.HealthCheck).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Status code should be 200 OK")
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type should be application/json")

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err, "Should be able to decode response body")

	assert.Equal(t, "ok", response["status"], "Response status should be 'ok'")
	assert.Equal(t, "AIFAA API", response["service"], "Response service name should be 'AIFAA API'")

	// Check timestamp roughly
	timestampStr, ok := response["timestamp"].(string)
	require.True(t, ok, "Timestamp should be a string")

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	require.NoError(t, err, "Timestamp should be in RFC3339 format")

	// Check if the timestamp is recent (e.g., within the last 5 seconds)
	assert.WithinDuration(t, time.Now(), timestamp, 5*time.Second, "Timestamp should be recent")
}
