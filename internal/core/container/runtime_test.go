package container

import (
	"os/exec"
	"testing"
)

// MockRuntime implementation moved to test_utils.go

func TestNewProvider(t *testing.T) {
	provider := NewProvider()
	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}
	if len(provider.runtimes) != 2 {
		t.Fatalf("Expected 2 runtimes, got %d", len(provider.runtimes))
	}
}

func TestProviderDetectRuntime(t *testing.T) {
	tests := []struct {
		name         string
		runtimes     []Runtime
		expectError  bool
		expectedName string
	}{
		{
			name:        "no runtimes available",
			runtimes:    []Runtime{NewMockRuntime("docker", false), NewMockRuntime("podman", false)},
			expectError: true,
		},
		{
			name:         "docker available",
			runtimes:     []Runtime{NewMockRuntime("docker", true), NewMockRuntime("podman", false)},
			expectError:  false,
			expectedName: "docker",
		},
		{
			name:         "podman available",
			runtimes:     []Runtime{NewMockRuntime("docker", false), NewMockRuntime("podman", true)},
			expectError:  false,
			expectedName: "podman",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := CreateTestProvider(tt.runtimes...)
			err := provider.DetectRuntime()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
			if !tt.expectError && provider.selected.Name() != tt.expectedName {
				t.Errorf("Expected %s runtime, got %s", tt.expectedName, provider.selected.Name())
			}
		})
	}
}

func TestDockerCheckAvailable(t *testing.T) {
	// Skip if docker is not installed
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Skip("Docker not available, skipping test")
	}

	docker := NewDockerRuntime()
	err = docker.CheckAvailable()
	if err != nil {
		// Skip if Docker daemon is not running
		t.Skipf("Docker daemon not running, skipping test: %v", err)
	}
}

func TestDockerComposeOperations(t *testing.T) {
	// Use enhanced mock runtime for compose operations
	mock := NewMockRuntime("docker", true).WithPortMappings(map[string]string{"5432/tcp": "5432"})

	// Test ComposeUp
	err := mock.ComposeUp([]string{"docker-compose.yaml"}, []string{"service1"}, false)
	if err != nil {
		t.Fatalf("Unexpected error in ComposeUp: %v", err)
	}

	// Test ComposeDown
	err = mock.ComposeDown([]string{"docker-compose.yaml"}, []string{"service1"})
	if err != nil {
		t.Fatalf("Unexpected error in ComposeDown: %v", err)
	}
}

func TestExecInContainer(t *testing.T) {
	// Use mock runtime for exec operation
	mock := NewMockRuntime("docker", true)

	// Test ExecInContainer
	err := mock.ExecInContainer("container1", "echo hello", false)
	if err != nil {
		t.Fatalf("Unexpected error in ExecInContainer: %v", err)
	}
}

func TestPodmanCheckAvailable(t *testing.T) {
	// Skip if podman is not installed
	_, err := exec.LookPath("podman")
	if err != nil {
		t.Skip("Podman not available, skipping test")
	}

	podman := NewPodmanRuntime()
	err = podman.CheckAvailable()
	if err != nil {
		// Skip if Podman is not properly configured
		t.Skipf("Podman not properly configured, skipping test: %v", err)
	}
}

func TestComposeFilesHandling(t *testing.T) {
	mockRuntime := NewMockRuntime("docker", true)
	composeFiles := []string{"docker-compose.yaml", "docker-compose-persist.yaml"}
	services := []string{"postgres", "mysql"}

	err := mockRuntime.ComposeUp(composeFiles, services, true)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
