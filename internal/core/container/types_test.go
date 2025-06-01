package container

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestComposeServiceJSONMarshaling(t *testing.T) {
	tests := []struct {
		name     string
		service  ComposeService
		wantJSON string
	}{
		{
			name: "service with dependencies and image",
			service: func() ComposeService {
				svc := CreateTestComposeService([]string{"db"})
				svc.ContainerName = "web_container"
				svc.Image = "nginx:latest"
				return svc
			}(),
			wantJSON: `{"depends_on":{"db":{"condition":"service_started"}},"container_name":"web_container","image":"nginx:latest"}`,
		},
		{
			name: "service with only image",
			service: ComposeService{
				Image: "postgres:13",
			},
			wantJSON: `{"depends_on":null,"image":"postgres:13"}`,
		},
		{
			name:     "empty service",
			service:  ComposeService{},
			wantJSON: `{"depends_on":null}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.service)
			if err != nil {
				t.Fatalf("Failed to marshal ComposeService: %v", err)
			}

			if string(data) != tt.wantJSON {
				t.Errorf("Expected JSON %s, got %s", tt.wantJSON, string(data))
			}

			// Test unmarshaling back
			var unmarshaled ComposeService
			err = json.Unmarshal(data, &unmarshaled)
			if err != nil {
				t.Fatalf("Failed to unmarshal ComposeService: %v", err)
			}

			if !reflect.DeepEqual(tt.service, unmarshaled) {
				t.Errorf("Unmarshaled service doesn't match original. Expected %+v, got %+v", tt.service, unmarshaled)
			}
		})
	}
}

func TestComposeConfigJSONMarshaling(t *testing.T) {
	config := CreateTestComposeConfig(map[string][]string{
		"web": {"db"},
		"db":  {},
	})
	// Add images to the services
	config.Services["web"] = func() ComposeService {
		svc := config.Services["web"]
		svc.Image = "nginx:latest"
		return svc
	}()
	config.Services["db"] = func() ComposeService {
		svc := config.Services["db"]
		svc.Image = "postgres:13"
		return svc
	}()

	data, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal ComposeConfig: %v", err)
	}

	// Test unmarshaling back
	var unmarshaled ComposeConfig
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ComposeConfig: %v", err)
	}

	if !reflect.DeepEqual(config, unmarshaled) {
		t.Errorf("Unmarshaled config doesn't match original. Expected %+v, got %+v", config, unmarshaled)
	}
}

func TestImagePullProgressStruct(t *testing.T) {
	progress := ImagePullProgress{
		Status:       "downloading",
		Progress:     45.5,
		CurrentLayer: "layer123",
		TotalLayers:  10,
		Downloaded:   1024 * 1024,     // 1MB
		Total:        5 * 1024 * 1024, // 5MB
		Speed:        "500 KB/s",
		ETA:          "8s",
	}

	// Test JSON marshaling
	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("Failed to marshal ImagePullProgress: %v", err)
	}

	// Test unmarshaling back
	var unmarshaled ImagePullProgress
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal ImagePullProgress: %v", err)
	}

	if !reflect.DeepEqual(progress, unmarshaled) {
		t.Errorf("Unmarshaled progress doesn't match original. Expected %+v, got %+v", progress, unmarshaled)
	}
}

func TestImagePullProgressError(t *testing.T) {
	progress := ImagePullProgress{
		Status: "error",
		Error:  "Failed to pull image: network timeout",
	}

	// Test JSON marshaling
	data, err := json.Marshal(progress)
	if err != nil {
		t.Fatalf("Failed to marshal ImagePullProgress with error: %v", err)
	}

	// Verify error field is included
	expectedJSON := `{"status":"error","progress":0,"currentLayer":"","totalLayers":0,"downloaded":0,"total":0,"speed":"","eta":"","error":"Failed to pull image: network timeout"}`
	if string(data) != expectedJSON {
		t.Errorf("Expected JSON %s, got %s", expectedJSON, string(data))
	}
}

// Provider tests are covered in provider_test.go

func TestDockerRuntimeStruct(t *testing.T) {
	runtime := NewDockerRuntime()

	if runtime == nil {
		t.Fatal("NewDockerRuntime returned nil")
	}

	if runtime.Name() != "docker" {
		t.Errorf("Expected runtime name 'docker', got '%s'", runtime.Name())
	}

	// Test that caching fields are properly initialized as nil/empty
	if runtime.parsedComposeConfig != nil {
		t.Error("Expected parsedComposeConfig to be nil initially")
	}

	if runtime.cachedComposeFilesKey != "" {
		t.Error("Expected cachedComposeFilesKey to be empty initially")
	}
}

func TestPodmanRuntimeStruct(t *testing.T) {
	runtime := NewPodmanRuntime()

	if runtime == nil {
		t.Fatal("NewPodmanRuntime returned nil")
	}

	if runtime.Name() != "podman" {
		t.Errorf("Expected runtime name 'podman', got '%s'", runtime.Name())
	}

	// Test that caching fields are properly initialized as nil/empty
	if runtime.parsedComposeConfig != nil {
		t.Error("Expected parsedComposeConfig to be nil initially")
	}

	if runtime.cachedComposeFilesKey != "" {
		t.Error("Expected cachedComposeFilesKey to be empty initially")
	}
}
