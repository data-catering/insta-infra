package container

import (
	"testing"
)

func TestNewProviderDetailed(t *testing.T) {
	provider := NewProvider()

	if provider == nil {
		t.Fatal("NewProvider returned nil")
	}

	if len(provider.runtimes) != 2 {
		t.Errorf("Expected 2 runtimes, got %d", len(provider.runtimes))
	}

	// Check that selected is initially nil
	if provider.selected != nil {
		t.Error("Expected selected runtime to be nil initially")
	}
}

func TestProviderDetectRuntimeDetailed(t *testing.T) {
	tests := []struct {
		name            string
		runtimes        []Runtime
		expectError     bool
		expectedRuntime string
	}{
		{
			name: "no runtimes available",
			runtimes: []Runtime{
				NewMockRuntime("docker", false),
				NewMockRuntime("podman", false),
			},
			expectError: true,
		},
		{
			name: "docker available first",
			runtimes: []Runtime{
				NewMockRuntime("docker", true),
				NewMockRuntime("podman", false),
			},
			expectError:     false,
			expectedRuntime: "docker",
		},
		{
			name: "podman available when docker is not",
			runtimes: []Runtime{
				NewMockRuntime("docker", false),
				NewMockRuntime("podman", true),
			},
			expectError:     false,
			expectedRuntime: "podman",
		},
		{
			name: "both available - should select first one",
			runtimes: []Runtime{
				NewMockRuntime("docker", true),
				NewMockRuntime("podman", true),
			},
			expectError:     false,
			expectedRuntime: "docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{
				runtimes: tt.runtimes,
			}

			err := provider.DetectRuntime()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError {
				if provider.selected == nil {
					t.Error("Expected runtime to be selected")
				} else if provider.selected.Name() != tt.expectedRuntime {
					t.Errorf("Expected %s runtime, got %s", tt.expectedRuntime, provider.selected.Name())
				}
			}
		})
	}
}

func TestProviderSelectedRuntime(t *testing.T) {
	provider := &Provider{}

	// Test when no runtime is selected
	if selected := provider.SelectedRuntime(); selected != nil {
		t.Error("Expected nil when no runtime selected")
	}

	// Test when runtime is selected
	mockRuntime := NewMockRuntime("docker", true)
	provider.selected = mockRuntime

	if selected := provider.SelectedRuntime(); selected != mockRuntime {
		t.Error("Expected selected runtime to be returned")
	}
}

func TestProviderSetRuntime(t *testing.T) {
	tests := []struct {
		name        string
		runtimes    []Runtime
		setRuntime  string
		expectError bool
	}{
		{
			name: "set existing available runtime",
			runtimes: []Runtime{
				NewMockRuntime("docker", true),
				NewMockRuntime("podman", false),
			},
			setRuntime:  "docker",
			expectError: false,
		},
		{
			name: "set existing unavailable runtime",
			runtimes: []Runtime{
				NewMockRuntime("docker", false),
				NewMockRuntime("podman", true),
			},
			setRuntime:  "docker",
			expectError: true,
		},
		{
			name: "set non-existent runtime",
			runtimes: []Runtime{
				NewMockRuntime("docker", true),
				NewMockRuntime("podman", true),
			},
			setRuntime:  "containerd",
			expectError: true,
		},
		{
			name: "case insensitive runtime name",
			runtimes: []Runtime{
				NewMockRuntime("docker", true),
				NewMockRuntime("podman", false),
			},
			setRuntime:  "DOCKER",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &Provider{
				runtimes: tt.runtimes,
			}

			err := provider.SetRuntime(tt.setRuntime)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !tt.expectError {
				if provider.selected == nil {
					t.Error("Expected runtime to be selected")
				} else if provider.selected.Name() != "docker" { // Assuming lowercase conversion
					t.Errorf("Expected docker runtime to be selected, got %s", provider.selected.Name())
				}
			}
		})
	}
}
