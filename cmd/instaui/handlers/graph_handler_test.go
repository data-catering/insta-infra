package handlers

import (
	"errors"
	"testing"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

func TestGraphHandler_NewGraphHandler(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	instaDir := "/test/insta"
	depHandler := NewDependencyHandler(mockRuntime, instaDir, nil)

	handler := NewGraphHandler(mockRuntime, instaDir, depHandler)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	if handler.containerRuntime != mockRuntime {
		t.Errorf("Expected containerRuntime to be %v, got %v", mockRuntime, handler.containerRuntime)
	}
	if handler.instaDir != instaDir {
		t.Errorf("Expected instaDir to be %s, got %s", instaDir, handler.instaDir)
	}
	if handler.dependencyHandler != depHandler {
		t.Errorf("Expected dependencyHandler to be %v, got %v", depHandler, handler.dependencyHandler)
	}
}

func TestGraphHandler_GetDependencyGraph_Success(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getDependenciesFunc: func(serviceName string, composeFiles []string) ([]string, error) {
			// Mock dependencies for the compose config
			switch serviceName {
			case "grafana":
				return []string{"postgres"}, nil
			default:
				return []string{}, nil
			}
		},
		getContainerNameFunc: func(serviceName string, composeFiles []string) (string, error) {
			return serviceName, nil // Use the new simple naming
		},
		getContainerStatusFunc: func(containerName string) (string, error) {
			return "running", nil
		},
	}
	serviceHandler := NewServiceHandler(mockRuntime, "/test/insta")
	depHandler := NewDependencyHandler(mockRuntime, "/test/insta", serviceHandler)
	handler := NewGraphHandler(mockRuntime, "/test/insta", depHandler)

	graph, err := handler.GetDependencyGraph()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if graph == nil {
		t.Fatal("Expected graph to be returned, got nil")
	}
	if len(graph.Nodes) == 0 {
		t.Error("Expected nodes to be created")
	}

	// The graph should contain all services from core.Services
	// Check that we have some nodes at least
	if len(graph.Nodes) == 0 {
		t.Error("Expected at least one node to be created")
	}
}

func TestGraphHandler_GetDependencyGraph_ServiceDetailsError(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getContainerNameFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("container runtime not available")
		},
		getAllDependenciesRecursiveFunc: func(serviceName string, composeFiles []string) ([]string, error) {
			return nil, errors.New("dependencies error")
		},
	}
	serviceHandler := NewServiceHandler(mockRuntime, "/test/insta")
	depHandler := NewDependencyHandler(mockRuntime, "/test/insta", serviceHandler)
	handler := NewGraphHandler(mockRuntime, "/test/insta", depHandler)

	graph, err := handler.GetDependencyGraph()

	// getAllServiceDetails doesn't return an error, it captures individual errors in fields
	// So we should get a successful graph but with error statuses
	if err != nil {
		t.Fatalf("Expected no error from GetDependencyGraph, got %v", err)
	}
	if graph == nil {
		t.Fatal("Expected graph to be returned, got nil")
	}
	if len(graph.Nodes) == 0 {
		t.Error("Expected nodes to be created even with service errors")
	}
}

func TestGraphHandler_GetServiceDependencyGraph_Success(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getDependenciesFunc: func(serviceName string, composeFiles []string) ([]string, error) {
			switch serviceName {
			case "grafana":
				return []string{"postgres"}, nil
			default:
				return []string{}, nil
			}
		},
		getContainerNameFunc: func(serviceName string, composeFiles []string) (string, error) {
			return serviceName, nil
		},
		getContainerStatusFunc: func(containerName string) (string, error) {
			return "running", nil
		},
	}
	serviceHandler := NewServiceHandler(mockRuntime, "/test/insta")
	depHandler := NewDependencyHandler(mockRuntime, "/test/insta", serviceHandler)
	handler := NewGraphHandler(mockRuntime, "/test/insta", depHandler)

	graph, err := handler.GetServiceDependencyGraph("grafana")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if graph == nil {
		t.Fatal("Expected graph to be returned, got nil")
	}
	if len(graph.Nodes) < 1 {
		t.Errorf("Expected at least 1 node (grafana), got %d", len(graph.Nodes))
	}

	// Verify the target service is included
	foundTarget := false
	for _, node := range graph.Nodes {
		if node.ID == "grafana" {
			foundTarget = true
			break
		}
	}
	if !foundTarget {
		t.Error("Expected target service 'grafana' to be in focused graph")
	}
}

