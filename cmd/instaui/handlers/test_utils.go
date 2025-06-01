package handlers

import (
	"fmt"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// mockContainerRuntime provides a comprehensive mock implementation for testing.
// This consolidates the duplicate mockContainerRuntime implementations across handler test files.
type mockContainerRuntime struct {
	nameFunc                        func() string
	checkAvailableFunc              func() error
	composeUpFunc                   func([]string, []string, bool) error
	composeDownFunc                 func([]string, []string) error
	execInContainerFunc             func(string, string, bool) error
	getPortMappingsFunc             func(string) (map[string]string, error)
	getContainerNameFunc            func(string, []string) (string, error)
	getAllDependenciesRecursiveFunc func(string, []string, bool) ([]string, error)
	getContainerLogsFunc            func(string, int) ([]string, error)
	streamContainerLogsFunc         func(string, chan<- string, <-chan struct{}) error
	checkImageExistsFunc            func(string) (bool, error)
	getImageInfoFunc                func(string, []string) (string, error)
	pullImageWithProgressFunc       func(string, chan<- container.ImagePullProgress, <-chan struct{}) error
	getContainerStatusFunc          func(string) (string, error)
	getAllContainerStatusesFunc     func() (map[string]string, error)

	// Default return values
	defaultName            string
	defaultPortMappings    map[string]string
	defaultDependencies    []string
	defaultLogs            []string
	defaultContainerStatus string
	defaultImageExists     bool
}

// newMockContainerRuntime creates a new mock container runtime with sensible defaults
func newMockContainerRuntime() *mockContainerRuntime {
	return &mockContainerRuntime{
		defaultName:            "mock",
		defaultPortMappings:    map[string]string{},
		defaultDependencies:    []string{},
		defaultLogs:            []string{"test log line"},
		defaultContainerStatus: "running",
		defaultImageExists:     true,
	}
}

// withComposeUp sets a custom ComposeUp function
func (m *mockContainerRuntime) withComposeUp(fn func([]string, []string, bool) error) *mockContainerRuntime {
	m.composeUpFunc = fn
	return m
}

// withComposeDown sets a custom ComposeDown function
func (m *mockContainerRuntime) withComposeDown(fn func([]string, []string) error) *mockContainerRuntime {
	m.composeDownFunc = fn
	return m
}

// withGetPortMappings sets a custom GetPortMappings function
func (m *mockContainerRuntime) withGetPortMappings(fn func(string) (map[string]string, error)) *mockContainerRuntime {
	m.getPortMappingsFunc = fn
	return m
}

// withGetContainerName sets a custom GetContainerName function
func (m *mockContainerRuntime) withGetContainerName(fn func(string, []string) (string, error)) *mockContainerRuntime {
	m.getContainerNameFunc = fn
	return m
}

// withGetAllDependenciesRecursive sets a custom GetAllDependenciesRecursive function
func (m *mockContainerRuntime) withGetAllDependenciesRecursive(fn func(string, []string, bool) ([]string, error)) *mockContainerRuntime {
	m.getAllDependenciesRecursiveFunc = fn
	return m
}

// withGetContainerLogs sets a custom GetContainerLogs function
func (m *mockContainerRuntime) withGetContainerLogs(fn func(string, int) ([]string, error)) *mockContainerRuntime {
	m.getContainerLogsFunc = fn
	return m
}

// withStreamContainerLogs sets a custom StreamContainerLogs function
func (m *mockContainerRuntime) withStreamContainerLogs(fn func(string, chan<- string, <-chan struct{}) error) *mockContainerRuntime {
	m.streamContainerLogsFunc = fn
	return m
}

// withGetAllContainerStatuses sets a custom GetAllContainerStatuses function
func (m *mockContainerRuntime) withGetAllContainerStatuses(fn func() (map[string]string, error)) *mockContainerRuntime {
	m.getAllContainerStatusesFunc = fn
	return m
}

// withGetContainerStatus sets a custom GetContainerStatus function
func (m *mockContainerRuntime) withGetContainerStatus(fn func(string) (string, error)) *mockContainerRuntime {
	m.getContainerStatusFunc = fn
	return m
}

// Runtime interface implementation
func (m *mockContainerRuntime) Name() string {
	if m.nameFunc != nil {
		return m.nameFunc()
	}
	return m.defaultName
}

func (m *mockContainerRuntime) CheckAvailable() error {
	if m.checkAvailableFunc != nil {
		return m.checkAvailableFunc()
	}
	return nil
}

func (m *mockContainerRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	if m.composeUpFunc != nil {
		return m.composeUpFunc(composeFiles, services, quiet)
	}
	return nil
}

func (m *mockContainerRuntime) ComposeDown(composeFiles []string, services []string) error {
	if m.composeDownFunc != nil {
		return m.composeDownFunc(composeFiles, services)
	}
	return nil
}

func (m *mockContainerRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	if m.execInContainerFunc != nil {
		return m.execInContainerFunc(containerName, cmd, interactive)
	}
	return nil
}

func (m *mockContainerRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	if m.getPortMappingsFunc != nil {
		return m.getPortMappingsFunc(containerName)
	}
	return m.defaultPortMappings, nil
}

func (m *mockContainerRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string, isContainer bool) ([]string, error) {
	if m.getAllDependenciesRecursiveFunc != nil {
		return m.getAllDependenciesRecursiveFunc(serviceName, composeFiles, isContainer)
	}
	return m.defaultDependencies, nil
}

func (m *mockContainerRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
	if m.getContainerNameFunc != nil {
		return m.getContainerNameFunc(serviceName, composeFiles)
	}
	return fmt.Sprintf("test_%s_1", serviceName), nil
}

func (m *mockContainerRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	if m.getContainerLogsFunc != nil {
		return m.getContainerLogsFunc(containerName, tailLines)
	}
	return m.defaultLogs, nil
}

func (m *mockContainerRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
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

func (m *mockContainerRuntime) CheckImageExists(imageName string) (bool, error) {
	if m.checkImageExistsFunc != nil {
		return m.checkImageExistsFunc(imageName)
	}
	return m.defaultImageExists, nil
}

// CheckMultipleImagesExist checks if multiple images exist locally in a single call
func (m *mockContainerRuntime) CheckMultipleImagesExist(imageNames []string) (map[string]bool, error) {
	result := make(map[string]bool)
	for _, imageName := range imageNames {
		if m.checkImageExistsFunc != nil {
			exists, err := m.checkImageExistsFunc(imageName)
			if err != nil {
				return nil, err
			}
			result[imageName] = exists
		} else {
			result[imageName] = m.defaultImageExists
		}
	}
	return result, nil
}

func (m *mockContainerRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	if m.getImageInfoFunc != nil {
		return m.getImageInfoFunc(serviceName, composeFiles)
	}
	return fmt.Sprintf("test/%s:latest", serviceName), nil
}

// GetMultipleImageInfo returns image information for multiple services from compose files
func (m *mockContainerRuntime) GetMultipleImageInfo(serviceNames []string, composeFiles []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, serviceName := range serviceNames {
		if m.getImageInfoFunc != nil {
			imageInfo, err := m.getImageInfoFunc(serviceName, composeFiles)
			if err != nil {
				// Skip services with errors rather than failing the whole operation
				continue
			}
			result[serviceName] = imageInfo
		} else {
			result[serviceName] = fmt.Sprintf("test/%s:latest", serviceName)
		}
	}
	return result, nil
}

func (m *mockContainerRuntime) PullImageWithProgress(imageName string, progressChan chan<- container.ImagePullProgress, stopChan <-chan struct{}) error {
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

func (m *mockContainerRuntime) GetContainerStatus(containerName string) (string, error) {
	if m.getContainerStatusFunc != nil {
		return m.getContainerStatusFunc(containerName)
	}
	return m.defaultContainerStatus, nil
}

// GetAllContainerStatuses returns all running containers managed by compose
func (m *mockContainerRuntime) GetAllContainerStatuses() (map[string]string, error) {
	if m.getAllContainerStatusesFunc != nil {
		return m.getAllContainerStatusesFunc()
	}
	// Return some mock container names for testing that match the expected patterns
	if m.defaultContainerStatus == "running" {
		// Return container names that match the patterns used in the optimized handlers
		return map[string]string{"postgres": "running", "redis": "running", "insta_postgres_1": "running", "insta_redis_1": "running"}, nil
	}
	return map[string]string{}, nil
}

// Test utilities

// Helper function for string contains check (used across tests)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && containsHelper(s[1:], substr)
}

func containsHelper(s, substr string) bool {
	if len(s) < len(substr) {
		return false
	}
	if s[:len(substr)] == substr {
		return true
	}
	return containsHelper(s[1:], substr)
}
