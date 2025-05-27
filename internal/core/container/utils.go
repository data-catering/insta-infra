package container

import (
	"fmt"
	"os"
	"os/exec"
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
