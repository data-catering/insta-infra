package container

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodmanRuntime_Name(t *testing.T) {
	podman := NewPodmanRuntime()
	if podman.Name() != "podman" {
		t.Errorf("Expected name 'podman', got '%s'", podman.Name())
	}
}

func TestPodmanRuntime_CheckAvailable_PodmanNotInstalled(t *testing.T) {
	// This test will only pass if podman is not installed
	// We can't easily mock exec.LookPath, so we'll skip if podman is available
	if _, err := exec.LookPath("podman"); err == nil {
		t.Skip("Podman is installed, skipping test for podman not installed")
	}

	podman := NewPodmanRuntime()
	err := podman.CheckAvailable()
	if err == nil {
		t.Error("Expected error when podman is not installed")
	}
	if !strings.Contains(err.Error(), "podman not found") {
		t.Errorf("Expected 'podman not found' error, got: %v", err)
	}
}

func TestPodmanRuntime_CheckAvailable_PodmanInstalled(t *testing.T) {
	// Skip if podman is not installed
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("Podman not available, skipping test")
	}

	podman := NewPodmanRuntime()
	err := podman.CheckAvailable()

	// We expect this to either succeed (if Podman machine is running) or fail with machine/daemon error
	if err != nil {
		// Check that it's a machine/daemon error, not a "not found" error
		if strings.Contains(err.Error(), "podman not found") {
			t.Error("Podman is installed but CheckAvailable returned 'not found' error")
		}
		// It's okay if machine is not running - that's expected in CI
		t.Logf("Podman machine not running (expected in CI): %v", err)
	} else {
		t.Log("Podman is available and running")
	}
}

// Test the parsePodmanPullOutput function
func TestPodmanRuntime_parsePodmanPullOutput(t *testing.T) {
	podman := NewPodmanRuntime()

	tests := []struct {
		name           string
		line           string
		expectedStatus string
	}{
		{
			name:           "trying to pull",
			line:           "Trying to pull docker.io/library/redis:alpine...",
			expectedStatus: "starting",
		},
		{
			name:           "getting image source signatures",
			line:           "Getting image source signatures",
			expectedStatus: "downloading",
		},
		{
			name:           "copying blob done",
			line:           "Copying blob 8bc3a26b84da done",
			expectedStatus: "downloading",
		},
		{
			name:           "writing manifest",
			line:           "Writing manifest to image destination",
			expectedStatus: "downloading",
		},
		{
			name:           "storing signatures",
			line:           "Storing signatures",
			expectedStatus: "complete",
		},
		{
			name:           "unrecognized line",
			line:           "Some random output",
			expectedStatus: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			progress := podman.parsePodmanPullOutput(tt.line)

			if progress.Status != tt.expectedStatus {
				t.Errorf("Expected status '%s', got '%s'", tt.expectedStatus, progress.Status)
			}

			// For downloading status, check that progress is reasonable
			if progress.Status == "downloading" && progress.Progress < 0 {
				t.Errorf("Expected non-negative progress, got %f", progress.Progress)
			}

			// For complete status, check that progress is 100
			if progress.Status == "complete" && progress.Progress != 100.0 {
				t.Errorf("Expected progress 100.0 for complete status, got %f", progress.Progress)
			}
		})
	}
}

// Test PodmanRuntime struct initialization
func TestPodmanRuntime_StructInitialization(t *testing.T) {
	podman := NewPodmanRuntime()

	if podman == nil {
		t.Fatal("NewPodmanRuntime() returned nil")
	}

	if podman.Name() != "podman" {
		t.Errorf("Expected name 'podman', got '%s'", podman.Name())
	}

	// Test that the struct is properly initialized
	if podman.parsedComposeConfig != nil {
		t.Error("Expected parsedComposeConfig to be nil initially")
	}

	if podman.cachedComposeFilesKey != "" {
		t.Error("Expected cachedComposeFilesKey to be empty initially")
	}
}

// Test containerExistsAnywhere function behavior
func TestPodmanRuntime_containerExistsAnywhere(t *testing.T) {
	podman := NewPodmanRuntime()

	// Test with a definitely non-existent container
	exists := podman.containerExistsAnywhere("definitely-nonexistent-container-12345")
	if exists {
		t.Error("Expected container to not exist")
	}
}

