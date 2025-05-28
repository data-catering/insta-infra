package container

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// RuntimeStatus represents the detailed status of a container runtime
type RuntimeStatus struct {
	Name              string `json:"name"`              // "docker" or "podman"
	IsInstalled       bool   `json:"isInstalled"`       // Binary exists in PATH
	IsRunning         bool   `json:"isRunning"`         // Daemon/service is running
	IsAvailable       bool   `json:"isAvailable"`       // Fully functional (installed + running + compose)
	HasCompose        bool   `json:"hasCompose"`        // Compose plugin/tool available
	Version           string `json:"version"`           // Version string if available
	Error             string `json:"error"`             // Error message if any
	CanAutoStart      bool   `json:"canAutoStart"`      // Can we attempt to start the service
	InstallationGuide string `json:"installationGuide"` // Platform-specific installation instructions
	StartupCommand    string `json:"startupCommand"`    // Command to start the service (if applicable)
	RequiresMachine   bool   `json:"requiresMachine"`   // True for Podman on macOS
	MachineStatus     string `json:"machineStatus"`     // Status of podman machine (if applicable)
}

// SystemRuntimeStatus represents the overall container runtime situation
type SystemRuntimeStatus struct {
	HasAnyRuntime     bool            `json:"hasAnyRuntime"`     // At least one runtime is available
	PreferredRuntime  string          `json:"preferredRuntime"`  // Which runtime to recommend
	AvailableRuntimes []string        `json:"availableRuntimes"` // List of available runtime names
	RuntimeStatuses   []RuntimeStatus `json:"runtimeStatuses"`   // Detailed status for each runtime
	RecommendedAction string          `json:"recommendedAction"` // What the user should do next
	CanProceed        bool            `json:"canProceed"`        // Whether insta-infra can function
	Platform          string          `json:"platform"`          // Operating system
}

// GetDetailedRuntimeStatus provides comprehensive runtime status information
func GetDetailedRuntimeStatus() *SystemRuntimeStatus {
	status := &SystemRuntimeStatus{
		Platform:          runtime.GOOS,
		RuntimeStatuses:   make([]RuntimeStatus, 0, 2),
		AvailableRuntimes: make([]string, 0, 2),
	}

	// Check Docker
	dockerStatus := checkDockerStatus()
	status.RuntimeStatuses = append(status.RuntimeStatuses, dockerStatus)
	if dockerStatus.IsAvailable {
		status.AvailableRuntimes = append(status.AvailableRuntimes, "docker")
		status.HasAnyRuntime = true
		if status.PreferredRuntime == "" {
			status.PreferredRuntime = "docker"
		}
	}

	// Check Podman
	podmanStatus := checkPodmanStatus()
	status.RuntimeStatuses = append(status.RuntimeStatuses, podmanStatus)
	if podmanStatus.IsAvailable {
		status.AvailableRuntimes = append(status.AvailableRuntimes, "podman")
		status.HasAnyRuntime = true
		if status.PreferredRuntime == "" {
			status.PreferredRuntime = "podman"
		}
	}

	// Determine recommended action
	status.CanProceed = status.HasAnyRuntime
	if status.HasAnyRuntime {
		status.RecommendedAction = fmt.Sprintf("Ready to use %s", status.PreferredRuntime)
	} else {
		status.RecommendedAction = determineRecommendedAction(status.RuntimeStatuses)
	}

	return status
}

// checkDockerStatus provides detailed Docker status
func checkDockerStatus() RuntimeStatus {
	status := RuntimeStatus{
		Name:              "docker",
		InstallationGuide: getDockerInstallationGuide(),
	}

	// Check if Docker binary exists using enhanced detection
	dockerPath := findBinaryInCommonPaths("docker", getCommonDockerPaths())
	if dockerPath == "" {
		status.Error = "Docker not installed"
		return status
	}
	status.IsInstalled = true

	// Check if Docker daemon is running
	cmd := exec.Command(dockerPath, "info")
	if err := cmd.Run(); err != nil {
		status.Error = "Docker daemon not running"
		status.CanAutoStart = canStartDockerService()
		status.StartupCommand = getDockerStartupCommand()
		return status
	}
	status.IsRunning = true

	// Get Docker version
	if versionCmd := exec.Command(dockerPath, "version", "--format", "{{.Server.Version}}"); versionCmd != nil {
		if output, err := versionCmd.Output(); err == nil {
			status.Version = strings.TrimSpace(string(output))
		}
	}

	// Check Docker Compose
	cmd = exec.Command(dockerPath, "compose", "version")
	if err := cmd.Run(); err != nil {
		status.Error = "Docker Compose plugin not available"
		return status
	}
	status.HasCompose = true
	status.IsAvailable = true

	return status
}

