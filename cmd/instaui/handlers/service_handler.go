package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// ServiceHandler handles all service-related operations
type ServiceHandler struct {
	*BaseHandler

	// Status tracking for immediate updates
	stoppedServices     map[string]bool
	startingServices    map[string]bool
	statusTrackingMutex sync.RWMutex
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(runtime container.Runtime, instaDir string) *ServiceHandler {
	return &ServiceHandler{
		BaseHandler:      NewBaseHandler(runtime, instaDir),
		stoppedServices:  make(map[string]bool),
		startingServices: make(map[string]bool),
	}
}

// ListServices returns a list of all available services from core.Services
// All services in core.Services should be shown to users regardless of their container names
func (h *ServiceHandler) ListServices() []models.ServiceInfo {
	var serviceList []models.ServiceInfo

	// Show all services from core.Services - this is the canonical list
	for _, service := range core.Services {
		serviceList = append(serviceList, models.ServiceInfo(service))
	}

	// Sort services by name for consistent ordering in the UI
	sort.Slice(serviceList, func(i, j int) bool {
		return serviceList[i].Name < serviceList[j].Name
	})

	return serviceList
}

// GetAllServiceDetails fetches status and dependencies for all services concurrently
func (h *ServiceHandler) GetAllServiceDetails() ([]models.ServiceDetailInfo, error) {
	baseServices := h.ListServices()
	detailsList := make([]models.ServiceDetailInfo, len(baseServices))

	var wg sync.WaitGroup
	resultsChan := make(chan models.ServiceDetailInfo, len(baseServices))

	composeFiles := h.getComposeFiles()

	for _, sInfo := range baseServices {
		wg.Add(1)
		go func(service models.ServiceInfo) {
			defer wg.Done()
			var detail models.ServiceDetailInfo
			detail.Name = service.Name
			detail.Type = service.Type

			// Fetch Status
			status, statusErr := h.getServiceStatusInternal(service.Name, composeFiles)
			detail.Status = status
			if statusErr != nil {
				detail.StatusError = statusErr.Error()
			}

			// Fetch Dependencies
			deps, depsErr := h.getServiceDependenciesInternal(service.Name, composeFiles)
			detail.Dependencies = deps
			if depsErr != nil {
				detail.DependenciesError = depsErr.Error()
			}
			resultsChan <- detail
		}(sInfo)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	i := 0
	for detail := range resultsChan {
		detailsList[i] = detail
		i++
	}

	// Sort the final list by name for consistent UI display
	sort.Slice(detailsList, func(k, j int) bool {
		return detailsList[k].Name < detailsList[j].Name
	})

	return detailsList, nil
}

// GetServiceStatus returns the status of a given service
func (h *ServiceHandler) GetServiceStatus(serviceName string) (string, error) {
	composeFiles := h.getComposeFiles()
	return h.getServiceStatusInternal(serviceName, composeFiles)
}

// GetServiceDependencies returns a list of dependencies for a given service
func (h *ServiceHandler) GetServiceDependencies(serviceName string) ([]string, error) {
	composeFiles := h.getComposeFiles()
	return h.getServiceDependenciesInternal(serviceName, composeFiles)
}

// StartService starts a specific service with optional persistence
func (h *ServiceHandler) StartService(serviceName string, persist bool) error {
	services := []string{serviceName}

	if persist {
		// Create data directory structure
		dataDir := filepath.Join(h.InstaDir(), "data")
		if err := os.MkdirAll(dataDir, 0755); err != nil {
			return fmt.Errorf("failed to create data directory: %w", err)
		}

		// Create persist directories for the service
		persistDir := filepath.Join(dataDir, serviceName, "persist")
		if err := os.MkdirAll(persistDir, 0755); err != nil {
			return fmt.Errorf("failed to create persist directory for %s: %w", serviceName, err)
		}
	}

	composeFiles := []string{filepath.Join(h.InstaDir(), "docker-compose.yaml")}
	if persist {
		composeFiles = append(composeFiles, filepath.Join(h.InstaDir(), "docker-compose-persist.yaml"))
	}

	// Invalidate container cache after starting services (container state might change)
	defer h.invalidateContainerCache()

	err := h.Runtime().ComposeUp(composeFiles, services, true)
	if err != nil {
		return err
	}

	// Clear the stopped status for this service since it's now starting
	h.statusTrackingMutex.Lock()
	delete(h.stoppedServices, serviceName)
	h.statusTrackingMutex.Unlock()

	return nil
}

// StartServiceWithStatusUpdate starts a service and returns updated status information
// This provides immediate status updates without requiring additional Docker/Podman calls
func (h *ServiceHandler) StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]models.ServiceStatus, error) {
	// Mark service as starting immediately
	h.markServiceStarting(serviceName)

	// Start the service
	err := h.StartService(serviceName, persist)
	if err != nil {
		// Clear starting status if start failed
		h.clearServiceStarting(serviceName)
		return nil, err
	}

	// Return optimized status update focusing on the started service
	return h.getOptimizedStatusUpdate(serviceName), nil
}

