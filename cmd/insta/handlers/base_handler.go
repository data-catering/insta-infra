package handlers

import (
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// BaseHandler contains common functionality shared across all handlers
type BaseHandler struct {
	containerRuntime container.Runtime
	instaDir         string
	logger           Logger
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(runtime container.Runtime, instaDir string, logger Logger) *BaseHandler {
	return &BaseHandler{
		containerRuntime:  runtime,
		instaDir:          instaDir,
		logger:            logger,
	}
}

// Runtime returns the container runtime
func (h *BaseHandler) Runtime() container.Runtime {
	return h.containerRuntime
}

// InstaDir returns the insta directory
func (h *BaseHandler) InstaDir() string {
	return h.instaDir
}

// Log logs a message using the handler's logger
func (h *BaseHandler) Log(message string) {
	if h.logger != nil {
		h.logger.Log(message)
	}
}
