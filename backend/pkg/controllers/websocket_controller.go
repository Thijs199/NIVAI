package controllers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

/**
 * Client represents a connected WebSocket client.
 * Manages the connection and message handling for a single client.
 */
type Client struct {
	// The WebSocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Reference to the hub for broadcasting
	hub *Hub
}

/**
 * Hub maintains active clients and broadcasts messages to them.
 * Implements the pub/sub pattern for WebSocket communication.
 */
type Hub struct {
	// Registered clients map
	clients map[*Client]bool

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast message to all clients
	broadcast chan []byte

	// Mutex for concurrent access to clients map
	mu sync.Mutex
}

// WebSocket connection upgrader with configuration
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin for development
	// In production, this should be restricted
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

/**
 * NewHub creates a new hub instance.
 * Initializes channels and client map for the hub.
 *
 * @return A new Hub instance ready to be run
 */
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
		mu:         sync.Mutex{},
	}
}

/**
 * Run starts the hub's main loop.
 * Handles client registration, unregistration, and message broadcasting.
 * Should be run in a goroutine.
 */
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			// Register new client
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()

		case client := <-h.unregister:
			// Unregister client and close connection
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			// Broadcast message to all connected clients
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- message:
					// Message sent successfully
				default:
					// Failed to send, remove client
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

/**
 * readPump pumps messages from the WebSocket connection to the hub.
 * Continuously reads from the WebSocket and forwards messages to the hub.
 * Must be run in a separate goroutine.
 */
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Forward the message to the hub for broadcasting
		c.hub.broadcast <- message
	}
}

/**
 * writePump pumps messages from the hub to the WebSocket connection.
 * Continuously sends messages from the client's send channel to the WebSocket.
 * Must be run in a separate goroutine.
 */
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	for {
		message, ok := <-c.send
		if !ok {
			// The hub closed the channel
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("Error writing to WebSocket: %v", err)
			return
		}
	}
}

/**
 * WebSocketHandler upgrades HTTP connections to WebSocket connections.
 * Creates a new client for each WebSocket connection and starts its pumps.
 *
 * @param w The HTTP response writer
 * @param r The HTTP request
 */
// WebSocketHandler becomes ServeHTTP, a method of Hub
func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}

	// Create a new client
	client := &Client{
		conn: conn,
		send: make(chan []byte, 256),
		hub:  h, // Use the hub instance 'h'
	}

	// Register the client
	client.hub.register <- client // Register to the specific hub instance

	// Start the client's read and write pumps in goroutines
	go client.writePump()
	go client.readPump()
}