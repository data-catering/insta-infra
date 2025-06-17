package main

import (
	"fmt"
	"testing"

	"github.com/data-catering/insta-infra/v2/cmd/insta/handlers"
	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
)

// TestMockRuntimeInfo provides a mock implementation for testing
type TestMockRuntimeInfo struct {
	containers map[string]string // containerName -> status
	errorMode  bool              // Simulate errors when true
}

func NewTestMockRuntimeInfo() *TestMockRuntimeInfo {
	return &TestMockRuntimeInfo{
		containers: make(map[string]string),
		errorMode:  false,
	}
}

func (m *TestMockRuntimeInfo) CheckContainerStatus(containerName string) (string, error) {
	if m.errorMode {
		return "error", fmt.Errorf("mock error checking container status")
	}
	if status, exists := m.containers[containerName]; exists {
		return status, nil
	}
	return "stopped", nil
}

func (m *TestMockRuntimeInfo) GetContainerLogs(containerName string, lines int) ([]string, error) {
	if m.errorMode {
		return nil, fmt.Errorf("mock error getting container logs")
	}
	return []string{"mock log line 1", "mock log line 2"}, nil
}

func (m *TestMockRuntimeInfo) StartService(serviceName string, persist bool) error {
	if m.errorMode {
		return fmt.Errorf("mock error starting service")
	}
	m.containers[serviceName] = "running"
	return nil
}

func (m *TestMockRuntimeInfo) StopService(serviceName string) error {
	if m.errorMode {
		return fmt.Errorf("mock error stopping service")
	}
	m.containers[serviceName] = "stopped"
	return nil
}

func (m *TestMockRuntimeInfo) GetAllContainerStatuses() (map[string]string, error) {
	if m.errorMode {
		return nil, fmt.Errorf("mock error getting all container statuses")
	}
	return m.containers, nil
}

// TestMockLogger provides a simple logger for testing
type TestMockLogger struct {
	logs []string
}

func NewTestMockLogger() *TestMockLogger {
	return &TestMockLogger{logs: make([]string, 0)}
}

func (l *TestMockLogger) Log(message string) {
	l.logs = append(l.logs, message)
}

func TestServiceManagerImplementsServiceHandlerInterface(t *testing.T) {
	// Verify that ServiceManager implements ServiceHandlerInterface
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)

	// Verify it implements the interface
	var _ handlers.ServiceHandlerInterface = sm

	t.Log("✓ ServiceManager successfully implements ServiceHandlerInterface")
}

func TestServiceManagerInitialization(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)

	// Test that service manager is properly initialized
	if sm == nil {
		t.Fatal("ServiceManager should not be nil")
	}

	// Load services from core definitions
	err := sm.LoadServices()
	if err != nil {
		t.Fatalf("Failed to load services: %v", err)
	}

	// Test that services are loaded from core
	services := sm.ListServices()
	if len(services) == 0 {
		t.Error("Expected services to be loaded from core definitions")
	}

	// Verify some core services are present
	foundPostgres := false
	foundMySQL := false
	for _, service := range services {
		if service.Name == "postgres" {
			foundPostgres = true
		}
		if service.Name == "mysql" {
			foundMySQL = true
		}
	}

	if !foundPostgres {
		t.Error("Expected postgres service to be loaded")
	}
	if !foundMySQL {
		t.Error("Expected mysql service to be loaded")
	}
}

func TestServiceManagerBasicOperations(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("ListServices", func(t *testing.T) {
		services := sm.ListServices()
		if services == nil {
			t.Error("ListServices returned nil")
		}
		if len(services) == 0 {
			t.Error("Expected at least some services")
		}
	})

	t.Run("ListEnhancedServices", func(t *testing.T) {
		services := sm.ListEnhancedServices()
		if services == nil {
			t.Error("ListEnhancedServices returned nil")
		}
		// Should be same as ListServices
		regularServices := sm.ListServices()
		if len(services) != len(regularServices) {
			t.Error("ListEnhancedServices should return same count as ListServices")
		}
	})

	t.Run("GetService", func(t *testing.T) {
		// Test existing service
		service, exists := sm.GetService("postgres")
		if !exists {
			t.Error("Expected postgres service to exist")
		}
		if service == nil {
			t.Error("Expected service to not be nil")
		}
		if service.Name != "postgres" {
			t.Errorf("Expected service name 'postgres', got %s", service.Name)
		}

		// Test non-existent service
		_, exists = sm.GetService("non-existent")
		if exists {
			t.Error("Expected non-existent service to not exist")
		}
	})
}

