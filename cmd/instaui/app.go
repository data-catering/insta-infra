package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/handlers"
	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// Version information - these will be set during build via ldflags
var (
	version   = "dev"
	buildTime = "unknown"
)

//go:embed resources/docker-compose.yaml resources/docker-compose-persist.yaml
//go:embed all:resources/data
var embedFS embed.FS

// App struct - Maintains Wails compatibility
type App struct {
	ctx              context.Context
	containerRuntime container.Runtime
	instaDir         string
	runtimeInitError error
	bundledCLIPath   string // Path to bundled CLI binary

	// Handler instances for delegating business logic
	serviceHandler    handlers.ServiceHandlerInterface
	connectionHandler *handlers.ConnectionHandler
	logsHandler       *handlers.LogsHandler
	imageHandler      *handlers.ImageHandler
	dependencyHandler *handlers.DependencyHandler
	graphHandler      *handlers.GraphHandler

	// Simple logging for debugging
	logMutex sync.RWMutex
	logLines []string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		logLines: make([]string, 0, 1000), // Keep last 1000 log lines
	}
}

// getBundledCLIPath returns the path to the bundled CLI binary if it exists
func (a *App) getBundledCLIPath() string {
	if a.bundledCLIPath != "" {
		return a.bundledCLIPath
	}

	// Get the executable path
	execPath, err := os.Executable()
	if err != nil {
		a.logMessage(fmt.Sprintf("Failed to get executable path: %v", err))
		return ""
	}

	var cliPath string

	if runtime.GOOS == "darwin" {
		// On macOS, check if we're inside an app bundle
		if strings.Contains(execPath, ".app/Contents/MacOS/") {
			// We're in an app bundle, look for the CLI in Resources
			appBundle := strings.Split(execPath, ".app/Contents/MacOS/")[0] + ".app"
			cliPath = filepath.Join(appBundle, "Contents", "Resources", "insta-cli")
		}
	} else if runtime.GOOS == "linux" {
		// On Linux, look for CLI binary alongside the main executable
		execDir := filepath.Dir(execPath)
		cliPath = filepath.Join(execDir, "insta-cli")
	} else if runtime.GOOS == "windows" {
		// On Windows, look for CLI binary alongside the main executable
		execDir := filepath.Dir(execPath)
		cliPath = filepath.Join(execDir, "insta-cli.exe")
	}

	// Check if the CLI binary exists
	if cliPath != "" {
		if _, err := os.Stat(cliPath); err == nil {
			a.bundledCLIPath = cliPath
			a.logMessage(fmt.Sprintf("Found bundled CLI binary at: %s", cliPath))
			return cliPath
		}
	}

	a.logMessage("No bundled CLI binary found, will use system container runtime")
	return ""
}

// executeCLICommand executes a command using the bundled CLI binary
func (a *App) executeCLICommand(args ...string) ([]byte, error) {
	cliPath := a.getBundledCLIPath()
	if cliPath == "" {
		return nil, fmt.Errorf("bundled CLI binary not available")
	}

	cmd := exec.Command(cliPath, args...)

	// Set up environment for the CLI
	cmd.Env = os.Environ()

	// Add common paths for Docker/Podman
	if runtime.GOOS == "darwin" {
		// Use the wrapper script on macOS for better environment setup
		wrapperPath := filepath.Join(filepath.Dir(cliPath), "insta-wrapper.sh")
		if _, err := os.Stat(wrapperPath); err == nil {
			cmd = exec.Command(wrapperPath, args...)
			cmd.Env = os.Environ()
		}
	}

	return cmd.Output()
}

// logMessage adds a message to the internal log buffer
func (a *App) logMessage(message string) {
	a.logMutex.Lock()
	defer a.logMutex.Unlock()

	// Keep only last 1000 lines
	if len(a.logLines) >= 1000 {
		a.logLines = a.logLines[1:]
	}

	a.logLines = append(a.logLines, message)
	fmt.Println(message) // Also print to console
}

