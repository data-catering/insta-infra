package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/handlers"
	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// App struct - Maintains Wails compatibility
type App struct {
	ctx              context.Context
	containerRuntime container.Runtime
	instaDir         string
	runtimeInitError error

	// Handler instances for delegating business logic
	serviceHandler    *handlers.ServiceHandler
	connectionHandler *handlers.ConnectionHandler
	logsHandler       *handlers.LogsHandler
	imageHandler      *handlers.ImageHandler
	dependencyHandler *handlers.DependencyHandler
	graphHandler      *handlers.GraphHandler
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Determine instaDir
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory for UI: %v\n", err)
		a.runtimeInitError = fmt.Errorf("failed to get home directory: %w", err)
		return
	}

	instaDir := os.Getenv("INSTA_HOME")
	if instaDir == "" {
		a.instaDir = filepath.Join(homeDir, ".insta")
	} else {
		a.instaDir = instaDir
	}
	fmt.Printf("UI insta directory: %s\n", a.instaDir)

	// Initialize container runtime provider
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		fmt.Printf("Error detecting container runtime for UI: %v\n", err)
		a.runtimeInitError = fmt.Errorf("container runtime initialization failed: %w. Please ensure Docker or Podman is running and properly configured", err)
		return
	}

	a.containerRuntime = provider.SelectedRuntime()
	a.runtimeInitError = nil
	fmt.Printf("UI using container runtime: %s\n", a.containerRuntime.Name())

	// Initialize handlers
	a.initializeHandlers()
}

// initializeHandlers creates and configures all business logic handlers
func (a *App) initializeHandlers() {
	if a.containerRuntime == nil || a.instaDir == "" {
		fmt.Printf("Warning: Cannot initialize handlers - missing runtime or instaDir\n")
		return
	}

	a.serviceHandler = handlers.NewServiceHandler(a.containerRuntime, a.instaDir)
	a.connectionHandler = handlers.NewConnectionHandler(a.containerRuntime, a.instaDir)
	a.logsHandler = handlers.NewLogsHandler(a.containerRuntime, a.instaDir, a.ctx)
	a.imageHandler = handlers.NewImageHandler(a.containerRuntime, a.instaDir, a.ctx)
	a.dependencyHandler = handlers.NewDependencyHandler(a.containerRuntime, a.instaDir, a.serviceHandler)
	a.graphHandler = handlers.NewGraphHandler(a.containerRuntime, a.instaDir, a.dependencyHandler)

	fmt.Printf("All handlers initialized successfully\n")
}

// checkInitialization ensures the app is properly initialized before handler calls
func (a *App) checkInitialization() error {
	if a.runtimeInitError != nil {
		return a.runtimeInitError
	}
	if a.containerRuntime == nil {
		return fmt.Errorf("container runtime not available")
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
		return "unknown"
	}
	return a.containerRuntime.Name()
}

// GetRuntimeStatus returns detailed information about container runtime availability
func (a *App) GetRuntimeStatus() *container.SystemRuntimeStatus {
	return container.GetDetailedRuntimeStatus()
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
