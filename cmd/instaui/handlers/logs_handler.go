package handlers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// LogsHandler handles service logging and streaming operations
type LogsHandler struct {
	*BaseHandler
	ctx             context.Context
	logStreams      map[string]chan struct{}
	logStreamsMutex sync.RWMutex
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(runtime container.Runtime, instaDir string, ctx context.Context) *LogsHandler {
	return &LogsHandler{
		BaseHandler: NewBaseHandler(runtime, instaDir),
		ctx:         ctx,
		logStreams:  make(map[string]chan struct{}),
	}
}

// GetServiceLogs returns recent logs from a service container (running or exited)
func (h *LogsHandler) GetServiceLogs(serviceName string, tailLines int) ([]string, error) {
	// Try to get the proper container name for the service
	containerName, err := h.Runtime().GetContainerName(serviceName, h.getComposeFiles())
	if err != nil {
		// If GetContainerName fails, fall back to using service name directly
		// This handles cases where the container exists but compose files might not be found
		containerName = serviceName
	}

	// Try to get logs from the container (works for both running and stopped containers)
	logs, err := h.Runtime().GetContainerLogs(containerName, tailLines)
	if err != nil {
		// If using the resolved container name fails, try with the original service name
		// This provides a fallback in case the container naming doesn't match expectations
		if containerName != serviceName {
			logs, err = h.Runtime().GetContainerLogs(serviceName, tailLines)
			if err != nil {
				return nil, fmt.Errorf("failed to get logs for service %s (tried both '%s' and '%s'): %w", serviceName, containerName, serviceName, err)
			}
		} else {
			return nil, fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
		}
	}
	return logs, nil
}

// StartLogStream starts streaming logs for a service using Wails events
func (h *LogsHandler) StartLogStream(serviceName string) error {
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

	// Try to get the proper container name for the service
	// This will work for both running and stopped containers
	containerName, err := h.Runtime().GetContainerName(serviceName, h.getComposeFiles())
	if err != nil {
		// If GetContainerName fails, fall back to using service name directly
		// This handles cases where the container exists but compose files might not be found
		containerName = serviceName
	}

	// Get all running containers to check if this service is currently running
	currentContainers, err := h.getCurrentContainers()
	if err != nil {
		// If we can't get current containers, we'll still try to get logs from the container
		// but we won't be able to stream live logs
		currentContainers = make(map[string]string)
	}

	// Check if the service is currently running
	var isRunning bool
	if status, exists := currentContainers[serviceName]; exists {
		isRunning = (status == "running" || status == "restarting")
	}

	// If not running, try to get historical logs from the container (stopped container)
	if !isRunning {
		// Try to get logs from the container (works for both running and stopped containers)
		logs, err := h.Runtime().GetContainerLogs(containerName, 1000)

		// Clean up the stop channel since we won't be streaming
		h.logStreamsMutex.Lock()
		delete(h.logStreams, serviceName)
		h.logStreamsMutex.Unlock()

		if err != nil {
			// If we can't get logs, it might mean the container doesn't exist at all
			// This is not necessarily an error - just means no logs are available
			if h.ctx != nil {
				runtime.EventsEmit(h.ctx, "service-log", map[string]interface{}{
					"serviceName": serviceName,
					"log":         []string{fmt.Sprintf("No logs available for service '%s' (container may not exist or have no logs)", serviceName)},
					"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
				})
			}
			return nil // Don't return error, just no logs available
		}

		// Emit the historical logs
		if h.ctx != nil {
			runtime.EventsEmit(h.ctx, "service-log", map[string]interface{}{
				"serviceName": serviceName,
				"log":         logs,
				"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
			})
		}
		return nil
	}

	// Service is running, set up live log streaming
	// Create channels for log streaming
	logChan := make(chan string, 100) // Buffer for logs

	// Start log streaming in a goroutine
	go func() {
		defer func() {
			close(logChan)
			// Clean up the stop channel when streaming ends
			h.logStreamsMutex.Lock()
			delete(h.logStreams, serviceName)
			h.logStreamsMutex.Unlock()

			// Emit stop event to frontend
			if h.ctx != nil {
				runtime.EventsEmit(h.ctx, "service-logs-stopped", map[string]interface{}{
					"serviceName": serviceName,
				})
			}
		}()

		if err := h.Runtime().StreamContainerLogs(containerName, logChan, stopChan); err != nil {
			// Send error event to frontend
			if h.ctx != nil {
				runtime.EventsEmit(h.ctx, "service-logs-error", map[string]interface{}{
					"serviceName": serviceName,
					"error":       err.Error(),
				})
			}
		}
	}()

	// Forward logs from channel to Wails events
	go func() {
		for logLine := range logChan {
			if logLine != "" {
				// Emit log event to frontend
				if h.ctx != nil {
					runtime.EventsEmit(h.ctx, "service-log", map[string]interface{}{
						"serviceName": serviceName,
						"log":         logLine,
						"timestamp":   fmt.Sprintf("%d", time.Now().Unix()),
					})
				}
			}
		}
	}()

	return nil
}

// getComposeFiles returns the list of compose files to use
func (h *LogsHandler) getComposeFiles() []string {
	return []string{
		fmt.Sprintf("%s/docker-compose.yaml", h.InstaDir()),
		fmt.Sprintf("%s/docker-compose-persist.yaml", h.InstaDir()),
	}
}

// StopLogStream stops log streaming for a service
func (h *LogsHandler) StopLogStream(serviceName string) error {
	h.logStreamsMutex.Lock()
	defer h.logStreamsMutex.Unlock()

	stopChan, exists := h.logStreams[serviceName]
	if !exists {
		return fmt.Errorf("no active log stream found for service %s", serviceName)
	}

	// Signal the streaming goroutine to stop
	close(stopChan)
	// Remove from active streams (will also be cleaned up by the goroutine)
	delete(h.logStreams, serviceName)

	return nil
}
