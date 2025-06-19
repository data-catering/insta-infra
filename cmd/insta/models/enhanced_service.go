package models

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/data-catering/insta-infra/v2/internal/core"
	"gopkg.in/yaml.v3"
)

/*
SERVICE-TO-CONTAINER MAPPING ARCHITECTURE

This file implements the core service-to-container mapping logic for insta-infra.
Understanding this architecture is critical for proper dependency resolution and UI display.

CORE PRINCIPLES:
1. Service Names -> Service Cards: Service names in internal/core/models.go define the service cards shown in the UI
2. Service Name -> Compose Service: Each service name maps to a Docker Compose service with the same name
3. Compose Service -> Container: Each Compose service has a container_name that may differ from the service name
4. Dependencies by Container Names: All dependency tracking uses container names, not service names
5. UI Container Mapping: Container names map back to service cards with the same name (when they exist)

EXAMPLES:
- postgres service (models.go) -> postgres compose service -> postgres-data container -> depends on postgres container
- airflow service (models.go) -> airflow compose service -> airflow container -> depends on airflow-init container -> depends on postgres compose service -> postgres-data container -> depends on postgres-server container -> postgres container
- cassandra service (models.go) -> cassandra compose service -> cassandra-data container -> depends on cassandra container

DEPENDENCY RESOLUTION:
- Service cards show dependencies based on their corresponding container's dependencies
- Dependencies are shown as container names in the UI
- Recursive dependencies traverse the container dependency graph
- Some containers may not have corresponding service cards (e.g., postgres-data, airflow-init)

IMPLEMENTATION NOTES:
- GetAllRunningServices() returns container statuses by container name
- resolveDependencies() works with container names for accurate dependency tracking
- Status monitoring uses container names to match real Docker/Podman container states
- Frontend maps container names back to service cards where possible
*/

// EnhancedService represents a complete service definition with all available information
type EnhancedService struct {
	// Basic service information from internal/core/models.go
	Name             string `json:"name"`
	Type             string `json:"type"`
	ConnectionCmd    string `json:"connection_cmd"`
	DefaultUser      string `json:"default_user,omitempty"`
	DefaultPassword  string `json:"default_password,omitempty"`
	RequiresPassword bool   `json:"requires_password"`

	// Runtime status information
	Status      string    `json:"status"` // "running", "stopped", "starting", "stopping", "error"
	StatusError string    `json:"status_error,omitempty"`
	LastUpdated time.Time `json:"last_updated"`

	// Container information from compose files
	ContainerName   string   `json:"container_name,omitempty"`
	ImageName       string   `json:"image_name,omitempty"`
	ImageExists     bool     `json:"image_exists"`
	ImagePullStatus string   `json:"image_pull_status,omitempty"` // "idle", "downloading", "complete", "error"
	AllContainers   []string `json:"all_containers"`              // All containers for this service

	// Dependencies (recursive, by container name)
	DirectDependencies    []string `json:"direct_dependencies"`    // Immediate dependencies
	RecursiveDependencies []string `json:"recursive_dependencies"` // All dependencies including self (flattened)
	DependsOnMe           []string `json:"depends_on_me"`          // Services that depend on this one

	// Port information
	ExposedPorts  []PortMapping `json:"exposed_ports"`  // Ports exposed to host
	InternalPorts []PortMapping `json:"internal_ports"` // Internal container ports

	// Web URLs
	WebUrls []WebURL `json:"web_urls,omitempty"` // Available web interfaces

	// Additional metadata from compose files
	HealthCheck   *HealthCheck `json:"health_check,omitempty"`
	Environment   []string     `json:"environment,omitempty"`
	Volumes       []string     `json:"volumes,omitempty"`
	Networks      []string     `json:"networks,omitempty"`
	RestartPolicy string       `json:"restart_policy,omitempty"`
}

// PortMapping represents a port mapping between host and container
type PortMapping struct {
	HostPort      string        `json:"host_port"`
	ContainerPort string        `json:"container_port"`
	Protocol      string        `json:"protocol,omitempty"` // "tcp", "udp"
	Type          core.PortType `json:"type"`               // Port type from core enum
	Description   string        `json:"description,omitempty"`
}

// WebURL represents a web interface available for the service
type WebURL struct {
	Name         string        `json:"name"`           // "Admin UI", "API", "Metrics"
	URL          string        `json:"url"`            // "http://localhost:8080"
	Port         string        `json:"port"`           // "8080"
	Path         string        `json:"path,omitempty"` // "/admin", "/api/v1"
	Type         core.PortType `json:"type"`           // Type of web interface
	Description  string        `json:"description,omitempty"`
	RequiresAuth bool          `json:"requires_auth,omitempty"`
}

