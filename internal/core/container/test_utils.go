package container

import (
	"fmt"
	"os"
	"os/exec"
)

// MockRuntime implements Runtime interface for testing across all test files.
// This consolidates the previously duplicated MockRuntime and MockRuntimeForProvider.
type MockRuntime struct {
	name                   string
	available              bool
	composeUpError         error
	composeDownError       error
	execError              error
	portMappings           map[string]string
	portMappingsError      error
	containerName          string
	containerNameError     error
	containerLogs          []string
	containerLogsError     error
	imageExists            bool
	imageExistsError       error
	imageInfo              string
	imageInfoError         error
	containerStatus        string
	containerStatusError   error
	runningContainers      map[string]string
	runningContainersError error
}

// NewMockRuntime creates a new mock runtime with sensible defaults
func NewMockRuntime(name string, available bool) *MockRuntime {
	return &MockRuntime{
		name:            name,
		available:       available,
		portMappings:    make(map[string]string),
		imageExists:     true,
		containerStatus: "running",
	}
}

// WithPortMappings sets the port mappings for the mock runtime
func (m *MockRuntime) WithPortMappings(mappings map[string]string) *MockRuntime {
	m.portMappings = mappings
	return m
}

// WithImageExists sets whether images exist for the mock runtime
func (m *MockRuntime) WithImageExists(exists bool) *MockRuntime {
	m.imageExists = exists
	return m
}

// WithLogs sets the log lines for the mock runtime
func (m *MockRuntime) WithLogs(logs []string) *MockRuntime {
	m.containerLogs = logs
	return m
}

// WithContainerStatus sets the container status for the mock runtime
func (m *MockRuntime) WithContainerStatus(status string) *MockRuntime {
	m.containerStatus = status
	return m
}

// Runtime interface implementation
func (m *MockRuntime) Name() string {
	return m.name
}

func (m *MockRuntime) CheckAvailable() error {
	if !m.available {
		return &exec.Error{Name: m.name, Err: exec.ErrNotFound}
	}
	return nil
}

func (m *MockRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	return m.composeUpError
}

func (m *MockRuntime) ComposeDown(composeFiles []string, services []string) error {
	return m.composeDownError
}

func (m *MockRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	return m.execError
}

func (m *MockRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	return m.portMappings, m.portMappingsError
}

func (m *MockRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
	if m.containerName != "" {
		return m.containerName, m.containerNameError
	}
	return fmt.Sprintf("mock_%s_1", serviceName), m.containerNameError
}

// GetAllDependenciesRecursive returns all dependencies recursively for a service from compose files
func (m *MockRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string, isContainer bool) ([]string, error) {
	// For testing, return empty dependencies by default
	return []string{}, nil
}

func (m *MockRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	return m.containerLogs, m.containerLogsError
}

func (m *MockRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	go func() {
		defer close(logChan)
		for _, log := range m.containerLogs {
			select {
			case logChan <- log:
			case <-stopChan:
				return
			}
		}
	}()
	return nil
}

func (m *MockRuntime) CheckImageExists(imageName string) (bool, error) {
	return m.imageExists, m.imageExistsError
}

// CheckMultipleImagesExist checks if multiple images exist locally in a single call
func (m *MockRuntime) CheckMultipleImagesExist(imageNames []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, imageName := range imageNames {
		result[imageName] = m.imageExists
	}
	return result, nil
}

func (m *MockRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	return m.imageInfo, m.imageInfoError
}

// GetMultipleImageInfo returns image information for multiple services from compose files
func (m *MockRuntime) GetMultipleImageInfo(serviceNames []string, composeFiles []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, serviceName := range serviceNames {
		result[serviceName] = fmt.Sprintf("mock/%s:latest", serviceName)
	}
	return result, nil
}

func (m *MockRuntime) PullImageWithProgress(imageName string, progressChan chan<- ImagePullProgress, stopChan <-chan struct{}) error {
	go func() {
		defer close(progressChan)

		progress := []ImagePullProgress{
			{Status: "downloading", Progress: 25.0},
			{Status: "downloading", Progress: 75.0},
			{Status: "complete", Progress: 100.0},
		}

		for _, p := range progress {
			select {
			case progressChan <- p:
			case <-stopChan:
				return
			}
		}
	}()
	return nil
}

func (m *MockRuntime) GetContainerStatus(containerName string) (string, error) {
	return m.containerStatus, m.containerStatusError
}

// GetAllContainerStatuses returns all running containers managed by compose
func (m *MockRuntime) GetAllContainerStatuses() (map[string]string, error) {
	return m.runningContainers, m.runningContainersError
}

// Test state accessors
func (m *MockRuntime) WasComposeUpCalled() bool   { return m.composeUpError != nil }
func (m *MockRuntime) WasComposeDownCalled() bool { return m.composeDownError != nil }
func (m *MockRuntime) WasExecCalled() bool        { return m.execError != nil }

// Reset clears all call tracking state
func (m *MockRuntime) Reset() {
	m.composeUpError = nil
	m.composeDownError = nil
	m.execError = nil
	m.portMappings = nil
	m.portMappingsError = nil
	m.containerName = ""
	m.containerNameError = nil
	m.containerLogs = nil
	m.containerLogsError = nil
	m.imageExists = true
	m.imageExistsError = nil
	m.imageInfo = ""
	m.imageInfoError = nil
	m.containerStatus = "running"
	m.containerStatusError = nil
	m.runningContainers = map[string]string{}
	m.runningContainersError = nil
}

// Common test utilities

// CreateTestComposeService creates a test compose service with the given dependencies
func CreateTestComposeService(deps []string) ComposeService {
	dependsOn := make(map[string]struct {
		Condition string `json:"condition"`
	})

	for _, dep := range deps {
		dependsOn[dep] = struct {
			Condition string `json:"condition"`
		}{Condition: "service_started"}
	}

	return ComposeService{DependsOn: dependsOn}
}

// CreateTestComposeConfig creates a test compose config with the given services
func CreateTestComposeConfig(services map[string][]string) ComposeConfig {
	config := ComposeConfig{
		Services: make(map[string]ComposeService),
	}

	for serviceName, deps := range services {
		config.Services[serviceName] = CreateTestComposeService(deps)
	}

	return config
}

// CreateTestProvider creates a provider with mock runtimes for testing
func CreateTestProvider(runtimes ...Runtime) *Provider {
	return &Provider{runtimes: runtimes}
}

// SaveAndRestoreEnvVars saves environment variables and returns a restore function
func SaveAndRestoreEnvVars(vars []string) func() {
	originalValues := make(map[string]string)
	for _, envVar := range vars {
		originalValues[envVar] = os.Getenv(envVar)
	}

	return func() {
		for envVar, originalValue := range originalValues {
			if originalValue == "" {
				os.Unsetenv(envVar)
			} else {
				os.Setenv(envVar, originalValue)
			}
		}
	}
}
