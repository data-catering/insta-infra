package handlers

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

// ProgressCallback is a function type for broadcasting progress updates
type ProgressCallback func(serviceName, imageName string, progress float64, status string)

// ImageHandler handles image-related HTTP requests with simplified logic
type ImageHandler struct {
	runtime          container.Runtime
	serviceManager   *models.ServiceManager
	logger           Logger
	ctx              context.Context
	progressCallback ProgressCallback
}

// NewImageHandler creates a new simplified image handler
func NewImageHandler(runtime container.Runtime, instaDir string, ctx context.Context, logger Logger) *ImageHandler {
	return NewImageHandlerWithCallback(runtime, instaDir, ctx, logger, nil)
}

// NewImageHandlerWithCallback creates a new image handler with progress callback
func NewImageHandlerWithCallback(runtime container.Runtime, instaDir string, ctx context.Context, logger Logger, progressCallback ProgressCallback) *ImageHandler {
	// Create runtime info adapter for service manager
	runtimeInfo := NewRuntimeInfoAdapter(runtime)

	// Create service manager to get image information from enhanced services
	serviceManager := models.NewServiceManager(instaDir, runtimeInfo, logger)

	// Load services to get image information
	if err := serviceManager.LoadServices(); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to load services for image handler: %v", err))
	}

	return &ImageHandler{
		runtime:          runtime,
		serviceManager:   serviceManager,
		logger:           logger,
		ctx:              ctx,
		progressCallback: progressCallback,
	}
}

// Note: RuntimeInfoAdapter is defined in runtime_adapter.go

// resolveImageName resolves environment variables in image names
// Supports ${VAR:-default} syntax used in Docker Compose
func (h *ImageHandler) resolveImageName(imageName string) string {
	if imageName == "" {
		return imageName
	}

	// Pattern to match ${VAR:-default} syntax
	re := regexp.MustCompile(`\$\{([^}]+)\}`)

	return re.ReplaceAllStringFunc(imageName, func(match string) string {
		// Remove ${ and }
		content := match[2 : len(match)-1]

		// Split on :- to get variable name and default value
		parts := strings.SplitN(content, ":-", 2)
		varName := parts[0]

		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// Return default value if provided
		if len(parts) > 1 {
			return parts[1]
		}

		// Return empty string if no default
		return ""
	})
}

// GetAllImageStatuses returns the status of all images for all services
func (h *ImageHandler) GetAllImageStatuses() (map[string]ImageStatus, error) {
	h.logger.Log("Getting all image statuses")

	imageStatuses := make(map[string]ImageStatus)

	// Get all enhanced services
	for _, service := range h.serviceManager.ListServices() {
		if service.ImageName == "" {
			continue // Skip services without images
		}

		// Resolve environment variables in image name
		resolvedImageName := h.resolveImageName(service.ImageName)

		// Check if image exists locally
		exists, err := h.runtime.CheckImageExists(resolvedImageName)
		if err != nil {
			h.logger.Log(fmt.Sprintf("Failed to check image existence for %s (resolved: %s): %v", service.Name, resolvedImageName, err))
			imageStatuses[service.Name] = ImageStatus{
				ServiceName: service.Name,
				ImageName:   service.ImageName,
				Status:      "error",
				Error:       err.Error(),
			}
			continue
		}

		status := "missing"
		if exists {
			status = "available"
		}

		imageStatuses[service.Name] = ImageStatus{
			ServiceName: service.Name,
			ImageName:   service.ImageName,
			Status:      status,
		}
	}

	h.logger.Log(fmt.Sprintf("Retrieved status for %d images", len(imageStatuses)))
	return imageStatuses, nil
}

// GetImageStatus returns the status of a specific service's image
func (h *ImageHandler) GetImageStatus(serviceName string) (ImageStatus, error) {
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return ImageStatus{}, fmt.Errorf("service %s not found", serviceName)
	}

	if service.ImageName == "" {
		return ImageStatus{
			ServiceName: serviceName,
			Status:      "no_image",
		}, nil
	}

	// Resolve environment variables in image name
	resolvedImageName := h.resolveImageName(service.ImageName)

	// Check if image exists locally
	exists, err := h.runtime.CheckImageExists(resolvedImageName)
	if err != nil {
		return ImageStatus{
			ServiceName: serviceName,
			ImageName:   service.ImageName,
			Status:      "error",
			Error:       err.Error(),
		}, err
	}

	status := "missing"
	if exists {
		status = "available"
	}

	return ImageStatus{
		ServiceName: serviceName,
		ImageName:   service.ImageName,
		Status:      status,
	}, nil
}

