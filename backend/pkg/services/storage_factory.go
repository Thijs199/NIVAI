package services

import (
	"errors"
	"fmt"
	"os"
)

/**
 * StorageType represents the type of storage service to use.
 */
type StorageType string

const (
	// AzureBlobStorageType represents Azure Blob Storage
	AzureBlobStorageType StorageType = "azure_blob"
	
	// LocalFileStorageType represents local file system storage
	LocalFileStorageType StorageType = "local_file"
)

/**
 * StorageFactory creates and configures storage services based on configuration.
 * Implements the Factory design pattern to abstract storage implementation creation.
 */
type StorageFactory struct {}

/**
 * NewStorageFactory creates a new storage factory instance.
 *
 * @return A new storage factory
 */
func NewStorageFactory() *StorageFactory {
	return &StorageFactory{}
}

/**
 * CreateStorage creates and returns the appropriate storage service based on configuration.
 * Selects between Azure Blob Storage and Local File Storage based on environment variables.
 *
 * @param storageType The type of storage to create
 * @return A configured storage service or error
 */
func (f *StorageFactory) CreateStorage(storageType StorageType) (StorageService, error) {
	switch storageType {
	case AzureBlobStorageType:
		// Get Azure credentials from environment
		accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
		accountKey := os.Getenv("AZURE_STORAGE_KEY")
		containerName := os.Getenv("AZURE_STORAGE_CONTAINER")
		
		// Validate required values
		if accountName == "" || accountKey == "" || containerName == "" {
			return nil, errors.New("missing required Azure Storage configuration")
		}
		
		// Create and return Azure blob storage service
		return NewAzureBlobStorage(accountName, accountKey, containerName)
		
	case LocalFileStorageType:
		// Get base path from environment
		basePath := os.Getenv("EXTERNAL_DATA_PATH")
		
		// Validate required values
		if basePath == "" {
			return nil, errors.New("missing required Local Storage configuration: EXTERNAL_DATA_PATH")
		}
		
		// Create and return local file storage service
		return NewLocalFileStorage(basePath)
		
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

/**
 * CreateDefaultStorage creates a storage service based on environment variables.
 * Automatically determines which storage type to use based on available configuration.
 *
 * @return A configured storage service or error
 */
func (f *StorageFactory) CreateDefaultStorage() (StorageService, error) {
	// First, check if external data path is set for local file storage
	if externalPath := os.Getenv("EXTERNAL_DATA_PATH"); externalPath != "" {
		// Verify the path exists and is accessible
		if _, err := os.Stat(externalPath); err == nil {
			return f.CreateStorage(LocalFileStorageType)
		}
	}
	
	// If local storage isn't configured, try Azure Blob
	if accountName := os.Getenv("AZURE_STORAGE_ACCOUNT"); accountName != "" {
		return f.CreateStorage(AzureBlobStorageType)
	}
	
	// No storage configuration found
	return nil, errors.New("no valid storage configuration found")
}