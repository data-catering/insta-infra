package handlers

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// GraphHandler handles dependency graph visualization operations
type GraphHandler struct {
	containerRuntime  container.Runtime
	instaDir          string
	dependencyHandler *DependencyHandler
}

// NewGraphHandler creates a new graph handler
func NewGraphHandler(runtime container.Runtime, instaDir string, depHandler *DependencyHandler) *GraphHandler {
	return &GraphHandler{
		containerRuntime:  runtime,
		instaDir:          instaDir,
		dependencyHandler: depHandler,
	}
}

// GetDependencyGraph returns the complete dependency graph for all services
func (h *GraphHandler) GetDependencyGraph() (*models.DependencyGraph, error) {
	graph := &models.DependencyGraph{
		Nodes: make([]models.GraphNode, 0),
		Edges: make([]models.GraphEdge, 0),
	}

	// Get compose files to parse all services (including init containers)
	composeFiles := h.getComposeFiles()

	// Parse compose configuration to get all services and their relationships
	composeConfig, err := h.getComposeConfig(composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compose config: %w", err)
	}

	// Create nodes for all services in compose file
	nodeCount := 0
	serviceNodes := make(map[string]models.GraphNode)

	for serviceName, serviceConfig := range composeConfig.Services {
		// Get container name (use explicit container_name or generate default)
		containerName := h.getContainerNameFromConfig(serviceName, serviceConfig)

		// Get service status using the new method that handles auxiliary services
		status := h.GetServiceStatus(serviceName, composeFiles)

		// Determine service type
		serviceType := h.guessServiceType(serviceName)

		// Calculate position
		position := h.calculateNodePosition(serviceName, nodeCount)

		dependencies := h.extractDependencyNames(serviceConfig)
		if dependencies == nil {
			dependencies = make([]string, 0)
		}

		node := models.GraphNode{
			ID:     serviceName,
			Label:  h.getDisplayLabel(containerName), // Show user-friendly label
			Type:   "service",
			Status: status,
			Position: models.NodePosition{
				X: position.X,
				Y: position.Y,
			},
			Data: models.NodeData{
				ServiceName:  serviceName,
				Type:         serviceType,
				Status:       status,
				Health:       h.getHealthFromStatus(status),
				Dependencies: dependencies,
				Color:        h.getNodeColor(status),
			},
		}

		graph.Nodes = append(graph.Nodes, node)
		serviceNodes[serviceName] = node
		nodeCount++
	}

	// Create edges for dependencies
	edgeCount := 0
	for serviceName, serviceConfig := range composeConfig.Services {
		dependencies := h.extractDependencyNames(serviceConfig)

		for _, dependency := range dependencies {
			// Check if dependency service exists in our nodes
			if _, exists := serviceNodes[dependency]; exists {
				sourceNode := serviceNodes[dependency]
				targetNode := serviceNodes[serviceName]

				edge := models.GraphEdge{
					ID:     fmt.Sprintf("edge_%d", edgeCount),
					Source: serviceName,
					Target: dependency,
					Type:   "smoothstep",
					Data: models.EdgeData{
						Label:    "depends on",
						Type:     "dependency",
						Animated: h.shouldAnimateEdge(sourceNode.Status, targetNode.Status),
						Color:    h.getEdgeColor(sourceNode.Status, targetNode.Status),
					},
				}

				graph.Edges = append(graph.Edges, edge)
				edgeCount++
			}
		}
	}

	return graph, nil
}

