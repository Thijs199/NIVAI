package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ContextKey type for request context keys
type ContextKey string

const (
	// RequestIDKey is the key used to store request ID in context
	RequestIDKey ContextKey = "requestID"

	// UserIDKey is the key used to store authenticated user ID in context
	UserIDKey ContextKey = "userID"
)

/**
 * Logger middleware logs HTTP requests with timing information.
 * Captures request method, path, status code, and response time.
 *
 * @param next The next handler in the chain
 * @return An http.Handler that performs logging
 */
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper to capture status code
		wrapper := newResponseWriter(w)

		// Process request
		next.ServeHTTP(wrapper, r)

		// Calculate request duration
		duration := time.Since(start)

		// Get request ID from context if available
		requestID := "unknown"
		if id, ok := r.Context().Value(RequestIDKey).(string); ok {
			requestID = id
		}

		// Log request details
		log.Printf(
			"[%s] %s %s %d %s",
			requestID,
			r.Method,
			r.URL.Path,
			wrapper.status,
			duration,
		)
	})
}

/**
 * CORS middleware adds Cross-Origin Resource Sharing headers to responses.
 * Configures which origins, methods, and headers are allowed.
 *
 * @param next The next handler in the chain
 * @return An http.Handler that handles CORS
 */
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

/**
 * RequestID middleware adds a unique ID to each request.
 * This ID is used for request tracing and debugging.
 *
 * @param next The next handler in the chain
 * @return An http.Handler that adds a request ID
 */
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generate a new UUID for the request
		requestID := uuid.New().String()

		// Add request ID to response headers for tracing
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

		// Serve request with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

/**
 * Authenticate middleware validates JWT tokens for protected routes.
 * Extracts and validates the token from the Authorization header.
 *
 * @param next The next handler in the chain
 * @return An http.Handler that performs authentication
 */
func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		// Check if the header has the correct format
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// TODO: Implement actual JWT validation logic
		// This is a placeholder - in a real implementation, we would:
		// 1. Parse and validate JWT token
		// 2. Check expiration time
		// 3. Extract user ID or other claims

		// For now, assume token is valid and add mock user ID to context
		ctx := context.WithValue(r.Context(), UserIDKey, "mock-user-id")

		// Pass the request with the authenticated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter is a wrapper for http.ResponseWriter that captures the status code
type responseWriter struct {
	http.ResponseWriter
	status int
}

// newResponseWriter creates a new responseWriter
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

// WriteHeader captures the status code and forwards to the embedded ResponseWriter
func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}