// Test JSON marshaling reusing existing test utilities
func TestPodmanRuntime_ComposeServiceJSONMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		service ComposeService
		wantErr bool
	}{
		{
			name: "service with dependencies and image",
			service: ComposeService{
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{
					"db": {Condition: "service_started"},
				},
				Image: "web:latest",
			},
			wantErr: false,
		},
		{
			name: "service with only image",
			service: ComposeService{
				Image: "postgres:13",
			},
			wantErr: false,
		},
		{
			name:    "empty service",
			service: ComposeService{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Try to unmarshal back to verify structure
				var unmarshaled ComposeService
				if err := json.Unmarshal(data, &unmarshaled); err != nil {
					t.Errorf("Failed to unmarshal back: %v", err)
				}
			}
		})
	}
}

// Test JSON marshaling of ComposeConfig
func TestPodmanRuntime_ComposeConfigJSONMarshaling(t *testing.T) {
	config := ComposeConfig{
		Services: map[string]ComposeService{
			"web": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{
					"db": {Condition: "service_started"},
				},
				Image: "web:latest",
			},
			"db": {
				Image: "postgres:13",
			},
		},
	}

	data, err := json.Marshal(config)
	if err != nil {
		t.Errorf("json.Marshal() error = %v", err)
		return
	}

	var unmarshaled ComposeConfig
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal back: %v", err)
		return
	}

	if len(unmarshaled.Services) != len(config.Services) {
		t.Errorf("Expected %d services after unmarshal, got %d", len(config.Services), len(unmarshaled.Services))
	}
}

// Test ImagePullProgress struct (reusing from docker tests)
func TestPodmanRuntime_ImagePullProgressStruct(t *testing.T) {
	progress := ImagePullProgress{
		Status:       "downloading",
		Progress:     50.0,
		CurrentLayer: "layer1",
		TotalLayers:  5,
		Downloaded:   1024,
		Total:        2048,
		Speed:        "1.2 MB/s",
		ETA:          "30s",
		Error:        "",
	}

	// Test JSON marshaling
	data, err := json.Marshal(progress)
	if err != nil {
		t.Errorf("Failed to marshal ImagePullProgress: %v", err)
	}

	var unmarshaled ImagePullProgress
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal ImagePullProgress: %v", err)
	}

	if unmarshaled.Status != progress.Status {
		t.Errorf("Expected status %s, got %s", progress.Status, unmarshaled.Status)
	}
	if unmarshaled.Progress != progress.Progress {
		t.Errorf("Expected progress %f, got %f", progress.Progress, unmarshaled.Progress)
	}
}

// Test ImagePullProgress with error
func TestPodmanRuntime_ImagePullProgressError(t *testing.T) {
	progress := ImagePullProgress{
		Status: "error",
		Error:  "Failed to pull image",
	}

	if progress.Status != "error" {
		t.Errorf("Expected status 'error', got '%s'", progress.Status)
	}
	if progress.Error != "Failed to pull image" {
		t.Errorf("Expected error 'Failed to pull image', got '%s'", progress.Error)
	}
}

// Test utility functions using existing test utilities
func TestPodmanRuntime_ParsePortMappings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:     "single port mapping",
			input:    "5432/tcp -> 0.0.0.0:5432",
			expected: map[string]string{"5432/tcp": "5432"},
		},
		{
			name:     "multiple port mappings",
			input:    "5432/tcp -> 0.0.0.0:5432\n8080/tcp -> 0.0.0.0:8080",
			expected: map[string]string{"5432/tcp": "5432", "8080/tcp": "8080"},
		},
		{
			name:     "port mapping with different host port",
			input:    "5432/tcp -> 0.0.0.0:15432",
			expected: map[string]string{"5432/tcp": "15432"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:     "port mapping with IPv6",
			input:    "5432/tcp -> [::]:5432",
			expected: map[string]string{"5432/tcp": "5432"},
		},
		{
			name:     "malformed input - no arrow",
			input:    "5432/tcp 0.0.0.0:5432",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePortMappings(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d mappings, got %d", len(tt.expected), len(result))
			}

			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("Expected %s -> %s, got %s -> %s", k, v, k, result[k])
				}
			}
		})
	}
}