// getOptimizedStatusUpdate returns status for all services with optimized checking for the target service
func (h *ServiceHandler) getOptimizedStatusUpdate(targetService string) map[string]models.ServiceStatus {
	allServices := h.ListServices()
	statusMap := make(map[string]models.ServiceStatus)
	composeFiles := h.getComposeFiles()

	for _, service := range allServices {
		var status string
		var err error

		if service.Name == targetService {
			// For the target service, get fresh status (it should be starting)
			status, err = h.getServiceStatusInternal(service.Name, composeFiles)
		} else {
			// For other services, use efficient checking
			if h.isServiceMarkedStopped(service.Name) {
				status = "stopped"
			} else if h.isServiceMarkedStarting(service.Name) {
				status = "starting"
			} else {
				// Quick check if still running (only for non-target services)
				runningContainers, containerErr := h.getRunningContainers()
				if containerErr == nil && h.isServiceRunning(service.Name, composeFiles, runningContainers) {
					status = "running"
				} else {
					status = "stopped"
				}
			}
		}

		statusMap[service.Name] = models.ServiceStatus{
			ServiceName: service.Name,
			Status:      status,
		}

		if err != nil {
			statusMap[service.Name] = models.ServiceStatus{
				ServiceName: service.Name,
				Status:      "error",
				Error:       err.Error(),
			}
		}
	}

	return statusMap
}

// StopService stops a specific service
func (h *ServiceHandler) StopService(serviceName string) error {
	services := []string{serviceName}
	composeFiles := []string{
		filepath.Join(h.InstaDir(), "docker-compose.yaml"),
		filepath.Join(h.InstaDir(), "docker-compose-persist.yaml"),
	}

	// Invalidate container cache after stopping services (container state might change)
	defer h.invalidateContainerCache()

	err := h.Runtime().ComposeDown(composeFiles, services)
	if err != nil {
		return err
	}

	// Mark the service as stopped for immediate status updates
	h.markServiceStopped(serviceName)

	return nil
}

// StopAllServices stops all running services
func (h *ServiceHandler) StopAllServices() error {
	// Empty services array means stop all services
	services := []string{}
	composeFiles := []string{
		filepath.Join(h.InstaDir(), "docker-compose.yaml"),
		filepath.Join(h.InstaDir(), "docker-compose-persist.yaml"),
	}

	// Invalidate container cache after stopping all services
	defer h.invalidateContainerCache()

	err := h.Runtime().ComposeDown(composeFiles, services)
	if err != nil {
		return err
	}

	// Mark all services as stopped for immediate status updates
	h.markAllServicesStopped()

	return nil
}

