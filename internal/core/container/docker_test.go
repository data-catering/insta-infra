package container

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDockerRuntime_Name(t *testing.T) {
	docker := NewDockerRuntime()
	if docker.Name() != "docker" {
		t.Errorf("Expected name 'docker', got '%s'", docker.Name())
	}
}

func TestDockerRuntime_CheckAvailable_DockerNotInstalled(t *testing.T) {
	// This test will only pass if docker is not installed
	// We can't easily mock exec.LookPath, so we'll skip if docker is available
	if _, err := exec.LookPath("docker"); err == nil {
		t.Skip("Docker is installed, skipping test for docker not installed")
	}

	docker := NewDockerRuntime()
	err := docker.CheckAvailable()
	if err == nil {
		t.Error("Expected error when docker is not installed")
	}
	if !strings.Contains(err.Error(), "docker not found") {
		t.Errorf("Expected 'docker not found' error, got: %v", err)
	}
}

func TestDockerRuntime_CheckAvailable_DockerInstalled(t *testing.T) {
	// Skip if docker is not installed
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping test")
	}

	docker := NewDockerRuntime()
	err := docker.CheckAvailable()

	// We expect this to either succeed (if Docker daemon is running) or fail with daemon error
	if err != nil {
		// Check that it's a daemon error, not a "not found" error
		if strings.Contains(err.Error(), "docker not found") {
			t.Error("Docker is installed but CheckAvailable returned 'not found' error")
		}
		// It's okay if daemon is not running - that's expected in CI
		t.Logf("Docker daemon not running (expected in CI): %v", err)
	} else {
		t.Log("Docker is available and running")
	}
}