func TestGraphHandler_GetServiceDependencyGraph_ServiceNotFound(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getAllDependenciesRecursiveFunc: func(serviceName string, composeFiles []string) ([]string, error) {
			return []string{}, nil
		},
		getContainerNameFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		},
		getPortMappingsFunc: func(containerName string) (map[string]string, error) {
			return map[string]string{"80/tcp": "80"}, nil
		},
	}
	serviceHandler := NewServiceHandler(mockRuntime, "/test/insta")
	depHandler := NewDependencyHandler(mockRuntime, "/test/insta", serviceHandler)
	handler := NewGraphHandler(mockRuntime, "/test/insta", depHandler)

	graph, err := handler.GetServiceDependencyGraph("nonexistent")

	if err == nil {
		t.Fatal("Expected error for non-existent service, got nil")
	}
	if graph != nil {
		t.Errorf("Expected graph to be nil, got %v", graph)
	}
	if !contains(err.Error(), "service 'nonexistent' not found") {
		t.Errorf("Expected error to contain service not found message, got %s", err.Error())
	}
}

func TestGraphHandler_GetServiceDependencyGraph_ServiceDetailsError(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getContainerNameFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("container runtime not available")
		},
		getAllDependenciesRecursiveFunc: func(serviceName string, composeFiles []string) ([]string, error) {
			return nil, errors.New("dependencies error")
		},
	}
	serviceHandler := NewServiceHandler(mockRuntime, "/test/insta")
	depHandler := NewDependencyHandler(mockRuntime, "/test/insta", serviceHandler)
	handler := NewGraphHandler(mockRuntime, "/test/insta", depHandler)

	graph, err := handler.GetServiceDependencyGraph("postgres")

	// getAllServiceDetails doesn't return an error, it captures individual errors in fields
	// So we should get a successful graph (as postgres exists in core.Services)
	if err != nil {
		t.Fatalf("Expected no error from GetServiceDependencyGraph, got %v", err)
	}
	if graph == nil {
		t.Fatal("Expected graph to be returned, got nil")
	}
	if len(graph.Nodes) == 0 {
		t.Error("Expected at least the target service node to be created")
	}
}

func TestGraphHandler_getNodeColor(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	testCases := []struct {
		status      string
		expected    string
	}{
		{"running", "#10b981"}, // green
		{"stopped", "#6b7280"},    // gray
		{"error", "#ef4444"},  // red
		{"unknown", "#94a3b8"},  // light gray
		{"invalid", "#94a3b8"},  // default light gray
	}

	for _, tc := range testCases {
		result := handler.getNodeColor(tc.status)
		if result != tc.expected {
			t.Errorf("getNodeColor(%s): expected %s, got %s", tc.status, tc.expected, result)
		}
	}
}

func TestGraphHandler_getEdgeColor(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	testCases := []struct {
		sourceStatus string
		targetStatus string
		expected     string
	}{
		{"running", "running", "#10b981"}, // green - healthy connection
		{"error", "running", "#ef4444"},   // red - problematic connection
		{"running", "error", "#ef4444"},   // red - problematic connection
		{"stopped", "running", "#6b7280"}, // gray - inactive connection
		{"running", "stopped", "#6b7280"}, // gray - inactive connection
		{"unknown", "unknown", "#94a3b8"}, // default gray
	}

	for _, tc := range testCases {
		result := handler.getEdgeColor(tc.sourceStatus, tc.targetStatus)
		if result != tc.expected {
			t.Errorf("getEdgeColor(%s, %s): expected %s, got %s", tc.sourceStatus, tc.targetStatus, tc.expected, result)
		}
	}
}

