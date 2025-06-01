package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// Performance test to measure expensive operations
func TestPerformance_ExpensiveOperations(t *testing.T) {
	// Skip if no docker available
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	// Setup test directory
	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)

	// Create test compose files
	createTestComposeFiles(t, instaDir)

	// Initialize container runtime
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	t.Run("BaseHandler_fetchRunningContainers", func(t *testing.T) {
		baseHandler := NewBaseHandler(runtime, instaDir)

		start := time.Now()
		containers, err := baseHandler.getCurrentContainers()
		duration := time.Since(start)

		if err != nil {
			t.Logf("fetchRunningContainers failed (expected if no services running): %v", err)
		}

		t.Logf("fetchRunningContainers took %v, found %d containers", duration, len(containers))

		// Flag if it takes more than 2 seconds
		if duration > 2*time.Second {
			t.Logf("WARNING: fetchRunningContainers is slow (%v)", duration)
		}
	})

	t.Run("ContainerRuntime_GetContainerName_Performance", func(t *testing.T) {
		composeFiles := []string{
			filepath.Join(instaDir, "docker-compose.yaml"),
			filepath.Join(instaDir, "docker-compose-persist.yaml"),
		}

		// Test GetContainerName for multiple services
		testServices := []string{"postgres", "redis", "mysql", "grafana", "superset"}

		totalDuration := time.Duration(0)
		for _, serviceName := range testServices {
			start := time.Now()
			containerName, err := runtime.GetContainerName(serviceName, composeFiles)
			duration := time.Since(start)
			totalDuration += duration

			t.Logf("GetContainerName(%s) took %v, result: %s", serviceName, duration, containerName)
			if err != nil {
				t.Logf("  Error: %v", err)
			}

			// Flag if it takes more than 200ms per service
			if duration > 200*time.Millisecond {
				t.Logf("WARNING: GetContainerName(%s) is slow (%v)", serviceName, duration)
			}
		}

		avgDuration := totalDuration / time.Duration(len(testServices))
		t.Logf("Average GetContainerName duration: %v", avgDuration)
	})

	t.Run("ContainerRuntime_GetPortMappings_Performance", func(t *testing.T) {
		composeFiles := []string{
			filepath.Join(instaDir, "docker-compose.yaml"),
			filepath.Join(instaDir, "docker-compose-persist.yaml"),
		}

		// First get container names, then test port mappings
		testServices := []string{"postgres", "redis", "mysql"}

		for _, serviceName := range testServices {
			containerName, err := runtime.GetContainerName(serviceName, composeFiles)
			if err != nil {
				t.Logf("Skipping %s: %v", serviceName, err)
				continue
			}

			start := time.Now()
			portMappings, err := runtime.GetPortMappings(containerName)
			duration := time.Since(start)

			t.Logf("GetPortMappings(%s) took %v", containerName, duration)
			if err != nil {
				t.Logf("  Error: %v", err)
			} else {
				t.Logf("  Found %d port mappings", len(portMappings))
			}

			// Flag if it takes more than 300ms per container
			if duration > 300*time.Millisecond {
				t.Logf("WARNING: GetPortMappings(%s) is slow (%v)", containerName, duration)
			}
		}
	})
}

// Test to measure the impact of avoiding expensive GetContainerName calls
func TestPerformance_ServiceToContainerMapping(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	composeFiles := []string{
		filepath.Join(instaDir, "docker-compose.yaml"),
		filepath.Join(instaDir, "docker-compose-persist.yaml"),
	}

	t.Run("Traditional_GetContainerName_Approach", func(t *testing.T) {
		// Simulate the old approach: call GetContainerName for each service
		testServices := []string{"postgres", "redis", "mysql", "grafana", "superset"}

		start := time.Now()
		containerNames := make(map[string]string)
		for _, serviceName := range testServices {
			containerName, err := runtime.GetContainerName(serviceName, composeFiles)
			if err == nil {
				containerNames[serviceName] = containerName
			}
		}
		duration := time.Since(start)

		t.Logf("Traditional approach: %d services took %v", len(testServices), duration)
		t.Logf("Found %d container names", len(containerNames))
	})

	t.Run("Optimized_DirectPS_Approach", func(t *testing.T) {
		// Simulate the new approach: get all running containers at once
		start := time.Now()
		runningContainers, err := runtime.GetAllContainerStatuses()
		if err != nil {
			t.Logf("GetAllContainerStatuses failed: %v", err)
		}
		duration := time.Since(start)

		t.Logf("Optimized approach: got %d containers in %v", len(runningContainers), duration)

		// Show the performance difference
		if duration < 50*time.Millisecond {
			t.Logf("SUCCESS: Optimized approach is very fast (<%v)", 50*time.Millisecond)
		}
	})
}

