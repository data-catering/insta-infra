package container

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

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

// parsePortMappings parses port mappings output (shared between runtimes)
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

// extractDependencies extracts dependencies from compose config (shared between runtimes)
func extractDependencies(config ComposeConfig, service string) []string {
	var dependencies []string
	if serviceConfig, ok := config.Services[service]; ok {
		for dep := range serviceConfig.DependsOn {
			dependencies = append(dependencies, dep)
		}
	}
	return dependencies
}

// executeCommand executes shell commands with robust error handling
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

// getServiceNames returns a list of service names from the compose config
func getServiceNames(config *ComposeConfig) []string {
	names := make([]string, 0, len(config.Services))
	for name := range config.Services {
		names = append(names, name)
	}
	return names
}

// setPodmanEnvVars sets environment variables specific to podman
func setPodmanEnvVars() {
	os.Setenv("COMPOSE_PROVIDER", "podman")
}

// getCommonDockerPaths returns common Docker installation paths for the current OS
func getCommonDockerPaths() []string {
	switch runtime.GOOS {
	case "darwin": // macOS
		return []string{
			"/usr/local/bin/docker",                                  // Intel Mac with Homebrew
			"/opt/homebrew/bin/docker",                               // Apple Silicon Mac with Homebrew
			"/Applications/Docker.app/Contents/Resources/bin/docker", // Docker Desktop
		}
	case "linux":
		return []string{
			"/usr/bin/docker",                     // System package manager
			"/usr/local/bin/docker",               // Manual installation
			"/opt/docker/bin/docker",              // Alternative installation
			"/snap/bin/docker",                    // Snap package
			"/var/lib/flatpak/exports/bin/docker", // Flatpak
		}
	case "windows":
		return []string{
			"C:\\Program Files\\Docker\\Docker\\resources\\bin\\docker.exe", // Docker Desktop
			"C:\\ProgramData\\chocolatey\\bin\\docker.exe",                  // Chocolatey
			"C:\\tools\\docker\\docker.exe",                                 // Manual installation
		}
	default:
		return []string{}
	}
}

// getCommonPodmanPaths returns common Podman installation paths for the current OS
func getCommonPodmanPaths() []string {
	switch runtime.GOOS {
	case "darwin": // macOS
		return []string{
			"/usr/local/bin/podman",    // Intel Mac with Homebrew
			"/opt/homebrew/bin/podman", // Apple Silicon Mac with Homebrew
		}
	case "linux":
		return []string{
			"/usr/bin/podman",                     // System package manager
			"/usr/local/bin/podman",               // Manual installation
			"/opt/podman/bin/podman",              // Alternative installation
			"/snap/bin/podman",                    // Snap package
			"/var/lib/flatpak/exports/bin/podman", // Flatpak
		}
	case "windows":
		return []string{
			"C:\\Program Files\\RedHat\\Podman\\podman.exe", // Official installer
			"C:\\ProgramData\\chocolatey\\bin\\podman.exe",  // Chocolatey
			"C:\\tools\\podman\\podman.exe",                 // Manual installation
		}
	default:
		return []string{}
	}
}

// findBinaryInCommonPaths tries to find a binary in common installation paths
func findBinaryInCommonPaths(binaryName string, commonPaths []string) string {
	// First check for custom path via environment variable
	customPath := getCustomBinaryPath(binaryName)
	if customPath != "" {
		if _, err := os.Stat(customPath); err == nil {
			return customPath
		}
		// Log warning if custom path is invalid but continue with fallback
		fmt.Printf("Warning: Custom %s path '%s' not found, falling back to standard detection\n", binaryName, customPath)
	}

	// Try the standard PATH lookup
	if path, err := exec.LookPath(binaryName); err == nil {
		return path
	}

	// Try common installation locations
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// getCustomBinaryPath returns custom binary path from environment variables
func getCustomBinaryPath(binaryName string) string {
	switch binaryName {
	case "docker":
		return os.Getenv("INSTA_DOCKER_PATH")
	case "podman":
		return os.Getenv("INSTA_PODMAN_PATH")
	default:
		return ""
	}
}