// Test extractDependencies using existing test utilities
func TestPodmanRuntime_ExtractDependencies(t *testing.T) {
	tests := []struct {
		name         string
		config       ComposeConfig
		serviceName  string
		expectedDeps []string
	}{
		{
			name: "service with depends_on dependencies",
			config: ComposeConfig{
				Services: map[string]ComposeService{
					"web": {
						DependsOn: map[string]struct {
							Condition string `json:"condition"`
						}{
							"db":    {Condition: "service_started"},
							"redis": {Condition: "service_started"},
						},
					},
				},
			},
			serviceName:  "web",
			expectedDeps: []string{"db", "redis"},
		},
		{
			name: "service with no dependencies",
			config: ComposeConfig{
				Services: map[string]ComposeService{
					"db": {},
				},
			},
			serviceName:  "db",
			expectedDeps: []string{},
		},
		{
			name:         "non-existent service",
			config:       ComposeConfig{Services: map[string]ComposeService{}},
			serviceName:  "nonexistent",
			expectedDeps: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractDependencies(tt.config, tt.serviceName)

			if len(result) != len(tt.expectedDeps) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDeps), len(result))
			}

			// Convert to maps for easier comparison
			expectedMap := make(map[string]bool)
			for _, dep := range tt.expectedDeps {
				expectedMap[dep] = true
			}

			for _, dep := range result {
				if !expectedMap[dep] {
					t.Errorf("Unexpected dependency: %s", dep)
				}
			}
		})
	}
}

// Test getServiceNames using existing test utilities
func TestPodmanRuntime_GetServiceNames(t *testing.T) {
	tests := []struct {
		name     string
		config   ComposeConfig
		expected []string
	}{
		{
			name: "multiple services",
			config: ComposeConfig{
				Services: map[string]ComposeService{
					"web":   {},
					"db":    {},
					"redis": {},
				},
			},
			expected: []string{"web", "db", "redis"},
		},
		{
			name: "single service",
			config: ComposeConfig{
				Services: map[string]ComposeService{
					"web": {},
				},
			},
			expected: []string{"web"},
		},
		{
			name:     "no services",
			config:   ComposeConfig{Services: map[string]ComposeService{}},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getServiceNames(&tt.config)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d services, got %d", len(tt.expected), len(result))
			}

			// Convert to maps for easier comparison (order doesn't matter)
			expectedMap := make(map[string]bool)
			for _, service := range tt.expected {
				expectedMap[service] = true
			}

			for _, service := range result {
				if !expectedMap[service] {
					t.Errorf("Unexpected service: %s", service)
				}
			}
		})
	}
}

// Test setDefaultEnvVars function (reusing from docker tests)
func TestPodmanRuntime_SetDefaultEnvVars(t *testing.T) {
	// Save original environment
	restore := SaveAndRestoreEnvVars([]string{"DB_USER", "DB_USER_PASSWORD", "MYSQL_USER", "MYSQL_PASSWORD"})
	defer restore()

	// Clear environment variables
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_USER_PASSWORD")
	os.Unsetenv("MYSQL_USER")
	os.Unsetenv("MYSQL_PASSWORD")

	// Call setDefaultEnvVars
	setDefaultEnvVars()

	// Check that environment variables are set
	if os.Getenv("DB_USER") != "root" {
		t.Errorf("Expected DB_USER to be 'root', got '%s'", os.Getenv("DB_USER"))
	}

	if os.Getenv("DB_USER_PASSWORD") != "root" {
		t.Errorf("Expected DB_USER_PASSWORD to be 'root', got '%s'", os.Getenv("DB_USER_PASSWORD"))
	}

	if os.Getenv("MYSQL_USER") != "root" {
		t.Errorf("Expected MYSQL_USER to be 'root', got '%s'", os.Getenv("MYSQL_USER"))
	}

	if os.Getenv("MYSQL_PASSWORD") != "root" {
		t.Errorf("Expected MYSQL_PASSWORD to be 'root', got '%s'", os.Getenv("MYSQL_PASSWORD"))
	}
}