// PullImage pulls an image for a specific service
func (h *ImageHandler) PullImage(serviceName string, progressChan chan<- container.ImagePullProgress) error {
	h.logger.Log(fmt.Sprintf("Pulling image for service: %s", serviceName))

	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return fmt.Errorf("service %s not found", serviceName)
	}

	if service.ImageName == "" {
		return fmt.Errorf("service %s has no image defined", serviceName)
	}

	// Create stop channel for the pull operation
	stopChan := make(chan struct{})

	// Resolve environment variables in image name
	resolvedImageName := h.resolveImageName(service.ImageName)

	// Start the image pull
	err := h.runtime.PullImageWithProgress(resolvedImageName, progressChan, stopChan)
	if err != nil {
		h.logger.Log(fmt.Sprintf("Failed to pull image %s (resolved: %s) for service %s: %v", service.ImageName, resolvedImageName, serviceName, err))
		return err
	}

	h.logger.Log(fmt.Sprintf("Successfully pulled image %s (resolved: %s) for service %s", service.ImageName, resolvedImageName, serviceName))

	// Update the service's image status in the service manager
	h.updateServiceImageStatus(serviceName)

	return nil
}

// PullAllImages pulls all missing images for all services
func (h *ImageHandler) PullAllImages(progressChan chan<- container.ImagePullProgress) error {
	h.logger.Log("Pulling all missing images")

	// Get current image statuses
	imageStatuses, err := h.GetAllImageStatuses()
	if err != nil {
		return fmt.Errorf("failed to get image statuses: %w", err)
	}

	// Pull missing images
	for serviceName, status := range imageStatuses {
		if status.Status == "missing" {
			h.logger.Log(fmt.Sprintf("Pulling missing image for service: %s", serviceName))

			err := h.PullImage(serviceName, progressChan)
			if err != nil {
				h.logger.Log(fmt.Sprintf("Failed to pull image for service %s: %v", serviceName, err))
				// Continue with other images
				continue
			}
		}
	}

	h.logger.Log("Completed pulling all missing images")
	return nil
}

// ListAllImages returns all available images in the system
func (h *ImageHandler) ListAllImages() ([]string, error) {
	h.logger.Log("Listing all available images")

	images, err := h.runtime.ListAllImages()
	if err != nil {
		h.logger.Log(fmt.Sprintf("Failed to list images: %v", err))
		return nil, err
	}

	h.logger.Log(fmt.Sprintf("Found %d images", len(images)))
	return images, nil
}

// GetServiceImages returns image information for all services
func (h *ImageHandler) GetServiceImages() (map[string]ServiceImageInfo, error) {
	h.logger.Log("Getting service image information")

	serviceImages := make(map[string]ServiceImageInfo)

	// Get all enhanced services
	for _, service := range h.serviceManager.ListServices() {
		imageInfo := ServiceImageInfo{
			ServiceName: service.Name,
			ImageName:   service.ImageName,
		}

		if service.ImageName != "" {
			// Check if image exists locally
			exists, err := h.runtime.CheckImageExists(service.ImageName)
			if err != nil {
				imageInfo.Status = "error"
				imageInfo.Error = err.Error()
			} else {
				if exists {
					imageInfo.Status = "available"
				} else {
					imageInfo.Status = "missing"
				}
			}
		} else {
			imageInfo.Status = "no_image"
		}

		serviceImages[service.Name] = imageInfo
	}

	h.logger.Log(fmt.Sprintf("Retrieved image information for %d services", len(serviceImages)))
	return serviceImages, nil
}

