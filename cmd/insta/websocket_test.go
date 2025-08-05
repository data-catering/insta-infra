package main

import (
	"testing"
	"time"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
)

// MockLogger implements Logger interface for testing
type MockWSLogger struct {
	messages []string
}

func (m *MockWSLogger) Log(message string) {
	m.messages = append(m.messages, message)
}

func TestWebSocketBroadcasterCreation(t *testing.T) {
	logger := &MockWSLogger{}
	broadcaster := NewWebSocketBroadcaster(logger)

	if broadcaster == nil {
		t.Error("Expected WebSocket broadcaster to be created, got nil")
	}

	if broadcaster.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if broadcaster.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if broadcaster.broadcast == nil {
		t.Error("Expected broadcast channel to be initialized")
	}
}

func TestWebSocketMessageTypes(t *testing.T) {
	// Test that enhanced service message types are defined
	expectedTypes := []string{
		WSMsgEnhancedServiceUpdate,
		WSMsgServiceDependencyUpdate,
	}

	for _, msgType := range expectedTypes {
		if msgType == "" {
			t.Errorf("Expected message type to be defined, got empty string")
		}
	}
}

func TestServiceStatusUpdateBroadcast(t *testing.T) {
	logger := &MockWSLogger{}
	broadcaster := NewWebSocketBroadcaster(logger)

	// Test broadcasting service status update
	broadcaster.BroadcastServiceStatusUpdate("test-service", "running")

	// Since we don't have actual WebSocket connections, just verify the method doesn't panic
	// and the broadcaster is properly initialized
	if broadcaster.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", broadcaster.GetClientCount())
	}
}

func TestEnhancedServiceUpdateBroadcast(t *testing.T) {
	logger := &MockWSLogger{}
	broadcaster := NewWebSocketBroadcaster(logger)

	// Create a test enhanced service
	service := &models.EnhancedService{
		Name:   "test-service",
		Type:   "Database",
		Status: "running",
	}

	// Test broadcasting enhanced service update
	broadcaster.BroadcastEnhancedServiceUpdate("test-service", service, "status")

	// Verify method doesn't panic and broadcaster is functional
	if broadcaster.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", broadcaster.GetClientCount())
	}
}

func TestAppLogsBroadcast(t *testing.T) {
	logger := &MockWSLogger{}
	broadcaster := NewWebSocketBroadcaster(logger)

	// Test broadcasting application logs
	broadcaster.BroadcastAppLogs("Test log message", "info")

	// Verify method doesn't panic
	if broadcaster.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", broadcaster.GetClientCount())
	}
}

func TestImagePullProgressBroadcast(t *testing.T) {
	logger := &MockWSLogger{}
	broadcaster := NewWebSocketBroadcaster(logger)

	// Test broadcasting image pull progress
	broadcaster.BroadcastImagePullProgress("test-service", "test-image", 50.0, "downloading")

	// Verify method doesn't panic
	if broadcaster.GetClientCount() != 0 {
		t.Errorf("Expected 0 clients, got %d", broadcaster.GetClientCount())
	}
}

func TestWebSocketBroadcasterStartStop(t *testing.T) {
	logger := &MockWSLogger{}
	broadcaster := NewWebSocketBroadcaster(logger)

	// Start the broadcaster
	broadcaster.Start()

	// Give it a moment to start
	time.Sleep(10 * time.Millisecond)

	// Check that start message was logged
	found := false
	for _, msg := range logger.messages {
		if msg == "WebSocket broadcaster started" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected start message to be logged")
	}

	// Stop the broadcaster
	broadcaster.Stop()

	// Give it a moment to stop
	time.Sleep(10 * time.Millisecond)

	// Check that stop message was logged
	found = false
	for _, msg := range logger.messages {
		if msg == "WebSocket broadcaster stopped" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected stop message to be logged")
	}
}
