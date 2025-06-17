package handlers

import (
	"fmt"
	"path/filepath"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// RuntimeInfoAdapter adapts container.Runtime to models.RuntimeInfo interface
type RuntimeInfoAdapter struct {
	runtime  container.Runtime
	instaDir string
}

// NewRuntimeInfoAdapter creates a new RuntimeInfoAdapter
func NewRuntimeInfoAdapter(runtime container.Runtime) *RuntimeInfoAdapter {
	return NewRuntimeInfoAdapterWithDir(runtime, "")
}

// NewRuntimeInfoAdapterWithDir creates a new RuntimeInfoAdapter with instaDir
func NewRuntimeInfoAdapterWithDir(runtime container.Runtime, instaDir string) *RuntimeInfoAdapter {
	return &RuntimeInfoAdapter{
		runtime:  runtime,
		instaDir: instaDir,
	}
}

func (r *RuntimeInfoAdapter) CheckContainerStatus(containerName string) (string, error) {
	status, err := r.runtime.GetContainerStatus(containerName)
	if err != nil {
		return "error", err
	}
	return status, nil
}

func (r *RuntimeInfoAdapter) GetContainerLogs(containerName string, lines int) ([]string, error) {
	logs, err := r.runtime.GetContainerLogs(containerName, lines)
	if err != nil {
		return nil, err
	}
	return logs, nil
}

func (r *RuntimeInfoAdapter) StartService(serviceName string, persist bool) error {
	// Determine compose files to use with absolute paths
	var composeFiles []string
	if r.instaDir != "" {
		composeFiles = []string{filepath.Join(r.instaDir, "docker-compose.yaml")}
		if persist {
			composeFiles = append(composeFiles, filepath.Join(r.instaDir, "docker-compose-persist.yaml"))
		}
	} else {
		// Fallback to relative paths if instaDir not set
		composeFiles = []string{"docker-compose.yaml"}
		if persist {
			composeFiles = append(composeFiles, "docker-compose-persist.yaml")
		}
	}

	// Start the service using compose up
	return r.runtime.ComposeUp(composeFiles, []string{serviceName}, true)
}

func (r *RuntimeInfoAdapter) StopService(serviceName string) error {
	// Use both compose files for stopping with absolute paths
	fmt.Println("Stopping service: ", serviceName)
	var composeFiles []string
	if r.instaDir != "" {
		composeFiles = []string{
			filepath.Join(r.instaDir, "docker-compose.yaml"),
			filepath.Join(r.instaDir, "docker-compose-persist.yaml"),
		}
	} else {
		// Fallback to relative paths if instaDir not set
		composeFiles = []string{"docker-compose.yaml", "docker-compose-persist.yaml"}
	}

	// Stop the service using compose down
	return r.runtime.ComposeDown(composeFiles, []string{serviceName})
}

func (r *RuntimeInfoAdapter) GetAllContainerStatuses() (map[string]string, error) {
	// Use the runtime's GetAllContainerStatuses method which calls docker ps -a or podman ps -a
	return r.runtime.GetAllContainerStatuses()
}