// GetAppLogs returns the internal application logs for debugging
func (a *App) GetAppLogs() []string {
	a.logMutex.RLock()
	defer a.logMutex.RUnlock()

	// Return a copy of the log lines
	logs := make([]string, len(a.logLines))
	copy(logs, a.logLines)
	return logs
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logMessage("Starting insta-infra UI application...")

	// Check for bundled CLI first
	bundledCLI := a.getBundledCLIPath()
	if bundledCLI != "" {
		a.logMessage(fmt.Sprintf("Using bundled CLI binary: %s", bundledCLI))
		// When using bundled CLI, we'll delegate container operations to it
		// rather than trying to access container runtime directly
		a.initializeForBundledCLI()
		return
	}

	// Fallback to direct container runtime access (development mode)
	a.logMessage("No bundled CLI found, using direct container runtime access")
	a.initializeForDirectRuntime()
}

// initializeForBundledCLI sets up the app to use the bundled CLI binary
func (a *App) initializeForBundledCLI() {
	// Determine instaDir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		a.logMessage(fmt.Sprintf("Error getting home directory for UI: %v", err))
		a.runtimeInitError = fmt.Errorf("failed to get home directory: %w", err)
		return
	}

	instaDir := os.Getenv("INSTA_HOME")
	if instaDir == "" {
		a.instaDir = filepath.Join(homeDir, ".insta")
	} else {
		a.instaDir = instaDir
	}
	a.logMessage(fmt.Sprintf("UI insta directory: %s", a.instaDir))

	// Extract Docker Compose files if needed
	if err := a.ensureComposeFiles(); err != nil {
		a.logMessage(fmt.Sprintf("Error extracting compose files: %v", err))
		a.runtimeInitError = fmt.Errorf("failed to extract compose files: %w", err)
		return
	}

	// Test the bundled CLI
	output, err := a.executeCLICommand("--version")
	if err != nil {
		a.logMessage(fmt.Sprintf("Error testing bundled CLI: %v", err))
		a.runtimeInitError = fmt.Errorf("bundled CLI not working: %w", err)
		return
	}

	a.logMessage(fmt.Sprintf("Bundled CLI version: %s", string(output)))
	a.runtimeInitError = nil

	// Initialize handlers for bundled CLI mode
	a.logMessage("Initializing handlers for bundled CLI mode...")
	a.initializeHandlersForCLI()
}

// initializeForDirectRuntime sets up the app to use direct container runtime access
func (a *App) initializeForDirectRuntime() {
	// Determine instaDir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		a.logMessage(fmt.Sprintf("Error getting home directory for UI: %v", err))
		a.runtimeInitError = fmt.Errorf("failed to get home directory: %w", err)
		return
	}

	instaDir := os.Getenv("INSTA_HOME")
	if instaDir == "" {
		a.instaDir = filepath.Join(homeDir, ".insta")
	} else {
		a.instaDir = instaDir
	}
	a.logMessage(fmt.Sprintf("UI insta directory: %s", a.instaDir))

	// Extract Docker Compose files if needed
	if err := a.ensureComposeFiles(); err != nil {
		a.logMessage(fmt.Sprintf("Error extracting compose files: %v", err))
		a.runtimeInitError = fmt.Errorf("failed to extract compose files: %w", err)
		return
	}

	// Log custom paths if set
	if dockerPath := os.Getenv("INSTA_DOCKER_PATH"); dockerPath != "" {
		a.logMessage(fmt.Sprintf("Custom Docker path: %s", dockerPath))
	}
	if podmanPath := os.Getenv("INSTA_PODMAN_PATH"); podmanPath != "" {
		a.logMessage(fmt.Sprintf("Custom Podman path: %s", podmanPath))
	}

	// Initialize container runtime provider
	a.logMessage("Detecting container runtime...")
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		a.logMessage(fmt.Sprintf("Error detecting container runtime for UI: %v", err))
		a.runtimeInitError = fmt.Errorf("container runtime initialization failed: %w. Please ensure Docker or Podman is running and properly configured", err)
		return
	}

	a.containerRuntime = provider.SelectedRuntime()
	a.runtimeInitError = nil
	a.logMessage(fmt.Sprintf("UI using container runtime: %s", a.containerRuntime.Name()))

	// Initialize handlers
	a.logMessage("Initializing handlers...")
	a.initializeHandlers()
}

