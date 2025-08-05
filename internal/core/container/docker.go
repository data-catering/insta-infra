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

func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{}
}

func (d *DockerRuntime) Name() string {
	return "docker"
}

func (d *DockerRuntime) CheckAvailable() error {
	// Try to find Docker binary in PATH or common installation locations
	dockerPath := findBinaryInCommonPaths("docker", getCommonDockerPaths())
	if dockerPath == "" {
		return fmt.Errorf("docker not found in PATH or common locations")
	}

	// Store the docker path for future use
	d.dockerPath = dockerPath

	// Check if Docker daemon is running
	cmd := exec.Command(d.dockerPath, "info")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker daemon not running")
	}

	// Check Docker Compose plugin
	cmd = exec.Command(d.dockerPath, "compose", "version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose plugin not available")
	}

	return nil
}

// getDockerCommand returns the path to the docker binary
func (d *DockerRuntime) getDockerCommand() string {
	if d.dockerPath != "" {
		return d.dockerPath
	}
	// Fallback to "docker" if path not set (shouldn't happen after CheckAvailable)
	return "docker"
}

func (d *DockerRuntime) ComposeUp(composeFiles []string, services []string, quiet bool) error {
	// Ensure the insta network exists
	networkCmd := exec.Command(d.getDockerCommand(), "network", "create", "--driver", "bridge", "insta-network")
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

	cmd := exec.Command(d.getDockerCommand(), args...)
	// Set working directory to the directory containing the first compose file
	cmd.Dir = filepath.Dir(composeFiles[0])

	// Capture both stdout and stderr since docker compose writes to both
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the error indicates Docker daemon is not running, return a specific error
		if strings.Contains(err.Error(), "Cannot connect to the Docker daemon") {
			return fmt.Errorf("docker daemon not running")
		}
		// Include the command output in the error message for better debugging
		if len(output) > 0 {
			return fmt.Errorf("docker compose up failed: %s\nOutput: %s", err, string(output))
		}
		return fmt.Errorf("docker compose up failed: %s", err)
	}
	return nil
}

func (d *DockerRuntime) ComposeDown(composeFiles []string, services []string) error {
	// Set default environment variables
	setDefaultEnvVars()

	// TODO: Only stop containers that do not have another dependency running (for example, if keycloak is stopped and airflow is also running, we should not stop postgres)

	// Get all dependencies of the services
	for _, service := range services {
		dependencies, err := d.GetAllDependenciesRecursive(service, composeFiles, false)
		if err != nil {
			return fmt.Errorf("failed to get all dependencies of the services: %w", err)
		}
		services = append(services, dependencies...)
	}

	// Stop containers
	stopArgs := []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		stopArgs = append(stopArgs, "-f", file)
	}
	stopArgs = append(stopArgs, "stop")
	stopArgs = append(stopArgs, services...)

	stopCmd := exec.Command(d.getDockerCommand(), stopArgs...)
	stopCmd.Dir = filepath.Dir(composeFiles[0])

	// Capture both stdout and stderr since docker compose writes to both
	output, err := stopCmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("docker compose stop failed: %s\nOutput: %s", err, string(output))
		}
		return fmt.Errorf("docker compose stop failed: %s", err)
	}

	// Remove stopped containers but preserve volumes
	rmArgs := []string{"--log-level", "error", "compose", "--project-name", "insta"}
	for _, file := range composeFiles {
		rmArgs = append(rmArgs, "-f", file)
	}
	rmArgs = append(rmArgs, "rm", "-f")
	rmArgs = append(rmArgs, services...)

	rmCmd := exec.Command(d.getDockerCommand(), rmArgs...)
	rmCmd.Dir = filepath.Dir(composeFiles[0])

	// Capture both stdout and stderr since docker compose writes to both
	output, err = rmCmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("docker compose rm failed: %s\nOutput: %s", err, string(output))
		}
		return fmt.Errorf("docker compose rm failed: %s", err)
	}

	return nil
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

	execCmd := exec.Command(d.getDockerCommand(), args...)
	execCmd.Stdin = os.Stdin
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr

	// For interactive commands, we need to keep the original behavior
	// since we want to see the output in real-time
	if interactive {
		return executeCommand(execCmd, fmt.Sprintf("failed to execute command in container %s", containerName))
	}

	// For non-interactive commands, use CombinedOutput for better error handling
	output, err := execCmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return fmt.Errorf("failed to execute command in container %s: %s\nOutput: %s", containerName, err, string(output))
		}
		return fmt.Errorf("failed to execute command in container %s: %s", containerName, err)
	}

	return nil
}

