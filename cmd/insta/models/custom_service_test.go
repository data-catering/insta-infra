package models

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/data-catering/insta-infra/v2/internal/core"
)

func TestNewCustomServiceRegistry(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test creating new registry
	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create custom service registry: %v", err)
	}

	// Check that custom directory was created
	customDir := filepath.Join(tempDir, "custom")
	if _, err := os.Stat(customDir); os.IsNotExist(err) {
		t.Error("Custom directory was not created")
	}

	// Check that registry was initialized properly
	if registry.CustomDir != customDir {
		t.Errorf("Expected CustomDir %s, got %s", customDir, registry.CustomDir)
	}

	if registry.Services == nil {
		t.Error("Services map was not initialized")
	}

	if registry.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", registry.Version)
	}
}

func TestAddCustomService(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test adding a valid custom service
	composeContent := `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      - NGINX_HOST=localhost
  db:
    image: postgres:13
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_PASSWORD=password
`

	metadata, err := registry.AddCustomService("test-app", "Test application", composeContent)
	if err != nil {
		t.Fatalf("Failed to add custom service: %v", err)
	}

	// Check metadata
	if metadata.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got %s", metadata.Name)
	}

	if metadata.Description != "Test application" {
		t.Errorf("Expected description 'Test application', got %s", metadata.Description)
	}

	if len(metadata.Services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(metadata.Services))
	}

	// Check that file was created
	filePath := filepath.Join(registry.CustomDir, metadata.Filename)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Compose file was not created")
	}

	// Check that metadata was saved
	metadataFile := filepath.Join(registry.CustomDir, "metadata.json")
	if _, err := os.Stat(metadataFile); os.IsNotExist(err) {
		t.Error("Metadata file was not created")
	}
}

func TestAddCustomServiceInvalidYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test adding invalid YAML
	invalidContent := `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
    environment:
      - NGINX_HOST=localhost
  - invalid yaml structure
`

	_, err = registry.AddCustomService("invalid-app", "Invalid app", invalidContent)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}

func TestAddCustomServiceMissingServices(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Test adding compose file without services section
	invalidContent := `
version: '3.8'
networks:
  default:
    driver: bridge
`

	_, err = registry.AddCustomService("no-services", "No services app", invalidContent)
	if err == nil {
		t.Error("Expected error for missing services section, got nil")
	}
}

func TestUpdateCustomService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Add initial service
	originalContent := `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`

	metadata, err := registry.AddCustomService("test-app", "Test application", originalContent)
	if err != nil {
		t.Fatalf("Failed to add custom service: %v", err)
	}

	// Add a small delay to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	// Update the service
	updatedContent := `
services:
  web:
    image: nginx:alpine
    ports:
      - "8081:80"
  api:
    image: node:16
    ports:
      - "3000:3000"
`

	updatedMetadata, err := registry.UpdateCustomService(metadata.ID, "updated-app", "Updated application", updatedContent)
	if err != nil {
		t.Fatalf("Failed to update custom service: %v", err)
	}

	// Check updated metadata
	if updatedMetadata.Name != "updated-app" {
		t.Errorf("Expected updated name 'updated-app', got %s", updatedMetadata.Name)
	}

	if updatedMetadata.Description != "Updated application" {
		t.Errorf("Expected updated description 'Updated application', got %s", updatedMetadata.Description)
	}

	if len(updatedMetadata.Services) != 2 {
		t.Errorf("Expected 2 services after update, got %d", len(updatedMetadata.Services))
	}

	// Check that UpdatedAt was changed (allow for same time in case of very fast execution)
	if updatedMetadata.UpdatedAt.Before(metadata.UpdatedAt) {
		t.Error("UpdatedAt was not updated properly")
	}
}

func TestRemoveCustomService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Add service
	composeContent := `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`

	metadata, err := registry.AddCustomService("test-app", "Test application", composeContent)
	if err != nil {
		t.Fatalf("Failed to add custom service: %v", err)
	}

	// Check that service exists
	_, err = registry.GetCustomService(metadata.ID)
	if err != nil {
		t.Fatalf("Service should exist: %v", err)
	}

	// Remove service
	err = registry.RemoveCustomService(metadata.ID)
	if err != nil {
		t.Fatalf("Failed to remove custom service: %v", err)
	}

	// Check that service no longer exists
	_, err = registry.GetCustomService(metadata.ID)
	if err == nil {
		t.Error("Service should not exist after removal")
	}

	// Check that file was removed
	filePath := filepath.Join(registry.CustomDir, metadata.Filename)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Compose file should have been removed")
	}
}

