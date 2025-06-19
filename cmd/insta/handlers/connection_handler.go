package handlers

import (
	"fmt"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// ConnectionHandler handles service connection information with simplified logic
type ConnectionHandler struct {
	serviceManager *models.ServiceManager
	runtime        container.Runtime
	logger         Logger
}

// NewConnectionHandler creates a new simplified connection handler
func NewConnectionHandler(runtime container.Runtime, instaDir string, logger Logger) *ConnectionHandler {
	// Create runtime info adapter for service manager
	runtimeInfo := NewRuntimeInfoAdapter(runtime)

	// Create service manager to get enhanced service information
	serviceManager := models.NewServiceManager(instaDir, runtimeInfo, logger)

	// Load services to get connection and port information
	if err := serviceManager.LoadServices(); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to load services for connection handler: %v", err))
	}

	return &ConnectionHandler{
		serviceManager: serviceManager,
		runtime:        runtime,
		logger:         logger,
	}
}

// OpenServiceInBrowser opens the service's web interface in the browser
func (h *ConnectionHandler) OpenServiceInBrowser(serviceName string) (*models.EnhancedServiceConnectionInfo, error) {
	h.logger.Log(fmt.Sprintf("Getting web URL for service %s to open in browser", serviceName))

	// Get enhanced service data
	enhancedService, exists := h.serviceManager.GetServiceByContainerName(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Update service status to ensure we have current availability
	h.serviceManager.UpdateServiceStatus(serviceName)

	// Check if service has web UI
	if len(enhancedService.WebUrls) == 0 {
		return nil, fmt.Errorf("service %s does not have a web interface", serviceName)
	}

	// Return connection info with web URLs for frontend to handle browser opening
	connectionInfo := &models.EnhancedServiceConnectionInfo{
		ServiceName:           serviceName,
		Available:             true,
		Status:                enhancedService.Status,
		ContainerName:         enhancedService.ContainerName,
		ImageName:             enhancedService.ImageName,
		Type:                  enhancedService.Type,
		Username:              enhancedService.DefaultUser,
		Password:              enhancedService.DefaultPassword,
		ConnectionCommand:     enhancedService.ConnectionCmd,
		WebUrls:               enhancedService.WebUrls,
		ExposedPorts:          enhancedService.ExposedPorts,
		InternalPorts:         enhancedService.InternalPorts,
		DirectDependencies:    enhancedService.DirectDependencies,
		RecursiveDependencies: enhancedService.RecursiveDependencies,
		AllContainers:         enhancedService.AllContainers,
	}

	// Build connection strings and credentials
	connectionInfo.ConnectionStrings = h.buildConnectionStrings(serviceName, enhancedService)
	connectionInfo.Credentials = h.buildCredentials(serviceName, enhancedService)

	h.logger.Log(fmt.Sprintf("Returning connection info for service %s with %d web URLs", serviceName, len(enhancedService.WebUrls)))
	return connectionInfo, nil
}

// GetConnectionInfo returns enhanced connection information for a service
// This is an additional method that provides the full enhanced service data
func (h *ConnectionHandler) GetConnectionInfo(serviceName string) (*models.EnhancedService, error) {
	h.logger.Log(fmt.Sprintf("Getting enhanced connection info for service: %s", serviceName))

	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	h.logger.Log(fmt.Sprintf("Retrieved enhanced connection info for service: %s", serviceName))
	return service, nil
}

// GetEnhancedServiceConnectionInfo returns enhanced connection information for a service
// This provides more detailed connection information using the enhanced service model
func (h *ConnectionHandler) GetEnhancedServiceConnectionInfo(serviceName string) (*models.EnhancedServiceConnectionInfo, error) {
	h.logger.Log(fmt.Sprintf("Getting enhanced connection info for service: %s", serviceName))

	// Get enhanced service data
	enhancedService, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Build enhanced connection info
	enhancedConnectionInfo := &models.EnhancedServiceConnectionInfo{
		ServiceName:           serviceName,
		Available:             true,
		Status:                enhancedService.Status,
		ContainerName:         enhancedService.ContainerName,
		ImageName:             enhancedService.ImageName,
		Type:                  enhancedService.Type,
		Username:              enhancedService.DefaultUser,
		Password:              enhancedService.DefaultPassword,
		ConnectionCommand:     enhancedService.ConnectionCmd,
		WebUrls:               enhancedService.WebUrls,
		ExposedPorts:          enhancedService.ExposedPorts,
		InternalPorts:         enhancedService.InternalPorts,
		DirectDependencies:    enhancedService.DirectDependencies,
		RecursiveDependencies: enhancedService.RecursiveDependencies,
		AllContainers:         enhancedService.AllContainers,
	}

	// Build connection strings and credentials
	enhancedConnectionInfo.ConnectionStrings = h.buildConnectionStrings(serviceName, enhancedService)
	enhancedConnectionInfo.Credentials = h.buildCredentials(serviceName, enhancedService)

	h.logger.Log(fmt.Sprintf("Retrieved enhanced connection info for service: %s", serviceName))
	return enhancedConnectionInfo, nil
}

// GetAllConnectionInfo returns enhanced connection information for all services
func (h *ConnectionHandler) GetAllConnectionInfo() ([]*models.EnhancedService, error) {
	h.logger.Log("Getting enhanced connection info for all services")

	services := h.serviceManager.ListServices()
	h.logger.Log(fmt.Sprintf("Retrieved enhanced connection info for %d services", len(services)))
	return services, nil
}

// GetServicePorts returns port information for a service
func (h *ConnectionHandler) GetServicePorts(serviceName string) ([]models.PortMapping, error) {
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service.ExposedPorts, nil
}

// GetServiceWebURLs returns web URLs for a service
func (h *ConnectionHandler) GetServiceWebURLs(serviceName string) ([]models.WebURL, error) {
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service.WebUrls, nil
}

// GetServiceDependencies returns dependency information for a service
func (h *ConnectionHandler) GetServiceDependencies(serviceName string) ([]string, []string, error) {
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return nil, nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service.DirectDependencies, service.RecursiveDependencies, nil
}

// RefreshConnectionInfo refreshes connection information from compose files
func (h *ConnectionHandler) RefreshConnectionInfo() error {
	h.logger.Log("Refreshing connection information from compose files")

	if err := h.serviceManager.LoadServices(); err != nil {
		h.logger.Log(fmt.Sprintf("Failed to refresh connection info: %v", err))
		return err
	}

	h.logger.Log("Successfully refreshed connection information")
	return nil
}

// Helper methods

// getCurrentContainers gets current running containers from the runtime
func (h *ConnectionHandler) getCurrentContainers() (map[string]bool, error) {
	containers := make(map[string]bool)

	// Use the runtime to get running containers
	// This is a simplified implementation - in a real scenario,
	// we would call the appropriate runtime method
	for serviceName := range core.Services {
		// Check if any containers for this service are running
		enhancedService, exists := h.serviceManager.GetService(serviceName)
		if !exists {
			continue
		}

		for _, containerName := range enhancedService.AllContainers {
			status, err := h.serviceManager.UpdateServiceStatus(serviceName)
			if err == nil && status == "running" {
				containers[containerName] = true
			}
		}
	}

	return containers, nil
}

// buildConnectionString builds a connection string for the service
func (h *ConnectionHandler) buildConnectionString(serviceName string, service core.Service, hostPort string) string {
	// Build a basic connection string based on service type
	switch service.Type {
	case "Database":
		switch serviceName {
		case "postgres":
			return fmt.Sprintf("postgresql://%s:%s@localhost:%s/postgres",
				service.DefaultUser, service.DefaultPassword, hostPort)
		case "mysql", "mariadb":
			return fmt.Sprintf("mysql://%s:%s@localhost:%s/",
				service.DefaultUser, service.DefaultPassword, hostPort)
		case "mongodb":
			return fmt.Sprintf("mongodb://%s:%s@localhost:%s/",
				service.DefaultUser, service.DefaultPassword, hostPort)
		case "redis":
			if service.DefaultPassword != "" {
				return fmt.Sprintf("redis://:%s@localhost:%s", service.DefaultPassword, hostPort)
			}
			return fmt.Sprintf("redis://localhost:%s", hostPort)
		case "elasticsearch":
			return fmt.Sprintf("http://%s:%s@localhost:%s",
				service.DefaultUser, service.DefaultPassword, hostPort)
		}
	case "Messaging":
		switch serviceName {
		case "kafka":
			return fmt.Sprintf("localhost:%s", hostPort)
		case "rabbitmq":
			return fmt.Sprintf("amqp://%s:%s@localhost:%s/",
				service.DefaultUser, service.DefaultPassword, hostPort)
		}
	}

	// Default to HTTP for services with web interfaces or basic TCP
	return fmt.Sprintf("localhost:%s", hostPort)
}

// buildConnectionStrings builds connection strings for all ports of an enhanced service
func (h *ConnectionHandler) buildConnectionStrings(serviceName string, enhancedService *models.EnhancedService) []models.ConnectionString {
	var connectionStrings []models.ConnectionString

	// Get core service for type information
	coreService, exists := core.Services[serviceName]
	if !exists {
		return connectionStrings
	}

	// Get runtime name for exec commands
	runtimeName := h.runtime.Name()

	// Add container exec command for database services
	if coreService.Type == "Database" && coreService.ConnectionCmd != "" {
		connectionStrings = append(connectionStrings, models.ConnectionString{
			Description:      "Container Connection",
			ConnectionString: fmt.Sprintf("%s exec -it %s sh -c \"%s\"", runtimeName, enhancedService.ContainerName, coreService.ConnectionCmd),
			Note:             "Run this command in your terminal to connect directly to the container",
		})
	}

	// Build connection strings for each exposed port
	for _, port := range enhancedService.ExposedPorts {
		var description string
		var connectionString string
		var note string

		// Create descriptive key based on port type and description
		if port.Description != "" {
			description = port.Description
		} else {
			switch port.Type {
			case core.PortTypeDatabase:
				description = "Database Connection"
			case core.PortTypeWebUI:
				description = "Web Interface"
			case core.PortTypeAdmin:
				description = "Admin Interface"
			case core.PortTypeAPI:
				description = "API Endpoint"
			case core.PortTypeMetrics:
				description = "Metrics Endpoint"
			default:
				description = fmt.Sprintf("Port %s", port.HostPort)
			}
		}

		// Build connection string based on port type
		switch port.Type {
		case core.PortTypeDatabase:
			connectionString = h.buildConnectionString(serviceName, coreService, port.HostPort)
			note = "Use this connection string to connect to the database from external tools"
		case core.PortTypeWebUI, core.PortTypeAdmin, core.PortTypeAPI:
			connectionString = fmt.Sprintf("http://localhost:%s", port.HostPort)
			note = "Open this URL in your browser"
		case core.PortTypeMetrics:
			connectionString = fmt.Sprintf("http://localhost:%s", port.HostPort)
			note = "Metrics endpoint for monitoring tools"
		default:
			connectionString = fmt.Sprintf("localhost:%s", port.HostPort)
			note = fmt.Sprintf("TCP connection on port %s", port.HostPort)
		}

		connectionStrings = append(connectionStrings, models.ConnectionString{
			Description:      description,
			ConnectionString: connectionString,
			Note:             note,
		})
	}

	return connectionStrings
}

// buildCredentials builds credential information for a service
func (h *ConnectionHandler) buildCredentials(serviceName string, enhancedService *models.EnhancedService) []models.Credential {
	var credentials []models.Credential

	// Get core service for credential information
	coreService, exists := core.Services[serviceName]
	if !exists {
		return credentials
	}

	// Add username credential if available
	if coreService.DefaultUser != "" {
		credentials = append(credentials, models.Credential{
			Description: "Default Username",
			Value:       coreService.DefaultUser,
			Note:        "Default username for authentication",
		})
	}

	// Add password credential if available
	if coreService.DefaultPassword != "" {
		credentials = append(credentials, models.Credential{
			Description: "Default Password",
			Value:       coreService.DefaultPassword,
			Note:        "Default password for authentication",
		})
	}

	return credentials
}