// HealthCheck represents health check configuration
type HealthCheck struct {
	Test        []string `json:"test,omitempty"`
	Interval    string   `json:"interval,omitempty"`
	Timeout     string   `json:"timeout,omitempty"`
	Retries     int      `json:"retries,omitempty"`
	StartPeriod string   `json:"start_period,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshaling for HealthCheck to handle both string and array formats for test
func (hc *HealthCheck) UnmarshalYAML(value *yaml.Node) error {
	// Create a temporary struct for basic fields
	type tempHealthCheck struct {
		Interval    string `yaml:"interval,omitempty"`
		Timeout     string `yaml:"timeout,omitempty"`
		Retries     int    `yaml:"retries,omitempty"`
		StartPeriod string `yaml:"start_period,omitempty"`
	}

	var temp tempHealthCheck
	if err := value.Decode(&temp); err != nil {
		return err
	}

	// Copy basic fields
	hc.Interval = temp.Interval
	hc.Timeout = temp.Timeout
	hc.Retries = temp.Retries
	hc.StartPeriod = temp.StartPeriod

	// Handle the test field specially
	for i := 0; i < len(value.Content); i += 2 {
		if value.Content[i].Value == "test" {
			testNode := value.Content[i+1]

			if testNode.Kind == yaml.ScalarNode {
				// Single string format
				hc.Test = []string{testNode.Value}
			} else if testNode.Kind == yaml.SequenceNode {
				// Array format
				var testArray []string
				if err := testNode.Decode(&testArray); err != nil {
					return fmt.Errorf("failed to decode test array: %w", err)
				}
				hc.Test = testArray
			}
			break
		}
	}

	return nil
}

// ComposeFileParser handles parsing of docker-compose files
type ComposeFileParser struct {
	composeFiles   []string
	cachedServices map[string]*ComposeServiceConfig
	lastModified   map[string]time.Time
}

// ComposeServiceConfig represents parsed service configuration from compose files
type ComposeServiceConfig struct {
	ContainerName string              `yaml:"container_name,omitempty" json:"container_name,omitempty"`
	Image         string              `yaml:"image,omitempty" json:"image,omitempty"`
	DependsOn     FlexibleDependsOn   `yaml:"depends_on,omitempty" json:"depends_on,omitempty"`
	Ports         FlexibleStringArray `yaml:"ports,omitempty" json:"ports,omitempty"`
	Environment   FlexibleStringArray `yaml:"environment,omitempty" json:"environment,omitempty"`
	Volumes       FlexibleStringArray `yaml:"volumes,omitempty" json:"volumes,omitempty"`
	Networks      FlexibleStringArray `yaml:"networks,omitempty" json:"networks,omitempty"`
	HealthCheck   *HealthCheck        `yaml:"healthcheck,omitempty" json:"healthcheck,omitempty"`
	Restart       string              `yaml:"restart,omitempty" json:"restart,omitempty"`
	Command       FlexibleCommand     `yaml:"command,omitempty" json:"command,omitempty"`

	// Additional Docker Compose fields that we need to ignore/handle
	Entrypoint FlexibleCommand        `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	EnvFile    FlexibleStringArray    `yaml:"env_file,omitempty" json:"env_file,omitempty"`
	Hostname   string                 `yaml:"hostname,omitempty" json:"hostname,omitempty"`
	User       string                 `yaml:"user,omitempty" json:"user,omitempty"`
	CapAdd     FlexibleStringArray    `yaml:"cap_add,omitempty" json:"cap_add,omitempty"`
	Ulimits    map[string]interface{} `yaml:"ulimits,omitempty" json:"ulimits,omitempty"`
	Labels     map[string]string      `yaml:"labels,omitempty" json:"labels,omitempty"`
}

// FlexibleDependsOn handles both simple array and complex map formats for depends_on
type FlexibleDependsOn map[string]ComposeDependency

// UnmarshalYAML implements custom YAML unmarshaling for FlexibleDependsOn
func (fdo *FlexibleDependsOn) UnmarshalYAML(value *yaml.Node) error {
	if *fdo == nil {
		*fdo = make(map[string]ComposeDependency)
	}

	switch value.Kind {
	case yaml.SequenceNode:
		// Handle array format: ["postgres", "redis"]
		var deps []string
		if err := value.Decode(&deps); err != nil {
			return err
		}
		for _, dep := range deps {
			(*fdo)[dep] = ComposeDependency{Condition: "service_started"}
		}

	case yaml.MappingNode:
		// Handle map format: {postgres: {condition: service_healthy}}
		var deps map[string]ComposeDependency
		if err := value.Decode(&deps); err != nil {
			return err
		}
		for k, v := range deps {
			(*fdo)[k] = v
		}

	default:
		return fmt.Errorf("depends_on must be either an array or a map")
	}

	return nil
}

// FlexibleStringArray handles both single string and array formats
type FlexibleStringArray []string

// UnmarshalYAML implements custom YAML unmarshaling for FlexibleStringArray
func (fsa *FlexibleStringArray) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		// Single string format
		*fsa = []string{value.Value}

	case yaml.SequenceNode:
		// Array format
		var arr []string
		if err := value.Decode(&arr); err != nil {
			return fmt.Errorf("failed to decode array at line %d: %w", value.Line, err)
		}
		*fsa = arr

	case yaml.MappingNode:
		// Map format - convert to "KEY=value" format
		var m map[string]interface{}
		if err := value.Decode(&m); err != nil {
			return fmt.Errorf("failed to decode map at line %d: %w", value.Line, err)
		}

		var result []string
		for key, val := range m {
			result = append(result, fmt.Sprintf("%s=%v", key, val))
		}
		*fsa = result

	default:
		return fmt.Errorf("field at line %d column %d must be either a string, array, or map, got %v", value.Line, value.Column, value.Kind)
	}

	return nil
}

// FlexibleCommand handles both string, array, and map formats for command
type FlexibleCommand struct {
	Value interface{}
}

// UnmarshalYAML for handling flexible command formats
func (fc *FlexibleCommand) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		// Single string format
		fc.Value = value.Value

	case yaml.SequenceNode:
		// Array format
		var arr []string
		if err := value.Decode(&arr); err != nil {
			return err
		}
		fc.Value = arr

	case yaml.MappingNode:
		// Map format (less common but possible)
		var m map[string]interface{}
		if err := value.Decode(&m); err != nil {
			return err
		}
		fc.Value = m

	default:
		return fmt.Errorf("command must be a string, array, or map")
	}

	return nil
}

// ComposeDependency represents a dependency relationship in compose files
type ComposeDependency struct {
	Condition string `yaml:"condition,omitempty" json:"condition,omitempty"` // "service_started", "service_healthy", etc.
}

// ServiceManager handles all service operations with simplified logic
type ServiceManager struct {
	services    map[string]*EnhancedService
	instaDir    string
	parser      *ComposeFileParser
	runtimeInfo RuntimeInfo
	logger      Logger
}

// Logger interface for service manager
type Logger interface {
	Log(message string)
}