func TestValidateComposeContent(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		shouldErr bool
	}{
		{
			name: "valid compose file",
			content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
			shouldErr: false,
		},
		{
			name: "invalid YAML",
			content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  - invalid structure
`,
			shouldErr: true,
		},
		{
			name: "missing services section",
			content: `
version: '3.8'
networks:
  default:
    driver: bridge
`,
			shouldErr: true,
		},
		{
			name: "empty services section",
			content: `
services: {}
`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateComposeContent(tt.content)
			if tt.shouldErr && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestSyncWithFilesystem(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Manually create a compose file in the custom directory
	customDir := filepath.Join(tempDir, "custom")
	composeFile := filepath.Join(customDir, "manual-service.yaml")
	composeContent := `
services:
  manual:
    image: alpine:latest
    command: sleep infinity
`

	err = os.WriteFile(composeFile, []byte(composeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write manual compose file: %v", err)
	}

	// Sync with filesystem
	err = registry.syncWithFilesystem()
	if err != nil {
		t.Fatalf("Failed to sync with filesystem: %v", err)
	}

	// Check that the manual file was discovered
	found := false
	for _, metadata := range registry.Services {
		if metadata.Filename == "manual-service.yaml" {
			found = true
			if len(metadata.Services) != 1 || metadata.Services[0] != "manual" {
				t.Errorf("Expected service 'manual', got %v", metadata.Services)
			}
			break
		}
	}

	if !found {
		t.Error("Manual compose file was not discovered during sync")
	}
}

func TestRegisterCustomServices(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create registry: %v", err)
	}

	// Add a custom service
	composeContent := `
services:
  custom-web:
    image: nginx:latest
    ports:
      - "8080:80"
      - "8443:443"
    expose:
      - "9000"
`

	_, err = registry.AddCustomService("test-web", "Test web service", composeContent)
	if err != nil {
		t.Fatalf("Failed to add custom service: %v", err)
	}

	// Store original services count
	originalCount := len(core.Services)

	// Register all custom services
	err = registry.RegisterAllCustomServices()
	if err != nil {
		t.Fatalf("Failed to register custom services: %v", err)
	}

	// Check that custom service was registered
	customService, exists := core.Services["custom-web"]
	if !exists {
		t.Error("Custom service was not registered in core.Services")
	}

	if customService.Type != "Custom" {
		t.Errorf("Expected service type 'Custom', got %s", customService.Type)
	}

	if len(customService.Ports) != 3 {
		t.Errorf("Expected 3 ports, got %d", len(customService.Ports))
	}

	// Clean up
	registry.UnregisterAllCustomServices()

	// Check that service was unregistered
	_, exists = core.Services["custom-web"]
	if exists {
		t.Error("Custom service was not unregistered from core.Services")
	}

	// Check that original services are still there
	if len(core.Services) != originalCount {
		t.Errorf("Expected %d services after cleanup, got %d", originalCount, len(core.Services))
	}
}

func TestParsePortMapping(t *testing.T) {
	tests := []struct {
		portStr  string
		expected *core.ServicePort
	}{
		{
			portStr: "8080:80",
			expected: &core.ServicePort{
				InternalPort: 80,
				Type:         core.PortTypeWebUI,
				Name:         "Port 80",
				Description:  "Custom service port 80",
			},
		},
		{
			portStr: "3000:3000/tcp",
			expected: &core.ServicePort{
				InternalPort: 3000,
				Type:         core.PortTypeWebUI,
				Name:         "Port 3000",
				Description:  "Custom service port 3000",
			},
		},
		{
			portStr: "127.0.0.1:5432:5432",
			expected: &core.ServicePort{
				InternalPort: 5432,
				Type:         core.PortTypeDatabase,
				Name:         "Port 5432",
				Description:  "Custom service port 5432",
			},
		},
		{
			portStr:  "invalid",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.portStr, func(t *testing.T) {
			result := parsePortMapping(tt.portStr)
			if tt.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %+v", result)
				}
				return
			}

			if result == nil {
				t.Error("Expected port mapping, got nil")
				return
			}

			if result.InternalPort != tt.expected.InternalPort {
				t.Errorf("Expected InternalPort %d, got %d", tt.expected.InternalPort, result.InternalPort)
			}

			if result.Type != tt.expected.Type {
				t.Errorf("Expected Type %s, got %s", tt.expected.Type, result.Type)
			}
		})
	}
}

func TestListCustomServices(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create custom service registry: %v", err)
	}

	// Initially should be empty
	services := registry.ListCustomServices()
	if len(services) != 0 {
		t.Errorf("Expected empty list, got %d services", len(services))
	}

	// Add a service
	composeContent := `version: '3.8'
services:
  test-app:
    image: nginx:latest
    ports:
      - "8080:80"`

	metadata, err := registry.AddCustomService("test-service", "Test service", composeContent)
	if err != nil {
		t.Fatalf("Failed to add custom service: %v", err)
	}

	// Should now have one service
	services = registry.ListCustomServices()
	if len(services) != 1 {
		t.Errorf("Expected 1 service, got %d", len(services))
	}

	if services[0].ID != metadata.ID {
		t.Errorf("Expected service ID %s, got %s", metadata.ID, services[0].ID)
	}
}

func TestGetAllCustomComposeFiles(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create custom service registry: %v", err)
	}

	// Initially should be empty
	files := registry.GetAllCustomComposeFiles()
	if len(files) != 0 {
		t.Errorf("Expected empty list, got %d files", len(files))
	}

	// Add a service
	composeContent := `version: '3.8'
services:
  test-app:
    image: nginx:latest`

	_, err = registry.AddCustomService("test-service", "Test service", composeContent)
	if err != nil {
		t.Fatalf("Failed to add custom service: %v", err)
	}

	// Should now have one file
	files = registry.GetAllCustomComposeFiles()
	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	// Check that the file path exists
	if _, err := os.Stat(files[0]); os.IsNotExist(err) {
		t.Errorf("Custom compose file does not exist: %s", files[0])
	}
}

func TestGetServiceClashes(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	registry, err := NewCustomServiceRegistry(tempDir)
	if err != nil {
		t.Fatalf("Failed to create custom service registry: %v", err)
	}

	// Initially should be empty
	clashes := registry.GetServiceClashes()
	if len(clashes) != 0 {
		t.Errorf("Expected empty clashes, got %d", len(clashes))
	}

	// Clear clashes (should not error)
	registry.ClearServiceClashes()

	// Verify still empty
	clashes = registry.GetServiceClashes()
	if len(clashes) != 0 {
		t.Errorf("Expected empty clashes after clear, got %d", len(clashes))
	}
}