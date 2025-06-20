package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// LogsHandler handles service logging operations with simplified logic
type LogsHandler struct {
	serviceManager  *models.ServiceManager
	runtime         container.Runtime
	logger          Logger
	ctx             context.Context
	logStreams      map[string]chan struct{}
	logStreamsMutex sync.RWMutex
}

// NewLogsHandler creates a new simplified logs handler
func NewLogsHandler(runtime container.Runtime, instaDir string, ctx context.Context, logger Logger) *LogsHandler {
	// Create runtime info adapter for service manager
	runtimeInfo := NewRuntimeInfoAdapter(runtime)

	// Create service manager to get enhanced service information
	serviceManager := models.NewServiceManager(instaDir, runtimeInfo, logger)

	// Load services to get container information
	if err := serviceManager.LoadServices(); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to load services for logs handler: %v", err))
	}

	return &LogsHandler{
		serviceManager: serviceManager,
		runtime:        runtime,
		logger:         logger,
		ctx:            ctx,
		logStreams:     make(map[string]chan struct{}),
	}
}

// GetServiceLogs returns recent logs from all containers belonging to a service
func (h *LogsHandler) GetServiceLogs(serviceName string, tailLines int) ([]string, error) {
	h.logger.Log(fmt.Sprintf("Getting logs for service: %s", serviceName))

	// Get enhanced service information
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	var allLogs []string

	// Get logs from all containers belonging to this service
	if len(service.AllContainers) == 0 {
		// Fallback to service name if no containers are defined
		logs, err := h.runtime.GetContainerLogs(serviceName, tailLines)
		if err != nil {
			return nil, fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
		}
		return logs, nil
	}

	// Collect logs from all containers
	for _, containerName := range service.AllContainers {
		logs, err := h.runtime.GetContainerLogs(containerName, tailLines)
		if err != nil {
			// Log the error but continue with other containers
			h.logger.Log(fmt.Sprintf("Warning: Failed to get logs from container %s: %v", containerName, err))
			continue
		}

		// Add container prefix to distinguish logs from different containers
		if len(service.AllContainers) > 1 {
			for _, log := range logs {
				allLogs = append(allLogs, fmt.Sprintf("[%s] %s", containerName, log))
			}
		} else {
			allLogs = append(allLogs, logs...)
		}
	}

	h.logger.Log(fmt.Sprintf("Retrieved %d log lines for service: %s", len(allLogs), serviceName))
	return allLogs, nil
}

// StartLogStream starts streaming logs for all containers of a service
func (h *LogsHandler) StartLogStream(serviceName string) error {
	h.logger.Log(fmt.Sprintf("Starting log stream for service: %s", serviceName))

	// Check if stream is already active
	h.logStreamsMutex.Lock()
	if _, exists := h.logStreams[serviceName]; exists {
		h.logStreamsMutex.Unlock()
		return fmt.Errorf("log stream already active for service %s", serviceName)
	}

	// Create stop channel for this service
	stopChan := make(chan struct{})
	h.logStreams[serviceName] = stopChan
	h.logStreamsMutex.Unlock()

	// Get enhanced service information
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		// Clean up stop channel
		h.logStreamsMutex.Lock()
		delete(h.logStreams, serviceName)
		h.logStreamsMutex.Unlock()
		return fmt.Errorf("service %s not found", serviceName)
	}

	// Check if service is running
	if service.Status != "running" {
		// Service is not running, clean up and return
		h.logStreamsMutex.Lock()
		delete(h.logStreams, serviceName)
		h.logStreamsMutex.Unlock()
		h.logger.Log(fmt.Sprintf("Service %s is not running, no live log stream available", serviceName))
		return nil
	}

	// Get containers to stream from
	containersToStream := service.AllContainers
	if len(containersToStream) == 0 {
		// Fallback to service name
		containersToStream = []string{serviceName}
	}

	// Start streaming logs from all containers
	for _, containerName := range containersToStream {
		go h.streamContainerLogs(serviceName, containerName, stopChan)
	}

	h.logger.Log(fmt.Sprintf("Started log stream for service %s with %d containers", serviceName, len(containersToStream)))
	return nil
}

// streamContainerLogs streams logs from a specific container
func (h *LogsHandler) streamContainerLogs(serviceName, containerName string, stopChan <-chan struct{}) {
	h.logger.Log(fmt.Sprintf("Starting log stream for container: %s", containerName))

	logChan := make(chan string, 100)

	// Start the actual log streaming
	go func() {
		defer close(logChan)
		if err := h.runtime.StreamContainerLogs(containerName, logChan, stopChan); err != nil {
			h.logger.Log(fmt.Sprintf("Error streaming logs for container %s: %v", containerName, err))
		}
	}()

	// Forward logs (in a real implementation, this would go to WebSocket broadcaster)
	go func() {
		for log := range logChan {
			// Log processing is now handled by WebSocket broadcaster
			// This is a simplified implementation that just logs locally
			h.logger.Log(fmt.Sprintf("[%s][%s] %s", serviceName, containerName, log))
		}
	}()
}

// StopLogStream stops log streaming for a service
func (h *LogsHandler) StopLogStream(serviceName string) error {
	h.logger.Log(fmt.Sprintf("Stopping log stream for service: %s", serviceName))

	h.logStreamsMutex.Lock()
	defer h.logStreamsMutex.Unlock()

	stopChan, exists := h.logStreams[serviceName]
	if !exists {
		return fmt.Errorf("no active log stream found for service %s", serviceName)
	}

	// Signal all streaming goroutines to stop
	close(stopChan)
	// Remove from active streams
	delete(h.logStreams, serviceName)

	h.logger.Log(fmt.Sprintf("Stopped log stream for service: %s", serviceName))
	return nil
}

// GetActiveLogStreams returns a list of services with active log streams
func (h *LogsHandler) GetActiveLogStreams() []string {
	h.logStreamsMutex.RLock()
	defer h.logStreamsMutex.RUnlock()

	streams := make([]string, 0, len(h.logStreams))
	for serviceName := range h.logStreams {
		streams = append(streams, serviceName)
	}

	return streams
}

// RefreshServiceInfo refreshes service information from compose files
func (h *LogsHandler) RefreshServiceInfo() error {
	h.logger.Log("Refreshing service information for logs handler")

	if err := h.serviceManager.LoadServices(); err != nil {
		h.logger.Log(fmt.Sprintf("Failed to refresh service info: %v", err))
		return err
	}

	h.logger.Log("Successfully refreshed service information")
	return nil
}
