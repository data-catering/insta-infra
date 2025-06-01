package container

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func NewPodmanRuntime() *PodmanRuntime {
	return &PodmanRuntime{}
}

func (p *PodmanRuntime) Name() string {
	return "podman"
}

func (p *PodmanRuntime) CheckAvailable() error {
	// Try to find Podman binary in PATH or common installation locations
	podmanPath := findBinaryInCommonPaths("podman", getCommonPodmanPaths())
	if podmanPath == "" {
		return fmt.Errorf("podman not found in PATH or common locations")
	}

	// Store the podman path for future use
	p.podmanPath = podmanPath

	// Check podman version
	cmd := exec.Command(p.podmanPath, "version", "--format", "{{.Version}}")
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get podman version: %w", err)
	}

	// Check if podman machine is running (for macOS)
	cmd = exec.Command(p.podmanPath, "machine", "list", "--format", "{{.Name}}")
	if output, err := cmd.Output(); err != nil || len(strings.TrimSpace(string(output))) == 0 {
		return fmt.Errorf("podman machine not running")
	}

	// Set podman-specific environment variables before checking compose
	setPodmanEnvVars()

	// Check compose plugin
	cmd = exec.Command(p.podmanPath, "compose", "version")
	if err := cmd.Run(); err != nil {
		// Try podman-compose as fallback
		if _, fallbackErr := exec.LookPath("podman-compose"); fallbackErr != nil {
			return fmt.Errorf("neither podman compose plugin nor podman-compose found: %w", err)
		}
	}

	// Check if running rootless
	cmd = exec.Command(p.podmanPath, "info", "--format", "{{.Host.Security.Rootless}}")
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

	validateCmd := exec.Command(p.getPodmanCommand(), validateArgs...)
	validateCmd.Stdout = os.Stdout
	validateCmd.Stderr = os.Stderr
	validateCmd.Dir = filepath.Dir(composeFiles[0])

	if err := executeCommand(validateCmd, "compose files validation failed"); err != nil {
		return err
	}

	// Ensure the insta network exists
	networkCmd := exec.Command(p.getPodmanCommand(), "network", "create", "--driver", "bridge", "insta-network")
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

	upCmd := exec.Command(p.getPodmanCommand(), upArgs...)
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

	stopCmd := exec.Command(p.getPodmanCommand(), stopArgs...)
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

	rmCmd := exec.Command(p.getPodmanCommand(), rmArgs...)
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

	execCmd := exec.Command(p.getPodmanCommand(), args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	return executeCommand(execCmd, fmt.Sprintf("failed to execute command in container %s", containerName))
}

func (p *PodmanRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	cmd := exec.Command(p.getPodmanCommand(), "port", containerName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get port mappings for container %s: %w", containerName, err)
	}

	return parsePortMappings(string(output)), nil
}

func (p *PodmanRuntime) getOrParseComposeConfig(composeFiles []string) (*ComposeConfig, error) {
	currentFilesKey := strings.Join(composeFiles, "|")
	if p.cachedComposeFilesKey == currentFilesKey && p.parsedComposeConfig != nil {
		return p.parsedComposeConfig, nil
	}

	args := []string{"compose"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "config", "--format", "json")

	cmd := exec.Command(p.getPodmanCommand(), args...)
	if len(composeFiles) > 0 {
		cmd.Dir = filepath.Dir(composeFiles[0])
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to get podman compose configuration: %s\n%s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to get podman compose configuration: %w", err)
	}

	var config ComposeConfig
	if err := json.Unmarshal(output, &config); err != nil {
		return nil, fmt.Errorf("failed to parse podman compose configuration: %w", err)
	}

	p.parsedComposeConfig = &config
	p.cachedComposeFilesKey = currentFilesKey
	return p.parsedComposeConfig, nil
}

func (p *PodmanRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	config, err := p.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return "", fmt.Errorf("failed to get compose config: %w", err)
	}

	if service, exists := config.Services[serviceName]; exists {
		return service.Image, nil
	}

	return "", fmt.Errorf("service %s not found in compose configuration", serviceName)
}

// GetMultipleImageInfo returns image information for multiple services from compose files
func (p *PodmanRuntime) GetMultipleImageInfo(serviceNames []string, composeFiles []string) (map[string]string, error) {
	if len(serviceNames) == 0 {
		return make(map[string]string), nil
	}

	// Parse compose config once for all services
	config, err := p.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get compose config for image info: %w", err)
	}

	result := make(map[string]string)
	for _, serviceName := range serviceNames {
		service, exists := config.Services[serviceName]
		if !exists {
			// Skip services not found in compose files rather than erroring
			continue
		}

		if service.Image != "" {
			result[serviceName] = service.Image
		}
	}

	return result, nil
}