// Integration tests that require Podman to be available
func TestPodmanRuntime_Integration(t *testing.T) {
	// Skip if podman is not available
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("Podman not available, skipping integration tests")
	}

	podman := NewPodmanRuntime()

	t.Run("GetPortMappings with nonexistent container", func(t *testing.T) {
		_, err := podman.GetPortMappings("definitely-nonexistent-container-12345")
		if err == nil {
			t.Error("Expected error for nonexistent container")
		}
	})

	t.Run("GetContainerStatus with nonexistent container", func(t *testing.T) {
		status, err := podman.GetContainerStatus("definitely-nonexistent-container-12345")
		if err != nil {
			// If Podman machine is not running, this will error, which is expected
			t.Logf("GetContainerStatus failed (expected if Podman machine not running): %v", err)
			return
		}
		if status != "not_found" {
			t.Errorf("Expected status 'not_found', got '%s'", status)
		}
	})

	t.Run("CheckImageExists with nonexistent image", func(t *testing.T) {
		exists, err := podman.CheckImageExists("definitely-nonexistent-image:12345")
		if err != nil {
			t.Errorf("CheckImageExists should not error, got: %v", err)
		}
		if exists {
			t.Error("Expected image to not exist")
		}
	})

	t.Run("GetContainerLogs with nonexistent container", func(t *testing.T) {
		_, err := podman.GetContainerLogs("definitely-nonexistent-container-12345", 10)
		if err == nil {
			t.Error("Expected error for nonexistent container")
		}
	})

	t.Run("ExecInContainer with nonexistent container", func(t *testing.T) {
		err := podman.ExecInContainer("definitely-nonexistent-container-12345", "echo test", false)
		if err == nil {
			t.Error("Expected error for nonexistent container")
		}
	})
}

