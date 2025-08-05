package handlers

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
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
	// Get all compose files including custom ones
	composeFiles := r.getAllComposeFiles(persist)

	// Start the service using compose up
	err := r.runtime.ComposeUp(composeFiles, []string{serviceName}, true)
	if err != nil {
		return err
	}

	// Wait 5 seconds and check for stuck containers (only if not in test mode)
	if r.runtime.Name() != "test" {
		time.Sleep(5 * time.Second)

		// Get the first container that should start for this service
		firstContainer, err := r.getFirstContainerForService(serviceName, composeFiles)
		if err != nil {
			// If we can't determine the first container, skip the check
			return nil
		}

		// Check if the first container is stuck in 'created' status
		status, statusErr := r.runtime.GetContainerStatus(firstContainer)
		if statusErr == nil && status == "created" {
			// Return a structured error that the frontend can recognize and handle with a nice UI
			return fmt.Errorf("STUCK_CONTAINER:%s:The container '%s' is stuck in 'created' status. This usually indicates %s needs to be restarted.", firstContainer, firstContainer, r.runtime.Name())
		}
	}

	return nil
}

// getFirstContainerForService returns the first container that should start when starting a service
// This is typically the deepest dependency in the dependency chain
func (r *RuntimeInfoAdapter) getFirstContainerForService(serviceName string, composeFiles []string) (string, error) {
	// Get all dependencies for the service (as container names)
	dependencies, err := r.runtime.GetAllDependenciesRecursive(serviceName, composeFiles, true)
	if err != nil {
		return "", err
	}

	// If there are dependencies, the first container to start is the last one in the dependency list
	// (since dependencies are returned in order from service to deepest dependency)
	if len(dependencies) > 0 {
		return dependencies[len(dependencies)-1], nil
	}

	// If no dependencies, get the container name for the service itself
	containerName, err := r.runtime.GetContainerName(serviceName, composeFiles)
	if err != nil {
		return "", err
	}

	return containerName, nil
}

// getAllComposeFiles returns all compose files including built-in and custom files
func (r *RuntimeInfoAdapter) getAllComposeFiles(persist bool) []string {
	var composeFiles []string
	
	// Add built-in compose files
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
	
	// Add custom compose files if instaDir is available
	if r.instaDir != "" {
		customRegistry, err := models.NewCustomServiceRegistry(r.instaDir)
		if err == nil {
			customFiles := customRegistry.GetAllCustomComposeFiles()
			composeFiles = append(composeFiles, customFiles...)
		}
		// Silently ignore errors - custom services are optional
	}
	
	return composeFiles
}

func (r *RuntimeInfoAdapter) StopService(serviceName string) error {
	// Use all compose files for stopping (including custom files)
	fmt.Println("Stopping service: ", serviceName)
	composeFiles := r.getAllComposeFiles(true) // Use true to include persist files

	// Stop the service using compose down
	return r.runtime.ComposeDown(composeFiles, []string{serviceName})
}

func (r *RuntimeInfoAdapter) GetAllContainerStatuses() (map[string]string, error) {
	// Use the runtime's GetAllContainerStatuses method which calls docker ps -a or podman ps -a
	return r.runtime.GetAllContainerStatuses()
}

// GetRuntimeName returns the name of the underlying container runtime
func (r *RuntimeInfoAdapter) GetRuntimeName() string {
	if r.runtime != nil {
		return r.runtime.Name()
	}
	return "unknown"
}
