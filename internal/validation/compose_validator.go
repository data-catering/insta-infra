package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// ValidationResult represents the outcome of compose file validation
type ValidationResult struct {
	Valid         bool                    `json:"valid"`
	Errors        []ValidationError       `json:"errors,omitempty"`
	Warnings      []ValidationWarning     `json:"warnings,omitempty"`
	Suggestions   []ValidationSuggestion  `json:"suggestions,omitempty"`
	ServiceCount  int                     `json:"service_count"`
	VolumeCount   int                     `json:"volume_count"`
	NetworkCount  int                     `json:"network_count"`
}

// ValidationError represents a validation error with context
type ValidationError struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name,omitempty"`
	Field       string `json:"field,omitempty"`
	Line        int    `json:"line,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name,omitempty"`
	Field       string `json:"field,omitempty"`
	Suggestion  string `json:"suggestion,omitempty"`
}

// ValidationSuggestion represents an improvement suggestion
type ValidationSuggestion struct {
	Type        string `json:"type"`
	Message     string `json:"message"`
	ServiceName string `json:"service_name,omitempty"`
	Example     string `json:"example,omitempty"`
}

// ComposeFile represents the structure of a docker-compose file
type ComposeFile struct {
	Version  string                 `yaml:"version,omitempty"`
	Services map[string]interface{} `yaml:"services"`
	Volumes  map[string]interface{} `yaml:"volumes,omitempty"`
	Networks map[string]interface{} `yaml:"networks,omitempty"`
	Secrets  map[string]interface{} `yaml:"secrets,omitempty"`
	Configs  map[string]interface{} `yaml:"configs,omitempty"`
}

// ComposeValidator provides comprehensive Docker Compose file validation
type ComposeValidator struct {
	// Known service names from insta-infra for dependency validation
	KnownServices map[string]bool
}

// NewComposeValidator creates a new compose file validator
func NewComposeValidator(knownServices []string) *ComposeValidator {
	serviceMap := make(map[string]bool)
	for _, service := range knownServices {
		serviceMap[service] = true
	}
	
	// Add common database and infrastructure services that users might reference
	commonServices := []string{
		"postgres", "mysql", "redis", "mongodb", "elasticsearch", "kibana",
		"grafana", "prometheus", "jaeger", "zipkin", "consul", "etcd",
		"rabbitmq", "kafka", "zookeeper", "cassandra", "influxdb",
		"memcached", "nginx", "apache", "haproxy", "traefik",
	}
	
	for _, service := range commonServices {
		serviceMap[service] = true
	}
	
	return &ComposeValidator{
		KnownServices: serviceMap,
	}
}

// ValidateComposeContent performs comprehensive validation of compose file content
func (v *ComposeValidator) ValidateComposeContent(content string) *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []ValidationWarning{},
		Suggestions: []ValidationSuggestion{},
	}

	// Parse YAML
	var compose ComposeFile
	if err := yaml.Unmarshal([]byte(content), &compose); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Type:       "yaml_syntax",
			Message:    fmt.Sprintf("Invalid YAML syntax: %v", err),
			Suggestion: "Check for proper indentation, quotes, and YAML formatting",
		})
		return result
	}

	// Basic structure validation
	if err := v.validateBasicStructure(&compose, result); err != nil {
		result.Valid = false
		return result
	}

	// Set counts
	result.ServiceCount = len(compose.Services)
	if compose.Volumes != nil {
		result.VolumeCount = len(compose.Volumes)
	}
	if compose.Networks != nil {
		result.NetworkCount = len(compose.Networks)
	}

	// Validate each service
	for serviceName, serviceConfig := range compose.Services {
		v.validateService(serviceName, serviceConfig, &compose, result)
	}

	// Validate volumes
	if compose.Volumes != nil {
		v.validateVolumes(compose.Volumes, result)
	}

	// Validate networks
	if compose.Networks != nil {
		v.validateNetworks(compose.Networks, result)
	}

	// Cross-service validation
	v.validateServiceDependencies(&compose, result)
	v.validatePortConflicts(&compose, result)
	v.validateVolumeUsage(&compose, result)

	// Best practices and suggestions
	v.addBestPracticeSuggestions(&compose, result)

	return result
}

