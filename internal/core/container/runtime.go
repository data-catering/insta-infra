package container

import (
	"fmt"
	"strings"
)

// Runtime represents a container runtime (docker, podman, etc)
type Runtime interface {
	// Name returns the name of the runtime
	Name() string
	// CheckAvailable checks if the runtime and compose are available
	CheckAvailable() error
	// ComposeUp starts services with the given compose files and options
	ComposeUp(composeFiles []string, services []string, quiet bool) error
	// ComposeDown stops services
	ComposeDown(composeFiles []string, services []string) error
	// ExecInContainer executes a command in a container
	ExecInContainer(containerName string, cmd string, interactive bool) error
	// GetPortMappings returns port mappings for a container
	GetPortMappings(containerName string) (map[string]string, error)
	// GetDependencies returns dependencies for a service
	GetDependencies(service string, composeFiles []string) ([]string, error)
	// GetAllDependenciesRecursive returns all direct and indirect dependencies for a service
	GetAllDependenciesRecursive(serviceName string, composeFiles []string) ([]string, error)
	// GetContainerName returns the container name for a service
	GetContainerName(serviceName string, composeFiles []string) (string, error)
	// GetContainerLogs returns recent logs from a container
	GetContainerLogs(containerName string, tailLines int) ([]string, error)
	// StreamContainerLogs streams logs from a container in real-time
	StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error
	// CheckImageExists checks if an image exists locally
	CheckImageExists(imageName string) (bool, error)
	// GetImageInfo returns information about a service's image from compose files
	GetImageInfo(serviceName string, composeFiles []string) (string, error)
	// PullImageWithProgress pulls an image and reports progress
	PullImageWithProgress(imageName string, progressChan chan<- ImagePullProgress, stopChan <-chan struct{}) error
	// GetContainerStatus returns the status of a container
	GetContainerStatus(containerName string) (string, error)
}

// Provider struct moved to types.go

// NewProvider creates a new runtime provider with all supported runtimes
func NewProvider() *Provider {
	return &Provider{
		runtimes: []Runtime{
			NewDockerRuntime(),
			NewPodmanRuntime(),
		},
	}
}

// DetectRuntime tries to detect and select an available container runtime
func (p *Provider) DetectRuntime() error {
	for _, rt := range p.runtimes {
		if err := rt.CheckAvailable(); err == nil {
			p.selected = rt
			return nil
		}
	}
	return fmt.Errorf("no supported container runtime found (tried: docker, podman)")
}

// SelectedRuntime returns the selected container runtime
func (p *Provider) SelectedRuntime() Runtime {
	return p.selected
}

// SetRuntime explicitly sets the container runtime
func (p *Provider) SetRuntime(name string) error {
	name = strings.ToLower(name)
	for _, rt := range p.runtimes {
		if rt.Name() == name {
			if err := rt.CheckAvailable(); err != nil {
				return fmt.Errorf("runtime %s is not available: %w", name, err)
			}
			p.selected = rt
			return nil
		}
	}
	return fmt.Errorf("unsupported runtime: %s (supported: docker, podman)", name)
}
