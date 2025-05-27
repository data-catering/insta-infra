package container

// ComposeService represents a service configuration in a Docker Compose file
type ComposeService struct {
	DependsOn map[string]struct {
		Condition string `json:"condition"`
	} `json:"depends_on"`
	ContainerName string `json:"container_name,omitempty"`
	Image         string `json:"image,omitempty"`
}

// ComposeConfig represents a Docker Compose configuration
type ComposeConfig struct {
	Services map[string]ComposeService `json:"services"`
}

// ImagePullProgress represents the progress of pulling a Docker/Podman image
type ImagePullProgress struct {
	Status       string  `json:"status"`       // "downloading", "extracting", "complete", "error"
	Progress     float64 `json:"progress"`     // 0.0 to 100.0
	CurrentLayer string  `json:"currentLayer"` // Current layer being processed
	TotalLayers  int     `json:"totalLayers"`  // Total number of layers
	Downloaded   int64   `json:"downloaded"`   // Bytes downloaded
	Total        int64   `json:"total"`        // Total bytes
	Speed        string  `json:"speed"`        // Download speed (e.g., "1.2 MB/s")
	ETA          string  `json:"eta"`          // Estimated time remaining
	Error        string  `json:"error"`        // Error message if any
}

// Provider manages container runtime detection and selection
type Provider struct {
	runtimes []Runtime
	selected Runtime
}

// DockerRuntime implements Runtime interface for Docker
type DockerRuntime struct {
	parsedComposeConfig   *ComposeConfig
	cachedComposeFilesKey string
}

// PodmanRuntime implements Runtime interface for Podman
type PodmanRuntime struct {
	parsedComposeConfig   *ComposeConfig
	cachedComposeFilesKey string
}