// validateBasicStructure ensures the compose file has required structure
func (v *ComposeValidator) validateBasicStructure(compose *ComposeFile, result *ValidationResult) error {
	if compose.Services == nil {
		result.Errors = append(result.Errors, ValidationError{
			Type:       "missing_services",
			Message:    "Compose file must contain a 'services' section",
			Suggestion: "Add a 'services:' section with at least one service definition",
		})
		return fmt.Errorf("missing services section")
	}

	if len(compose.Services) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Type:       "empty_services",
			Message:    "Services section cannot be empty",
			Suggestion: "Define at least one service in the services section",
		})
		return fmt.Errorf("empty services section")
	}

	// Check version (if specified)
	if compose.Version != "" {
		if !v.isValidComposeVersion(compose.Version) {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "version_deprecated",
				Message:    fmt.Sprintf("Compose version '%s' may be deprecated", compose.Version),
				Suggestion: "Consider using version '3.8' or removing version field (modern Docker Compose doesn't require it)",
			})
		}
	}

	return nil
}

// validateService validates an individual service configuration
func (v *ComposeValidator) validateService(serviceName string, serviceConfig interface{}, compose *ComposeFile, result *ValidationResult) {
	serviceMap, ok := serviceConfig.(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_service",
			Message:     fmt.Sprintf("Service '%s' must be a map/object", serviceName),
			ServiceName: serviceName,
			Suggestion:  "Ensure the service definition uses proper YAML map syntax",
		})
		return
	}

	// Validate service name
	if !v.isValidServiceName(serviceName) {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_service_name",
			Message:     fmt.Sprintf("Service name '%s' contains invalid characters", serviceName),
			ServiceName: serviceName,
			Suggestion:  "Service names should only contain letters, numbers, underscores, and hyphens",
		})
	}

	// Check for required image or build
	if !v.hasImageOrBuild(serviceMap) {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "missing_image_or_build",
			Message:     fmt.Sprintf("Service '%s' must specify either 'image' or 'build'", serviceName),
			ServiceName: serviceName,
			Suggestion:  "Add an 'image: <image-name>' or 'build: <build-context>' field",
		})
	}

	// Validate specific fields
	v.validateServicePorts(serviceName, serviceMap, result)
	v.validateServiceVolumes(serviceName, serviceMap, result)
	v.validateServiceEnvironment(serviceName, serviceMap, result)
	v.validateServiceNetworks(serviceName, serviceMap, result)
	v.validateServiceDependsOn(serviceName, serviceMap, compose, result)
	v.validateServiceHealthcheck(serviceName, serviceMap, result)
	v.validateServiceRestart(serviceName, serviceMap, result)
}

// validateServicePorts validates port configurations
func (v *ComposeValidator) validateServicePorts(serviceName string, serviceMap map[string]interface{}, result *ValidationResult) {
	ports, exists := serviceMap["ports"]
	if !exists {
		return
	}

	portsList, ok := ports.([]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_ports",
			Message:     fmt.Sprintf("Service '%s' ports must be a list", serviceName),
			ServiceName: serviceName,
			Field:       "ports",
			Suggestion:  "Use YAML list syntax: ports: ['8080:80', '3000:3000']",
		})
		return
	}

	for i, port := range portsList {
		if err := v.validatePortMapping(fmt.Sprintf("%v", port)); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Type:        "invalid_port_mapping",
				Message:     fmt.Sprintf("Service '%s' has invalid port mapping at index %d: %v", serviceName, i, err),
				ServiceName: serviceName,
				Field:       "ports",
				Suggestion:  "Use format 'HOST_PORT:CONTAINER_PORT' (e.g., '8080:80')",
			})
		}
	}
}

