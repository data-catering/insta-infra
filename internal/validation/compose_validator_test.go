package validation

import (
	"testing"
)

func TestValidateComposeContent_BasicValidation(t *testing.T) {
	validator := NewComposeValidator([]string{"postgres", "redis"})

	tests := []struct {
		name      string
		content   string
		wantValid bool
		wantError string
	}{
		{
			name: "valid compose file",
			content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
			wantValid: true,
		},
		{
			name: "invalid YAML",
			content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  - invalid structure
`,
			wantValid: false,
			wantError: "yaml:",
		},
		{
			name: "missing services section",
			content: `
version: "3.8"
volumes:
  data:
`,
			wantValid: false,
			wantError: "services",
		},
		{
			name: "empty services section",
			content: `
services: {}
`,
			wantValid: false,
			wantError: "empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateComposeContent(tt.content)
			
			if result.Valid != tt.wantValid {
				t.Errorf("ValidateComposeContent() valid = %v, want %v", result.Valid, tt.wantValid)
			}
			
			if !tt.wantValid && tt.wantError != "" {
				hasExpectedError := false
				for _, err := range result.Errors {
					if containsString(err.Message, tt.wantError) {
						hasExpectedError = true
						break
					}
				}
				if !hasExpectedError {
					t.Errorf("ValidateComposeContent() expected error containing '%s', got errors: %v", tt.wantError, result.Errors)
				}
			}
		})
	}
}

func TestValidateComposeContent_ServiceValidation(t *testing.T) {
	validator := NewComposeValidator([]string{"postgres", "redis"})

	tests := []struct {
		name           string
		content        string
		expectErrors   int
		expectWarnings int
	}{
		{
			name: "service without image or build",
			content: `
services:
  web:
    ports:
      - "8080:80"
`,
			expectErrors: 1,
		},
		{
			name: "invalid port mapping",
			content: `
services:
  web:
    image: nginx
    ports:
      - "invalid:port"
`,
			expectErrors: 1,
		},
		{
			name: "valid service with dependencies",
			content: `
services:
  web:
    image: nginx
    depends_on:
      - postgres
  postgres:
    image: postgres:13
`,
			expectErrors: 0,
		},
		{
			name: "dependency on external service",
			content: `
services:
  web:
    image: nginx
    depends_on:
      - redis
`,
			expectErrors:   0,
			expectWarnings: 0, // Should have suggestions about external dependency
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidateComposeContent(tt.content)
			
			if len(result.Errors) != tt.expectErrors {
				t.Errorf("ValidateComposeContent() errors = %d, want %d. Errors: %v", 
					len(result.Errors), tt.expectErrors, result.Errors)
			}
			
			if tt.expectWarnings > 0 && len(result.Warnings) != tt.expectWarnings {
				t.Errorf("ValidateComposeContent() warnings = %d, want %d. Warnings: %v", 
					len(result.Warnings), tt.expectWarnings, result.Warnings)
			}
		})
	}
}

func TestValidateComposeContent_PortConflicts(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web1:
    image: nginx
    ports:
      - "8080:80"
  web2:
    image: nginx
    ports:
      - "8080:80"
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should have port conflict error
	hasPortConflict := false
	for _, err := range result.Errors {
		if err.Type == "port_conflict" {
			hasPortConflict = true
			break
		}
	}
	
	if !hasPortConflict {
		t.Error("Expected port conflict error, but didn't find one")
	}
}

func TestValidateComposeContent_CircularDependencies(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web:
    image: nginx
    depends_on:
      - api
  api:
    image: node
    depends_on:
      - web
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should have circular dependency error
	hasCircularDep := false
	for _, err := range result.Errors {
		if err.Type == "circular_dependency" {
			hasCircularDep = true
			break
		}
	}
	
	if !hasCircularDep {
		t.Error("Expected circular dependency error, but didn't find one")
	}
}

func TestValidateComposeContent_BestPractices(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web:
    image: nginx
    environment:
      - ENV1=value1
      - ENV2=value2
      - ENV3=value3
      - ENV4=value4
      - ENV5=value5
      - ENV6=value6
      - ENV7=value7
      - ENV8=value8
      - ENV9=value9
      - ENV10=value10
      - ENV11=value11
  db:
    image: postgres:13
    volumes:
      - ./data:/var/lib/postgresql/data
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should have suggestions
	if len(result.Suggestions) == 0 {
		t.Error("Expected suggestions for best practices, but got none")
	}
	
	// Should suggest env file for many environment variables
	hasEnvFileSuggestion := false
	for _, suggestion := range result.Suggestions {
		if suggestion.Type == "use_env_file" {
			hasEnvFileSuggestion = true
			break
		}
	}
	
	if !hasEnvFileSuggestion {
		t.Error("Expected env file suggestion for many environment variables")
	}
}

func TestValidateComposeContent_HealthchecksAndRestart(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web:
    image: nginx
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
  db:
    image: postgres:13
    restart: invalid-policy
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should have restart policy error
	hasRestartError := false
	for _, err := range result.Errors {
		if err.Type == "invalid_restart_policy" {
			hasRestartError = true
			break
		}
	}
	
	if !hasRestartError {
		t.Error("Expected invalid restart policy error")
	}
	
	// Should suggest healthcheck for database
	hasHealthcheckSuggestion := false
	for _, suggestion := range result.Suggestions {
		if suggestion.Type == "add_healthcheck" && suggestion.ServiceName == "db" {
			hasHealthcheckSuggestion = true
			break
		}
	}
	
	if !hasHealthcheckSuggestion {
		t.Error("Expected healthcheck suggestion for database service")
	}
}

func TestValidateComposeContent_VolumeUsage(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web:
    image: nginx
    volumes:
      - used_volume:/data
volumes:
  used_volume:
  unused_volume:
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should warn about unused volume
	hasUnusedVolumeWarning := false
	for _, warning := range result.Warnings {
		if warning.Type == "unused_volume" {
			hasUnusedVolumeWarning = true
			break
		}
	}
	
	if !hasUnusedVolumeWarning {
		t.Error("Expected unused volume warning")
	}
}

func TestValidateComposeContent_ServiceNames(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  valid-service:
    image: nginx
  invalid@service:
    image: nginx
  "another-invalid service":
    image: nginx
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should have service name errors
	invalidNameErrors := 0
	for _, err := range result.Errors {
		if err.Type == "invalid_service_name" {
			invalidNameErrors++
		}
	}
	
	if invalidNameErrors != 2 {
		t.Errorf("Expected 2 invalid service name errors, got %d", invalidNameErrors)
	}
}

func TestValidateComposeContent_NetworkValidation(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web:
    image: nginx
    networks:
      - valid_network
      - invalid-NETWORK
networks:
  valid_network:
  invalid-NETWORK:
    driver: unknown_driver
`
	
	result := validator.ValidateComposeContent(content)
	
	// Should have network warnings
	hasNetworkWarnings := false
	for _, warning := range result.Warnings {
		if warning.Type == "invalid_network_name" || warning.Type == "unknown_network_driver" {
			hasNetworkWarnings = true
			break
		}
	}
	
	if !hasNetworkWarnings {
		t.Error("Expected network validation warnings")
	}
}

func TestValidateComposeContent_Statistics(t *testing.T) {
	validator := NewComposeValidator([]string{})
	
	content := `
services:
  web:
    image: nginx
  api:
    image: node
volumes:
  data1:
  data2:
networks:
  frontend:
  backend:
`
	
	result := validator.ValidateComposeContent(content)
	
	if result.ServiceCount != 2 {
		t.Errorf("Expected 2 services, got %d", result.ServiceCount)
	}
	
	if result.VolumeCount != 2 {
		t.Errorf("Expected 2 volumes, got %d", result.VolumeCount)
	}
	
	if result.NetworkCount != 2 {
		t.Errorf("Expected 2 networks, got %d", result.NetworkCount)
	}
}

// Helper function to check if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(s) > len(substr) && (s[:len(substr)] == substr || 
		s[len(s)-len(substr):] == substr || 
		indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}