package handlers

import (
	"errors"
	"testing"

	"github.com/data-catering/insta-infra/v2/internal/core"
)

func TestConnectionHandler_NewConnectionHandler(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	instaDir := "/test/insta"

	handler := NewConnectionHandler(mockRuntime, instaDir)

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

func TestConnectionHandler_GetServiceConnectionInfo_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"5432/tcp": "5432"}, nil
		})
	handler := NewConnectionHandler(mockRuntime, "/test/insta")

	connInfo, err := handler.GetServiceConnectionInfo("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if connInfo == nil {
		t.Fatal("Expected connection info to be returned, got nil")
	}
	if connInfo.ServiceName != "postgres" {
		t.Errorf("Expected service name 'postgres', got '%s'", connInfo.ServiceName)
	}
	if !connInfo.Available {
		t.Error("Expected service to be available")
	}
	if connInfo.HostPort != "5432" {
		t.Errorf("Expected host port '5432', got '%s'", connInfo.HostPort)
	}
	if connInfo.ContainerPort != "5432" {
		t.Errorf("Expected container port '5432', got '%s'", connInfo.ContainerPort)
	}
}

func TestConnectionHandler_GetServiceConnectionInfo_ServiceNotRunning(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return nil, errors.New("container not running")
		})
	handler := NewConnectionHandler(mockRuntime, "/test/insta")

	connInfo, err := handler.GetServiceConnectionInfo("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if connInfo == nil {
		t.Fatal("Expected connection info to be returned, got nil")
	}
	if connInfo.Available {
		t.Error("Expected service to not be available")
	}
	if connInfo.Error == "" {
		t.Error("Expected error message to be set")
	}
}

func TestConnectionHandler_GetServiceConnectionInfo_ContainerNameError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		})
	handler := NewConnectionHandler(mockRuntime, "/test/insta")

	// Use a real service name that exists in core.Services but mock the container name error
	connInfo, err := handler.GetServiceConnectionInfo("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if connInfo == nil {
		t.Fatal("Expected connection info to be returned, got nil")
	}
	if connInfo.Available {
		t.Error("Expected service to not be available")
	}
	if !contains(connInfo.Error, "service not running") {
		t.Errorf("Expected error to contain 'service not running', got %s", connInfo.Error)
	}
}

func TestConnectionHandler_GetServiceConnectionInfo_UnknownService(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	handler := NewConnectionHandler(mockRuntime, "/test/insta")

	connInfo, err := handler.GetServiceConnectionInfo("unknown_service")

	if err == nil {
		t.Fatal("Expected error for unknown service, got nil")
	}
	if !contains(err.Error(), "unknown service") {
		t.Errorf("Expected error to contain 'unknown service', got %s", err.Error())
	}
	if connInfo != nil {
		t.Errorf("Expected connection info to be nil for unknown service, got %v", connInfo)
	}
}

func TestConnectionHandler_GetServiceConnectionInfo_WebUIService(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_grafana_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"3000/tcp": "3000"}, nil
		})
	handler := NewConnectionHandler(mockRuntime, "/test/insta")

	connInfo, err := handler.GetServiceConnectionInfo("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !connInfo.HasWebUI {
		t.Error("Expected Grafana to have web UI")
	}
	if connInfo.WebURL != "http://localhost:3000" {
		t.Errorf("Expected web URL 'http://localhost:3000', got '%s'", connInfo.WebURL)
	}
}

func TestConnectionHandler_hasWebUI(t *testing.T) {
	handler := NewConnectionHandler(nil, "/test/insta")

	testCases := []struct {
		serviceName string
		expected    bool
	}{
		{"grafana", true},
		{"superset", true},
		{"minio", true},
		{"postgres", false},
		{"redis", false},
		{"mysql", false},
		{"unknown", false},
	}

	for _, tc := range testCases {
		result := handler.hasWebUI(tc.serviceName)
		if result != tc.expected {
			t.Errorf("hasWebUI(%s): expected %t, got %t", tc.serviceName, tc.expected, result)
		}
	}
}

func TestConnectionHandler_buildConnectionString(t *testing.T) {
	handler := NewConnectionHandler(nil, "/test/insta")

	testCases := []struct {
		serviceName string
		serviceType string
		hostPort    string
		expected    string
	}{
		{"postgres", "Database", "5432", "postgresql://postgres:postgres@localhost:5432/postgres"},
		{"mysql", "Database", "3306", "mysql://root:root@localhost:3306/"},
		{"redis", "Database", "6379", "redis://localhost:6379"},
		{"grafana", "WebUI", "3000", "http://localhost:3000"},
		{"unknown", "Unknown", "8080", "localhost:8080"},
	}

	for _, tc := range testCases {
		// Create a mock service with the right type and credentials
		service := createMockService(tc.serviceName, tc.serviceType)
		result := handler.buildConnectionString(tc.serviceName, service, tc.hostPort)

		if !contains(result, tc.hostPort) {
			t.Errorf("buildConnectionString(%s): expected result to contain port %s, got %s",
				tc.serviceName, tc.hostPort, result)
		}
	}
}

func TestConnectionHandler_getComposeFiles(t *testing.T) {
	handler := NewConnectionHandler(nil, "/test/insta")

	composeFiles := handler.getComposeFiles()

	if len(composeFiles) == 0 {
		t.Error("Expected at least one compose file")
	}
	if !contains(composeFiles[0], "/test/insta/docker-compose.yaml") {
		t.Errorf("Expected first compose file to contain '/test/insta/docker-compose.yaml', got %s", composeFiles[0])
	}
}

// Helper function to create mock services for testing
func createMockService(name, serviceType string) core.Service {
	service := core.Service{
		Name: name,
		Type: serviceType,
	}

	// Set default credentials based on service type
	switch name {
	case "postgres":
		service.DefaultUser = "postgres"
		service.DefaultPassword = "postgres"
	case "mysql":
		service.DefaultUser = "root"
		service.DefaultPassword = "root"
	case "redis":
		service.DefaultUser = ""
		service.DefaultPassword = ""
	}

	return service
}