// RuntimeInfo provides access to container runtime operations
type RuntimeInfo interface {
	CheckContainerStatus(containerName string) (string, error)
	GetContainerLogs(containerName string, lines int) ([]string, error)
	StartService(serviceName string, persist bool) error
	StopService(serviceName string) error
	GetAllContainerStatuses() (map[string]string, error)
}

// RuntimeInfoWithName extends RuntimeInfo to provide runtime name detection
type RuntimeInfoWithName interface {
	RuntimeInfo
	GetRuntimeName() string
}

// NewServiceManager creates a new service manager instance
func NewServiceManager(instaDir string, runtimeInfo RuntimeInfo, logger Logger) *ServiceManager {
	return &ServiceManager{
		services:    make(map[string]*EnhancedService),
		instaDir:    instaDir,
		parser:      NewComposeFileParser(instaDir),
		runtimeInfo: runtimeInfo,
		logger:      logger,
	}
}

// NewComposeFileParser creates a new compose file parser
func NewComposeFileParser(instaDir string) *ComposeFileParser {
	return &ComposeFileParser{
		composeFiles:   []string{},
		cachedServices: make(map[string]*ComposeServiceConfig),
		lastModified:   make(map[string]time.Time),
	}
}

// LoadComposeFiles loads and parses docker-compose files
func (p *ComposeFileParser) LoadComposeFiles(instaDir string) error {
	// Define compose files to parse
	composeFiles := []string{
		filepath.Join(instaDir, "docker-compose.yaml"),
		filepath.Join(instaDir, "docker-compose-persist.yaml"),
	}

	p.composeFiles = composeFiles

	// Parse each compose file
	for _, filePath := range composeFiles {
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue // Skip non-existent files
		}

		// Check if file has been modified since last parse
		if p.shouldReparse(filePath) {
			err := p.parseComposeFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to parse compose file %s: %w", filePath, err)
			}
		}
	}

	return nil
}

// shouldReparse checks if a compose file should be reparsed based on modification time
func (p *ComposeFileParser) shouldReparse(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil {
		return true // Parse if we can't get file info
	}

	lastMod, exists := p.lastModified[filePath]
	if !exists || stat.ModTime().After(lastMod) {
		p.lastModified[filePath] = stat.ModTime()
		return true
	}

	return false
}

// parseComposeFile parses a single docker-compose file
func (p *ComposeFileParser) parseComposeFile(filePath string) error {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML
	var composeFile struct {
		Services map[string]ComposeServiceConfig `yaml:"services"`
	}

	err = yaml.Unmarshal(data, &composeFile)
	if err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	// Store parsed services in cache, merging with existing configurations
	for serviceName, serviceConfig := range composeFile.Services {
		if existingConfig, exists := p.cachedServices[serviceName]; exists {
			// Merge with existing configuration
			mergedConfig := p.mergeServiceConfigs(existingConfig, &serviceConfig)
			p.cachedServices[serviceName] = mergedConfig
		} else {
			// New service, store as-is
			p.cachedServices[serviceName] = &serviceConfig
		}
	}

	return nil
}

// mergeServiceConfigs merges two service configurations, with the new config taking precedence for non-empty fields
func (p *ComposeFileParser) mergeServiceConfigs(existing *ComposeServiceConfig, new *ComposeServiceConfig) *ComposeServiceConfig {
	merged := *existing // Start with existing config

	// Merge fields from new config, preferring non-empty values
	if new.ContainerName != "" {
		merged.ContainerName = new.ContainerName
	}
	if new.Image != "" {
		merged.Image = new.Image
	}
	if new.Restart != "" {
		merged.Restart = new.Restart
	}
	if new.Hostname != "" {
		merged.Hostname = new.Hostname
	}
	if new.User != "" {
		merged.User = new.User
	}

	// Merge slice fields by appending (avoiding duplicates)
	merged.Ports = p.mergeStringArrays(existing.Ports, new.Ports)
	merged.Environment = p.mergeStringArrays(existing.Environment, new.Environment)
	merged.Volumes = p.mergeStringArrays(existing.Volumes, new.Volumes)
	merged.Networks = p.mergeStringArrays(existing.Networks, new.Networks)
	merged.EnvFile = p.mergeStringArrays(existing.EnvFile, new.EnvFile)
	merged.CapAdd = p.mergeStringArrays(existing.CapAdd, new.CapAdd)

	// Merge maps
	if len(new.DependsOn) > 0 {
		if merged.DependsOn == nil {
			merged.DependsOn = make(FlexibleDependsOn)
		}
		for k, v := range new.DependsOn {
			merged.DependsOn[k] = v
		}
	}

	if len(new.Labels) > 0 {
		if merged.Labels == nil {
			merged.Labels = make(map[string]string)
		}
		for k, v := range new.Labels {
			merged.Labels[k] = v
		}
	}

	if len(new.Ulimits) > 0 {
		if merged.Ulimits == nil {
			merged.Ulimits = make(map[string]interface{})
		}
		for k, v := range new.Ulimits {
			merged.Ulimits[k] = v
		}
	}

	// Override complex objects if present in new config
	if new.HealthCheck != nil {
		merged.HealthCheck = new.HealthCheck
	}
	if new.Command.Value != nil {
		merged.Command = new.Command
	}
	if new.Entrypoint.Value != nil {
		merged.Entrypoint = new.Entrypoint
	}

	return &merged
}

// mergeStringArrays merges two FlexibleStringArray slices, avoiding duplicates
func (p *ComposeFileParser) mergeStringArrays(existing, new FlexibleStringArray) FlexibleStringArray {
	if len(new) == 0 {
		return existing
	}
	if len(existing) == 0 {
		return new
	}

	// Create a set to track existing items
	existingSet := make(map[string]bool)
	result := make(FlexibleStringArray, len(existing))
	copy(result, existing)

	for _, item := range existing {
		existingSet[item] = true
	}

	// Add new items that don't already exist
	for _, item := range new {
		if !existingSet[item] {
			result = append(result, item)
		}
	}

	return result
}

