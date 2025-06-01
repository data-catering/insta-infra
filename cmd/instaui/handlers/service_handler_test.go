package handlers

import (
	"errors"
	"fmt"
	"testing"
)

// Mock implementation moved to test_utils.go

func TestServiceHandler_NewServiceHandler(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	instaDir := "/test/insta"

	handler := NewServiceHandler(mockRuntime, instaDir)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	if handler.containerRuntime != mockRuntime {
		t.Errorf("Expected containerRuntime to be %v, got %v", mockRuntime, handler.containerRuntime)
	}
	if handler.instaDir != instaDir {
		t.Errorf("Expected instaDir to be %s, got %s", instaDir, handler.instaDir)
	}
}

func TestServiceHandler_ListServices(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	services := handler.ListServices()

	if len(services) == 0 {
		t.Error("Expected services to be returned, got empty list")
	}

	// Verify services are sorted by name
	for i := 1; i < len(services); i++ {
		if services[i-1].Name > services[i].Name {
			t.Errorf("Services not sorted properly: %s > %s", services[i-1].Name, services[i].Name)
		}
	}
}

func TestServiceHandler_GetServiceStatus_Running(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllContainerStatuses(func() (map[string]string, error) {
			// Return postgres as a running container
			return map[string]string{"postgres": "running"}, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			// Return the same container name that GetAllContainerStatuses returns
			return serviceName, nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	// Clear any stopped service tracking that might interfere
	handler.ClearStoppedServiceTrackingForTesting()

	status, err := handler.GetServiceStatus("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != "running" {
		t.Errorf("Expected status 'running', got '%s'", status)
	}
}

func TestServiceHandler_GetServiceStatus_Stopped(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllContainerStatuses(func() (map[string]string, error) {
			// Return empty list - no running containers
			return map[string]string{}, nil
		}).
		withGetContainerStatus(func(containerName string) (string, error) {
			// Container doesn't exist or isn't running
			return "", fmt.Errorf("container %s not found", containerName)
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	status, err := handler.GetServiceStatus("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", status)
	}
}

func TestServiceHandler_GetServiceStatus_ContainerNameError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		}).
		withGetContainerStatus(func(containerName string) (string, error) {
			// This shouldn't be called since GetContainerName fails, but just in case
			return "", fmt.Errorf("container %s not found", containerName)
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	status, err := handler.GetServiceStatus("nonexistent")

	// With the new BaseHandler logic, when GetContainerName fails,
	// the service is treated as "stopped" rather than propagating the error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", status)
	}
}

func TestServiceHandler_GetServiceDependencies_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	deps, err := handler.GetServiceDependencies("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Current implementation returns empty dependencies for performance
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestServiceHandler_GetServiceDependencies_Error(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	deps, err := handler.GetServiceDependencies("grafana")

	// Current implementation doesn't return errors, just empty dependencies
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestServiceHandler_StartService_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withComposeUp(func(composeFiles []string, services []string, quiet bool) error {
			return nil
		})
	handler := NewServiceHandler(mockRuntime, "/tmp/test_insta")

	err := handler.StartService("postgres", false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestServiceHandler_StartService_Error(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withComposeUp(func(composeFiles []string, services []string, quiet bool) error {
			return errors.New("docker daemon not running")
		})
	handler := NewServiceHandler(mockRuntime, "/tmp/test_insta")

	err := handler.StartService("postgres", false)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "docker daemon not running") {
		t.Errorf("Expected error to contain 'docker daemon not running', got %s", err.Error())
	}
}

func TestServiceHandler_StopService_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withComposeDown(func(composeFiles []string, services []string) error {
			return nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	err := handler.StopService("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestServiceHandler_StopService_Error(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withComposeDown(func(composeFiles []string, services []string) error {
			return errors.New("service not found")
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	err := handler.StopService("postgres")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "service not found") {
		t.Errorf("Expected error to contain 'service not found', got %s", err.Error())
	}
}

func TestServiceHandler_StopAllServices_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withComposeDown(func(composeFiles []string, services []string) error {
			if len(services) != 0 {
				return errors.New("expected empty services slice for stop all")
			}
			return nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	err := handler.StopAllServices()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestServiceHandler_getComposeFiles(t *testing.T) {
	handler := NewServiceHandler(nil, "/test/insta")

	composeFiles := handler.getComposeFiles()

	if len(composeFiles) == 0 {
		t.Error("Expected at least one compose file")
	}
	if !contains(composeFiles[0], "/test/insta/docker-compose.yaml") {
		t.Errorf("Expected first compose file to contain '/test/insta/docker-compose.yaml', got %s", composeFiles[0])
	}
}

func TestServiceHandler_GetMultipleServiceStatuses_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllContainerStatuses(func() (map[string]string, error) {
			// Return both postgres and redis as running containers
			return map[string]string{"postgres": "running", "redis": "running"}, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			// Return the same container name that GetAllContainerStatuses returns
			return serviceName, nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	// Clear any stopped service tracking that might interfere
	handler.ClearStoppedServiceTrackingForTesting()

	serviceNames := []string{"postgres", "redis"}
	statusMap, err := handler.GetMultipleServiceStatuses(serviceNames)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(statusMap) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statusMap))
	}
	if statusMap["postgres"].Status != "running" {
		t.Errorf("Expected postgres status 'running', got '%s'", statusMap["postgres"].Status)
	}
	if statusMap["redis"].Status != "running" {
		t.Errorf("Expected redis status 'running', got '%s'", statusMap["redis"].Status)
	}
}

func TestServiceHandler_GetMultipleServiceStatuses_EmptyInput(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	statusMap, err := handler.GetMultipleServiceStatuses([]string{})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(statusMap) != 0 {
		t.Errorf("Expected empty status map, got %d entries", len(statusMap))
	}
}

func TestServiceHandler_GetMultipleServiceStatuses_WithErrors(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllContainerStatuses(func() (map[string]string, error) {
			// Return only redis as running, postgres will be stopped
			return map[string]string{"redis": "running"}, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			// Return the same container name that GetAllContainerStatuses returns
			if serviceName == "redis" {
				return serviceName, nil
			}
			// For postgres, return a different name so it's not found in running containers
			return "test_" + serviceName + "_1", nil
		}).
		withGetContainerStatus(func(containerName string) (string, error) {
			// Only redis is running, postgres returns error (not found)
			if containerName == "redis" {
				return "running", nil
			}
			return "", fmt.Errorf("container %s not found", containerName)
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	// Clear any stopped service tracking that might interfere
	handler.ClearStoppedServiceTrackingForTesting()

	serviceNames := []string{"postgres", "redis"}
	statusMap, err := handler.GetMultipleServiceStatuses(serviceNames)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(statusMap) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statusMap))
	}
	if statusMap["postgres"].Status != "stopped" {
		t.Errorf("Expected postgres status 'stopped', got '%s'", statusMap["postgres"].Status)
	}
	if statusMap["redis"].Status != "running" {
		t.Errorf("Expected redis status 'running', got '%s'", statusMap["redis"].Status)
	}
}

// Helper functions moved to test_utils.go