func (d *DockerRuntime) GetPortMappings(containerName string) (map[string]string, error) {
	cmd := exec.Command(d.getDockerCommand(), "port", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get port mappings for container %s: %w", containerName, err)
	}

	return parsePortMappings(string(output)), nil
}

func (d *DockerRuntime) getOrParseComposeConfig(composeFiles []string) (*ComposeConfig, error) {
	currentFilesKey := strings.Join(composeFiles, "|")
	if d.cachedComposeFilesKey == currentFilesKey && d.parsedComposeConfig != nil {
		return d.parsedComposeConfig, nil
	}

	args := []string{"--log-level", "error", "compose"}
	for _, file := range composeFiles {
		args = append(args, "-f", file)
	}
	args = append(args, "config", "--format", "json")

	cmd := exec.Command(d.getDockerCommand(), args...)
	if len(composeFiles) > 0 {
		cmd.Dir = filepath.Dir(composeFiles[0])
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) > 0 {
			return nil, fmt.Errorf("failed to get docker compose configuration: %s\nOutput: %s", err, string(output))
		}
		return nil, fmt.Errorf("failed to get docker compose configuration: %s", err)
	}

	var config ComposeConfig
	if err := json.Unmarshal(output, &config); err != nil {
		return nil, fmt.Errorf("failed to parse docker compose configuration: %w", err)
	}

	d.parsedComposeConfig = &config
	d.cachedComposeFilesKey = currentFilesKey
	return d.parsedComposeConfig, nil
}

func (d *DockerRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	config, err := d.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return "", fmt.Errorf("failed to get compose config: %w", err)
	}

	if service, exists := config.Services[serviceName]; exists {
		return service.Image, nil
	}

	return "", fmt.Errorf("service %s not found in compose configuration", serviceName)
}

// GetMultipleImageInfo returns image information for multiple services from compose files
func (d *DockerRuntime) GetMultipleImageInfo(serviceNames []string, composeFiles []string) (map[string]string, error) {
	if len(serviceNames) == 0 {
		return make(map[string]string), nil
	}

	// Parse compose config once for all services
	config, err := d.getOrParseComposeConfig(composeFiles)
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

// PullImageWithProgress pulls a Docker image and reports progress
func (d *DockerRuntime) PullImageWithProgress(imageName string, progressChan chan<- ImagePullProgress, stopChan <-chan struct{}) error {
	cmd := exec.Command(d.getDockerCommand(), "pull", imageName)

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

		layerProgress := make(map[string]float64)
		var totalLayers int

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Parse Docker pull output
			progress := d.parseDockerPullOutput(line, layerProgress, &totalLayers)
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
			// Try to parse it as well
			layerProgress := make(map[string]float64)
			var totalLayers int
			progress := d.parseDockerPullOutput(line, layerProgress, &totalLayers)
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
				errorMsg += fmt.Sprintf("\nDocker output: %s", errorOutput.String())
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

// parseDockerPullOutput parses Docker pull output and extracts progress information
func (d *DockerRuntime) parseDockerPullOutput(line string, layerProgress map[string]float64, totalLayers *int) ImagePullProgress {
	// Standard Docker pull output format examples:
	// "v3.3.0: Pulling from ankane/blazer"
	// "6e771e15690e: Already exists"
	// "9521bbc382b8: Pulling fs layer"
	// "9521bbc382b8: Pull complete"
	// "Status: Downloaded newer image for ankane/blazer:v3.3.0"

	if strings.Contains(line, "Pulling from") {
		return ImagePullProgress{
			Status: "starting",
		}
	}

	if strings.Contains(line, "Already exists") {
		// Extract layer ID and mark as complete
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			layerID := strings.TrimSpace(parts[0])
			layerProgress[layerID] = 100.0
			*totalLayers = len(layerProgress)
		}

		// Calculate overall progress
		total := 0.0
		for _, p := range layerProgress {
			total += p
		}
		var progress float64
		if len(layerProgress) > 0 {
			progress = total / float64(len(layerProgress))
		}

		return ImagePullProgress{
			Status:       "downloading",
			Progress:     progress,
			TotalLayers:  *totalLayers,
			CurrentLayer: "Layer exists",
		}
	}

	if strings.Contains(line, "Pull complete") {
		// Extract layer ID and mark as complete
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			layerID := strings.TrimSpace(parts[0])
			layerProgress[layerID] = 100.0
			*totalLayers = len(layerProgress)
		}

		// Calculate overall progress
		total := 0.0
		for _, p := range layerProgress {
			total += p
		}
		var progress float64
		if len(layerProgress) > 0 {
			progress = total / float64(len(layerProgress))
		}

		return ImagePullProgress{
			Status:       "downloading",
			Progress:     progress,
			TotalLayers:  *totalLayers,
			CurrentLayer: "Layer complete",
		}
	}

	if strings.Contains(line, "Pulling fs layer") || strings.Contains(line, "Downloading") {
		// Extract layer ID and mark as in progress
		parts := strings.Split(line, ":")
		if len(parts) > 0 {
			layerID := strings.TrimSpace(parts[0])
			layerProgress[layerID] = 50.0 // Mark as in progress
			*totalLayers = len(layerProgress)
		}

		// Calculate overall progress
		total := 0.0
		for _, p := range layerProgress {
			total += p
		}
		var progress float64
		if len(layerProgress) > 0 {
			progress = total / float64(len(layerProgress))
		}

		return ImagePullProgress{
			Status:       "downloading",
			Progress:     progress,
			TotalLayers:  *totalLayers,
			CurrentLayer: "Downloading layer",
		}
	}

	if strings.Contains(line, "Status: Downloaded newer image") || strings.Contains(line, "Status: Image is up to date") {
		return ImagePullProgress{
			Status:   "complete",
			Progress: 100.0,
		}
	}

	return ImagePullProgress{} // Empty progress for unrecognized lines
}

