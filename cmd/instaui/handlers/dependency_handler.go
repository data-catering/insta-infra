package handlers

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// DependencyHandler handles dependency management operations
type DependencyHandler struct {
	*BaseHandler
	serviceHandler ServiceHandlerInterface // Reference to service handler for status tracking
}

// NewDependencyHandler creates a new dependency handler
func NewDependencyHandler(runtime container.Runtime, instaDir string, serviceHandler ServiceHandlerInterface) *DependencyHandler {
	return &DependencyHandler{
		BaseHandler:    NewBaseHandler(runtime, instaDir),
		serviceHandler: serviceHandler,
	}
}

// GetDependencyStatus returns detailed status information about dependencies of a service
// Uses a fast approach: assumes dependencies are stopped unless the service is running,
// then only checks dependencies for running services
func (h *DependencyHandler) GetDependencyStatus(serviceName string) (*models.DependencyStatus, error) {
	// If we don't have a service handler (e.g., in tests), fall back to the old behavior
	if h.serviceHandler == nil {
		composeFiles := h.getComposeFiles()
		return h.getLegacyDependencyStatus(serviceName, composeFiles)
	}

	// Create a default empty status that we can return quickly if anything fails
	defaultStatus := &models.DependencyStatus{
		ServiceName:          serviceName,
		Dependencies:         []models.DependencyInfo{},
		FailedDependencies:   []string{},
		AllDependenciesReady: true, // Assume ready if we can't determine dependencies
		CanStart:             true,
		RequiredCount:        0,
		RunningCount:         0,
		ErrorCount:           0,
	}

	// First check if the service itself is running (with timeout protection)
	serviceStatus, err := h.serviceHandler.GetServiceStatus(serviceName)
	if err != nil {
		// If we can't get service status, assume it's stopped and return default
		return defaultStatus, nil
	}

	// If service is in error state, return immediately
	if serviceStatus == "error" {
		return defaultStatus, nil
	}

	// Get direct dependencies (not recursive) for faster lookup
	// First try to get dependencies from the service handler if it supports it
	var directDeps []string
	
	if deps, depErr := h.serviceHandler.GetServiceDependencies(serviceName); depErr == nil {
		directDeps = deps
	} else {
		// Fallback to runtime-based dependency lookup
		composeFiles := h.getComposeFiles()
		var depErr error
		directDeps, depErr = h.Runtime().GetDependencies(serviceName, composeFiles)
		if depErr != nil {
			// If we can't get dependencies, return default status (no dependencies)
			return defaultStatus, nil
		}
	}

	status := &models.DependencyStatus{
		ServiceName:        serviceName,
		Dependencies:       make([]models.DependencyInfo, 0, len(directDeps)),
		FailedDependencies: make([]string, 0),
	}

	// If the service is not running, assume all dependencies are stopped (fast path)
	if serviceStatus != "running" && serviceStatus != "starting" {
		for i, depName := range directDeps {
			depInfo := models.DependencyInfo{
				ServiceName:  depName,
				Required:     true,
				StartupOrder: i + 1,
				HasLogs:      true,
				Status:       "stopped",
				Health:       "unknown",
			}

			// Get service info to determine type
			if service := h.getServiceByName(depName); service != nil {
				depInfo.Type = service.Type
			}

			status.Dependencies = append(status.Dependencies, depInfo)
			status.RequiredCount++
		}

		// Service can start if no dependencies or all are assumed ready
		status.AllDependenciesReady = len(directDeps) == 0
		status.CanStart = status.AllDependenciesReady
		return status, nil
	}

	// Service is running/starting, so check actual dependency statuses
	// But do this with error protection to avoid hanging
	for i, depName := range directDeps {
		depInfo := models.DependencyInfo{
			ServiceName:  depName,
			Required:     true,
			StartupOrder: i + 1,
			HasLogs:      true,
			Status:       "unknown", // Default to unknown
			Health:       "unknown",
		}

		// Get service info to determine type
		if service := h.getServiceByName(depName); service != nil {
			depInfo.Type = service.Type
		}

		// Get current status with error protection
		depStatus, err := h.serviceHandler.GetServiceStatus(depName)
		if err != nil {
			// If we can't get dependency status, assume it's stopped
			depStatus = "stopped"
			depInfo.Status = "stopped"
			depInfo.Health = "unknown"
		} else {
			depInfo.Status = depStatus

			// Analyze status and health
			switch depStatus {
			case "running":
				depInfo.Health = "healthy"
				status.RunningCount++
			case "error", "failed", "exited":
				depInfo.Health = "unhealthy"
				status.ErrorCount++
				status.FailedDependencies = append(status.FailedDependencies, depName)
				depInfo.Error = "Service failed to start"
			case "starting":
				depInfo.Health = "unknown"
			default:
				depInfo.Health = "unknown"
			}
		}

		if depInfo.Required {
			status.RequiredCount++
		}

		status.Dependencies = append(status.Dependencies, depInfo)
	}

	// Determine if service can start (all required dependencies running)
	status.AllDependenciesReady = (status.RunningCount == status.RequiredCount && status.ErrorCount == 0)
	status.CanStart = status.AllDependenciesReady

	return status, nil
}

