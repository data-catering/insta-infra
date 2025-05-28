package handlers

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
)

// CLIServiceHandler handles service operations using the bundled CLI binary
type CLIServiceHandler struct {
	cliPath  string
	instaDir string

	// Status tracking for immediate updates (similar to ServiceHandler)
	stoppedServices     map[string]bool
	startingServices    map[string]bool
	statusTrackingMutex sync.RWMutex
}

// NewCLIServiceHandler creates a new CLI-based service handler
func NewCLIServiceHandler(cliPath, instaDir string) *CLIServiceHandler {
	return &CLIServiceHandler{
		cliPath:          cliPath,
		instaDir:         instaDir,
		stoppedServices:  make(map[string]bool),
		startingServices: make(map[string]bool),
	}
}

// executeCLI executes a CLI command and returns the output
func (h *CLIServiceHandler) executeCLI(args ...string) ([]byte, error) {
	cmd := exec.Command(h.cliPath, args...)
	// Set working directory to instaDir for proper compose file access
	cmd.Dir = h.instaDir
	return cmd.Output()
}

// ListServices lists all available services using the CLI
func (h *CLIServiceHandler) ListServices() []models.ServiceInfo {
	// Use the core.Services directly since the CLI just lists service names
	// This ensures consistency with the direct runtime approach
	var serviceList []models.ServiceInfo

	for _, service := range core.Services {
		serviceList = append(serviceList, models.ServiceInfo(service))
	}

	// Sort services by name for consistent ordering in the UI
	sort.Slice(serviceList, func(i, j int) bool {
		return serviceList[i].Name < serviceList[j].Name
	})

	return serviceList
}

// GetServiceStatus gets the status of a specific service using container runtime commands
func (h *CLIServiceHandler) GetServiceStatus(serviceName string) (string, error) {
	// Check if service is marked as starting first (immediate update)
	if h.isServiceMarkedStarting(serviceName) {
		return "starting", nil
	}

	// Check if service is marked as stopped first (immediate update)
	if h.isServiceMarkedStopped(serviceName) {
		return "stopped", nil
	}

	// Use docker/podman ps to check if containers are running
	// This is more reliable than trying to parse CLI output
	status, err := h.getServiceStatusFromContainers(serviceName)
	if err != nil {
		return "error", fmt.Errorf("failed to get service status: %w", err)
	}

	// Clear starting status if service is now running
	if status == "running" {
		h.clearServiceStarting(serviceName)
	}

	return status, nil
}

// getServiceStatusFromContainers checks service status using container runtime
func (h *CLIServiceHandler) getServiceStatusFromContainers(serviceName string) (string, error) {
	// Try docker first, then podman
	runtimes := []string{"docker", "podman"}

	for _, runtime := range runtimes {
		// Check if runtime is available
		if _, err := exec.LookPath(runtime); err != nil {
			continue
		}

		// Get containers with the service label
		cmd := exec.Command(runtime, "ps", "--filter", fmt.Sprintf("label=com.docker.compose.service=%s", serviceName), "--format", "{{.Status}}")
		output, err := cmd.Output()
		if err != nil {
			continue
		}

		outputStr := strings.TrimSpace(string(output))
		if outputStr == "" {
			// No containers found, service is stopped
			return "stopped", nil
		}

		// Parse container status
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Check for various status indicators
			if strings.Contains(strings.ToLower(line), "up") {
				return "running", nil
			} else if strings.Contains(strings.ToLower(line), "starting") {
				return "starting", nil
			} else if strings.Contains(strings.ToLower(line), "restarting") {
				return "starting", nil
			} else if strings.Contains(strings.ToLower(line), "exited") {
				return "stopped", nil
			}
		}

		// If we found containers but couldn't determine status, assume running
		return "running", nil
	}

	return "stopped", fmt.Errorf("no container runtime available")
}

// StartService starts a service using the CLI
func (h *CLIServiceHandler) StartService(serviceName string, persist bool) error {
	// Mark service as starting immediately
	h.markServiceStarting(serviceName)

	args := []string{serviceName}
	if persist {
		args = append([]string{"-p"}, args...)
	}

	_, err := h.executeCLI(args...)
	if err != nil {
		// Clear starting status if start failed
		h.clearServiceStarting(serviceName)
		return fmt.Errorf("failed to start service %s: %w", serviceName, err)
	}

	// Clear the stopped status for this service since it's now starting
	h.statusTrackingMutex.Lock()
	delete(h.stoppedServices, serviceName)
	h.statusTrackingMutex.Unlock()

	return nil
}

