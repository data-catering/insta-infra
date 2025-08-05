package handlers

import (
	"fmt"
	"net/http"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
	"github.com/gin-gonic/gin"
)

// CustomServiceHandler handles custom service management operations
type CustomServiceHandler struct {
	instaDir string
	logger   Logger
}

// NewCustomServiceHandler creates a new custom service handler
func NewCustomServiceHandler(instaDir string, logger Logger) *CustomServiceHandler {
	return &CustomServiceHandler{
		instaDir: instaDir,
		logger:   logger,
	}
}

// CustomComposeRequest represents the request body for custom compose operations
type CustomComposeRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
}

// CustomComposeResponse represents the response for custom compose operations
type CustomComposeResponse struct {
	ID          string                            `json:"id"`
	Name        string                            `json:"name"`
	Description string                            `json:"description"`
	Filename    string                            `json:"filename"`
	CreatedAt   string                            `json:"created_at"`
	UpdatedAt   string                            `json:"updated_at"`
	Services    []string                          `json:"services"`
	Metadata    *models.CustomServiceMetadata     `json:"metadata,omitempty"`
}

// ValidationResult represents the result of compose file validation
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// UploadCustomCompose handles POST /api/v1/custom/compose
// @Summary Upload a new custom compose file
// @Description Upload and register a new custom docker compose file
// @Tags custom
// @Accept json
// @Produce json
// @Param request body CustomComposeRequest true "Custom compose configuration"
// @Success 201 {object} CustomComposeResponse "Custom compose file created successfully"
// @Failure 400 {object} EnhancedErrorResponse "Invalid request or compose file"
// @Failure 500 {object} EnhancedErrorResponse "Internal server error"
// @Router /api/v1/custom/compose [post]
func (h *CustomServiceHandler) UploadCustomCompose(c *gin.Context) {
	if h.logger != nil {
		h.logger.Log("Uploading custom compose file")
	}

	var req CustomComposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate compose content
	if err := models.ValidateComposeContent(req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid compose content: %v", err)})
		return
	}

	// Create custom service registry
	registry, err := models.NewCustomServiceRegistry(h.instaDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize custom service registry: %v", err)})
		return
	}

	// Add the custom service
	metadata, err := registry.AddCustomService(req.Name, req.Description, req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to add custom service: %v", err)})
		return
	}

	// Create response
	response := CustomComposeResponse{
		ID:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Filename:    metadata.Filename,
		CreatedAt:   metadata.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   metadata.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Services:    metadata.Services,
		Metadata:    metadata,
	}

	if h.logger != nil {
		h.logger.Log(fmt.Sprintf("Successfully uploaded custom compose file: %s", metadata.ID))
	}

	c.JSON(http.StatusCreated, response)
}

// ListCustomCompose handles GET /api/v1/custom/compose
// @Summary List all custom compose files
// @Description Get a list of all registered custom compose files
// @Tags custom
// @Produce json
// @Success 200 {array} CustomComposeResponse "List of custom compose files"
// @Failure 500 {object} EnhancedErrorResponse "Internal server error"
// @Router /api/v1/custom/compose [get]
func (h *CustomServiceHandler) ListCustomCompose(c *gin.Context) {
	if h.logger != nil {
		h.logger.Log("Listing custom compose files")
	}

	// Create custom service registry
	registry, err := models.NewCustomServiceRegistry(h.instaDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize custom service registry: %v", err)})
		return
	}

	// Get all custom services
	customServices := registry.ListCustomServices()

	// Convert to response format
	var responses []CustomComposeResponse
	for _, metadata := range customServices {
		response := CustomComposeResponse{
			ID:          metadata.ID,
			Name:        metadata.Name,
			Description: metadata.Description,
			Filename:    metadata.Filename,
			CreatedAt:   metadata.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   metadata.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Services:    metadata.Services,
			Metadata:    metadata,
		}
		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, responses)
}

