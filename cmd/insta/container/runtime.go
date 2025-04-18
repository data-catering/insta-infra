package container

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// Common types for JSON parsing
type ComposeService struct {
	DependsOn map[string]struct {
		Condition string `json:"condition"`
	} `json:"depends_on"`
}

type ComposeConfig struct {
	Services map[string]ComposeService `json:"services"`
}

// Helper to parse port mappings output (shared between runtimes)
func parsePortMappings(output string) map[string]string {
	portMappings := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, " -> ")
		if len(parts) == 2 {
			containerPort := strings.TrimSpace(parts[0])
			hostPort := strings.TrimSpace(parts[1])
			// Extract just the port number from host port (e.g., "0.0.0.0:8080" -> "8080")
			if idx := strings.LastIndex(hostPort, ":"); idx != -1 {
				hostPort = hostPort[idx+1:]
			}
			portMappings[containerPort] = hostPort
		}
	}
	return portMappings
}

// Helper to extract dependencies from compose config (shared between runtimes)
func extractDependencies(config ComposeConfig, service string) []string {
	var dependencies []string
	if serviceConfig, ok := config.Services[service]; ok {
		for dep := range serviceConfig.DependsOn {
			dependencies = append(dependencies, dep)
		}
	}
	return dependencies
}

// Helper for executing shell commands with robust error handling
func executeCommand(cmd *exec.Cmd, errorPrefix string) error {
	if err := cmd.Run(); err != nil {
		// Capture the error output for more details
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s: %s\n%s", errorPrefix, err, string(exitErr.Stderr))
		}
		return fmt.Errorf("%s: %w", errorPrefix, err)
	}
	return nil
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
	// Set working directory to the directory containing the first compose file
	cmd.Dir = filepath.Dir(composeFiles[0])
	
	if err := cmd.Run(); err != nil {
		// If the error indicates Docker daemon is not running, return a specific error
		if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			return fmt.Errorf("docker daemon not running")
		}
		// Use the helper for error handling
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("docker compose up failed: %s\n%s", err, string(exitErr.Stderr))
		}
		return fmt.Errorf("docker compose up failed: %w", err)
	}
	return nil
}

func (d *DockerRuntime) ComposeDown(composeFiles []string, services []string) error {
	// Set default environment variables
	setDefaultEnvVars()

	// Stop containers
	stopArgs := []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		stopArgs = append(stopArgs, "-f", file)
	}
	stopArgs = append(stopArgs, "stop")
	stopArgs = append(stopArgs, services...)

	stopCmd := exec.Command("docker", stopArgs...)
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr
	stopCmd.Dir = filepath.Dir(composeFiles[0])
	
	if err := executeCommand(stopCmd, "docker compose stop failed"); err != nil {
		return err
	}

	// Remove stopped containers but preserve volumes
	rmArgs := []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		rmArgs = append(rmArgs, "-f", file)
	}
	rmArgs = append(rmArgs, "rm", "-f")
	rmArgs = append(rmArgs, services...)

	rmCmd := exec.Command("docker", rmArgs...)
	rmCmd.Stdout = os.Stdout
	rmCmd.Stderr = os.Stderr
	rmCmd.Dir = filepath.Dir(composeFiles[0])
	
	return executeCommand(rmCmd, "docker compose rm failed")
}

func (d *DockerRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	args := []string{"exec"}
	if interactive && os.Getenv("TESTING") != "true" {
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
	
	return executeCommand(execCmd, fmt.Sprintf("failed to execute command in container %s", containerName))
}

func (d *DockerRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	cmd := exec.Command("docker", "port", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get port mappings for container %s: %w", containerName, err)
	}

	return parsePortMappings(string(output)), nil
}

func (d *DockerRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	args := []string{"--log-level", "error", "compose"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "config", "--format", "json")

	cmd := exec.Command("docker", args...)
	cmd.Dir = filepath.Dir(composeFiles[0])
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get docker compose configuration: %w", err)
	}

	var config ComposeConfig
	if err := json.Unmarshal(output, &config); err != nil {
		return nil, fmt.Errorf("failed to parse docker compose configuration: %w", err)
	}

	return extractDependencies(config, service), nil
}

// PodmanRuntime implements Runtime interface for Podman
type PodmanRuntime struct{}

func NewPodmanRuntime() *PodmanRuntime {
	return &PodmanRuntime{}
}

