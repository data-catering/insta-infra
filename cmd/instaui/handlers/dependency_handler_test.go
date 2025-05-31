package handlers

import (
	"errors"
	"testing"
)

func TestDependencyHandler_NewDependencyHandler(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	instaDir := "/test/insta"
	serviceHandler := &ServiceHandler{}

	handler := NewDependencyHandler(mockRuntime, instaDir, serviceHandler)

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

func TestDependencyHandler_GetDependencyStatus_Success(t *testing.T) {
	dependencies := []string{"postgres", "redis"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			if contains(containerName, "postgres") {
				return map[string]string{"5432/tcp": "5432"}, nil
			}
			return map[string]string{"6379/tcp": "6379"}, nil
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetDependencyStatus("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status == nil {
		t.Fatal("Expected status to be returned, got nil")
	}
	if status.ServiceName != "grafana" {
		t.Errorf("Expected service name 'grafana', got '%s'", status.ServiceName)
	}
	if len(status.Dependencies) != len(dependencies) {
		t.Errorf("Expected %d dependencies, got %d", len(dependencies), len(status.Dependencies))
	}
	if status.RequiredCount != len(dependencies) {
		t.Errorf("Expected required count %d, got %d", len(dependencies), status.RequiredCount)
	}
	if status.RunningCount != len(dependencies) {
		t.Errorf("Expected running count %d, got %d", len(dependencies), status.RunningCount)
	}
	if !status.AllDependenciesReady {
		t.Error("Expected all dependencies to be ready")
	}
	if !status.CanStart {
		t.Error("Expected service to be able to start")
	}
}

func TestDependencyHandler_GetDependencyStatus_GetDependenciesError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return nil, errors.New("compose file not found")
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetDependencyStatus("grafana")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if status != nil {
		t.Errorf("Expected status to be nil, got %v", status)
	}
	if !contains(err.Error(), "failed to get dependencies for service") {
		t.Errorf("Expected error to contain 'failed to get dependencies for service', got %s", err.Error())
	}
}

func TestDependencyHandler_GetDependencyStatus_ServiceStatusError(t *testing.T) {
	dependencies := []string{"postgres"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetDependencyStatus("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status == nil {
		t.Fatal("Expected status to be returned, got nil")
	}
	// With the new BaseHandler logic, when GetContainerName fails, 
	// the service is treated as "stopped" rather than "error"
	if status.ErrorCount != 0 {
		t.Errorf("Expected error count 0, got %d", status.ErrorCount)
	}
	if len(status.Dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(status.Dependencies))
	}
	if status.Dependencies[0].Status != "stopped" {
		t.Errorf("Expected dependency status 'stopped', got '%s'", status.Dependencies[0].Status)
	}
	if status.AllDependenciesReady {
		t.Error("Expected all dependencies to not be ready")
	}
	if status.CanStart {
		t.Error("Expected service to not be able to start")
	}
}

func TestDependencyHandler_GetDependencyStatus_MixedStatuses(t *testing.T) {
	dependencies := []string{"postgres", "redis"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			if contains(containerName, "postgres") {
				return map[string]string{"5432/tcp": "5432"}, nil // Running
			}
			return nil, errors.New("container not running") // Stopped
		})
	
	// Add GetContainerStatus mock to properly handle the enhanced status detection
	mockRuntime.getContainerStatusFunc = func(containerName string) (string, error) {
		if contains(containerName, "postgres") {
			return "running", nil
		}
		return "not_found", nil // Redis container not found (stopped)
	}
	
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetDependencyStatus("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	// Based on the mock setup: postgres is running, redis is stopped
	if status.RunningCount != 1 {
		t.Errorf("Expected running count 1, got %d", status.RunningCount)
	}
	if status.RequiredCount != 2 {
		t.Errorf("Expected required count 2, got %d", status.RequiredCount)
	}
	if status.AllDependenciesReady {
		t.Error("Expected all dependencies to not be ready (redis is stopped)")
	}
	if status.CanStart {
		t.Error("Expected service to not be able to start (redis dependency not ready)")
	}
}

func TestDependencyHandler_StartAllDependencies_Success(t *testing.T) {
	dependencies := []string{"postgres", "redis"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return nil, errors.New("container not running") // All stopped initially
		}).
		withComposeUp(func(composeFiles []string, services []string, quiet bool) error {
			return nil
		})
	handler := NewDependencyHandler(mockRuntime, "/tmp/test_insta", nil)

	err := handler.StartAllDependencies("grafana", false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDependencyHandler_StartAllDependencies_GetDependenciesError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return nil, errors.New("compose file not found")
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartAllDependencies("grafana", false)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "failed to get dependencies for service") {
		t.Errorf("Expected error to contain 'failed to get dependencies for service', got %s", err.Error())
	}
}