// Test the parsePortMappings utility function directly
func TestDockerRuntime_ParsePortMappings(t *testing.T) {
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

// Test the extractDependencies utility function directly
func TestDockerRuntime_ExtractDependencies(t *testing.T) {
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

// Test the getServiceNames utility function directly
func TestDockerRuntime_GetServiceNames(t *testing.T) {
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

// Test JSON marshaling of ComposeService
func TestDockerRuntime_ComposeServiceJSONMarshaling(t *testing.T) {
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
func TestDockerRuntime_ComposeConfigJSONMarshaling(t *testing.T) {
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

// Test ImagePullProgress struct
func TestDockerRuntime_ImagePullProgressStruct(t *testing.T) {
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
func TestDockerRuntime_ImagePullProgressError(t *testing.T) {
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

// Test DockerRuntime struct initialization
func TestDockerRuntime_StructInitialization(t *testing.T) {
	docker := NewDockerRuntime()

	if docker == nil {
		t.Fatal("NewDockerRuntime() returned nil")
	}

	if docker.Name() != "docker" {
		t.Errorf("Expected name 'docker', got '%s'", docker.Name())
	}

	// Test that the struct is properly initialized
	if docker.parsedComposeConfig != nil {
		t.Error("Expected parsedComposeConfig to be nil initially")
	}

	if docker.cachedComposeFilesKey != "" {
		t.Error("Expected cachedComposeFilesKey to be empty initially")
	}
}

// Test setDefaultEnvVars function
func TestDockerRuntime_SetDefaultEnvVars(t *testing.T) {
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

// Test parseDockerPullOutput function
func TestDockerRuntime_parseDockerPullOutput(t *testing.T) {
	docker := NewDockerRuntime()
	layerProgress := make(map[string]float64)
	totalLayers := 0

	tests := []struct {
		name           string
		line           string
		expectedStatus string
	}{
		{
			name:           "pulling from repository",
			line:           "v3.3.0: Pulling from ankane/blazer",
			expectedStatus: "starting",
		},
		{
			name:           "layer already exists",
			line:           "6e771e15690e: Already exists",
			expectedStatus: "downloading",
		},
		{
			name:           "layer pull complete",
			line:           "9521bbc382b8: Pull complete",
			expectedStatus: "downloading",
		},
		{
			name:           "pulling fs layer",
			line:           "9521bbc382b8: Pulling fs layer",
			expectedStatus: "downloading",
		},
		{
			name:           "downloading layer",
			line:           "9521bbc382b8: Downloading",
			expectedStatus: "downloading",
		},
		{
			name:           "download complete",
			line:           "Status: Downloaded newer image for ankane/blazer:v3.3.0",
			expectedStatus: "complete",
		},
		{
			name:           "image up to date",
			line:           "Status: Image is up to date for postgres:13",
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
			progress := docker.parseDockerPullOutput(tt.line, layerProgress, &totalLayers)

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

// Integration tests that require Docker to be available
func TestDockerRuntime_Integration(t *testing.T) {
	// Skip if docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping integration tests")
	}

	docker := NewDockerRuntime()

	t.Run("GetPortMappings with nonexistent container", func(t *testing.T) {
		_, err := docker.GetPortMappings("definitely-nonexistent-container-12345")
		if err == nil {
			t.Error("Expected error for nonexistent container")
		}
	})

	t.Run("GetContainerStatus with nonexistent container", func(t *testing.T) {
		status, err := docker.GetContainerStatus("definitely-nonexistent-container-12345")
		if err != nil {
			// If Docker daemon is not running, this will error, which is expected
			t.Logf("GetContainerStatus failed (expected if Docker daemon not running): %v", err)
			return
		}
		if status != "not_found" {
			t.Errorf("Expected status 'not_found', got '%s'", status)
		}
	})

	t.Run("CheckImageExists with nonexistent image", func(t *testing.T) {
		exists, err := docker.CheckImageExists("definitely-nonexistent-image:12345")
		if err != nil {
			t.Errorf("CheckImageExists should not error, got: %v", err)
		}
		if exists {
			t.Error("Expected image to not exist")
		}
	})

	t.Run("GetContainerLogs with nonexistent container", func(t *testing.T) {
		_, err := docker.GetContainerLogs("definitely-nonexistent-container-12345", 10)
		if err == nil {
			t.Error("Expected error for nonexistent container")
		}
	})

	t.Run("ExecInContainer with nonexistent container", func(t *testing.T) {
		err := docker.ExecInContainer("definitely-nonexistent-container-12345", "echo test", false)
		if err == nil {
			t.Error("Expected error for nonexistent container")
		}
	})
}

// Test compose file operations with temporary files
func TestDockerRuntime_ComposeFileOperations(t *testing.T) {
	// Skip if docker is not available
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available, skipping compose file tests")
	}

	docker := NewDockerRuntime()
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
		image, err := docker.GetImageInfo("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Docker daemon is not running, which is okay
			t.Logf("GetImageInfo failed (expected if Docker daemon not running): %v", err)
			return
		}
		if image != "hello-world" {
			t.Errorf("Expected image 'hello-world', got '%s'", image)
		}
	})

	t.Run("GetImageInfo with nonexistent service", func(t *testing.T) {
		_, err := docker.GetImageInfo("nonexistent-service", []string{composeFile})
		if err == nil {
			t.Error("Expected error for nonexistent service")
		}
	})

	t.Run("GetDependencies with valid compose file", func(t *testing.T) {
		deps, err := docker.GetDependencies("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Docker daemon is not running, which is okay
			t.Logf("GetDependencies failed (expected if Docker daemon not running): %v", err)
			return
		}
		if len(deps) != 1 || deps[0] != "dependency" {
			t.Errorf("Expected dependencies ['dependency'], got %v", deps)
		}
	})

	t.Run("GetContainerName with valid compose file", func(t *testing.T) {
		name, err := docker.GetContainerName("test-service", []string{composeFile})
		if err != nil {
			// This might fail if Docker daemon is not running, which is okay
			t.Logf("GetContainerName failed (expected if Docker daemon not running): %v", err)
			return
		}
		// Should return the service name as fallback
		if name != "test-service" {
			t.Errorf("Expected container name 'test-service', got '%s'", name)
		}
	})
}

// Test streaming operations
func TestDockerRuntime_StreamingOperations(t *testing.T) {
	docker := NewDockerRuntime()

	t.Run("StreamContainerLogs with stop signal", func(t *testing.T) {
		logChan := make(chan string, 10)
		stopChan := make(chan struct{})

		// Start streaming in a goroutine
		go func() {
			time.Sleep(50 * time.Millisecond)
			close(stopChan)
		}()

		// This should handle the stop signal gracefully
		err := docker.StreamContainerLogs("nonexistent-container", logChan, stopChan)

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
		err := docker.PullImageWithProgress("nonexistent:latest", progressChan, stopChan)

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