// GetServiceDependencyGraph returns a focused dependency graph for a specific service
func (h *GraphHandler) GetServiceDependencyGraph(serviceName string) (*models.DependencyGraph, error) {
	composeFiles := h.getComposeFiles()

	graph := &models.DependencyGraph{
		Nodes: make([]models.GraphNode, 0),
		Edges: make([]models.GraphEdge, 0),
	}

	// Parse compose configuration
	composeConfig, err := h.getComposeConfig(composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse compose config: %w", err)
	}

	// Check if target service exists
	if _, exists := composeConfig.Services[serviceName]; !exists {
		return nil, fmt.Errorf("service '%s' not found", serviceName)
	}

	// Collect all services to include in the focused graph
	servicesToInclude := make(map[string]bool)
	servicesToInclude[serviceName] = true

	// Add all dependencies recursively
	h.addDependenciesRecursiveFromConfig(serviceName, composeConfig, servicesToInclude)

	// Add all dependents recursively
	h.addDependentsRecursiveFromConfig(serviceName, composeConfig, servicesToInclude)

	// Create nodes for included services with better positioning
	nodePositions := h.calculateFocusedNodePositions(servicesToInclude, serviceName)

	for currentServiceName, include := range servicesToInclude {
		if !include {
			continue
		}

		serviceConfig := composeConfig.Services[currentServiceName]
		containerName := h.getContainerNameFromConfig(currentServiceName, serviceConfig)
		status := h.GetServiceStatus(currentServiceName, composeFiles)

		serviceType := h.guessServiceType(currentServiceName)
		position := nodePositions[currentServiceName]

		dependencies := h.extractDependencyNames(serviceConfig)
		if dependencies == nil {
			dependencies = make([]string, 0)
		}

		node := models.GraphNode{
			ID:       currentServiceName,
			Label:    h.getDisplayLabel(containerName), // Show user-friendly label
			Type:     "service",
			Status:   status,
			Position: position,
			Data: models.NodeData{
				ServiceName:  currentServiceName,
				Type:         serviceType,
				Status:       status,
				Health:       h.getHealthFromStatus(status),
				Dependencies: dependencies,
				Color:        h.getNodeColor(status),
			},
		}

		graph.Nodes = append(graph.Nodes, node)
	}

	// Create edges for dependencies within the included services
	edgeCount := 0
	for currentServiceName, include := range servicesToInclude {
		if !include {
			continue
		}

		serviceConfig := composeConfig.Services[currentServiceName]
		dependencies := h.extractDependencyNames(serviceConfig)

		for _, dependency := range dependencies {
			if servicesToInclude[dependency] {
				// Find source and target nodes for status information
				var sourceStatus, targetStatus string
				for _, node := range graph.Nodes {
					if node.ID == dependency {
						sourceStatus = node.Status
					}
					if node.ID == currentServiceName {
						targetStatus = node.Status
					}
				}

				edge := models.GraphEdge{
					ID:     fmt.Sprintf("edge_%d", edgeCount),
					Source: currentServiceName,
					Target: dependency,
					Type:   "smoothstep",
					Data: models.EdgeData{
						Label:    "depends on",
						Type:     "dependency",
						Animated: h.shouldAnimateEdge(sourceStatus, targetStatus),
						Color:    h.getEdgeColor(sourceStatus, targetStatus),
					},
				}

				graph.Edges = append(graph.Edges, edge)
				edgeCount++
			}
		}
	}

	return graph, nil
}

// Helper methods for enhanced graph functionality

// getComposeConfig creates a simplified config showing only user-facing services
func (h *GraphHandler) getComposeConfig(composeFiles []string) (*container.ComposeConfig, error) {
	config := &container.ComposeConfig{
		Services: make(map[string]container.ComposeService),
	}

	// Define auxiliary services that should be hidden from the main graph
	auxiliaryServices := map[string]string{
		"postgres-server":   "postgres",   // postgres-server is internal, show postgres
		"cassandra-server":  "cassandra",  // cassandra-server is internal, show cassandra
		"clickhouse-server": "clickhouse", // clickhouse-server is internal, show clickhouse
	}

	// Add only user-facing services from core.Services
	for _, service := range core.Services {
		// Skip auxiliary services that should be hidden
		if _, isAuxiliary := auxiliaryServices[service.Name]; isAuxiliary {
			continue
		}

		// Get dependencies, but map them to user-facing services
		dependencies, err := h.containerRuntime.GetDependencies(service.Name, composeFiles)
		if err != nil {
			dependencies = []string{} // Continue with empty dependencies if error
		}

		// Map dependencies to user-facing services and container names
		userFacingDeps := h.mapToUserFacingServices(dependencies, auxiliaryServices, service.Name)

		// Get the actual running container name for this service
		containerName := h.getActualContainerName(service.Name)

		// Build depends_on map with user-facing dependencies
		dependsOn := make(map[string]struct {
			Condition string `json:"condition"`
		})
		for _, dep := range userFacingDeps {
			dependsOn[dep] = struct {
				Condition string `json:"condition"`
			}{Condition: "service_started"}
		}

		config.Services[service.Name] = container.ComposeService{
			DependsOn:     dependsOn,
			ContainerName: containerName,
		}
	}

	return config, nil
}