// checkPodmanStatus provides detailed Podman status
func checkPodmanStatus() RuntimeStatus {
	status := RuntimeStatus{
		Name:              "podman",
		InstallationGuide: getPodmanInstallationGuide(),
		RequiresMachine:   runtime.GOOS == "darwin", // macOS requires podman machine
	}

	// Check if Podman binary exists using enhanced detection
	podmanPath := findBinaryInCommonPaths("podman", getCommonPodmanPaths())
	if podmanPath == "" {
		status.Error = "Podman not installed"
		return status
	}
	status.IsInstalled = true

	// Get Podman version
	if versionCmd := exec.Command(podmanPath, "version", "--format", "{{.Version}}"); versionCmd != nil {
		if output, err := versionCmd.Output(); err == nil {
			status.Version = strings.TrimSpace(string(output))
		}
	}

	// Check if Podman machine is running (macOS)
	if status.RequiresMachine {
		machineCmd := exec.Command(podmanPath, "machine", "list", "--format", "{{.Name}} {{.Running}}")
		output, err := machineCmd.Output()
		if err != nil {
			status.Error = "Failed to check Podman machine status"
			status.MachineStatus = "unknown"
			status.CanAutoStart = true
			status.StartupCommand = "podman machine start"
			return status
		}

		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		hasRunningMachine := false

		for _, line := range lines {
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[1] == "true" {
				hasRunningMachine = true
				break
			}
		}

		if !hasRunningMachine {
			status.Error = "Podman machine not running"
			status.MachineStatus = "stopped"
			status.CanAutoStart = true
			status.StartupCommand = "podman machine start"
			return status
		}
		status.MachineStatus = "running"
	}

	// Check if Podman daemon is accessible
	cmd := exec.Command(podmanPath, "info")
	if err := cmd.Run(); err != nil {
		status.Error = "Podman not accessible"
		if !status.RequiresMachine {
			status.CanAutoStart = canStartPodmanService()
			status.StartupCommand = getPodmanStartupCommand()
		}
		return status
	}
	status.IsRunning = true

	// Check Podman Compose
	cmd = exec.Command(podmanPath, "compose", "version")
	if err := cmd.Run(); err != nil {
		// Try podman-compose as fallback
		if _, fallbackErr := exec.LookPath("podman-compose"); fallbackErr != nil {
			status.Error = "Neither podman compose plugin nor podman-compose found"
			return status
		}
	}
	status.HasCompose = true
	status.IsAvailable = true

	return status
}

// Platform-specific installation guides
func getDockerInstallationGuide() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Docker Desktop from https://docs.docker.com/desktop/install/mac-install/ or use Homebrew: brew install --cask docker"
	case "linux":
		return "Install Docker Engine from https://docs.docker.com/engine/install/ or use your package manager"
	case "windows":
		return "Install Docker Desktop from https://docs.docker.com/desktop/install/windows-install/"
	default:
		return "Visit https://docs.docker.com/get-docker/ for installation instructions"
	}
}

func getPodmanInstallationGuide() string {
	switch runtime.GOOS {
	case "darwin":
		return "Install Podman using Homebrew: brew install podman, then run 'podman machine init' and 'podman machine start'"
	case "linux":
		return "Install Podman from https://podman.io/getting-started/installation or use your package manager"
	case "windows":
		return "Install Podman from https://github.com/containers/podman/blob/main/docs/tutorials/podman-for-windows.md"
	default:
		return "Visit https://podman.io/getting-started/installation for installation instructions"
	}
}

// Platform-specific service startup capabilities
func canStartDockerService() bool {
	switch runtime.GOOS {
	case "darwin":
		// On macOS, Docker Desktop needs to be started manually or via open command
		return true
	case "linux":
		// On Linux, we can try to start the systemd service
		return true
	case "windows":
		// On Windows, Docker Desktop needs to be started manually
		return true
	default:
		return false
	}
}

func canStartPodmanService() bool {
	switch runtime.GOOS {
	case "linux":
		// On Linux, we can try to start the systemd service
		return true
	default:
		return false
	}
}

func getDockerStartupCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "open -a Docker"
	case "linux":
		return "sudo systemctl start docker"
	case "windows":
		return "Start Docker Desktop from the Start menu"
	default:
		return "Start Docker service"
	}
}

func getPodmanStartupCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "podman machine start"
	case "linux":
		return "sudo systemctl start podman"
	default:
		return "Start Podman service"
	}
}

func determineRecommendedAction(statuses []RuntimeStatus) string {
	// Check if any runtime is installed but not running
	for _, status := range statuses {
		if status.IsInstalled && !status.IsRunning && status.CanAutoStart {
			return fmt.Sprintf("Start %s service", status.Name)
		}
	}

	// Check if any runtime is partially set up
	for _, status := range statuses {
		if status.IsInstalled {
			return fmt.Sprintf("Configure %s", status.Name)
		}
	}

	// No runtime installed - recommend based on platform
	switch runtime.GOOS {
	case "darwin":
		return "Install Docker Desktop (recommended) or Podman"
	case "linux":
		return "Install Docker Engine or Podman"
	case "windows":
		return "Install Docker Desktop"
	default:
		return "Install Docker or Podman"
	}
}
