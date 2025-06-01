package main

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/internal"
	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// Version information - set during build via ldflags
var (
	version   = "dev"
	buildTime = "unknown"
)

//go:embed resources/docker-compose.yaml resources/docker-compose-persist.yaml
//go:embed all:resources/data
var embedFS embed.FS

// App struct contains the application context and dependencies
type App struct {
	ctx            context.Context
	config         *internal.AppConfig
	handlerManager *internal.HandlerManager
	logger         *internal.AppLogger
	initError      error
}

// NewApp creates a new App application struct
func NewApp() *App {
	logger := internal.NewAppLogger()
	return &App{
		logger:         logger,
		config:         internal.NewAppConfig(embedFS, version, logger),
		handlerManager: internal.NewHandlerManager(logger),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.logger.Log("Starting insta-infra UI application...")

	// Initialize configuration and files
	if err := a.config.Initialize(); err != nil {
		a.logger.Log(fmt.Sprintf("ERROR: Failed to initialize config: %v", err))
		a.initError = err
		return
	}

	a.logger.Log("Starting application initialization...")

	// Initialize handlers using container runtime
	a.logger.Log("Initializing handlers with container runtime...")
	if err := a.handlerManager.Initialize(a.config.InstaDir, ctx); err != nil {
		a.logger.Log(fmt.Sprintf("ERROR: Handler initialization failed: %v", err))
		a.initError = err
		return
	}

	a.logger.Log("Application startup completed successfully")
}

// checkReady ensures the app is properly initialized
func (a *App) checkReady() error {
	if a.initError != nil {
		return a.initError
	}
	if a.handlerManager.GetServiceHandler() == nil {
		return fmt.Errorf("service handler not available")
	}
	return nil
}

// ==================== SERVICE MANAGEMENT ====================

func (a *App) ListServices() []models.ServiceInfo {
	if err := a.checkReady(); err != nil {
		a.logger.Log(fmt.Sprintf("Error in ListServices: %v", err))
		return []models.ServiceInfo{}
	}
	return a.handlerManager.GetServiceHandler().ListServices()
}

func (a *App) GetAllServiceDetails() ([]models.ServiceDetailInfo, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().GetAllServicesWithStatusAndDependencies()
}

func (a *App) GetServiceStatus(serviceName string) (string, error) {
	if err := a.checkReady(); err != nil {
		return "error", err
	}
	return a.handlerManager.GetServiceHandler().GetServiceStatus(serviceName)
}

func (a *App) GetMultipleServiceStatuses(serviceNames []string) (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().GetMultipleServiceStatuses(serviceNames)
}

func (a *App) GetAllRunningServices() (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().GetAllRunningServices()
}

func (a *App) GetAllServicesWithStatusAndDependencies() ([]models.ServiceDetailInfo, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().GetAllServicesWithStatusAndDependencies()
}

func (a *App) StartService(serviceName string, persist bool) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	return a.handlerManager.GetServiceHandler().StartService(serviceName, persist)
}

func (a *App) StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().StartServiceWithStatusUpdate(serviceName, persist)
}

func (a *App) StopService(serviceName string) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	return a.handlerManager.GetServiceHandler().StopService(serviceName)
}

func (a *App) StopAllServices() error {
	if err := a.checkReady(); err != nil {
		return err
	}
	return a.handlerManager.GetServiceHandler().StopAllServices()
}

func (a *App) StopServiceWithStatusUpdate(serviceName string) (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().StopServiceWithStatusUpdate(serviceName)
}

func (a *App) StopAllServicesWithStatusUpdate() (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().StopAllServicesWithStatusUpdate()
}

func (a *App) RefreshStatusFromContainers() (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().RefreshStatusFromContainers()
}

func (a *App) CheckStartingServicesProgress() (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().CheckStartingServicesProgress()
}

func (a *App) GetAllDependencyStatuses() (map[string]models.ServiceStatus, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	return a.handlerManager.GetServiceHandler().GetAllDependencyStatuses()
}

// ==================== CONNECTION MANAGEMENT ====================

func (a *App) GetServiceConnectionInfo(serviceName string) (*models.ServiceConnectionInfo, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	handler := a.handlerManager.GetConnectionHandler()
	if handler == nil {
		return nil, fmt.Errorf("connection handler not available")
	}
	return handler.GetServiceConnectionInfo(serviceName)
}

func (a *App) OpenServiceInBrowser(serviceName string) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	handler := a.handlerManager.GetConnectionHandler()
	if handler == nil {
		return fmt.Errorf("connection handler not available")
	}
	return handler.OpenServiceInBrowser(serviceName)
}

// ==================== LOGGING ====================

func (a *App) GetServiceLogs(serviceName string, tailLines int) ([]string, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	handler := a.handlerManager.GetLogsHandler()
	if handler == nil {
		return nil, fmt.Errorf("logs handler not available")
	}
	return handler.GetServiceLogs(serviceName, tailLines)
}

