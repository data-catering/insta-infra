package main

import (
	"fmt"
	"testing"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
)

// TestAirflowDependencies tests the complete airflow dependency chain
func TestAirflowDependencies(t *testing.T) {
	// Create a mock runtime with all the necessary containers
	mockRuntime := NewEnhancedMockRuntimeInfo()

	// Set up container statuses for the complete airflow dependency chain
	mockRuntime.containers["airflow"] = "stopped"
	mockRuntime.containers["airflow-init"] = "stopped"
	mockRuntime.containers["postgres"] = "stopped"      // postgres-server container
	mockRuntime.containers["postgres-data"] = "stopped" // postgres container

	mockLogger := NewTestMockLogger()

	// Use real compose files for proper dependency resolution
	sm := models.NewServiceManager("./resources", mockRuntime, mockLogger)
	err := sm.LoadServices()
	if err != nil {
		t.Skipf("Cannot load services from compose files: %v", err)
	}

	t.Run("AirflowServiceExists", func(t *testing.T) {
		service, exists := sm.GetService("airflow")
		if !exists {
			t.Fatal("Expected airflow service to exist")
		}

		if service.Name != "airflow" {
			t.Errorf("Expected service name 'airflow', got '%s'", service.Name)
		}

		if service.ContainerName != "airflow" {
			t.Errorf("Expected container name 'airflow', got '%s'", service.ContainerName)
		}
	})

	t.Run("AirflowDependencyChain", func(t *testing.T) {
		service, exists := sm.GetService("airflow")
		if !exists {
			t.Fatal("Expected airflow service to exist")
		}

		// Test that airflow has the expected dependencies
		expectedDependencies := map[string]bool{
			"airflow":       true, // Should include itself
			"airflow-init":  true, // Direct dependency
			"postgres-data": true, // postgres container (from airflow-init -> postgres)
			"postgres":      true, // postgres-server container (from postgres-data -> postgres-server)
		}

		t.Logf("Airflow recursive dependencies: %v", service.RecursiveDependencies)

		if len(service.RecursiveDependencies) != len(expectedDependencies) {
			t.Errorf("Expected %d dependencies, got %d: %v",
				len(expectedDependencies), len(service.RecursiveDependencies), service.RecursiveDependencies)
		}

		for _, dep := range service.RecursiveDependencies {
			if !expectedDependencies[dep] {
				t.Errorf("Unexpected dependency: %s", dep)
			}
		}

		for expectedDep := range expectedDependencies {
			found := false
			for _, dep := range service.RecursiveDependencies {
				if dep == expectedDep {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Missing expected dependency: %s", expectedDep)
			}
		}
	})

	t.Run("AirflowDependencyStatuses", func(t *testing.T) {
		// Update all service statuses from containers
		_, err := sm.UpdateAllServiceStatuses()
		if err != nil {
			t.Errorf("UpdateAllServiceStatuses failed: %v", err)
		}

		// Get the airflow service after status update
		service, exists := sm.GetService("airflow")
		if !exists {
			t.Fatal("Expected airflow service to exist")
		}

		// Verify airflow service status is properly updated
		if service.Status != "stopped" {
			t.Errorf("Expected airflow status 'stopped', got '%s'", service.Status)
		}

		// Check dependency services have proper statuses (not 'unknown')
		dependencyServices := []string{"airflow-init"}

		for _, depServiceName := range dependencyServices {
			depService, exists := sm.GetService(depServiceName)
			if exists {
				if depService.Status == "unknown" || depService.Status == "" {
					t.Errorf("Dependency service %s has unknown/empty status: '%s'", depServiceName, depService.Status)
				}
				t.Logf("Dependency service %s status: %s", depServiceName, depService.Status)
			}
		}

		// Also check postgres service status since it's part of the chain
		postgresService, exists := sm.GetService("postgres")
		if exists {
			if postgresService.Status == "unknown" || postgresService.Status == "" {
				t.Errorf("Postgres service has unknown/empty status: '%s'", postgresService.Status)
			}
			t.Logf("Postgres service status: %s", postgresService.Status)
		}
	})

	t.Run("AirflowDependencyStatusWithRunningContainers", func(t *testing.T) {
		// Set some dependencies as running to test status propagation
		mockRuntime.containers["airflow-init"] = "running"
		mockRuntime.containers["postgres"] = "running"
		mockRuntime.containers["postgres-data"] = "running"

		// Update statuses
		_, err := sm.UpdateAllServiceStatuses()
		if err != nil {
			t.Errorf("UpdateAllServiceStatuses failed: %v", err)
		}

		// Check that dependency statuses are properly updated
		airflowInitService, exists := sm.GetService("airflow-init")
		if exists {
			if airflowInitService.Status != "running" {
				t.Errorf("Expected airflow-init status 'running', got '%s'", airflowInitService.Status)
			}
		}

		postgresService, exists := sm.GetService("postgres")
		if exists {
			if postgresService.Status != "running" {
				t.Errorf("Expected postgres status 'running', got '%s'", postgresService.Status)
			}
		}
	})

	t.Run("GetAllRunningServicesIncludesDependencies", func(t *testing.T) {
		// Set airflow and its dependencies as running
		mockRuntime.containers["airflow"] = "running"
		mockRuntime.containers["postgres-data"] = "running" // postgres service container

		runningServices, err := sm.GetAllRunningServices()
		if err != nil {
			t.Errorf("GetAllRunningServices failed: %v", err)
		}

		// Should be indexed by container name now
		// Note: airflow-init is not a service in models.go, so it won't appear in GetAllRunningServices
		// postgres service uses postgres-data as container name
		expectedRunningContainers := []string{"airflow", "postgres-data"}

		for _, containerName := range expectedRunningContainers {
			if _, exists := runningServices[containerName]; !exists {
				t.Errorf("Expected running service indexed by container name '%s' to exist", containerName)
			}
		}

		t.Logf("Running services found: %d", len(runningServices))
		for containerName, service := range runningServices {
			t.Logf("Running service: %s -> %s (status: %s)", containerName, service.Name, service.Status)
		}
	})

	t.Run("DependencyStatusesIncludeVirtualServices", func(t *testing.T) {
		// Set some containers as running
		mockRuntime.containers["airflow"] = "running"
		mockRuntime.containers["airflow-init"] = "running"
		mockRuntime.containers["postgres"] = "stopped"
		mockRuntime.containers["postgres-data"] = "running"

		dependencyStatuses, err := sm.GetAllDependencyStatuses()
		if err != nil {
			t.Errorf("GetAllDependencyStatuses failed: %v", err)
		}

		// Should include virtual services for dependency containers
		if _, exists := dependencyStatuses["airflow-init"]; !exists {
			t.Error("Expected airflow-init to be included in dependency statuses")
		}

		if _, exists := dependencyStatuses["postgres"]; !exists {
			t.Error("Expected postgres to be included in dependency statuses")
		}

		// Check that airflow-init has the correct status and is marked as dependency
		if airflowInit := dependencyStatuses["airflow-init"]; airflowInit != nil {
			if airflowInit.Status != "running" {
				t.Errorf("Expected airflow-init status 'running', got '%s'", airflowInit.Status)
			}
			if airflowInit.Type != "dependency" {
				t.Errorf("Expected airflow-init type 'dependency', got '%s'", airflowInit.Type)
			}
			t.Logf("Airflow-init dependency: status=%s, type=%s", airflowInit.Status, airflowInit.Type)
		}

		t.Logf("Total dependency statuses found: %d", len(dependencyStatuses))
		for containerName, service := range dependencyStatuses {
			t.Logf("Dependency: %s -> %s (status: %s, type: %s)",
				containerName, service.Name, service.Status, service.Type)
		}
	})
}

// TestAirflowDependenciesIntegration tests airflow dependencies with actual compose file parsing
func TestAirflowDependenciesIntegration(t *testing.T) {
	// Test with the real compose files to ensure airflow dependencies are properly parsed
	mockRuntime := NewEnhancedMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	// Create service manager with real compose files path (if available)
	sm := models.NewServiceManager("./resources", mockRuntime, mockLogger)
	err := sm.LoadServices()
	if err != nil {
		t.Skipf("Skipping integration test, cannot load services: %v", err)
	}

	t.Run("AirflowComposeIntegration", func(t *testing.T) {
		service, exists := sm.GetService("airflow")
		if !exists {
			t.Skip("airflow service not found in compose files")
		}

		t.Logf("Airflow service container name: %s", service.ContainerName)
		t.Logf("Airflow direct dependencies: %v", service.DirectDependencies)
		t.Logf("Airflow recursive dependencies: %v", service.RecursiveDependencies)

		// Verify airflow has airflow-init as dependency
		hasAirflowInit := false
		for _, dep := range service.RecursiveDependencies {
			if dep == "airflow-init" {
				hasAirflowInit = true
				break
			}
		}
		if !hasAirflowInit {
			t.Error("Expected airflow to have airflow-init as dependency")
		}

		// Check that postgres containers are also in the dependency chain
		hasPostgres := false
		hasPostgresData := false
		for _, dep := range service.RecursiveDependencies {
			if dep == "postgres" {
				hasPostgres = true
			}
			if dep == "postgres-data" {
				hasPostgresData = true
			}
		}

		if !hasPostgres && !hasPostgresData {
			t.Error("Expected airflow to have postgres dependencies in chain")
		}

		t.Logf("Airflow dependency chain validation: airflow-init=%v, postgres=%v, postgres-data=%v",
			hasAirflowInit, hasPostgres, hasPostgresData)
	})
}

// EnhancedMockRuntimeInfo provides better mock runtime for integration testing
type EnhancedMockRuntimeInfo struct {
	*TestMockRuntimeInfo
}

func NewEnhancedMockRuntimeInfo() *EnhancedMockRuntimeInfo {
	base := NewTestMockRuntimeInfo()
	return &EnhancedMockRuntimeInfo{
		TestMockRuntimeInfo: base,
	}
}

// Override GetAllContainerStatuses to return a more comprehensive set of containers
func (m *EnhancedMockRuntimeInfo) GetAllContainerStatuses() (map[string]string, error) {
	if m.errorMode {
		return nil, fmt.Errorf("mock error mode enabled")
	}

	// Return a comprehensive set of container statuses including all airflow dependencies
	result := make(map[string]string)

	// Copy existing containers
	for name, status := range m.containers {
		result[name] = status
	}

	// Add airflow dependency containers if they don't exist
	defaultContainers := map[string]string{
		"airflow":       "stopped",
		"airflow-init":  "stopped",
		"postgres":      "stopped",
		"postgres-data": "stopped",
	}

	for name, status := range defaultContainers {
		if _, exists := result[name]; !exists {
			result[name] = status
		}
	}

	return result, nil
}
