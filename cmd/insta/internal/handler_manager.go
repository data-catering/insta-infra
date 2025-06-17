package internal

import (
	"context"
	"fmt"

	"github.com/data-catering/insta-infra/v2/cmd/insta/handlers"
	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// HandlerManager manages all application handlers
type HandlerManager struct {
	serviceHandler    handlers.ServiceHandlerInterface
	connectionHandler *handlers.ConnectionHandler
	logsHandler       *handlers.LogsHandler
	imageHandler      *handlers.ImageHandler

	containerRuntime container.Runtime
	logger           *AppLogger
}

// NewHandlerManager creates a new handler manager
func NewHandlerManager(logger *AppLogger) *HandlerManager {
	return &HandlerManager{
		logger: logger,
	}
}

// Initialize initializes handlers using ServiceManager directly
func (h *HandlerManager) Initialize(instaDir string, ctx context.Context) error {
	return h.InitializeWithCallback(instaDir, ctx, nil)
}

// InitializeWithCallback initializes handlers with a progress callback
func (h *HandlerManager) InitializeWithCallback(instaDir string, ctx context.Context, progressCallback handlers.ProgressCallback) error {
	if instaDir == "" {
		return fmt.Errorf("missing instaDir")
	}

	// Try to detect and initialize container runtime, but don't fail if none is found
	provider := container.NewProvider()
	var runtime container.Runtime

	if err := provider.DetectRuntime(); err != nil {
		h.logger.Log(fmt.Sprintf("Warning: No container runtime detected: %v", err))
		h.logger.Log("Handlers initialized in runtime setup mode - some features will be limited until a container runtime is configured")

		// Set runtime to nil - handlers will need to handle this gracefully
		h.containerRuntime = nil
		h.serviceHandler = nil
		h.connectionHandler = nil
		h.logsHandler = nil
		h.imageHandler = nil

		return nil // Don't fail - allow UI to start for runtime setup
	}

	runtime = provider.SelectedRuntime()
	h.containerRuntime = runtime

	// Create runtime info adapter with instaDir
	runtimeInfo := handlers.NewRuntimeInfoAdapterWithDir(runtime, instaDir)

	// Create service manager directly (it implements ServiceHandlerInterface)
	serviceManager := models.NewServiceManager(instaDir, runtimeInfo, h.logger)

	// Load services from compose files
	if err := serviceManager.LoadServices(); err != nil {
		h.logger.Log(fmt.Sprintf("Warning: Failed to load services: %v", err))
	}

	h.serviceHandler = serviceManager
	h.connectionHandler = handlers.NewConnectionHandler(runtime, instaDir, h.logger)
	h.logsHandler = handlers.NewLogsHandler(runtime, instaDir, ctx, h.logger)
	h.imageHandler = handlers.NewImageHandlerWithCallback(runtime, instaDir, ctx, h.logger, progressCallback)

	h.logger.Log(fmt.Sprintf("Handlers initialized with %s runtime", runtime.Name()))
	return nil
}

// GetServiceHandler returns the current service handler
func (h *HandlerManager) GetServiceHandler() handlers.ServiceHandlerInterface {
	return h.serviceHandler
}

// GetConnectionHandler returns the connection handler
func (h *HandlerManager) GetConnectionHandler() *handlers.ConnectionHandler {
	return h.connectionHandler
}

// GetLogsHandler returns the logs handler
func (h *HandlerManager) GetLogsHandler() *handlers.LogsHandler {
	return h.logsHandler
}

// GetImageHandler returns the image handler
func (h *HandlerManager) GetImageHandler() *handlers.ImageHandler {
	return h.imageHandler
}

// GetContainerRuntime returns the current container runtime
func (h *HandlerManager) GetContainerRuntime() container.Runtime {
	return h.containerRuntime
}

// ReinitializeRuntime attempts to reinitialize the container runtime
func (h *HandlerManager) ReinitializeRuntime(instaDir string, ctx context.Context) error {
	return h.Initialize(instaDir, ctx)
}
