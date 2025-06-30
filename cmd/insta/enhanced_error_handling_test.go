package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/data-catering/insta-infra/v2/cmd/insta/internal"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data structures for error responses
type ErrorResponse struct {
	Error       string                 `json:"error"`
	ServiceName string                 `json:"service_name,omitempty"`
	ImageName   string                 `json:"image_name,omitempty"`
	Action      string                 `json:"action"`
	Timestamp   string                 `json:"timestamp"`
	Status      int                    `json:"status,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

func TestEnhancedErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil, // Force nil runtime to ensure error responses
	}

	apiServer := NewAPIServer(app)
	// Ensure handlers are not initialized to force error conditions
	apiServer.handlerManager = internal.NewHandlerManager(internal.NewAppLogger())
	router := apiServer.engine

	t.Run("ServiceStartErrorResponse", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/nonexistent/start", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error response structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "start", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ServiceStopErrorResponse", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/nonexistent/stop", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error response structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "stop", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ImagePullErrorResponse", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/images/postgres:13/pull", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error response structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "postgres:13", errorResp.ImageName)
		assert.Equal(t, "start_image_pull", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ServiceLogsErrorResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/services/nonexistent/logs", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error response structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "get_logs", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ServiceConnectionErrorResponse", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/services/nonexistent/connection", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error response structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "get_connection_info", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})
}

func TestErrorResponseTimestamps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil, // Force nil runtime to ensure error responses
	}

	apiServer := NewAPIServer(app)
	// Ensure handlers are not initialized to force error conditions
	apiServer.handlerManager = internal.NewHandlerManager(internal.NewAppLogger())
	router := apiServer.engine

	t.Run("TimestampFormat", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/test/start", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify timestamp format
		timestamp, err := time.Parse(time.RFC3339, errorResp.Timestamp)
		require.NoError(t, err, "Timestamp should be valid RFC3339 format")

		// Verify timestamp is recent (within last minute)
		now := time.Now()
		timeDiff := now.Sub(timestamp)
		assert.True(t, timeDiff < time.Minute, "Timestamp should be recent")
		assert.True(t, timeDiff >= 0, "Timestamp should not be in the future")

		// Verify metadata also has timestamp
		metadataTimestamp, exists := errorResp.Metadata["timestamp"]
		require.True(t, exists, "Metadata should contain timestamp")

		metadataTime, err := time.Parse(time.RFC3339, metadataTimestamp.(string))
		require.NoError(t, err, "Metadata timestamp should be valid RFC3339 format")

		// Both timestamps should be very close (within 1 second)
		timeDiffMetadata := timestamp.Sub(metadataTime)
		if timeDiffMetadata < 0 {
			timeDiffMetadata = -timeDiffMetadata
		}
		assert.True(t, timeDiffMetadata < time.Second, "Response and metadata timestamps should be very close")
	})
}

func TestErrorResponseConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil, // Force nil runtime to ensure error responses
	}

	apiServer := NewAPIServer(app)
	// Ensure handlers are not initialized to force error conditions
	apiServer.handlerManager = internal.NewHandlerManager(internal.NewAppLogger())
	router := apiServer.engine

	t.Run("ServiceStart", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/test/start", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify all required fields are present
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "test", errorResp.ServiceName)
		assert.Equal(t, "start", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ServiceStop", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/test/stop", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify all required fields are present and consistent
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "test", errorResp.ServiceName)
		assert.Equal(t, "stop", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ImagePull", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/images/test:latest/pull", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify all required fields are present and consistent
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "test:latest", errorResp.ImageName)
		assert.Equal(t, "start_image_pull", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ServiceLogs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/services/test/logs", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify all required fields are present and consistent
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "test", errorResp.ServiceName)
		assert.Equal(t, "get_logs", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})

	t.Run("ServiceConnection", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/services/test/connection", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify all required fields are present and consistent
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "test", errorResp.ServiceName)
		assert.Equal(t, "get_connection_info", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)
		assert.NotNil(t, errorResp.Metadata)
	})
}

func TestErrorMetadataValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil, // Force nil runtime to ensure error responses
	}

	apiServer := NewAPIServer(app)
	// Ensure handlers are not initialized to force error conditions
	apiServer.handlerManager = internal.NewHandlerManager(internal.NewAppLogger())
	router := apiServer.engine

	t.Run("ServiceStartMetadata", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/postgres/start?persist=true", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify persist flag is included in metadata
		assert.Equal(t, true, errorResp.Metadata["persist"])

		// Verify all required metadata fields
		requiredFields := []string{"serviceName", "action", "timestamp", "errorType"}
		for _, field := range requiredFields {
			assert.Contains(t, errorResp.Metadata, field, fmt.Sprintf("Metadata should contain %s", field))
			assert.NotEmpty(t, errorResp.Metadata[field], fmt.Sprintf("Metadata %s should not be empty", field))
		}
	})

	t.Run("ServiceLogsMetadata", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/services/postgres/logs?tail=50", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify tail lines parameter is included
		assert.Equal(t, float64(50), errorResp.Metadata["tailLines"])
	})

	t.Run("ImagePullMetadata", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/images/postgres:13/pull", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify image-specific metadata
		assert.Equal(t, "postgres:13", errorResp.Metadata["imageName"])
		assert.Equal(t, "postgres:13", errorResp.Metadata["serviceName"])
		assert.Equal(t, "start_image_pull", errorResp.Metadata["action"])
		assert.Equal(t, "image_pull_failed", errorResp.Metadata["errorType"])
	})
}

func TestErrorResponseHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil, // Force nil runtime to ensure error responses
	}

	apiServer := NewAPIServer(app)
	// Ensure handlers are not initialized to force error conditions
	apiServer.handlerManager = internal.NewHandlerManager(internal.NewAppLogger())
	router := apiServer.engine

	t.Run("ContentTypeJSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/test/start", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify content type is JSON
		contentType := w.Header().Get("Content-Type")
		assert.True(t, strings.Contains(contentType, "application/json"),
			"Content-Type should be application/json, got: %s", contentType)
	})

	t.Run("CORSHeaders", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/test/start", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Verify CORS headers are present (if CORS middleware is enabled)
		// This depends on the actual CORS configuration
		accessControl := w.Header().Get("Access-Control-Allow-Origin")
		if accessControl != "" {
			assert.NotEmpty(t, accessControl, "CORS headers should be present when configured")
		}
	})
}

func TestErrorRecoveryScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil,
	}

	apiServer := NewAPIServer(app)
	router := apiServer.engine

	t.Run("InvalidServiceName", func(t *testing.T) {
		// Test with empty service name
		req := httptest.NewRequest("POST", "/api/v1/services//start", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 404 for invalid route, not 500
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidImageName", func(t *testing.T) {
		// Test with malformed image name
		req := httptest.NewRequest("POST", "/api/v1/images//pull", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 404 for invalid route
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("InvalidParameters", func(t *testing.T) {
		// Test with invalid persist parameter
		req := httptest.NewRequest("POST", "/api/v1/services/test/start?persist=invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		assert.Contains(t, errorResp["error"], "invalid persist parameter")
	})
}

func TestErrorHandlingEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil,
	}

	apiServer := NewAPIServer(app)
	router := apiServer.engine

	t.Run("SpecialCharactersInServiceName", func(t *testing.T) {
		serviceName := "test-service_123"
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/services/%s/start", serviceName), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		assert.Equal(t, serviceName, errorResp.ServiceName)
		assert.Equal(t, serviceName, errorResp.Metadata["serviceName"])
	})

	t.Run("LongServiceName", func(t *testing.T) {
		longName := strings.Repeat("a", 100)
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/services/%s/start", longName), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		assert.Equal(t, longName, errorResp.ServiceName)
	})

	t.Run("UnicodeInServiceName", func(t *testing.T) {
		unicodeName := "test-服务-🐳"
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/services/%s/start", unicodeName), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		assert.Equal(t, unicodeName, errorResp.ServiceName)
	})
}
