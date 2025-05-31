package container

import (
	"fmt"
	"os"
	"os/exec"
)

// MockRuntime implements Runtime interface for testing across all test files.
// This consolidates the previously duplicated MockRuntime and MockRuntimeForProvider.
type MockRuntime struct {
	name                  string
	available             bool
	composeUpCalled       bool
	composeDownCalled     bool
	execCalled            bool
	lastServices          []string
	lastCmd               string
	portMappings          map[string]string
	dependencies          []string
	containerNames        map[string]string
	imageExists           bool
	logs                  []string
	containerStatus       string
	shouldFailComposeUp   bool
	shouldFailComposeDown bool
	shouldFailExec        bool
}

// NewMockRuntime creates a new mock runtime with sensible defaults
func NewMockRuntime(name string, available bool) *MockRuntime {
	return &MockRuntime{
		name:            name,
		available:       available,
		portMappings:    make(map[string]string),
		dependencies:    []string{},
		containerNames:  make(map[string]string),
		imageExists:     true,
		logs:            []string{"Mock log line 1", "Mock log line 2", "Mock log line 3"},
		containerStatus: "running",
	}
}

// WithPortMappings sets the port mappings for the mock runtime
func (m *MockRuntime) WithPortMappings(mappings map[string]string) *MockRuntime {
	m.portMappings = mappings
	return m
}

// WithDependencies sets the dependencies for the mock runtime
func (m *MockRuntime) WithDependencies(deps []string) *MockRuntime {
	m.dependencies = deps
	return m
}

// WithContainerNames sets the container name mappings for the mock runtime
func (m *MockRuntime) WithContainerNames(names map[string]string) *MockRuntime {
	m.containerNames = names
	return m
}

// WithImageExists sets whether images exist for the mock runtime
func (m *MockRuntime) WithImageExists(exists bool) *MockRuntime {
	m.imageExists = exists
	return m
}

// WithLogs sets the log lines for the mock runtime
func (m *MockRuntime) WithLogs(logs []string) *MockRuntime {
	m.logs = logs
	return m
}

// WithContainerStatus sets the container status for the mock runtime
func (m *MockRuntime) WithContainerStatus(status string) *MockRuntime {
	m.containerStatus = status
	return m
}

// WithFailures configures the mock to fail on specific operations
func (m *MockRuntime) WithFailures(composeUp, composeDown, exec bool) *MockRuntime {
	m.shouldFailComposeUp = composeUp
	m.shouldFailComposeDown = composeDown
	m.shouldFailExec = exec
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
	if m.shouldFailComposeUp {
		return fmt.Errorf("mock compose up failure")
	}
	m.composeUpCalled = true
	m.lastServices = services
	return nil
}

func (m *MockRuntime) ComposeDown(composeFiles []string, services []string) error {
	if m.shouldFailComposeDown {
		return fmt.Errorf("mock compose down failure")
	}
	m.composeDownCalled = true
	m.lastServices = services
	return nil
}

func (m *MockRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	if m.shouldFailExec {
		return fmt.Errorf("mock exec failure")
	}
	m.execCalled = true
	m.lastCmd = cmd
	return nil
}

func (m *MockRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	if m.portMappings != nil {
		return m.portMappings, nil
	}
	return map[string]string{}, nil
}

func (m *MockRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	return m.dependencies, nil
}

func (m *MockRuntime) GetAllDependenciesRecursive(service string, composeFiles []string) ([]string, error) {
	return m.dependencies, nil
}

func (m *MockRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
	if cn, ok := m.containerNames[serviceName]; ok && cn != "" {
		return cn, nil
	}
	return fmt.Sprintf("mock_%s_1", serviceName), nil
}

func (m *MockRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	return m.logs, nil
}

func (m *MockRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	go func() {
		defer close(logChan)
		for _, log := range m.logs {
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
	return m.imageExists, nil
}

func (m *MockRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	return fmt.Sprintf("mock/%s:latest", serviceName), nil
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
	return m.containerStatus, nil
}

// Test state accessors
func (m *MockRuntime) WasComposeUpCalled() bool   { return m.composeUpCalled }
func (m *MockRuntime) WasComposeDownCalled() bool { return m.composeDownCalled }
func (m *MockRuntime) WasExecCalled() bool        { return m.execCalled }
func (m *MockRuntime) GetLastServices() []string  { return m.lastServices }
func (m *MockRuntime) GetLastCmd() string         { return m.lastCmd }

// Reset clears all call tracking state
func (m *MockRuntime) Reset() {
	m.composeUpCalled = false
	m.composeDownCalled = false
	m.execCalled = false
	m.lastServices = nil
	m.lastCmd = ""
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
