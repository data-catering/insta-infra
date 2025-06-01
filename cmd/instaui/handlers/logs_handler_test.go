package handlers

import (
	"errors"
	"testing"
)

func TestLogsHandler_NewLogsHandler(t *testing.T) {
	mockRuntime := newMockContainerRuntime()
	instaDir := "/test/insta"
	handler := NewLogsHandler(mockRuntime, instaDir, nil)

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

func TestLogsHandler_GetServiceLogs_Success(t *testing.T) {
	expectedLogs := []string{"log line 1", "log line 2", "log line 3"}
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			return expectedLogs, nil
		})
	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	logs, err := handler.GetServiceLogs("postgres", 100)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(logs) != len(expectedLogs) {
		t.Errorf("Expected %d log lines, got %d", len(expectedLogs), len(logs))
	}
	for i, log := range logs {
		if log != expectedLogs[i] {
			t.Errorf("Expected log line %d to be '%s', got '%s'", i, expectedLogs[i], log)
		}
	}
}

func TestLogsHandler_GetServiceLogs_ContainerNameError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "", errors.New("service not found")
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			// Simulate that logs can be retrieved using the fallback service name
			if containerName == "nonexistent" {
				return []string{"fallback log line 1", "fallback log line 2"}, nil
			}
			return nil, errors.New("container not found")
		})
	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	logs, err := handler.GetServiceLogs("nonexistent", 100)

	if err != nil {
		t.Fatalf("Expected no error with fallback, got %v", err)
	}
	if len(logs) != 2 {
		t.Errorf("Expected 2 log lines from fallback, got %d", len(logs))
	}
	if logs[0] != "fallback log line 1" {
		t.Errorf("Expected first log to be 'fallback log line 1', got '%s'", logs[0])
	}
}

func TestLogsHandler_GetServiceLogs_BothContainerNameAndFallbackFail(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "resolved_container_name", nil
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			// Both resolved name and fallback service name fail
			return nil, errors.New("container not found")
		})
	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	logs, err := handler.GetServiceLogs("nonexistent", 100)

	if err == nil {
		t.Fatal("Expected error when both container names fail, got nil")
	}
	if logs != nil {
		t.Errorf("Expected logs to be nil, got %v", logs)
	}
	if !contains(err.Error(), "tried both 'resolved_container_name' and 'nonexistent'") {
		t.Errorf("Expected error to mention both container names tried, got %s", err.Error())
	}
}

func TestLogsHandler_GetServiceLogs_FallbackToServiceNameSuccess(t *testing.T) {
	expectedLogs := []string{"fallback log 1", "fallback log 2"}
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "resolved_container_name", nil
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			// Resolved container name fails, but service name succeeds
			if containerName == "resolved_container_name" {
				return nil, errors.New("resolved container not found")
			}
			if containerName == "postgres" {
				return expectedLogs, nil
			}
			return nil, errors.New("container not found")
		})
	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	logs, err := handler.GetServiceLogs("postgres", 100)

	if err != nil {
		t.Fatalf("Expected no error with fallback, got %v", err)
	}
	if len(logs) != len(expectedLogs) {
		t.Errorf("Expected %d log lines, got %d", len(expectedLogs), len(logs))
	}
	for i, log := range logs {
		if log != expectedLogs[i] {
			t.Errorf("Expected log line %d to be '%s', got '%s'", i, expectedLogs[i], log)
		}
	}
}

func TestLogsHandler_GetServiceLogs_GetLogsError(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			return nil, errors.New("container not running")
		})
	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	logs, err := handler.GetServiceLogs("postgres", 100)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if logs != nil {
		t.Errorf("Expected logs to be nil, got %v", logs)
	}
	if !contains(err.Error(), "failed to get logs for service") {
		t.Errorf("Expected error to contain 'failed to get logs for service', got %s", err.Error())
	}
}

func TestLogsHandler_StartLogStream_Success(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		}).
		withStreamContainerLogs(func(containerName string, logChan chan<- string, stopChan <-chan struct{}) error {
			// Simulate successful streaming with proper channel handling
			go func() {
				defer func() {
					// Recover from any panic if logChan is closed
					if r := recover(); r != nil {
						// Channel was closed, test is ending
					}
				}()

				select {
				case logChan <- "test log line":
				case <-stopChan:
					return
				}
			}()
			return nil
		})

	// Set the mock to return running containers so the optimized approach works
	mockRuntime.defaultContainerStatus = "running"

	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartLogStream("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that the stream was registered
	handler.logStreamsMutex.RLock()
	_, exists := handler.logStreams["postgres"]
	handler.logStreamsMutex.RUnlock()

	if !exists {
		t.Error("Expected log stream to be registered")
	}

	// Clean up the stream to avoid race conditions
	handler.StopLogStream("postgres")
}