// validateServiceVolumes validates volume mounts
func (v *ComposeValidator) validateServiceVolumes(serviceName string, serviceMap map[string]interface{}, result *ValidationResult) {
	volumes, exists := serviceMap["volumes"]
	if !exists {
		return
	}

	volumesList, ok := volumes.([]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_volumes",
			Message:     fmt.Sprintf("Service '%s' volumes must be a list", serviceName),
			ServiceName: serviceName,
			Field:       "volumes",
			Suggestion:  "Use YAML list syntax: volumes: ['./data:/data', 'named_volume:/app']",
		})
		return
	}

	for i, volume := range volumesList {
		volumeStr := fmt.Sprintf("%v", volume)
		if err := v.validateVolumeMount(volumeStr); err != nil {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:        "volume_mount_warning",
				Message:     fmt.Sprintf("Service '%s' volume mount at index %d: %v", serviceName, i, err),
				ServiceName: serviceName,
				Suggestion:  "Use format 'SOURCE:TARGET' or 'SOURCE:TARGET:MODE'",
			})
		}
	}
}

// validateServiceEnvironment validates environment variables
func (v *ComposeValidator) validateServiceEnvironment(serviceName string, serviceMap map[string]interface{}, result *ValidationResult) {
	env, exists := serviceMap["environment"]
	if !exists {
		return
	}

	switch envType := env.(type) {
	case []interface{}:
		// List format: ["KEY=value", "KEY2=value2"]
		for i, envVar := range envType {
			envStr := fmt.Sprintf("%v", envVar)
			if !strings.Contains(envStr, "=") {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:        "env_var_format",
					Message:     fmt.Sprintf("Service '%s' environment variable at index %d may be missing value", serviceName, i),
					ServiceName: serviceName,
					Suggestion:  "Use format 'KEY=value' or ensure the variable is defined elsewhere",
				})
			}
		}
	case map[string]interface{}:
		// Map format: {KEY: value, KEY2: value2}
		for key, value := range envType {
			if value == nil {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:        "env_var_null",
					Message:     fmt.Sprintf("Service '%s' environment variable '%s' has null value", serviceName, key),
					ServiceName: serviceName,
					Suggestion:  "Provide a value or use ${VAR} for external variables",
				})
			}
		}
	default:
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_environment",
			Message:     fmt.Sprintf("Service '%s' environment must be a list or map", serviceName),
			ServiceName: serviceName,
			Field:       "environment",
			Suggestion:  "Use list format ['KEY=value'] or map format {KEY: value}",
		})
	}
}

// validateServiceNetworks validates network configuration
func (v *ComposeValidator) validateServiceNetworks(serviceName string, serviceMap map[string]interface{}, result *ValidationResult) {
	networks, exists := serviceMap["networks"]
	if !exists {
		return
	}

	switch networksType := networks.(type) {
	case []interface{}:
		// List format
		for _, network := range networksType {
			networkName := fmt.Sprintf("%v", network)
			if !v.isValidNetworkName(networkName) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:        "invalid_network_name",
					Message:     fmt.Sprintf("Service '%s' references network '%s' with invalid characters", serviceName, networkName),
					ServiceName: serviceName,
					Suggestion:  "Network names should be lowercase with no special characters",
				})
			}
		}
	case map[string]interface{}:
		// Map format with aliases
		for networkName := range networksType {
			if !v.isValidNetworkName(networkName) {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:        "invalid_network_name",
					Message:     fmt.Sprintf("Service '%s' references network '%s' with invalid characters", serviceName, networkName),
					ServiceName: serviceName,
					Suggestion:  "Network names should be lowercase with no special characters",
				})
			}
		}
	}
}