// StopService stops a service using the CLI
func (h *CLIServiceHandler) StopService(serviceName string) error {
	_, err := h.executeCLI("-d", serviceName)
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", serviceName, err)
	}

	// Mark the service as stopped for immediate status updates
	h.markServiceStopped(serviceName)

	return nil
}

// GetServiceDependencies gets the dependencies of a service
func (h *CLIServiceHandler) GetServiceDependencies(serviceName string) ([]string, error) {
	// Since the CLI doesn't have a direct dependency command, we'll use the compose files
	// to determine dependencies, similar to how the direct runtime approach works
	composeFiles := h.getComposeFiles()

	// Try to use docker/podman compose config to get dependencies
	runtimes := []string{"docker", "podman"}

	for _, runtime := range runtimes {
		if _, err := exec.LookPath(runtime); err != nil {
			continue
		}

		// Build compose command
		var cmd *exec.Cmd
		if runtime == "docker" {
			args := []string{"compose"}
			for _, file := range composeFiles {
				args = append(args, "-f", file)
			}
			args = append(args, "config", "--services")
			cmd = exec.Command("docker", args...)
		} else {
			// For podman, try podman-compose first, then podman compose
			if _, err := exec.LookPath("podman-compose"); err == nil {
				args := []string{}
				for _, file := range composeFiles {
					args = append(args, "-f", file)
				}
				args = append(args, "config")
				cmd = exec.Command("podman-compose", args...)
			} else {
				args := []string{"compose"}
				for _, file := range composeFiles {
					args = append(args, "-f", file)
				}
				args = append(args, "config", "--services")
				cmd = exec.Command("podman", args...)
			}
		}

		cmd.Dir = h.instaDir
		_, err := cmd.Output()
		if err != nil {
			continue
		}

		// For now, return empty dependencies since parsing compose config is complex
		// This could be enhanced to parse the actual dependencies from compose config
		return []string{}, nil
	}

	return []string{}, nil
}

// getComposeFiles returns the compose files to use
func (h *CLIServiceHandler) getComposeFiles() []string {
	return []string{
		filepath.Join(h.instaDir, "docker-compose.yaml"),
		filepath.Join(h.instaDir, "docker-compose-persist.yaml"),
	}
}

// GetMultipleServiceStatuses gets statuses for multiple services
func (h *CLIServiceHandler) GetMultipleServiceStatuses(serviceNames []string) (map[string]models.ServiceStatus, error) {
	statuses := make(map[string]models.ServiceStatus)

	// Use goroutines for concurrent status checking
	var wg sync.WaitGroup
	statusChan := make(chan models.ServiceStatus, len(serviceNames))

	for _, serviceName := range serviceNames {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			status, err := h.GetServiceStatus(name)
			serviceStatus := models.ServiceStatus{
				ServiceName: name,
				Status:      status,
			}
			if err != nil {
				serviceStatus.Status = "error"
				serviceStatus.Error = err.Error()
			}
			statusChan <- serviceStatus
		}(serviceName)
	}

	go func() {
		wg.Wait()
		close(statusChan)
	}()

	for status := range statusChan {
		statuses[status.ServiceName] = status
	}

	return statuses, nil
}

// GetAllRunningServices gets all running services
func (h *CLIServiceHandler) GetAllRunningServices() (map[string]models.ServiceStatus, error) {
	// Get all available services first
	allServices := h.ListServices()

	// Create result map with all services as "stopped" by default
	statusMap := make(map[string]models.ServiceStatus)
	for _, service := range allServices {
		statusMap[service.Name] = models.ServiceStatus{
			ServiceName: service.Name,
			Status:      "stopped",
		}
	}

	// Get statuses for all services
	serviceNames := make([]string, len(allServices))
	for i, service := range allServices {
		serviceNames[i] = service.Name
	}

	statuses, err := h.GetMultipleServiceStatuses(serviceNames)
	if err != nil {
		return statusMap, err
	}

	// Update the status map with actual statuses
	for serviceName, status := range statuses {
		statusMap[serviceName] = status
	}

	return statusMap, nil
}

// GetAllServiceDependencies gets dependencies for all services
func (h *CLIServiceHandler) GetAllServiceDependencies() (map[string][]string, error) {
	allServices := h.ListServices()
	dependencyMap := make(map[string][]string)

	for _, service := range allServices {
		dependencies, err := h.GetServiceDependencies(service.Name)
		if err != nil {
			// If we can't get dependencies for a service, just use empty list
			dependencies = []string{}
		}
		dependencyMap[service.Name] = dependencies
	}

	return dependencyMap, nil
}