func (p *PodmanRuntime) Name() string {
	return "podman"
}

// setPodmanEnvVars sets environment variables specific to podman
func setPodmanEnvVars() {
	os.Setenv("COMPOSE_PROVIDER", "podman")
}

func (p *PodmanRuntime) CheckAvailable() error {
	if _, err := exec.LookPath("podman"); err != nil {
		return fmt.Errorf("podman not found")
	}

	// Check podman version
	cmd := exec.Command("podman", "version", "--format", "{{.Version}}")
	_, err := cmd.Output()
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
	output, err := cmd.Output()
	if err == nil && string(output) == "true" {
		fmt.Printf("Warning: Running in rootless mode. Some features may require additional configuration.\n")
	}

	return nil
}

func (p *PodmanRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	// Set environment variables
	setDefaultEnvVars()
	setPodmanEnvVars()

	// Validate all compose files together
	validateArgs := []string{"--log-level", "error", "compose"}
	for _, file := range composeFiles {
		validateArgs = append(validateArgs, "-f", file)
	}
	validateArgs = append(validateArgs, "config", "--quiet")

	validateCmd := exec.Command("podman", validateArgs...)
	validateCmd.Stdout = os.Stdout
	validateCmd.Stderr = os.Stderr
	validateCmd.Dir = filepath.Dir(composeFiles[0])
	
	if err := executeCommand(validateCmd, "compose files validation failed"); err != nil {
		return err
	}

	// Ensure the insta network exists
	networkCmd := exec.Command("podman", "network", "create", "--driver", "bridge", "insta-network")
	_ = networkCmd.Run() // Ignore error if network already exists

	// Start services
	upArgs := []string{"compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		upArgs = append(upArgs, "-f", file)
	}
	upArgs = append(upArgs, "up", "-d")
	if quiet {
		upArgs = append(upArgs, "--quiet-pull")
	}
	upArgs = append(upArgs, services...)

	upCmd := exec.Command("podman", upArgs...)
	upCmd.Stderr = os.Stderr
	upCmd.Dir = filepath.Dir(composeFiles[0])
	
	return executeCommand(upCmd, "podman compose up failed")
}

func (p *PodmanRuntime) ComposeDown(composeFiles []string, services []string) error {
	// Set environment variables
	setDefaultEnvVars()
	setPodmanEnvVars()

	// Stop containers
	stopArgs := []string{"compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		stopArgs = append(stopArgs, "-f", file)
	}
	stopArgs = append(stopArgs, "stop")
	stopArgs = append(stopArgs, services...)

	stopCmd := exec.Command("podman", stopArgs...)
	stopCmd.Stdout = os.Stdout
	stopCmd.Stderr = os.Stderr
	stopCmd.Dir = filepath.Dir(composeFiles[0])
	
	if err := executeCommand(stopCmd, "podman compose stop failed"); err != nil {
		return err
	}

	// Remove stopped containers but preserve volumes
	rmArgs := []string{"compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		rmArgs = append(rmArgs, "-f", file)
	}
	rmArgs = append(rmArgs, "rm", "-f")
	rmArgs = append(rmArgs, services...)

	rmCmd := exec.Command("podman", rmArgs...)
	rmCmd.Stdout = os.Stdout
	rmCmd.Stderr = os.Stderr
	rmCmd.Dir = filepath.Dir(composeFiles[0])
	
	return executeCommand(rmCmd, "podman compose rm failed")
}

func (p *PodmanRuntime) ExecInContainer(containerName string, cmd string, interactive bool) error {
	args := []string{"exec"}
	if interactive && os.Getenv("TESTING") != "true" {
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
	
	return executeCommand(execCmd, fmt.Sprintf("failed to execute command in container %s", containerName))
}

func (p *PodmanRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	cmd := exec.Command("podman", "port", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get port mappings for container %s: %w", containerName, err)
	}

	return parsePortMappings(string(output)), nil
}

func (p *PodmanRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	args := []string{"compose"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "config", "--format", "json")

	cmd := exec.Command("podman", args...)
	cmd.Dir = filepath.Dir(composeFiles[0])
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get podman compose configuration: %w", err)
	}

	var config ComposeConfig
	if err := json.Unmarshal(output, &config); err != nil {
		return nil, fmt.Errorf("failed to parse podman compose configuration: %w", err)
	}

	return extractDependencies(config, service), nil
}
