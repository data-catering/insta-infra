package handlers

import (
	"os"
	"testing"
	"time"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// TestServiceStatusConsistency tests that different status checking methods return consistent results
// This is the core integration test for UI service management
func TestServiceStatusConsistency(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping integration test")
	}

	runtime := container.NewDockerRuntime()

	tempDir := t.TempDir()
	instaDir := tempDir + "/.insta"

	// Create the insta directory
	err := os.MkdirAll(instaDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create insta directory: %v", err)
	}

	createTestComposeFiles(t, instaDir)

	handler := NewServiceHandler(runtime, instaDir)

	// Test with redis running (lighter than postgres, no dependencies)
	t.Log("Starting redis for consistency test...")
	_ = handler.StopService("redis")
	time.Sleep(200 * time.Millisecond) // Brief wait for cleanup

	err = handler.StartService("redis", false)
	if err != nil {
		t.Fatalf("Failed to start redis: %v", err)
	}

	// Wait for startup with intelligent wait - redis starts much faster than postgres
	maxWait := 8 * time.Second
	interval := 500 * time.Millisecond
	start := time.Now()
	for time.Since(start) < maxWait {
		status, err := handler.GetServiceStatus("redis")
		if err == nil && status == "running" {
			t.Logf("Redis became running after %v", time.Since(start))
			break
		}
		time.Sleep(interval)
	}

	// Test all status checking methods return consistent results
	status1, err := handler.GetServiceStatus("redis")
	if err != nil {
		t.Fatalf("GetServiceStatus failed: %v", err)
	}

	allStatuses, err := handler.GetAllRunningServices()
	if err != nil {
		t.Fatalf("GetAllRunningServices failed: %v", err)
	}
	status2 := allStatuses["redis"].Status

	composeFiles := handler.getComposeFiles()
	status3, err := handler.getServiceStatusInternal("redis", composeFiles)
	if err != nil {
		t.Fatalf("getServiceStatusInternal failed: %v", err)
	}

	// All methods should return consistent results
	t.Logf("Status methods - GetServiceStatus: %s, GetAllRunningServices: %s, getServiceStatusInternal: %s", status1, status2, status3)

	if status1 != status2 {
		t.Errorf("Inconsistency: GetServiceStatus (%s) != GetAllRunningServices (%s)", status1, status2)
	}
	if status1 != status3 {
		t.Errorf("Inconsistency: GetServiceStatus (%s) != getServiceStatusInternal (%s)", status1, status3)
	}

	// All statuses should be "running" for a healthy redis
	if status1 != "running" {
		t.Errorf("Expected status to be 'running', got '%s'", status1)
	}

	// Cleanup
	_ = handler.StopService("redis")
}