// GetServiceConfig returns the parsed configuration for a service
func (p *ComposeFileParser) GetServiceConfig(serviceName string) (*ComposeServiceConfig, bool) {
	config, exists := p.cachedServices[serviceName]
	return config, exists
}

// GetAllServiceConfigs returns all parsed service configurations
func (p *ComposeFileParser) GetAllServiceConfigs() map[string]*ComposeServiceConfig {
	return p.cachedServices
}

// LoadServices initializes all enhanced services by combining static definitions and compose file data
func (sm *ServiceManager) LoadServices() error {

	// First: Parse compose files to get enhancement data
	err := sm.parser.LoadComposeFiles(sm.instaDir)
	if err != nil {
		return err
	}

	// Second: Create services from core.Services (authoritative source)
	for serviceName, coreService := range core.Services {
		enhanced := &EnhancedService{
			Name:                  coreService.Name,
			Type:                  coreService.Type,
			ConnectionCmd:         coreService.ConnectionCmd,
			DefaultUser:           coreService.DefaultUser,
			DefaultPassword:       coreService.DefaultPassword,
			RequiresPassword:      coreService.RequiresPassword,
			Status:                "stopped", // Default status
			LastUpdated:           time.Now(),
			ContainerName:         serviceName,           // Default container name same as service name
			AllContainers:         []string{serviceName}, // Default to service name
			DirectDependencies:    []string{},            // Initialize empty dependency lists
			RecursiveDependencies: []string{serviceName}, // Include self as per architecture
			DependsOnMe:           []string{},            // Initialize empty
			ExposedPorts:          []PortMapping{},       // Initialize empty port lists
			InternalPorts:         []PortMapping{},       // Initialize empty port lists
			WebUrls:               []WebURL{},            // Initialize empty URL list
			Environment:           []string{},            // Initialize empty environment list
			Volumes:               []string{},            // Initialize empty volume list
			Networks:              []string{},            // Initialize empty network list
		}

		sm.services[serviceName] = enhanced
	}

	// Third: Enhance all services with compose file data
	for serviceName, service := range sm.services {
		sm.enhanceServiceWithComposeData(serviceName, service)
	}

	// Fourth: Resolve dependencies after all services are loaded
	sm.resolveDependencies()

	return nil
}

// getComposeServiceName returns the correct compose service name for a core service name
// This handles cases where the core service name doesn't match the compose service name
func (sm *ServiceManager) getComposeServiceName(coreServiceName string) string {
	// Special mappings for services where core name != compose service name
	serviceNameMapping := map[string]string{
		"postgres": "postgres-server", // postgres core service maps to postgres-server compose service
	}

	if composeServiceName, exists := serviceNameMapping[coreServiceName]; exists {
		return composeServiceName
	}

	// Default: core service name = compose service name
	return coreServiceName
}

// enhanceServiceWithComposeData adds compose file information to a service
func (sm *ServiceManager) enhanceServiceWithComposeData(serviceName string, service *EnhancedService) {
	// Get the correct compose service name (may be different from core service name)
	composeServiceName := sm.getComposeServiceName(serviceName)

	// Get parsed compose configuration using the mapped service name
	config, exists := sm.parser.GetServiceConfig(composeServiceName)
	if !exists {
		// Service not found in compose files, set defaults
		service.ContainerName = serviceName
		service.ImageName = serviceName
		service.AllContainers = []string{serviceName}
		// Ensure slices are initialized even if not found in compose
		if service.ExposedPorts == nil {
			service.ExposedPorts = []PortMapping{}
		}
		if service.WebUrls == nil {
			service.WebUrls = []WebURL{}
		}
		return
	}

	// Set container information from compose
	if config.ContainerName != "" {
		service.ContainerName = config.ContainerName
	} else {
		service.ContainerName = serviceName
	}

	if config.Image != "" {
		service.ImageName = config.Image
	}

	// Set additional compose metadata
	service.Environment = []string(config.Environment)
	service.Volumes = []string(config.Volumes)
	service.Networks = []string(config.Networks)
	service.RestartPolicy = config.Restart
	service.HealthCheck = config.HealthCheck

	// Find all containers for this service (services with same prefix or related containers)
	allContainers := []string{service.ContainerName}
	for otherServiceName, otherConfig := range sm.parser.GetAllServiceConfigs() {
		if otherServiceName != serviceName && otherServiceName != composeServiceName {
			// Check if this is a related container (e.g., postgres-data for postgres)
			if sm.isRelatedContainer(serviceName, otherServiceName) {
				if otherConfig.ContainerName != "" {
					allContainers = append(allContainers, otherConfig.ContainerName)
				} else {
					allContainers = append(allContainers, otherServiceName)
				}
			}
		}
	}
	service.AllContainers = allContainers

	// Parse port mappings from compose and map to known service ports
	service.ExposedPorts = sm.parsePortMappingsWithServiceKnowledge(serviceName, []string(config.Ports))

	// Generate web URLs based on mapped port information
	service.WebUrls = sm.generateWebURLsFromServicePorts(serviceName, service.ExposedPorts)
}

// isRelatedContainer checks if two services are related (e.g., postgres and postgres-data)
func (sm *ServiceManager) isRelatedContainer(mainService, otherService string) bool {
	// Simple heuristic: if other service name starts with main service name followed by "-"
	// Example: postgres-data is related to postgres
	return len(otherService) > len(mainService) &&
		otherService[:len(mainService)] == mainService &&
		len(otherService) > len(mainService) &&
		otherService[len(mainService)] == '-'
}

// parsePortMappingsWithServiceKnowledge converts compose port strings to PortMapping structs using service knowledge
func (sm *ServiceManager) parsePortMappingsWithServiceKnowledge(serviceName string, ports []string) []PortMapping {
	var mappings []PortMapping

	// Get service definition from core
	coreService, exists := core.Services[serviceName]
	if !exists {
		// Fall back to old method if service not found
		return sm.parsePortMappings(ports)
	}

	for _, portStr := range ports {
		mapping := sm.parsePortStringWithServiceKnowledge(portStr, coreService.Ports)
		if mapping != nil {
			mappings = append(mappings, *mapping)
		}
	}

	return mappings
}