// GetAllServicesWithStatusAndDependencies gets complete service information
func (h *CLIServiceHandler) GetAllServicesWithStatusAndDependencies() ([]models.ServiceDetailInfo, error) {
	// Get basic services list
	allServices := h.ListServices()

	// Get all statuses efficiently
	statusMap, err := h.GetAllRunningServices()
	if err != nil {
		return nil, fmt.Errorf("failed to get service statuses: %w", err)
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

// StartServiceWithStatusUpdate starts a service and returns updated statuses
func (h *CLIServiceHandler) StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]models.ServiceStatus, error) {
	err := h.StartService(serviceName, persist)
	if err != nil {
		return nil, err
	}

	// Return updated status
	status, _ := h.GetServiceStatus(serviceName)
	return map[string]models.ServiceStatus{
		serviceName: {
			ServiceName: serviceName,
			Status:      status,
		},
	}, nil
}

// StopServiceWithStatusUpdate stops a service and returns updated statuses
func (h *CLIServiceHandler) StopServiceWithStatusUpdate(serviceName string) (map[string]models.ServiceStatus, error) {
	err := h.StopService(serviceName)
	if err != nil {
		return nil, err
	}

	// Return updated status
	status, _ := h.GetServiceStatus(serviceName)
	return map[string]models.ServiceStatus{
		serviceName: {
			ServiceName: serviceName,
			Status:      status,
		},
	}, nil
}

// StopAllServices stops all services
func (h *CLIServiceHandler) StopAllServices() error {
	_, err := h.executeCLI("-d")
	if err != nil {
		return err
	}

	// Mark all services as stopped for immediate status updates
	h.markAllServicesStopped()

	return nil
}

// StopAllServicesWithStatusUpdate stops all services and returns updated statuses
func (h *CLIServiceHandler) StopAllServicesWithStatusUpdate() (map[string]models.ServiceStatus, error) {
	err := h.StopAllServices()
	if err != nil {
		return nil, err
	}

	// Return updated statuses for all services (all stopped)
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

// RefreshStatusFromContainers refreshes status from containers
func (h *CLIServiceHandler) RefreshStatusFromContainers() (map[string]models.ServiceStatus, error) {
	// Clear the stopped service tracking to get fresh status
	h.clearStoppedServiceTracking()

	// Get fresh status from containers
	return h.GetAllRunningServices()
}

// CheckStartingServicesProgress checks progress of starting services
func (h *CLIServiceHandler) CheckStartingServicesProgress() (map[string]models.ServiceStatus, error) {
	// Get current statuses for all services
	return h.GetAllRunningServices()
}

// Status tracking methods (similar to ServiceHandler)

// markServiceStopped marks a service as stopped for immediate status updates
func (h *CLIServiceHandler) markServiceStopped(serviceName string) {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	h.stoppedServices[serviceName] = true
	// Clear starting status when marking as stopped
	delete(h.startingServices, serviceName)
}

// markAllServicesStopped marks all services as stopped for immediate status updates
func (h *CLIServiceHandler) markAllServicesStopped() {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()

	// Mark all known services as stopped
	allServices := h.ListServices()
	for _, service := range allServices {
		h.stoppedServices[service.Name] = true
	}
	// Clear all starting statuses
	h.startingServices = make(map[string]bool)
}

// clearStoppedServiceTracking clears the stopped service tracking
func (h *CLIServiceHandler) clearStoppedServiceTracking() {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	h.stoppedServices = make(map[string]bool)
	h.startingServices = make(map[string]bool)
}

// isServiceMarkedStopped checks if a service is marked as stopped in our tracking
func (h *CLIServiceHandler) isServiceMarkedStopped(serviceName string) bool {
	h.statusTrackingMutex.RLock()
	defer h.statusTrackingMutex.RUnlock()
	return h.stoppedServices[serviceName]
}

// markServiceStarting marks a service as starting for immediate status updates
func (h *CLIServiceHandler) markServiceStarting(serviceName string) {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	h.startingServices[serviceName] = true
	// Clear stopped status when marking as starting
	delete(h.stoppedServices, serviceName)
}

// clearServiceStarting clears the starting status for a service
func (h *CLIServiceHandler) clearServiceStarting(serviceName string) {
	h.statusTrackingMutex.Lock()
	defer h.statusTrackingMutex.Unlock()
	delete(h.startingServices, serviceName)
}

// isServiceMarkedStarting checks if a service is marked as starting in our tracking
func (h *CLIServiceHandler) isServiceMarkedStarting(serviceName string) bool {
	h.statusTrackingMutex.RLock()
	defer h.statusTrackingMutex.RUnlock()
	return h.startingServices[serviceName]
}