// getServiceStatusInternal is an internal helper for getting service status
func (h *ServiceHandler) getServiceStatusInternal(serviceName string, composeFiles []string) (string, error) {
	// Check if service is marked as starting first (immediate update without Docker/Podman call)
	if h.isServiceMarkedStarting(serviceName) {
		return "starting", nil
	}

	// Check if service is marked as stopped first (immediate update without Docker/Podman call)
	if h.isServiceMarkedStopped(serviceName) {
		return "stopped", nil
	}

	// Get all running containers first
	runningContainers, err := h.getRunningContainers()
	if err != nil {
		return "stopped", fmt.Errorf("could not get running containers: %w", err)
	}

	// Check if the main service is running
	isMainServiceRunning := h.isServiceRunning(serviceName, composeFiles, runningContainers)

	// If main service is running, check if any critical dependencies have failed
	if isMainServiceRunning {
		// Clear starting status since service is now running
		h.clearServiceStarting(serviceName)

		// Check for dependency failures that should affect main service status
		if hasCriticalDependencyFailures, err := h.checkCriticalDependencyFailures(serviceName, composeFiles); err == nil && hasCriticalDependencyFailures {
			return "failed", nil
		}
		return "running", nil
	}

	// If main service is not running, check if it failed due to dependency failures
	if hasCriticalDependencyFailures, err := h.checkCriticalDependencyFailures(serviceName, composeFiles); err == nil && hasCriticalDependencyFailures {
		// Clear starting status since service failed
		h.clearServiceStarting(serviceName)
		return "failed", nil
	}

	// Check if the main service container exists but is not running (failed to start)
	containerName, err := h.Runtime().GetContainerName(serviceName, composeFiles)
	if err == nil {
		containerStatus, err := h.Runtime().GetContainerStatus(containerName)
		if err == nil {
			switch containerStatus {
			case "error", "exited":
				// Clear starting status since service failed
				h.clearServiceStarting(serviceName)
				return "failed", nil
			case "created":
				// Container was created but never started - likely dependency failure
				h.clearServiceStarting(serviceName)
				return "failed", nil
			}
		}
	}

	return "stopped", nil
}

// checkCriticalDependencyFailures checks if any critical dependencies have failed
func (h *ServiceHandler) checkCriticalDependencyFailures(serviceName string, composeFiles []string) (bool, error) {
	// Get all dependencies
	deps, err := h.Runtime().GetAllDependenciesRecursive(serviceName, composeFiles)
	if err != nil {
		return false, err
	}

	// Check each dependency for failure
	for _, depName := range deps {
		// Get container name for dependency
		containerName, err := h.Runtime().GetContainerName(depName, composeFiles)
		if err != nil {
			continue // Skip if we can't get container name
		}

		// Get container status
		containerStatus, err := h.Runtime().GetContainerStatus(containerName)
		if err != nil {
			continue // Skip if we can't get status
		}

		// Check for critical failure patterns
		if containerStatus == "error" || containerStatus == "dead" {
			// Container failed or is in dead state
			return true, nil
		}

		// For init containers, "completed" status is success, but "error" is failure
		if strings.Contains(depName, "-init") || strings.Contains(depName, "-data") {
			if containerStatus == "error" {
				// Init container failed (exited with non-zero code)
				return true, nil
			}
			// "completed" status means init container succeeded, so continue checking other deps
		} else {
			// For non-init containers, any exited or error state is a failure
			if containerStatus == "exited" || containerStatus == "error" {
				return true, nil
			}
		}
	}

	return false, nil
}

// getServiceDependenciesInternal is an internal helper for getting service dependencies
func (h *ServiceHandler) getServiceDependenciesInternal(serviceName string, composeFiles []string) ([]string, error) {
	deps, err := h.Runtime().GetAllDependenciesRecursive(serviceName, composeFiles)
	if err != nil {
		return nil, fmt.Errorf("could not get recursive dependencies for %s: %w", serviceName, err)
	}
	sort.Strings(deps) // Sort for consistent UI display
	return deps, nil
}