func (p *PodmanRuntime) PullImageWithProgress(imageName string, progressChan chan<- ImagePullProgress, stopChan <-chan struct{}) error {
	cmd := exec.Command(p.getPodmanCommand(), "pull", imageName)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start image pull: %w", err)
	}

	// Channel to signal when command is done
	doneChan := make(chan error, 1)

	// Capture error output for debugging
	var errorOutput strings.Builder

	// Start goroutine to read pull output and parse progress
	go func() {
		defer stdout.Close()
		scanner := bufio.NewScanner(stdout)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Parse Podman pull output
			progress := p.parsePodmanPullOutput(line)
			if progress.Status != "" {
				select {
				case progressChan <- progress:
				case <-stopChan:
					return
				}
			}
		}
	}()

	// Start goroutine to read error output
	go func() {
		defer stderr.Close()
		scanner := bufio.NewScanner(stderr)

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			errorOutput.WriteString(line + "\n")

			// Some progress information might come through stderr
			progress := p.parsePodmanPullOutput(line)
			if progress.Status != "" {
				select {
				case progressChan <- progress:
				case <-stopChan:
					return
				}
			}
		}
	}()

	// Wait for stop signal or command completion
	go func() {
		doneChan <- cmd.Wait()
	}()

	select {
	case <-stopChan:
		// Kill the process
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		progressChan <- ImagePullProgress{
			Status: "cancelled",
			Error:  "Image pull cancelled by user",
		}
		return nil
	case err := <-doneChan:
		if err != nil {
			errorMsg := fmt.Sprintf("Image pull failed: %s", err.Error())
			if errorOutput.Len() > 0 {
				errorMsg += fmt.Sprintf("\nPodman output: %s", errorOutput.String())
			}

			progressChan <- ImagePullProgress{
				Status: "error",
				Error:  errorMsg,
			}
			return fmt.Errorf("failed to pull image: %w", err)
		}

		progressChan <- ImagePullProgress{
			Status:   "complete",
			Progress: 100.0,
		}
		return nil
	}
}

func (p *PodmanRuntime) parsePodmanPullOutput(line string) ImagePullProgress {
	// Podman pull output format examples:
	// "Trying to pull docker.io/library/redis:alpine..."
	// "Getting image source signatures"
	// "Copying blob 8bc3a26b84da done"
	// "Writing manifest to image destination"

	if strings.Contains(line, "Trying to pull") {
		return ImagePullProgress{
			Status: "starting",
		}
	}

	if strings.Contains(line, "Getting image source signatures") {
		return ImagePullProgress{
			Status:   "downloading",
			Progress: 10.0,
		}
	}

	if strings.Contains(line, "Copying blob") && strings.Contains(line, "done") {
		return ImagePullProgress{
			Status:   "downloading",
			Progress: 70.0,
		}
	}

	if strings.Contains(line, "Writing manifest") {
		return ImagePullProgress{
			Status:   "downloading",
			Progress: 90.0,
		}
	}

	if strings.Contains(line, "Storing signatures") {
		return ImagePullProgress{
			Status:   "complete",
			Progress: 100.0,
		}
	}

	return ImagePullProgress{} // Empty progress for unrecognized lines
}

func (p *PodmanRuntime) GetContainerStatus(containerName string) (string, error) {
	// Check all containers (including stopped ones)
	cmd := exec.Command(p.getPodmanCommand(), "ps", "-a", "--format", "{{.Names}}\t{{.Status}}", "--filter", fmt.Sprintf("name=^%s$", containerName))
	output, err := cmd.Output()
	if err != nil {
		return "not_found", fmt.Errorf("failed to check container status: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		// No container found with this name
		return "not_found", nil
	}

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) >= 2 && strings.TrimSpace(parts[0]) == containerName {
			status := strings.TrimSpace(parts[1])

			// Normalize Podman status to our standard status values
			statusLower := strings.ToLower(status)
			if strings.Contains(statusLower, "up") {
				return "running", nil
			} else if strings.Contains(statusLower, "exited") {
				// Check exit code for init containers
				if strings.Contains(statusLower, "exited (0)") {
					return "completed", nil
				}
				return "error", nil
			} else if strings.Contains(statusLower, "created") {
				return "stopped", nil
			} else if strings.Contains(statusLower, "paused") {
				return "paused", nil
			} else if strings.Contains(statusLower, "restarting") {
				return "restarting", nil
			}

			// Default for unknown status
			return "unknown", nil
		}
	}

	return "not_found", nil
}

// getPodmanCommand returns the path to the podman binary
func (p *PodmanRuntime) getPodmanCommand() string {
	if p.podmanPath != "" {
		return p.podmanPath
	}
	// Fallback to "podman" if path not set (shouldn't happen after CheckAvailable)
	return "podman"
}

