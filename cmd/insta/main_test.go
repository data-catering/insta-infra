package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractDataFiles(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test extraction of embedded files
	t.Run("extract embedded files", func(t *testing.T) {
		err := extractDataFiles(tempDir, embedFS)
		if err != nil {
			t.Fatalf("extractDataFiles failed: %v", err)
		}

		// Check if data directory structure is created
		dataDir := filepath.Join(tempDir, "data")
		if _, err := os.Stat(dataDir); os.IsNotExist(err) {
			t.Errorf("Expected data directory to be created, but it doesn't exist")
		}

		// Verify persist directories are not extracted
		serviceDirs, err := os.ReadDir(dataDir)
		if err != nil {
			t.Errorf("Failed to read data directory: %v", err)
			return
		}

		for _, svcDir := range serviceDirs {
			if !svcDir.IsDir() {
				continue
			}

			persistDir := filepath.Join(dataDir, svcDir.Name(), "persist")
			if _, err := os.Stat(persistDir); !os.IsNotExist(err) {
				t.Errorf("Persist directory should not exist: %s", persistDir)
			}
		}
	})
}

func TestCleanup(t *testing.T) {
	// Create temporary test directory
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test file in the temp directory
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create an App with the temp directory
	app := &App{
		instaDir: tempDir,
	}

	// Test cleanup
	if err := app.Cleanup(); err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Verify the directory was removed
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Errorf("Expected temp directory to be removed, but it still exists")
	}
}

func TestListServices(t *testing.T) {
	app := &App{}

	// Make sure the services are defined
	if len(Services) == 0 {
		t.Fatal("No services defined")
	}

	// Test listing services
	err := app.listServices()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

// Mock test for connectToService - can't fully test without a container runtime
func TestConnectToService(t *testing.T) {
	// Test with invalid service name
	t.Run("invalid service", func(t *testing.T) {
		app := &App{}
		err := app.connectToService("nonexistent-service")
		if err == nil {
			t.Error("Expected error for nonexistent service, got nil")
		}
	})

	// Test with empty service name
	t.Run("empty service name", func(t *testing.T) {
		app := &App{}
		err := app.connectToService("")
		if err == nil {
			t.Error("Expected error for empty service name, got nil")
		}
	})
}

// TestNewApp - partial test since we can't fully test the runtime detection
func TestNewApp(t *testing.T) {
	// This is a partial test since we can't fully test the runtime detection
	// and file extraction without mocking more dependencies

	// Just verify it doesn't panic
	t.Run("should not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("NewApp panicked: %v", r)
			}
		}()

		_, err := NewApp("")
		if err != nil {
			// It's okay to get an error if docker/podman isn't available
			// but we shouldn't panic
			t.Logf("NewApp returned error (expected if no runtime available): %v", err)
		}
	})
}