// RefreshImageStatuses refreshes image status information from the runtime
func (h *ImageHandler) RefreshImageStatuses() (map[string]ImageStatus, error) {
	h.logger.Log("Refreshing image statuses")

	// Reload services to pick up any changes
	if err := h.serviceManager.LoadServices(); err != nil {
		h.logger.Log(fmt.Sprintf("Warning: Failed to reload services: %v", err))
	}

	// Get fresh image statuses
	return h.GetAllImageStatuses()
}

// ImageStatus represents the status of an image for a service
type ImageStatus struct {
	ServiceName string `json:"service_name"`
	ImageName   string `json:"image_name,omitempty"`
	Status      string `json:"status"` // "available", "missing", "error", "no_image"
	Error       string `json:"error,omitempty"`
}

// ServiceImageInfo represents image information for a service
type ServiceImageInfo struct {
	ServiceName string `json:"service_name"`
	ImageName   string `json:"image_name,omitempty"`
	Status      string `json:"status"` // "available", "missing", "error", "no_image"
	Error       string `json:"error,omitempty"`
}

// Service-based image management methods

// CheckImageExists checks if an image exists for a service
func (h *ImageHandler) CheckImageExists(serviceName string) (bool, error) {
	status, err := h.GetImageStatus(serviceName)
	if err != nil {
		return false, err
	}
	return status.Status == "available", nil
}

// GetAllImages returns all available images
func (h *ImageHandler) GetAllImages() ([]string, error) {
	return h.ListAllImages()
}

// StartImagePull starts pulling an image for a service
func (h *ImageHandler) StartImagePull(serviceName string) error {
	// Create a progress channel to receive updates
	progressChan := make(chan container.ImagePullProgress, 10)

	// Start pull in background and handle progress updates
	go func() {
		defer close(progressChan)

		// Start the pull operation
		go func() {
			h.PullImage(serviceName, progressChan)
		}()

		// Process progress updates and broadcast them
		for progress := range progressChan {
			h.logger.Log(fmt.Sprintf("Image pull progress for %s: %.1f%% - %s", serviceName, progress.Progress, progress.Status))

			// Broadcast progress via callback if available
			if h.progressCallback != nil {
				// Get the actual image name for the service
				service, exists := h.serviceManager.GetService(serviceName)
				imageName := serviceName // fallback
				if exists && service.ImageName != "" {
					imageName = service.ImageName
				}
				h.progressCallback(serviceName, imageName, progress.Progress, progress.Status)
			}
		}
	}()

	return nil
}

// StopImagePull stops image pull for a service
func (h *ImageHandler) StopImagePull(serviceName string) error {
	// In simplified implementation, we don't track individual pulls
	// Just return success
	h.logger.Log(fmt.Sprintf("Stop image pull requested for service: %s", serviceName))
	return nil
}

// GetImagePullProgress returns image pull progress for a service
func (h *ImageHandler) GetImagePullProgress(serviceName string) (*models.ImagePullProgress, error) {
	// In simplified implementation, just return current status
	status, err := h.GetImageStatus(serviceName)
	if err != nil {
		return &models.ImagePullProgress{
			Status:      "error",
			ServiceName: serviceName,
		}, err
	}

	var progressStatus string
	switch status.Status {
	case "available":
		progressStatus = "complete"
	case "missing":
		progressStatus = "idle"
	case "error":
		progressStatus = "error"
	default:
		progressStatus = "idle"
	}

	return &models.ImagePullProgress{
		Status:      progressStatus,
		ServiceName: serviceName,
	}, nil
}

// updateServiceImageStatus updates the image status for a service after successful operations
func (h *ImageHandler) updateServiceImageStatus(serviceName string) {
	service, exists := h.serviceManager.GetService(serviceName)
	if !exists {
		return
	}

	if service.ImageName == "" {
		service.ImageExists = false
		return
	}

	// Resolve environment variables and check if image exists
	resolvedImageName := h.resolveImageName(service.ImageName)
	exists, err := h.runtime.CheckImageExists(resolvedImageName)
	if err != nil {
		h.logger.Log(fmt.Sprintf("Failed to check image status for %s: %v", serviceName, err))
		return
	}

	// Update the service's image status
	service.ImageExists = exists
	h.logger.Log(fmt.Sprintf("Updated image status for %s: exists=%v", serviceName, exists))
}
