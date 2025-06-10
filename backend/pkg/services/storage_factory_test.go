package services_test

import (
	"os"
	"path/filepath"
	"testing"
	"time" // Required by fileInfoMock, even if not directly by all tests

	"nivai/backend/pkg/services" // Adjust import path as necessary

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
	return os.Stat(name) // Fallback to real os.Stat if mock not set
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


func TestStorageFactory_CreateStorage(t *testing.T) {
	factory := services.NewStorageFactory()

	t.Run("Azure Blob Storage type", func(t *testing.T) {
		t.Setenv("AZURE_STORAGE_ACCOUNT", "testaccount")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=") // Longer, more realistic fake base64
		t.Setenv("AZURE_STORAGE_CONTAINER", "testcontainer")

		storage, err := factory.CreateStorage(services.AzureBlobStorageType)
		require.NoError(t, err)
		assert.NotNil(t, storage)
		// Check if it's the AzureBlobStorage type (requires type assertion or reflection,
		// or checking a known method/field if interfaces were different)
		// For now, NotNil is a basic check. A more specific type check is desirable.
		// Example: _, ok := storage.(*services.AzureBlobStorage); assert.True(t, ok)
		// This requires AzureBlobStorage to be an exported type if this test is in services_test.
		// If AzureBlobStorage is not exported, we can't directly assert its type from services_test.
		// Let's assume we can check via a known behavior or a Type() string method if added.
		// For this test, we will rely on the fact that no error means success.
	})

	t.Run("Azure Blob Storage missing config", func(t *testing.T) {
		t.Setenv("AZURE_STORAGE_ACCOUNT", "") // Missing account
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdA==")
		t.Setenv("AZURE_STORAGE_CONTAINER", "testcontainer")

		_, err := factory.CreateStorage(services.AzureBlobStorageType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required Azure Storage configuration")
	})

	t.Run("Local File Storage type", func(t *testing.T) {
		// Create a temporary directory for the test
		tempDir, err := os.MkdirTemp("", "localstorage_test")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		t.Setenv("EXTERNAL_DATA_PATH", tempDir)

		storage, err := factory.CreateStorage(services.LocalFileStorageType)
		require.NoError(t, err)
		assert.NotNil(t, storage)
		// Similar to Azure, direct type assertion like:
		// _, ok := storage.(*services.LocalFileStorage); assert.True(t, ok)
		// requires LocalFileStorage to be exported or tested within 'package services'.
	})

	t.Run("Local File Storage missing config", func(t *testing.T) {
		t.Setenv("EXTERNAL_DATA_PATH", "") // Missing path

		_, err := factory.CreateStorage(services.LocalFileStorageType)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required Local Storage configuration")
	})

    t.Run("Local File Storage path is not a directory", func(t *testing.T) {
        tempFile, err := os.CreateTemp("", "not_a_dir")
        require.NoError(t, err)
        defer os.Remove(tempFile.Name())
        tempFile.Close()

        t.Setenv("EXTERNAL_DATA_PATH", tempFile.Name())
        _, err = factory.CreateStorage(services.LocalFileStorageType)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "base path must be a directory")
    })

	t.Run("Unsupported storage type", func(t *testing.T) {
		_, err := factory.CreateStorage(services.StorageType("unknown_type"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported storage type")
	})
}

func TestStorageFactory_CreateDefaultStorage(t *testing.T) {
	factory := services.NewStorageFactory()
    originalOsStat := services.OsStat // Store original os.Stat
    services.OsStat = patchedOsStat      // Patch os.Stat
    defer func() { services.OsStat = originalOsStat }() // Restore original
    // Re-enabled OsStat patching.

	// Cleanup function to unset all relevant env vars
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
		// Check if it's LocalFileStorage (indirectly, e.g. by trying to use a feature specific to it if possible, or by type name if exposed)
        // For now, assert.NotNil and no error is the main check.
	})

	t.Run("Local storage configured but path invalid, fallback to Azure", func(t *testing.T) {
		cleanupEnv()
		t.Setenv("EXTERNAL_DATA_PATH", "/nonexistentpath_for_testing_stat_fail")
		t.Setenv("AZURE_STORAGE_ACCOUNT", "testaccount_azure")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=") // Longer fake base64
		t.Setenv("AZURE_STORAGE_CONTAINER", "testcontainer_azure")

        mockOsStat = func(name string) (os.FileInfo, error) {
            assert.Equal(t, "/nonexistentpath_for_testing_stat_fail", name)
            return nil, os.ErrNotExist // Simulate os.Stat failing
        }
        defer func() { mockOsStat = nil }()

		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
		// This should be an Azure storage instance.
	})

	t.Run("Only Azure storage configured", func(t *testing.T) {
		cleanupEnv()
		t.Setenv("AZURE_STORAGE_ACCOUNT", "azure_only_account")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=") // Longer fake base64
		t.Setenv("AZURE_STORAGE_CONTAINER", "azure_only_container")
        mockOsStat = func(name string) (os.FileInfo, error) {
            // This shouldn't be called if EXTERNAL_DATA_PATH is not set
            t.Fatalf("os.Stat should not be called when EXTERNAL_DATA_PATH is not set")
            return nil, nil
        }
        defer func() { mockOsStat = nil }()


		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
		// This should be an Azure storage instance.
	})

	t.Run("Local storage valid, Azure also configured (local preferred)", func(t *testing.T) {
		cleanupEnv()
		tempDir, _ := os.MkdirTemp("", "default_local_preferred")
		defer os.RemoveAll(tempDir)
		t.Setenv("EXTERNAL_DATA_PATH", tempDir)
		t.Setenv("AZURE_STORAGE_ACCOUNT", "azure_preferred_account")
		t.Setenv("AZURE_STORAGE_KEY", "dGVzdGtleV9tdXN0X2JlX2xvbmdlcl9hbmRfZW5jb2RlZF9jb3JyZWN0bHlhYmMxMjM0NTY3ODkwYWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXo=") // Longer fake base64
		t.Setenv("AZURE_STORAGE_CONTAINER", "azure_preferred_container")

        mockOsStat = func(name string) (os.FileInfo, error) {
            return &fileInfoMock{name: filepath.Base(tempDir), isDir: true}, nil
        }
        defer func() { mockOsStat = nil }()

		storage, err := factory.CreateDefaultStorage()
		require.NoError(t, err)
		assert.NotNil(t, storage)
		// This should be a Local storage instance.
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

// Note: To make the os.Stat patching cleaner for CreateDefaultStorage tests,
// the services.StorageFactory would ideally accept an osStat func as a parameter,
// or OsStat could be a package-level variable function in 'services' that can be swapped in tests.
// For this test, I'll assume we can add `var OsStat = os.Stat` to `storage_factory.go` (or a similar file in services package)
// and then patch `services.OsStat` in these tests.
// If `services.AzureBlobStorage` or `services.LocalFileStorage` structs are not exported,
// type assertions like `_, ok := storage.(*services.LocalFileStorage)` will not work from `services_test` package.
// The tests will rely on `NoError` and `NotNil` for type correctness in such cases.
