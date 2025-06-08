package services

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// FileUploadInfo contains information about an uploaded file
type FileUploadInfo struct {
	Path     string // Storage path
	Provider string // Storage provider name
	Size     int64  // File size in bytes
	Format   string // File format/extension
}

/**
 * StorageService defines the interface for file storage operations.
 * Abstracts operations for uploading, retrieving, and managing stored files.
 */
type StorageService interface {
	// UploadFile uploads a file to storage
	UploadFile(file multipart.File, path string) (*FileUploadInfo, error)

	// GetFile retrieves a file from storage
	GetFile(path string) (io.ReadCloser, error)

	// DeleteFile removes a file from storage
	DeleteFile(path string) error

	// GetStreamURL generates a URL for streaming the file
	GetStreamURL(path string) (string, error)

	// GetFileMetadata retrieves metadata about a stored file
	GetFileMetadata(path string) (map[string]string, error)
}

/**
 * AzureBlobStorage implements the StorageService interface using Azure Blob Storage.
 */
type AzureBlobStorage struct {
	accountName   string
	accountKey    string
	containerName string
	credential    *azblob.SharedKeyCredential
	serviceURL    azblob.ServiceURL
	containerURL  azblob.ContainerURL
}

/**
 * NewAzureBlobStorage creates a new Azure Blob Storage service client.
 * Initializes the connection to Azure Blob Storage using provided credentials.
 *
 * @param accountName Azure storage account name
 * @param accountKey Azure storage account key
 * @param containerName Azure blob container name
 * @return A new storage service client or error
 */
func NewAzureBlobStorage(accountName, accountKey, containerName string) (StorageService, error) {
	// Validate parameters
	if accountName == "" || accountKey == "" || containerName == "" {
		return nil, errors.New("azure credentials cannot be empty")
	}

	// Create credential
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}

	// Create pipeline
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			MaxTries:      3,
			TryTimeout:    30 * time.Second,
			RetryDelay:    3 * time.Second,
			MaxRetryDelay: 30 * time.Second,
		},
	})

	// Create service URL
	serviceURL := azblob.NewServiceURL(
		url.URL{
			Scheme: "https",
			Host:   accountName + ".blob.core.windows.net",
		},
		pipeline,
	)

	// Get container URL
	containerURL := serviceURL.NewContainerURL(containerName)

	return &AzureBlobStorage{
		accountName:   accountName,
		accountKey:    accountKey,
		containerName: containerName,
		credential:    credential,
		serviceURL:    serviceURL,
		containerURL:  containerURL,
	}, nil
}

/**
 * UploadFile uploads a file to Azure Blob Storage.
 * Streams the file to the specified path in the storage container.
 *
 * @param file The file to upload
 * @param path The destination path in the storage
 * @return Upload information or error
 */
func (s *AzureBlobStorage) UploadFile(file multipart.File, path string) (*FileUploadInfo, error) {
	ctx := context.Background()

	// Create blob URL
	blobURL := s.containerURL.NewBlockBlobURL(path)

	// Upload file
	info, err := azblob.UploadStreamToBlockBlob(
		ctx,
		file,
		blobURL,
		azblob.UploadStreamToBlockBlobOptions{
			BufferSize: 2 * 1024 * 1024, // 2MB buffer
			MaxBuffers: 3,
		},
	)
	if err != nil {
		return nil, err
	}

	// Return upload info
	return &FileUploadInfo{
		Path:     path,
		Provider: "azure_blob",
		Size:     info.ContentLength,
		Format:   strings.TrimPrefix(filepath.Ext(path), "."),
	}, nil
}

/**
 * GetFile retrieves a file from Azure Blob Storage.
 * Downloads the blob from the specified path.
 *
 * @param path The path of the file in storage
 * @return A reader for the file content or error
 */
func (s *AzureBlobStorage) GetFile(path string) (io.ReadCloser, error) {
	ctx := context.Background()

	// Create blob URL
	blobURL := s.containerURL.NewBlockBlobURL(path)

	// Download blob
	response, err := blobURL.Download(ctx, 0, azblob.CountToEnd, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	// Create a reader from the response
	reader := response.Body(azblob.RetryReaderOptions{
		MaxRetries: 3,
	})

	return reader, nil
}

/**
 * DeleteFile removes a file from Azure Blob Storage.
 * Deletes the blob at the specified path.
 *
 * @param path The path of the file to delete
 * @return Error if deletion fails
 */
func (s *AzureBlobStorage) DeleteFile(path string) error {
	ctx := context.Background()

	// Create blob URL
	blobURL := s.containerURL.NewBlockBlobURL(path)

	// Delete blob
	_, err := blobURL.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
	return err
}

/**
 * GetStreamURL generates a URL for streaming a file from Azure Blob Storage.
 * Creates a Shared Access Signature (SAS) URL with temporary access.
 *
 * @param path The path of the file in storage
 * @return A temporary URL for accessing the file or error
 */
func (s *AzureBlobStorage) GetStreamURL(path string) (string, error) {
	// Create blob URL
	blobURL := s.containerURL.NewBlockBlobURL(path)

	// Create SAS token for the blob
	sasQueryParams, err := azblob.BlobSASSignatureValues{
		Protocol:      azblob.SASProtocolHTTPS,
		ExpiryTime:    time.Now().Add(1 * time.Hour), // URL valid for 1 hour
		ContainerName: s.containerName,
		BlobName:      path,
		Permissions:   azblob.BlobSASPermissions{Read: true}.String(),
	}.NewSASQueryParameters(s.credential)

	if err != nil {
		return "", err
	}

	// Construct the SAS URL
	qp := sasQueryParams.Encode()
	return blobURL.URL().String() + "?" + qp, nil
}

/**
 * GetFileMetadata retrieves metadata for a file in Azure Blob Storage.
 * Fetches properties and metadata of the blob.
 *
 * @param path The path of the file in storage
 * @return A map of metadata or error
 */
func (s *AzureBlobStorage) GetFileMetadata(path string) (map[string]string, error) {
	ctx := context.Background()

	// Create blob URL
	blobURL := s.containerURL.NewBlockBlobURL(path)

	// Get blob properties
	props, err := blobURL.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
	if err != nil {
		return nil, err
	}

	// Extract metadata into a map
	metadata := make(map[string]string)
	for k, v := range props.Metadata() {
		metadata[k] = v
	}

	// Add content properties
	metadata["content-length"] = string(props.ContentLength())
	metadata["content-type"] = string(props.ContentType())
	metadata["last-modified"] = props.LastModified().Format(time.RFC3339)

	return metadata, nil
}