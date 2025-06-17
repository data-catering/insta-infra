package models

import "github.com/data-catering/insta-infra/v2/internal/core"

// ServiceInfo defines the structure for basic service information (name, etc.)
// Still useful for the initial list of services before details are fetched.
type ServiceInfo core.Service

// ServiceDetailInfo holds comprehensive details for a single service.
type ServiceDetailInfo struct {
	Name              string   `json:"name"`
	Type              string   `json:"type"`
	Status            string   `json:"status"`
	StatusError       string   `json:"statusError,omitempty"` // Store error message if status check fails
	Dependencies      []string `json:"dependencies"`
	DependenciesError string   `json:"dependenciesError,omitempty"` // Store error message if dependency check fails
}

// ServiceStatus holds just the status information for a service
// Used for progressive loading where we fetch statuses separately from basic service info
type ServiceStatus struct {
	ServiceName string `json:"serviceName"`
	Status      string `json:"status"`          // "running", "stopped", "error"
	Error       string `json:"error,omitempty"` // Error message if status check failed
}

// ServiceContainerDetails holds detailed container information for drill-down functionality
type ServiceContainerDetails struct {
	ServiceName string          `json:"serviceName"`
	Containers  []ContainerInfo `json:"containers"`
}

// ContainerInfo holds information about a specific container
type ContainerInfo struct {
	Name        string `json:"name"`        // Container name (e.g., "postgres", "postgres-data")
	ServiceName string `json:"serviceName"` // Service name from compose (e.g., "postgres-server", "postgres")
	Status      string `json:"status"`      // Container status ("running", "completed", "stopped", etc.)
	Role        string `json:"role"`        // Role in the service ("main", "auxiliary")
	Description string `json:"description"` // User-friendly description
}
