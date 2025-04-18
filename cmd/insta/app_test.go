package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// MockRuntime for testing App
type MockRuntime struct {
	name              string
	available         bool
	composeUpCalled   bool
	composeDownCalled bool
	execCalled        bool
	composeFiles      []string
	services          []string
	lastContainer     string
	lastCmd           string
	portMappings      map[string]string
	dependencies      []string
}

func NewMockRuntime() *MockRuntime {
	return &MockRuntime{
		name:         "mock-runtime",
		portMappings: make(map[string]string),
		dependencies: []string{},
	}
}

func (m *MockRuntime) Name() string {
	return m.name
}

func (m *MockRuntime) CheckAvailable() error {
	return nil
}

func (m *MockRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	m.composeUpCalled = true
	m.composeFiles = composeFiles
	m.services = services
	return nil
}

func (m *MockRuntime) ComposeDown(composeFiles []string, services []string) error {
	m.composeDownCalled = true
	m.composeFiles = composeFiles
	m.services = services
	return nil
}

func (m *MockRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	m.execCalled = true
	m.lastContainer = containerName
	m.lastCmd = cmd
	return nil
}

func (m *MockRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	// Only return port mappings for specific containers
	switch containerName {
	case "postgres", "mysql":
		return m.portMappings, nil
	default:
		// Return empty map for services without ports
		return map[string]string{}, nil
	}
}

func (m *MockRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	return m.dependencies, nil
}

func TestAppWithMockRuntime(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "app-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Setup App with mock runtime
	mockRuntime := NewMockRuntime()
	
	// Add some port mappings to the mock runtime
	mockRuntime.portMappings["5432/tcp"] = "5432"
	
	app := &App{
		dataDir:  filepath.Join(tempDir, "data"),
		instaDir: tempDir,
		runtime:  mockRuntime,
	}

	// Test starting services
	t.Run("start services", func(t *testing.T) {
		err := app.startServices([]string{"postgres", "mysql"}, false)
		if err != nil {
			t.Errorf("startServices failed: %v", err)
		}

		if !mockRuntime.composeUpCalled {
			t.Error("Expected ComposeUp to be called")
		}

		if len(mockRuntime.services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(mockRuntime.services))
		}

		if !contains(mockRuntime.services, "postgres") || !contains(mockRuntime.services, "mysql") {
			t.Errorf("Expected services to include postgres and mysql, got %v", mockRuntime.services)
		}
	})

	// Reset mock state
	mockRuntime.composeUpCalled = false
	mockRuntime.services = nil

	// Test persisting data
	t.Run("start services with persist", func(t *testing.T) {
		err := app.startServices([]string{"postgres"}, true)
		if err != nil {
			t.Errorf("startServices with persist failed: %v", err)
		}

		if !mockRuntime.composeUpCalled {
			t.Error("Expected ComposeUp to be called")
		}

		if !contains(mockRuntime.services, "postgres") {
			t.Errorf("Expected services to include postgres, got %v", mockRuntime.services)
		}

		// Check if data directory was created
		dataDirExists, err := dirExists(app.dataDir)
		if err != nil {
			t.Fatalf("Error checking data directory: %v", err)
		}
		if !dataDirExists {
			t.Error("Expected data directory to be created")
		}

		// Check if compose files includes persist file
		persistFileIncluded := false
		for _, file := range mockRuntime.composeFiles {
			if strings.Contains(file, "persist") {
				persistFileIncluded = true
				break
			}
		}
		if !persistFileIncluded {
			t.Error("Expected persist compose file to be included")
		}
	})

	// Test stopping services
	t.Run("stop services", func(t *testing.T) {
		err := app.stopServices([]string{"postgres"})
		if err != nil {
			t.Errorf("stopServices failed: %v", err)
		}

		if !mockRuntime.composeDownCalled {
			t.Error("Expected ComposeDown to be called")
		}

		if !contains(mockRuntime.services, "postgres") {
			t.Errorf("Expected services to include postgres, got %v", mockRuntime.services)
		}
	})

	// Test connecting to a service
	t.Run("connect to service", func(t *testing.T) {
		err := app.connectToService("postgres")
		if err != nil {
			t.Errorf("connectToService failed: %v", err)
		}

		if !mockRuntime.execCalled {
			t.Error("Expected ExecInContainer to be called")
		}

		if mockRuntime.lastContainer != "postgres" {
			t.Errorf("Expected container name postgres, got %s", mockRuntime.lastContainer)
		}

		// Verify command includes credentials for postgres
		if !strings.Contains(mockRuntime.lastCmd, "POSTGRES_USER") {
			t.Errorf("Expected command to include POSTGRES_USER, got %s", mockRuntime.lastCmd)
		}
	})
}

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// Helper function to check if a directory exists
func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}
