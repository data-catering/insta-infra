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
		requiresPassword   bool
		passwordCredential bool
	}{
		{
			serviceName:        "postgres",
			expectedCmd:        "PGPASSWORD=${POSTGRES_PASSWORD:-postgres} psql -U${POSTGRES_USER:-postgres}",
			requiresPassword:   true,
			passwordCredential: true,
		},
		{
			serviceName:        "mysql",
			expectedCmd:        "mysql -u ${MYSQL_USER:-root} -p${MYSQL_PASSWORD:-root}",
			requiresPassword:   true,
			passwordCredential: true,
		},
		{
			serviceName:        "cassandra",
			expectedCmd:        "cqlsh",
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
