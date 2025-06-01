package main

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestApp_NewApp(t *testing.T) {
	app := NewApp()

	if app == nil {
		t.Fatal("Expected app to be created, got nil")
	}
	if app.ctx != nil {
		t.Error("Expected ctx to be nil initially")
	}
	if app.handlerManager == nil {
		t.Error("Expected handlerManager to be initialized")
	}
	if app.config == nil {
		t.Error("Expected config to be initialized")
	}
	if app.initError != nil {
		t.Error("Expected initError to be nil initially")
	}
}

func TestApp_startup_Success(t *testing.T) {
	app := NewApp()
	ctx := context.Background()

	// Set a temporary home directory for testing
	tempDir := t.TempDir()
	os.Setenv("HOME", tempDir)
	defer os.Unsetenv("HOME")

	// For native apps, startup will fail without a bundled CLI
	app.startup(ctx)

	if app.ctx != ctx {
		t.Errorf("Expected ctx to be set")
	}
	// For native apps without bundled CLI, initError should be set
	if app.initError == nil {
		t.Error("Expected init error since no bundled CLI is available in test environment")
	}
}

func TestApp_checkReady_InitError(t *testing.T) {
	app := NewApp()
	app.initError = errors.New("init failed")

	err := app.checkReady()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if err.Error() != "init failed" {
		t.Errorf("Expected error 'init failed', got %s", err.Error())
	}
}

func TestApp_ListServices_NotInitialized(t *testing.T) {
	app := NewApp()
	app.initError = errors.New("not initialized")

	services := app.ListServices()

	if len(services) != 0 {
		t.Error("Expected empty services list when not initialized")
	}
}

func TestApp_GetServiceStatus_NotInitialized(t *testing.T) {
	app := NewApp()
	app.initError = errors.New("not initialized")

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
	app.initError = errors.New("not initialized")

	err := app.StartService("postgres", false)

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestApp_StopService_NotInitialized(t *testing.T) {
	app := NewApp()
	app.initError = errors.New("not initialized")

	err := app.StopService("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
}

func TestApp_GetServiceConnectionInfo_NotInitialized(t *testing.T) {
	app := NewApp()
	app.initError = errors.New("not initialized")

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
	app.initError = errors.New("not initialized")

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
	app.initError = errors.New("not initialized")

	exists, err := app.CheckImageExists("postgres")

	if err == nil {
		t.Fatal("Expected error when not initialized")
	}
	if exists {
		t.Error("Expected exists to be false when not initialized")
	}
}

func TestApp_BasicFunctionality(t *testing.T) {
	app := NewApp()

	// Test that basic methods work even when not fully initialized
	services := app.ListServices()
	if services == nil {
		t.Error("Expected services list to be returned")
	}

	// Test service status (should return error when not initialized)
	_, err := app.GetServiceStatus("postgres")
	if err == nil {
		t.Error("Expected error when not initialized")
	}

	// Test connection info (should return error when not initialized)
	_, err = app.GetServiceConnectionInfo("postgres")
	if err == nil {
		t.Error("Expected error when not initialized")
	}

	// Test logs (should return error when not initialized)
	_, err = app.GetServiceLogs("postgres", 100)
	if err == nil {
		t.Error("Expected error when not initialized")
	}

	// Test image check (should return error when not initialized)
	_, err = app.CheckImageExists("postgres")
	if err == nil {
		t.Error("Expected error when not initialized")
	}
}