// validateServiceDependsOn validates service dependencies
func (v *ComposeValidator) validateServiceDependsOn(serviceName string, serviceMap map[string]interface{}, compose *ComposeFile, result *ValidationResult) {
	dependsOn, exists := serviceMap["depends_on"]
	if !exists {
		return
	}

	var dependencies []string
	switch depType := dependsOn.(type) {
	case []interface{}:
		for _, dep := range depType {
			dependencies = append(dependencies, fmt.Sprintf("%v", dep))
		}
	case map[string]interface{}:
		for dep := range depType {
			dependencies = append(dependencies, dep)
		}
	default:
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_depends_on",
			Message:     fmt.Sprintf("Service '%s' depends_on must be a list or map", serviceName),
			ServiceName: serviceName,
			Field:       "depends_on",
			Suggestion:  "Use list format ['service1', 'service2'] or map format with conditions",
		})
		return
	}

	// Check if dependencies exist
	for _, dep := range dependencies {
		if _, exists := compose.Services[dep]; !exists {
			// Check if it's a known external service
			if v.KnownServices[dep] {
				result.Suggestions = append(result.Suggestions, ValidationSuggestion{
					Type:        "external_dependency",
					Message:     fmt.Sprintf("Service '%s' depends on '%s' which appears to be an external insta-infra service", serviceName, dep),
					ServiceName: serviceName,
					Example:     "This dependency will be resolved by insta-infra's built-in services",
				})
			} else {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Type:        "missing_dependency",
					Message:     fmt.Sprintf("Service '%s' depends on '%s' which is not defined in this compose file", serviceName, dep),
					ServiceName: serviceName,
					Suggestion:  "Define the service in this file or ensure it's available in insta-infra",
				})
			}
		}
	}
}

// validateServiceHealthcheck validates healthcheck configuration
func (v *ComposeValidator) validateServiceHealthcheck(serviceName string, serviceMap map[string]interface{}, result *ValidationResult) {
	healthcheck, exists := serviceMap["healthcheck"]
	if !exists {
		// Suggest healthcheck for database services
		if v.isDatabaseService(serviceMap) {
			result.Suggestions = append(result.Suggestions, ValidationSuggestion{
				Type:        "add_healthcheck",
				Message:     fmt.Sprintf("Consider adding a healthcheck to service '%s'", serviceName),
				ServiceName: serviceName,
				Example:     "healthcheck:\n  test: ['CMD', 'curl', '-f', 'http://localhost:8080/health']\n  interval: 30s\n  timeout: 10s\n  retries: 3",
			})
		}
		return
	}

	healthMap, ok := healthcheck.(map[string]interface{})
	if !ok {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_healthcheck",
			Message:     fmt.Sprintf("Service '%s' healthcheck must be a map", serviceName),
			ServiceName: serviceName,
			Field:       "healthcheck",
			Suggestion:  "Use healthcheck with test, interval, timeout, retries fields",
		})
		return
	}

	// Check for required test field
	if _, hasTest := healthMap["test"]; !hasTest {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "missing_healthcheck_test",
			Message:     fmt.Sprintf("Service '%s' healthcheck must specify a test", serviceName),
			ServiceName: serviceName,
			Field:       "healthcheck.test",
			Suggestion:  "Add a test field like: test: ['CMD', 'curl', '-f', 'http://localhost/health']",
		})
	}
}

// validateServiceRestart validates restart policy
func (v *ComposeValidator) validateServiceRestart(serviceName string, serviceMap map[string]interface{}, result *ValidationResult) {
	restart, exists := serviceMap["restart"]
	if !exists {
		result.Suggestions = append(result.Suggestions, ValidationSuggestion{
			Type:        "add_restart_policy",
			Message:     fmt.Sprintf("Consider adding a restart policy to service '%s'", serviceName),
			ServiceName: serviceName,
			Example:     "restart: unless-stopped",
		})
		return
	}

	restartStr := fmt.Sprintf("%v", restart)
	validPolicies := []string{"no", "always", "on-failure", "unless-stopped"}
	
	valid := false
	for _, policy := range validPolicies {
		if restartStr == policy {
			valid = true
			break
		}
	}

	if !valid {
		result.Errors = append(result.Errors, ValidationError{
			Type:        "invalid_restart_policy",
			Message:     fmt.Sprintf("Service '%s' has invalid restart policy '%s'", serviceName, restartStr),
			ServiceName: serviceName,
			Field:       "restart",
			Suggestion:  "Use one of: no, always, on-failure, unless-stopped",
		})
	}
}

