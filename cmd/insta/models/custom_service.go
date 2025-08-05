package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/data-catering/insta-infra/v2/internal/validation"
	"github.com/data-catering/insta-infra/v2/internal/core"
	"gopkg.in/yaml.v3"
)

// CustomServiceMetadata represents metadata for a custom compose file
type CustomServiceMetadata struct {
	ID          string    `json:"id"`          // Unique identifier for the custom service
	Name        string    `json:"name"`        // Display name for the custom service
	Description string    `json:"description"` // User-provided description
	Filename    string    `json:"filename"`    // Actual filename on disk
	CreatedAt   time.Time `json:"created_at"`  // When it was added
	UpdatedAt   time.Time `json:"updated_at"`  // When it was last modified
	Services    []string  `json:"services"`    // List of service names defined in the compose file
	Warnings    []string  `json:"warnings"`    // List of warnings (e.g., name clashes)
}

// ServiceClash represents a name clash between custom and built-in services
type ServiceClash struct {
	ServiceName     string `json:"service_name"`     // Name of the clashing service
	CustomServiceID string `json:"custom_service_id"` // ID of the custom service causing the clash
	CustomFileName  string `json:"custom_file_name"`  // Filename of the custom compose file
	ClashType       string `json:"clash_type"`       // Type of clash ("built-in", "custom")
	Resolution      string `json:"resolution"`       // How the clash was resolved
}

// CustomServiceRegistry manages the metadata registry for custom compose files
type CustomServiceRegistry struct {
	CustomDir    string                             `json:"-"`              // Directory where custom files are stored
	MetadataFile string                             `json:"-"`              // Path to metadata.json file
	Services     map[string]*CustomServiceMetadata `json:"services"`       // Map of service ID to metadata
	LastSync     time.Time                         `json:"last_sync"`      // Last time registry was synced
	Version      string                            `json:"version"`        // Registry format version
	Clashes      []ServiceClash                    `json:"clashes"`        // Record of service name clashes
}

// NewCustomServiceRegistry creates a new custom service registry
func NewCustomServiceRegistry(instaDir string) (*CustomServiceRegistry, error) {
	customDir := filepath.Join(instaDir, "custom")
	metadataFile := filepath.Join(customDir, "metadata.json")

	// Create custom directory if it doesn't exist
	if err := os.MkdirAll(customDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create custom directory: %w", err)
	}

	registry := &CustomServiceRegistry{
		CustomDir:    customDir,
		MetadataFile: metadataFile,
		Services:     make(map[string]*CustomServiceMetadata),
		Version:      "1.0",
		Clashes:      []ServiceClash{},
	}

	// Load existing metadata if it exists
	if err := registry.loadMetadata(); err != nil {
		// If metadata file doesn't exist, that's fine - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load existing metadata: %w", err)
		}
	}

	// Sync with filesystem to ensure consistency
	if err := registry.syncWithFilesystem(); err != nil {
		return nil, fmt.Errorf("failed to sync with filesystem: %w", err)
	}

	return registry, nil
}

// loadMetadata loads the metadata.json file
func (r *CustomServiceRegistry) loadMetadata() error {
	data, err := os.ReadFile(r.MetadataFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, r)
}

// saveMetadata saves the current metadata to metadata.json
func (r *CustomServiceRegistry) saveMetadata() error {
	r.LastSync = time.Now()
	
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(r.MetadataFile, data, 0644)
}