// initializeHandlers creates and configures all business logic handlers for direct runtime access
func (a *App) initializeHandlers() {
	if a.containerRuntime == nil || a.instaDir == "" {
		a.logMessage("Warning: Cannot initialize handlers - missing runtime or instaDir")
		return
	}

	a.serviceHandler = handlers.NewServiceHandler(a.containerRuntime, a.instaDir)
	a.connectionHandler = handlers.NewConnectionHandler(a.containerRuntime, a.instaDir)
	a.logsHandler = handlers.NewLogsHandler(a.containerRuntime, a.instaDir, a.ctx)
	a.imageHandler = handlers.NewImageHandler(a.containerRuntime, a.instaDir, a.ctx)
	a.dependencyHandler = handlers.NewDependencyHandler(a.containerRuntime, a.instaDir, a.serviceHandler)
	a.graphHandler = handlers.NewGraphHandler(a.containerRuntime, a.instaDir, a.dependencyHandler)

	a.logMessage("All handlers initialized successfully")
}

// initializeHandlersForCLI creates handlers that use the bundled CLI instead of direct runtime access
func (a *App) initializeHandlersForCLI() {
	if a.instaDir == "" {
		a.logMessage("Warning: Cannot initialize CLI handlers - missing instaDir")
		return
	}

	// Create CLI-based service handler
	cliServiceHandler := handlers.NewCLIServiceHandler(a.getBundledCLIPath(), a.instaDir)
	a.serviceHandler = cliServiceHandler

	// For other handlers, we'll still need the container runtime for some operations
	// Try to initialize container runtime for non-service operations
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err == nil {
		a.containerRuntime = provider.SelectedRuntime()
		a.connectionHandler = handlers.NewConnectionHandler(a.containerRuntime, a.instaDir)
		a.logsHandler = handlers.NewLogsHandler(a.containerRuntime, a.instaDir, a.ctx)
		a.imageHandler = handlers.NewImageHandler(a.containerRuntime, a.instaDir, a.ctx)
		a.dependencyHandler = handlers.NewDependencyHandler(a.containerRuntime, a.instaDir, a.serviceHandler)
		a.graphHandler = handlers.NewGraphHandler(a.containerRuntime, a.instaDir, a.dependencyHandler)
		a.logMessage("CLI-based service handler with container runtime support initialized successfully")
	} else {
		a.logMessage("CLI-based service handler initialized (container runtime not available for other operations)")
	}
}

// checkInitialization ensures the app is properly initialized before handler calls
func (a *App) checkInitialization() error {
	if a.runtimeInitError != nil {
		return a.runtimeInitError
	}

	// Check if we have either container runtime or bundled CLI
	if a.containerRuntime == nil && a.getBundledCLIPath() == "" {
		return fmt.Errorf("neither container runtime nor bundled CLI available")
	}

	if a.instaDir == "" {
		return fmt.Errorf("insta directory not determined")
	}
	return nil
}

// ==================== SERVICE MANAGEMENT ====================

func (a *App) ListServices() []models.ServiceInfo {
	if err := a.checkInitialization(); err != nil {
		fmt.Printf("Error in ListServices: %v\n", err)
		return []models.ServiceInfo{}
	}

	return a.serviceHandler.ListServices()
}

func (a *App) GetAllServiceDetails() ([]models.ServiceDetailInfo, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.GetAllServicesWithStatusAndDependencies()
}

func (a *App) GetServiceStatus(serviceName string) (string, error) {
	if err := a.checkInitialization(); err != nil {
		return "error", err
	}
	return a.serviceHandler.GetServiceStatus(serviceName)
}

func (a *App) GetServiceDependencies(serviceName string) ([]string, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.GetServiceDependencies(serviceName)
}

// GetMultipleServiceStatuses fetches statuses for multiple services efficiently
// This enables progressive loading where the UI shows services immediately and loads statuses separately
func (a *App) GetMultipleServiceStatuses(serviceNames []string) (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.GetMultipleServiceStatuses(serviceNames)
}

