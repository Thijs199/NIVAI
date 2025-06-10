package services_test

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"nivai/backend/pkg/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOsStat is used to control os.Stat behavior for testing CreateDefaultStorage
var mockOsStat func(name string) (os.FileInfo, error)

// Patched os.Stat that uses our mock
func patchedOsStat(name string) (os.FileInfo, error) {
	if mockOsStat != nil {
		return mockOsStat(name)
	}
	return os.Stat(name)
}

// fileInfoMock is a simple mock for os.FileInfo
type fileInfoMock struct {
	name    string
	isDir   bool
	modTime time.Time
	size    int64
	mode    os.FileMode
}

func (fim *fileInfoMock) Name() string       { return fim.name }
func (fim *fileInfoMock) Size() int64        { return fim.size }
func (fim *fileInfoMock) Mode() os.FileMode  { return fim.mode }
func (fim *fileInfoMock) ModTime() time.Time { return fim.modTime }
func (fim *fileInfoMock) IsDir() bool        { return fim.isDir }
func (fim *fileInfoMock) Sys() interface{}   { return nil }

// Simplified mockMultipartFile
type mockMultipartFile struct {
    *bytes.Reader
}

func (mf *mockMultipartFile) Close() error { return nil }

func newMockMultipartFile(content string) multipart.File {
    return &mockMultipartFile{
        Reader: bytes.NewReader([]byte(content)),
    }
}

func TestNewLocalFileStorage(t *testing.T) {
    t.Run("Valid base path", func(t *testing.T) {
        tempDir, err := os.MkdirTemp("", "localfs_new_valid")
        require.NoError(t, err)
        defer os.RemoveAll(tempDir)

        fs, err := services.NewLocalFileStorage(tempDir)
        require.NoError(t, err)
        assert.NotNil(t, fs)
    })

    t.Run("Empty base path", func(t *testing.T) {
        _, err := services.NewLocalFileStorage("")
        assert.Error(t, err)
        if err != nil {
            assert.Contains(t, err.Error(), "base path cannot be empty")
        }
    })

    t.Run("Base path is a file, not a directory", func(t *testing.T) {
        tempFile, err := os.CreateTemp("", "localfs_new_file")
        require.NoError(t, err)
        defer os.Remove(tempFile.Name())
        tempFile.Close()

        _, err = services.NewLocalFileStorage(tempFile.Name())
        assert.Error(t, err)
        if err != nil {
            assert.Contains(t, err.Error(), "base path must be a directory")
        }
    })

    t.Run("Base path does not exist", func(t *testing.T) {
        nonExistentPath := filepath.Join(os.TempDir(), "localfs_new_nonexistent", "somerandompath_full")
        _, err := services.NewLocalFileStorage(nonExistentPath)
        assert.Error(t, err)
        if err != nil {
            assert.True(t, strings.Contains(err.Error(), "error accessing base path") || strings.Contains(err.Error(), "no such file or directory"))
        }
    })
}