// syncWithFilesystem syncs the metadata registry with actual files on disk
func (r *CustomServiceRegistry) syncWithFilesystem() error {
	// Get all .yaml and .yml files in custom directory
	files, err := filepath.Glob(filepath.Join(r.CustomDir, "*.y*ml"))
	if err != nil {
		return fmt.Errorf("failed to list custom compose files: %w", err)
	}

	// Track which services we've seen on disk
	foundServices := make(map[string]bool)

	// Process each file
	for _, file := range files {
		filename := filepath.Base(file)
		
		// Skip metadata.json
		if filename == "metadata.json" {
			continue
		}

		// Find existing metadata entry or create new one
		var metadata *CustomServiceMetadata
		var serviceID string

		// Look for existing entry by filename
		for id, service := range r.Services {
			if service.Filename == filename {
				metadata = service
				serviceID = id
				break
			}
		}

		// If not found, create new metadata entry
		if metadata == nil {
			serviceID = generateServiceID(filename)
			// Ensure unique ID
			for r.Services[serviceID] != nil {
				serviceID = generateServiceID(filename + "_" + fmt.Sprintf("%d", time.Now().UnixNano()))
			}

			metadata = &CustomServiceMetadata{
				ID:        serviceID,
				Filename:  filename,
				CreatedAt: time.Now(),
			}
		}

		// Update file info
		fileInfo, err := os.Stat(file)
		if err != nil {
			continue // Skip files we can't stat
		}
		metadata.UpdatedAt = fileInfo.ModTime()

		// Parse compose file to extract service names and metadata
		if err := r.updateMetadataFromFile(metadata, file); err != nil {
			// Log error but continue - we don't want sync to fail for one bad file
			fmt.Fprintf(os.Stderr, "Warning: failed to parse custom compose file %s: %v\n", filename, err)
		}

		r.Services[serviceID] = metadata
		foundServices[serviceID] = true
	}

	// Remove metadata entries for files that no longer exist
	for serviceID := range r.Services {
		if !foundServices[serviceID] {
			delete(r.Services, serviceID)
		}
	}

	// Save updated metadata
	return r.saveMetadata()
}

// updateMetadataFromFile updates metadata by parsing the compose file
func (r *CustomServiceRegistry) updateMetadataFromFile(metadata *CustomServiceMetadata, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse YAML to extract basic structure
	var compose map[string]interface{}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract service names
	var serviceNames []string
	if services, ok := compose["services"].(map[string]interface{}); ok {
		for serviceName := range services {
			serviceNames = append(serviceNames, serviceName)
		}
	}
	metadata.Services = serviceNames

	// If name is not set, derive from filename
	if metadata.Name == "" {
		metadata.Name = strings.TrimSuffix(metadata.Filename, filepath.Ext(metadata.Filename))
	}

	return nil
}

// generateServiceID generates a unique service ID from filename
func generateServiceID(filename string) string {
	// Remove extension and replace special characters
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ToLower(name)
	
	// Add timestamp suffix to ensure uniqueness
	return fmt.Sprintf("custom_%s_%d", name, time.Now().Unix())
}

// AddCustomService adds a new custom compose file
func (r *CustomServiceRegistry) AddCustomService(name, description, content string) (*CustomServiceMetadata, error) {
	// Validate content is valid YAML
	var compose map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &compose); err != nil {
		return nil, fmt.Errorf("invalid YAML content: %w", err)
	}

	// Validate it's a docker compose file
	if _, ok := compose["services"]; !ok {
		return nil, fmt.Errorf("compose file must contain 'services' section")
	}

	// Generate filename
	filename := generateFilename(name)
	filePath := filepath.Join(r.CustomDir, filename)

	// Ensure filename is unique
	counter := 1
	for {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			break
		}
		filename = generateFilename(fmt.Sprintf("%s_%d", name, counter))
		filePath = filepath.Join(r.CustomDir, filename)
		counter++
	}

	// Write file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write compose file: %w", err)
	}

	// Create metadata
	serviceID := generateServiceID(filename)
	metadata := &CustomServiceMetadata{
		ID:          serviceID,
		Name:        name,
		Description: description,
		Filename:    filename,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Update metadata from file content
	if err := r.updateMetadataFromFile(metadata, filePath); err != nil {
		// Clean up file if metadata extraction fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to extract metadata from compose file: %w", err)
	}

	// Add to registry
	r.Services[serviceID] = metadata

	// Save metadata
	if err := r.saveMetadata(); err != nil {
		// Clean up file if saving metadata fails
		os.Remove(filePath)
		delete(r.Services, serviceID)
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return metadata, nil
}

// UpdateCustomService updates an existing custom compose file
func (r *CustomServiceRegistry) UpdateCustomService(serviceID, name, description, content string) (*CustomServiceMetadata, error) {
	metadata, exists := r.Services[serviceID]
	if !exists {
		return nil, fmt.Errorf("custom service not found: %s", serviceID)
	}

	// Validate content is valid YAML
	var compose map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &compose); err != nil {
		return nil, fmt.Errorf("invalid YAML content: %w", err)
	}

	// Validate it's a docker compose file
	if _, ok := compose["services"]; !ok {
		return nil, fmt.Errorf("compose file must contain 'services' section")
	}

	// Write updated content
	filePath := filepath.Join(r.CustomDir, metadata.Filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write compose file: %w", err)
	}

	// Update metadata from file content first
	if err := r.updateMetadataFromFile(metadata, filePath); err != nil {
		return nil, fmt.Errorf("failed to extract metadata from compose file: %w", err)
	}

	// Update metadata fields after parsing (to ensure UpdatedAt is current)
	metadata.Name = name
	metadata.Description = description
	metadata.UpdatedAt = time.Now()

	// Save metadata
	if err := r.saveMetadata(); err != nil {
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return metadata, nil
}

// RemoveCustomService removes a custom compose file
func (r *CustomServiceRegistry) RemoveCustomService(serviceID string) error {
	metadata, exists := r.Services[serviceID]
	if !exists {
		return fmt.Errorf("custom service not found: %s", serviceID)
	}

	// Remove file
	filePath := filepath.Join(r.CustomDir, metadata.Filename)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove compose file: %w", err)
	}

	// Remove from registry
	delete(r.Services, serviceID)

	// Save metadata
	return r.saveMetadata()
}

