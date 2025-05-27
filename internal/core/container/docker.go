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

	cmd := exec.Command("docker", args...)
	if len(composeFiles) > 0 {
		cmd.Dir = filepath.Dir(composeFiles[0])
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("failed to get docker compose configuration: %s\n%s", err, string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to get docker compose configuration: %w", err)
	}

	var config ComposeConfig
	if err := json.Unmarshal(output, &config); err != nil {
		return nil, fmt.Errorf("failed to parse docker compose configuration: %w", err)
	}

	d.parsedComposeConfig = &config
	d.cachedComposeFilesKey = currentFilesKey
	return d.parsedComposeConfig, nil
}

func (d *DockerRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	config, err := d.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get compose config for dependencies: %w", err)
	}
	return extractDependencies(*config, service), nil
}

func (d *DockerRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string) ([]string, error) {
	// `toProcessQueue` acts as a queue for services whose dependencies need to be fetched.
	// Initialize with direct dependencies of serviceName, as serviceName is not its own dependency.
	toProcessQueue, err := d.GetDependencies(serviceName, composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial dependencies for %s: %w", serviceName, err)
	}

	// `allFoundDependencies` will store all unique dependencies discovered.
	allFoundDependencies := make(map[string]bool)
	// Add initial direct dependencies to the set and prepare the queue.
	// We use a new slice for the queue to avoid modifying the original direct deps slice if it's from a cache.
	queue := make([]string, 0, len(toProcessQueue))
	for _, dep := range toProcessQueue {
		if !allFoundDependencies[dep] { // Ensure initial deps are unique if GetDependencies somehow returns duplicates
			allFoundDependencies[dep] = true
			queue = append(queue, dep)
		}
	}

	head := 0
	for head < len(queue) {
		currentServiceToExplore := queue[head]
		head++

		children, err := d.GetDependencies(currentServiceToExplore, composeFiles)
		if err != nil {
			// If fetching dependencies for a sub-service fails, we might have an incomplete list.
			// For now, we'll propagate the error. Consider if partial results are acceptable.
			return nil, fmt.Errorf("failed to get dependencies for intermediate service %s during recursive search: %w", currentServiceToExplore, err)
		}

		for _, child := range children {
			if !allFoundDependencies[child] { // If this is a newly discovered dependency
				allFoundDependencies[child] = true
				queue = append(queue, child) // Add it to the queue to process its dependencies later.
			}
		}
	}

	result := make([]string, 0, len(allFoundDependencies))
	for dep := range allFoundDependencies {
		result = append(result, dep)
	}
	// Sorting is handled in app.go, so not strictly needed here.
	return result, nil
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

func (d *DockerRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	args := []string{"logs", "--tail", fmt.Sprintf("%d", tailLines), containerName}
	cmd := exec.Command("docker", args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get logs for container %s: %w", containerName, err)
	}

	// Split output into individual log lines
	logLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(logLines) == 1 && logLines[0] == "" {
		return []string{}, nil // Return empty slice for empty logs
	}

	return logLines, nil
}

func (d *DockerRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	args := []string{"logs", "--follow", "--tail", "50", containerName}
	cmd := exec.Command("docker", args...)

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

	// Start goroutine to read from stdout
	go func() {
		defer func() {
			stdout.Close()
		}()

		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				// Split by lines and send each line
				lines := strings.Split(strings.TrimSpace(string(buf[:n])), "\n")
				for _, line := range lines {
					if line != "" {
						select {
						case logChan <- line:
						case <-stopChan:
							return
						}
					}
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Start goroutine to read from stderr
	go func() {
		defer func() {
			stderr.Close()
		}()

		buf := make([]byte, 1024)
		for {
			n, err := stderr.Read(buf)
			if n > 0 {
				// Split by lines and send each line
				lines := strings.Split(strings.TrimSpace(string(buf[:n])), "\n")
				for _, line := range lines {
					if line != "" {
						select {
						case logChan <- line:
						case <-stopChan:
							return
						}
					}
				}
			}
			if err != nil {
				return
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
		return nil
	case err := <-doneChan:
		return err
	}
}

// containerExists checks if a container with the given name exists and is running
func (d *DockerRuntime) containerExists(containerName string) bool {
	cmd := exec.Command("docker", "ps", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=^%s$", containerName))
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

// containerExistsAnywhere checks if a container with the given name exists (running or stopped)
func (d *DockerRuntime) containerExistsAnywhere(containerName string) bool {
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}", "--filter", fmt.Sprintf("name=^%s$", containerName))
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

// GetContainerStatus returns the status of a container (running, exited, created, etc.)
func (d *DockerRuntime) GetContainerStatus(containerName string) (string, error) {
	// Check all containers (including stopped ones)
	cmd := exec.Command("docker", "ps", "-a", "--format", "{{.Names}}\t{{.Status}}", "--filter", fmt.Sprintf("name=^%s$", containerName))
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

			// Normalize Docker status to our standard status values
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

// CheckImageExists checks if a Docker image exists locally
func (d *DockerRuntime) CheckImageExists(imageName string) (bool, error) {
	cmd := exec.Command("docker", "image", "inspect", imageName)
	return cmd.Run() == nil, nil
}

// GetImageInfo returns the image name for a service from compose files
func (d *DockerRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	config, err := d.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return "", fmt.Errorf("failed to get compose config for image info: %w", err)
	}

	service, exists := config.Services[serviceName]
	if !exists {
		return "", fmt.Errorf("service %s not found in compose files", serviceName)
	}

	if service.Image == "" {
		return "", fmt.Errorf("no image specified for service %s", serviceName)
	}

	return service.Image, nil
}

// PullImageWithProgress pulls a Docker image and reports progress
func (d *DockerRuntime) PullImageWithProgress(imageName string, progressChan chan<- ImagePullProgress, stopChan <-chan struct{}) error {
	cmd := exec.Command("docker", "pull", imageName)

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