// Test compose file operations with temporary files
func TestPodmanRuntime_ComposeFileOperations(t *testing.T) {
	// Skip if podman is not available
	if _, err := exec.LookPath("podman"); err != nil {
		t.Skip("Podman not available, skipping compose file tests")
	}

	podman := NewPodmanRuntime()
	tempDir := t.TempDir()

	// Create a test compose file
	composeFile := filepath.Join(tempDir, "docker-compose.yaml")
	composeContent := `
version: '3.8'
services:
  test-service:
    image: hello-world
    depends_on:
      - dependency
  dependency:
    image: alpine:latest
`

	err := os.WriteFile(composeFile, []byte(composeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test compose file: %v", err)
	}

	t.Run("GetImageInfo with valid compose file", func(t *testing.T) {
		image, err := podman.GetImageInfo("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Podman machine is not running, which is okay
			t.Logf("GetImageInfo failed (expected if Podman machine not running): %v", err)
			return
		}
		if image != "hello-world" {
			t.Errorf("Expected image 'hello-world', got '%s'", image)
		}
	})

	t.Run("GetImageInfo with nonexistent service", func(t *testing.T) {
		_, err := podman.GetImageInfo("nonexistent-service", []string{composeFile})
		if err == nil {
			t.Error("Expected error for nonexistent service")
		}
	})

	t.Run("GetDependencies with valid compose file", func(t *testing.T) {
		deps, err := podman.GetDependencies("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Podman machine is not running, which is okay
			t.Logf("GetDependencies failed (expected if Podman machine not running): %v", err)
			return
		}
		if len(deps) != 1 || deps[0] != "dependency" {
			t.Errorf("Expected dependencies ['dependency'], got %v", deps)
		}
	})

	t.Run("GetContainerName with valid compose file", func(t *testing.T) {
		name, err := podman.GetContainerName("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Podman machine is not running, which is okay
			t.Logf("GetContainerName failed (expected if Podman machine not running): %v", err)
			return
		}
		// Should return the service name as fallback
		if name != "test-service" {
			t.Errorf("Expected container name 'test-service', got '%s'", name)
		}
	})

	t.Run("GetAllDependenciesRecursive with valid compose file", func(t *testing.T) {
		deps, err := podman.GetAllDependenciesRecursive("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Podman machine is not running, which is okay
			t.Logf("GetAllDependenciesRecursive failed (expected if Podman machine not running): %v", err)
			return
		}
		if len(deps) != 1 || deps[0] != "dependency" {
			t.Errorf("Expected dependencies ['dependency'], got %v", deps)
		}
	})
}

// Test streaming operations
func TestPodmanRuntime_StreamingOperations(t *testing.T) {
	podman := NewPodmanRuntime()

	t.Run("StreamContainerLogs with stop signal", func(t *testing.T) {
		logChan := make(chan string, 10)
		stopChan := make(chan struct{})

		// Start streaming in a goroutine
		go func() {
			time.Sleep(50 * time.Millisecond)
			close(stopChan)
		}()

		// This should handle the stop signal gracefully
		err := podman.StreamContainerLogs("nonexistent-container", logChan, stopChan)

		// We expect an error since the container doesn't exist, but it should handle the stop signal
		if err == nil {
			t.Log("StreamContainerLogs completed without error")
		} else {
			t.Logf("StreamContainerLogs failed as expected: %v", err)
		}
	})

	t.Run("PullImageWithProgress with stop signal", func(t *testing.T) {
		progressChan := make(chan ImagePullProgress, 10)
		stopChan := make(chan struct{})

		// Start pulling in a goroutine and stop it quickly
		go func() {
			time.Sleep(50 * time.Millisecond)
			close(stopChan)
		}()

		// This should handle the stop signal gracefully
		err := podman.PullImageWithProgress("nonexistent:latest", progressChan, stopChan)

		// We expect this to either fail (image doesn't exist) or be cancelled
		if err == nil {
			t.Log("PullImageWithProgress completed without error")
		} else {
			t.Logf("PullImageWithProgress failed as expected: %v", err)
		}

		// Check if we received any progress updates
		close(progressChan)
		progressCount := 0
		for range progressChan {
			progressCount++
		}
		t.Logf("Received %d progress updates", progressCount)
	})
}

// Test Podman-specific functionality
func TestPodmanRuntime_PodmanSpecific(t *testing.T) {
	podman := NewPodmanRuntime()

	t.Run("GetContainerName with different naming patterns", func(t *testing.T) {
		// Test the candidate name generation logic
		name, err := podman.GetContainerName("test-service", []string{})
		if err != nil {
			t.Errorf("GetContainerName should not error for basic case: %v", err)
		}
		// Should return the service name as fallback when no containers exist
		if name != "test-service" {
			t.Errorf("Expected container name 'test-service', got '%s'", name)
		}
	})

	t.Run("parsePodmanPullOutput with various progress states", func(t *testing.T) {
		testLines := []string{
			"Trying to pull docker.io/library/redis:alpine...",
			"Getting image source signatures",
			"Copying blob 8bc3a26b84da done",
			"Writing manifest to image destination",
			"Storing signatures",
		}

		for _, line := range testLines {
			progress := podman.parsePodmanPullOutput(line)
			if progress.Status == "" && line != "" {
				t.Logf("Line '%s' produced empty status (may be expected)", line)
			}
		}
	})
}

func TestPodmanRuntimeCheckAvailableWithCustomPath(t *testing.T) {
	// Test that Podman detection works with custom path via environment variable

	// Save original environment
	originalPath := os.Getenv("INSTA_PODMAN_PATH")
	defer func() {
		if originalPath == "" {
			os.Unsetenv("INSTA_PODMAN_PATH")
		} else {
			os.Setenv("INSTA_PODMAN_PATH", originalPath)
		}
	}()

	// Create a temporary podman binary
	tempDir := t.TempDir()
	podmanPath := filepath.Join(tempDir, "podman")

	// Create a mock podman script that responds to basic commands
	podmanScript := `#!/bin/bash
case "$1" in
    "version")
        echo "4.0.0"
        exit 0
        ;;
    "machine")
        if [ "$2" = "list" ]; then
            echo "test-machine"
            exit 0
        fi
        ;;
    "compose")
        if [ "$2" = "version" ]; then
            echo "Podman Compose version"
            exit 0
        fi
        ;;
    "info")
        echo '{"host":{"security":{"rootless":false}}}'
        exit 0
        ;;
esac
exit 1
`

	err := os.WriteFile(podmanPath, []byte(podmanScript), 0755)
	require.NoError(t, err)

	// Set custom Podman path
	os.Setenv("INSTA_PODMAN_PATH", podmanPath)

	// Test that findBinaryInCommonPaths finds the custom path
	foundPath := findBinaryInCommonPaths("podman", getCommonPodmanPaths())
	assert.Equal(t, podmanPath, foundPath)

	// Test that the found podman binary works
	cmd := exec.Command(foundPath, "version", "--format", "{{.Version}}")
	err = cmd.Run()
	assert.NoError(t, err)
}