// GetAllContainerStatuses returns all current containers (including stopped ones) managed by compose
// Returns back the container names and statuses in a map
func (d *DockerRuntime) GetAllContainerStatuses() (map[string]string, error) {
	// Use docker ps to get all running containers with compose labels
	// This is much faster than checking each service individually
	cmd := exec.Command(d.getDockerCommand(), "ps", "-a",
		"--filter", "label=com.docker.compose.project=insta",
		"--format", "{{.Names}},{{.State}},{{.Status}}")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get running containers: %w", err)
	}

	// Example output:
	//  ~ % docker ps -a --format "{{.Names}} {{.State}} {{.Status}}"
	// postgres-data,exited,Exited (0) About an hour ago
	// postgres,running,Up About an hour (healthy)
	// jaeger,running,Up About an hour (health: starting)
	// redis,running,Up About an hour

	containerStatuses := make(map[string]string)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ",")
		if len(parts) >= 3 {
			containerName := parts[0]
			state := parts[1]
			status := parts[2] // This is the detailed status of the container

			// Parse detailed health status from the status string
			finalStatus := d.parseDetailedContainerStatus(state, status)
			containerStatuses[containerName] = finalStatus
		}
	}

	return containerStatuses, nil
}

// parseDetailedContainerStatus provides more granular status parsing including health checks
func (d *DockerRuntime) parseDetailedContainerStatus(state, status string) string {
	statusLower := strings.ToLower(status)
	stateLower := strings.ToLower(state)

	// Handle container states first
	if strings.Contains(stateLower, "created") {
		return "created"
	} else if strings.Contains(stateLower, "exited") {
		return "stopped"
	} else if strings.Contains(stateLower, "restarting") {
		return "restarting"
	} else if strings.Contains(stateLower, "paused") {
		return "paused"
	} else if strings.Contains(stateLower, "dead") {
		return "error"
	}

	// For running containers, check health status
	if strings.Contains(stateLower, "running") {
		// Check for health check states in the status
		if strings.Contains(statusLower, "(healthy)") {
			return "running-healthy"
		} else if strings.Contains(statusLower, "(unhealthy)") {
			return "running-unhealthy"
		} else if strings.Contains(statusLower, "(health: starting)") || strings.Contains(statusLower, "health: starting") {
			return "running-health-starting"
		} else if strings.Contains(statusLower, "starting") {
			// Container is starting up
			return "starting"
		} else {
			// Running but no health check configured
			return "running"
		}
	}

	// Default fallback
	return "unknown"
}

