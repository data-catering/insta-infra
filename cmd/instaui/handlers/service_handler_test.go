package handlers

import (
	"errors"
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
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"5432/tcp": "5432"}, nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

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
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return nil, errors.New("container not running")
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
	dependencies := []string{"postgres", "redis"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	deps, err := handler.GetServiceDependencies("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(deps) != len(dependencies) {
		t.Errorf("Expected %d dependencies, got %d", len(dependencies), len(deps))
	}
}

func TestServiceHandler_GetServiceDependencies_Error(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return nil, errors.New("compose file not found")
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	deps, err := handler.GetServiceDependencies("grafana")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if deps != nil {
		t.Errorf("Expected deps to be nil, got %v", deps)
	}
	if !contains(err.Error(), "could not get recursive dependencies") {
		t.Errorf("Expected error to contain 'could not get recursive dependencies', got %s", err.Error())
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
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			if contains(containerName, "postgres") {
				return map[string]string{"5432/tcp": "5432"}, nil
			}
			return map[string]string{"6379/tcp": "6379"}, nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

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
	if statusMap["postgres"].ServiceName != "postgres" {
		t.Errorf("Expected postgres serviceName 'postgres', got '%s'", statusMap["postgres"].ServiceName)
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
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			if serviceName == "postgres" {
				return "", errors.New("service not found")
			}
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"6379/tcp": "6379"}, nil
		})
	handler := NewServiceHandler(mockRuntime, "/test/insta")

	serviceNames := []string{"postgres", "redis"}
	statusMap, err := handler.GetMultipleServiceStatuses(serviceNames)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(statusMap) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statusMap))
	}
	// With the new BaseHandler logic, when GetContainerName fails, 
	// the service is treated as "stopped" rather than "error"
	if statusMap["postgres"].Status != "stopped" {
		t.Errorf("Expected postgres status 'stopped', got '%s'", statusMap["postgres"].Status)
	}
	if statusMap["redis"].Status != "running" {
		t.Errorf("Expected redis status 'running', got '%s'", statusMap["redis"].Status)
	}
}

// Helper functions moved to test_utils.go