// validateVolumes validates top-level volumes
func (v *ComposeValidator) validateVolumes(volumes map[string]interface{}, result *ValidationResult) {
	for volumeName, volumeConfig := range volumes {
		if !v.isValidVolumeName(volumeName) {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "invalid_volume_name",
				Message:    fmt.Sprintf("Volume name '%s' contains invalid characters", volumeName),
				Suggestion: "Volume names should be lowercase with underscores or hyphens",
			})
		}

		// Check if volume config is valid
		if volumeConfig != nil {
			if volumeMap, ok := volumeConfig.(map[string]interface{}); ok {
				if driver, hasDriver := volumeMap["driver"]; hasDriver {
					driverStr := fmt.Sprintf("%v", driver)
					if driverStr != "local" && driverStr != "" {
						result.Suggestions = append(result.Suggestions, ValidationSuggestion{
							Type:    "volume_driver",
							Message: fmt.Sprintf("Volume '%s' uses driver '%s'", volumeName, driverStr),
							Example: "Ensure the driver is available in your environment",
						})
					}
				}
			}
		}
	}
}

// validateNetworks validates top-level networks
func (v *ComposeValidator) validateNetworks(networks map[string]interface{}, result *ValidationResult) {
	for networkName, networkConfig := range networks {
		if !v.isValidNetworkName(networkName) {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "invalid_network_name",
				Message:    fmt.Sprintf("Network name '%s' contains invalid characters", networkName),
				Suggestion: "Network names should be lowercase with no special characters",
			})
		}

		// Check network configuration
		if networkConfig != nil {
			if networkMap, ok := networkConfig.(map[string]interface{}); ok {
				if driver, hasDriver := networkMap["driver"]; hasDriver {
					driverStr := fmt.Sprintf("%v", driver)
					if driverStr != "bridge" && driverStr != "host" && driverStr != "overlay" && driverStr != "macvlan" && driverStr != "none" {
						result.Warnings = append(result.Warnings, ValidationWarning{
							Type:       "unknown_network_driver",
							Message:    fmt.Sprintf("Network '%s' uses unknown driver '%s'", networkName, driverStr),
							Suggestion: "Common drivers are: bridge, host, overlay, macvlan, none",
						})
					}
				}
			}
		}
	}
}

// validateServiceDependencies checks for circular dependencies
func (v *ComposeValidator) validateServiceDependencies(compose *ComposeFile, result *ValidationResult) {
	// Build dependency graph
	deps := make(map[string][]string)
	for serviceName, serviceConfig := range compose.Services {
		serviceMap, ok := serviceConfig.(map[string]interface{})
		if !ok {
			continue
		}

		dependsOn, exists := serviceMap["depends_on"]
		if !exists {
			continue
		}

		var dependencies []string
		switch depType := dependsOn.(type) {
		case []interface{}:
			for _, dep := range depType {
				dependencies = append(dependencies, fmt.Sprintf("%v", dep))
			}
		case map[string]interface{}:
			for dep := range depType {
				dependencies = append(dependencies, dep)
			}
		}

		deps[serviceName] = dependencies
	}

	// Check for circular dependencies using DFS
	visited := make(map[string]bool)
	inStack := make(map[string]bool)

	var dfs func(string) bool
	dfs = func(service string) bool {
		visited[service] = true
		inStack[service] = true

		for _, dep := range deps[service] {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if inStack[dep] {
				result.Errors = append(result.Errors, ValidationError{
					Type:        "circular_dependency",
					Message:     fmt.Sprintf("Circular dependency detected involving services '%s' and '%s'", service, dep),
					ServiceName: service,
					Suggestion:  "Remove circular dependencies by restructuring service relationships",
				})
				return true
			}
		}

		inStack[service] = false
		return false
	}

	for service := range deps {
		if !visited[service] {
			dfs(service)
		}
	}
}

// validatePortConflicts checks for port conflicts between services
func (v *ComposeValidator) validatePortConflicts(compose *ComposeFile, result *ValidationResult) {
	usedPorts := make(map[string]string) // port -> service name

	for serviceName, serviceConfig := range compose.Services {
		serviceMap, ok := serviceConfig.(map[string]interface{})
		if !ok {
			continue
		}

		ports, exists := serviceMap["ports"]
		if !exists {
			continue
		}

		portsList, ok := ports.([]interface{})
		if !ok {
			continue
		}

		for _, port := range portsList {
			portStr := fmt.Sprintf("%v", port)
			hostPort := v.extractHostPort(portStr)
			if hostPort != "" {
				if existingService, exists := usedPorts[hostPort]; exists {
					result.Errors = append(result.Errors, ValidationError{
						Type:        "port_conflict",
						Message:     fmt.Sprintf("Port %s is used by both services '%s' and '%s'", hostPort, existingService, serviceName),
						ServiceName: serviceName,
						Field:       "ports",
						Suggestion:  "Use different host ports for each service",
					})
				} else {
					usedPorts[hostPort] = serviceName
				}
			}
		}
	}
}

