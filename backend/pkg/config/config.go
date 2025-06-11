package config

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration structure
type Config struct {
	// Server configuration
	Server struct {
		Port string `json:"port"`
		Host string `json:"host"`
	} `json:"server"`

	// Database configurations
	Database struct {
		Postgres struct {
			Host     string `json:"host"`
			Port     string `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
			DBName   string `json:"dbname"`
			SSLMode  string `json:"sslmode"`
		} `json:"postgres"`

		Redis struct {
			Host     string `json:"host"`
			Port     string `json:"port"`
			Password string `json:"password"`
			DB       int    `json:"db"`
		} `json:"redis"`
	} `json:"database"`

	// Storage configuration
	Storage struct {
		AzureBlobStorage struct {
			AccountName   string `json:"account_name"`
			AccountKey    string `json:"account_key"`
			ContainerName string `json:"container_name"`
		} `json:"azure_blob_storage"`
	} `json:"storage"`
}

// Load loads the configuration from a file and environment variables
func Load() (*Config, error) {
	// Initialize default configuration
	config := &Config{}

	// Default server configuration
	config.Server.Port = getEnvOrDefault("SERVER_PORT", "8080")
	config.Server.Host = getEnvOrDefault("SERVER_HOST", "0.0.0.0")

	// Default database configuration
	config.Database.Postgres.Host = getEnvOrDefault("DB_HOST", "localhost")
	config.Database.Postgres.Port = getEnvOrDefault("DB_PORT", "5432")
	config.Database.Postgres.User = getEnvOrDefault("DB_USER", "postgres")
	config.Database.Postgres.Password = getEnvOrDefault("DB_PASSWORD", "password")
	config.Database.Postgres.DBName = getEnvOrDefault("DB_NAME", "nivai")
	config.Database.Postgres.SSLMode = getEnvOrDefault("DB_SSL_MODE", "disable")

	// Default Redis configuration
	config.Database.Redis.Host = getEnvOrDefault("REDIS_HOST", "localhost")
	config.Database.Redis.Port = getEnvOrDefault("REDIS_PORT", "6379")
	config.Database.Redis.Password = getEnvOrDefault("REDIS_PASSWORD", "")

	// Try to load configuration from file if it exists
	configPath := getEnvOrDefault("CONFIG_PATH", "config.json")
	if _, err := os.Stat(configPath); err == nil {
		file, err := os.Open(configPath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		decoder := json.NewDecoder(file)
		if err := decoder.Decode(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

// getEnvOrDefault retrieves the value of the environment variable named by the key
// or returns the default value if the environment variable is not set
func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