// GetAllRunningServices uses docker/podman ps to quickly get all running services
// This is much faster than the progressive loading approach - single fast command
func (a *App) GetAllRunningServices() (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.GetAllRunningServices()
}

// GetAllServiceDependencies efficiently gets dependencies for all services by reading compose file once
// This is much faster than calling GetAllDependenciesRecursive for each service individually
func (a *App) GetAllServiceDependencies() (map[string][]string, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.GetAllServiceDependencies()
}

// GetAllServicesWithStatusAndDependencies efficiently gets services, statuses, and dependencies in a single call
// This is the most efficient way to get complete service information
func (a *App) GetAllServicesWithStatusAndDependencies() ([]models.ServiceDetailInfo, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.GetAllServicesWithStatusAndDependencies()
}

func (a *App) StartService(serviceName string, persist bool) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.serviceHandler.StartService(serviceName, persist)
}

// StartServiceWithStatusUpdate starts a service and returns updated status information
// This provides immediate status updates without requiring additional Docker/Podman calls
func (a *App) StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.StartServiceWithStatusUpdate(serviceName, persist)
}

func (a *App) StopService(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.serviceHandler.StopService(serviceName)
}

func (a *App) StopAllServices() error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.serviceHandler.StopAllServices()
}

// StopServiceWithStatusUpdate stops a service and returns updated status information
// This provides immediate status updates without requiring additional Docker/Podman calls
func (a *App) StopServiceWithStatusUpdate(serviceName string) (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.StopServiceWithStatusUpdate(serviceName)
}

// StopAllServicesWithStatusUpdate stops all services and returns updated status information
// This provides immediate status updates without requiring additional Docker/Podman calls
func (a *App) StopAllServicesWithStatusUpdate() (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.StopAllServicesWithStatusUpdate()
}

// RefreshStatusFromContainers clears any cached status and gets fresh status from containers
// This should be called when the UI wants to refresh from actual container state
func (a *App) RefreshStatusFromContainers() (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.RefreshStatusFromContainers()
}

// CheckStartingServicesProgress checks if any services marked as starting have completed their transition
// Returns updated statuses only for services that have changed state
func (a *App) CheckStartingServicesProgress() (map[string]models.ServiceStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.serviceHandler.CheckStartingServicesProgress()
}

// ==================== CONNECTION MANAGEMENT ====================

func (a *App) GetServiceConnectionInfo(serviceName string) (*models.ServiceConnectionInfo, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.connectionHandler.GetServiceConnectionInfo(serviceName)
}

func (a *App) OpenServiceInBrowser(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.connectionHandler.OpenServiceInBrowser(serviceName)
}

// ==================== LOGGING ====================

func (a *App) GetServiceLogs(serviceName string, tailLines int) ([]string, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.logsHandler.GetServiceLogs(serviceName, tailLines)
}

func (a *App) StartLogStream(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.logsHandler.StartLogStream(serviceName)
}

func (a *App) StopLogStream(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.logsHandler.StopLogStream(serviceName)
}

// ==================== IMAGE MANAGEMENT ====================

func (a *App) CheckImageExists(serviceName string) (bool, error) {
	if err := a.checkInitialization(); err != nil {
		return false, err
	}
	return a.imageHandler.CheckImageExists(serviceName)
}

func (a *App) GetImagePullProgress(serviceName string) (*models.ImagePullProgress, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.imageHandler.GetImagePullProgress(serviceName)
}

func (a *App) StartImagePull(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.imageHandler.StartImagePull(serviceName)
}

func (a *App) StopImagePull(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.imageHandler.StopImagePull(serviceName)
}

func (a *App) GetImageInfo(serviceName string) (string, error) {
	if err := a.checkInitialization(); err != nil {
		return "", err
	}
	return a.imageHandler.GetImageInfo(serviceName)
}

// ==================== DEPENDENCY MANAGEMENT ====================

func (a *App) GetDependencyStatus(serviceName string) (*models.DependencyStatus, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.dependencyHandler.GetDependencyStatus(serviceName)
}