func TestServiceManagerStatusOperations(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("GetServiceStatus", func(t *testing.T) {
		// Test with existing service
		status, err := sm.GetServiceStatus("postgres")
		if err != nil {
			t.Errorf("GetServiceStatus failed: %v", err)
		}
		if status != "stopped" {
			t.Errorf("Expected status 'stopped', got %s", status)
		}

		// Test with non-existent service
		_, err = sm.GetServiceStatus("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent service")
		}
	})

	t.Run("GetAllRunningServices", func(t *testing.T) {
		// Initially no services should be running
		running, err := sm.GetAllRunningServices()
		if err != nil {
			t.Errorf("GetAllRunningServices failed: %v", err)
		}
		if len(running) != 0 {
			t.Errorf("Expected 0 running services, got %d", len(running))
		}

		// Start a service in mock runtime
		mockRuntime.containers["postgres"] = "running"

		// Now should find running services
		running, err = sm.GetAllRunningServices()
		if err != nil {
			t.Errorf("GetAllRunningServices failed: %v", err)
		}
		if len(running) != 1 {
			t.Errorf("Expected 1 running service, got %d", len(running))
		}
		if _, exists := running["postgres"]; !exists {
			t.Error("Expected postgres to be in running services")
		}
	})
}

func TestServiceManagerServiceControl(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("StartService", func(t *testing.T) {
		err := sm.StartService("postgres", false)
		if err != nil {
			t.Errorf("StartService failed: %v", err)
		}

		// Verify service was started in mock runtime
		if status, exists := mockRuntime.containers["postgres"]; !exists || status != "running" {
			t.Error("Expected postgres to be running in mock runtime")
		}

		// Verify logging (StartService itself doesn't log, but status operations might)
		// Note: StartService may not generate logs directly
	})

	t.Run("StopService", func(t *testing.T) {
		// First start a service
		sm.StartService("postgres", false)

		err := sm.StopService("postgres")
		if err != nil {
			t.Errorf("StopService failed: %v", err)
		}

		// Verify service was stopped
		if status, exists := mockRuntime.containers["postgres"]; !exists || status != "stopped" {
			t.Error("Expected postgres to be stopped in mock runtime")
		}
	})

	t.Run("StopAllServices", func(t *testing.T) {
		// Start multiple services
		mockRuntime.containers["postgres"] = "running"
		mockRuntime.containers["mysql"] = "running"
		mockRuntime.containers["redis"] = "running"

		err := sm.StopAllServices()
		if err != nil {
			t.Errorf("StopAllServices failed: %v", err)
		}

		// Verify all services were stopped
		for serviceName, status := range mockRuntime.containers {
			if status != "stopped" {
				t.Errorf("Expected service %s to be stopped, got %s", serviceName, status)
			}
		}
	})
}

func TestServiceManagerDependencyOperations(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("GetServiceDependencies", func(t *testing.T) {
		deps, err := sm.GetServiceDependencies("postgres")
		if err != nil {
			t.Errorf("GetServiceDependencies failed: %v", err)
		}
		if deps == nil {
			t.Error("Expected dependencies to not be nil")
		}
		// Should at least include itself
		found := false
		for _, dep := range deps {
			if dep == "postgres" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected service to include itself in recursive dependencies")
		}
	})

	t.Run("GetAllServiceDependencies", func(t *testing.T) {
		deps, err := sm.GetAllServiceDependencies()
		if err != nil {
			t.Errorf("GetAllServiceDependencies failed: %v", err)
		}
		if deps == nil {
			t.Error("Expected dependencies to not be nil")
		}
		if len(deps) == 0 {
			t.Error("Expected at least some services to have dependencies")
		}
	})
}

