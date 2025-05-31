package main

import (
	"fmt"
	"os"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// MockContainerRuntime provides a comprehensive mock implementation for testing.
// This consolidates the duplicate mockContainerRuntime implementations across test files.
type MockContainerRuntime struct {
	name                            string
	checkAvailableFunc              func() error
	composeUpFunc                   func(composeFiles []string, services []string, quiet bool) error
	composeDownFunc                 func(composeFiles []string, services []string) error
	execInContainerFunc             func(containerName string, cmd string, interactive bool) error
	getPortMappingsFunc             func(containerName string) (map[string]string, error)
	getDependenciesFunc             func(service string, composeFiles []string) ([]string, error)
	getContainerNameFunc            func(serviceName string, composeFiles []string) (string, error)
	getAllDependenciesRecursiveFunc func(serviceName string, composeFiles []string) ([]string, error)
	getContainerLogsFunc            func(containerName string, tailLines int) ([]string, error)
	streamContainerLogsFunc         func(containerName string, logChan chan<- string, stopChan <-chan struct{}) error
	checkImageExistsFunc            func(imageName string) (bool, error)
	getImageInfoFunc                func(serviceName string, composeFiles []string) (string, error)
	pullImageWithProgressFunc       func(imageName string, progressChan chan<- container.ImagePullProgress, stopChan <-chan struct{}) error
	getContainerStatusFunc          func(containerName string) (string, error)

	// Default return values
	defaultPortMappings    map[string]string
	defaultDependencies    []string
	defaultLogs            []string
	defaultContainerStatus string
	defaultImageExists     bool
}

// NewMockContainerRuntime creates a new mock container runtime with sensible defaults
func NewMockContainerRuntime(name string) *MockContainerRuntime {
	return &MockContainerRuntime{
		name:                   name,
		defaultPortMappings:    map[string]string{"80/tcp": "8080"},
		defaultDependencies:    []string{},
		defaultLogs:            []string{"test log"},
		defaultContainerStatus: "running",
		defaultImageExists:     true,
	}
}

// WithCheckAvailable sets a custom CheckAvailable function
func (m *MockContainerRuntime) WithCheckAvailable(fn func() error) *MockContainerRuntime {
	m.checkAvailableFunc = fn
	return m
}

// WithComposeUp sets a custom ComposeUp function
func (m *MockContainerRuntime) WithComposeUp(fn func([]string, []string, bool) error) *MockContainerRuntime {
	m.composeUpFunc = fn
	return m
}

// WithComposeDown sets a custom ComposeDown function
func (m *MockContainerRuntime) WithComposeDown(fn func([]string, []string) error) *MockContainerRuntime {
	m.composeDownFunc = fn
	return m
}

// WithExecInContainer sets a custom ExecInContainer function
func (m *MockContainerRuntime) WithExecInContainer(fn func(string, string, bool) error) *MockContainerRuntime {
	m.execInContainerFunc = fn
	return m
}

// WithGetPortMappings sets a custom GetPortMappings function
func (m *MockContainerRuntime) WithGetPortMappings(fn func(string) (map[string]string, error)) *MockContainerRuntime {
	m.getPortMappingsFunc = fn
	return m
}

// WithGetDependencies sets a custom GetDependencies function
func (m *MockContainerRuntime) WithGetDependencies(fn func(string, []string) ([]string, error)) *MockContainerRuntime {
	m.getDependenciesFunc = fn
	return m
}

// WithGetContainerName sets a custom GetContainerName function
func (m *MockContainerRuntime) WithGetContainerName(fn func(string, []string) (string, error)) *MockContainerRuntime {
	m.getContainerNameFunc = fn
	return m
}

// WithGetAllDependenciesRecursive sets a custom GetAllDependenciesRecursive function
func (m *MockContainerRuntime) WithGetAllDependenciesRecursive(fn func(string, []string) ([]string, error)) *MockContainerRuntime {
	m.getAllDependenciesRecursiveFunc = fn
	return m
}

// WithGetContainerLogs sets a custom GetContainerLogs function
func (m *MockContainerRuntime) WithGetContainerLogs(fn func(string, int) ([]string, error)) *MockContainerRuntime {
	m.getContainerLogsFunc = fn
	return m
}

// WithStreamContainerLogs sets a custom StreamContainerLogs function
func (m *MockContainerRuntime) WithStreamContainerLogs(fn func(string, chan<- string, <-chan struct{}) error) *MockContainerRuntime {
	m.streamContainerLogsFunc = fn
	return m
}

// WithCheckImageExists sets a custom CheckImageExists function
func (m *MockContainerRuntime) WithCheckImageExists(fn func(string) (bool, error)) *MockContainerRuntime {
	m.checkImageExistsFunc = fn
	return m
}

// WithGetImageInfo sets a custom GetImageInfo function
func (m *MockContainerRuntime) WithGetImageInfo(fn func(string, []string) (string, error)) *MockContainerRuntime {
	m.getImageInfoFunc = fn
	return m
}

// WithPullImageWithProgress sets a custom PullImageWithProgress function
func (m *MockContainerRuntime) WithPullImageWithProgress(fn func(string, chan<- container.ImagePullProgress, <-chan struct{}) error) *MockContainerRuntime {
	m.pullImageWithProgressFunc = fn
	return m
}

// WithGetContainerStatus sets a custom GetContainerStatus function
func (m *MockContainerRuntime) WithGetContainerStatus(fn func(string) (string, error)) *MockContainerRuntime {
	m.getContainerStatusFunc = fn
	return m
}

// WithDefaults sets default return values for common operations
func (m *MockContainerRuntime) WithDefaults(portMappings map[string]string, dependencies []string, logs []string, containerStatus string, imageExists bool) *MockContainerRuntime {
	if portMappings != nil {
		m.defaultPortMappings = portMappings
	}
	if dependencies != nil {
		m.defaultDependencies = dependencies
	}
	if logs != nil {
		m.defaultLogs = logs
	}
	if containerStatus != "" {
		m.defaultContainerStatus = containerStatus
	}
	m.defaultImageExists = imageExists
	return m
}

// Runtime interface implementation
func (m *MockContainerRuntime) Name() string {
	if m.name != "" {
		return m.name
	}
	return "mock"
}

func (m *MockContainerRuntime) CheckAvailable() error {
	if m.checkAvailableFunc != nil {
		return m.checkAvailableFunc()
	}
	return nil
}

func (m *MockContainerRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	if m.composeUpFunc != nil {
		return m.composeUpFunc(composeFiles, services, quiet)
	}
	return nil
}

func (m *MockContainerRuntime) ComposeDown(composeFiles []string, services []string) error {
	if m.composeDownFunc != nil {
		return m.composeDownFunc(composeFiles, services)
	}
	return nil
}

func (m *MockContainerRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	if m.execInContainerFunc != nil {
		return m.execInContainerFunc(containerName, cmd, interactive)
	}
	return nil
}

func (m *MockContainerRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	if m.getPortMappingsFunc != nil {
		return m.getPortMappingsFunc(containerName)
	}
	return m.defaultPortMappings, nil
}

func (m *MockContainerRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	if m.getDependenciesFunc != nil {
		return m.getDependenciesFunc(service, composeFiles)
	}
	return m.defaultDependencies, nil
}

func (m *MockContainerRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
	if m.getContainerNameFunc != nil {
		return m.getContainerNameFunc(serviceName, composeFiles)
	}
	return fmt.Sprintf("test_%s_1", serviceName), nil
}

func (m *MockContainerRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string) ([]string, error) {
	if m.getAllDependenciesRecursiveFunc != nil {
		return m.getAllDependenciesRecursiveFunc(serviceName, composeFiles)
	}
	return m.defaultDependencies, nil
}

func (m *MockContainerRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	if m.getContainerLogsFunc != nil {
		return m.getContainerLogsFunc(containerName, tailLines)
	}
	return m.defaultLogs, nil
}

func (m *MockContainerRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	if m.streamContainerLogsFunc != nil {
		return m.streamContainerLogsFunc(containerName, logChan, stopChan)
	}
	go func() {
		defer close(logChan)
		for _, log := range m.defaultLogs {
			select {
			case logChan <- log:
			case <-stopChan:
				return
			}
		}
	}()
	return nil
}

func (m *MockContainerRuntime) CheckImageExists(imageName string) (bool, error) {
	if m.checkImageExistsFunc != nil {
		return m.checkImageExistsFunc(imageName)
	}
	return m.defaultImageExists, nil
}

func (m *MockContainerRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	if m.getImageInfoFunc != nil {
		return m.getImageInfoFunc(serviceName, composeFiles)
	}
	return fmt.Sprintf("%s:latest", serviceName), nil
}

func (m *MockContainerRuntime) PullImageWithProgress(imageName string, progressChan chan<- container.ImagePullProgress, stopChan <-chan struct{}) error {
	if m.pullImageWithProgressFunc != nil {
		return m.pullImageWithProgressFunc(imageName, progressChan, stopChan)
	}
	// Default implementation
	go func() {
		defer close(progressChan)
		progress := []container.ImagePullProgress{
			{Status: "downloading", Progress: 50.0},
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

func (m *MockContainerRuntime) GetContainerStatus(containerName string) (string, error) {
	if m.getContainerStatusFunc != nil {
		return m.getContainerStatusFunc(containerName)
	}
	return m.defaultContainerStatus, nil
}

// Test utilities

// CreateTempDir creates a temporary directory for testing and returns cleanup function
func CreateTempDir() (string, func()) {
	tempDir, err := os.MkdirTemp("", "instaui-test-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp dir: %v", err))
	}
	return tempDir, func() { os.RemoveAll(tempDir) }
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
