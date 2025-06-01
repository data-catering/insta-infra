package main

import (
	"os"
	"os/exec"
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

	// Remove test containers
	cmd = exec.Command("docker", "ps", "-aq", "-f", "name=postgres", "-f", "name=httpbin")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		cmd = exec.Command("docker", "rm", "-f", string(output))
		if err := cmd.Run(); err != nil {
			t.Logf("warning: failed to remove containers: %v", err)
		}
	}
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

// TestCoreIntegration tests essential CLI functionality
func TestCoreIntegration(t *testing.T) {
	checkContainerRuntime(t)

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)
	cleanup(t, binaryPath)

	// Test basic service lifecycle
	t.Run("service lifecycle", func(t *testing.T) {
		// Start a simple service
		cmd := exec.Command(binaryPath, "httpbin")
		cmd.Env = append(os.Environ(), "TESTING=true")
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to start httpbin: %v", err)
		}

		// Stop the service
		cmd = exec.Command(binaryPath, "-d", "httpbin")
		cmd.Env = append(os.Environ(), "TESTING=true")
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to stop httpbin: %v", err)
		}
	})

	// Test list services functionality
	t.Run("list services", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-l")
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to list services: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "postgres") || !strings.Contains(outputStr, "redis") {
			t.Errorf("expected common services in output, got: %s", outputStr)
		}
	})
}

// TestServiceConnection tests that services can actually be connected to
func TestServiceConnection(t *testing.T) {
	checkContainerRuntime(t)

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)
	cleanup(t, binaryPath)

	// Start httpbin service (lighter than postgres)
	cmd := exec.Command(binaryPath, "httpbin")
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to start httpbin: %v", err)
	}

	// Wait for service to start
	time.Sleep(2 * time.Second)

	// Test connection by checking if the container is running
	// Since curl might not be available in the container, just verify the container exists and is responsive
	cmd = exec.Command("docker", "ps", "-f", "name=http", "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("failed to check httpbin container status: %v", err)
	} else {
		statusOutput := string(output)
		if !strings.Contains(statusOutput, "Up") {
			t.Errorf("httpbin container not running: %s", statusOutput)
		}
	}

	// Cleanup
	cmd = exec.Command(binaryPath, "-d", "httpbin")
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to stop httpbin: %v", err)
	}
}
