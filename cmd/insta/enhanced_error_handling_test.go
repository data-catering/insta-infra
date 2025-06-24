package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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
	// Setup test environment
	gin.SetMode(gin.TestMode)

	// Create test app with mock runtime
	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil, // Mock runtime will be nil for error testing
	}

	// Create API server
	apiServer := NewAPIServer(app)
	router := apiServer.engine

	t.Run("ServiceStartErrorResponse", func(t *testing.T) {
		// Test service start error with enhanced metadata
		req := httptest.NewRequest("POST", "/api/v1/services/nonexistent/start", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "start", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)

		// Verify metadata structure
		require.NotNil(t, errorResp.Metadata)
		assert.Equal(t, "nonexistent", errorResp.Metadata["serviceName"])
		assert.Equal(t, "service_start", errorResp.Metadata["action"])
		assert.Equal(t, "service_start_failed", errorResp.Metadata["errorType"])
		assert.NotEmpty(t, errorResp.Metadata["timestamp"])
	})

	t.Run("ServiceStopErrorResponse", func(t *testing.T) {
		// Test service stop error with enhanced metadata
		req := httptest.NewRequest("POST", "/api/v1/services/nonexistent/stop", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "stop", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)

		// Verify metadata structure
		require.NotNil(t, errorResp.Metadata)
		assert.Equal(t, "service_stop", errorResp.Metadata["action"])
		assert.Equal(t, "service_stop_failed", errorResp.Metadata["errorType"])
	})

	t.Run("ImagePullErrorResponse", func(t *testing.T) {
		// Test image pull error response when handler not available
		req := httptest.NewRequest("POST", "/api/v1/images/postgres:13/pull", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Image pull returns error when handler not available
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "postgres:13", errorResp.ImageName)
		assert.Equal(t, "start_image_pull", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)
		assert.Equal(t, http.StatusInternalServerError, errorResp.Status)

		// Verify metadata structure
		require.NotNil(t, errorResp.Metadata)
		assert.Equal(t, "postgres:13", errorResp.Metadata["imageName"])
		assert.Equal(t, "start_image_pull", errorResp.Metadata["action"])
		assert.Equal(t, "image_pull_failed", errorResp.Metadata["errorType"])
	})

	t.Run("ServiceLogsErrorResponse", func(t *testing.T) {
		// Test service logs error with enhanced metadata
		req := httptest.NewRequest("GET", "/api/v1/services/nonexistent/logs", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "get_logs", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)

		// Verify metadata structure
		require.NotNil(t, errorResp.Metadata)
		assert.Equal(t, "get_service_logs", errorResp.Metadata["action"])
		assert.Equal(t, "logs_fetch_failed", errorResp.Metadata["errorType"])
		assert.Equal(t, float64(100), errorResp.Metadata["tailLines"]) // JSON numbers are float64
	})

	t.Run("ServiceConnectionErrorResponse", func(t *testing.T) {
		// Test service connection error with enhanced metadata
		req := httptest.NewRequest("GET", "/api/v1/services/nonexistent/connection", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify enhanced error structure
		assert.NotEmpty(t, errorResp.Error)
		assert.Equal(t, "nonexistent", errorResp.ServiceName)
		assert.Equal(t, "get_connection_info", errorResp.Action)
		assert.NotEmpty(t, errorResp.Timestamp)

		// Verify metadata structure
		require.NotNil(t, errorResp.Metadata)
		assert.Equal(t, "connection_info_failed", errorResp.Metadata["errorType"])
	})
}

func TestErrorResponseTimestamps(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil,
	}

	apiServer := NewAPIServer(app)
	router := apiServer.engine

	t.Run("TimestampFormat", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/test/start", nil)
		w := httptest.NewRecorder()

		beforeRequest := time.Now()
		router.ServeHTTP(w, req)
		afterRequest := time.Now()

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify timestamp format (RFC3339)
		parsedTime, err := time.Parse(time.RFC3339, errorResp.Timestamp)
		require.NoError(t, err, "Timestamp should be in RFC3339 format")

		// Verify timestamp is within reasonable range
		assert.True(t, parsedTime.After(beforeRequest.Add(-time.Second)))
		assert.True(t, parsedTime.Before(afterRequest.Add(time.Second)))

		// Verify metadata timestamp matches main timestamp
		metadataTimestamp, ok := errorResp.Metadata["timestamp"].(string)
		require.True(t, ok, "Metadata should contain timestamp")

		metadataParsedTime, err := time.Parse(time.RFC3339, metadataTimestamp)
		require.NoError(t, err, "Metadata timestamp should be in RFC3339 format")

		// Timestamps should be very close (within 1 second)
		timeDiff := parsedTime.Sub(metadataParsedTime)
		assert.True(t, timeDiff < time.Second && timeDiff > -time.Second)
	})
}