// getLegacyDependencyStatus provides the old behavior for backward compatibility (mainly for tests)
func (h *DependencyHandler) getLegacyDependencyStatus(serviceName string, composeFiles []string) (*models.DependencyStatus, error) {
	// Get all dependencies recursively (old behavior)
	allDeps, err := h.Runtime().GetAllDependenciesRecursive(serviceName, composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies for service %s: %w", serviceName, err)
	}

	status := &models.DependencyStatus{
		ServiceName:        serviceName,
		Dependencies:       make([]models.DependencyInfo, 0, len(allDeps)),
		FailedDependencies: make([]string, 0),
	}

	// Get status for each dependency using the old method
	for i, depName := range allDeps {
		depInfo := models.DependencyInfo{
			ServiceName:  depName,
			Required:     true,
			StartupOrder: i + 1,
			HasLogs:      true,
		}

		// Get service info to determine type
		if service := h.getServiceByName(depName); service != nil {
			depInfo.Type = service.Type
		}

		// Get current status using internal method
		depStatus, err := h.GetServiceStatusInternal(depName, composeFiles)
		if err != nil {
			depStatus = "error"
			depInfo.Error = err.Error()
		}

		depInfo.Status = depStatus

		// Analyze status and health
		switch depStatus {
		case "running":
			depInfo.Health = "healthy"
			status.RunningCount++
		case "error", "failed", "exited":
			depInfo.Health = "unhealthy"
			status.ErrorCount++
			status.FailedDependencies = append(status.FailedDependencies, depName)
			if depInfo.Error == "" {
				depInfo.Error = "Service failed to start"
			}
		case "starting":
			depInfo.Health = "unknown"
		default:
			depInfo.Health = "unknown"
		}

		if depInfo.Required {
			status.RequiredCount++
		}

		status.Dependencies = append(status.Dependencies, depInfo)
	}

	// Determine if service can start
	status.AllDependenciesReady = (status.RunningCount == status.RequiredCount && status.ErrorCount == 0)
	status.CanStart = status.AllDependenciesReady

	return status, nil
}

// StartAllDependencies starts all dependencies of a service in the correct order
func (h *DependencyHandler) StartAllDependencies(serviceName string, persist bool) error {
	composeFiles := h.getComposeFiles()
	if persist {
		composeFiles = append(composeFiles, filepath.Join(h.InstaDir(), "docker-compose-persist.yaml"))
	}

	// Get all dependencies recursively
	allDeps, err := h.Runtime().GetAllDependenciesRecursive(serviceName, composeFiles)
	if err != nil {
		return fmt.Errorf("failed to get dependencies for service %s: %w", serviceName, err)
	}

	if len(allDeps) == 0 {
		return fmt.Errorf("service %s has no dependencies to start", serviceName)
	}

	// Start each dependency individually
	for _, depName := range allDeps {
		fmt.Printf("Starting dependency: %s\n", depName)

		// Check if dependency is already running
		status, err := h.GetServiceStatusInternal(depName, composeFiles)
		if err == nil && status == "running" {
			fmt.Printf("Dependency %s is already running, skipping\n", depName)
			continue
		}

		// Start the dependency using service handler logic
		if err := h.startSingleService(depName, persist); err != nil {
			return fmt.Errorf("failed to start dependency %s: %w", depName, err)
		}
	}

	return nil
}