// getContainerNameFromConfig gets the container name from service config
func (h *GraphHandler) getContainerNameFromConfig(serviceName string, serviceConfig container.ComposeService) string {
	if serviceConfig.ContainerName != "" {
		return serviceConfig.ContainerName
	}
	// Generate default container name based on compose naming convention
	return serviceName
}

// getDisplayLabel returns the user-friendly label for display in the graph
// For the simplified approach, we just show the actual container name
func (h *GraphHandler) getDisplayLabel(containerName string) string {
	// For the simplified approach, just show the container name
	return containerName
}

// mapToUserFacingServices maps auxiliary service dependencies to user-facing services
// This filters out internal dependencies to avoid self-references
func (h *GraphHandler) mapToUserFacingServices(dependencies []string, auxiliaryServices map[string]string, currentService string) []string {
	var userFacingDeps []string

	for _, dep := range dependencies {
		// If this dependency is an auxiliary service, map it to the user-facing service
		if userFacingService, isAuxiliary := auxiliaryServices[dep]; isAuxiliary {
			// Avoid self-dependencies: if the mapped service is the same as current service, skip it
			if userFacingService != currentService {
				userFacingDeps = append(userFacingDeps, userFacingService)
			}
		} else {
			// Otherwise, keep the dependency as-is (unless it's the same as current service)
			if dep != currentService {
				userFacingDeps = append(userFacingDeps, dep)
			}
		}
	}

	return userFacingDeps
}

// getActualContainerName returns the actual running container name for a service
// This handles the auxiliary service mapping to show the right container name
func (h *GraphHandler) getActualContainerName(serviceName string) string {
	// For auxiliary services, we want to show the actual running container name
	auxiliaryToMainContainer := map[string]string{
		"postgres":   "postgres",   // postgres service -> postgres container (from postgres-server)
		"cassandra":  "cassandra",  // cassandra service -> cassandra container (from cassandra-server)
		"clickhouse": "clickhouse", // clickhouse service -> clickhouse container (from clickhouse-server)
	}

	if containerName, isAuxiliary := auxiliaryToMainContainer[serviceName]; isAuxiliary {
		return containerName
	}

	// For regular services, try to get the container name from the runtime
	composeFiles := h.getComposeFiles()
	containerName, err := h.containerRuntime.GetContainerName(serviceName, composeFiles)
	if err != nil {
		// Fallback to service name
		return serviceName
	}

	return containerName
}

// getContainerStatus gets the actual status of a container
func (h *GraphHandler) getContainerStatus(containerName string) string {
	// Use the container runtime to get the actual container status
	status, err := h.containerRuntime.GetContainerStatus(containerName)
	if err != nil {
		return "not_found"
	}
	return status
}