// Test connection handler performance
func TestPerformance_ConnectionHandler(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	handler := NewConnectionHandler(runtime, instaDir)

	t.Run("GetServiceConnectionInfo_Performance", func(t *testing.T) {
		testServices := []string{"postgres", "redis", "mysql", "grafana", "superset"}

		for _, serviceName := range testServices {
			start := time.Now()
			connInfo, err := handler.GetServiceConnectionInfo(serviceName)
			duration := time.Since(start)

			t.Logf("GetServiceConnectionInfo(%s) took %v", serviceName, duration)
			if err != nil {
				t.Logf("  Error: %v", err)
			} else {
				t.Logf("  Available: %v", connInfo.Available)
			}

			// Connection info should be fast (under 100ms)
			if duration > 100*time.Millisecond {
				t.Logf("WARNING: GetServiceConnectionInfo(%s) is slow (%v)", serviceName, duration)
			}
		}
	})
}

// Test to measure UI startup performance - this is what causes "loading services" delay
func TestPerformance_UIStartupLoadingServices(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	t.Run("GetAllServicesWithStatusAndDependencies_UIStartup", func(t *testing.T) {
		serviceHandler := NewServiceHandler(runtime, instaDir)

		start := time.Now()
		serviceDetails, err := serviceHandler.GetAllServicesWithStatusAndDependencies()
		duration := time.Since(start)

		if err != nil {
			t.Logf("GetAllServicesWithStatusAndDependencies failed: %v", err)
		}

		t.Logf("UI Startup: GetAllServicesWithStatusAndDependencies took %v for %d services",
			duration, len(serviceDetails))

		// This should be fast (under 1 second) for good UX
		if duration > 1*time.Second {
			t.Logf("WARNING: UI startup is slow (%v) - users will see 'loading services' for too long", duration)
		}

		// Log some sample results
		for i, detail := range serviceDetails {
			if i < 5 { // Show first 5 services
				t.Logf("  Service: %s, Status: %s, Dependencies: %d",
					detail.Name, detail.Status, len(detail.Dependencies))
			}
		}
	})

	t.Run("GetAllRunningServices_RefreshButton", func(t *testing.T) {
		serviceHandler := NewServiceHandler(runtime, instaDir)

		start := time.Now()
		statusMap, err := serviceHandler.GetAllRunningServices()
		duration := time.Since(start)

		if err != nil {
			t.Logf("GetAllRunningServices failed: %v", err)
		}

		t.Logf("Refresh: GetAllRunningServices took %v for %d services",
			duration, len(statusMap))

		// This should be very fast (under 500ms) for responsive refresh
		if duration > 500*time.Millisecond {
			t.Logf("WARNING: Refresh is slow (%v) - users will experience lag when clicking refresh", duration)
		}

		// Count statuses
		runningCount := 0
		stoppedCount := 0
		errorCount := 0
		for _, status := range statusMap {
			switch status.Status {
			case "running":
				runningCount++
			case "stopped":
				stoppedCount++
			case "error":
				errorCount++
			}
		}
		t.Logf("  Status breakdown: %d running, %d stopped, %d errors",
			runningCount, stoppedCount, errorCount)
	})
}