func TestLogsHandler_StartLogStream_StoppedServiceWithLogs(t *testing.T) {
	expectedLogs := []string{"historical log 1", "historical log 2"}
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_nonexistent_1", nil
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			// Simulate that the stopped container has historical logs
			return expectedLogs, nil
		})

	// Set the mock to return no running containers so the service appears stopped
	mockRuntime.defaultContainerStatus = "stopped"

	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartLogStream("nonexistent")

	// Should not return an error for stopped containers with logs
	if err != nil {
		t.Fatalf("Expected no error for stopped container with logs, got %v", err)
	}

	// Check that no stream was registered (since service is stopped, no live streaming)
	handler.logStreamsMutex.RLock()
	_, exists := handler.logStreams["nonexistent"]
	handler.logStreamsMutex.RUnlock()

	if exists {
		t.Error("Expected no log stream to be registered for stopped service")
	}
}

func TestLogsHandler_StartLogStream_StoppedServiceNoLogs(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_nonexistent_1", nil
		}).
		withGetContainerLogs(func(containerName string, tailLines int) ([]string, error) {
			// Simulate that the container doesn't exist or has no logs
			return nil, errors.New("container not found")
		})

	// Set the mock to return no running containers so the service appears stopped
	mockRuntime.defaultContainerStatus = "stopped"

	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	err := handler.StartLogStream("nonexistent")

	// Should not return an error even if no logs are available
	if err != nil {
		t.Fatalf("Expected no error even when no logs available, got %v", err)
	}

	// Check that no stream was registered
	handler.logStreamsMutex.RLock()
	_, exists := handler.logStreams["nonexistent"]
	handler.logStreamsMutex.RUnlock()

	if exists {
		t.Error("Expected no log stream to be registered")
	}
}

func TestLogsHandler_StartLogStream_AlreadyActive(t *testing.T) {
	mockRuntime := newMockContainerRuntime().
		withGetContainerName(func(serviceName string, composeFiles []string) (string, error) {
			return "test_postgres_1", nil
		})
	handler := NewLogsHandler(mockRuntime, "/test/insta", nil)

	// Manually add a stream to simulate it already being active
	handler.logStreamsMutex.Lock()
	handler.logStreams["postgres"] = make(chan struct{})
	handler.logStreamsMutex.Unlock()

	err := handler.StartLogStream("postgres")

	if err == nil {
		t.Fatal("Expected error for already active stream, got nil")
	}
	if !contains(err.Error(), "log stream already active") {
		t.Errorf("Expected error to contain 'log stream already active', got %s", err.Error())
	}
}

func TestLogsHandler_StopLogStream_Success(t *testing.T) {
	handler := NewLogsHandler(newMockContainerRuntime(), "/test/insta", nil)

	// Manually add a stream
	stopChan := make(chan struct{})
	handler.logStreamsMutex.Lock()
	handler.logStreams["postgres"] = stopChan
	handler.logStreamsMutex.Unlock()

	err := handler.StopLogStream("postgres")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that the stream was removed
	handler.logStreamsMutex.RLock()
	_, exists := handler.logStreams["postgres"]
	handler.logStreamsMutex.RUnlock()

	if exists {
		t.Error("Expected log stream to be removed")
	}
}

func TestLogsHandler_StopLogStream_NotFound(t *testing.T) {
	handler := NewLogsHandler(newMockContainerRuntime(), "/test/insta", nil)

	err := handler.StopLogStream("nonexistent")

	if err == nil {
		t.Fatal("Expected error for non-existent stream, got nil")
	}
	if !contains(err.Error(), "no active log stream found") {
		t.Errorf("Expected error to contain 'no active log stream found', got %s", err.Error())
	}
}

func TestLogsHandler_getComposeFiles(t *testing.T) {
	handler := NewLogsHandler(nil, "/test/insta", nil)

	composeFiles := handler.getComposeFiles()

	if len(composeFiles) == 0 {
		t.Error("Expected at least one compose file")
	}
	if !contains(composeFiles[0], "/test/insta/docker-compose.yaml") {
		t.Errorf("Expected first compose file to contain '/test/insta/docker-compose.yaml', got %s", composeFiles[0])
	}
}
