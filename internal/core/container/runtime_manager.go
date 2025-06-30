package container

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

// RuntimeManager provides functionality to manage container runtime services
type RuntimeManager struct{}

// NewRuntimeManager creates a new runtime manager
func NewRuntimeManager() *RuntimeManager {
	return &RuntimeManager{}
}

// StartupResult represents the result of attempting to start a runtime service
type StartupResult struct {
	Success              bool   `json:"success"`
	Message              string `json:"message"`
	Error                string `json:"error"`
	RequiresManualAction bool   `json:"requiresManualAction"`
}

// getDockerPath returns the resolved Docker binary path
func getDockerPath() string {
	// Check for custom path first
	if customPath := os.Getenv("INSTA_DOCKER_PATH"); customPath != "" {
		return customPath
	}

	// Try PATH first
	if path, err := exec.LookPath("docker"); err == nil {
		return path
	}

	// Try common paths
	commonPaths := []string{
		"/opt/homebrew/bin/docker",            // Homebrew on Apple Silicon
		"/usr/local/bin/docker",               // Homebrew on Intel Mac
		"/usr/bin/docker",                     // System package
		"/snap/bin/docker",                    // Snap package
		"/var/lib/flatpak/exports/bin/docker", // Flatpak
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "docker" // Fallback
}

// getPodmanPath returns the resolved Podman binary path
func getPodmanPath() string {
	// Check for custom path first
	if customPath := os.Getenv("INSTA_PODMAN_PATH"); customPath != "" {
		return customPath
	}

	// Try PATH first
	if path, err := exec.LookPath("podman"); err == nil {
		return path
	}

	// Try common paths
	commonPaths := []string{
		"/opt/homebrew/bin/podman",            // Homebrew on Apple Silicon
		"/usr/local/bin/podman",               // Homebrew on Intel Mac
		"/usr/bin/podman",                     // System package
		"/snap/bin/podman",                    // Snap package
		"/var/lib/flatpak/exports/bin/podman", // Flatpak
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return "podman" // Fallback
}

// AttemptStartRuntime tries to start the specified container runtime
func (rm *RuntimeManager) AttemptStartRuntime(runtimeName string) *StartupResult {
	switch runtimeName {
	case "docker":
		return rm.startDocker()
	case "podman":
		return rm.startPodman()
	default:
		return &StartupResult{
			Success: false,
			Error:   fmt.Sprintf("Unknown runtime: %s", runtimeName),
		}
	}
}

// startDocker attempts to start Docker service
func (rm *RuntimeManager) startDocker() *StartupResult {
	switch runtime.GOOS {
	case "darwin":
		return rm.startDockerMacOS()
	case "linux":
		return rm.startDockerLinux()
	case "windows":
		return rm.startDockerWindows()
	default:
		return &StartupResult{
			Success:              false,
			Error:                "Docker startup not supported on this platform",
			RequiresManualAction: true,
		}
	}
}

// startPodman attempts to start Podman service
func (rm *RuntimeManager) startPodman() *StartupResult {
	switch runtime.GOOS {
	case "darwin":
		return rm.startPodmanMacOS()
	case "linux":
		return rm.startPodmanLinux()
	case "windows":
		return rm.startPodmanWindows()
	default:
		return &StartupResult{
			Success:              false,
			Error:                "Podman startup not supported on this platform",
			RequiresManualAction: true,
		}
	}
}

// Platform-specific Docker startup implementations

func (rm *RuntimeManager) startDockerMacOS() *StartupResult {
	// Try to start Docker Desktop
	cmd := exec.Command("open", "-a", "Docker")
	if err := cmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to start Docker Desktop: %v", err),
			Message:              "Please start Docker Desktop manually from Applications",
			RequiresManualAction: true,
		}
	}

	// Wait a bit and check if Docker is starting
	time.Sleep(3 * time.Second)

	// Check if Docker is responding (it may take time to fully start)
	dockerPath := getDockerPath()
	checkCmd := exec.Command(dockerPath, "info")
	if err := checkCmd.Run(); err == nil {
		return &StartupResult{
			Success: true,
			Message: "Docker Desktop started successfully",
		}
	}

	return &StartupResult{
		Success:              true,
		Message:              "Docker Desktop is starting up. This may take a minute or two.",
		RequiresManualAction: false,
	}
}

func (rm *RuntimeManager) startDockerLinux() *StartupResult {
	// Try to start Docker service using systemctl
	cmd := exec.Command("sudo", "systemctl", "start", "docker")
	if err := cmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to start Docker service: %v", err),
			Message:              "Please run 'sudo systemctl start docker' manually",
			RequiresManualAction: true,
		}
	}

	// Wait a moment and verify
	time.Sleep(2 * time.Second)
	dockerPath := getDockerPath()
	checkCmd := exec.Command(dockerPath, "info")
	if err := checkCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                "Docker service started but not responding",
			Message:              "Docker service may still be starting up",
			RequiresManualAction: false,
		}
	}

	return &StartupResult{
		Success: true,
		Message: "Docker service started successfully",
	}
}

