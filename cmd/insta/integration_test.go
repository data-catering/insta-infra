package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binaryPath := "./insta"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	return binaryPath
}

func cleanup(t *testing.T, binaryPath string) {
	t.Helper()

	// Stop all services
	cmd := exec.Command(binaryPath, "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Just log the error but continue with cleanup
		t.Logf("warning: failed to stop services during cleanup: %v", err)
	}

	// Remove all containers with name containing "postgres" or "httpbin"
	cmd = exec.Command("docker", "ps", "-aq", "-f", "name=postgres", "-f", "name=httpbin")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		cmd = exec.Command("docker", "rm", "-f", string(output))
		if err := cmd.Run(); err != nil {
			t.Logf("warning: failed to remove containers: %v", err)
		}
	}

	// Wait for containers to be fully removed
	time.Sleep(2 * time.Second)
}

func checkContainerRuntime(t *testing.T) {
	t.Helper()

	// Check for Docker
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "info")
		if err := cmd.Run(); err == nil {
			return
		}
	}

	// Check for Podman
	if _, err := exec.LookPath("podman"); err == nil {
		cmd := exec.Command("podman", "info")
		if err := cmd.Run(); err == nil {
			return
		}
	}

	t.Skip("No container runtime available, skipping integration tests")
}

func TestInstaBinary(t *testing.T) {
	checkContainerRuntime(t)

	// Build binary if it doesn't exist
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Clean up before test
	cleanup(t, binaryPath)

	// Test cases
	tests := []struct {
		name    string
		command []string
		wantErr bool
	}{
		{
			name:    "list services",
			command: []string{"-l"},
			wantErr: false,
		},
		{
			name:    "start httpbin",
			command: []string{"httpbin"},
			wantErr: false,
		},
		{
			name:    "start postgres with persistence",
			command: []string{"-p", "postgres"},
			wantErr: false,
		},
		{
			name:    "stop httpbin",
			command: []string{"-d", "httpbin"},
			wantErr: false,
		},
		{
			name:    "stop all services",
			command: []string{"-d"},
			wantErr: false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.command...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), "TESTING=true")

			err := cmd.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("command %v failed: %v", tt.command, err)
			}

			// Add a small delay between commands to allow services to start/stop
			time.Sleep(2 * time.Second)
		})
	}
}

func TestDataPersistence(t *testing.T) {
	checkContainerRuntime(t)

	// Build binary if it doesn't exist
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Clean up before test
	cleanup(t, binaryPath)

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	// Start a service with persistence
	cmd := exec.Command(binaryPath, "-p", "postgres")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}

	// Wait for service to start
	time.Sleep(5 * time.Second)

	// Verify data directory was created
	persistDir := filepath.Join(homeDir, ".insta", "data", "postgres", "persist")
	if _, err := os.Stat(persistDir); err != nil {
		t.Errorf("persist directory not created: %v", err)
	}

	// Stop the service
	cmd = exec.Command(binaryPath, "-d", "postgres")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to stop postgres: %v", err)
	}
}

func TestServiceConnection(t *testing.T) {
	checkContainerRuntime(t)

	// Build binary if it doesn't exist
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Clean up before test
	cleanup(t, binaryPath)

	// Start postgres service
	cmd := exec.Command(binaryPath, "postgres")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to start postgres: %v", err)
	}

	// Wait for service to start
	time.Sleep(10 * time.Second)

	// Test connecting to postgres using a non-interactive command with correct credentials
	cmd = exec.Command(binaryPath, "-c", "postgres", "--", "psql", "-U", "postgres", "-d", "postgres", "-c", "SELECT 1;")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "TESTING=true", "PGPASSWORD=postgres")
	if err := cmd.Run(); err != nil {
		t.Errorf("failed to connect to postgres: %v", err)
	}

	// Stop postgres
	cmd = exec.Command(binaryPath, "-d", "postgres")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to stop postgres: %v", err)
	}
}

func TestErrorHandling(t *testing.T) {
	checkContainerRuntime(t)

	// Build binary if it doesn't exist
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Clean up before test
	cleanup(t, binaryPath)

	tests := []struct {
		name    string
		command []string
		wantErr bool
	}{
		{
			name:    "connect to non-existent service",
			command: []string{"-c", "non-existent-service"},
			wantErr: true,
		},
		{
			name:    "stop non-existent service",
			command: []string{"-d", "non-existent-service"},
			wantErr: true,
		},
		{
			name:    "start non-existent service",
			command: []string{"non-existent-service"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.command...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Env = append(os.Environ(), "TESTING=true")

			err := cmd.Run()
			if (err != nil) != tt.wantErr {
				t.Errorf("command %v should have failed: %v", tt.command, err)
			}
		})
	}
}

// TestIntegration_KeycloakConnectionTable verifies the connection info table output for keycloak.
func TestIntegration_KeycloakConnectionTable(t *testing.T) {
	checkContainerRuntime(t)

	// Build binary if it doesn't exist
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Clean up before test
	cleanup(t, binaryPath)

	// Start keycloak and capture output
	cmd := exec.Command(binaryPath, "keycloak")
	cmd.Env = append(os.Environ(), "TESTING=true")
	// Use CombinedOutput to capture stdout and stderr
	outputBytes, err := cmd.CombinedOutput()
	output := string(outputBytes)
	if err != nil {
		t.Fatalf("failed to start keycloak: %v\nOutput:\n%s", err, output)
	}

	// --- Assertions ---
	// Check for table header and footer presence
	if !strings.Contains(output, "Connection Information Table") {
		t.Errorf("Output missing table header. Got:\n%s", output)
	}
	if !strings.Contains(output, "└─────") { // Check for part of the footer
		t.Errorf("Output missing table footer. Got:\n%s", output)
	}

	// Define expected connection details
	expectedLines := []string{
		"│ keycloak                │ keycloak:8080                │ localhost:8082       │ host.docker.internal:8082    │ admin      │ admin      │",
		"│ postgres                │ postgres:5432                │ localhost:5432       │ host.docker.internal:5432    │ postgres   │ postgres   │",
	}

	// Check if each expected line is present in the output
	for _, expectedLine := range expectedLines {
		if !strings.Contains(output, expectedLine) {
			t.Errorf("Output does not contain expected connection info line.\nExpected:\n%s\nGot:\n%s", expectedLine, output)
		}
	}

	// Stop the service
	cmd = exec.Command(binaryPath, "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to stop keycloak and postgres: %v", err)
	}
}
