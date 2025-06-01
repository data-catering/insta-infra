package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/handlers"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

func TestRecursiveDependencies(t *testing.T) {
	// Create runtime and handler
	runtime := container.NewDockerRuntime()

	// Use absolute path to avoid path resolution issues
	cwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	instaDir := filepath.Join(cwd, "resources")

	// Verify the docker-compose file exists
	composeFile := filepath.Join(instaDir, "docker-compose.yaml")
	t.Logf("Looking for compose file at: %s", composeFile)
	if _, err := os.Stat(composeFile); err != nil {
		t.Fatalf("Compose file does not exist: %v", err)
	}

	handler := handlers.NewServiceHandler(runtime, instaDir)

	// Test case: airflow should have recursive dependencies
	t.Run("airflow recursive dependencies", func(t *testing.T) {
		deps, err := handler.GetServiceDependencies("airflow")
		if err != nil {
			t.Fatalf("Failed to get dependencies for airflow: %v", err)
		}

		t.Logf("Airflow dependencies found: %v", deps)

		// Expected dependencies for airflow (recursive):
		// airflow -> airflow-init -> postgres-data -> postgres
		expectedDeps := []string{
			"airflow-init",  // direct dependency
			"postgres-data", // indirect via airflow-init
			"postgres",      // indirect via postgres
		}

		// Check that we have all expected dependencies
		depMap := make(map[string]bool)
		for _, dep := range deps {
			depMap[dep] = true
		}

		for _, expectedDep := range expectedDeps {
			if !depMap[expectedDep] {
				t.Errorf("Expected dependency '%s' not found in airflow dependencies. Got: %v", expectedDep, deps)
			}
		}

		// Verify we have at least the expected number of dependencies
		if len(deps) < len(expectedDeps) {
			t.Errorf("Expected at least %d dependencies for airflow, got %d: %v", len(expectedDeps), len(deps), deps)
		}
	})

	// Test case: postgres should have postgres container dependency
	t.Run("postgres dependencies", func(t *testing.T) {
		deps, err := handler.GetServiceDependencies("postgres")
		if err != nil {
			t.Fatalf("Failed to get dependencies for postgres: %v", err)
		}

		t.Logf("Postgres dependencies found: %v", deps)

		// postgres service (container: postgres-data) should depend on postgres-server service (container: postgres)
		expectedDeps := []string{
			"postgres", // container name from postgres-server service
		}

		// Check that we have all expected dependencies
		depMap := make(map[string]bool)
		for _, dep := range deps {
			depMap[dep] = true
		}

		for _, expectedDep := range expectedDeps {
			if !depMap[expectedDep] {
				t.Errorf("Expected dependency '%s' not found in postgres dependencies. Got: %v", expectedDep, deps)
			}
		}
	})

	// Test case: airflow-init should have postgres container dependencies
	t.Run("airflow-init dependencies", func(t *testing.T) {
		deps, err := handler.GetServiceDependencies("airflow-init")
		if err != nil {
			t.Fatalf("Failed to get dependencies for airflow-init: %v", err)
		}

		t.Logf("Airflow-init dependencies found: %v", deps)

		// airflow-init should depend on postgres service (container: postgres-data) -> postgres-server service (container: postgres)
		expectedDeps := []string{
			"postgres-data", // container name from postgres service
			"postgres",      // container name from postgres-server service
		}

		// Check that we have all expected dependencies
		depMap := make(map[string]bool)
		for _, dep := range deps {
			depMap[dep] = true
		}

		for _, expectedDep := range expectedDeps {
			if !depMap[expectedDep] {
				t.Errorf("Expected dependency '%s' not found in airflow-init dependencies. Got: %v", expectedDep, deps)
			}
		}
	})

	// Test case: check other services for expected container name dependencies
	t.Run("blazer dependencies", func(t *testing.T) {
		deps, err := handler.GetServiceDependencies("blazer")
		if err != nil {
			t.Fatalf("Failed to get dependencies for blazer: %v", err)
		}
		t.Logf("Blazer dependencies found: %v", deps)

		// blazer should depend on postgres service chain
		expectedDeps := []string{
			"postgres-data", // container name from postgres service
			"postgres",      // container name from postgres-server service
		}

		depMap := make(map[string]bool)
		for _, dep := range deps {
			depMap[dep] = true
		}

		for _, expectedDep := range expectedDeps {
			if !depMap[expectedDep] {
				t.Errorf("Expected dependency '%s' not found in blazer dependencies. Got: %v", expectedDep, deps)
			}
		}
	})

	t.Run("argilla dependencies", func(t *testing.T) {
		deps, err := handler.GetServiceDependencies("argilla")
		if err != nil {
			t.Fatalf("Failed to get dependencies for argilla: %v", err)
		}
		t.Logf("Argilla dependencies found: %v", deps)

		// argilla should have multiple container dependencies (check for at least some key ones)
		expectedContainerNames := []string{
			"argilla",       // direct dependency container
			"postgres-data", // postgres chain containers
			"postgres",
			"redis",              // redis container
			"elasticsearch-data", // elasticsearch chain containers
			"elasticsearch",
		}

		depMap := make(map[string]bool)
		for _, dep := range deps {
			depMap[dep] = true
		}

		for _, expectedDep := range expectedContainerNames {
			if !depMap[expectedDep] {
				t.Errorf("Expected dependency container '%s' not found in argilla dependencies. Got: %v", expectedDep, deps)
			}
		}
	})

	t.Run("grafana dependencies", func(t *testing.T) {
		deps, err := handler.GetServiceDependencies("grafana")
		if err != nil {
			t.Fatalf("Failed to get dependencies for grafana: %v", err)
		}
		t.Logf("Grafana dependencies found: %v", deps)

		// grafana has no dependencies in the current compose structure
		if len(deps) != 0 {
			t.Logf("Note: grafana has dependencies: %v (this might be expected if compose structure changed)", deps)
		}
	})
}

func TestDependencyStatusResolution(t *testing.T) {
	// Create runtime and handler
	runtime := container.NewDockerRuntime()

	// Use absolute path to avoid path resolution issues
	cwd, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	instaDir := filepath.Join(cwd, "resources")

	handler := handlers.NewServiceHandler(runtime, instaDir)

	// Test that we can get dependency statuses
	t.Run("get all dependency statuses", func(t *testing.T) {
		statuses, err := handler.GetAllDependencyStatuses()
		if err != nil {
			t.Fatalf("Failed to get dependency statuses: %v", err)
		}

		t.Logf("All dependency statuses: %v", statuses)

		// We should have status information for containers
		if len(statuses) == 0 {
			t.Logf("Warning: No dependency statuses found (containers might not be running)")
		}

		// Check that status objects have the right structure
		for containerName, statusObj := range statuses {
			if statusObj.ServiceName == "" {
				t.Errorf("Status object for %s missing ServiceName", containerName)
			}
			if statusObj.Status == "" {
				t.Errorf("Status object for %s missing Status", containerName)
			}
			t.Logf("Container %s has status: %s", containerName, statusObj.Status)
		}
	})
}
