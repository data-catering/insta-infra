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

func (p *PodmanRuntime) GetDependencies(service string, composeFiles []string) ([]string, error) {
	config, err := p.getOrParseComposeConfig(composeFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to get compose config for dependencies: %w", err)
	}
	return extractDependencies(*config, service), nil
}

func (p *PodmanRuntime) GetAllDependenciesRecursive(serviceName string, composeFiles []string) ([]string, error) {
	// `toProcessQueue` acts as a queue for services whose dependencies need to be fetched.
	// Initialize with direct dependencies of serviceName, as serviceName is not its own dependency.
	toProcessQueue, err := p.GetDependencies(serviceName, composeFiles)
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

		children, err := p.GetDependencies(currentServiceToExplore, composeFiles)
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

func (p *PodmanRuntime) GetContainerName(serviceName string, composeFiles []string) (string, error) {
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

func (p *PodmanRuntime) GetContainerLogs(containerName string, tailLines int) ([]string, error) {
	args := []string{"logs", "--tail", fmt.Sprintf("%d", tailLines), containerName}
	cmd := exec.Command(p.getPodmanCommand(), args...)

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

func (p *PodmanRuntime) StreamContainerLogs(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
	args := []string{"logs", "--follow", "--tail", "50", containerName}
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

func (p *PodmanRuntime) GetImageInfo(serviceName string, composeFiles []string) (string, error) {
	config, err := p.getOrParseComposeConfig(composeFiles)
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