func TestLocalFileStorage_Operations(t *testing.T) {
    baseDir, mkdirErr := os.MkdirTemp("", "localfs_ops_base_full")
    require.NoError(t, mkdirErr, "Failed to create temp base dir")
    defer os.RemoveAll(baseDir)

    var fs services.StorageService
    var fsErr error
    fs, fsErr = services.NewLocalFileStorage(baseDir)
    require.NoError(t, fsErr, "Failed to create LocalFileStorage instance")
    require.NotNil(t, fs, "LocalFileStorage instance should not be nil")

    fileName := "test_upload.txt"
    fileContent := "Hello, Local Storage!"
    uploadPath := filepath.Join("subdir", fileName)

    t.Run("UploadFile success", func(t *testing.T) {
        mockFile := newMockMultipartFile(fileContent)

        uploadInfo, err := fs.UploadFile(mockFile, uploadPath)
        require.NoError(t, err)
        require.NotNil(t, uploadInfo)

        assert.Equal(t, uploadPath, uploadInfo.Path)
        assert.Equal(t, "local_file", uploadInfo.Provider)
        assert.Equal(t, int64(len(fileContent)), uploadInfo.Size)
        assert.Equal(t, "txt", uploadInfo.Format)

        fullDiskPath := filepath.Join(baseDir, uploadPath)
        contentBytes, readErr := os.ReadFile(fullDiskPath)
        require.NoError(t, readErr)
        assert.Equal(t, fileContent, string(contentBytes))
    })

    t.Run("GetFile success", func(t *testing.T) {
        reader, err := fs.GetFile(uploadPath)
        require.NoError(t, err)
        defer reader.Close()

        contentBytes, err := io.ReadAll(reader)
        require.NoError(t, err)
        assert.Equal(t, fileContent, string(contentBytes))
    })

    t.Run("GetFile not found", func(t *testing.T) {
        _, err := fs.GetFile("nonexistent/file.txt")
        assert.Error(t, err)
        if err != nil {
            assert.Contains(t, err.Error(), "file not found")
        }
    })

    t.Run("GetFileMetadata success", func(t *testing.T) {
        metadata, err := fs.GetFileMetadata(uploadPath)
        require.NoError(t, err)
        require.NotNil(t, metadata)

        assert.Equal(t, fmt.Sprintf("%d", len(fileContent)), metadata["content-length"])
        assert.Equal(t, fileName, metadata["name"])
        assert.Equal(t, "false", metadata["is-directory"])

        _, parseErr := time.Parse(time.RFC3339, metadata["last-modified"])
        assert.NoError(t, parseErr, "last-modified should be a valid RFC3339 timestamp")
    })

    t.Run("GetFileMetadata not found", func(t *testing.T) {
        _, err := fs.GetFileMetadata("nonexistent/metadata.txt")
        assert.Error(t, err)
        if err != nil {
            assert.Contains(t, err.Error(), "file not found")
        }
    })

    t.Run("GetStreamURL success", func(t *testing.T) {
        streamURL, err := fs.GetStreamURL(uploadPath)
        require.NoError(t, err)

        expectedAbsPath, absErr := filepath.Abs(filepath.Join(baseDir, uploadPath))
        require.NoError(t, absErr)
        assert.True(t, strings.HasPrefix(streamURL, "file://"), "URL should start with file://")
        assert.True(t, strings.HasSuffix(streamURL, filepath.ToSlash(expectedAbsPath)), "URL should end with correct path")
    })

    t.Run("GetStreamURL not found", func(t *testing.T) {
        _, err := fs.GetStreamURL("nonexistent/stream.txt")
        assert.Error(t, err)
        if err != nil {
            assert.Contains(t, err.Error(), "file not found")
        }
    })

    t.Run("DeleteFile success", func(t *testing.T) {
        err := fs.DeleteFile(uploadPath)
        require.NoError(t, err)

        fullDiskPath := filepath.Join(baseDir, uploadPath)
        _, statErr := os.Stat(fullDiskPath)
        assert.True(t, os.IsNotExist(statErr), "File should not exist after deletion")
    })

    t.Run("DeleteFile not found", func(t *testing.T) {
        err := fs.DeleteFile("nonexistent/delete_me.txt")
        assert.Error(t, err)
        if err != nil {
            assert.Contains(t, err.Error(), "file not found")
        }
    })

    t.Run("UploadFile to a non-existent nested directory", func(t *testing.T) {
        nestedPath := filepath.Join("deeply", "nested", "dir", "another_upload.txt")
        nestedFileContent := "nested content"
        mockFile := newMockMultipartFile(nestedFileContent)

        uploadInfo, err := fs.UploadFile(mockFile, nestedPath)
        require.NoError(t, err)
        require.NotNil(t, uploadInfo)
        assert.Equal(t, nestedPath, uploadInfo.Path)

        fullDiskPath := filepath.Join(baseDir, nestedPath)
        contentBytes, readErr := os.ReadFile(fullDiskPath)
        require.NoError(t, readErr)
        assert.Equal(t, nestedFileContent, string(contentBytes))

        delErr := fs.DeleteFile(nestedPath)
        assert.NoError(t, delErr, "Cleanup of nested file failed")
    })
}