// validateVolumeUsage checks if defined volumes are actually used
func (v *ComposeValidator) validateVolumeUsage(compose *ComposeFile, result *ValidationResult) {
	if compose.Volumes == nil {
		return
	}

	usedVolumes := make(map[string]bool)

	// Find all volume references in services
	for _, serviceConfig := range compose.Services {
		serviceMap, ok := serviceConfig.(map[string]interface{})
		if !ok {
			continue
		}

		volumes, exists := serviceMap["volumes"]
		if !exists {
			continue
		}

		volumesList, ok := volumes.([]interface{})
		if !ok {
			continue
		}

		for _, volume := range volumesList {
			volumeStr := fmt.Sprintf("%v", volume)
			if volumeName := v.extractVolumeName(volumeStr); volumeName != "" {
				usedVolumes[volumeName] = true
			}
		}
	}

	// Check for unused volumes
	for volumeName := range compose.Volumes {
		if !usedVolumes[volumeName] {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Type:       "unused_volume",
				Message:    fmt.Sprintf("Volume '%s' is defined but not used by any service", volumeName),
				Suggestion: "Remove unused volume definitions or add volume mounts to services",
			})
		}
	}
}

// addBestPracticeSuggestions adds suggestions for best practices
func (v *ComposeValidator) addBestPracticeSuggestions(compose *ComposeFile, result *ValidationResult) {
	serviceCount := len(compose.Services)

	// Suggest using named volumes for data persistence
	hasNamedVolumes := compose.Volumes != nil && len(compose.Volumes) > 0
	hasDataServices := false
	
	for _, serviceConfig := range compose.Services {
		if v.isDatabaseService(serviceConfig) {
			hasDataServices = true
			break
		}
	}

	if hasDataServices && !hasNamedVolumes {
		result.Suggestions = append(result.Suggestions, ValidationSuggestion{
			Type:    "use_named_volumes",
			Message: "Consider using named volumes for database data persistence",
			Example: "volumes:\n  db_data:\nservices:\n  db:\n    volumes:\n      - db_data:/var/lib/postgresql/data",
		})
	}

	// Suggest using environment files for many environment variables
	totalEnvVars := 0
	for _, serviceConfig := range compose.Services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			if env, exists := serviceMap["environment"]; exists {
				switch envType := env.(type) {
				case []interface{}:
					totalEnvVars += len(envType)
				case map[string]interface{}:
					totalEnvVars += len(envType)
				}
			}
		}
	}

	if totalEnvVars > 10 {
		result.Suggestions = append(result.Suggestions, ValidationSuggestion{
			Type:    "use_env_file",
			Message: "Consider using .env files for environment variables",
			Example: "env_file:\n  - .env\n  - ./config/.env.local",
		})
	}

	// Suggest resource limits for production
	hasResourceLimits := false
	for _, serviceConfig := range compose.Services {
		if serviceMap, ok := serviceConfig.(map[string]interface{}); ok {
			if _, exists := serviceMap["deploy"]; exists {
				hasResourceLimits = true
				break
			}
		}
	}

	if serviceCount > 1 && !hasResourceLimits {
		result.Suggestions = append(result.Suggestions, ValidationSuggestion{
			Type:    "add_resource_limits",
			Message: "Consider adding resource limits for production deployments",
			Example: "deploy:\n  resources:\n    limits:\n      memory: 512M\n      cpus: '0.5'",
		})
	}
}

// Helper methods