// GetAllContainerStatuses returns all current containers (including stopped ones) managed by compose
// Returns back the container names and statuses in a map
func (p *PodmanRuntime) GetAllContainerStatuses() (map[string]string, error) {
	// Use podman ps to get all running containers with compose labels
	// This is much faster than checking each service individually
	cmd := exec.Command(p.getPodmanCommand(), "ps", "-a",
		"--filter", "label=com.docker.compose.project=insta",
		"--format", "{{.Names}},{{.State}},{{.Status}}")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get running containers: %w", err)
	}

	containerStatuses := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ",")
		if len(parts) > 1 {
			state := parts[1]
			status := parts[2] // This is the status of the container
			if strings.Contains(state, "created") {
				containerStatuses[parts[0]] = "created"
			} else if strings.Contains(state, "running") && strings.Contains(status, "starting") {
				containerStatuses[parts[0]] = "starting"
			} else if strings.Contains(state, "running") && strings.Contains(status, "unhealthy") {
				containerStatuses[parts[0]] = "error"
			} else if strings.Contains(state, "running") {
				containerStatuses[parts[0]] = "running"
			} else if strings.Contains(state, "exited") {
				containerStatuses[parts[0]] = "stopped"
			} else if strings.Contains(state, "restarting") {
				containerStatuses[parts[0]] = "restarting"
			} else {
				containerStatuses[parts[0]] = "stopped"
			}
		}
	}

	return containerStatuses, nil
}

func (p *PodmanRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
	config, err := p.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return "", err
	}

	// First, check if there's an explicit container name in the compose config
	if serviceConfig, ok := config.Services[serviceName]; ok {
		if serviceConfig.ContainerName != "" {
			return serviceConfig.ContainerName, nil
		}
	}

	// For Podman, try different naming patterns similar to Docker
	candidateNames := []string{
		serviceName,                            // Direct service name (e.g., "airflow-init")
		fmt.Sprintf("insta_%s_1", serviceName), // Default compose pattern
		fmt.Sprintf("insta-%s-1", serviceName), // Alternative dash pattern
	}

	// Check which container actually exists
	for _, candidateName := range candidateNames {
		if p.containerExistsAnywhere(candidateName) {
			return candidateName, nil
		}
	}

	// If none exist, return the most likely candidate (service name itself)
	return serviceName, nil
}

// GetAllDependenciesRecursive returns all dependencies recursively for a service from compose files
// Returns container names (not service names) for UI display purposes
func (p *PodmanRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string, isContainer bool) ([]string, error) {
	config, err := p.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get compose config: %w", err)
	}

	visited := make(map[string]bool)
	dependencyServices := collectDependenciesRecursive(serviceName, config, visited)

	if !isContainer {
		return dependencyServices, nil
	}

	// Convert service names to container names
	var containerNames []string
	for _, serviceName := range dependencyServices {
		containerName, err := p.GetContainerName(serviceName, composeFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to get container name for service %s: %w", serviceName, err)
		}
		containerNames = append(containerNames, containerName)
	}

	return containerNames, nil
}

func (p *PodmanRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	args := buildLogCommand(containerName, tailLines, false)
	cmd := exec.Command(p.getPodmanCommand(), args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for container %s: %w", containerName, err)
	}

	return parseLogOutput(output), nil
}

func (p *PodmanRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	args := buildLogCommand(containerName, 0, true)
	cmd := exec.Command(p.getPodmanCommand(), args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start log streaming: %w", err)
	}

	// Channel to signal when command is done
	doneChan := make(chan error, 1)

	// Use shared log streaming utilities
	streamLogsFromPipes(stdout, stderr, logChan, stopChan)

	// Wait for stop signal or command completion
	go func() {
		doneChan <- cmd.Wait()
	}()

	select {
	case <-stopChan:
		// Kill the process
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil
	case err := <-doneChan:
		return err
	}
}

// containerExistsAnywhere checks if a container with the given name exists (running or stopped)
func (p *PodmanRuntime) containerExistsAnywhere(containerName string) bool {
	cmd := exec.Command(p.getPodmanCommand(), "ps", "-a", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=^%s$", containerName))
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	containerNames := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, name := range containerNames {
		if strings.TrimSpace(name) == containerName {
			return true
		}
	}
	return false
}

func (p *PodmanRuntime) CheckImageExists(imageName string) (bool, error) {
	cmd := exec.Command(p.getPodmanCommand(), "image", "inspect", imageName)
	return cmd.Run() == nil, nil
}

// CheckMultipleImagesExist checks if multiple Podman images exist locally in a single call
func (p *PodmanRuntime) CheckMultipleImagesExist(imageNames []string) (map[string]bool, error) {
	if len(imageNames) == 0 {
		return make(map[string]bool), nil
	}

	// Use podman images command to get all local images in one call
	cmd := exec.Command(p.getPodmanCommand(), "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list podman images: %w", err)
	}

	// Parse the output to create a set of existing images
	existingImages := make(map[string]bool)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "<none>:<none>" {
			existingImages[line] = true
		}
	}

	// Check each requested image
	result := make(map[string]bool)
	for _, imageName := range imageNames {
		// Handle different image name formats
		exists := false

		// Check exact match first
		if existingImages[imageName] {
			exists = true
		} else {
			// If no tag specified, check with :latest
			if !strings.Contains(imageName, ":") {
				if existingImages[imageName+":latest"] {
					exists = true
				}
			}

			// Also check without tag (some images might be listed differently)
			if !exists && strings.Contains(imageName, ":") {
				baseImage := strings.Split(imageName, ":")[0]
				for existingImage := range existingImages {
					if strings.HasPrefix(existingImage, baseImage+":") {
						exists = true
						break
					}
				}
			}
		}

		result[imageName] = exists
	}

	return result, nil
}