func (a *App) StartAllDependencies(serviceName string, persist bool) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.dependencyHandler.StartAllDependencies(serviceName, persist)
}

func (a *App) StopDependencyChain(serviceName string) error {
	if err := a.checkInitialization(); err != nil {
		return err
	}
	return a.dependencyHandler.StopDependencyChain(serviceName)
}

// ==================== GRAPH VISUALIZATION ====================

func (a *App) GetDependencyGraph() (*models.DependencyGraph, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.graphHandler.GetDependencyGraph()
}

func (a *App) GetServiceDependencyGraph(serviceName string) (*models.DependencyGraph, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.graphHandler.GetServiceDependencyGraph(serviceName)
}

func (a *App) GetServiceContainerDetails(serviceName string) (*models.ServiceContainerDetails, error) {
	if err := a.checkInitialization(); err != nil {
		return nil, err
	}
	return a.graphHandler.GetServiceContainerDetails(serviceName)
}

// ==================== RUNTIME MANAGEMENT ====================

// GetCurrentRuntime returns the name of the currently active container runtime
func (a *App) GetCurrentRuntime() string {
	if a.containerRuntime == nil {
		return "none"
	}
	return a.containerRuntime.Name()
}

// GetRuntimeStatus returns detailed information about container runtime availability
func (a *App) GetRuntimeStatus() *container.SystemRuntimeStatus {
	return container.GetDetailedRuntimeStatus()
}

// SetCustomDockerPath sets a custom Docker binary path
func (a *App) SetCustomDockerPath(path string) error {
	if path == "" {
		a.logMessage("Clearing custom Docker path")
		os.Unsetenv("INSTA_DOCKER_PATH")
	} else {
		// Validate the path exists
		if _, err := os.Stat(path); err != nil {
			a.logMessage(fmt.Sprintf("Failed to set Docker path - binary not found: %s", path))
			return fmt.Errorf("docker binary not found at path: %s", path)
		}
		a.logMessage(fmt.Sprintf("Setting custom Docker path: %s", path))
		os.Setenv("INSTA_DOCKER_PATH", path)
	}
	return nil
}

// SetCustomPodmanPath sets a custom Podman binary path
func (a *App) SetCustomPodmanPath(path string) error {
	if path == "" {
		a.logMessage("Clearing custom Podman path")
		os.Unsetenv("INSTA_PODMAN_PATH")
	} else {
		// Validate the path exists
		if _, err := os.Stat(path); err != nil {
			a.logMessage(fmt.Sprintf("Failed to set Podman path - binary not found: %s", path))
			return fmt.Errorf("podman binary not found at path: %s", path)
		}
		a.logMessage(fmt.Sprintf("Setting custom Podman path: %s", path))
		os.Setenv("INSTA_PODMAN_PATH", path)
	}
	return nil
}

// GetCustomDockerPath returns the current custom Docker path
func (a *App) GetCustomDockerPath() string {
	return os.Getenv("INSTA_DOCKER_PATH")
}

// GetCustomPodmanPath returns the current custom Podman path
func (a *App) GetCustomPodmanPath() string {
	return os.Getenv("INSTA_PODMAN_PATH")
}

// AttemptStartRuntime tries to start the specified container runtime service
func (a *App) AttemptStartRuntime(runtimeName string) *container.StartupResult {
	manager := container.NewRuntimeManager()
	return manager.AttemptStartRuntime(runtimeName)
}

// WaitForRuntimeReady waits for a runtime to become available and reinitializes the app
func (a *App) WaitForRuntimeReady(runtimeName string, maxWaitSeconds int) *container.StartupResult {
	manager := container.NewRuntimeManager()
	result := manager.WaitForRuntimeReady(runtimeName, maxWaitSeconds)

	// If the runtime is now ready, try to reinitialize our app
	if result.Success {
		provider := container.NewProvider()
		if err := provider.DetectRuntime(); err == nil {
			a.containerRuntime = provider.SelectedRuntime()
			a.runtimeInitError = nil
			a.initializeHandlers()
			fmt.Printf("Successfully reinitialized with %s runtime\n", a.containerRuntime.Name())
		}
	}

	return result
}

