package models

// DependencyInfo represents detailed information about a service dependency
type DependencyInfo struct {
	ServiceName  string `json:"serviceName"`
	Status       string `json:"status"`       // "running", "stopped", "starting", "stopping", "error"
	Type         string `json:"type"`         // "database", "messaging", etc.
	Health       string `json:"health"`       // "healthy", "unhealthy", "unknown"
	Required     bool   `json:"required"`     // true for required dependencies, false for optional
	StartupOrder int    `json:"startupOrder"` // order in which dependencies should be started
	Error        string `json:"error,omitempty"`

	// Enhanced failure detection fields
	FailureReason   string `json:"failureReason,omitempty"`   // Specific reason for failure (e.g., "init container failed")
	ContainerStatus string `json:"containerStatus,omitempty"` // Raw container status from runtime
	ExitCode        int    `json:"exitCode,omitempty"`        // Exit code if container exited
	HasLogs         bool   `json:"hasLogs"`                   // Whether logs are available for investigation
	LastFailureTime string `json:"lastFailureTime,omitempty"` // Timestamp of last failure
}

// DependencyStatus represents the complete dependency status for a service
type DependencyStatus struct {
	ServiceName          string           `json:"serviceName"`
	Dependencies         []DependencyInfo `json:"dependencies"`
	AllDependenciesReady bool             `json:"allDependenciesReady"`
	CanStart             bool             `json:"canStart"`
	RequiredCount        int              `json:"requiredCount"`
	RunningCount         int              `json:"runningCount"`
	ErrorCount           int              `json:"errorCount"`

	// Enhanced failure tracking
	FailedDependencies []string `json:"failedDependencies,omitempty"` // List of failed dependency names
}
