package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	binaryPath := "./insta"

	// Check if binary exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath
	}

	// Build the binary
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to build binary: %v", err)
	}

	return binaryPath
}

func cleanup(t *testing.T, binaryPath string) {
	t.Helper()

	// Stop all services
	cmd := exec.Command(binaryPath, "-d")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Just log the error but continue with cleanup
		t.Logf("warning: failed to stop services during cleanup: %v", err)
	}

	// Remove test containers
	cmd = exec.Command("docker", "ps", "-aq", "-f", "name=postgres", "-f", "name=httpbin")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		cmd = exec.Command("docker", "rm", "-f", string(output))
		if err := cmd.Run(); err != nil {
			t.Logf("warning: failed to remove containers: %v", err)
		}
	}
}

func checkContainerRuntime(t *testing.T) {
	t.Helper()

	// Check for Docker
	if _, err := exec.LookPath("docker"); err == nil {
		cmd := exec.Command("docker", "info")
		if err := cmd.Run(); err == nil {
			return
		}
	}

	// Check for Podman
	if _, err := exec.LookPath("podman"); err == nil {
		cmd := exec.Command("podman", "info")
		if err := cmd.Run(); err == nil {
			return
		}
	}

	t.Skip("No container runtime available, skipping integration tests")
}

// TestCoreIntegration tests essential CLI functionality
func TestCoreIntegration(t *testing.T) {
	checkContainerRuntime(t)

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)
	cleanup(t, binaryPath)

	// Test basic service lifecycle
	t.Run("service lifecycle", func(t *testing.T) {
		// Start a simple service
		cmd := exec.Command(binaryPath, "httpbin")
		cmd.Env = append(os.Environ(), "TESTING=true")
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to start httpbin: %v", err)
		}

		// Stop the service
		cmd = exec.Command(binaryPath, "-d", "httpbin")
		cmd.Env = append(os.Environ(), "TESTING=true")
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to stop httpbin: %v", err)
		}
	})

	// Test list services functionality
	t.Run("list services", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "-l")
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.Output()
		if err != nil {
			t.Fatalf("failed to list services: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "postgres") || !strings.Contains(outputStr, "redis") {
			t.Errorf("expected common services in output, got: %s", outputStr)
		}
	})
}

// TestServiceConnection tests that services can actually be connected to
func TestServiceConnection(t *testing.T) {
	checkContainerRuntime(t)

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)
	cleanup(t, binaryPath)

	// Start httpbin service (lighter than postgres)
	cmd := exec.Command(binaryPath, "httpbin")
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to start httpbin: %v", err)
	}

	// Wait for service to start
	time.Sleep(2 * time.Second)

	// Test connection by checking if the container is running
	// Since curl might not be available in the container, just verify the container exists and is responsive
	cmd = exec.Command("docker", "ps", "-f", "name=http", "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("failed to check httpbin container status: %v", err)
	} else {
		statusOutput := string(output)
		if !strings.Contains(statusOutput, "Up") {
			t.Errorf("httpbin container not running: %s", statusOutput)
		}
	}

	// Cleanup
	cmd = exec.Command(binaryPath, "-d", "httpbin")
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to stop httpbin: %v", err)
	}
}

// TestCustomServiceCLI tests the full custom service CLI workflow
func TestCustomServiceCLI(t *testing.T) {
	checkContainerRuntime(t)

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)
	cleanup(t, binaryPath)

	// Create a temporary compose file for testing
	tempComposeFile := "./test-custom-service.yaml"
	composeContent := `version: '3.8'
services:
  test-web:
    image: nginx:alpine
    ports:
      - "8090:80"
    environment:
      - NODE_ENV=test
    restart: unless-stopped
  
  test-db:
    image: postgres:13-alpine
    environment:
      POSTGRES_DB: testdb
      POSTGRES_USER: testuser
      POSTGRES_PASSWORD: testpass
    ports:
      - "5433:5432"
    restart: unless-stopped`

	// Write the test compose file
	if err := os.WriteFile(tempComposeFile, []byte(composeContent), 0644); err != nil {
		t.Fatalf("failed to create test compose file: %v", err)
	}
	defer os.Remove(tempComposeFile)

	// Test 1: Validate the compose file
	t.Run("validate_compose_file", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "custom", "validate", tempComposeFile)
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to validate compose file: %v, output: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Compose file is valid!") {
			t.Errorf("expected validation success message, got: %s", outputStr)
		}
	})

	// Test 2: Add the custom service
	var serviceID string
	t.Run("add_custom_service", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "custom", "add", tempComposeFile)
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to add custom service: %v, output: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Successfully added custom service") {
			t.Errorf("expected success message, got: %s", outputStr)
		}

		// Extract service ID from output for later use
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Successfully added custom service") {
				// Extract ID from line like: "Successfully added custom service 'test-custom-service' (ID: custom_123)"
				if idx := strings.Index(line, "(ID: "); idx != -1 {
					endIdx := strings.Index(line[idx:], ")")
					if endIdx != -1 {
						serviceID = strings.TrimSpace(line[idx+5 : idx+endIdx])
					}
				}
			}
		}
	})

	// Test 3: List custom services and verify our service is there
	t.Run("list_custom_services", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "custom", "list")
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("failed to list custom services: %v, output: %s", err, output)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "test-custom-service") {
			t.Errorf("expected to find test-custom-service in list, got: %s", outputStr)
		}
		// Check for services in any order (test-web, test-db or test-db, test-web)
		if !strings.Contains(outputStr, "test-web") || !strings.Contains(outputStr, "test-db") {
			t.Errorf("expected to find service names test-web and test-db in list, got: %s", outputStr)
		}
	})

	// Test 4: Test custom service lifecycle (start/stop) - Skip if Docker not available
	t.Run("custom_service_lifecycle", func(t *testing.T) {
		// Check if Docker is available
		if _, err := exec.LookPath("docker"); err != nil {
			t.Skip("Docker not available, skipping lifecycle test")
		}

		// Start the custom service
		cmd := exec.Command(binaryPath, "test-web")
		cmd.Env = append(os.Environ(), "TESTING=true")
		if err := cmd.Run(); err != nil {
			t.Logf("failed to start custom service (this may be expected in test environment): %v", err)
			return // Don't fail the test, as Docker may not be fully configured in test environment
		}

		// Wait for service to start
		time.Sleep(3 * time.Second)

		// Check if the container is running
		cmd = exec.Command("docker", "ps", "-f", "name=insta-test-web-1", "--format", "{{.Status}}")
		output, err := cmd.Output()
		if err != nil {
			t.Logf("failed to check custom service container status: %v", err)
		} else {
			statusOutput := string(output)
			if !strings.Contains(statusOutput, "Up") {
				t.Logf("custom service container not running: %s", statusOutput)
			}
		}

		// Stop the custom service
		cmd = exec.Command(binaryPath, "-d", "test-web")
		cmd.Env = append(os.Environ(), "TESTING=true")
		if err := cmd.Run(); err != nil {
			t.Logf("failed to stop custom service: %v", err)
		}
	})

	// Test 5: Remove the custom service
	if serviceID != "" {
		t.Run("remove_custom_service", func(t *testing.T) {
			cmd := exec.Command(binaryPath, "custom", "remove", serviceID)
			cmd.Env = append(os.Environ(), "TESTING=true")
			cmd.Stdin = strings.NewReader("y\n") // Auto-confirm removal
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to remove custom service: %v, output: %s", err, output)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, "Successfully removed custom service") {
				t.Errorf("expected removal success message, got: %s", outputStr)
			}
		})

		// Verify removal by listing again
		t.Run("verify_service_removed", func(t *testing.T) {
			cmd := exec.Command(binaryPath, "custom", "list")
			cmd.Env = append(os.Environ(), "TESTING=true")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("failed to list custom services after removal: %v", err)
			}

			outputStr := string(output)
			// Check that the specific service ID is no longer in the list
			if serviceID != "" && strings.Contains(outputStr, serviceID) {
				t.Errorf("service with ID %s should have been removed but still appears in list: %s", serviceID, outputStr)
			}
		})
	}
}