// GetCustomService returns metadata for a specific custom service
func (r *CustomServiceRegistry) GetCustomService(serviceID string) (*CustomServiceMetadata, error) {
	metadata, exists := r.Services[serviceID]
	if !exists {
		return nil, fmt.Errorf("custom service not found: %s", serviceID)
	}
	return metadata, nil
}

// ListCustomServices returns all custom service metadata
func (r *CustomServiceRegistry) ListCustomServices() []*CustomServiceMetadata {
	var services []*CustomServiceMetadata
	for _, metadata := range r.Services {
		services = append(services, metadata)
	}
	return services
}

// GetCustomServiceContent returns the content of a custom compose file
func (r *CustomServiceRegistry) GetCustomServiceContent(serviceID string) (string, error) {
	metadata, exists := r.Services[serviceID]
	if !exists {
		return "", fmt.Errorf("custom service not found: %s", serviceID)
	}

	filePath := filepath.Join(r.CustomDir, metadata.Filename)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read compose file: %w", err)
	}

	return string(data), nil
}

// GetAllCustomComposeFiles returns paths to all custom compose files
func (r *CustomServiceRegistry) GetAllCustomComposeFiles() []string {
	var files []string
	for _, metadata := range r.Services {
		filePath := filepath.Join(r.CustomDir, metadata.Filename)
		files = append(files, filePath)
	}
	return files
}

// ValidateComposeContent validates that content is a valid docker compose file
func ValidateComposeContent(content string) error {
	// Use the comprehensive validator for basic validation
	validator := validation.NewComposeValidator([]string{}) // We'll enhance this with known services later
	result := validator.ValidateComposeContent(content)
	
	if !result.Valid {
		// Return the first error for simple error reporting
		if len(result.Errors) > 0 {
			return fmt.Errorf("%s", result.Errors[0].Message)
		}
	}
	
	return nil
}

// ValidateComposeContentDetailed validates content and returns detailed validation results
func ValidateComposeContentDetailed(content string, knownServices []string) *validation.ValidationResult {
	validator := validation.NewComposeValidator(knownServices)
	return validator.ValidateComposeContent(content)
}

// generateFilename generates a safe filename from a service name
func generateFilename(name string) string {
	// Replace unsafe characters
	safe := strings.ReplaceAll(name, " ", "_")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	safe = strings.ReplaceAll(safe, ":", "_")
	safe = strings.ReplaceAll(safe, "*", "_")
	safe = strings.ReplaceAll(safe, "?", "_")
	safe = strings.ReplaceAll(safe, "\"", "_")
	safe = strings.ReplaceAll(safe, "<", "_")
	safe = strings.ReplaceAll(safe, ">", "_")
	safe = strings.ReplaceAll(safe, "|", "_")
	
	// Ensure it ends with .yaml
	if !strings.HasSuffix(safe, ".yaml") && !strings.HasSuffix(safe, ".yml") {
		safe += ".yaml"
	}
	
	return safe
}