func TestStorageFactory_CreateDefaultStorage(t *testing.T) { // Copied from previous correct version
	factory := services.NewStorageFactory()
    originalOsStat := services.OsStat
    services.OsStat = patchedOsStat
    defer func() { services.OsStat = originalOsStat }()

	cleanupEnv := func() {
		os.Unsetenv("EXTERNAL_DATA_PATH")
		os.Unsetenv("AZURE_STORAGE_ACCOUNT")
		os.Unsetenv("AZURE_STORAGE_KEY")
		os.Unsetenv("AZURE_STORAGE_CONTAINER")
	}
	defer cleanupEnv()

	t.Run("Local storage configured and path valid", func(t *testing.T) {
		cleanupEnv()
		tempDir, _ := os.MkdirTemp("", "default_local_valid")
		defer os.RemoveAll(tempDir)
		t.Setenv("EXTERNAL_DATA_PATH", tempDir)

        mockOsStat = func(name string) (os.FileInfo, error) {
            assert.Equal(t, tempDir, name)
            return &fileInfoMock{name: filepath.Base(tempDir), isDir: true}, nil
        }
        defer func() { mockOsStat = nil }()

		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
	})

	t.Run("Local storage configured but path invalid, fallback to Azure", func(t *testing.T) {
		cleanupEnv()
		t.Setenv("EXTERNAL_DATA_PATH", "/nonexistentpath_for_testing_stat_fail")
		t.Setenv("AZURE_STORAGE_ACCOUNT", "testaccount_azure")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=")
		t.Setenv("AZURE_STORAGE_CONTAINER", "testcontainer_azure")

        mockOsStat = func(name string) (os.FileInfo, error) {
            assert.Equal(t, "/nonexistentpath_for_testing_stat_fail", name)
            return nil, os.ErrNotExist
        }
        defer func() { mockOsStat = nil }()

		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
	})

	t.Run("Only Azure storage configured", func(t *testing.T) {
		cleanupEnv()
		t.Setenv("AZURE_STORAGE_ACCOUNT", "azure_only_account")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=")
		t.Setenv("AZURE_STORAGE_CONTAINER", "azure_only_container")
        mockOsStat = func(name string) (os.FileInfo, error) {
            t.Fatalf("os.Stat should not be called when EXTERNAL_DATA_PATH is not set")
            return nil, nil
        }
        defer func() { mockOsStat = nil }()

		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
	})

	t.Run("Local storage valid, Azure also configured (local preferred)", func(t *testing.T) {
		cleanupEnv()
		tempDir, _ := os.MkdirTemp("", "default_local_preferred")
		defer os.RemoveAll(tempDir)
		t.Setenv("EXTERNAL_DATA_PATH", tempDir)
		t.Setenv("AZURE_STORAGE_ACCOUNT", "azure_preferred_account")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=")
		t.Setenv("AZURE_STORAGE_CONTAINER", "azure_preferred_container")

        mockOsStat = func(name string) (os.FileInfo, error) {
            return &fileInfoMock{name: filepath.Base(tempDir), isDir: true}, nil
        }
        defer func() { mockOsStat = nil }()

		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
	})

	t.Run("No storage configuration found", func(t *testing.T) {
		cleanupEnv()
        mockOsStat = func(name string) (os.FileInfo, error) {
            t.Fatalf("os.Stat should not be called when EXTERNAL_DATA_PATH is not set")
            return nil, nil
        }
        defer func() { mockOsStat = nil }()

		_, err := factory.CreateDefaultStorage()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no valid storage configuration found")
	})
}
// Note: The mockMultipartFile is a basic stand-in.
// The LocalFileStorage.UploadFile method uses io.Copy, which works with io.Reader.
// It doesn't explicitly use multipart.FileHeader for anything other than perhaps logging or metadata in a more complex system.
// The FileUploadInfo also takes Format from filepath.Ext(path) not from header.
// The current implementation of LocalFileStorage.UploadFile does not use the header argument.