// GetServiceStatus gets the status of a service using the simplified approach
func (h *GraphHandler) GetServiceStatus(serviceName string, composeFiles []string) string {
	// Use the dependency handler's status checking which includes stopped service tracking
	if h.dependencyHandler != nil {
		status, err := h.dependencyHandler.GetServiceStatusInternal(serviceName, composeFiles)
		if err == nil {
			return status
		}
	}

	// Fallback to direct checking
	// Get the actual container name that should be running for this service
	containerName := h.getActualContainerName(serviceName)

	// Get the container status
	status := h.getContainerStatus(containerName)

	// For auxiliary services like postgres, if the main container is running, show as running
	if status == "running" {
		return "running"
	}

	// If container is not found, return stopped
	if status == "not_found" {
		return "stopped"
	}

	// Return the actual status (completed, error, etc.)
	return status
}

// extractDependencyNames extracts dependency names from service config
func (h *GraphHandler) extractDependencyNames(serviceConfig container.ComposeService) []string {
	dependencies := make([]string, 0)
	for depName := range serviceConfig.DependsOn {
		dependencies = append(dependencies, depName)
	}
	return dependencies
}

// getHealthFromStatus converts status to health
func (h *GraphHandler) getHealthFromStatus(status string) string {
	switch status {
	case "running":
		return "healthy"
	case "completed":
		return "completed"
	case "stopped":
		return "stopped"
	case "error":
		return "unhealthy"
	case "not_found":
		return "not_found"
	case "paused":
		return "paused"
	case "restarting":
		return "starting"
	default:
		return "unknown"
	}
}

// addDependenciesRecursiveFromConfig adds dependencies recursively from compose config
func (h *GraphHandler) addDependenciesRecursiveFromConfig(serviceName string, composeConfig *container.ComposeConfig, servicesToInclude map[string]bool) {
	serviceConfig, exists := composeConfig.Services[serviceName]
	if !exists {
		return
	}

	for depName := range serviceConfig.DependsOn {
		if depName == "" || servicesToInclude[depName] {
			continue
		}
		servicesToInclude[depName] = true
		h.addDependenciesRecursiveFromConfig(depName, composeConfig, servicesToInclude)
	}
}

// addDependentsRecursiveFromConfig adds dependents recursively from compose config
func (h *GraphHandler) addDependentsRecursiveFromConfig(serviceName string, composeConfig *container.ComposeConfig, servicesToInclude map[string]bool) {
	for currentServiceName, serviceConfig := range composeConfig.Services {
		if currentServiceName == "" || servicesToInclude[currentServiceName] {
			continue
		}

		// Check if this service depends on our target service
		for depName := range serviceConfig.DependsOn {
			if depName == serviceName {
				servicesToInclude[currentServiceName] = true
				h.addDependentsRecursiveFromConfig(currentServiceName, composeConfig, servicesToInclude)
				break
			}
		}
	}
}

// getNodeColor determines the node color based on service status and type
func (h *GraphHandler) getNodeColor(status string) string {
	switch status {
		case "running":
			return "#10b981" // green
		case "completed":
			return "#22c55e" // light green for completed init containers
		case "stopped":
			return "#6b7280" // gray
		case "error":
			return "#ef4444" // red
		case "not_found":
			return "#9ca3af" // light gray for not found
		case "paused":
			return "#f59e0b" // amber for paused
		case "restarting":
			return "#3b82f6" // blue for restarting
		default:
			return "#94a3b8" // light gray
	}
}

// getEdgeColor determines the edge color based on source and target status
func (h *GraphHandler) getEdgeColor(sourceStatus, targetStatus string) string {
	if sourceStatus == "running" && targetStatus == "running" {
		return "#10b981" // green - healthy connection
	} else if sourceStatus == "error" || targetStatus == "error" {
		return "#ef4444" // red - problematic connection
	} else if sourceStatus == "stopped" || targetStatus == "stopped" {
		return "#6b7280" // gray - inactive connection
	}
	return "#94a3b8" // default gray
}

// shouldAnimateEdge determines if an edge should be animated based on status
func (h *GraphHandler) shouldAnimateEdge(sourceStatus, targetStatus string) bool {
	// Animate edges where services are running
	return sourceStatus == "running" && targetStatus == "running"
}

