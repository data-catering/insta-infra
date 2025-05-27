package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// DependencyHandler handles dependency management operations
type DependencyHandler struct {
	*BaseHandler
	serviceHandler *ServiceHandler // Reference to service handler for status tracking
}

// NewDependencyHandler creates a new dependency handler
func NewDependencyHandler(runtime container.Runtime, instaDir string, serviceHandler *ServiceHandler) *DependencyHandler {
	return &DependencyHandler{
		BaseHandler:    NewBaseHandler(runtime, instaDir),
		serviceHandler: serviceHandler,
	}
}

// GetDependencyStatus returns detailed status information about all dependencies of a service
func (h *DependencyHandler) GetDependencyStatus(serviceName string) (*models.DependencyStatus, error) {
	composeFiles := h.getComposeFiles()

	// Get all dependencies recursively
	allDeps, err := h.Runtime().GetAllDependenciesRecursive(serviceName, composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get dependencies for service %s: %w", serviceName, err)
	}

	status := &models.DependencyStatus{
		ServiceName:        serviceName,
		Dependencies:       make([]models.DependencyInfo, 0, len(allDeps)),
		FailedDependencies: make([]string, 0),
	}

	// Get status for each dependency
	for i, depName := range allDeps {
		depInfo := models.DependencyInfo{
			ServiceName:  depName,
			Required:     true, // For now, treat all dependencies as required
			StartupOrder: i + 1,
			HasLogs:      true, // Assume logs are available for investigation
		}

		// Get service info to determine type
		if service := h.getServiceByName(depName); service != nil {
			depInfo.Type = service.Type
		}

		// Get current status with enhanced failure detection
		depStatus, failureReason := h.getEnhancedServiceStatus(depName, composeFiles)
		depInfo.Status = depStatus
		depInfo.FailureReason = failureReason

		// Analyze status and health
		switch depStatus {
		case "running":
			depInfo.Health = "healthy"
			status.RunningCount++
		case "error", "failed", "exited":
			depInfo.Health = "unhealthy"
			status.ErrorCount++
			status.FailedDependencies = append(status.FailedDependencies, depName)

			// Set error information
			if failureReason != "" {
				depInfo.Error = failureReason
			} else {
				depInfo.Error = "Service failed to start"
			}
		case "completed":
			// For init containers that completed successfully
			depInfo.Health = "healthy"
			status.RunningCount++ // Count as "running" for dependency purposes
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

// getEnhancedServiceStatus provides detailed status information including failure detection
func (h *DependencyHandler) getEnhancedServiceStatus(serviceName string, composeFiles []string) (status, failureReason string) {
	// First try to get basic status
	basicStatus, err := h.GetServiceStatusInternal(serviceName, composeFiles)
	if err != nil {
		return "error", err.Error()
	}

	// Always try to get container information for better status detection
	containerName, err := h.Runtime().GetContainerName(serviceName, composeFiles)
	if err != nil {
		// If we can't get container name, return basic status
		return basicStatus, ""
	}

	// Get detailed container status
	containerStatus, err := h.Runtime().GetContainerStatus(containerName)
	if err != nil {
		// If we can't get container status, return basic status
		return basicStatus, ""
	}

	// Analyze container status for failure patterns
	switch containerStatus {
	case "running":
		return "running", ""
	case "completed":
		// Init containers that completed successfully
		return "completed", ""
	case "error":
		// Container failed (exited with non-zero code)
		failureReason = h.analyzeContainerFailure(serviceName, containerName)
		return "failed", failureReason
	case "exited":
		// Container exited - could be success or failure depending on context
		if strings.Contains(serviceName, "-init") || strings.Contains(serviceName, "-data") {
			// For init containers, exited usually means failed (completed would be success)
			failureReason = h.analyzeContainerFailure(serviceName, containerName)
			return "failed", failureReason
		} else {
			// For regular containers, exited is always a failure
			failureReason = h.analyzeContainerFailure(serviceName, containerName)
			return "failed", failureReason
		}
	case "created":
		return "stopped", "Container created but not started"
	case "dead":
		return "failed", "Container is in dead state"
	case "paused":
		return "stopped", "Container is paused"
	case "restarting":
		return "starting", "Container is restarting"
	case "not_found":
		return "stopped", "Container not found"
	default:
		// For unknown statuses or when container doesn't exist
		if basicStatus == "running" {
			return "running", ""
		}
		// If basic status is stopped but we have container info, it might have failed
		if containerStatus != "" {
			return "stopped", fmt.Sprintf("Container status: %s", containerStatus)
		}
		return basicStatus, ""
	}
}

// analyzeContainerFailure provides specific failure analysis for common patterns
func (h *DependencyHandler) analyzeContainerFailure(serviceName, containerName string) string {
	// Get recent logs to analyze failure patterns
	logs, err := h.Runtime().GetContainerLogs(containerName, 10)
	if err != nil {
		return "Container failed - logs unavailable"
	}

	// Analyze logs for common failure patterns
	if len(logs) == 0 {
		return "Container exited - no logs available"
	}

	// Join all log lines for analysis
	logText := strings.ToLower(strings.Join(logs, " "))

	// Check for common Airflow init failures
	if strings.Contains(serviceName, "airflow") && strings.Contains(serviceName, "init") {
		if containsAny(logText, []string{"syntax error", "unexpected token"}) {
			return "Script syntax error - check init.sh script"
		}
		if containsAny(logText, []string{"database", "connection", "postgres"}) {
			return "Database connection failed - ensure PostgreSQL is running"
		}
		if containsAny(logText, []string{"permission", "denied", "access"}) {
			return "Permission denied - check file permissions"
		}
		if containsAny(logText, []string{"timeout", "timed out"}) {
			return "Operation timed out - database may be slow to respond"
		}
		return "Initialization failed - check container logs for details"
	}

	// Check for database connection issues
	if containsAny(logText, []string{"connection refused", "could not connect", "database"}) {
		return "Database connection failed"
	}

	// Check for permission issues
	if containsAny(logText, []string{"permission denied", "access denied", "forbidden"}) {
		return "Permission or access denied"
	}

	// Check for resource issues
	if containsAny(logText, []string{"out of memory", "disk space", "no space"}) {
		return "Resource constraint (memory/disk)"
	}

	// Check for port binding issues
	if containsAny(logText, []string{"port already in use", "address already in use", "bind"}) {
		return "Port binding failed - port may be in use"
	}

	// Generic failure with last log line
	if len(logs) > 0 {
		lastLog := logs[len(logs)-1]
		if len(lastLog) > 100 {
			lastLog = lastLog[:100] + "..."
		}
		return fmt.Sprintf("Container exited - last log: %s", lastLog)
	}

	return "Container exited unexpectedly"
}

// containsAny checks if text contains any of the given substrings (case-insensitive)
func containsAny(text string, substrings []string) bool {
	textLower := strings.ToLower(text)
	for _, substr := range substrings {
		if strings.Contains(textLower, strings.ToLower(substr)) {
			return true
		}
	}
	return false
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
		return h.serviceHandler.getServiceStatusInternal(serviceName, composeFiles)
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
