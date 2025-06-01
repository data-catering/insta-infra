package handlers

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// BaseHandler contains common functionality shared across all handlers
type BaseHandler struct {
	containerRuntime container.Runtime
	instaDir         string

	// Container status caching
	runningContainersCache      map[string]string
	runningContainersCacheTime  time.Time
	runningContainersCacheMutex sync.RWMutex
	containerCacheTTL           time.Duration
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(runtime container.Runtime, instaDir string) *BaseHandler {
	return &BaseHandler{
		containerRuntime:  runtime,
		instaDir:          instaDir,
		containerCacheTTL: 3 * time.Second, // Cache container status for 3 seconds
	}
}

// getComposeFiles returns the list of compose files to use
func (h *BaseHandler) getComposeFiles() []string {
	baseComposeFile := filepath.Join(h.instaDir, "docker-compose.yaml")
	persistComposeFile := filepath.Join(h.instaDir, "docker-compose-persist.yaml")

	composeFiles := []string{baseComposeFile}
	if fileExists(persistComposeFile) {
		composeFiles = append(composeFiles, persistComposeFile)
	}

	return composeFiles
}

// getRunningContainers gets all running containers with caching to improve performance
func (h *BaseHandler) getCurrentContainers() (map[string]string, error) {
	h.runningContainersCacheMutex.RLock()

	// Check if cache is still valid
	if h.runningContainersCache != nil && time.Since(h.runningContainersCacheTime) < h.containerCacheTTL {
		defer h.runningContainersCacheMutex.RUnlock()
		// Return a copy of the cache to avoid concurrent map access
		result := make(map[string]string)
		for k, v := range h.runningContainersCache {
			result[k] = v
		}
		return result, nil
	}

	h.runningContainersCacheMutex.RUnlock()

	// Cache is stale or missing, fetch new data
	h.runningContainersCacheMutex.Lock()
	defer h.runningContainersCacheMutex.Unlock()

	// Double-check in case another goroutine updated while we waited
	if h.runningContainersCache != nil && time.Since(h.runningContainersCacheTime) < h.containerCacheTTL {
		result := make(map[string]string)
		for k, v := range h.runningContainersCache {
			result[k] = v
		}
		return result, nil
	}

	// Fetch fresh data
	currentContainers, err := h.fetchCurrentContainers()
	if err != nil {
		return nil, err
	}

	// Update cache
	h.runningContainersCache = currentContainers
	h.runningContainersCacheTime = time.Now()

	// Return a copy
	result := make(map[string]string)
	for k, v := range currentContainers {
		result[k] = v
	}
	return result, nil
}

// fetchCurrentContainers does the actual work of getting current containers and their statuses
func (h *BaseHandler) fetchCurrentContainers() (map[string]string, error) {
	currentContainers := make(map[string]string)

	// Use direct docker/podman ps to get all current containers at once
	// This is much faster than calling GetContainerName + GetPortMappings for each service
	containers, err := h.containerRuntime.GetAllContainerStatuses()
	if err != nil {
		return nil, err
	}

	// Convert to map for fast lookup
	for containerName, status := range containers {
		currentContainers[containerName] = status
	}

	return currentContainers, nil
}

// isServiceRunning checks if a service is running using the container runtime's methods
func (h *BaseHandler) isServiceRunning(serviceName string, composeFiles []string, currentContainers map[string]string) bool {
	// Use the container runtime's GetContainerName method to get the proper container name
	containerName, err := h.containerRuntime.GetContainerName(serviceName, composeFiles)
	if err != nil {
		return false
	}

	// Check if this container is in the running containers list
	return currentContainers[containerName] == "running"
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Runtime returns the container runtime
func (h *BaseHandler) Runtime() container.Runtime {
	return h.containerRuntime
}

// InstaDir returns the insta directory
func (h *BaseHandler) InstaDir() string {
	return h.instaDir
}

// invalidateContainerCache clears the running containers cache
// Call this when containers might have changed (start/stop operations)
func (h *BaseHandler) invalidateContainerCache() {
	h.runningContainersCacheMutex.Lock()
	defer h.runningContainersCacheMutex.Unlock()
	h.runningContainersCache = nil
	h.runningContainersCacheTime = time.Time{}
}