// RegisterAllCustomServices registers all custom services with the core service registry
func (r *CustomServiceRegistry) RegisterAllCustomServices() error {
	for _, metadata := range r.Services {
		if err := r.registerCustomService(metadata); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to register custom service %s: %v\n", metadata.Name, err)
		}
	}
	return nil
}

// registerCustomService registers a single custom service with the core service registry
func (r *CustomServiceRegistry) registerCustomService(metadata *CustomServiceMetadata) error {
	// Get compose file content to extract port information
	content, err := r.GetCustomServiceContent(metadata.ID)
	if err != nil {
		return fmt.Errorf("failed to read compose content: %w", err)
	}

	// Parse compose file to extract service information
	var compose map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &compose); err != nil {
		return fmt.Errorf("failed to parse compose YAML: %w", err)
	}

	services, ok := compose["services"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("compose file must contain 'services' section")
	}

	// Clear any existing warnings for this metadata
	metadata.Warnings = []string{}

	// Register each service defined in the compose file
	for serviceName := range services {
		// Check for name clashes with built-in services
		existingService, hasBuiltIn := core.Services[serviceName]
		var hasCustomClash bool

		// Check if this is already a custom service from another file
		if hasBuiltIn && existingService.Type == "Custom" {
			hasCustomClash = true
		}

		// Create a core.Service entry for this custom service
		customService := core.Service{
			Name:             serviceName,
			Type:             "Custom",
			ConnectionCmd:    "bash", // Default to bash for custom services
			RequiresPassword: false,
			Ports:            extractPortsFromCompose(services[serviceName]),
		}

		// Handle name clashes according to precedence rules
		if hasBuiltIn && !hasCustomClash {
			// Clash with built-in service - custom takes precedence
			clash := ServiceClash{
				ServiceName:     serviceName,
				CustomServiceID: metadata.ID,
				CustomFileName:  metadata.Filename,
				ClashType:       "built-in",
				Resolution:      "custom-service-takes-precedence",
			}
			r.addOrUpdateClash(clash)
			
			warning := fmt.Sprintf("Service '%s' overrides built-in service of the same name", serviceName)
			metadata.Warnings = append(metadata.Warnings, warning)
			
			fmt.Fprintf(os.Stderr, "Warning: Custom service '%s' in file '%s' overrides built-in service\n", 
				serviceName, metadata.Filename)
		} else if hasCustomClash {
			// Clash with another custom service - latest takes precedence
			clash := ServiceClash{
				ServiceName:     serviceName,
				CustomServiceID: metadata.ID,
				CustomFileName:  metadata.Filename,
				ClashType:       "custom",
				Resolution:      "latest-custom-service-takes-precedence",
			}
			r.addOrUpdateClash(clash)
			
			warning := fmt.Sprintf("Service '%s' overrides another custom service of the same name", serviceName)
			metadata.Warnings = append(metadata.Warnings, warning)
			
			fmt.Fprintf(os.Stderr, "Warning: Custom service '%s' in file '%s' overrides another custom service\n", 
				serviceName, metadata.Filename)
		}

		// Register the service (this will override any existing service)
		core.RegisterCustomService(serviceName, customService)
	}

	return nil
}