// GetAllServiceDependencies efficiently gets dependencies for all services using runtime's cached methods
// This leverages the runtime's built-in caching for compose config parsing
func (h *ServiceHandler) GetAllServiceDependencies() (map[string][]string, error) {
	// Get all services that we care about (from filtered list)
	allServices := h.ListServices()
	dependencyMap := make(map[string][]string)
	composeFiles := h.getComposeFiles()

	// Use the runtime's cached GetAllDependenciesRecursive method for each service
	// The runtime caches the compose config parsing, so this is efficient
	for _, service := range allServices {
		dependencies, err := h.Runtime().GetAllDependenciesRecursive(service.Name, composeFiles)
		if err != nil {
			// If we can't get dependencies for a service, just use empty list
			dependencies = []string{}
		}
		dependencyMap[service.Name] = dependencies
	}

	return dependencyMap, nil
}

// GetAllRunningServices uses docker/podman ps to quickly get all running services
// This is much faster than checking each service individually
func (h *ServiceHandler) GetAllRunningServices() (map[string]models.ServiceStatus, error) {
	// Get all available services first (filtered list)
	allServices := h.ListServices()

	// Create result map with all services as "stopped" by default
	statusMap := make(map[string]models.ServiceStatus)
	for _, service := range allServices {
		statusMap[service.Name] = models.ServiceStatus{
			ServiceName: service.Name,
			Status:      "stopped",
		}
	}

	composeFiles := h.getComposeFiles()

	// Match running containers to service names and detect failures
	for _, service := range allServices {
		// Check if service is marked as stopped first (immediate update)
		if h.isServiceMarkedStopped(service.Name) {
			// Keep as stopped (already set as default)
			continue
		}

		// Use the improved status detection logic that can detect failures
		status, err := h.getServiceStatusInternal(service.Name, composeFiles)
		if err != nil {
			statusMap[service.Name] = models.ServiceStatus{
				ServiceName: service.Name,
				Status:      "error",
				Error:       err.Error(),
			}
		} else {
			statusMap[service.Name] = models.ServiceStatus{
				ServiceName: service.Name,
				Status:      status,
			}
		}
	}

	return statusMap, nil
}

// GetAllServicesWithStatusAndDependencies efficiently gets services, statuses, and dependencies in a single call
// This combines the fast approaches for both statuses and dependencies
func (h *ServiceHandler) GetAllServicesWithStatusAndDependencies() ([]models.ServiceDetailInfo, error) {
	// Get basic services list (filtered)
	allServices := h.ListServices()

	// Get all statuses efficiently with single docker ps call
	statusMap, err := h.GetAllRunningServices()
	if err != nil {
		return nil, fmt.Errorf("failed to get service statuses: %w", err)
	}

	// Get all dependencies efficiently with single compose config read (cached)
	dependencyMap, err := h.GetAllServiceDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to get service dependencies: %w", err)
	}

	// Combine everything into ServiceDetailInfo
	detailsList := make([]models.ServiceDetailInfo, 0, len(allServices))
	for _, service := range allServices {
		detail := models.ServiceDetailInfo{
			Name:         service.Name,
			Type:         service.Type,
			Status:       "stopped",  // Default
			Dependencies: []string{}, // Default
		}

		// Get status from status map
		if statusInfo, exists := statusMap[service.Name]; exists {
			detail.Status = statusInfo.Status
			if statusInfo.Error != "" {
				detail.StatusError = statusInfo.Error
			}
		}

		// Get dependencies from dependency map
		if deps, exists := dependencyMap[service.Name]; exists {
			detail.Dependencies = deps
		}

		detailsList = append(detailsList, detail)
	}

	// Sort by name for consistent UI display
	sort.Slice(detailsList, func(i, j int) bool {
		return detailsList[i].Name < detailsList[j].Name
	})

	return detailsList, nil
}