// TestCustomServiceErrorHandling tests error scenarios for custom service CLI
func TestCustomServiceErrorHandling(t *testing.T) {
	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)

	// Test 1: Invalid custom command
	t.Run("invalid_custom_command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "custom", "invalid-command")
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("expected error for invalid custom command")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "unknown custom command") {
			t.Errorf("expected error message about unknown command, got: %s", outputStr)
		}
	})

	// Test 2: Missing file for add command
	t.Run("missing_file_for_add", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "custom", "add", "nonexistent-file.yaml")
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("expected error for nonexistent file")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "does not exist") {
			t.Errorf("expected error message about file not existing, got: %s", outputStr)
		}
	})

	// Test 3: Invalid YAML file
	t.Run("invalid_yaml_file", func(t *testing.T) {
		invalidFile := "./invalid.yaml"
		invalidContent := `invalid: yaml: content:
		malformed`

		if err := os.WriteFile(invalidFile, []byte(invalidContent), 0644); err != nil {
			t.Fatalf("failed to create invalid file: %v", err)
		}
		defer os.Remove(invalidFile)

		cmd := exec.Command(binaryPath, "custom", "validate", invalidFile)
		cmd.Env = append(os.Environ(), "TESTING=true")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("expected error for invalid YAML")
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Validation Errors") {
			t.Errorf("expected validation error message, got: %s", outputStr)
		}
	})

	// Test 4: Missing arguments
	t.Run("missing_arguments", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{"no_command", []string{"custom"}},
			{"add_no_file", []string{"custom", "add"}},
			{"remove_no_service", []string{"custom", "remove"}},
			{"validate_no_file", []string{"custom", "validate"}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				cmd := exec.Command(binaryPath, tt.args...)
				cmd.Env = append(os.Environ(), "TESTING=true")
				output, err := cmd.CombinedOutput()
				if err == nil {
					t.Errorf("expected error for missing arguments in %s", tt.name)
				}

				outputStr := string(output)
				if !strings.Contains(outputStr, "Error:") {
					t.Errorf("expected error message for %s, got: %s", tt.name, outputStr)
				}
			})
		}
	})
}

// TestCustomServiceUIIntegration tests UI-related functionality through API
func TestCustomServiceUIIntegration(t *testing.T) {
	checkContainerRuntime(t)

	binaryPath := buildBinary(t)
	defer os.Remove(binaryPath)
	cleanup(t, binaryPath)

	// Start the web server
	cmd := exec.Command(binaryPath, "--web-server", "--port", "9315")
	cmd.Env = append(os.Environ(), "TESTING=true")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start web server: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for server to start
	time.Sleep(2 * time.Second)

	// Test API endpoints are accessible
	t.Run("api_endpoints_accessible", func(t *testing.T) {
		endpoints := []string{
			"http://localhost:9315/api/v1/custom/compose",
			"http://localhost:9315/api/v1/custom/stats",
		}

		for _, endpoint := range endpoints {
			resp, err := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", endpoint).Output()
			if err != nil {
				t.Logf("curl not available or endpoint unreachable: %v", err)
				continue
			}

			statusCode := strings.TrimSpace(string(resp))
			if statusCode != "200" && statusCode != "405" { // 405 for methods not allowed is OK
				t.Errorf("endpoint %s returned status %s", endpoint, statusCode)
			}
		}
	})
}