// extractPortsFromCompose extracts port information from a compose service definition
func extractPortsFromCompose(serviceConfig interface{}) []core.ServicePort {
	var ports []core.ServicePort

	serviceMap, ok := serviceConfig.(map[string]interface{})
	if !ok {
		return ports
	}

	// Extract ports from the ports section
	if portsSection, exists := serviceMap["ports"]; exists {
		if portsList, ok := portsSection.([]interface{}); ok {
			for _, portEntry := range portsList {
				if portStr, ok := portEntry.(string); ok {
					// Parse port mapping (e.g., "8080:80", "3000:3000/tcp")
					if port := parsePortMapping(portStr); port != nil {
						ports = append(ports, *port)
					}
				}
			}
		}
	}

	// Extract ports from expose section
	if exposeSection, exists := serviceMap["expose"]; exists {
		if exposeList, ok := exposeSection.([]interface{}); ok {
			for _, portEntry := range exposeList {
				if portStr, ok := portEntry.(string); ok {
					if portNum := parsePort(portStr); portNum > 0 {
						ports = append(ports, core.ServicePort{
							InternalPort: portNum,
							Type:         core.PortTypeOther,
							Name:         fmt.Sprintf("Port %d", portNum),
							Description:  fmt.Sprintf("Exposed port %d", portNum),
						})
					}
				}
			}
		}
	}

	return ports
}

// parsePortMapping parses a docker compose port mapping string
func parsePortMapping(portStr string) *core.ServicePort {
	// Handle various formats: "8080:80", "3000:3000/tcp", "127.0.0.1:8080:80"
	parts := strings.Split(portStr, ":")
	if len(parts) < 2 {
		return nil
	}

	// Get the container port (last part)
	containerPortStr := parts[len(parts)-1]
	
	// Remove protocol if present
	if strings.Contains(containerPortStr, "/") {
		containerPortStr = strings.Split(containerPortStr, "/")[0]
	}

	containerPort := parsePort(containerPortStr)
	if containerPort <= 0 {
		return nil
	}

	// Determine port type based on common ports
	portType := determinePortType(containerPort)

	return &core.ServicePort{
		InternalPort: containerPort,
		Type:         portType,
		Name:         fmt.Sprintf("Port %d", containerPort),
		Description:  fmt.Sprintf("Custom service port %d", containerPort),
	}
}

// parsePort parses a port string to integer
func parsePort(portStr string) int {
	portNum := 0
	fmt.Sscanf(portStr, "%d", &portNum)
	return portNum
}

// determinePortType determines the port type based on common port numbers
func determinePortType(port int) core.PortType {
	switch port {
	case 80, 8080, 3000, 8000, 8081, 8082, 8088, 8090, 3001, 3002, 4000, 4200, 5000, 5173:
		return core.PortTypeWebUI
	case 443, 8443:
		return core.PortTypeWebUI
	case 3306, 5432, 1433, 1521, 27017, 6379, 11211, 5984, 9200, 9300:
		return core.PortTypeDatabase
	case 8001, 8002, 9001, 9002:
		return core.PortTypeAdmin
	case 9000:
		return core.PortTypeWebUI // Common web UI port
	case 9090, 9100, 8086, 9093, 9094:
		return core.PortTypeMetrics
	case 5672, 15672, 1883, 8883, 61616, 9092, 29092, 4222, 8222:
		return core.PortTypeMessaging
	default:
		// Try to guess based on port ranges
		if port >= 8000 && port <= 8999 {
			return core.PortTypeWebUI
		}
		if port >= 9000 && port <= 9999 {
			return core.PortTypeAPI
		}
		return core.PortTypeOther
	}
}

// addOrUpdateClash adds or updates a service clash record
func (r *CustomServiceRegistry) addOrUpdateClash(clash ServiceClash) {
	// Remove any existing clash for the same service name
	for i := len(r.Clashes) - 1; i >= 0; i-- {
		if r.Clashes[i].ServiceName == clash.ServiceName {
			r.Clashes = append(r.Clashes[:i], r.Clashes[i+1:]...)
		}
	}
	
	// Add the new clash record
	r.Clashes = append(r.Clashes, clash)
}

// GetServiceClashes returns all service clash records
func (r *CustomServiceRegistry) GetServiceClashes() []ServiceClash {
	return r.Clashes
}

// ClearServiceClashes removes all clash records
func (r *CustomServiceRegistry) ClearServiceClashes() {
	r.Clashes = []ServiceClash{}
}

// UnregisterAllCustomServices removes all custom services from the core service registry
func (r *CustomServiceRegistry) UnregisterAllCustomServices() {
	for _, metadata := range r.Services {
		for _, serviceName := range metadata.Services {
			core.UnregisterCustomService(serviceName)
		}
	}
}