package main

import (
	"context"
	"errors"
	"os"
	"testing"
)

// Mock implementation moved to test_utils.go

// Mock methods moved to test_utils.go

func TestApp_NewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}
	if app.ctx != nil {
		t.Error("Expected ctx to be nil initially")
	}
	if app.containerRuntime != nil {
		t.Error("Expected containerRuntime to be nil initially")
	}
	if app.instaDir != "" {
		t.Error("Expected instaDir to be empty initially")
	}
	if app.runtimeInitError != nil {
		t.Error("Expected runtimeInitError to be nil initially")
	}
}

func TestApp_startup_Success(t *testing.T) {
	app := NewApp()
	ctx := context.Background()

	// Set a temporary home directory for testing
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// Mock a working environment
	app.startup(ctx)

	if app.ctx != ctx {
		t.Errorf("Expected ctx to be set")
	}
	if app.instaDir == "" {
		t.Error("Expected instaDir to be set")
	}
	if app.runtimeInitError == nil {
		// This might fail since we don't have a real container runtime
		// but that's expected in a unit test environment
		t.Log("Note: Container runtime initialization may fail in test environment, which is expected")
	}
}

func TestApp_startup_WithInstaHome(t *testing.T) {
	app := NewApp()
	ctx := context.Background()

	// Set a temporary home directory for testing first
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	customInstaDir := "/tmp/custom-insta"
	os.Setenv("INSTA_HOME", customInstaDir)
	defer os.Unsetenv("INSTA_HOME")

	app.startup(ctx)

	if app.instaDir != customInstaDir {
		t.Errorf("Expected instaDir to be %s, got %s", customInstaDir, app.instaDir)
	}
}

func TestApp_checkInitialization_Success(t *testing.T) {
	app := NewApp()
	app.containerRuntime = NewMockContainerRuntime("mock")
	app.instaDir = "/test/insta"
	app.runtimeInitError = nil

	err := app.checkInitialization()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestApp_checkInitialization_RuntimeInitError(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("runtime failed")

	err := app.checkInitialization()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "runtime failed" {
		t.Errorf("Expected error 'runtime failed', got %s", err.Error())
	}
}

func TestApp_checkInitialization_NoContainerRuntime(t *testing.T) {
	app := NewApp()
	app.containerRuntime = nil
	app.instaDir = "/test/insta"
	app.runtimeInitError = nil

	err := app.checkInitialization()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "neither container runtime nor bundled CLI available" {
		t.Errorf("Expected error 'neither container runtime nor bundled CLI available', got %s", err.Error())
	}
}

func TestApp_checkInitialization_NoInstaDir(t *testing.T) {
	app := NewApp()
	app.containerRuntime = NewMockContainerRuntime("mock")
	app.instaDir = ""
	app.runtimeInitError = nil

	err := app.checkInitialization()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "insta directory not determined" {
		t.Errorf("Expected error 'insta directory not determined', got %s", err.Error())
	}
}

func TestApp_initializeHandlers(t *testing.T) {
	app := NewApp()
	app.containerRuntime = NewMockContainerRuntime("mock")
	app.instaDir = "/test/insta"
	app.ctx = context.Background()

	app.initializeHandlers()

	if app.serviceHandler == nil {
		t.Error("Expected serviceHandler to be initialized")
	}
	if app.connectionHandler == nil {
		t.Error("Expected connectionHandler to be initialized")
	}
	if app.logsHandler == nil {
		t.Error("Expected logsHandler to be initialized")
	}
	if app.imageHandler == nil {
		t.Error("Expected imageHandler to be initialized")
	}
	if app.dependencyHandler == nil {
		t.Error("Expected dependencyHandler to be initialized")
	}
	if app.graphHandler == nil {
		t.Error("Expected graphHandler to be initialized")
	}
}

func TestApp_initializeHandlers_MissingRuntime(t *testing.T) {
	app := NewApp()
	app.containerRuntime = nil
	app.instaDir = "/test/insta"

	app.initializeHandlers()

	// Should not initialize handlers when runtime is missing
	if app.serviceHandler != nil {
		t.Error("Expected serviceHandler to be nil when runtime is missing")
	}
}

func TestApp_initializeHandlers_MissingInstaDir(t *testing.T) {
	app := NewApp()
	app.containerRuntime = NewMockContainerRuntime("mock")
	app.instaDir = ""

	app.initializeHandlers()

	// Should not initialize handlers when instaDir is missing
	if app.serviceHandler != nil {
		t.Error("Expected serviceHandler to be nil when instaDir is missing")
	}
}

// Test app methods that delegate to handlers with initialization check

func TestApp_ListServices_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	services := app.ListServices()

	if len(services) != 0 {
		t.Error("Expected empty services list when not initialized")
	}
}

func TestApp_GetServiceStatus_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	status, err := app.GetServiceStatus("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if status != "error" {
		t.Errorf("Expected status 'error', got '%s'", status)
	}
}

func TestApp_StartService_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	err := app.StartService("postgres", false)

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestApp_StopService_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	err := app.StopService("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestApp_GetServiceConnectionInfo_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	info, err := app.GetServiceConnectionInfo("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if info != nil {
		t.Error("Expected info to be nil when not initialized")
	}
}

func TestApp_GetServiceLogs_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	logs, err := app.GetServiceLogs("postgres", 100)

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if logs != nil {
		t.Error("Expected logs to be nil when not initialized")
	}
}

func TestApp_CheckImageExists_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	exists, err := app.CheckImageExists("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if exists {
		t.Error("Expected exists to be false when not initialized")
	}
}

func TestApp_GetDependencyStatus_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	status, err := app.GetDependencyStatus("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if status != nil {
		t.Error("Expected status to be nil when not initialized")
	}
}

func TestApp_GetDependencyGraph_NotInitialized(t *testing.T) {
	app := NewApp()
	app.runtimeInitError = errors.New("not initialized")

	graph, err := app.GetDependencyGraph()

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if graph != nil {
		t.Error("Expected graph to be nil when not initialized")
	}
}

// Test successful delegation to initialized handlers

func TestApp_SuccessfulDelegation(t *testing.T) {
	app := NewApp()
	app.containerRuntime = NewMockContainerRuntime("mock")
	app.instaDir = "/test/insta"
	app.ctx = context.Background()
	app.runtimeInitError = nil

	// Initialize handlers
	app.initializeHandlers()

	// Test that methods can be called without initialization errors
	// Note: The actual functionality is tested in handler tests

	services := app.ListServices()
	if services == nil {
		t.Error("Expected services list to be returned")
	}

	_, err := app.GetServiceStatus("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.GetServiceDependencies("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	err = app.StartService("postgres", false)
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	err = app.StopService("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.GetServiceConnectionInfo("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.GetServiceLogs("postgres", 100)
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.CheckImageExists("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.GetDependencyStatus("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.GetDependencyGraph()
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}

	_, err = app.GetServiceDependencyGraph("postgres")
	if err != nil {
		t.Errorf("Expected no initialization error, got %v", err)
	}
}