// parsePortMappings converts compose port strings to PortMapping structs with basic type detection (fallback)
func (sm *ServiceManager) parsePortMappings(ports []string) []PortMapping {
	var mappings []PortMapping

	for _, portStr := range ports {
		mapping := sm.parsePortString(portStr)
		if mapping != nil {
			mappings = append(mappings, *mapping)
		}
	}

	return mappings
}

// parsePortStringWithServiceKnowledge parses a port string using service knowledge
func (sm *ServiceManager) parsePortStringWithServiceKnowledge(portStr string, servicePorts []core.ServicePort) *PortMapping {
	// Remove quotes and whitespace
	portStr = strings.Trim(portStr, "\"' ")
	if portStr == "" {
		return nil
	}

	// Expand environment variables before parsing
	portStr = expandEnvVars(portStr)

	// Default values
	mapping := &PortMapping{
		Protocol: "tcp",
		Type:     core.PortTypeOther,
	}

	// Parse different port formats:
	// "8080:80"
	// "127.0.0.1:8080:80"
	// "8080:80/tcp"
	// "8080"

	// Split by protocol if present
	protocolSplit := strings.Split(portStr, "/")
	portPart := protocolSplit[0]
	if len(protocolSplit) > 1 {
		mapping.Protocol = protocolSplit[1]
	}

	// Parse port mapping
	parts := strings.Split(portPart, ":")
	switch len(parts) {
	case 1:
		// Just a port number "8080"
		mapping.HostPort = parts[0]
		mapping.ContainerPort = parts[0]
	case 2:
		// "host:container" format "8080:80"
		mapping.HostPort = parts[0]
		mapping.ContainerPort = parts[1]
	case 3:
		// "ip:host:container" format "127.0.0.1:8080:80"
		mapping.HostPort = parts[1]
		mapping.ContainerPort = parts[2]
	default:
		return nil
	}

	// Try to match the container port to a known service port
	if containerPortNum, err := strconv.Atoi(mapping.ContainerPort); err == nil {
		for _, servicePort := range servicePorts {
			if servicePort.InternalPort == containerPortNum {
				mapping.Type = servicePort.Type
				mapping.Description = servicePort.Description
				break
			}
		}
	}

	return mapping
}

// expandEnvVars expands environment variables in the format ${VAR:-default}
func expandEnvVars(input string) string {
	// Regular expression to match ${VAR:-default} patterns
	re := regexp.MustCompile(`\$\{([^}:]+)(:-([^}]*))?\}`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		// Extract variable name and default value
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		varName := parts[1]
		defaultValue := ""
		if len(parts) >= 4 && parts[3] != "" {
			defaultValue = parts[3]
		}

		// Get environment variable value or use default
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return defaultValue
	})
}

// parsePortString parses a single port string and returns a PortMapping (fallback method)
func (sm *ServiceManager) parsePortString(portStr string) *PortMapping {
	// Remove quotes and whitespace
	portStr = strings.Trim(portStr, "\"' ")
	if portStr == "" {
		return nil
	}

	// Expand environment variables before parsing
	portStr = expandEnvVars(portStr)

	// Default values
	mapping := &PortMapping{
		Protocol: "tcp",
		Type:     core.PortTypeOther,
	}

	// Parse different port formats:
	// "8080:80"
	// "127.0.0.1:8080:80"
	// "8080:80/tcp"
	// "8080"

	// Split by protocol if present
	protocolSplit := strings.Split(portStr, "/")
	portPart := protocolSplit[0]
	if len(protocolSplit) > 1 {
		mapping.Protocol = protocolSplit[1]
	}

	// Parse port mapping
	parts := strings.Split(portPart, ":")
	switch len(parts) {
	case 1:
		// Just a port number "8080"
		mapping.HostPort = parts[0]
		mapping.ContainerPort = parts[0]
	case 2:
		// "host:container" format "8080:80"
		mapping.HostPort = parts[0]
		mapping.ContainerPort = parts[1]
	case 3:
		// "ip:host:container" format "127.0.0.1:8080:80"
		mapping.HostPort = parts[1]
		mapping.ContainerPort = parts[2]
	default:
		return nil
	}

	// No port type detection - use OTHER for fallback
	mapping.Type = core.PortTypeOther

	return mapping
}

// generateWebURLsFromServicePorts creates web URLs based on port mappings and service knowledge
func (sm *ServiceManager) generateWebURLsFromServicePorts(serviceName string, portMappings []PortMapping) []WebURL {
	var webUrls []WebURL

	// Get service definition from core
	coreService, exists := core.Services[serviceName]
	if !exists {
		// Fall back to old method if service not found
		return sm.generateWebURLs(serviceName, portMappings)
	}

	for _, port := range portMappings {
		if port.Type == core.PortTypeWebUI || port.Type == core.PortTypeAdmin || port.Type == core.PortTypeAPI {
			// Find the matching service port definition
			var servicePort *core.ServicePort
			if containerPortNum, err := strconv.Atoi(port.ContainerPort); err == nil {
				for _, sp := range coreService.Ports {
					if sp.InternalPort == containerPortNum {
						servicePort = &sp
						break
					}
				}
			}

			webUrl := WebURL{
				Port: port.HostPort,
				URL:  fmt.Sprintf("http://localhost:%s", port.HostPort),
				Type: port.Type,
			}

			// Use service port information if available
			if servicePort != nil {
				webUrl.Name = servicePort.Name
				webUrl.Description = servicePort.Description
				webUrl.Path = servicePort.Path
				webUrl.RequiresAuth = servicePort.RequiresAuth
			} else {
				// Fallback to generic names
				switch port.Type {
				case core.PortTypeWebUI:
					webUrl.Name = "Web Interface"
					webUrl.Path = "/"
				case core.PortTypeAdmin:
					webUrl.Name = "Admin Interface"
					webUrl.Path = "/admin"
				case core.PortTypeAPI:
					webUrl.Name = "API"
					webUrl.Path = "/"
				}
			}

			// Append path to URL if specified
			if webUrl.Path != "" && webUrl.Path != "/" {
				webUrl.URL = fmt.Sprintf("http://localhost:%s%s", port.HostPort, webUrl.Path)
			}

			webUrls = append(webUrls, webUrl)
		}
	}

	return webUrls
}