func TestGraphHandler_shouldAnimateEdge(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	testCases := []struct {
		sourceStatus string
		targetStatus string
		expected     bool
	}{
		{"running", "running", true},  // Should animate when both running
		{"running", "stopped", false}, // Should not animate when target stopped
		{"stopped", "running", false}, // Should not animate when source stopped
		{"error", "error", false},     // Should not animate when both error
		{"unknown", "unknown", false}, // Should not animate when both unknown
	}

	for _, tc := range testCases {
		result := handler.shouldAnimateEdge(tc.sourceStatus, tc.targetStatus)
		if result != tc.expected {
			t.Errorf("shouldAnimateEdge(%s, %s): expected %t, got %t", tc.sourceStatus, tc.targetStatus, tc.expected, result)
		}
	}
}

func TestGraphHandler_calculateNodePosition(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	position1 := handler.calculateNodePosition("postgres", 0)
	position2 := handler.calculateNodePosition("redis", 1)

	// Check that positions are different
	if position1.X == position2.X && position1.Y == position2.Y {
		t.Error("Expected different positions for different services")
	}

	// Check that positions are consistently calculated
	position1Again := handler.calculateNodePosition("postgres", 0)
	if position1.X != position1Again.X || position1.Y != position1Again.Y {
		t.Error("Expected consistent positioning for same service and index")
	}

	// Check that coordinates are within reasonable bounds
	if position1.X < 0 || position1.Y < 0 || position1.X > 1000 || position1.Y > 1000 {
		t.Errorf("Position coordinates seem unreasonable: X=%f, Y=%f", position1.X, position1.Y)
	}
}

func TestGraphHandler_calculateFocusedNodePositions(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	servicesToInclude := map[string]bool{
		"grafana":  true,
		"postgres": true,
		"redis":    true,
	}

	positions := handler.calculateFocusedNodePositions(servicesToInclude, "grafana")

	// Check that target service is in center
	grafanaPos := positions["grafana"]
	if grafanaPos.X != 400 || grafanaPos.Y != 300 {
		t.Errorf("Expected target service at center (400, 300), got (%f, %f)", grafanaPos.X, grafanaPos.Y)
	}

	// Check that other services have positions
	if _, exists := positions["postgres"]; !exists {
		t.Error("Expected postgres to have a position")
	}
	if _, exists := positions["redis"]; !exists {
		t.Error("Expected redis to have a position")
	}

	// Check that positions are different
	postgresPos := positions["postgres"]
	redisPos := positions["redis"]
	if postgresPos.X == redisPos.X && postgresPos.Y == redisPos.Y {
		t.Error("Expected different positions for different services")
	}
}

func TestGraphHandler_guessServiceType(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	testCases := []struct {
		serviceName string
		expected    string
	}{
		{"postgres", "Database"},
		{"mysql", "Database"},
		{"redis", "Database"},
		{"mongodb", "Database"},
		{"elasticsearch", "Database"},
		{"kafka", "Messaging"},
		{"rabbitmq", "Messaging"},
		{"activemq", "Messaging"},
		{"prometheus", "Monitoring"},
		{"grafana", "Monitoring"},
		{"jaeger", "Monitoring"},
		{"airflow", "Job Orchestrator"},
		{"celery", "Job Orchestrator"},
		{"worker", "Job Orchestrator"},
		{"unknown-service", "Service"},
		{"", "Service"},
	}

	for _, tc := range testCases {
		result := handler.guessServiceType(tc.serviceName)
		if result != tc.expected {
			t.Errorf("guessServiceType(%s): expected %s, got %s", tc.serviceName, tc.expected, result)
		}
	}
}