// GetCustomCompose handles GET /api/v1/custom/compose/:id
// @Summary Get a specific custom compose file
// @Description Get details and content of a specific custom compose file
// @Tags custom
// @Produce json
// @Param id path string true "Custom compose ID"
// @Success 200 {object} CustomComposeResponse "Custom compose file details"
// @Failure 404 {object} EnhancedErrorResponse "Custom compose file not found"
// @Failure 500 {object} EnhancedErrorResponse "Internal server error"
// @Router /api/v1/custom/compose/{id} [get]
func (h *CustomServiceHandler) GetCustomCompose(c *gin.Context) {
	serviceID := c.Param("id")
	
	if h.logger != nil {
		h.logger.Log(fmt.Sprintf("Getting custom compose file: %s", serviceID))
	}

	// Create custom service registry
	registry, err := models.NewCustomServiceRegistry(h.instaDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize custom service registry: %v", err)})
		return
	}

	// Get the custom service
	metadata, err := registry.GetCustomService(serviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Custom service not found: %v", err)})
		return
	}

	// Get content
	content, err := registry.GetCustomServiceContent(serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get custom service content: %v", err)})
		return
	}

	// Create response with content
	response := CustomComposeResponse{
		ID:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Filename:    metadata.Filename,
		CreatedAt:   metadata.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   metadata.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Services:    metadata.Services,
		Metadata:    metadata,
	}

	// Add content to response as additional field
	c.JSON(http.StatusOK, gin.H{
		"compose": response,
		"content": content,
	})
}

// UpdateCustomCompose handles PUT /api/v1/custom/compose/:id
// @Summary Update an existing custom compose file
// @Description Update the name, description, or content of an existing custom compose file
// @Tags custom
// @Accept json
// @Produce json
// @Param id path string true "Custom compose ID"
// @Param request body CustomComposeRequest true "Updated custom compose configuration"
// @Success 200 {object} CustomComposeResponse "Custom compose file updated successfully"
// @Failure 400 {object} EnhancedErrorResponse "Invalid request or compose file"
// @Failure 404 {object} EnhancedErrorResponse "Custom compose file not found"
// @Failure 500 {object} EnhancedErrorResponse "Internal server error"
// @Router /api/v1/custom/compose/{id} [put]
func (h *CustomServiceHandler) UpdateCustomCompose(c *gin.Context) {
	serviceID := c.Param("id")
	
	if h.logger != nil {
		h.logger.Log(fmt.Sprintf("Updating custom compose file: %s", serviceID))
	}

	var req CustomComposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	// Validate compose content
	if err := models.ValidateComposeContent(req.Content); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid compose content: %v", err)})
		return
	}

	// Create custom service registry
	registry, err := models.NewCustomServiceRegistry(h.instaDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize custom service registry: %v", err)})
		return
	}

	// Update the custom service
	metadata, err := registry.UpdateCustomService(serviceID, req.Name, req.Description, req.Content)
	if err != nil {
		if err.Error() == fmt.Sprintf("custom service not found: %s", serviceID) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to update custom service: %v", err)})
		}
		return
	}

	// Create response
	response := CustomComposeResponse{
		ID:          metadata.ID,
		Name:        metadata.Name,
		Description: metadata.Description,
		Filename:    metadata.Filename,
		CreatedAt:   metadata.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   metadata.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Services:    metadata.Services,
		Metadata:    metadata,
	}

	if h.logger != nil {
		h.logger.Log(fmt.Sprintf("Successfully updated custom compose file: %s", serviceID))
	}

	c.JSON(http.StatusOK, response)
}

// DeleteCustomCompose handles DELETE /api/v1/custom/compose/:id
// @Summary Delete a custom compose file
// @Description Remove a custom compose file and unregister its services
// @Tags custom
// @Produce json
// @Param id path string true "Custom compose ID"
// @Success 204 "Custom compose file deleted successfully"
// @Failure 404 {object} EnhancedErrorResponse "Custom compose file not found"
// @Failure 500 {object} EnhancedErrorResponse "Internal server error"
// @Router /api/v1/custom/compose/{id} [delete]
func (h *CustomServiceHandler) DeleteCustomCompose(c *gin.Context) {
	serviceID := c.Param("id")
	
	if h.logger != nil {
		h.logger.Log(fmt.Sprintf("Deleting custom compose file: %s", serviceID))
	}

	// Create custom service registry
	registry, err := models.NewCustomServiceRegistry(h.instaDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize custom service registry: %v", err)})
		return
	}

	// Delete the custom service
	err = registry.RemoveCustomService(serviceID)
	if err != nil {
		if err.Error() == fmt.Sprintf("custom service not found: %s", serviceID) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete custom service: %v", err)})
		}
		return
	}

	if h.logger != nil {
		h.logger.Log(fmt.Sprintf("Successfully deleted custom compose file: %s", serviceID))
	}

	c.Status(http.StatusNoContent)
}