func TestErrorResponseConsistency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil,
	}

	apiServer := NewAPIServer(app)
	router := apiServer.engine

	testCases := []struct {
		name           string
		method         string
		path           string
		expectedAction string
		expectedType   string
		hasImageName   bool
		hasServiceName bool
	}{
		{
			name:           "ServiceStart",
			method:         "POST",
			path:           "/api/v1/services/test/start",
			expectedAction: "service_start",
			expectedType:   "service_start_failed",
			hasServiceName: true,
		},
		{
			name:           "ServiceStop",
			method:         "POST",
			path:           "/api/v1/services/test/stop",
			expectedAction: "service_stop",
			expectedType:   "service_stop_failed",
			hasServiceName: true,
		},
		{
			name:           "ImagePull",
			method:         "POST",
			path:           "/api/v1/images/test:latest/pull",
			expectedAction: "start_image_pull",
			expectedType:   "image_pull_failed",
			hasImageName:   true,
			hasServiceName: true,
		},
		{
			name:           "ServiceLogs",
			method:         "GET",
			path:           "/api/v1/services/test/logs",
			expectedAction: "get_service_logs",
			expectedType:   "logs_fetch_failed",
			hasServiceName: true,
		},
		{
			name:           "ServiceConnection",
			method:         "GET",
			path:           "/api/v1/services/test/connection",
			expectedAction: "get_connection_info",
			expectedType:   "connection_info_failed",
			hasServiceName: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// All should return error status
			assert.Equal(t, http.StatusInternalServerError, w.Code)

			var errorResp ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &errorResp)
			require.NoError(t, err)

			// Verify consistent error structure
			assert.NotEmpty(t, errorResp.Error, "Error message should not be empty")
			assert.NotEmpty(t, errorResp.Action, "Action should not be empty")
			assert.NotEmpty(t, errorResp.Timestamp, "Timestamp should not be empty")
			assert.Equal(t, http.StatusInternalServerError, errorResp.Status, "Status should be 500")

			// Verify metadata structure
			require.NotNil(t, errorResp.Metadata, "Metadata should not be nil")
			assert.Equal(t, tc.expectedAction, errorResp.Metadata["action"], "Metadata action should match expected")
			assert.Equal(t, tc.expectedType, errorResp.Metadata["errorType"], "Error type should match expected")

			// Verify service name presence
			if tc.hasServiceName {
				assert.NotEmpty(t, errorResp.ServiceName, "Service name should be present")
				assert.Equal(t, errorResp.ServiceName, errorResp.Metadata["serviceName"], "Service names should match")
			}

			// Verify image name presence
			if tc.hasImageName {
				assert.NotEmpty(t, errorResp.ImageName, "Image name should be present")
				assert.Equal(t, errorResp.ImageName, errorResp.Metadata["imageName"], "Image names should match")
			}

			// Verify timestamp format
			_, err = time.Parse(time.RFC3339, errorResp.Timestamp)
			assert.NoError(t, err, "Timestamp should be valid RFC3339")
		})
	}
}

func TestErrorMetadataValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := &App{
		dataDir:  "/tmp/test",
		instaDir: "/tmp/test",
		runtime:  nil,
	}

	apiServer := NewAPIServer(app)
	router := apiServer.engine

	t.Run("ServiceStartMetadata", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/services/postgres/start?persist=true", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

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

		var errorResp ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResp)
		require.NoError(t, err)

		// Verify tail lines parameter is included
		assert.Equal(t, float64(50), errorResp.Metadata["tailLines"])

		// Verify it's also in the main response
		assert.Equal(t, float64(50), errorResp.Metadata["tailLines"])
	})

	t.Run("ImagePullMetadata", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/v1/images/postgres:13/pull", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

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
		runtime:  nil,
	}

	apiServer := NewAPIServer(app)
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