func TestDependencyHandler_StartAllDependencies_NoDependencies(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return []string{}, nil
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartAllDependencies("redis", false)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "has no dependencies to start") {
		t.Errorf("Expected error to contain 'has no dependencies to start', got %s", err.Error())
	}
}

func TestDependencyHandler_StartAllDependencies_StartServiceError(t *testing.T) {
	dependencies := []string{"postgres"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return nil, errors.New("container not running")
		}).
		withComposeUp(func(composeFiles []string, services []string, quiet bool) error {
			return errors.New("docker daemon not running")
		})
	handler := NewDependencyHandler(mockRuntime, "/tmp/test_insta", nil)

	err := handler.StartAllDependencies("grafana", false)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "failed to start dependency") {
		t.Errorf("Expected error to contain 'failed to start dependency', got %s", err.Error())
	}
}

func TestDependencyHandler_StartAllDependencies_SkipRunning(t *testing.T) {
	dependencies := []string{"postgres", "redis"}
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return dependencies, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			if contains(containerName, "postgres") {
				return map[string]string{"5432/tcp": "5432"}, nil // Running
			}
			return nil, errors.New("container not running") // Not running
		}).
		withComposeUp(func(composeFiles []string, services []string, quiet bool) error {
			// Should only be called for redis, not postgres
			if len(services) != 1 || services[0] != "redis" {
				return errors.New("unexpected service started")
			}
			return nil
		})
	handler := NewDependencyHandler(mockRuntime, "/tmp/test_insta", nil)

	err := handler.StartAllDependencies("grafana", false)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDependencyHandler_StopDependencyChain_Success(t *testing.T) {
	// Mock dependencies: grafana depends on postgres, superset depends on postgres
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			switch serviceName {
			case "grafana":
				return []string{"postgres"}, nil
			case "superset":
				return []string{"postgres"}, nil
			default:
				return []string{}, nil
			}
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"80/tcp": "80"}, nil // All running
		}).
		withComposeDown(func(composeFiles []string, services []string) error {
			return nil
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	err := handler.StopDependencyChain("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDependencyHandler_StopDependencyChain_NoDependents(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return []string{}, nil // No dependencies for any service
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_redis_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"6379/tcp": "6379"}, nil
		}).
		withComposeDown(func(composeFiles []string, services []string) error {
			return nil
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	err := handler.StopDependencyChain("redis")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDependencyHandler_StopDependencyChain_StopServiceError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			return []string{}, nil
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_redis_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"6379/tcp": "6379"}, nil
		}).
		withComposeDown(func(composeFiles []string, services []string) error {
			return errors.New("failed to stop container")
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	err := handler.StopDependencyChain("redis")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "failed to stop service") {
		t.Errorf("Expected error to contain 'failed to stop service', got %s", err.Error())
	}
}

func TestDependencyHandler_getServiceByName_Found(t *testing.T) {
	handler := NewDependencyHandler(nil, "/test/insta", nil)

	service := handler.getServiceByName("postgres")

	if service == nil {
		t.Fatal("Expected service to be found, got nil")
	}
	if service.Name != "postgres" {
		t.Errorf("Expected service name 'postgres', got '%s'", service.Name)
	}
}

func TestDependencyHandler_getServiceByName_NotFound(t *testing.T) {
	handler := NewDependencyHandler(nil, "/test/insta", nil)

	service := handler.getServiceByName("nonexistent")

	if service != nil {
		t.Errorf("Expected service to be nil, got %v", service)
	}
}

func TestDependencyHandler_getServiceStatusInternal_Running(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"5432/tcp": "5432"}, nil
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetServiceStatusInternal("postgres", []string{"compose.yaml"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != "running" {
		t.Errorf("Expected status 'running', got '%s'", status)
	}
}

func TestDependencyHandler_getServiceStatusInternal_Stopped(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return nil, errors.New("container not running")
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetServiceStatusInternal("postgres", []string{"compose.yaml"})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", status)
	}
}

func TestDependencyHandler_getServiceStatusInternal_ContainerNameError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		})
	handler := NewDependencyHandler(mockRuntime, "/test/insta", nil)

	status, err := handler.GetServiceStatusInternal("nonexistent", []string{"compose.yaml"})

	// With the new BaseHandler logic, when GetContainerName fails, 
	// the service is treated as "stopped" rather than propagating the error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if status != "stopped" {
		t.Errorf("Expected status 'stopped', got '%s'", status)
	}
}

func TestDependencyHandler_getComposeFiles(t *testing.T) {
	handler := NewDependencyHandler(nil, "/test/insta", nil)

	composeFiles := handler.getComposeFiles()

	if len(composeFiles) == 0 {
		t.Error("Expected at least one compose file")
	}
	if !contains(composeFiles[0], "/test/insta/docker-compose.yaml") {
		t.Errorf("Expected first compose file to contain '/test/insta/docker-compose.yaml', got %s", composeFiles[0])
	}
}
