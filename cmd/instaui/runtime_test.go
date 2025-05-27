package main

import (
	"context"
	"testing"
)

func TestApp_GetRuntimeStatus(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	// Test GetRuntimeStatus method
	status := app.GetRuntimeStatus()

	// Should return a valid status object
	if status == nil {
		t.Fatal("GetRuntimeStatus returned nil")
	}

	// Should have platform information
	if status.Platform == "" {
		t.Error("Platform should not be empty")
	}

	// Should have runtime statuses
	if status.RuntimeStatuses == nil {
		t.Error("RuntimeStatuses should not be nil")
	}

	// Should have at least one runtime status (even if not available)
	if len(status.RuntimeStatuses) == 0 {
		t.Error("Should have at least one runtime status")
	}

	// Should have recommended action
	if status.RecommendedAction == "" {
		t.Error("RecommendedAction should not be empty")
	}

	t.Logf("Platform: %s", status.Platform)
	t.Logf("Can proceed: %v", status.CanProceed)
	t.Logf("Preferred runtime: %s", status.PreferredRuntime)
	t.Logf("Recommended action: %s", status.RecommendedAction)
	t.Logf("Number of runtime statuses: %d", len(status.RuntimeStatuses))

	for _, rt := range status.RuntimeStatuses {
		t.Logf("Runtime %s: installed=%v, running=%v, available=%v",
			rt.Name, rt.IsInstalled, rt.IsRunning, rt.IsAvailable)
	}
}

func TestApp_AttemptStartRuntime(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	// Test AttemptStartRuntime with a non-existent runtime
	result := app.AttemptStartRuntime("nonexistent")

	// Should return a result object
	if result == nil {
		t.Fatal("AttemptStartRuntime returned nil")
	}

	// Should not be successful for non-existent runtime
	if result.Success {
		t.Error("Should not be successful for non-existent runtime")
	}

	// Should have an error message
	if result.Error == "" {
		t.Error("Should have an error message for non-existent runtime")
	}

	t.Logf("Result for nonexistent runtime: success=%v, error=%s", result.Success, result.Error)
}

func TestApp_WaitForRuntimeReady(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	// Test WaitForRuntimeReady with a non-existent runtime
	result := app.WaitForRuntimeReady("nonexistent", 1)

	// Should return a result object
	if result == nil {
		t.Fatal("WaitForRuntimeReady returned nil")
	}

	// Should not be successful for non-existent runtime
	if result.Success {
		t.Error("Should not be successful for non-existent runtime")
	}

	t.Logf("Wait result for nonexistent runtime: success=%v, error=%s", result.Success, result.Error)
}

func TestApp_ReinitializeRuntime(t *testing.T) {
	app := NewApp()
	app.startup(context.Background())

	// Test ReinitializeRuntime
	err := app.ReinitializeRuntime()

	// Should not panic and should return an error (since no runtime is available in test)
	if err == nil {
		t.Log("ReinitializeRuntime succeeded (runtime available)")
	} else {
		t.Logf("ReinitializeRuntime failed as expected: %v", err)
	}
}