// Test to measure the performance bottleneck in GetAllRunningServices
func TestPerformance_GetAllRunningServices_Bottleneck(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	serviceHandler := NewServiceHandler(runtime, instaDir)

	t.Run("Current_GetAllRunningServices_Implementation", func(t *testing.T) {
		// This tests the current implementation that calls getServiceStatusInternal for each service
		start := time.Now()
		statusMap, err := serviceHandler.GetAllRunningServices()
		duration := time.Since(start)

		if err != nil {
			t.Logf("GetAllRunningServices failed: %v", err)
		}

		t.Logf("Current implementation: %d services took %v", len(statusMap), duration)

		// This is likely slow because it calls getServiceStatusInternal for each service
		if duration > 2*time.Second {
			t.Logf("BOTTLENECK IDENTIFIED: Current implementation is too slow (%v)", duration)
		}
	})

	t.Run("Optimized_BaseHandler_fetchRunningContainers", func(t *testing.T) {
		// This tests the optimized base handler approach
		baseHandler := NewBaseHandler(runtime, instaDir)

		start := time.Now()
		runningContainers, err := baseHandler.fetchCurrentContainers()
		duration := time.Since(start)

		if err != nil {
			t.Logf("fetchCurrentContainers failed: %v", err)
		}

		t.Logf("Optimized base handler: got %d containers in %v", len(runningContainers), duration)

		// This should be fast due to our previous optimizations
		if duration < 100*time.Millisecond {
			t.Logf("SUCCESS: Base handler is fast (%v)", duration)
		}
	})

	t.Run("ServiceStatusInternal_PerService_Cost", func(t *testing.T) {
		// Measure the cost of getServiceStatusInternal for individual services
		testServices := []string{"postgres", "redis", "mysql", "grafana", "superset"}
		composeFiles := serviceHandler.getComposeFiles()

		totalDuration := time.Duration(0)
		for _, serviceName := range testServices {
			start := time.Now()
			status, err := serviceHandler.getServiceStatusInternal(serviceName, composeFiles)
			duration := time.Since(start)
			totalDuration += duration

			t.Logf("getServiceStatusInternal(%s) took %v, status: %s", serviceName, duration, status)
			if err != nil {
				t.Logf("  Error: %v", err)
			}

			// Each service check should be fast
			if duration > 200*time.Millisecond {
				t.Logf("WARNING: Individual service status check is slow (%v)", duration)
			}
		}

		avgDuration := totalDuration / time.Duration(len(testServices))
		t.Logf("Average getServiceStatusInternal duration: %v", avgDuration)

		// Extrapolate to all 76 services
		estimatedTotal := avgDuration * 76
		t.Logf("Estimated total for all 76 services: %v", estimatedTotal)

		if estimatedTotal > 5*time.Second {
			t.Logf("PERFORMANCE ISSUE: Estimated total time for all services is too high (%v)", estimatedTotal)
		}
	})
}

// Test to compare different approaches for getting service statuses
func TestPerformance_ServiceStatusApproaches_Comparison(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	serviceHandler := NewServiceHandler(runtime, instaDir)
	baseHandler := NewBaseHandler(runtime, instaDir)

	t.Run("Approach1_Current_GetAllRunningServices", func(t *testing.T) {
		// Current approach: calls getServiceStatusInternal for each service
		start := time.Now()
		statusMap, err := serviceHandler.GetAllRunningServices()
		duration := time.Since(start)

		if err != nil {
			t.Logf("Current approach failed: %v", err)
		}

		t.Logf("Approach 1 (Current): %d services in %v", len(statusMap), duration)
	})

	t.Run("Approach2_Optimized_SingleDockerPS", func(t *testing.T) {
		// Optimized approach: single docker ps call + pattern matching
		start := time.Now()

		// Get all services
		allServices := serviceHandler.ListServices()

		// Get running containers with single call
		currentContainers, err := baseHandler.getCurrentContainers()
		if err != nil {
			t.Logf("Failed to get current containers: %v", err)
			return
		}

		// Map services to statuses using pattern matching
		statusMap := make(map[string]models.ServiceStatus)
		composeFiles := serviceHandler.getComposeFiles()

		for _, service := range allServices {
			status := "stopped" // Default

			// Check if service is running using optimized container matching
			if baseHandler.isServiceRunning(service.Name, composeFiles, currentContainers) {
				status = "running"
			}

			statusMap[service.Name] = models.ServiceStatus{
				ServiceName: service.Name,
				Status:      status,
			}
		}

		duration := time.Since(start)
		t.Logf("Approach 2 (Optimized): %d services in %v", len(statusMap), duration)

		// Count running services
		runningCount := 0
		for _, status := range statusMap {
			if status.Status == "running" {
				runningCount++
			}
		}
		t.Logf("  Found %d running services", runningCount)
	})

	t.Run("Approach3_UltraFast_ContainersOnly", func(t *testing.T) {
		// Ultra-fast approach: just get running containers, don't map to services
		start := time.Now()

		currentContainers, err := runtime.GetAllContainerStatuses()
		if err != nil {
			t.Logf("Failed to get current containers: %v", err)
			return
		}

		duration := time.Since(start)
		t.Logf("Approach 3 (Ultra-fast): got %d containers in %v", len(currentContainers), duration)

		// This is what the UI should use for initial load - just show running containers
		// Service definitions come from compose files, not from container inspection
	})
}