func TestGraphHandler_addDependenciesRecursive(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	// Create test data using compose config format
	composeConfig := &container.ComposeConfig{
		Services: map[string]container.ComposeService{
			"grafana": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{
					"postgres": {Condition: "service_started"},
				},
			},
			"postgres": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{},
			},
			"superset": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{
					"postgres": {Condition: "service_started"},
					"redis":    {Condition: "service_started"},
				},
			},
			"redis": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{},
			},
		},
	}

	servicesToInclude := make(map[string]bool)
	servicesToInclude["grafana"] = true

	handler.addDependenciesRecursiveFromConfig("grafana", composeConfig, servicesToInclude)

	// Should include postgres dependency
	if !servicesToInclude["postgres"] {
		t.Error("Expected postgres to be included as dependency")
	}

	// Should not include unrelated services
	if servicesToInclude["redis"] {
		t.Error("Expected redis to not be included")
	}
	if servicesToInclude["superset"] {
		t.Error("Expected superset to not be included")
	}
}

func TestGraphHandler_addDependentsRecursive(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	// Create test data: both grafana and superset depend on postgres
	composeConfig := &container.ComposeConfig{
		Services: map[string]container.ComposeService{
			"grafana": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{
					"postgres": {Condition: "service_started"},
				},
			},
			"postgres": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{},
			},
			"superset": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{
					"postgres": {Condition: "service_started"},
				},
			},
			"redis": {
				DependsOn: map[string]struct {
					Condition string `json:"condition"`
				}{},
			},
		},
	}

	servicesToInclude := make(map[string]bool)
	servicesToInclude["postgres"] = true

	handler.addDependentsRecursiveFromConfig("postgres", composeConfig, servicesToInclude)

	// Should include services that depend on postgres
	if !servicesToInclude["grafana"] {
		t.Error("Expected grafana to be included as dependent")
	}
	if !servicesToInclude["superset"] {
		t.Error("Expected superset to be included as dependent")
	}

	// Should not include unrelated services
	if servicesToInclude["redis"] {
		t.Error("Expected redis to not be included")
	}
}

func TestGraphHandler_getAllServiceDetails_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetAllDependenciesRecursive(func(serviceName string, composeFiles []string) ([]string, error) {
			switch serviceName {
			case "grafana":
				return []string{"postgres"}, nil
			default:
				return []string{}, nil
			}
		}).
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		}).
		withGetPortMappings(func(containerName string) (map[string]string, error) {
			return map[string]string{"80/tcp": "80"}, nil
		})
	serviceHandler := NewServiceHandler(mockRuntime, "/test/insta")
	depHandler := NewDependencyHandler(mockRuntime, "/test/insta", serviceHandler)
	handler := NewGraphHandler(mockRuntime, "/test/insta", depHandler)

	services, err := handler.getAllServiceDetails()

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(services) == 0 {
		t.Error("Expected services to be returned")
	}

	// Verify service details
	foundGrafana := false
	for _, service := range services {
		if service.Name == "grafana" {
			foundGrafana = true
			if service.Status != "running" {
				t.Errorf("Expected grafana status 'running', got '%s'", service.Status)
			}
			if len(service.Dependencies) != 1 {
				t.Errorf("Expected grafana to have 1 dependency, got %d", len(service.Dependencies))
			}
		}
	}
	if !foundGrafana {
		t.Error("Expected to find grafana service in details")
	}
}

func TestGraphHandler_getComposeFiles(t *testing.T) {
	handler := NewGraphHandler(nil, "/test/insta", nil)

	composeFiles := handler.getComposeFiles()

	if len(composeFiles) == 0 {
		t.Error("Expected at least one compose file")
	}
	if !contains(composeFiles[0], "/test/insta/docker-compose.yaml") {
		t.Errorf("Expected first compose file to contain '/test/insta/docker-compose.yaml', got %s", composeFiles[0])
	}
}
