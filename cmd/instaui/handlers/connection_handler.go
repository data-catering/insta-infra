package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// ConnectionHandler handles service connection and web UI operations
type ConnectionHandler struct {
	*BaseHandler
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(runtime container.Runtime, instaDir string) *ConnectionHandler {
	return &ConnectionHandler{
		BaseHandler: NewBaseHandler(runtime, instaDir),
	}
}

// GetServiceConnectionInfo returns connection details for a specific service
func (h *ConnectionHandler) GetServiceConnectionInfo(serviceName string) (*models.ServiceConnectionInfo, error) {
	// Get the service definition
	service, exists := core.Services[serviceName]
	if !exists {
		return nil, fmt.Errorf("unknown service: %s", serviceName)
	}

	// Get compose files
	composeFiles := h.getComposeFiles()

	// Get all running containers first
	runningContainers, err := h.getRunningContainers()
	if err != nil {
		return &models.ServiceConnectionInfo{
			ServiceName: serviceName,
			Available:   false,
			Error:       fmt.Sprintf("could not get running containers: %v", err),
		}, nil
	}

	// Check if service is running and get the actual container name
	containerName, err := h.getActualRunningContainerName(serviceName, composeFiles, runningContainers)
	if err != nil {
		return &models.ServiceConnectionInfo{
			ServiceName: serviceName,
			Available:   false,
			Error:       "service not running",
		}, nil
	}

	// Get port mappings
	portMappings, err := h.Runtime().GetPortMappings(containerName)
	if err != nil || len(portMappings) == 0 {
		return &models.ServiceConnectionInfo{
			ServiceName: serviceName,
			Available:   false,
			Error:       "service not running or no port mappings available",
		}, nil
	}

	// Use the first port mapping found
	var hostPort, containerPort string
	for cPort, hPort := range portMappings {
		// Extract container port number (e.g., "5432/tcp" -> "5432")
		parts := strings.Split(cPort, "/")
		if len(parts) > 0 {
			containerPort = parts[0]
		} else {
			containerPort = cPort
		}
		hostPort = hPort
		break // Use the first mapping
	}

	// Determine if service has web UI and build URL
	hasWebUI := h.hasWebUI(serviceName)
	var webURL string
	if hasWebUI && hostPort != "" {
		webURL = fmt.Sprintf("http://localhost:%s", hostPort)
	}

	// Build connection string for databases and other services
	connectionString := h.buildConnectionString(serviceName, service, hostPort)

	return &models.ServiceConnectionInfo{
		ServiceName:       serviceName,
		HasWebUI:          hasWebUI,
		WebURL:            webURL,
		HostPort:          hostPort,
		ContainerPort:     containerPort,
		Username:          service.DefaultUser,
		Password:          service.DefaultPassword,
		ConnectionCommand: service.ConnectionCmd,
		ConnectionString:  connectionString,
		Available:         true,
	}, nil
}

// OpenServiceInBrowser opens a service's web UI in the default browser (if available)
func (h *ConnectionHandler) OpenServiceInBrowser(serviceName string) error {
	connInfo, err := h.GetServiceConnectionInfo(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get connection info: %w", err)
	}

	if !connInfo.Available {
		return fmt.Errorf("service is not available: %s", connInfo.Error)
	}

	if !connInfo.HasWebUI || connInfo.WebURL == "" {
		return fmt.Errorf("service %s does not have a web UI", serviceName)
	}

	// Use the context to open URL in browser
	return h.openURL(connInfo.WebURL)
}

// hasWebUI determines if a service has a web UI that can be opened in browser
func (h *ConnectionHandler) hasWebUI(serviceName string) bool {
	webUIServices := map[string]bool{
		"grafana":       true,
		"superset":      true,
		"minio":         true,
		"keycloak":      true,
		"jenkins":       true,
		"kibana":        true,
		"metabase":      true,
		"airflow":       true,
		"jaeger":        true,
		"prometheus":    true,
		"amundsen":      true,
		"datahub":       true,
		"cvat":          true,
		"doccano":       true,
		"argilla":       true,
		"blazer":        true,
		"evidence":      true,
		"redash":        true,
		"sonarqube":     true,
		"vault":         true,
		"traefik":       true,
		"kong":          true,
		"flink":         true,
		"mlflow":        true,
		"temporal":      true,
		"prefect-data":  true,
		"ray":           true,
		"label-studio":  true,
		"opensearch":    true,
		"loki":          true,
		"elasticsearch": true, // Sometimes has web UI through plugins
	}
	return webUIServices[serviceName]
}

// buildConnectionString creates connection strings for different service types
func (h *ConnectionHandler) buildConnectionString(serviceName string, service core.Service, hostPort string) string {
	if hostPort == "" {
		return ""
	}

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
		case "cassandra":
			return fmt.Sprintf("cassandra://localhost:%s", hostPort)
		case "neo4j":
			return fmt.Sprintf("bolt://localhost:%s", hostPort)
		case "influxdb":
			return fmt.Sprintf("http://localhost:%s", hostPort)
		}
	case "Messaging":
		switch serviceName {
		case "kafka":
			return fmt.Sprintf("localhost:%s", hostPort)
		case "rabbitmq":
			return fmt.Sprintf("amqp://%s:%s@localhost:%s/",
				service.DefaultUser, service.DefaultPassword, hostPort)
		case "nats":
			return fmt.Sprintf("nats://localhost:%s", hostPort)
		case "pulsar":
			return fmt.Sprintf("pulsar://localhost:%s", hostPort)
		}
	}

	// Default to HTTP for web services
	if h.hasWebUI(serviceName) {
		return fmt.Sprintf("http://localhost:%s", hostPort)
	}

	// Generic TCP connection
	return fmt.Sprintf("localhost:%s", hostPort)
}

// openURL opens a URL in the default browser using cross-platform commands
func (h *ConnectionHandler) openURL(url string) error {
	var cmd string
	var args []string

	// Detect OS and use appropriate command
	switch {
	case os.Getenv("WSL_DISTRO_NAME") != "": // WSL
		cmd = "cmd.exe"
		args = []string{"/c", "start", url}
	default: // macOS, Linux, Windows
		switch {
		case fileExists("/usr/bin/xdg-open"): // Linux
			cmd = "xdg-open"
			args = []string{url}
		case fileExists("/usr/bin/open"): // macOS
			cmd = "open"
			args = []string{url}
		default: // Windows or fallback
			cmd = "rundll32"
			args = []string{"url.dll,FileProtocolHandler", url}
		}
	}

	// Execute the command using exec package
	execCmd := exec.Command(cmd, args...)
	return execCmd.Start()
}





// fileExists checks if a file exists