// Test to measure dependency loading performance
func TestPerformance_DependencyLoading(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	serviceHandler := NewServiceHandler(runtime, instaDir)

	t.Run("GetAllServiceDependencies_Performance", func(t *testing.T) {
		start := time.Now()
		dependencyMap, err := serviceHandler.GetAllServiceDependencies()
		duration := time.Since(start)

		if err != nil {
			t.Logf("GetAllServiceDependencies failed: %v", err)
		}

		t.Logf("GetAllServiceDependencies: %d services took %v", len(dependencyMap), duration)

		// Dependencies should load reasonably fast
		if duration > 1*time.Second {
			t.Logf("WARNING: Dependency loading is slow (%v)", duration)
		}

		// Show some sample dependencies
		count := 0
		for serviceName, deps := range dependencyMap {
			if count < 5 && len(deps) > 0 {
				t.Logf("  %s has %d dependencies: %v", serviceName, len(deps), deps)
				count++
			}
		}
	})
}

// Test to measure the impact of the number of services on performance
func TestPerformance_ServiceCount_Impact(t *testing.T) {
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping performance tests")
	}

	tempDir := t.TempDir()
	instaDir := filepath.Join(tempDir, ".insta")
	os.MkdirAll(instaDir, 0755)
	createTestComposeFiles(t, instaDir)

	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		t.Skipf("No container runtime available: %v", err)
	}
	runtime := provider.SelectedRuntime()

	serviceHandler := NewServiceHandler(runtime, instaDir)

	// Test with different numbers of services to see how performance scales
	serviceCounts := []int{5, 10, 20, 50}

	for _, count := range serviceCounts {
		t.Run(fmt.Sprintf("ServiceCount_%d", count), func(t *testing.T) {
			// Get first N services
			allServices := serviceHandler.ListServices()
			if len(allServices) < count {
				t.Skipf("Not enough services available (need %d, have %d)", count, len(allServices))
			}

			testServices := allServices[:count]
			serviceNames := make([]string, len(testServices))
			for i, service := range testServices {
				serviceNames[i] = service.Name
			}

			start := time.Now()
			statusMap, err := serviceHandler.GetMultipleServiceStatuses(serviceNames)
			duration := time.Since(start)

			if err != nil {
				t.Logf("GetMultipleServiceStatuses failed for %d services: %v", count, err)
			}

			t.Logf("GetMultipleServiceStatuses for %d services took %v", count, duration)

			// Count statuses
			runningCount := 0
			stoppedCount := 0
			errorCount := 0
			for _, status := range statusMap {
				switch status.Status {
				case "running":
					runningCount++
				case "stopped":
					stoppedCount++
				case "error":
					errorCount++
				}
			}
			t.Logf("  Status breakdown: %d running, %d stopped, %d errors", runningCount, stoppedCount, errorCount)

			// Calculate per-service time
			perServiceTime := duration / time.Duration(count)
			t.Logf("  Average per service: %v", perServiceTime)

			// Extrapolate to full service count (76 services)
			fullEstimate := perServiceTime * 76
			t.Logf("  Estimated for 76 services: %v", fullEstimate)
		})
	}
}

