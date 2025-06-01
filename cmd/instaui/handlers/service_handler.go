package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// GetAllServiceDetails fetches status for all services concurrently
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
	statusMap, err := h.getOptimizedStatusUpdate(serviceName)
	if err != nil {
		return nil, err
	}

	return statusMap, nil
}

// getOptimizedStatusUpdate returns status for all services with optimized checking for the target service
func (h *ServiceHandler) getOptimizedStatusUpdate(targetService string) (map[string]models.ServiceStatus, error) {
	// Get all current containers
	currentContainers, err := h.getCurrentContainers()
	if err != nil {
		return nil, fmt.Errorf("failed to get current containers: %w", err)
	}

	statusMap := make(map[string]models.ServiceStatus)

	for container, status := range currentContainers {
		statusMap[container] = models.ServiceStatus{
			ServiceName: container,
			Status:      status,
		}
	}

	return statusMap, nil
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

	// Use getCurrentContainers for consistency with GetAllRunningServices
	currentContainers, err := h.getCurrentContainers()
	if err != nil {
		// If we can't get current containers, assume stopped
		return "stopped", nil
	}

	// Use direct service name as container name (guaranteed by compose architecture)
	mainContainerName := serviceName

	// Check the status of the main container from the current containers map
	containerStatus, exists := currentContainers[mainContainerName]
	if !exists {
		// Container doesn't exist
		return "stopped", nil
	}

	// Clear starting status and return appropriate status based on container state
	switch containerStatus {
	case "running":
		h.clearServiceStarting(serviceName)
		return "running", nil
	case "starting", "restarting":
		return "starting", nil
	case "completed":
		h.clearServiceStarting(serviceName)
		return "completed", nil
	case "error", "exited", "dead":
		h.clearServiceStarting(serviceName)
		return "failed", nil
	case "created":
		h.clearServiceStarting(serviceName)
		return "failed", nil
	default:
		return "stopped", nil
	}
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

	// Get all current containers with single optimized call
	currentContainers, err := h.getCurrentContainers()
	if err != nil {
		// If we can't get current containers, return all as stopped with error
		for serviceName := range statusMap {
			statusMap[serviceName] = models.ServiceStatus{
				ServiceName: serviceName,
				Status:      "error",
				Error:       fmt.Sprintf("could not get current containers: %v", err),
			}
		}
		return statusMap, nil
	}

	for container, status := range currentContainers {
		statusMap[container] = models.ServiceStatus{
			ServiceName: container,
			Status:      status,
		}
	}

	return statusMap, nil
}

// isServiceRunningFast checks if a service is running using direct container lookup
// Uses the direct service name as container name (guaranteed by compose architecture)
func (h *ServiceHandler) isServiceRunningFast(serviceName string, currentContainers map[string]string) bool {
	// Direct check: service name equals main container name
	// This is guaranteed by our compose file structure
	return currentContainers[serviceName] == "running"
}