// calculateNodePosition calculates initial position for a node using a simple algorithm
func (h *GraphHandler) calculateNodePosition(serviceName string, index int) models.NodePosition {
	// Simple circular arrangement algorithm
	const radius = 200
	const centerX = 400
	const centerY = 300

	// Convert string to number for consistent positioning
	hash := 0
	for _, char := range serviceName {
		hash = hash*31 + int(char)
	}

	// Use hash to determine angle for consistent positioning
	angle := float64(hash%360) * 3.14159 / 180

	// Add some variation based on index to avoid overlaps
	radiusVariation := radius + float64(index%3)*50

	x := centerX + radiusVariation*math.Cos(angle)
	y := centerY + radiusVariation*math.Sin(angle)

	return models.NodePosition{X: x, Y: y}
}

// calculateFocusedNodePositions creates a better layout for focused dependency graphs
func (h *GraphHandler) calculateFocusedNodePositions(servicesToInclude map[string]bool, targetService string) map[string]models.NodePosition {
	positions := make(map[string]models.NodePosition)

	// Count services to determine layout
	serviceCount := 0
	for _, include := range servicesToInclude {
		if include {
			serviceCount++
		}
	}

	// Place target service in the center
	positions[targetService] = models.NodePosition{X: 400, Y: 300}

	// Arrange other services in a circle around the target
	radius := float64(150)
	if serviceCount > 8 {
		radius = float64(200)
	}

	angle := 0.0
	angleStep := 2 * 3.14159 / float64(serviceCount-1) // -1 because target is already placed

	for currentServiceName, include := range servicesToInclude {
		if !include || currentServiceName == targetService || currentServiceName == "" {
			continue
		}

		x := 400 + radius*math.Cos(angle)
		y := 300 + radius*math.Sin(angle)

		positions[currentServiceName] = models.NodePosition{X: x, Y: y}

		angle += angleStep
	}

	return positions
}

// guessServiceType attempts to determine service type based on service name patterns
func (h *GraphHandler) guessServiceType(serviceName string) string {
	name := strings.ToLower(serviceName)

	// Database patterns
	if strings.Contains(name, "postgres") || strings.Contains(name, "mysql") || strings.Contains(name, "redis") ||
		strings.Contains(name, "mongo") || strings.Contains(name, "elasticsearch") || strings.Contains(name, "cassandra") ||
		strings.Contains(name, "influx") || strings.Contains(name, "neo4j") || strings.Contains(name, "mariadb") {
		return "Database"
	}

	// Messaging patterns
	if strings.Contains(name, "kafka") || strings.Contains(name, "rabbit") || strings.Contains(name, "nats") ||
		strings.Contains(name, "pulsar") || strings.Contains(name, "activemq") {
		return "Messaging"
	}

	// Monitoring patterns
	if strings.Contains(name, "prometheus") || strings.Contains(name, "grafana") || strings.Contains(name, "jaeger") ||
		strings.Contains(name, "elastic") || strings.Contains(name, "kibana") {
		return "Monitoring"
	}

	// Job/Task patterns
	if strings.Contains(name, "airflow") || strings.Contains(name, "celery") || strings.Contains(name, "worker") ||
		strings.Contains(name, "scheduler") || strings.Contains(name, "init") {
		return "Job Orchestrator"
	}

	// Default fallback
	return "Service"
}

// getAllServiceDetails is a simplified version that uses the service handler logic
func (h *GraphHandler) getAllServiceDetails() ([]models.ServiceDetailInfo, error) {
	// This would typically delegate to the service handler
	// For now, we'll implement a simplified version
	var serviceList []models.ServiceDetailInfo
	composeFiles := h.getComposeFiles()

	for _, service := range core.Services {
		detail := models.ServiceDetailInfo{
			Name: service.Name,
			Type: service.Type,
		}

		// Get status
		status, err := h.dependencyHandler.GetServiceStatusInternal(service.Name, composeFiles)
		detail.Status = status
		if err != nil {
			detail.StatusError = err.Error()
		}

		// Get dependencies
		deps, err := h.containerRuntime.GetAllDependenciesRecursive(service.Name, composeFiles)
		detail.Dependencies = deps
		if err != nil {
			detail.DependenciesError = err.Error()
		}

		serviceList = append(serviceList, detail)
	}

	return serviceList, nil
}