// ReinitializeRuntime attempts to reinitialize the container runtime after it becomes available
func (a *App) ReinitializeRuntime() error {
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		a.runtimeInitError = fmt.Errorf("container runtime still not available: %w", err)
		return a.runtimeInitError
	}

	a.containerRuntime = provider.SelectedRuntime()
	a.runtimeInitError = nil
	a.initializeHandlers()
	fmt.Printf("Successfully reinitialized with %s runtime\n", a.containerRuntime.Name())

	return nil
}

// ensureComposeFiles extracts Docker Compose files from embedded resources if needed
func (a *App) ensureComposeFiles() error {
	// Create insta directory if it doesn't exist
	if err := os.MkdirAll(a.instaDir, 0755); err != nil {
		return fmt.Errorf("failed to create insta directory: %w", err)
	}

	versionFilePath := filepath.Join(a.instaDir, ".version_synced")
	var syncedVersionBytes []byte
	syncedVersionBytes, readErr := os.ReadFile(versionFilePath)

	needsSync := false
	if readErr != nil { // Error reading (e.g., doesn't exist, permission issue)
		needsSync = true
		if !os.IsNotExist(readErr) {
			// Log if it's an unexpected error other than file not existing
			a.logMessage(fmt.Sprintf("Warning: failed to read version sync marker '%s': %v", versionFilePath, readErr))
		}
	} else {
		if strings.TrimSpace(string(syncedVersionBytes)) != version {
			needsSync = true
		}
	}

	if needsSync {
		a.logMessage(fmt.Sprintf("Performing one-time file synchronization for version %s...", version))

		// Extract docker-compose.yaml
		composePath := filepath.Join(a.instaDir, "docker-compose.yaml")
		composeContent, err := embedFS.ReadFile("resources/docker-compose.yaml")
		if err != nil {
			return fmt.Errorf("failed to read embedded docker-compose.yaml: %w", err)
		}
		if err := os.WriteFile(composePath, composeContent, 0644); err != nil {
			return fmt.Errorf("failed to write docker-compose.yaml: %w", err)
		}

		// Extract docker-compose-persist.yaml
		persistPath := filepath.Join(a.instaDir, "docker-compose-persist.yaml")
		persistContent, err := embedFS.ReadFile("resources/docker-compose-persist.yaml")
		if err != nil {
			return fmt.Errorf("failed to read embedded docker-compose-persist.yaml: %w", err)
		}
		if err := os.WriteFile(persistPath, persistContent, 0644); err != nil {
			return fmt.Errorf("failed to write docker-compose-persist.yaml: %w", err)
		}

		// Extract data files (this will also create the dataDir if needed)
		if err := a.extractDataFiles(); err != nil {
			return fmt.Errorf("failed to extract data files during sync: %w", err)
		}

		// Update .version_synced file
		if err := os.WriteFile(versionFilePath, []byte(version), 0644); err != nil {
			a.logMessage(fmt.Sprintf("Warning: failed to write synced version marker '%s': %v", versionFilePath, err))
			// Continue without returning error, as files are synced for this session. Next run will attempt to resync.
		}
		a.logMessage("File synchronization complete.")
	}

	return nil
}

// extractDataFiles extracts data files from embedded resources
func (a *App) extractDataFiles() error {
	// Create data directory in insta dir
	dataDir := filepath.Join(a.instaDir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	return fs.WalkDir(embedFS, "resources/data", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip all persist directories and files within them
		if strings.Contains(path, "persist") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Get relative path from resources/data to use for target
		relPath, err := filepath.Rel("resources/data", path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		if d.IsDir() {
			targetDir := filepath.Join(dataDir, relPath)
			if err := os.MkdirAll(targetDir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
			}
			return nil
		}

		content, err := embedFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		targetFile := filepath.Join(dataDir, relPath)
		targetDir := filepath.Dir(targetFile)

		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", targetDir, err)
		}

		if err := os.WriteFile(targetFile, content, 0755); err != nil {
			return fmt.Errorf("failed to write file %s: %w", targetFile, err)
		}

		return nil
	})
}