// GetMultipleServiceStatuses fetches statuses for multiple services concurrently
// This is optimized for progressive loading where we want to update statuses in batches
func (h *ServiceHandler) GetMultipleServiceStatuses(serviceNames []string) (map[string]models.ServiceStatus, error) {
	if len(serviceNames) == 0 {
		return make(map[string]models.ServiceStatus), nil
	}

	statusMap := make(map[string]models.ServiceStatus)
	resultsChan := make(chan models.ServiceStatus, len(serviceNames))
	var wg sync.WaitGroup

	composeFiles := h.getComposeFiles()

	for _, serviceName := range serviceNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			var status models.ServiceStatus
			status.ServiceName = name

			// Fetch status
			statusStr, statusErr := h.getServiceStatusInternal(name, composeFiles)
			status.Status = statusStr
			if statusErr != nil {
				status.Error = statusErr.Error()
				status.Status = "error"
			}

			resultsChan <- status
		}(serviceName)
	}

	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	for status := range resultsChan {
		statusMap[status.ServiceName] = status
	}

	return statusMap, nil
}

// markServiceStopped marks a service as stopped for immediate status updates
func (h *ServiceHandler) markServiceStopped(serviceName string) {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	h.stoppedServices[serviceName] = true
	// Clear starting status when marking as stopped
	delete(h.startingServices, serviceName)
}

// markAllServicesStopped marks all services as stopped for immediate status updates
func (h *ServiceHandler) markAllServicesStopped() {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()

	// Mark all known services as stopped
	allServices := h.ListServices()
	for _, service := range allServices {
		h.stoppedServices[service.Name] = true
	}
}

// clearStoppedServiceTracking clears the stopped service tracking
// This should be called when we want to refresh from actual container status
func (h *ServiceHandler) clearStoppedServiceTracking() {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	h.stoppedServices = make(map[string]bool)
	h.startingServices = make(map[string]bool)
}

// isServiceMarkedStopped checks if a service is marked as stopped in our tracking
func (h *ServiceHandler) isServiceMarkedStopped(serviceName string) bool {
	h.statusTrackingMutex.RLock()
	defer h.statusTrackingMutex.RUnlock()
	return h.stoppedServices[serviceName]
}

// markServiceStarting marks a service as starting for immediate status updates
func (h *ServiceHandler) markServiceStarting(serviceName string) {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	h.startingServices[serviceName] = true
	// Clear stopped status when marking as starting
	delete(h.stoppedServices, serviceName)
}

// clearServiceStarting clears the starting status for a service
func (h *ServiceHandler) clearServiceStarting(serviceName string) {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	delete(h.startingServices, serviceName)
	// Also clear stopped status when clearing starting status
	delete(h.stoppedServices, serviceName)
}

// isServiceMarkedStarting checks if a service is marked as starting in our tracking
func (h *ServiceHandler) isServiceMarkedStarting(serviceName string) bool {
	h.statusTrackingMutex.RLock()
	defer h.statusTrackingMutex.RUnlock()
	return h.startingServices[serviceName]
}

// Public methods for testing

// MarkServiceStoppedForTesting marks a service as stopped (for testing)
func (h *ServiceHandler) MarkServiceStoppedForTesting(serviceName string) {
	h.markServiceStopped(serviceName)
}

// MarkAllServicesStoppedForTesting marks all services as stopped (for testing)
func (h *ServiceHandler) MarkAllServicesStoppedForTesting() {
	h.markAllServicesStopped()
}

// ClearStoppedServiceTrackingForTesting clears the stopped service tracking (for testing)
func (h *ServiceHandler) ClearStoppedServiceTrackingForTesting() {
	h.clearStoppedServiceTracking()
}

// IsServiceMarkedStoppedForTesting checks if a service is marked as stopped (for testing)
func (h *ServiceHandler) IsServiceMarkedStoppedForTesting(serviceName string) bool {
	return h.isServiceMarkedStopped(serviceName)
}

