package handlers

import (
	"context"
	"fmt"
	"sync"

	"github.com/data-catering/insta-infra/v2/cmd/instaui/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ImageHandler handles Docker/Podman image management operations
type ImageHandler struct {
	*BaseHandler
	ctx             context.Context
	logStreams      map[string]chan struct{}
	logStreamsMutex sync.RWMutex
}

// NewImageHandler creates a new image handler
func NewImageHandler(runtime container.Runtime, instaDir string, ctx context.Context) *ImageHandler {
	return &ImageHandler{
		BaseHandler: NewBaseHandler(runtime, instaDir),
		ctx:         ctx,
		logStreams:  make(map[string]chan struct{}),
	}
}

// CheckImageExists checks if a Docker/Podman image exists locally for a service
func (h *ImageHandler) CheckImageExists(serviceName string) (bool, error) {
	composeFiles := h.getComposeFiles()

	// Get image name for the service
	imageName, err := h.Runtime().GetImageInfo(serviceName, composeFiles)
	if err != nil {
		return false, fmt.Errorf("failed to get image info for service %s: %w", serviceName, err)
	}

	// Check if image exists locally
	exists, err := h.Runtime().CheckImageExists(imageName)
	if err != nil {
		return false, fmt.Errorf("failed to check if image exists for service %s: %w", serviceName, err)
	}

	return exists, nil
}

// GetImagePullProgress checks the current image pull progress for a service
func (h *ImageHandler) GetImagePullProgress(serviceName string) (*models.ImagePullProgress, error) {
	h.logStreamsMutex.RLock()
	defer h.logStreamsMutex.RUnlock()

	// Check if there's an active image pull
	streamKey := fmt.Sprintf("image-pull-%s", serviceName)
	if _, exists := h.logStreams[streamKey]; exists {
		return &models.ImagePullProgress{
			Status:      "downloading",
			ServiceName: serviceName,
		}, nil
	}

	return &models.ImagePullProgress{
		Status:      "idle",
		ServiceName: serviceName,
	}, nil
}

// StartImagePull starts pulling the Docker/Podman image for a service with progress tracking
func (h *ImageHandler) StartImagePull(serviceName string) error {
	composeFiles := h.getComposeFiles()

	// Get image name for the service
	imageName, err := h.Runtime().GetImageInfo(serviceName, composeFiles)
	if err != nil {
		return fmt.Errorf("failed to get image info for service %s: %w", serviceName, err)
	}

	// Check if image pull is already in progress
	streamKey := fmt.Sprintf("image-pull-%s", serviceName)
	h.logStreamsMutex.Lock()
	if _, exists := h.logStreams[streamKey]; exists {
		h.logStreamsMutex.Unlock()
		return fmt.Errorf("image pull already in progress for service %s", serviceName)
	}

	// Create stop channel for this image pull
	stopChan := make(chan struct{})
	h.logStreams[streamKey] = stopChan
	h.logStreamsMutex.Unlock()

	// Start image pull in a goroutine
	go func() {
		defer func() {
			// Clean up the stop channel when pull ends
			h.logStreamsMutex.Lock()
			delete(h.logStreams, streamKey)
			h.logStreamsMutex.Unlock()

		}()

		// Create progress channel
		progressChan := make(chan container.ImagePullProgress, 10)

		// Start the pull process
		go func() {
			defer close(progressChan)
			if err := h.Runtime().PullImageWithProgress(imageName, progressChan, stopChan); err != nil {
				// Send final error progress
				progressChan <- container.ImagePullProgress{
					Status: "error",
					Error:  err.Error(),
				}
			}
		}()

		// Forward progress events to the frontend
		for progress := range progressChan {
			// Convert container.ImagePullProgress to our models.ImagePullProgress
			appProgress := models.ImagePullProgress{
				Status:       progress.Status,
				Progress:     progress.Progress,
				CurrentLayer: progress.CurrentLayer,
				TotalLayers:  progress.TotalLayers,
				Downloaded:   progress.Downloaded,
				Total:        progress.Total,
				Speed:        progress.Speed,
				ETA:          progress.ETA,
				Error:        progress.Error,
				ServiceName:  serviceName,
			}

			// Emit progress event to frontend
			if h.ctx != nil {
				runtime.EventsEmit(h.ctx, "image-pull-progress", appProgress)
			}
		}
	}()

	return nil
}

// StopImagePull stops the image pull process for a service
func (h *ImageHandler) StopImagePull(serviceName string) error {
	streamKey := fmt.Sprintf("image-pull-%s", serviceName)

	h.logStreamsMutex.Lock()
	defer h.logStreamsMutex.Unlock()

	stopChan, exists := h.logStreams[streamKey]
	if !exists {
		return fmt.Errorf("no active image pull found for service %s", serviceName)
	}

	// Signal the pull goroutine to stop
	close(stopChan)
	// Remove from active streams
	delete(h.logStreams, streamKey)

	return nil
}

// GetImageInfo returns the Docker/Podman image name for a service
func (h *ImageHandler) GetImageInfo(serviceName string) (string, error) {
	composeFiles := h.getComposeFiles()

	// Get image name for the service
	imageName, err := h.Runtime().GetImageInfo(serviceName, composeFiles)
	if err != nil {
		return "", fmt.Errorf("failed to get image info for service %s: %w", serviceName, err)
	}

	return imageName, nil
}
