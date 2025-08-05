package container

import (
	"os"
	"reflect"
	"testing"
)

func TestSetDefaultEnvVars(t *testing.T) {
	expectedVars := []string{
		"DB_USER",
		"DB_USER_PASSWORD",
		"ELASTICSEARCH_USER",
		"ELASTICSEARCH_PASSWORD",
		"MYSQL_USER",
		"MYSQL_PASSWORD",
	}

	// Use shared utility for environment variable management
	restore := SaveAndRestoreEnvVars(expectedVars)
	defer restore()

	// Clear environment variables first
	for _, envVar := range expectedVars {
		os.Unsetenv(envVar)
	}

	// Test setting default values
	setDefaultEnvVars()

	// Check that DB_USER is set to "root"
	if dbUser := os.Getenv("DB_USER"); dbUser != "root" {
		t.Errorf("Expected DB_USER to be 'root', got '%s'", dbUser)
	}

	// Check that MYSQL_PASSWORD is set to "root"
	if mysqlPass := os.Getenv("MYSQL_PASSWORD"); mysqlPass != "root" {
		t.Errorf("Expected MYSQL_PASSWORD to be 'root', got '%s'", mysqlPass)
	}

	// Check that ELASTICSEARCH_USER is set to "elastic"
	if esUser := os.Getenv("ELASTICSEARCH_USER"); esUser != "elastic" {
		t.Errorf("Expected ELASTICSEARCH_USER to be 'elastic', got '%s'", esUser)
	}
}

func TestSetPodmanEnvVars(t *testing.T) {
	// Use shared utility for environment variable management
	restore := SaveAndRestoreEnvVars([]string{"COMPOSE_PROVIDER"})
	defer restore()

	// Test setting Podman-specific environment variables
	setPodmanEnvVars()

	// Check that COMPOSE_PROVIDER is set to "podman"
	if provider := os.Getenv("COMPOSE_PROVIDER"); provider != "podman" {
		t.Errorf("Expected COMPOSE_PROVIDER to be 'podman', got '%s'", provider)
	}
}

func TestParsePortMappings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "single port mapping",
			input: "5432/tcp -> 0.0.0.0:5432",
			expected: map[string]string{
				"5432/tcp": "5432",
			},
		},
		{
			name:  "multiple port mappings",
			input: "5432/tcp -> 0.0.0.0:5432\n3306/tcp -> 0.0.0.0:3306",
			expected: map[string]string{
				"5432/tcp": "5432",
				"3306/tcp": "3306",
			},
		},
		{
			name:  "port mapping with different host port",
			input: "5432/tcp -> 0.0.0.0:15432",
			expected: map[string]string{
				"5432/tcp": "15432",
			},
		},
		{
			name:     "empty input",
			input:    "",
			expected: map[string]string{},
		},
		{
			name:  "port mapping with IPv6",
			input: "5432/tcp -> [::]:5432",
			expected: map[string]string{
				"5432/tcp": "5432",
			},
		},
		{
			name:     "malformed input - no arrow",
			input:    "5432/tcp 0.0.0.0:5432",
			expected: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePortMappings(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestContainersPath(t *testing.T) {
}

func TestGetServiceNames(t *testing.T) {
	tests := []struct {
		name     string
		config   *ComposeConfig
		expected []string
	}{
		{
			name: "multiple services",
			config: &ComposeConfig{
				Services: map[string]ComposeService{
					"web":   {},
					"db":    {},
					"redis": {},
				},
			},
			expected: []string{"db", "redis", "web"}, // Should contain all services (order may vary)
		},
		{
			name: "single service",
			config: &ComposeConfig{
				Services: map[string]ComposeService{
					"web": {},
				},
			},
			expected: []string{"web"},
		},
		{
			name: "no services",
			config: &ComposeConfig{
				Services: map[string]ComposeService{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getServiceNames(tt.config)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d services, got %d: %v vs %v", len(tt.expected), len(result), tt.expected, result)
				return
			}

			// Check that all expected services are present (order may vary due to map iteration)
			for _, expected := range tt.expected {
				found := false
				for _, actual := range result {
					if expected == actual {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected service '%s' not found in result %v", expected, result)
				}
			}
		})
	}
}
