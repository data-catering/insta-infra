package models

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
	ServiceName  string  `json:"serviceName"`  // Service name for tracking
}
