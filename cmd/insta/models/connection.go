package models

// EnhancedServiceConnectionInfo holds detailed connection information from enhanced service model
type EnhancedServiceConnectionInfo struct {
	ServiceName string `json:"serviceName"`
	Available   bool   `json:"available"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`

	// Service metadata
	ContainerName string `json:"containerName,omitempty"`
	ImageName     string `json:"imageName,omitempty"`
	Type          string `json:"type,omitempty"`

	// Authentication
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	ConnectionCommand string `json:"connectionCommand,omitempty"`

	// Enhanced connection data
	WebUrls           []WebURL           `json:"webUrls,omitempty"`
	ExposedPorts      []PortMapping      `json:"exposedPorts,omitempty"`
	InternalPorts     []PortMapping      `json:"internalPorts,omitempty"`
	ConnectionStrings []ConnectionString `json:"connectionStrings,omitempty"`
	Credentials       []Credential       `json:"credentials,omitempty"`

	// Dependencies
	DirectDependencies    []string `json:"directDependencies,omitempty"`
	RecursiveDependencies []string `json:"recursiveDependencies,omitempty"`
	AllContainers         []string `json:"allContainers,omitempty"`
}

// ConnectionString holds structured connection string information
type ConnectionString struct {
	Description      string `json:"description"`
	ConnectionString string `json:"connectionString"`
	Note             string `json:"note,omitempty"`
}

// Credential holds structured credential information
type Credential struct {
	Description string `json:"description"`
	Value       string `json:"value"`
	Note        string `json:"note,omitempty"`
}