// GetAllServicesWithStatusAndDependencies efficiently gets services and statuses in a single call
func (h *ServiceHandler) GetAllServicesWithStatusAndDependencies() ([]models.ServiceDetailInfo, error) {
	// Get basic services list (filtered)
	allServices := h.ListServices()

	// Get all current containers statuses
	statusMap, err := h.getCurrentContainers()
	if err != nil {
		return nil, fmt.Errorf("failed to get current containers statuses: %w", err)
	}

	// Get all dependencies efficiently
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
		if status, exists := statusMap[service.Name]; exists {
			detail.Status = status
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

// GetAllServicesStatuses gets all service statuses including stopped containers
// This is different from GetAllRunningServices which only checks running containers
func (h *ServiceHandler) GetAllServicesStatuses() (map[string]models.ServiceStatus, error) {
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

	// Check each service's actual container status (including stopped containers)
	for _, service := range allServices {
		// Check if service is marked as stopped first (immediate update)
		if h.isServiceMarkedStopped(service.Name) {
			continue // Keep as stopped (already set as default)
		}

		// Check if service is marked as starting first (immediate update)
		if h.isServiceMarkedStarting(service.Name) {
			statusMap[service.Name] = models.ServiceStatus{
				ServiceName: service.Name,
				Status:      "starting",
			}
			continue
		}

		// Get actual container status (checks both running and stopped containers)
		containerStatus, err := h.Runtime().GetContainerStatus(service.Name)
		if err != nil {
			// If there's an error getting status, keep as stopped but note the error
			statusMap[service.Name] = models.ServiceStatus{
				ServiceName: service.Name,
				Status:      "stopped",
				Error:       fmt.Sprintf("could not check container status: %v", err),
			}
			continue
		}

		// Map container status to service status
		var serviceStatus string
		switch containerStatus {
		case "running":
			serviceStatus = "running"
			h.clearServiceStarting(service.Name) // Clear starting status since service is now running
		case "starting", "restarting":
			serviceStatus = "starting"
		case "completed":
			serviceStatus = "completed"
			h.clearServiceStarting(service.Name)
		case "error", "exited", "dead":
			serviceStatus = "failed"
			h.clearServiceStarting(service.Name)
		case "created":
			serviceStatus = "failed"
			h.clearServiceStarting(service.Name)
		case "stopped", "not_found":
			serviceStatus = "stopped"
		case "paused":
			serviceStatus = "paused"
		default:
			serviceStatus = "stopped"
		}

		statusMap[service.Name] = models.ServiceStatus{
			ServiceName: service.Name,
			Status:      serviceStatus,
		}
	}

	return statusMap, nil
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
			currentContainers, err := h.getCurrentContainers()
			if err == nil && h.isServiceRunning(service.Name, composeFiles, currentContainers) {
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
	// This is very efficient since we know all services are now stopped otherwise error is thrown from above
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

// GetServiceDependencies returns the dependencies for a specific service (container names)
func (h *ServiceHandler) GetServiceDependencies(serviceName string) ([]string, error) {
	composeFiles := h.getComposeFiles()
	return h.Runtime().GetAllDependenciesRecursive(serviceName, composeFiles, true)
}

// GetAllServiceDependencies returns dependencies for all services (container names)
func (h *ServiceHandler) GetAllServiceDependencies() (map[string][]string, error) {
	// Get all available services first (filtered list)
	allServices := h.ListServices()

	// Create result map
	dependencyMap := make(map[string][]string)

	// Get dependencies for each service
	for _, service := range allServices {
		dependencies, err := h.GetServiceDependencies(service.Name)
		if err != nil {
			// If we can't get dependencies, log the error but continue
			// Set empty dependencies for this service
			dependencyMap[service.Name] = []string{}
		} else {
			dependencyMap[service.Name] = dependencies
		}
	}

	return dependencyMap, nil
}

// GetAllDependencyStatuses returns the statuses of all dependency containers (container names)
// This is used by the frontend to show individual container statuses in dependency lists
func (h *ServiceHandler) GetAllDependencyStatuses() (map[string]models.ServiceStatus, error) {
	// Get all service dependencies first
	dependencyMap, err := h.GetAllServiceDependencies()
	if err != nil {
		return nil, fmt.Errorf("failed to get service dependencies: %w", err)
	}

	// Collect all unique dependency container names
	allDependencyContainers := make(map[string]bool)
	for _, dependencies := range dependencyMap {
		for _, dep := range dependencies {
			allDependencyContainers[dep] = true
		}
	}

	// Create result map
	statusMap := make(map[string]models.ServiceStatus)

	// Get status for each dependency container
	for containerName := range allDependencyContainers {
		// Get actual container status (checks both running and stopped containers)
		containerStatus, err := h.Runtime().GetContainerStatus(containerName)
		if err != nil {
			statusMap[containerName] = models.ServiceStatus{
				ServiceName: containerName,
				Status:      "stopped",
				Error:       fmt.Sprintf("could not check container status: %v", err),
			}
			continue
		}

		// Map container status to service status
		var serviceStatus string
		switch containerStatus {
		case "running":
			serviceStatus = "running"
		case "starting", "restarting":
			serviceStatus = "starting"
		case "completed":
			serviceStatus = "completed"
		case "error", "exited", "dead":
			serviceStatus = "failed"
		case "created":
			serviceStatus = "failed"
		case "stopped", "not_found":
			serviceStatus = "stopped"
		case "paused":
			serviceStatus = "paused"
		default:
			serviceStatus = "unknown"
		}

		statusMap[containerName] = models.ServiceStatus{
			ServiceName: containerName,
			Status:      serviceStatus,
		}
	}

	return statusMap, nil
}

// getMainContainerForService determines the main container for a service
// In our compose architecture, the main container always has container_name matching the service name
func (h *ServiceHandler) getMainContainerForService(serviceName string, composeFiles []string) (string, error) {
	// Direct mapping: service name equals main container name
	// This is guaranteed by our compose file structure where each logical service
	// has a corresponding container with container_name: <serviceName>
	return serviceName, nil
}