// generateWebURLs creates web URLs based on port mappings (fallback method)
func (sm *ServiceManager) generateWebURLs(serviceName string, portMappings []PortMapping) []WebURL {
	var webUrls []WebURL

	for _, port := range portMappings {
		if port.Type == core.PortTypeWebUI || port.Type == core.PortTypeAdmin {
			webUrl := WebURL{
				Port: port.HostPort,
				URL:  fmt.Sprintf("http://localhost:%s", port.HostPort),
				Type: port.Type,
			}

			// Set generic name and path
			switch port.Type {
			case core.PortTypeWebUI:
				webUrl.Name = "Web Interface"
				webUrl.Path = "/"
			case core.PortTypeAdmin:
				webUrl.Name = "Admin Interface"
				webUrl.Path = "/admin"
			}

			webUrls = append(webUrls, webUrl)
		}
	}

	return webUrls
}

// resolveDependencies builds dependency trees for all services, collect all container names
func (sm *ServiceManager) resolveDependencies() {
	// Create mapping from container name to service for efficient lookup
	containerToService := make(map[string]*EnhancedService)
	for _, service := range sm.services {
		containerToService[service.ContainerName] = service
	}

	// First pass: collect direct dependencies by container names
	for serviceName, service := range sm.services {
		config, exists := sm.parser.GetServiceConfig(serviceName)
		if exists && config.DependsOn != nil {
			for depServiceName := range config.DependsOn {
				// Get the container name from the dependency's compose service config
				depConfig, depExists := sm.parser.GetServiceConfig(depServiceName)
				if depExists && depConfig.ContainerName != "" {
					service.DirectDependencies = append(service.DirectDependencies, depConfig.ContainerName)
				}
			}
		}
	}

	// Second pass: resolve recursive dependencies including self container names
	for _, service := range sm.services {
		visited := make(map[string]bool)
		recursiveDeps := []string{}

		// Include the service's own container name as per requirements
		recursiveDeps = append(recursiveDeps, service.ContainerName)
		visited[service.ContainerName] = true

		// Recursively collect all dependencies by container name
		sm.collectRecursiveDependencies(service.ContainerName, visited, &recursiveDeps, containerToService)

		service.RecursiveDependencies = recursiveDeps
	}

	// Third pass: build reverse dependencies (who depends on me) by container names
	for _, service := range sm.services {
		for _, depContainerName := range service.DirectDependencies {
			if depService, exists := containerToService[depContainerName]; exists {
				depService.DependsOnMe = append(depService.DependsOnMe, service.ContainerName)
			}
		}
	}
}

// collectRecursiveDependencies recursively collects all dependencies for a service container name
func (sm *ServiceManager) collectRecursiveDependencies(serviceContainerName string, visited map[string]bool, result *[]string, containerToService map[string]*EnhancedService) {
	// First, try to find the service in our service map
	service, serviceExists := containerToService[serviceContainerName]
	if serviceExists {
		// Service exists - use its direct dependencies
		for _, depContainerName := range service.DirectDependencies {
			if !visited[depContainerName] {
				visited[depContainerName] = true
				*result = append(*result, depContainerName)
				// Recursively collect dependencies of dependencies
				sm.collectRecursiveDependencies(depContainerName, visited, result, containerToService)
			}
		}
	} else {
		// Service doesn't exist in our service map, but it might be a compose service
		// Try to find it in compose configs and follow its dependencies
		sm.collectDependenciesFromComposeService(serviceContainerName, visited, result, containerToService)
	}
}

// collectDependenciesFromComposeService collects dependencies from compose services that are not in the service map
func (sm *ServiceManager) collectDependenciesFromComposeService(containerName string, visited map[string]bool, result *[]string, containerToService map[string]*EnhancedService) {
	// Find the compose service that has this container name
	allConfigs := sm.parser.GetAllServiceConfigs()

	var targetServiceConfig *ComposeServiceConfig

	// Look for the compose service with matching container name
	for _, config := range allConfigs {
		if config.ContainerName == containerName {
			targetServiceConfig = config
			break
		}
	}

	// If we didn't find by container name, try by service name
	if targetServiceConfig == nil {
		if config, exists := allConfigs[containerName]; exists {
			targetServiceConfig = config
		}
	}

	if targetServiceConfig != nil && targetServiceConfig.DependsOn != nil {
		// Process dependencies from the compose service
		for depServiceName := range targetServiceConfig.DependsOn {
			// Get the container name for this dependency
			var depContainerName string
			if depConfig, exists := allConfigs[depServiceName]; exists && depConfig.ContainerName != "" {
				depContainerName = depConfig.ContainerName
			} else {
				// Fallback to service name as container name
				depContainerName = depServiceName
			}

			if !visited[depContainerName] {
				visited[depContainerName] = true
				*result = append(*result, depContainerName)
				// Recursively collect dependencies of dependencies
				sm.collectRecursiveDependencies(depContainerName, visited, result, containerToService)
			}
		}
	}
}

// GetService returns an enhanced service by name
func (sm *ServiceManager) GetService(name string) (*EnhancedService, bool) {
	service, exists := sm.services[name]
	return service, exists
}

// GetServiceByContainerName return an enhanced service by container name
func (sm *ServiceManager) GetServiceByContainerName(containerName string) (*EnhancedService, bool) {
	for _, service := range sm.services {
		//log container name
		sm.logger.Log(fmt.Sprintf("Checking container name: %s", service.ContainerName))
		if service.ContainerName == containerName {
			return service, true
		}
	}
	return nil, false
}