// getComposeFiles returns the list of compose files to use
func (h *GraphHandler) getComposeFiles() []string {
	baseComposeFile := filepath.Join(h.instaDir, "docker-compose.yaml")
	persistComposeFile := filepath.Join(h.instaDir, "docker-compose-persist.yaml")

	composeFiles := []string{baseComposeFile}
	if _, err := os.Stat(persistComposeFile); err == nil {
		composeFiles = append(composeFiles, persistComposeFile)
	}

	return composeFiles
}

// GetServiceContainerDetails returns detailed container information for a service
// This is used for drill-down functionality to show internal containers
func (h *GraphHandler) GetServiceContainerDetails(serviceName string) (*models.ServiceContainerDetails, error) {
	composeFiles := h.getComposeFiles()

	details := &models.ServiceContainerDetails{
		ServiceName: serviceName,
		Containers:  make([]models.ContainerInfo, 0),
	}

	// For auxiliary services like postgres, show both the main and auxiliary containers
	auxiliaryServices := map[string][]string{
		"postgres":   {"postgres-server", "postgres"},
		"cassandra":  {"cassandra-server", "cassandra"},
		"clickhouse": {"clickhouse-server", "clickhouse"},
	}

	if containerServices, isAuxiliary := auxiliaryServices[serviceName]; isAuxiliary {
		// For auxiliary services, show all related containers
		for _, containerService := range containerServices {
			containerName, err := h.containerRuntime.GetContainerName(containerService, composeFiles)
			if err != nil {
				continue // Skip if we can't get container name
			}

			status := h.getContainerStatus(containerName)

			// Determine the role of this container
			role := "main"
			if containerService != serviceName+"-server" {
				role = "auxiliary"
			}

			containerInfo := models.ContainerInfo{
				Name:        containerName,
				ServiceName: containerService,
				Status:      status,
				Role:        role,
				Description: h.getContainerDescription(containerService, role),
			}

			details.Containers = append(details.Containers, containerInfo)
		}
	} else {
		// For regular services, just show the single container
		containerName, err := h.containerRuntime.GetContainerName(serviceName, composeFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to get container name for service %s: %w", serviceName, err)
		}

		status := h.getContainerStatus(containerName)

		containerInfo := models.ContainerInfo{
			Name:        containerName,
			ServiceName: serviceName,
			Status:      status,
			Role:        "main",
			Description: h.getContainerDescription(serviceName, "main"),
		}

		details.Containers = append(details.Containers, containerInfo)
	}

	return details, nil
}

// getContainerDescription returns a user-friendly description for a container
func (h *GraphHandler) getContainerDescription(serviceName string, role string) string {
	descriptions := map[string]map[string]string{
		"postgres-server": {
			"main": "Main PostgreSQL database server",
		},
		"postgres": {
			"auxiliary": "Data initialization container (completed after setup)",
		},
		"cassandra-server": {
			"main": "Main Cassandra database server",
		},
		"cassandra": {
			"auxiliary": "Data initialization container (completed after setup)",
		},
		"clickhouse-server": {
			"main": "Main ClickHouse database server",
		},
		"clickhouse": {
			"auxiliary": "Data initialization container (completed after setup)",
		},
	}

	if serviceDescriptions, exists := descriptions[serviceName]; exists {
		if description, exists := serviceDescriptions[role]; exists {
			return description
		}
	}

	// Default descriptions based on role
	switch role {
	case "main":
		return fmt.Sprintf("Main %s service container", serviceName)
	case "auxiliary":
		return fmt.Sprintf("Auxiliary container for %s", serviceName)
	default:
		return fmt.Sprintf("%s container", serviceName)
	}
}