func (a *App) StartLogStream(serviceName string) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	handler := a.handlerManager.GetLogsHandler()
	if handler == nil {
		return fmt.Errorf("logs handler not available")
	}
	return handler.StartLogStream(serviceName)
}

func (a *App) StopLogStream(serviceName string) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	handler := a.handlerManager.GetLogsHandler()
	if handler == nil {
		return fmt.Errorf("logs handler not available")
	}
	return handler.StopLogStream(serviceName)
}

func (a *App) GetAppLogs() []string {
	return a.logger.GetLogs()
}

// ==================== IMAGE MANAGEMENT ====================

func (a *App) CheckImageExists(serviceName string) (bool, error) {
	if err := a.checkReady(); err != nil {
		return false, err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return false, fmt.Errorf("image handler not available")
	}
	return handler.CheckImageExists(serviceName)
}

func (a *App) GetImagePullProgress(serviceName string) (*models.ImagePullProgress, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return nil, fmt.Errorf("image handler not available")
	}
	return handler.GetImagePullProgress(serviceName)
}

func (a *App) StartImagePull(serviceName string) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return fmt.Errorf("image handler not available")
	}
	return handler.StartImagePull(serviceName)
}

func (a *App) StopImagePull(serviceName string) error {
	if err := a.checkReady(); err != nil {
		return err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return fmt.Errorf("image handler not available")
	}
	return handler.StopImagePull(serviceName)
}

func (a *App) GetImageInfo(serviceName string) (string, error) {
	if err := a.checkReady(); err != nil {
		return "", err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return "", fmt.Errorf("image handler not available")
	}
	return handler.GetImageInfo(serviceName)
}

func (a *App) CheckMultipleImagesExist(serviceNames []string) (map[string]bool, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return nil, fmt.Errorf("image handler not available")
	}
	return handler.CheckMultipleImagesExist(serviceNames)
}

func (a *App) GetMultipleImageInfo(serviceNames []string) (map[string]string, error) {
	if err := a.checkReady(); err != nil {
		return nil, err
	}
	handler := a.handlerManager.GetImageHandler()
	if handler == nil {
		return nil, fmt.Errorf("image handler not available")
	}
	return handler.GetMultipleImageInfo(serviceNames)
}

// ==================== RUNTIME MANAGEMENT ====================

func (a *App) GetCurrentRuntime() string {
	runtime := a.handlerManager.GetContainerRuntime()
	if runtime == nil {
		return "none"
	}
	return runtime.Name()
}

func (a *App) GetRuntimeStatus() *container.SystemRuntimeStatus {
	return container.GetDetailedRuntimeStatus()
}

func (a *App) SetCustomDockerPath(path string) error {
	if path == "" {
		a.logger.Log("Clearing custom Docker path")
		os.Unsetenv("INSTA_DOCKER_PATH")
	} else {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("docker binary not found at path: %s", path)
		}
		a.logger.Log(fmt.Sprintf("Setting custom Docker path: %s", path))
		os.Setenv("INSTA_DOCKER_PATH", path)
	}
	return nil
}

func (a *App) SetCustomPodmanPath(path string) error {
	if path == "" {
		a.logger.Log("Clearing custom Podman path")
		os.Unsetenv("INSTA_PODMAN_PATH")
	} else {
		if _, err := os.Stat(path); err != nil {
			return fmt.Errorf("podman binary not found at path: %s", path)
		}
		a.logger.Log(fmt.Sprintf("Setting custom Podman path: %s", path))
		os.Setenv("INSTA_PODMAN_PATH", path)
	}
	return nil
}

func (a *App) GetCustomDockerPath() string {
	return os.Getenv("INSTA_DOCKER_PATH")
}

func (a *App) GetCustomPodmanPath() string {
	return os.Getenv("INSTA_PODMAN_PATH")
}

func (a *App) AttemptStartRuntime(runtimeName string) *container.StartupResult {
	manager := container.NewRuntimeManager()
	return manager.AttemptStartRuntime(runtimeName)
}

func (a *App) WaitForRuntimeReady(runtimeName string, maxWaitSeconds int) *container.StartupResult {
	manager := container.NewRuntimeManager()
	result := manager.WaitForRuntimeReady(runtimeName, maxWaitSeconds)

	if result.Success {
		if err := a.handlerManager.ReinitializeRuntime(a.config.InstaDir, a.ctx); err == nil {
			a.logger.Log(fmt.Sprintf("Successfully reinitialized with %s runtime", a.GetCurrentRuntime()))
		}
	}

	return result
}

func (a *App) ReinitializeRuntime() error {
	if err := a.handlerManager.ReinitializeRuntime(a.config.InstaDir, a.ctx); err != nil {
		a.initError = err
		return err
	}

	a.initError = nil
	a.logger.Log(fmt.Sprintf("Successfully reinitialized with %s runtime", a.GetCurrentRuntime()))
	return nil
}