// GetAllServices returns a map of all enhanced services
func (sm *ServiceManager) GetAllServices() map[string]*EnhancedService {
	return sm.services
}

// ListServices returns a slice of all enhanced services, sort by name
func (sm *ServiceManager) ListServices() []*EnhancedService {
	var services []*EnhancedService
	for _, service := range sm.services {
		services = append(services, service)
	}
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name < services[j].Name
	})
	return services
}

// StartService starts a service using the runtime info
func (sm *ServiceManager) StartService(serviceName string, persist bool) error {
	// Update status to starting
	if service, exists := sm.services[serviceName]; exists {
		service.Status = "starting"
		service.LastUpdated = time.Now()
	}

	// Use runtime to start the service
	err := sm.runtimeInfo.StartService(serviceName, persist)
	if err != nil {
		// Update status to error
		if service, exists := sm.services[serviceName]; exists {
			service.Status = "error"
			service.StatusError = err.Error()
			service.LastUpdated = time.Now()
		}
		return err
	}

	// Update status will be done via UpdateServiceStatus call
	return nil
}

// StopService stops a service using the runtime info
func (sm *ServiceManager) StopService(serviceName string) error {
	// Update status to stopping
	if service, exists := sm.services[serviceName]; exists {
		service.Status = "stopping"
		service.LastUpdated = time.Now()
	}

	// Use runtime to stop the service
	err := sm.runtimeInfo.StopService(serviceName)
	if err != nil {
		// Update status to error
		if service, exists := sm.services[serviceName]; exists {
			service.Status = "error"
			service.StatusError = err.Error()
			service.LastUpdated = time.Now()
		}
		return err
	}

	// Update status to stopped
	if service, exists := sm.services[serviceName]; exists {
		service.Status = "stopped"
		service.StatusError = ""
		service.LastUpdated = time.Now()
	}

	return nil
}

// UpdateServiceStatus updates the status of a service from the runtime
func (sm *ServiceManager) UpdateServiceStatus(serviceName string) (string, error) {
	service, exists := sm.services[serviceName]
	if !exists {
		return "", fmt.Errorf("service %s not found", serviceName)
	}

	// Get status from runtime
	status, err := sm.runtimeInfo.CheckContainerStatus(service.ContainerName)
	if err != nil {
		service.Status = "error"
		service.StatusError = err.Error()
		service.LastUpdated = time.Now()
		return "error", err
	}

	// Update service status
	service.Status = status
	service.StatusError = ""
	service.LastUpdated = time.Now()

	return status, nil
}

// UpdateAllServiceStatuses updates the status of all services efficiently using a single runtime call
func (sm *ServiceManager) UpdateAllServiceStatuses() (map[string]string, error) {
	// Get all container statuses in a SINGLE call to docker/podman
	allContainerStatuses, err := sm.runtimeInfo.GetAllContainerStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to get container statuses: %w", err)
	}

	// Update service statuses based on container statuses
	statusMap := make(map[string]string)
	for serviceName, service := range sm.services {
		containerName := service.ContainerName

		if containerStatus, exists := allContainerStatuses[containerName]; exists {
			service.Status = containerStatus
			service.StatusError = ""
			statusMap[serviceName] = containerStatus
		} else {
			// Container doesn't exist, mark as stopped
			service.Status = "stopped"
			service.StatusError = ""
			statusMap[serviceName] = "stopped"
		}
		service.LastUpdated = time.Now()
	}

	return statusMap, nil
}

// ===========================================
// ServiceHandlerInterface Implementation
// ===========================================

// GetServiceDependencies returns the dependency information for a service
func (sm *ServiceManager) GetServiceDependencies(serviceName string) ([]string, error) {
	service, exists := sm.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	return service.RecursiveDependencies, nil
}