// TestPerformance_IndividualServiceItemCalls tests the performance of calls made by each ServiceItem
func TestPerformance_IndividualServiceItemCalls(t *testing.T) {
	// This test demonstrates the performance issue with individual ServiceItem calls
	// In the real UI, each ServiceItem component makes these calls on mount

	t.Logf("Performance Analysis: Individual ServiceItem Calls")
	t.Logf("========================================================")
	t.Logf("Problem: Each ServiceItem component calls 3 methods on mount:")
	t.Logf("  1. GetImageInfo(serviceName) - ~20-50ms per call")
	t.Logf("  2. CheckImageExists(serviceName) - ~50-100ms per call")
	t.Logf("  3. GetDependencyStatus(serviceName) - ~10-30ms per call")
	t.Logf("")
	t.Logf("With 76 services, this means:")
	t.Logf("  - 76 × 3 = 228 concurrent API calls on UI startup")
	t.Logf("  - Each call involves Docker/Podman commands")
	t.Logf("  - Total time: 76 × (20+75+15)ms = ~8.36 seconds if sequential")
	t.Logf("  - Even with concurrency, this creates significant load")
	t.Logf("")
	t.Logf("Solution: Optimize frontend to:")
	t.Logf("  1. Lazy load these details (only when service is expanded)")
	t.Logf("  2. Batch the calls into fewer backend requests")
	t.Logf("  3. Cache results to avoid repeated calls")

	// Simulate the performance impact
	serviceCount := 76
	callsPerService := 3
	avgCallTime := 50 * time.Millisecond

	sequentialTime := time.Duration(serviceCount) * time.Duration(callsPerService) * avgCallTime
	concurrentTime := avgCallTime // Assuming perfect parallelization

	t.Logf("")
	t.Logf("Performance Impact:")
	t.Logf("  Sequential execution: %v", sequentialTime)
	t.Logf("  Concurrent execution (best case): %v", concurrentTime)
	t.Logf("  Realistic concurrent execution: %v", concurrentTime*4) // Account for resource limits
}

// Helper functions
func isDockerAvailable() bool {
	provider := container.NewProvider()
	return provider.DetectRuntime() == nil
}

func createTestComposeFiles(t *testing.T, instaDir string) {
	// Create a minimal docker-compose.yaml for testing with explicit container names
	composeContent := `version: '3.8'
services:
  postgres:
    image: postgres:13
    container_name: postgres
    environment:
      POSTGRES_PASSWORD: postgres
    ports:
      - "15432:5432"  # Use different port to avoid conflicts
    healthcheck:
      test: ["CMD", "pg_isready", "-U", "postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
  redis:
    image: redis:6
    container_name: redis
    ports:
      - "16379:6379"  # Use different port to avoid conflicts
  mysql:
    image: mysql:8
    container_name: mysql
    environment:
      MYSQL_ROOT_PASSWORD: root
    ports:
      - "13306:3306"  # Use different port to avoid conflicts
  grafana:
    image: grafana/grafana:latest
    container_name: grafana
    ports:
      - "13000:3000"  # Use different port to avoid conflicts
  superset:
    image: apache/superset:latest
    container_name: superset
    ports:
      - "18088:8088"  # Use different port to avoid conflicts
`

	composeFile := filepath.Join(instaDir, "docker-compose.yaml")
	if err := os.WriteFile(composeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("Failed to create test compose file: %v", err)
	}

	// Create empty persist file
	persistFile := filepath.Join(instaDir, "docker-compose-persist.yaml")
	if err := os.WriteFile(persistFile, []byte("version: '3.8'\nservices: {}\n"), 0644); err != nil {
		t.Fatalf("Failed to create test persist file: %v", err)
	}
}