// StopServiceWithStatusUpdate stops a service and returns updated status information
// This provides immediate status updates without requiring additional Docker/Podman calls
func (h *ServiceHandler) StopServiceWithStatusUpdate(serviceName string) (map[string]models.ServiceStatus, error) {
	// Stop the service first
	err := h.StopService(serviceName)
	if err != nil {
		return nil, err
	}

	// Return updated status map with the stopped service
	// We can do this efficiently since we know what changed
	allServices := h.ListServices()
	statusMap := make(map[string]models.ServiceStatus)

	for _, service := range allServices {
		status := "stopped"
		if service.Name == serviceName {
			// We know this service is now stopped
			status = "stopped"
		} else if h.isServiceMarkedStopped(service.Name) {
			// This service was previously marked as stopped
			status = "stopped"
		} else {
			// For other services, we need to check if they're still running
			// But we can optimize this by only checking services that might be affected
			composeFiles := h.getComposeFiles()
			runningContainers, err := h.getRunningContainers()
			if err == nil && h.isServiceRunning(service.Name, composeFiles, runningContainers) {
				status = "running"
			}
		}

		statusMap[service.Name] = models.ServiceStatus{
			ServiceName: service.Name,
			Status:      status,
		}
	}

	return statusMap, nil
}

// StopAllServicesWithStatusUpdate stops all services and returns updated status information
// This provides immediate status updates without requiring additional Docker/Podman calls
func (h *ServiceHandler) StopAllServicesWithStatusUpdate() (map[string]models.ServiceStatus, error) {
	// Stop all services first
	err := h.StopAllServices()
	if err != nil {
		return nil, err
	}

	// Return updated status map with all services stopped
	// This is very efficient since we know all services are now stopped
	allServices := h.ListServices()
	statusMap := make(map[string]models.ServiceStatus)

	for _, service := range allServices {
		statusMap[service.Name] = models.ServiceStatus{
			ServiceName: service.Name,
			Status:      "stopped",
		}
	}

	return statusMap, nil
}

// RefreshStatusFromContainers clears the stopped service tracking and gets fresh status from containers
// This should be called when the UI wants to refresh from actual container state
func (h *ServiceHandler) RefreshStatusFromContainers() (map[string]models.ServiceStatus, error) {
	// Clear the stopped service tracking to get fresh status
	h.clearStoppedServiceTracking()

	// Invalidate container cache to force fresh data
	h.invalidateContainerCache()

	// Get fresh status from containers
	return h.GetAllRunningServices()
}

// CheckStartingServicesProgress checks if any services marked as starting have completed their transition
// Returns updated statuses only for services that have changed state
func (h *ServiceHandler) CheckStartingServicesProgress() (map[string]models.ServiceStatus, error) {
	h.statusTrackingMutex.RLock()
	startingServiceNames := make([]string, 0, len(h.startingServices))
	for serviceName := range h.startingServices {
		startingServiceNames = append(startingServiceNames, serviceName)
	}
	h.statusTrackingMutex.RUnlock()

	// If no services are starting, return empty map
	if len(startingServiceNames) == 0 {
		return make(map[string]models.ServiceStatus), nil
	}

	// Check status for each starting service
	composeFiles := h.getComposeFiles()
	updatedStatuses := make(map[string]models.ServiceStatus)

	for _, serviceName := range startingServiceNames {
		currentStatus, err := h.getServiceStatusInternal(serviceName, composeFiles)
		if err != nil {
			updatedStatuses[serviceName] = models.ServiceStatus{
				ServiceName: serviceName,
				Status:      "error",
				Error:       err.Error(),
			}
		} else {
			// Only include in response if status has changed from starting
			if currentStatus != "starting" {
				updatedStatuses[serviceName] = models.ServiceStatus{
					ServiceName: serviceName,
					Status:      currentStatus,
				}
			}
		}
	}

	return updatedStatuses, nil
}