// ValidateCustomCompose handles POST /api/v1/custom/validate
// @Summary Validate a custom compose file
// @Description Validate the syntax and structure of a custom docker compose file without saving it
// @Tags custom
// @Accept json
// @Produce json
// @Param content body map[string]string true "Compose content to validate" example(content):"version: '3.8'\nservices:\n  web:\n    image: nginx"
// @Success 200 {object} ValidationResult "Validation result"
// @Failure 400 {object} EnhancedErrorResponse "Invalid request"
// @Router /api/v1/custom/validate [post]
func (h *CustomServiceHandler) ValidateCustomCompose(c *gin.Context) {
	if h.logger != nil {
		h.logger.Log("Validating custom compose content")
	}

	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid request: %v", err)})
		return
	}

	content, exists := req["content"]
	if !exists || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'content' field in request"})
		return
	}

	// Get known services for dependency validation (you may want to get this dynamically)
	knownServices := []string{
		"postgres", "mysql", "redis", "mongodb", "elasticsearch", "kibana",
		"grafana", "prometheus", "jaeger", "zipkin", "consul", "etcd",
		"rabbitmq", "kafka", "zookeeper", "cassandra", "influxdb",
		"memcached", "nginx", "apache", "haproxy", "traefik",
	}

	// Use detailed validation
	detailedResult := models.ValidateComposeContentDetailed(content, knownServices)
	
	// Convert to our API response format
	result := ValidationResult{
		Valid:    detailedResult.Valid,
		Errors:   []string{},
		Warnings: []string{},
	}
	
	// Convert detailed errors to simple strings for API compatibility
	for _, err := range detailedResult.Errors {
		if err.ServiceName != "" {
			result.Errors = append(result.Errors, fmt.Sprintf("[%s] %s", err.ServiceName, err.Message))
		} else {
			result.Errors = append(result.Errors, err.Message)
		}
	}
	
	// Convert detailed warnings to simple strings
	for _, warning := range detailedResult.Warnings {
		if warning.ServiceName != "" {
			result.Warnings = append(result.Warnings, fmt.Sprintf("[%s] %s", warning.ServiceName, warning.Message))
		} else {
			result.Warnings = append(result.Warnings, warning.Message)
		}
	}
	
	// Add suggestions as warnings for now
	for _, suggestion := range detailedResult.Suggestions {
		var msg string
		if suggestion.ServiceName != "" {
			msg = fmt.Sprintf("[%s] Suggestion: %s", suggestion.ServiceName, suggestion.Message)
		} else {
			msg = fmt.Sprintf("Suggestion: %s", suggestion.Message)
		}
		result.Warnings = append(result.Warnings, msg)
	}

	c.JSON(http.StatusOK, result)
}

// GetCustomServiceStats handles GET /api/v1/custom/stats
// @Summary Get statistics about custom services
// @Description Get summary statistics about registered custom compose files
// @Tags custom
// @Produce json
// @Success 200 {object} map[string]interface{} "Custom service statistics"
// @Failure 500 {object} EnhancedErrorResponse "Internal server error"
// @Router /api/v1/custom/stats [get]
func (h *CustomServiceHandler) GetCustomServiceStats(c *gin.Context) {
	if h.logger != nil {
		h.logger.Log("Getting custom service statistics")
	}

	// Create custom service registry
	registry, err := models.NewCustomServiceRegistry(h.instaDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to initialize custom service registry: %v", err)})
		return
	}

	// Get all custom services
	customServices := registry.ListCustomServices()

	// Calculate statistics
	totalFiles := len(customServices)
	totalServices := 0
	servicesByType := make(map[string]int)

	for _, metadata := range customServices {
		totalServices += len(metadata.Services)
		// Since all custom services are type "Custom", we'll track by compose file
		servicesByType["Custom"] = totalFiles
	}

	stats := map[string]interface{}{
		"total_custom_files":    totalFiles,
		"total_custom_services": totalServices,
		"services_by_type":      servicesByType,
		"average_services_per_file": func() float64 {
			if totalFiles == 0 {
				return 0
			}
			return float64(totalServices) / float64(totalFiles)
		}(),
	}

	c.JSON(http.StatusOK, stats)
}