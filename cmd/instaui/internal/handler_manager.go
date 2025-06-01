package internal

import (
	"context"
	"fmt"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/handlers"
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

// Initialize initializes handlers using direct container runtime
func (h *HandlerManager) Initialize(instaDir string, ctx context.Context) error {
	if instaDir == "" {
		return fmt.Errorf("missing instaDir")
	}

	// Detect and initialize container runtime
	provider := container.NewProvider()
	if err := provider.DetectRuntime(); err != nil {
		return fmt.Errorf("container runtime detection failed: %w", err)
	}

	runtime := provider.SelectedRuntime()
	h.containerRuntime = runtime
	h.serviceHandler = handlers.NewServiceHandler(runtime, instaDir)
	h.connectionHandler = handlers.NewConnectionHandler(runtime, instaDir)
	h.logsHandler = handlers.NewLogsHandler(runtime, instaDir, ctx)
	h.imageHandler = handlers.NewImageHandler(runtime, instaDir, ctx)

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