// StopDependencyChain stops a service and all services that depend on it
func (h *DependencyHandler) StopDependencyChain(serviceName string) error {
	composeFiles := h.getComposeFiles()

	// Find all services that depend on this service
	servicesToStop := []string{serviceName}

	// Get all services and check which ones depend on the target service
	for serviceKey := range core.Services {
		if serviceKey == serviceName {
			continue // Skip the service itself
		}

		// Get dependencies for this service
		deps, err := h.Runtime().GetAllDependenciesRecursive(serviceKey, composeFiles)
		if err != nil {
			fmt.Printf("Warning: failed to get dependencies for %s: %v\n", serviceKey, err)
			continue
		}

		// Check if the target service is in this service's dependencies
		for _, dep := range deps {
			if dep == serviceName {
				servicesToStop = append(servicesToStop, serviceKey)
				break
			}
		}
	}

	if len(servicesToStop) == 1 {
		fmt.Printf("No services depend on %s, stopping only %s\n", serviceName, serviceName)
	} else {
		fmt.Printf("Stopping %s and %d dependent services: %v\n", serviceName, len(servicesToStop)-1, servicesToStop[1:])
	}

	// Stop all affected services
	for _, serviceToStop := range servicesToStop {
		fmt.Printf("Stopping service: %s\n", serviceToStop)

		// Check if service is actually running
		status, err := h.GetServiceStatusInternal(serviceToStop, composeFiles)
		if err == nil && status != "running" {
			fmt.Printf("Service %s is not running, skipping\n", serviceToStop)
			continue
		}

		// Stop the service
		if err := h.stopSingleService(serviceToStop); err != nil {
			return fmt.Errorf("failed to stop service %s: %w", serviceToStop, err)
		}
	}

	return nil
}

// Helper methods

// getServiceByName is a helper method to find a service by name
func (h *DependencyHandler) getServiceByName(serviceName string) *core.Service {
	if service, exists := core.Services[serviceName]; exists {
		return &service
	}
	return nil
}

// GetServiceStatusInternal is a helper for getting service status with tracking
func (h *DependencyHandler) GetServiceStatusInternal(serviceName string, composeFiles []string) (string, error) {
	// Use the service handler's status checking which includes stopped service tracking
	if h.serviceHandler != nil {
		return h.serviceHandler.GetServiceStatus(serviceName)
	}

	// Fallback to direct checking if no service handler available
	// Get all running containers first
	runningContainers, err := h.getRunningContainers()
	if err != nil {
		return "stopped", fmt.Errorf("could not get running containers: %w", err)
	}

	// Use the new method to check if service is running
	if h.isServiceRunning(serviceName, composeFiles, runningContainers) {
		return "running", nil
	}

	return "stopped", nil
}

// startSingleService starts a single service (simplified version of ServiceHandler.StartService)
func (h *DependencyHandler) startSingleService(serviceName string, persist bool) error {
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

	return h.Runtime().ComposeUp(composeFiles, services, true)
}

// stopSingleService stops a single service (simplified version of ServiceHandler.StopService)
func (h *DependencyHandler) stopSingleService(serviceName string) error {
	services := []string{serviceName}
	composeFiles := []string{
		filepath.Join(h.InstaDir(), "docker-compose.yaml"),
		filepath.Join(h.InstaDir(), "docker-compose-persist.yaml"),
	}

	return h.Runtime().ComposeDown(composeFiles, services)
}
