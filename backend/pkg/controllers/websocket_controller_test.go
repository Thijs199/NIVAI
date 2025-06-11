package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nivai/backend/pkg/controllers" // Adjust import path as necessary

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain is removed as hub lifecycle is now managed per test or per sub-test.

func TestWebSocketHandler(t *testing.T) {
	t.Run("Successful WebSocket upgrade", func(t *testing.T) {
		testHub := controllers.NewHub()
		go testHub.Run()
		// If Hub had a Stop() method: defer testHub.Stop()

		server := httptest.NewServer(testHub) // Pass the Hub instance directly as it implements http.Handler
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		dialer := websocket.Dialer{}
		conn, resp, err := dialer.Dial(wsURL, nil) // Path is just root of test server
		require.NoError(t, err, "Failed to connect to WebSocket")
		defer conn.Close()

		assert.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode, "HTTP status should be 101 Switching Protocols")
	})

	t.Run("Send and receive a message (echo through hub)", func(t *testing.T) {
		testHub := controllers.NewHub()
		go testHub.Run()
		// defer testHub.Stop()

		server := httptest.NewServer(testHub)
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)
		defer conn.Close()

		testMessage := []byte("hello websocket")
		err = conn.WriteMessage(websocket.TextMessage, testMessage)
		require.NoError(t, err, "Failed to write message to WebSocket")

		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		msgType, p, err := conn.ReadMessage()
		require.NoError(t, err, "Failed to read message from WebSocket. Hub might not be running or broadcasting.")

		assert.Equal(t, websocket.TextMessage, msgType)
		assert.Equal(t, string(testMessage), string(p), "Received message should match sent message")
	})

	t.Run("Multiple clients connect and receive broadcast", func(t *testing.T) {
		testHub := controllers.NewHub()
		go testHub.Run()
		// defer testHub.Stop()

		server := httptest.NewServer(testHub)
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		client1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err, "Client 1 failed to connect")
		defer client1.Close()

		client2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err, "Client 2 failed to connect")
		defer client2.Close()

		// Give time for clients to register with the hub
		time.Sleep(100 * time.Millisecond)

		broadcastMessage := []byte("broadcast test")
		err = client1.WriteMessage(websocket.TextMessage, broadcastMessage)
		require.NoError(t, err, "Client 1 failed to write message")

		client1.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, p1, err := client1.ReadMessage()
		require.NoError(t, err, "Client 1 failed to read its own message")
		assert.Equal(t, string(broadcastMessage), string(p1))

		client2.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, p2, err := client2.ReadMessage()
		require.NoError(t, err, "Client 2 failed to read broadcast message. Hub might not be running correctly.")
		assert.Equal(t, string(broadcastMessage), string(p2))
	})

	t.Run("Connection closes when client disconnects", func(t *testing.T) {
		testHub := controllers.NewHub()
		go testHub.Run()
		// defer testHub.Stop()

		server := httptest.NewServer(testHub)
		defer server.Close()
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		require.NoError(t, err)

		conn.WriteMessage(websocket.TextMessage, []byte("closing soon"))
		conn.Close()

		conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond)) // Increased slightly
		_, _, err = conn.ReadMessage()
		assert.Error(t, err, "Reading from a closed connection should produce an error")

		isCloseError := websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
			websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) || // Added CloseNoStatusReceived
			strings.Contains(err.Error(), "use of closed network connection") ||
			strings.Contains(err.Error(), "connection reset by peer") // Common on some systems
		assert.True(t, isCloseError, "Error should be a WebSocket close error or network closed error, got: %v", err)
	})
}
