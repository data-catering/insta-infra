package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	logs []string
}

func (m *MockLogger) Log(message string) {
	m.logs = append(m.logs, message)
}

func TestCustomServiceHandler_UploadCustomCompose(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.POST("/custom/compose", handler.UploadCustomCompose)

	// Test valid compose file
	validCompose := CustomComposeRequest{
		Name:        "test-service",
		Description: "Test description",
		Content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
	}

	jsonData, _ := json.Marshal(validCompose)
	req, _ := http.NewRequest("POST", "/custom/compose", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response CustomComposeResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-service", response.Name)
	assert.Equal(t, "Test description", response.Description)
	assert.Equal(t, []string{"web"}, response.Services)
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Filename)
}

func TestCustomServiceHandler_UploadCustomCompose_InvalidYAML(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.POST("/custom/compose", handler.UploadCustomCompose)

	// Test invalid YAML
	invalidCompose := CustomComposeRequest{
		Name:        "invalid-service",
		Description: "Invalid description",
		Content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
  - invalid yaml structure
`,
	}

	jsonData, _ := json.Marshal(invalidCompose)
	req, _ := http.NewRequest("POST", "/custom/compose", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid compose content")
}

func TestCustomServiceHandler_ListCustomCompose(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.GET("/custom/compose", handler.ListCustomCompose)
	router.POST("/custom/compose", handler.UploadCustomCompose)

	// First, upload a compose file
	validCompose := CustomComposeRequest{
		Name:        "test-service",
		Description: "Test description",
		Content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
	}

	jsonData, _ := json.Marshal(validCompose)
	req, _ := http.NewRequest("POST", "/custom/compose", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	// Now list the compose files
	req, _ = http.NewRequest("GET", "/custom/compose", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responses []CustomComposeResponse
	err = json.Unmarshal(w.Body.Bytes(), &responses)
	require.NoError(t, err)

	assert.Len(t, responses, 1)
	assert.Equal(t, "test-service", responses[0].Name)
	assert.Equal(t, "Test description", responses[0].Description)
	assert.Equal(t, []string{"web"}, responses[0].Services)
}

func TestCustomServiceHandler_GetCustomCompose(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.GET("/custom/compose/:id", handler.GetCustomCompose)
	router.POST("/custom/compose", handler.UploadCustomCompose)

	// First, upload a compose file
	validCompose := CustomComposeRequest{
		Name:        "test-service",
		Description: "Test description",
		Content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
	}

	jsonData, _ := json.Marshal(validCompose)
	req, _ := http.NewRequest("POST", "/custom/compose", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var uploadResponse CustomComposeResponse
	err = json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	require.NoError(t, err)

	// Now get the specific compose file
	req, _ = http.NewRequest("GET", "/custom/compose/"+uploadResponse.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "compose")
	assert.Contains(t, response, "content")

	compose := response["compose"].(map[string]interface{})
	assert.Equal(t, "test-service", compose["name"])
	assert.Equal(t, "Test description", compose["description"])

	content := response["content"].(string)
	assert.Contains(t, content, "nginx:latest")
}

func TestCustomServiceHandler_ValidateCustomCompose(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.POST("/custom/validate", handler.ValidateCustomCompose)

	// Test valid compose content
	validRequest := map[string]string{
		"content": `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
	}

	jsonData, _ := json.Marshal(validRequest)
	req, _ := http.NewRequest("POST", "/custom/validate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ValidationResult
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.True(t, response.Valid)
	assert.Empty(t, response.Errors)
}

func TestCustomServiceHandler_ValidateCustomCompose_Invalid(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.POST("/custom/validate", handler.ValidateCustomCompose)

	// Test invalid compose content (no services section)
	invalidRequest := map[string]string{
		"content": `
version: '3.8'
networks:
  default:
    driver: bridge
`,
	}

	jsonData, _ := json.Marshal(invalidRequest)
	req, _ := http.NewRequest("POST", "/custom/validate", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ValidationResult
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Valid)
	assert.NotEmpty(t, response.Errors)
	assert.Contains(t, response.Errors[0], "services")
}

func TestCustomServiceHandler_DeleteCustomCompose(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.POST("/custom/compose", handler.UploadCustomCompose)
	router.DELETE("/custom/compose/:id", handler.DeleteCustomCompose)
	router.GET("/custom/compose", handler.ListCustomCompose)

	// First, upload a compose file
	validCompose := CustomComposeRequest{
		Name:        "test-service",
		Description: "Test description",
		Content: `
services:
  web:
    image: nginx:latest
    ports:
      - "8080:80"
`,
	}

	jsonData, _ := json.Marshal(validCompose)
	req, _ := http.NewRequest("POST", "/custom/compose", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var uploadResponse CustomComposeResponse
	err = json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	require.NoError(t, err)

	// Now delete the compose file
	req, _ = http.NewRequest("DELETE", "/custom/compose/"+uploadResponse.ID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify it's deleted by listing
	req, _ = http.NewRequest("GET", "/custom/compose", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var responses []CustomComposeResponse
	err = json.Unmarshal(w.Body.Bytes(), &responses)
	require.NoError(t, err)

	assert.Len(t, responses, 0)
}

func TestCustomServiceHandler_GetCustomServiceStats(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "insta-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup
	gin.SetMode(gin.TestMode)
	router := gin.New()
	logger := &MockLogger{}
	handler := NewCustomServiceHandler(tempDir, logger)

	router.GET("/custom/stats", handler.GetCustomServiceStats)

	// Get stats for empty registry
	req, _ := http.NewRequest("GET", "/custom/stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)

	assert.Equal(t, float64(0), stats["total_custom_files"])
	assert.Equal(t, float64(0), stats["total_custom_services"])
	assert.Equal(t, float64(0), stats["average_services_per_file"])
}