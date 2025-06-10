package routes

import (
	"net/http"
	"nivai/backend/pkg/config"
	"nivai/backend/pkg/controllers"
	"nivai/backend/pkg/middleware"
	"nivai/backend/pkg/models" // Added for VideoRepository
	"nivai/backend/pkg/services"

	"github.com/gorilla/mux"
)

/**
 * SetupRoutes creates and configures the main router for the API.
 * It registers all API endpoints and applies necessary middleware.
 *
 * @param cfg Configuration for the application
 * @param storage Storage service for file operations
 * @param videoRepo Repository for video data operations
 * @return The configured router
 */
func SetupRoutes(cfg *config.Config, storage services.StorageService, videoRepo models.VideoRepository) http.Handler {
	// Initialize router
	router := mux.NewRouter()

	// Apply common middleware to all routes
	router.Use(middleware.Logger)
	router.Use(middleware.CORS)
	router.Use(middleware.RequestID)

	// Create controller instances with dependencies
	videoController := controllers.NewVideoController(videoRepo, storage)
	// VideoService is needed for MatchController.
	videoServiceForMatch := services.NewVideoService(videoRepo, storage)
	matchController := controllers.NewMatchController(videoServiceForMatch)
	playerController := controllers.NewPlayerController()


	// API version prefix
	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint - no auth required
	apiRouter.HandleFunc("/health", controllers.HealthCheck).Methods("GET")

	// Auth endpoints
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", controllers.Login).Methods("POST")
	authRouter.HandleFunc("/refresh", controllers.RefreshToken).Methods("POST")

	// User endpoints - requires authentication
	userRouter := apiRouter.PathPrefix("/users").Subrouter()
	userRouter.Use(middleware.Authenticate)
	// userRouter.HandleFunc("", controllers.GetUsers).Methods("GET")
	// userRouter.HandleFunc("/{id}", controllers.GetUser).Methods("GET")

	// Video endpoints - requires authentication
	videoRouter := apiRouter.PathPrefix("/videos").Subrouter()
	videoRouter.Use(middleware.Authenticate)
	videoRouter.HandleFunc("", videoController.ListVideos).Methods("GET")
	videoRouter.HandleFunc("", videoController.UploadVideo).Methods("POST")
	videoRouter.HandleFunc("/{id}", videoController.GetVideo).Methods("GET")
	videoRouter.HandleFunc("/{id}", videoController.DeleteVideo).Methods("DELETE")

	// Analytics endpoints - requires authentication
	analyticsRouter := apiRouter.PathPrefix("/analytics").Subrouter()
	analyticsRouter.Use(middleware.Authenticate)
	analyticsRouter.HandleFunc("/matches/{id}", controllers.GetMatchAnalytics).Methods("GET")
	analyticsRouter.HandleFunc("/players/{id}", controllers.GetPlayerAnalytics).Methods("GET") // Player details by ID
	analyticsRouter.HandleFunc("/teams/{id}", controllers.GetTeamAnalytics).Methods("GET")
	analyticsRouter.HandleFunc("/players/image_search", playerController.SearchPlayerImage).Methods("GET") // Player image search by name

	// Matches list endpoint - requires authentication
	// This is a new top-level resource under /api/v1, similar to /videos or /users
	matchesRouter := apiRouter.PathPrefix("/matches").Subrouter()
	matchesRouter.Use(middleware.Authenticate)
	matchesRouter.HandleFunc("", matchController.ListMatches).Methods("GET")

	// WebSocket endpoint for real-time updates
	router.HandleFunc("/ws", controllers.WebSocketHandler)

	return router
}