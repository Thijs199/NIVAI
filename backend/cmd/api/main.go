package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"database/sql"
	"fmt"
	"syscall"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
	"nivai/backend/pkg/config"
	"nivai/backend/pkg/models"
	"nivai/backend/pkg/routes"
	"nivai/backend/pkg/services"
)

/**
 * Main entry point for the AIFAA API server.
 * Initializes configuration, sets up routes, and starts the HTTP server
 * with graceful shutdown capabilities.
 */
func main() {
	// Initialize logger
	logger := log.New(os.Stdout, "AIFAA API: ", log.LstdFlags)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage service
	logger.Println("Initializing storage service...")
	storageFactory := services.NewStorageFactory()
	storage, err := storageFactory.CreateDefaultStorage()

	if err != nil {
		logger.Printf("Warning: Could not initialize default storage: %v", err)
		// Check if we have an external data path configured
		if externalPath := os.Getenv("EXTERNAL_DATA_MOUNT"); externalPath != "" {
			logger.Printf("Attempting to use configured mount point: %s", externalPath)

			// Create directory if it doesn't exist
			if _, err := os.Stat(externalPath); os.IsNotExist(err) {
				logger.Printf("Creating mount directory: %s", externalPath)
				if err := os.MkdirAll(externalPath, 0755); err != nil {
					logger.Fatalf("Failed to create mount directory: %v", err)
				}
			}

			// Set environment variable expected by storage factory
			os.Setenv("EXTERNAL_DATA_PATH", externalPath)

			// Try to initialize storage again
			storage, err = storageFactory.CreateStorage(services.LocalFileStorageType)
			if err != nil {
				logger.Fatalf("Failed to initialize storage with mount point: %v", err)
			}
		} else {
			logger.Fatalf("No valid storage configuration found and no mount point specified")
		}
	}

	logger.Printf("Storage service initialized successfully")

	// Initialize database connection
	logger.Println("Initializing database connection...")
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Postgres.Host,
		cfg.Database.Postgres.Port,
		cfg.Database.Postgres.User,
		cfg.Database.Postgres.Password,
		cfg.Database.Postgres.DBName,
		cfg.Database.Postgres.SSLMode,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		logger.Fatalf("Failed to ping database: %v", err)
	}
	logger.Println("Database connection initialized successfully")

	// Create video repository
	videoRepo := models.NewPostgresVideoRepository(db)

	// Create router and register routes
	router := routes.SetupRoutes(cfg, storage, videoRepo)

	// Configure server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting server on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server gracefully
	logger.Println("Shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server exited properly")
}