func (rm *RuntimeManager) startDockerWindows() *StartupResult {
	// On Windows, we need to start Docker Desktop
	// Try using PowerShell to start Docker Desktop
	cmd := exec.Command("powershell", "-Command", "Start-Process", "'Docker Desktop'")
	if err := cmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to start Docker Desktop: %v", err),
			Message:              "Please start Docker Desktop manually from the Start menu",
			RequiresManualAction: true,
		}
	}

	return &StartupResult{
		Success:              true,
		Message:              "Docker Desktop is starting up. This may take a minute or two.",
		RequiresManualAction: false,
	}
}

// Platform-specific Podman startup implementations

func (rm *RuntimeManager) startPodmanMacOS() *StartupResult {
	// On macOS, Podman requires a machine to be started
	podmanPath := getPodmanPath()
	cmd := exec.Command(podmanPath, "machine", "start")
	if err := cmd.Run(); err != nil {
		// Try to initialize machine first if it doesn't exist
		initCmd := exec.Command(podmanPath, "machine", "init")
		if initErr := initCmd.Run(); initErr != nil {
			return &StartupResult{
				Success:              false,
				Error:                fmt.Sprintf("Failed to initialize Podman machine: %v", initErr),
				Message:              "Please run 'podman machine init' and 'podman machine start' manually",
				RequiresManualAction: true,
			}
		}

		// Try starting again after initialization
		if err := cmd.Run(); err != nil {
			return &StartupResult{
				Success:              false,
				Error:                fmt.Sprintf("Failed to start Podman machine: %v", err),
				Message:              "Please run 'podman machine start' manually",
				RequiresManualAction: true,
			}
		}
	}

	// Wait a moment and verify
	time.Sleep(3 * time.Second)
	checkCmd := exec.Command(podmanPath, "info")
	if err := checkCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                "Podman machine started but not responding",
			Message:              "Podman machine may still be starting up",
			RequiresManualAction: false,
		}
	}

	return &StartupResult{
		Success: true,
		Message: "Podman machine started successfully",
	}
}

func (rm *RuntimeManager) startPodmanLinux() *StartupResult {
	// Try to start Podman service using systemctl
	cmd := exec.Command("sudo", "systemctl", "start", "podman")
	if err := cmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to start Podman service: %v", err),
			Message:              "Please run 'sudo systemctl start podman' manually",
			RequiresManualAction: true,
		}
	}

	// Wait a moment and verify
	time.Sleep(2 * time.Second)
	podmanPath := getPodmanPath()
	checkCmd := exec.Command(podmanPath, "info")
	if err := checkCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                "Podman service started but not responding",
			Message:              "Podman service may still be starting up",
			RequiresManualAction: false,
		}
	}

	return &StartupResult{
		Success: true,
		Message: "Podman service started successfully",
	}
}

func (rm *RuntimeManager) startPodmanWindows() *StartupResult {
	// Podman on Windows is more complex and typically requires manual setup
	return &StartupResult{
		Success:              false,
		Error:                "Automatic Podman startup not supported on Windows",
		Message:              "Please refer to Podman Windows documentation for setup instructions",
		RequiresManualAction: true,
	}
}

// WaitForRuntimeReady waits for a runtime to become fully available
func (rm *RuntimeManager) WaitForRuntimeReady(runtimeName string, maxWaitSeconds int) *StartupResult {
	for i := 0; i < maxWaitSeconds; i++ {
		status := GetDetailedRuntimeStatus()
		for _, runtimeStatus := range status.RuntimeStatuses {
			if runtimeStatus.Name == runtimeName && runtimeStatus.IsAvailable {
				return &StartupResult{
					Success: true,
					Message: fmt.Sprintf("%s is ready", runtimeName),
				}
			}
		}
		time.Sleep(1 * time.Second)
	}

	return &StartupResult{
		Success:              false,
		Error:                fmt.Sprintf("%s did not become ready within %d seconds", runtimeName, maxWaitSeconds),
		RequiresManualAction: true,
	}
}

// RestartRuntime attempts to restart the specified container runtime
func (rm *RuntimeManager) RestartRuntime(runtimeName string) *StartupResult {
	switch runtimeName {
	case "docker":
		return rm.restartDocker()
	case "podman":
		return rm.restartPodman()
	default:
		return &StartupResult{
			Success: false,
			Error:   fmt.Sprintf("Unknown runtime: %s", runtimeName),
		}
	}
}

// restartDocker attempts to restart Docker service
func (rm *RuntimeManager) restartDocker() *StartupResult {
	switch runtime.GOOS {
	case "darwin":
		return rm.restartDockerMacOS()
	case "linux":
		return rm.restartDockerLinux()
	case "windows":
		return rm.restartDockerWindows()
	default:
		return &StartupResult{
			Success:              false,
			Error:                "Docker restart not supported on this platform",
			RequiresManualAction: true,
		}
	}
}