func (d *DockerRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
	config, err := d.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return "", err
	}

	// First, check if there's an explicit container name in the compose config
	if serviceConfig, ok := config.Services[serviceName]; ok {
		if serviceConfig.ContainerName != "" {
			return serviceConfig.ContainerName, nil
		}
	}

	// Try different naming patterns in order of preference:
	candidateNames := []string{
		serviceName,                            // Direct service name (e.g., "airflow-init")
		fmt.Sprintf("insta_%s_1", serviceName), // Default compose pattern
		fmt.Sprintf("insta-%s-1", serviceName), // Alternative dash pattern
	}

	// Check which container actually exists
	for _, candidateName := range candidateNames {
		if d.containerExistsAnywhere(candidateName) {
			return candidateName, nil
		}
	}

	// If none exist, return the most likely candidate (service name itself)
	return serviceName, nil
}

// GetAllDependenciesRecursive returns all dependencies recursively for a service from compose files
// Returns container names (if isContainer is true) or service names (if isContainer is false) for UI display purposes
func (d *DockerRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string, isContainer bool) ([]string, error) {
	config, err := d.getOrParseComposeConfig(composeFiles)
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
		containerName, err := d.GetContainerName(serviceName, composeFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to get container name for service %s: %w", serviceName, err)
		}
		containerNames = append(containerNames, containerName)
	}

	return containerNames, nil
}

func (d *DockerRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	args := buildLogCommand(containerName, tailLines, false)
	cmd := exec.Command(d.getDockerCommand(), args...)

	// Capture both stdout and stderr since docker logs writes to stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for container %s: %w", containerName, err)
	}

	return parseLogOutput(output), nil
}

func (d *DockerRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	args := buildLogCommand(containerName, 0, true)
	cmd := exec.Command(d.getDockerCommand(), args...)

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
func (d *DockerRuntime) containerExistsAnywhere(containerName string) bool {
	cmd := exec.Command(d.getDockerCommand(), "ps", "-a", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=^%s$", containerName))
	output, err := cmd.CombinedOutput()
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

// GetContainerStatus returns the status of a container (running, exited, created, etc.)
func (d *DockerRuntime) GetContainerStatus(containerName string) (string, error) {
	cmd := exec.Command(d.getDockerCommand(), "inspect", "--format", "{{.State.Status}} {{.State.ExitCode}}", containerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// If container doesn't exist, return stopped
		if strings.Contains(err.Error(), "No such container") {
			return "stopped", nil
		}
		return "error", fmt.Errorf("failed to get container status: %w", err)
	}

	statusLine := strings.TrimSpace(string(output))
	parts := strings.Split(statusLine, " ")
	if len(parts) < 2 {
		// Fallback if format doesn't match expected
		status := parts[0]
		switch status {
		case "running":
			return "running", nil
		case "exited":
			return "stopped", nil
		case "created":
			return "stopped", nil
		case "restarting":
			return "starting", nil
		case "removing":
			return "stopping", nil
		case "paused":
			return "paused", nil
		case "dead":
			return "error", nil
		default:
			return status, nil
		}
	}

	status := parts[0]
	exitCode := parts[1]

	switch status {
	case "running":
		return "running", nil
	case "exited":
		// Check exit code to differentiate between successful completion and failure
		if exitCode == "0" {
			return "completed", nil
		}
		return "error", nil
	case "created":
		return "stopped", nil
	case "restarting":
		return "starting", nil
	case "removing":
		return "stopping", nil
	case "paused":
		return "paused", nil
	case "dead":
		return "error", nil
	default:
		return status, nil
	}
}

// CheckImageExists checks if a Docker image exists locally
func (d *DockerRuntime) CheckImageExists(imageName string) (bool, error) {
	cmd := exec.Command(d.getDockerCommand(), "image", "inspect", imageName)
	return cmd.Run() == nil, nil
}

// CheckMultipleImagesExist checks if multiple Docker images exist locally in a single call
func (d *DockerRuntime) CheckMultipleImagesExist(imageNames []string) (map[string]bool, error) {
	if len(imageNames) == 0 {
		return make(map[string]bool), nil
	}

	// Use docker images command to get all local images in one call
	cmd := exec.Command(d.getDockerCommand(), "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %w", err)
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

// ListAllImages returns a list of all available Docker images
func (d *DockerRuntime) ListAllImages() ([]string, error) {
	cmd := exec.Command(d.getDockerCommand(), "images", "--format", "{{.Repository}}:{{.Tag}}")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list docker images: %w", err)
	}

	var images []string
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && line != "<none>:<none>" {
			images = append(images, line)
		}
	}

	return images, nil
}
