package handlers

import (
	"fmt"
	"testing"
)

func TestDebugBaseHandler(t *testing.T) {
	mockRuntime := &mockContainerRuntime{
		nameFunc: func() string {
			return "mock"
		},
		getContainerNameFunc: func(serviceName string, composeFiles []string) (string, error) {
			return "test_" + serviceName + "_1", nil
		},
		getPortMappingsFunc: func(containerName string) (map[string]string, error) {
			if containerName == "test_postgres_1" {
				return map[string]string{"5432/tcp": "5432"}, nil
			}
			if containerName == "test_redis_1" {
				return map[string]string{"6379/tcp": "6379"}, nil
			}
			return nil, fmt.Errorf("container not running")
		},
	}

	baseHandler := NewBaseHandler(mockRuntime, "/test/insta")

	// Test getRunningContainers
	runningContainers, err := baseHandler.getRunningContainers()
	if err != nil {
		t.Fatalf("getRunningContainers failed: %v", err)
	}

	t.Logf("Running containers: %v", runningContainers)

	// Test isServiceRunning for postgres
	composeFiles := baseHandler.getComposeFiles()
	isRunning := baseHandler.isServiceRunning("postgres", composeFiles, runningContainers)
	t.Logf("postgres isServiceRunning: %v", isRunning)

	// Test isServiceRunning for redis
	isRunning = baseHandler.isServiceRunning("redis", composeFiles, runningContainers)
	t.Logf("redis isServiceRunning: %v", isRunning)
}
