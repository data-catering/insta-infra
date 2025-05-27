package handlers

import (
	"errors"
	"testing"

	"github.com/data-catering/insta-infra/v2/internal/core/container"
)

func TestImageHandler_NewImageHandler(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	instaDir := "/test/insta"

	handler := NewImageHandler(mockRuntime, instaDir, nil)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
	if handler.containerRuntime != mockRuntime {
		t.Errorf("Expected containerRuntime to be %v, got %v", mockRuntime, handler.containerRuntime)
	}
	if handler.instaDir != instaDir {
		t.Errorf("Expected instaDir to be %s, got %s", instaDir, handler.instaDir)
	}
	if handler.logStreams == nil {
		t.Error("Expected logStreams map to be initialized")
	}
}

func TestImageHandler_CheckImageExists_Success(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "postgres:13", nil
		},
		checkImageExistsFunc: func(imageName string) (bool, error) {
			return true, nil
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	exists, err := handler.CheckImageExists("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected image to exist")
	}
}

func TestImageHandler_CheckImageExists_NotFound(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "postgres:13", nil
		},
		checkImageExistsFunc: func(imageName string) (bool, error) {
			return false, nil
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	exists, err := handler.CheckImageExists("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if exists {
		t.Error("Expected image to not exist")
	}
}

func TestImageHandler_CheckImageExists_GetImageInfoError(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found in compose")
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	exists, err := handler.CheckImageExists("nonexistent")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if exists {
		t.Error("Expected image to not exist when there's an error")
	}
	if !contains(err.Error(), "failed to get image info") {
		t.Errorf("Expected error to contain 'failed to get image info', got %s", err.Error())
	}
}

func TestImageHandler_CheckImageExists_CheckError(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "postgres:13", nil
		},
		checkImageExistsFunc: func(imageName string) (bool, error) {
			return false, errors.New("docker daemon not available")
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	exists, err := handler.CheckImageExists("postgres")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if exists {
		t.Error("Expected image to not exist when there's an error")
	}
	if !contains(err.Error(), "failed to check if image exists") {
		t.Errorf("Expected error to contain 'failed to check if image exists', got %s", err.Error())
	}
}

func TestImageHandler_GetImagePullProgress_Idle(t *testing.T) {
	handler := NewImageHandler(newMockContainerRuntime(), "/test/insta", nil)

	progress, err := handler.GetImagePullProgress("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if progress == nil {
		t.Fatal("Expected progress to be returned, got nil")
	}
	if progress.Status != "idle" {
		t.Errorf("Expected status 'idle', got '%s'", progress.Status)
	}
	if progress.ServiceName != "postgres" {
		t.Errorf("Expected service name 'postgres', got '%s'", progress.ServiceName)
	}
}

func TestImageHandler_GetImagePullProgress_InProgress(t *testing.T) {
	handler := NewImageHandler(newMockContainerRuntime(), "/test/insta", nil)

	// Manually add a pull stream to simulate in-progress
	streamKey := "image-pull-postgres"
	handler.logStreamsMutex.Lock()
	handler.logStreams[streamKey] = make(chan struct{})
	handler.logStreamsMutex.Unlock()

	progress, err := handler.GetImagePullProgress("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if progress == nil {
		t.Fatal("Expected progress to be returned, got nil")
	}
	if progress.Status != "downloading" {
		t.Errorf("Expected status 'downloading', got '%s'", progress.Status)
	}
	if progress.ServiceName != "postgres" {
		t.Errorf("Expected service name 'postgres', got '%s'", progress.ServiceName)
	}
}