func (v *ComposeValidator) isValidComposeVersion(version string) bool {
	validVersions := []string{"3.0", "3.1", "3.2", "3.3", "3.4", "3.5", "3.6", "3.7", "3.8", "3.9"}
	for _, valid := range validVersions {
		if version == valid {
			return true
		}
	}
	return false
}

func (v *ComposeValidator) isValidServiceName(name string) bool {
	// Service names should follow DNS naming conventions
	match, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`, name)
	return match && len(name) <= 63
}

func (v *ComposeValidator) isValidVolumeName(name string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`, name)
	return match
}

func (v *ComposeValidator) isValidNetworkName(name string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`, name)
	return match
}

func (v *ComposeValidator) hasImageOrBuild(serviceMap map[string]interface{}) bool {
	_, hasImage := serviceMap["image"]
	_, hasBuild := serviceMap["build"]
	return hasImage || hasBuild
}

func (v *ComposeValidator) validatePortMapping(port string) error {
	// Handle various port formats: "8080:80", "127.0.0.1:8080:80", "3000", "3000/tcp"
	if strings.Contains(port, ":") {
		parts := strings.Split(port, ":")
		if len(parts) < 2 || len(parts) > 3 {
			return fmt.Errorf("invalid port format")
		}
		
		// Validate the last part (container port)
		containerPort := parts[len(parts)-1]
		if strings.Contains(containerPort, "/") {
			containerPort = strings.Split(containerPort, "/")[0]
		}
		
		if _, err := strconv.Atoi(containerPort); err != nil {
			return fmt.Errorf("invalid container port")
		}
		
		// If there's a host port, validate it
		if len(parts) >= 2 {
			hostPortPart := parts[len(parts)-2]
			if hostPortPart != "" {
				if _, err := strconv.Atoi(hostPortPart); err != nil {
					return fmt.Errorf("invalid host port")
				}
			}
		}
	} else {
		// Single port
		singlePort := port
		if strings.Contains(port, "/") {
			singlePort = strings.Split(port, "/")[0]
		}
		if _, err := strconv.Atoi(singlePort); err != nil {
			return fmt.Errorf("invalid port number")
		}
	}
	
	return nil
}

func (v *ComposeValidator) validateVolumeMount(volume string) error {
	if !strings.Contains(volume, ":") {
		return fmt.Errorf("volume mount should specify source and target")
	}
	
	parts := strings.Split(volume, ":")
	if len(parts) < 2 {
		return fmt.Errorf("volume mount format should be SOURCE:TARGET[:MODE]")
	}
	
	// Check if source looks like a URL (which might be invalid)
	source := parts[0]
	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		return fmt.Errorf("volume source cannot be a URL")
	}
	
	return nil
}

func (v *ComposeValidator) isDatabaseService(serviceConfig interface{}) bool {
	serviceMap, ok := serviceConfig.(map[string]interface{})
	if !ok {
		return false
	}
	
	image, exists := serviceMap["image"]
	if !exists {
		return false
	}
	
	imageStr := strings.ToLower(fmt.Sprintf("%v", image))
	dbImages := []string{"postgres", "mysql", "mongodb", "redis", "elasticsearch", "cassandra", "mariadb"}
	
	for _, db := range dbImages {
		if strings.Contains(imageStr, db) {
			return true
		}
	}
	
	return false
}

func (v *ComposeValidator) extractHostPort(portMapping string) string {
	if !strings.Contains(portMapping, ":") {
		return ""
	}
	
	parts := strings.Split(portMapping, ":")
	if len(parts) >= 2 {
		// If it's IP:HOST:CONTAINER, return HOST
		if len(parts) == 3 {
			return parts[1]
		}
		// If it's HOST:CONTAINER, return HOST
		return parts[0]
	}
	
	return ""
}

func (v *ComposeValidator) extractVolumeName(volumeMount string) string {
	if !strings.Contains(volumeMount, ":") {
		return ""
	}
	
	parts := strings.Split(volumeMount, ":")
	source := parts[0]
	
	// If source doesn't start with . or /, it might be a named volume
	if !strings.HasPrefix(source, ".") && !strings.HasPrefix(source, "/") && !strings.Contains(source, "\\") {
		return source
	}
	
	return ""
}