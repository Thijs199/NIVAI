package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/**
 * LocalFileStorage implements the StorageService interface using the local file system.
 * This can be used for local development or for accessing a mounted file share.
 */
type LocalFileStorage struct {
	basePath string // Base path for file storage
}

/**
 * NewLocalFileStorage creates a new local file storage service.
 *
 * @param basePath The base directory path for file storage
 * @return A new storage service client or error
 */
func NewLocalFileStorage(basePath string) (StorageService, error) {
	// Validate parameters
	if basePath == "" {
		return nil, errors.New("base path cannot be empty")
	}

	// Check if directory exists
	info, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("error accessing base path: %v", err)
	}
	if !info.IsDir() {
		return nil, errors.New("base path must be a directory")
	}

	return &LocalFileStorage{
		basePath: basePath,
	}, nil
}

/**
 * UploadFile copies a file to the local storage path.
 * Ensures the destination directory exists and writes the file.
 *
 * @param file The file to upload
 * @param path The destination path in the storage
 * @return Upload information or error
 */
func (s *LocalFileStorage) UploadFile(file multipart.File, path string) (*FileUploadInfo, error) {
	// Create full path
	fullPath := filepath.Join(s.basePath, path)
	dirPath := filepath.Dir(fullPath)

	// Ensure directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy file contents
	written, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %v", err)
	}

	// Return upload info
	return &FileUploadInfo{
		Path:     path,
		Provider: "local_file",
		Size:     written,
		Format:   strings.TrimPrefix(filepath.Ext(path), "."),
	}, nil
}

/**
 * GetFile retrieves a file from local storage.
 * Opens the file at the specified path for reading.
 *
 * @param path The path of the file in storage
 * @return A reader for the file content or error
 */
func (s *LocalFileStorage) GetFile(path string) (io.ReadCloser, error) {
	// Create full path
	fullPath := filepath.Join(s.basePath, path)

	// Open file for reading
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file not found")
		}
		return nil, fmt.Errorf("failed to open file: %v", err)
	}

	return file, nil
}

/**
 * DeleteFile removes a file from local storage.
 * Deletes the file at the specified path.
 *
 * @param path The path of the file to delete
 * @return Error if deletion fails
 */
func (s *LocalFileStorage) DeleteFile(path string) error {
	// Create full path
	fullPath := filepath.Join(s.basePath, path)

	// Delete file
	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("file not found")
		}
		return fmt.Errorf("failed to delete file: %v", err)
	}

	return nil
}

/**
 * GetStreamURL generates a local file URL for streaming.
 * Since this is a local implementation, it returns a file:// URL.
 * Note: This may not work in all contexts due to browser security restrictions.
 *
 * @param path The path of the file in storage
 * @return A URL for accessing the file or error
 */
func (s *LocalFileStorage) GetStreamURL(path string) (string, error) {
	// Create full path
	fullPath := filepath.Join(s.basePath, path)

	// Check if file exists
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("file not found")
		}
		return "", fmt.Errorf("failed to access file: %v", err)
	}

	// Convert to absolute path for URL
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Return file URL (note: this has limited usability)
	return "file://" + absPath, nil
}

/**
 * GetFileMetadata retrieves metadata for a file in local storage.
 * Gets file information from the file system.
 *
 * @param path The path of the file in storage
 * @return A map of metadata or error
 */
func (s *LocalFileStorage) GetFileMetadata(path string) (map[string]string, error) {
	// Create full path
	fullPath := filepath.Join(s.basePath, path)

	// Get file stats
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file not found")
		}
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}

	// Extract metadata into a map
	metadata := make(map[string]string)
	metadata["content-length"] = fmt.Sprintf("%d", info.Size())
	metadata["last-modified"] = info.ModTime().Format(time.RFC3339)
	metadata["name"] = info.Name()
	metadata["is-directory"] = fmt.Sprintf("%t", info.IsDir())
	metadata["mode"] = info.Mode().String()

	return metadata, nil
}