// StopAllServices stops all running services
func (sm *ServiceManager) StopAllServices() error {
	if sm.logger != nil {
		sm.logger.Log("Stopping all services")
	}

	// Get all running services first
	allStatuses, err := sm.UpdateAllServiceStatuses()
	if err != nil {
		return fmt.Errorf("failed to get service statuses: %w", err)
	}

	var errors []string
	for serviceName, status := range allStatuses {
		// If status contains "running", stop the service
		if strings.Contains(status, "running") {
			if err := sm.StopService(serviceName); err != nil {
				errors = append(errors, fmt.Sprintf("failed to stop %s: %v", serviceName, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors stopping services: %s", strings.Join(errors, "; "))
	}

	if sm.logger != nil {
		sm.logger.Log("Successfully stopped all services")
	}
	return nil
}

// ListEnhancedServices returns all enhanced services (alias for ListServices)
func (sm *ServiceManager) ListEnhancedServices() []*EnhancedService {
	return sm.ListServices()
}

// GetMultipleServiceStatuses returns the status of multiple services
func (sm *ServiceManager) GetMultipleServiceStatuses(serviceNames []string) (map[string]*EnhancedService, error) {
	result := make(map[string]*EnhancedService)

	for _, serviceName := range serviceNames {
		if service, exists := sm.services[serviceName]; exists {
			// Update status from runtime
			sm.UpdateServiceStatus(serviceName)
			result[serviceName] = service
		}
	}

	return result, nil
}

// GetAllRunningServices returns all services that are currently running indexed by container name
// (This aligns with the service-to-container architecture where dependencies are tracked by container names)
func (sm *ServiceManager) GetAllRunningServices() (map[string]*EnhancedService, error) {
	result := make(map[string]*EnhancedService)

	// Update all statuses first
	_, err := sm.UpdateAllServiceStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to update service statuses: %w", err)
	}

	// Iterate through all services and find ones that are running (contains "running" in the status)
	for _, service := range sm.services {
		if strings.Contains(service.Status, "running") {
			containerName := service.ContainerName
			if containerName == "" {
				containerName = service.Name
			}
			result[containerName] = service
		}
	}

	return result, nil
}

// GetAllServiceDependencies returns dependency information for all services
func (sm *ServiceManager) GetAllServiceDependencies() (map[string][]string, error) {
	result := make(map[string][]string)

	for serviceName, service := range sm.services {
		result[serviceName] = service.RecursiveDependencies
	}

	return result, nil
}

// GetAllServicesWithStatusAndDependencies returns all services with current status and dependencies
func (sm *ServiceManager) GetAllServicesWithStatusAndDependencies() ([]*EnhancedService, error) {
	// Update all statuses first
	_, err := sm.UpdateAllServiceStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to update service statuses: %w", err)
	}

	return sm.ListServices(), nil
}

// GetAllDependencyStatuses returns all services that other services depend on, including dependency containers
func (sm *ServiceManager) GetAllDependencyStatuses() (map[string]*EnhancedService, error) {
	result := make(map[string]*EnhancedService)

	// Update all statuses first
	_, err := sm.UpdateAllServiceStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to update service statuses: %w", err)
	}

	// Get all container statuses for dependency containers
	allContainerStatuses, err := sm.runtimeInfo.GetAllContainerStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to get container statuses: %w", err)
	}

	// Add all services that are dependencies of other services
	for _, service := range sm.services {
		if len(service.DependsOnMe) > 0 {
			result[service.ContainerName] = service
		}
	}

	// Add dependency containers that are not services themselves
	processedContainers := make(map[string]bool)
	for _, service := range sm.services {
		for _, depContainerName := range service.RecursiveDependencies {
			// Skip if already processed or if it's a service we already added
			if processedContainers[depContainerName] || result[depContainerName] != nil {
				continue
			}
			processedContainers[depContainerName] = true

			// Check if this dependency container is not a service itself
			var isService bool
			for _, existingService := range sm.services {
				if existingService.ContainerName == depContainerName {
					isService = true
					break
				}
			}

			if !isService {
				// Create a virtual service for this dependency container
				status := "stopped" // Default status
				if containerStatus, hasStatus := allContainerStatuses[depContainerName]; hasStatus {
					status = containerStatus
				}

				virtualService := &EnhancedService{
					Name:          depContainerName,
					ContainerName: depContainerName,
					Status:        status,
					Type:          "dependency", // Mark as dependency container
					LastUpdated:   time.Now(),
				}
				result[depContainerName] = virtualService
			}
		}
	}

	return result, nil
}

// StartServiceWithStatusUpdate starts a service and returns updated status for all affected services
func (sm *ServiceManager) StartServiceWithStatusUpdate(serviceName string, persist bool) (map[string]*EnhancedService, error) {
	if sm.logger != nil {
		sm.logger.Log(fmt.Sprintf("Starting service with status update: %s (persist: %t)", serviceName, persist))
	}

	// Start the service
	err := sm.StartService(serviceName, persist)
	if err != nil {
		return nil, err
	}

	// Get the service and its dependencies
	service, exists := sm.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Update status for the service and all its dependencies
	affectedServices := make(map[string]*EnhancedService)

	// Include the service itself
	sm.UpdateServiceStatus(serviceName)
	affectedServices[serviceName] = service

	// Include dependencies
	for _, depName := range service.RecursiveDependencies {
		if depService, exists := sm.services[depName]; exists {
			sm.UpdateServiceStatus(depName)
			affectedServices[depName] = depService
		}
	}

	return affectedServices, nil
}

// StopServiceWithStatusUpdate stops a service and returns updated status for all affected services
func (sm *ServiceManager) StopServiceWithStatusUpdate(serviceName string) (map[string]*EnhancedService, error) {
	if sm.logger != nil {
		sm.logger.Log(fmt.Sprintf("Stopping service with status update: %s", serviceName))
	}

	// Stop the service
	err := sm.StopService(serviceName)
	if err != nil {
		return nil, err
	}

	// Get the service and services that depend on it
	service, exists := sm.services[serviceName]
	if !exists {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// Update status for the service and all services that depend on it
	affectedServices := make(map[string]*EnhancedService)

	// Include the service itself
	affectedServices[serviceName] = service

	// Include services that depend on this one
	for _, dependentName := range service.DependsOnMe {
		if depService, exists := sm.services[dependentName]; exists {
			sm.UpdateServiceStatus(dependentName)
			affectedServices[dependentName] = depService
		}
	}

	return affectedServices, nil
}

// StopAllServicesWithStatusUpdate stops all services and returns updated status for all services
func (sm *ServiceManager) StopAllServicesWithStatusUpdate() (map[string]*EnhancedService, error) {
	if sm.logger != nil {
		sm.logger.Log("Stopping all services with status update")
	}

	// Stop all services
	err := sm.StopAllServices()
	if err != nil {
		return nil, err
	}

	// Update all statuses
	_, err = sm.UpdateAllServiceStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to update service statuses: %w", err)
	}

	// Return all services
	return sm.services, nil
}

// RefreshStatusFromContainers refreshes the status of all services from container runtime
func (sm *ServiceManager) RefreshStatusFromContainers() (map[string]*EnhancedService, error) {
	if sm.logger != nil {
		sm.logger.Log("Refreshing status from containers")
	}

	// Update all statuses from containers
	_, err := sm.UpdateAllServiceStatuses()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh service statuses: %w", err)
	}

	// Return all services with updated statuses
	return sm.services, nil
}

// CheckStartingServicesProgress checks the progress of services that are in starting state
func (sm *ServiceManager) CheckStartingServicesProgress() (map[string]*EnhancedService, error) {
	if sm.logger != nil {
		sm.logger.Log("Checking starting services progress")
	}

	result := make(map[string]*EnhancedService)

	// Find services that are in starting state
	for serviceName, service := range sm.services {
		if service.Status == "starting" {
			// Update status to check if it's done starting
			sm.UpdateServiceStatus(serviceName)
			result[serviceName] = service
		}
	}

	return result, nil
}