func TestImageHandler_StartImagePull_Success(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "postgres:13", nil
		},
		pullImageWithProgressFunc: func(imageName string, progressChan chan<- container.ImagePullProgress, stopChan <-chan struct{}) error {
			// Simulate successful pull with proper channel handling
			go func() {
				defer func() {
					// Recover from any panic if progressChan is closed
					if r := recover(); r != nil {
						// Channel was closed, test is ending
					}
				}()

				select {
				case progressChan <- container.ImagePullProgress{
					Status:   "downloading",
					Progress: 50.0,
				}:
				case <-stopChan:
					return
				}

				select {
				case progressChan <- container.ImagePullProgress{
					Status:   "complete",
					Progress: 100.0,
				}:
				case <-stopChan:
					return
				}
			}()
			return nil
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartImagePull("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that the pull stream was registered
	streamKey := "image-pull-postgres"
	handler.logStreamsMutex.RLock()
	_, exists := handler.logStreams[streamKey]
	handler.logStreamsMutex.RUnlock()

	if !exists {
		t.Error("Expected image pull stream to be registered")
	}

	// Clean up the stream to avoid race conditions
	handler.StopImagePull("postgres")
}

func TestImageHandler_StartImagePull_GetImageInfoError(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartImagePull("nonexistent")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), "failed to get image info") {
		t.Errorf("Expected error to contain 'failed to get image info', got %s", err.Error())
	}
}

func TestImageHandler_StartImagePull_AlreadyInProgress(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "postgres:13", nil
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	// Manually add a pull stream to simulate already in progress
	streamKey := "image-pull-postgres"
	handler.logStreamsMutex.Lock()
	handler.logStreams[streamKey] = make(chan struct{})
	handler.logStreamsMutex.Unlock()

	err := handler.StartImagePull("postgres")

	if err == nil {
		t.Fatal("Expected error for already in progress pull, got nil")
	}
	if !contains(err.Error(), "image pull already in progress") {
		t.Errorf("Expected error to contain 'image pull already in progress', got %s", err.Error())
	}
}

func TestImageHandler_StopImagePull_Success(t *testing.T) {
	handler := NewImageHandler(newMockContainerRuntime(), "/test/insta", nil)

	// Manually add a pull stream
	streamKey := "image-pull-postgres"
	stopChan := make(chan struct{})
	handler.logStreamsMutex.Lock()
	handler.logStreams[streamKey] = stopChan
	handler.logStreamsMutex.Unlock()

	err := handler.StopImagePull("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that the stream was removed
	handler.logStreamsMutex.RLock()
	_, exists := handler.logStreams[streamKey]
	handler.logStreamsMutex.RUnlock()

	if exists {
		t.Error("Expected image pull stream to be removed")
	}
}

func TestImageHandler_StopImagePull_NotFound(t *testing.T) {
	handler := NewImageHandler(newMockContainerRuntime(), "/test/insta", nil)

	err := handler.StopImagePull("nonexistent")

	if err == nil {
		t.Fatal("Expected error for non-existent pull, got nil")
	}
	if !contains(err.Error(), "no active image pull found") {
		t.Errorf("Expected error to contain 'no active image pull found', got %s", err.Error())
	}
}

func TestImageHandler_GetImageInfo_Success(t *testing.T) {
	expectedImageName := "postgres:13"
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return expectedImageName, nil
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	imageName, err := handler.GetImageInfo("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if imageName != expectedImageName {
		t.Errorf("Expected image name '%s', got '%s'", expectedImageName, imageName)
	}
}

func TestImageHandler_GetImageInfo_Error(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		getImageInfoFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		},
	}
	handler := NewImageHandler(mockRuntime, "/test/insta", nil)

	imageName, err := handler.GetImageInfo("nonexistent")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if imageName != "" {
		t.Errorf("Expected empty image name, got '%s'", imageName)
	}
	if !contains(err.Error(), "failed to get image info") {
		t.Errorf("Expected error to contain 'failed to get image info', got %s", err.Error())
	}
}

func TestImageHandler_getComposeFiles(t *testing.T) {
	handler := NewImageHandler(nil, "/test/insta", nil)

	composeFiles := handler.getComposeFiles()

	if len(composeFiles) == 0 {
		t.Error("Expected at least one compose file")
	}
	if !contains(composeFiles[0], "/test/insta/docker-compose.yaml") {
		t.Errorf("Expected first compose file to contain '/test/insta/docker-compose.yaml', got %s", composeFiles[0])
	}
}
