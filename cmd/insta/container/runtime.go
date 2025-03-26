package container

import (
	"fmt"
	"os"
	"os/exec"
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
}

// Provider manages container runtime detection and selection
type Provider struct {
	runtimes []Runtime
	selected Runtime
}

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

// DockerRuntime implements Runtime interface for Docker
type DockerRuntime struct{}

func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{}
}

func (d *DockerRuntime) Name() string {
	return "docker"
}

func (d *DockerRuntime) CheckAvailable() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("docker not found")
	}

	// Check if Docker daemon is running
	cmd := exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon not running")
	}

	// Check Docker Compose plugin
	cmd = exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose plugin not available")
	}

	return nil
}

// setDefaultEnvVars sets default environment variables for container operations
func setDefaultEnvVars() {
	envVars := map[string]string{
		"DB_USER":                "root",
		"DB_USER_PASSWORD":       "root",
		"ELASTICSEARCH_USER":     "elastic",
		"ELASTICSEARCH_PASSWORD": "changeme",
		"MYSQL_USER":             "root",
		"MYSQL_PASSWORD":         "root",
	}

	for key, value := range envVars {
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
}

func (d *DockerRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	// Ensure the insta network exists
	networkCmd := exec.Command("docker", "network", "create", "--driver", "bridge", "insta-network")
	_ = networkCmd.Run() // Ignore error if network already exists

	// Set default environment variables
	setDefaultEnvVars()

	args := []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "up", "-d")
	if quiet {
		args = append(args, "--quiet-pull")
	}
	args = append(args, services...)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// If the error indicates Docker daemon is not running, return a specific error
		if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			return fmt.Errorf("docker daemon not running")
		}
		return fmt.Errorf("docker compose up failed: %w", err)
	}
	return nil
}

func (d *DockerRuntime) ComposeDown(composeFiles []string, services []string) error {
	// Set default environment variables
	setDefaultEnvVars()

	args := []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "stop")
	args = append(args, services...)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose stop failed: %w", err)
	}

	// Remove stopped containers but preserve volumes
	args = []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "rm", "-f")
	args = append(args, services...)

	cmd = exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose rm failed: %w", err)
	}

	return nil
}

func (d *DockerRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	args := []string{"exec"}
	if interactive && cmd == "" {
		args = append(args, "-it")
	}
	args = append(args, containerName)
	if cmd != "" {
		args = append(args, "bash", "-c", cmd)
	} else {
		args = append(args, "bash")
	}

	execCmd := exec.Command("docker", args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command in container %s: %w", containerName, err)
	}
	return nil
}

// PodmanRuntime implements Runtime interface for Podman
type PodmanRuntime struct{}

func NewPodmanRuntime() *PodmanRuntime {
	return &PodmanRuntime{}
}

func (p *PodmanRuntime) Name() string {
	return "podman"
}

func (p *PodmanRuntime) CheckAvailable() error {
	if _, err := exec.LookPath("podman"); err != nil {
		return fmt.Errorf("podman not found")
	}

	// Check podman version
	cmd := exec.Command("podman", "version", "--format", "{{.Version}}")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get podman version: %w", err)
	}

	// Check if podman machine is running (for macOS)
	cmd = exec.Command("podman", "machine", "list", "--format", "{{.Name}}")
	if output, err := cmd.Output(); err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return fmt.Errorf("podman machine not running")
	}

	// Set podman-specific environment variables before checking compose
	setPodmanEnvVars()

	// Check compose plugin
	cmd = exec.Command("podman", "compose", "version")
	if err := cmd.Run(); err != nil {
		// Try podman-compose as fallback
		if _, fallbackErr := exec.LookPath("podman-compose"); fallbackErr != nil {
			return fmt.Errorf("neither podman compose plugin nor podman-compose found: %w", err)
		}
	}

	// Check if running rootless
	cmd = exec.Command("podman", "info", "--format", "{{.Host.Security.Rootless}}")
	output, err = cmd.Output()
	if err == nil && string(output) == "true" {
		fmt.Printf("Warning: Running in rootless mode. Some features may require additional configuration.\n")
	}

	return nil
}

// setPodmanEnvVars sets environment variables specific to podman
func setPodmanEnvVars() {
	os.Setenv("COMPOSE_PROVIDER", "podman")
}

func (p *PodmanRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	// Validate compose files
	for _, file := range composeFiles {
		if err := p.validateComposeFile(file); err != nil {
			return fmt.Errorf("invalid compose file %s: %w", file, err)
		}
	}

	// Ensure the insta network exists
	networkCmd := exec.Command("podman", "network", "create", "--driver", "bridge", "insta-network")
	_ = networkCmd.Run() // Ignore error if network already exists

	// Set default environment variables
	setDefaultEnvVars()
	// Set podman-specific environment variables
	setPodmanEnvVars()

	args := []string{"compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "up", "-d")
	if quiet {
		args = append(args, "--quiet-pull")
	}
	args = append(args, services...)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman compose up failed: %w", err)
	}
	return nil
}

func (p *PodmanRuntime) validateComposeFile(file string) error {
	// Check if file exists
	if _, err := os.Stat(file); err != nil {
		return fmt.Errorf("compose file not found: %w", err)
	}

	// Set podman-specific environment variables before validation
	setPodmanEnvVars()

	// Validate compose file
	cmd := exec.Command("podman", "compose", "-f", file, "config", "--quiet")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compose file validation failed: %w", err)
	}

	return nil
}

func (p *PodmanRuntime) ComposeDown(composeFiles []string, services []string) error {
	// Set default environment variables
	setDefaultEnvVars()
	// Set podman-specific environment variables
	setPodmanEnvVars()

	args := []string{"compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "stop")
	args = append(args, services...)

	cmd := exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman compose stop failed: %w", err)
	}

	// Remove stopped containers but preserve volumes
	args = []string{"compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "rm", "-f")
	args = append(args, services...)

	cmd = exec.Command("podman", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("podman compose rm failed: %w", err)
	}

	return nil
}

func (p *PodmanRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	args := []string{"exec"}
	if interactive && cmd == "" {
		args = append(args, "-it")
	}
	args = append(args, containerName)
	if cmd != "" {
		args = append(args, "bash", "-c", cmd)
	} else {
		args = append(args, "bash")
	}

	execCmd := exec.Command("podman", args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command in container %s: %w", containerName, err)
	}
	return nil
}