// restartPodman attempts to restart Podman service
func (rm *RuntimeManager) restartPodman() *StartupResult {
	switch runtime.GOOS {
	case "darwin":
		return rm.restartPodmanMacOS()
	case "linux":
		return rm.restartPodmanLinux()
	case "windows":
		return rm.restartPodmanWindows()
	default:
		return &StartupResult{
			Success:              false,
			Error:                "Podman restart not supported on this platform",
			RequiresManualAction: true,
		}
	}
}

// Platform-specific Docker restart implementations

func (rm *RuntimeManager) restartDockerMacOS() *StartupResult {
	// On macOS, we need to quit and restart Docker Desktop
	// First try to quit Docker Desktop gracefully
	quitCmd := exec.Command("osascript", "-e", "quit app \"Docker Desktop\"")
	quitCmd.Run() // Ignore errors as Docker might not be running

	// Wait a moment for Docker to shut down
	time.Sleep(3 * time.Second)

	// Now start Docker Desktop again
	cmd := exec.Command("open", "-a", "Docker")
	if err := cmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to restart Docker Desktop: %v", err),
			Message:              "Please restart Docker Desktop manually from Applications",
			RequiresManualAction: true,
		}
	}

	// Wait for Docker to start up
	time.Sleep(5 * time.Second)

	return &StartupResult{
		Success: true,
		Message: "Docker Desktop restarted successfully. It may take a minute to fully initialize.",
	}
}

func (rm *RuntimeManager) restartDockerLinux() *StartupResult {
	// On Linux, restart the Docker service using systemctl
	restartCmd := exec.Command("sudo", "systemctl", "restart", "docker")
	if err := restartCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to restart Docker service: %v", err),
			Message:              "Please run 'sudo systemctl restart docker' manually",
			RequiresManualAction: true,
		}
	}

	// Wait a moment for the service to restart
	time.Sleep(3 * time.Second)

	// Verify Docker is responding
	dockerPath := getDockerPath()
	checkCmd := exec.Command(dockerPath, "info")
	if err := checkCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                "Docker service restarted but not responding",
			Message:              "Docker service may still be starting up",
			RequiresManualAction: false,
		}
	}

	return &StartupResult{
		Success: true,
		Message: "Docker service restarted successfully",
	}
}

func (rm *RuntimeManager) restartDockerWindows() *StartupResult {
	// On Windows, we need to restart Docker Desktop
	// Try to stop Docker Desktop first
	stopCmd := exec.Command("taskkill", "/F", "/IM", "Docker Desktop.exe")
	stopCmd.Run() // Ignore errors

	// Wait for shutdown
	time.Sleep(3 * time.Second)

	// Start Docker Desktop again
	cmd := exec.Command("powershell", "-Command", "Start-Process", "'Docker Desktop'")
	if err := cmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to restart Docker Desktop: %v", err),
			Message:              "Please restart Docker Desktop manually from the Start menu",
			RequiresManualAction: true,
		}
	}

	return &StartupResult{
		Success:              true,
		Message:              "Docker Desktop is restarting. This may take a minute or two.",
		RequiresManualAction: false,
	}
}

// Platform-specific Podman restart implementations

func (rm *RuntimeManager) restartPodmanMacOS() *StartupResult {
	// On macOS, restart the Podman machine
	stopCmd := exec.Command(getPodmanPath(), "machine", "stop")
	stopCmd.Run() // Ignore errors

	// Wait for shutdown
	time.Sleep(2 * time.Second)

	// Start the machine again
	startCmd := exec.Command(getPodmanPath(), "machine", "start")
	if err := startCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to restart Podman machine: %v", err),
			Message:              "Please run 'podman machine restart' manually",
			RequiresManualAction: true,
		}
	}

	return &StartupResult{
		Success: true,
		Message: "Podman machine restarted successfully",
	}
}

func (rm *RuntimeManager) restartPodmanLinux() *StartupResult {
	// On Linux, restart the Podman service using systemctl
	restartCmd := exec.Command("sudo", "systemctl", "restart", "podman")
	if err := restartCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                fmt.Sprintf("Failed to restart Podman service: %v", err),
			Message:              "Please run 'sudo systemctl restart podman' manually",
			RequiresManualAction: true,
		}
	}

	// Wait a moment for the service to restart
	time.Sleep(2 * time.Second)

	// Verify Podman is responding
	podmanPath := getPodmanPath()
	checkCmd := exec.Command(podmanPath, "info")
	if err := checkCmd.Run(); err != nil {
		return &StartupResult{
			Success:              false,
			Error:                "Podman service restarted but not responding",
			Message:              "Podman service may still be starting up",
			RequiresManualAction: false,
		}
	}

	return &StartupResult{
		Success: true,
		Message: "Podman service restarted successfully",
	}
}

func (rm *RuntimeManager) restartPodmanWindows() *StartupResult {
	// Podman on Windows is more complex and typically requires manual intervention
	return &StartupResult{
		Success:              false,
		Error:                "Automatic Podman restart not supported on Windows",
		Message:              "Please restart Podman manually",
		RequiresManualAction: true,
	}
}
