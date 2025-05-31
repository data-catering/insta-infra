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
	// Prepare compose files
	composeFiles := h.getComposeFiles()

	// First try to get container name (works for both running and exited containers)
	containerName, err := h.Runtime().GetContainerName(serviceName, composeFiles)
	if err != nil {
		return nil, fmt.Errorf("could not get container name for service %s: %w", serviceName, err)
	}

	// Get logs from container (works for both running and exited containers)
	logs, err := h.Runtime().GetContainerLogs(containerName, tailLines)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for service %s: %w", serviceName, err)
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

	// Prepare compose files
	composeFiles := h.getComposeFiles()

	// Get all running containers first
	runningContainers, err := h.getRunningContainers()
	if err != nil {
		// Clean up the stop channel if we can't get running containers
		h.logStreamsMutex.Lock()
		delete(h.logStreams, serviceName)
		h.logStreamsMutex.Unlock()
		return fmt.Errorf("could not get running containers: %w", err)
	}

	// Try to get container name first (works for both running and exited containers)
	containerName, err := h.Runtime().GetContainerName(serviceName, composeFiles)
	if err != nil {
		// Clean up the stop channel if we can't get container name
		h.logStreamsMutex.Lock()
		delete(h.logStreams, serviceName)
		h.logStreamsMutex.Unlock()
		return fmt.Errorf("could not get container name for service %s: %w", serviceName, err)
	}

	// Check if container is running for streaming (only running containers can be streamed)
	isRunning := h.isServiceRunning(serviceName, composeFiles, runningContainers)
	if !isRunning {
		// Check if container exists but is not running - we can't stream but should provide helpful error
		containerStatus, err := h.Runtime().GetContainerStatus(containerName)
		if err == nil && containerStatus != "not_found" {
			// Clean up the stop channel if service is not running
			h.logStreamsMutex.Lock()
			delete(h.logStreams, serviceName)
			h.logStreamsMutex.Unlock()
			return fmt.Errorf("service %s is not running (status: %s) - cannot stream logs from stopped containers", serviceName, containerStatus)
		}

		// Clean up the stop channel if service is not running
		h.logStreamsMutex.Lock()
		delete(h.logStreams, serviceName)
		h.logStreamsMutex.Unlock()
		return fmt.Errorf("service %s is not running", serviceName)
	}

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

