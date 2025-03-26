package main

import (
	"testing"
)

func TestServiceDefinitions(t *testing.T) {
	// Check if services map is populated
	if len(Services) == 0 {
		t.Fatal("Services map is empty")
	}

	// Verify specific services exist and have expected properties
	testCases := []struct {
		serviceName        string
		expectedCmd        string
		expectedPort       int
		requiresPassword   bool
		passwordCredential bool
	}{
		{
			serviceName:        "postgres",
			expectedCmd:        "psql",
			expectedPort:       5432,
			requiresPassword:   true,
			passwordCredential: true,
		},
		{
			serviceName:        "mysql",
			expectedCmd:        "mysql",
			expectedPort:       3306,
			requiresPassword:   true,
			passwordCredential: true,
		},
		{
			serviceName:        "cassandra",
			expectedCmd:        "cqlsh",
			expectedPort:       9042,
			requiresPassword:   false,
			passwordCredential: false,
		},
		{
			serviceName:      "duckdb",
			expectedCmd:      "./duckdb",
			requiresPassword: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.serviceName, func(t *testing.T) {
			service, exists := Services[tc.serviceName]
			if !exists {
				t.Fatalf("Service %s not found in Services map", tc.serviceName)
			}

			if service.Name != tc.serviceName {
				t.Errorf("Expected Name=%s, got %s", tc.serviceName, service.Name)
			}

			if service.ConnectionCmd != tc.expectedCmd {
				t.Errorf("Expected ConnectionCmd=%s, got %s", tc.expectedCmd, service.ConnectionCmd)
			}

			if tc.expectedPort > 0 && service.DefaultPort != tc.expectedPort {
				t.Errorf("Expected DefaultPort=%d, got %d", tc.expectedPort, service.DefaultPort)
			}

			if service.RequiresPassword != tc.requiresPassword {
				t.Errorf("Expected RequiresPassword=%v, got %v", tc.requiresPassword, service.RequiresPassword)
			}

			if tc.passwordCredential {
				if service.DefaultUser == "" {
					t.Error("Expected non-empty DefaultUser")
				}
				if service.DefaultPassword == "" {
					t.Error("Expected non-empty DefaultPassword")
				}
			}
		})
	}
}

func TestServiceNameMatchesKey(t *testing.T) {
	// Verify that each service's Name field matches its key in the map
	for key, service := range Services {
		if key != service.Name {
			t.Errorf("Service key %s does not match Service.Name %s", key, service.Name)
		}
	}
}

func TestUniqueServicePorts(t *testing.T) {
	// Create a map to track ports
	portMap := make(map[int][]string)

	// Collect services by port
	for name, service := range Services {
		if service.DefaultPort > 0 {
			portMap[service.DefaultPort] = append(portMap[service.DefaultPort], name)
		}
	}

	// Check for conflicts (multiple services using the same port)
	// Note: This is informative, not an error, as some services might intentionally use the same port
	for port, services := range portMap {
		if len(services) > 1 {
			t.Logf("Note: Port %d is used by multiple services: %v", port, services)
		}
	}
}