func TestServiceManagerEnhancedOperations(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("GetMultipleServiceStatuses", func(t *testing.T) {
		serviceNames := []string{"postgres", "mysql", "redis"}
		statuses, err := sm.GetMultipleServiceStatuses(serviceNames)
		if err != nil {
			t.Errorf("GetMultipleServiceStatuses failed: %v", err)
		}
		if len(statuses) != 3 {
			t.Errorf("Expected 3 service statuses, got %d", len(statuses))
		}
		for _, serviceName := range serviceNames {
			if _, exists := statuses[serviceName]; !exists {
				t.Errorf("Expected status for service %s", serviceName)
			}
		}
	})

	t.Run("StartServiceWithStatusUpdate", func(t *testing.T) {
		affectedServices, err := sm.StartServiceWithStatusUpdate("postgres", false)
		if err != nil {
			t.Errorf("StartServiceWithStatusUpdate failed: %v", err)
		}
		if len(affectedServices) == 0 {
			t.Error("Expected at least one affected service")
		}
		if _, exists := affectedServices["postgres"]; !exists {
			t.Error("Expected postgres to be in affected services")
		}
	})

	t.Run("RefreshStatusFromContainers", func(t *testing.T) {
		// Set some statuses in mock runtime
		mockRuntime.containers["postgres"] = "running"
		mockRuntime.containers["mysql"] = "stopped"

		refreshed, err := sm.RefreshStatusFromContainers()
		if err != nil {
			t.Errorf("RefreshStatusFromContainers failed: %v", err)
		}
		if refreshed == nil {
			t.Error("Expected refreshed services to not be nil")
		}
		if len(refreshed) == 0 {
			t.Error("Expected at least some services to be refreshed")
		}
	})

	t.Run("CheckStartingServicesProgress", func(t *testing.T) {
		// Manually set a service to starting state
		service, exists := sm.GetService("postgres")
		if !exists {
			t.Fatal("Expected postgres service to exist")
		}
		service.Status = "starting"

		progress, err := sm.CheckStartingServicesProgress()
		if err != nil {
			t.Errorf("CheckStartingServicesProgress failed: %v", err)
		}
		if progress == nil {
			t.Error("Expected progress to not be nil")
		}
	})
}

func TestServiceManagerErrorHandling(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("ErrorMode", func(t *testing.T) {
		// Enable error mode
		mockRuntime.errorMode = true

		// Test operations should handle errors gracefully
		_, err := sm.GetServiceStatus("postgres")
		if err == nil {
			t.Error("Expected error when runtime is in error mode")
		}

		err = sm.StartService("postgres", false)
		if err == nil {
			t.Error("Expected error when runtime is in error mode")
		}

		_, err = sm.GetAllRunningServices()
		if err == nil {
			t.Error("Expected error when runtime is in error mode")
		}
	})
}

func TestEnhancedServiceFields(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	t.Run("EnhancedServiceStructure", func(t *testing.T) {
		service, exists := sm.GetService("postgres")
		if !exists {
			t.Fatal("Expected postgres service to exist")
		}

		// Test that enhanced fields are present
		if service.Name == "" {
			t.Error("Expected service name to be set")
		}
		if service.Type == "" {
			t.Error("Expected service type to be set")
		}
		if service.ConnectionCmd == "" {
			t.Error("Expected connection command to be set")
		}
		if service.ContainerName == "" {
			t.Error("Expected container name to be set")
		}
		if service.AllContainers == nil {
			t.Error("Expected all containers to be initialized")
		}
		if service.RecursiveDependencies == nil {
			t.Error("Expected recursive dependencies to be initialized")
		}
		if service.ExposedPorts == nil {
			t.Error("Expected exposed ports to be initialized")
		}
		if service.WebUrls == nil {
			t.Error("Expected web URLs to be initialized")
		}
		// Note: These might be empty slices, which is fine

		// Test time fields
		if service.LastUpdated.IsZero() {
			t.Error("Expected last updated time to be set")
		}
	})

	t.Run("ServiceDataIntegrity", func(t *testing.T) {
		services := sm.ListServices()

		for _, service := range services {
			// Basic validation
			if service.Name == "" {
				t.Errorf("Service has empty name")
			}
			if service.ContainerName == "" {
				t.Errorf("Service %s has empty container name", service.Name)
			}
			if len(service.AllContainers) == 0 {
				t.Errorf("Service %s has no containers defined", service.Name)
			}
			if service.RecursiveDependencies == nil {
				t.Errorf("Service %s has nil recursive dependencies", service.Name)
			}

			// Verify recursive dependencies include self
			foundSelf := false
			for _, dep := range service.RecursiveDependencies {
				if dep == service.Name {
					foundSelf = true
					break
				}
			}
			if !foundSelf {
				t.Errorf("Service %s should include itself in recursive dependencies", service.Name)
			}
		}
	})
}

func TestServiceManagerLogging(t *testing.T) {
	mockRuntime := NewTestMockRuntimeInfo()
	mockLogger := NewTestMockLogger()

	sm := models.NewServiceManager("/tmp/test", mockRuntime, mockLogger)
	sm.LoadServices()

	// Perform operations that should generate logs
	sm.GetServiceStatus("postgres")
	sm.StartService("postgres", false)
	sm.StopService("postgres")
	sm.GetAllRunningServices()
	sm.RefreshStatusFromContainers()

	// Verify logs were generated
	if len(mockLogger.logs) == 0 {
		t.Error("Expected log messages to be generated")
	}

	// Check for specific log patterns
	found := false
	for _, log := range mockLogger.logs {
		if log == "Getting status for service: postgres" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected specific log message not found")
	}
}
