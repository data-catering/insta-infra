package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/data-catering/insta-infra/v2/cmd/insta/internal"
	"github.com/data-catering/insta-infra/v2/internal/core/container"
	"github.com/gin-gonic/gin"
)

//go:embed ui/dist/*
var uiFiles embed.FS

// WebSocket message types for real-time communication
const (
	// Service-related events
	WSMsgServiceStatusUpdate = "service_status_update"
	WSMsgServiceStarted      = "service_started"
	WSMsgServiceStopped      = "service_stopped"
	WSMsgServiceError        = "service_error"

	// Logging events
	WSMsgServiceLogs = "service_logs"
	WSMsgAppLogs     = "app_logs"
	WSMsgLogStream   = "log_stream"

	// Image/Container events
	WSMsgImagePullProgress = "image_pull_progress"
	WSMsgImagePullComplete = "image_pull_complete"
	WSMsgImagePullError    = "image_pull_error"

	// Runtime events
	WSMsgRuntimeStatusUpdate = "runtime_status_update"

	// Connection events
	WSMsgClientConnected    = "client_connected"
	WSMsgClientDisconnected = "client_disconnected"

	// Keep-alive
	WSMsgPing = "ping"
	WSMsgPong = "pong"
)

// WebSocket types are now defined in websocket.go

// APIServer represents the HTTP API server for insta-infra
type APIServer struct {
	app            *App
	engine         *gin.Engine
	port           int
	handlerManager *internal.HandlerManager
	logger         *internal.AppLogger
	ctx            context.Context

	// WebSocket broadcaster
	wsBroadcaster *WebSocketBroadcaster
}

// APIInfo represents the API information endpoint response
type APIInfo struct {
	Version    string    `json:"version"`
	APIVersion string    `json:"api_version"`
	BuildTime  string    `json:"build_time"`
	StartTime  time.Time `json:"start_time"`
	Features   []string  `json:"features"`
}

// HealthStatus represents the health check endpoint response
type HealthStatus struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Runtime   string    `json:"runtime,omitempty"`
}

// NewAPIServer creates a new API server instance
func NewAPIServer(app *App) *APIServer {
	// Set gin to release mode for production-like behavior
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()

	// Add middleware
	engine.Use(gin.Recovery())
	engine.Use(corsMiddleware())
	engine.Use(errorHandlingMiddleware())

	// Initialize logger and handler manager
	logger := internal.NewAppLogger()
	handlerManager := internal.NewHandlerManager(logger)
	ctx := context.Background()

	// Add request logging middleware with our logger
	engine.Use(requestLoggingMiddleware(logger))

	// Create WebSocket broadcaster first
	wsBroadcaster := NewWebSocketBroadcaster(logger)

	server := &APIServer{
		app:            app,
		engine:         engine,
		handlerManager: handlerManager,
		logger:         logger,
		ctx:            ctx,
		wsBroadcaster:  wsBroadcaster,
	}

	// Initialize handlers with progress callback
	progressCallback := func(serviceName, imageName string, progress float64, status string) {
		server.BroadcastImagePullProgress(serviceName, imageName, progress, status)
	}

	if err := handlerManager.InitializeWithCallback(app.instaDir, ctx, progressCallback); err != nil {
		logger.Log(fmt.Sprintf("Warning: Failed to initialize handlers: %v", err))
	} else {
		logger.Log("Handlers initialized successfully for HTTP API")
	}

	// Setup routes
	server.setupRoutes()

	// Start WebSocket broadcaster
	server.wsBroadcaster.Start()

	// Start service status monitoring
	server.startServiceStatusMonitor()

	return server
}

// corsMiddleware provides enhanced CORS headers for browser access
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// errorHandlingMiddleware provides consistent error response formatting
func errorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			switch err.Type {
			case gin.ErrorTypeBind:
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Invalid request format",
					"details": err.Error(),
				})
			case gin.ErrorTypePublic:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal server error",
				})
			}
		}
	}
}

// requestLoggingMiddleware provides enhanced request logging
func requestLoggingMiddleware(logger *internal.AppLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		clientIP := c.ClientIP()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Log(fmt.Sprintf("%s %s %d %v %s", method, path, status, duration, clientIP))
	}
}

// setupRoutes configures all API routes with versioning
func (s *APIServer) setupRoutes() {
	// API version 1 routes group
	v1 := s.engine.Group("/api/v1")

	// Health and info endpoints
	v1.GET("/health", s.healthCheck)
	v1.GET("/info", s.apiInfo)

	// Service Management endpoints
	services := v1.Group("/services")
	{
		services.GET("", s.listServices)
		services.GET("/:name/status", s.getServiceStatus)
		services.POST("/:name/start", s.startService)
		services.POST("/:name/stop", s.stopService)
		services.GET("/:name/connection", s.getServiceConnection)
		services.GET("/:name/connection/enhanced", s.getEnhancedServiceConnection)
		services.POST("/:name/open", s.openServiceConnection)
		services.GET("/:name/logs", s.getServiceLogs)
	}

	// Bulk service operations
	v1.GET("/services/all/status", s.getAllServiceStatuses)
	v1.GET("/services/all/details", s.getAllServiceDetails)
	v1.GET("/services/running", s.getRunningServices)
	v1.POST("/services/all/stop", s.stopAllServices)
	v1.POST("/services/status/refresh", s.refreshServiceStatus)

	// Image Management endpoints
	images := v1.Group("/images")
	{
		images.GET("/all/status", s.getAllImageStatuses)
		images.GET("/:name/exists", s.checkImageExists)
		images.POST("/:name/pull", s.startImagePull)
		images.DELETE("/:name/pull", s.stopImagePull)
		images.GET("/:name/progress", s.getImagePullProgress)
		images.GET("/:name/info", s.getImageInfo)
	}

	// Logging endpoints
	logs := v1.Group("/logs")
	{
		logs.GET("/app", s.getAppLogs)
		logs.GET("/app/entries", s.getAppLogEntries)
		logs.GET("/app/since/:timestamp", s.getAppLogsSince)
	}

	// Runtime Management endpoints
	runtime := v1.Group("/runtime")
	{
		runtime.GET("/status", s.getRuntimeStatus)
		runtime.GET("/current", s.getCurrentRuntime)
		runtime.POST("/start", s.startRuntime)
		runtime.POST("/reinitialize", s.reinitializeRuntime)
		runtime.PUT("/docker-path", s.setCustomDockerPath)
		runtime.PUT("/podman-path", s.setCustomPodmanPath)
		runtime.GET("/docker-path", s.getCustomDockerPath)
		runtime.GET("/podman-path", s.getCustomPodmanPath)
	}

	// WebSocket endpoint for real-time updates
	v1.GET("/ws", s.handleWebSocket)

	// Serve embedded frontend files
	s.setupStaticFileServing()
}

// setupStaticFileServing configures embedded frontend file serving
func (s *APIServer) setupStaticFileServing() {
	// Get the embedded filesystem
	distFS, err := fs.Sub(uiFiles, "ui/dist")
	if err != nil {
		s.logger.Log(fmt.Sprintf("Warning: Failed to setup embedded filesystem: %v", err))
		return
	}

	// Serve static assets (CSS, JS, fonts, etc.)
	assetsFS, err := fs.Sub(distFS, "assets")
	if err != nil {
		s.logger.Log(fmt.Sprintf("Warning: Failed to setup assets filesystem: %v", err))
	} else {
		s.engine.StaticFS("/assets", http.FS(assetsFS))
	}

	// Serve favicon and manifest files explicitly
	s.engine.GET("/favicon.svg", func(c *gin.Context) {
		content, err := fs.ReadFile(distFS, "favicon.svg")
		if err != nil {
			c.String(http.StatusNotFound, "Favicon not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", content)
	})

	s.engine.GET("/favicon-16x16.png", func(c *gin.Context) {
		content, err := fs.ReadFile(distFS, "favicon-16x16.png")
		if err != nil {
			c.String(http.StatusNotFound, "Favicon not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", content) // Currently SVG format
	})

	s.engine.GET("/favicon-32x32.png", func(c *gin.Context) {
		content, err := fs.ReadFile(distFS, "favicon-32x32.png")
		if err != nil {
			c.String(http.StatusNotFound, "Favicon not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", content) // Currently SVG format
	})

	s.engine.GET("/apple-touch-icon.png", func(c *gin.Context) {
		content, err := fs.ReadFile(distFS, "apple-touch-icon.png")
		if err != nil {
			c.String(http.StatusNotFound, "Apple touch icon not found")
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", content) // Currently SVG format
	})

	s.engine.GET("/site.webmanifest", func(c *gin.Context) {
		content, err := fs.ReadFile(distFS, "site.webmanifest")
		if err != nil {
			c.String(http.StatusNotFound, "Web manifest not found")
			return
		}
		c.Data(http.StatusOK, "application/manifest+json", content)
	})

	// Serve index.html at root
	s.engine.GET("/", func(c *gin.Context) {
		indexFile, err := distFS.Open("index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to load frontend")
			return
		}
		defer indexFile.Close()

		indexContent, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read frontend")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
	})

	// Handle SPA routing - serve index.html for all unmatched routes
	s.engine.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		indexContent, err := fs.ReadFile(distFS, "index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to read frontend")
			return
		}

		c.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
	})
}

// findAvailablePort finds an available port in the safe range 9310-9320
func (s *APIServer) findAvailablePort() (int, error) {
	for port := 9310; port <= 9320; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available ports in range 9310-9320")
}

// Start starts the HTTP server on an available port
func (s *APIServer) Start() error {
	port, err := s.findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	s.port = port
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Starting insta-infra web server on http://localhost:%d\n", port)
	return s.engine.Run(addr)
}

// StartOnPort starts the HTTP server on a specific port
func (s *APIServer) StartOnPort(port int) error {
	s.port = port
	addr := fmt.Sprintf(":%d", port)

	fmt.Printf("Starting insta-infra web server on http://localhost:%d\n", port)
	return s.engine.Run(addr)
}

// GetPort returns the current server port
func (s *APIServer) GetPort() int {
	return s.port
}

// StartWithBrowser starts the HTTP server and opens the browser
func (s *APIServer) StartWithBrowser() error {
	port, err := s.findAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	s.port = port
	url := fmt.Sprintf("http://localhost:%d", port)

	// Open browser in a goroutine so it doesn't block server startup
	go func() {
		time.Sleep(1 * time.Second) // Give server time to start
		if err := openBrowser(url); err != nil {
			s.logger.Log(fmt.Sprintf("Failed to open browser: %v", err))
			s.logger.Log(fmt.Sprintf("Please manually open: %s", url))
		} else {
			s.logger.Log(fmt.Sprintf("Opened browser to: %s", url))
		}
	}()

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Starting insta-infra web server on %s\n", url)
	return s.engine.Run(addr)
}

// openBrowser opens the default browser to the specified URL
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	default:
		return fmt.Errorf("unsupported platform")
	}

	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

// ==================== ENDPOINT HANDLERS ====================

// healthCheck provides a health check endpoint
func (s *APIServer) healthCheck(c *gin.Context) {
	status := HealthStatus{
		Status:    "ok",
		Timestamp: time.Now(),
		Runtime:   s.app.runtime.Name(),
	}
	c.JSON(http.StatusOK, status)
}

// apiInfo provides API version and feature information
func (s *APIServer) apiInfo(c *gin.Context) {
	info := APIInfo{
		Version:    version,
		APIVersion: "v1",
		BuildTime:  buildTime,
		StartTime:  time.Now(), // TODO: Store actual start time
		Features: []string{
			"service_management",
			"image_management",
			"logging",
			"runtime_management",
			"real_time_updates",
		},
	}
	c.JSON(http.StatusOK, info)
}

// Service Management Endpoints Implementation

func (s *APIServer) listServices(c *gin.Context) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Use enhanced services instead of legacy services
	services := serviceHandler.ListEnhancedServices()
	c.JSON(http.StatusOK, services)
}

func (s *APIServer) getServiceStatus(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	status, err := serviceHandler.GetServiceStatus(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"status":       status,
	})
}

func (s *APIServer) startService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	// Parse persist flag from query parameters
	persistStr := c.DefaultQuery("persist", "false")
	persist, err := strconv.ParseBool(persistStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid persist parameter, must be true or false"})
		return
	}

	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Broadcast that service is starting
	s.BroadcastServiceStatusUpdate(serviceName, "starting")

	err = serviceHandler.StartService(serviceName, persist)
	if err != nil {
		// Broadcast error status
		s.BroadcastServiceStatusUpdate(serviceName, "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Don't broadcast status here - let the status monitor detect and broadcast the actual status
	// This prevents conflicting updates and the "flash" effect in the UI

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"action":       "start",
		"persist":      persist,
		"status":       "starting",
	})
}

func (s *APIServer) stopService(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Broadcast that service is stopping
	s.BroadcastServiceStatusUpdate(serviceName, "stopping")

	err := serviceHandler.StopService(serviceName)
	if err != nil {
		// Broadcast error status
		s.BroadcastServiceStatusUpdate(serviceName, "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast successful stop initiation
	s.BroadcastServiceStatusUpdate(serviceName, "stopped")

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"action":       "stop",
		"status":       "stopping",
	})
}

func (s *APIServer) getServiceConnection(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	connectionHandler := s.handlerManager.GetConnectionHandler()
	if connectionHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "connection handler not available"})
		return
	}

	connection, err := connectionHandler.GetServiceConnectionInfo(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"connection":   connection,
	})
}

func (s *APIServer) getEnhancedServiceConnection(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	connectionHandler := s.handlerManager.GetConnectionHandler()
	if connectionHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "connection handler not available"})
		return
	}

	enhancedConnection, err := connectionHandler.GetEnhancedServiceConnectionInfo(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"connection":   enhancedConnection,
	})
}

func (s *APIServer) openServiceConnection(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	connectionHandler := s.handlerManager.GetConnectionHandler()
	if connectionHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "connection handler not available"})
		return
	}

	connectionInfo, err := connectionHandler.OpenServiceInBrowser(serviceName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"action":       "open_connection",
		"status":       "success",
		"connection":   connectionInfo,
	})
}

func (s *APIServer) getServiceLogs(c *gin.Context) {
	serviceName := c.Param("name")
	if serviceName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service name is required"})
		return
	}

	// Parse optional query parameters
	tailLinesStr := c.DefaultQuery("tail", "100")
	tailLines := 100
	if parsed, err := strconv.Atoi(tailLinesStr); err == nil && parsed > 0 {
		tailLines = parsed
	}

	logsHandler := s.handlerManager.GetLogsHandler()
	if logsHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "logs handler not available"})
		return
	}

	logs, err := logsHandler.GetServiceLogs(serviceName, tailLines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_name": serviceName,
		"logs":         logs,
		"tail_lines":   tailLines,
	})
}

func (s *APIServer) getAllServiceStatuses(c *gin.Context) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Get all services with their current statuses
	services, err := serviceHandler.GetAllServicesWithStatusAndDependencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to map of service statuses
	statuses := make(map[string]interface{})
	for _, service := range services {
		statuses[service.Name] = map[string]interface{}{
			"Status": service.Status,
			"status": service.Status, // Provide both capitalized and lowercase for compatibility
		}
	}

	c.JSON(http.StatusOK, statuses)
}

func (s *APIServer) getAllServiceDetails(c *gin.Context) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	details, err := serviceHandler.GetAllServicesWithStatusAndDependencies()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_details": details,
		"total_services":  len(details),
	})
}

func (s *APIServer) getRunningServices(c *gin.Context) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Get both main services and dependency containers
	mainServices, err := serviceHandler.GetAllRunningServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dependencyServices, err := serviceHandler.GetAllDependencyStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Merge both main services and dependency containers
	runningServices := make(map[string]interface{})
	runningCount := 0

	// Helper function to check if a service is considered "running"
	isRunning := func(status string) bool {
		return strings.Contains(status, "running")
	}

	// Add running main services
	for containerName, service := range mainServices {
		if isRunning(service.Status) {
			runningServices[containerName] = service
			runningCount++
		}
	}

	// Add running dependency containers
	for containerName, service := range dependencyServices {
		// Avoid duplicates (dependency containers that are also main services)
		if _, exists := runningServices[containerName]; !exists && isRunning(service.Status) {
			runningServices[containerName] = service
			runningCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"running_services": runningServices,
		"total_running":    runningCount,
	})
}

func (s *APIServer) stopAllServices(c *gin.Context) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Broadcast that all services are stopping
	s.BroadcastServiceStatusUpdate("all", "stopping")

	err := serviceHandler.StopAllServices()
	if err != nil {
		// Broadcast error status
		s.BroadcastServiceStatusUpdate("all", "failed", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast successful stop initiation
	s.BroadcastServiceStatusUpdate("all", "stopped")

	c.JSON(http.StatusOK, gin.H{
		"action": "stop_all",
		"status": "stopping",
	})
}

func (s *APIServer) refreshServiceStatus(c *gin.Context) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service handler not available"})
		return
	}

	// Get both main services and dependency containers
	mainServices, err := serviceHandler.GetAllRunningServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	dependencyServices, err := serviceHandler.GetAllDependencyStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Merge all statuses for broadcasting
	allStatuses := make(map[string]interface{})

	// Add main services
	for containerName, service := range mainServices {
		allStatuses[containerName] = service
		s.BroadcastServiceStatusUpdate(containerName, service.Status)
	}

	// Add dependency containers (avoid duplicates)
	for containerName, service := range dependencyServices {
		if _, exists := allStatuses[containerName]; !exists {
			allStatuses[containerName] = service
			s.BroadcastServiceStatusUpdate(containerName, service.Status)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"action":   "refresh_status",
		"statuses": allStatuses,
	})
}

func (s *APIServer) checkImageExists(c *gin.Context) {
	imageName := c.Param("name")
	if imageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image name is required"})
		return
	}

	imageHandler := s.handlerManager.GetImageHandler()
	if imageHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image handler not available"})
		return
	}

	exists, err := imageHandler.CheckImageExists(imageName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, exists)
}

func (s *APIServer) getAllImageStatuses(c *gin.Context) {
	imageHandler := s.handlerManager.GetImageHandler()
	if imageHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image handler not available"})
		return
	}

	// Use the image handler's proper method that handles environment variable resolution
	imageStatuses, err := imageHandler.GetAllImageStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get image statuses: %v", err)})
		return
	}

	// Convert to the expected API format
	apiImageStatuses := make(map[string]map[string]interface{})
	for serviceName, status := range imageStatuses {
		apiStatus := "missing"
		if status.Status == "available" {
			apiStatus = "ready"
		} else if status.Status == "error" {
			apiStatus = "error"
		}

		apiImageStatuses[serviceName] = map[string]interface{}{
			"status":     apiStatus,
			"exists":     status.Status == "available",
			"image_name": serviceName, // Keep service name for API compatibility
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"image_statuses": apiImageStatuses,
		"total_services": len(imageStatuses),
	})
}

func (s *APIServer) startImagePull(c *gin.Context) {
	imageName := c.Param("name")
	if imageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image name is required"})
		return
	}

	imageHandler := s.handlerManager.GetImageHandler()
	if imageHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image handler not available"})
		return
	}

	// Broadcast that image pull is starting
	s.BroadcastImagePullProgress(imageName, imageName, 0, "starting")

	err := imageHandler.StartImagePull(imageName)
	if err != nil {
		// Broadcast error status
		s.BroadcastImagePullProgress(imageName, imageName, 0, "error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast that image pull has been initiated
	s.BroadcastImagePullProgress(imageName, imageName, 0, "downloading")

	c.JSON(http.StatusOK, gin.H{
		"image_name": imageName,
		"action":     "pull",
		"status":     "downloading",
	})
}

func (s *APIServer) stopImagePull(c *gin.Context) {
	imageName := c.Param("name")
	if imageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image name is required"})
		return
	}

	imageHandler := s.handlerManager.GetImageHandler()
	if imageHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image handler not available"})
		return
	}

	// Broadcast that image pull is being cancelled
	s.BroadcastImagePullProgress(imageName, imageName, 0, "cancelling")

	err := imageHandler.StopImagePull(imageName)
	if err != nil {
		// Broadcast error status
		s.BroadcastImagePullProgress(imageName, imageName, 0, "error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast successful cancellation
	s.BroadcastImagePullProgress(imageName, imageName, 0, "cancelled")

	c.JSON(http.StatusOK, gin.H{
		"image_name": imageName,
		"action":     "stop_pull",
		"status":     "cancelled",
	})
}

func (s *APIServer) getImagePullProgress(c *gin.Context) {
	imageName := c.Param("name")
	if imageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image name is required"})
		return
	}

	imageHandler := s.handlerManager.GetImageHandler()
	if imageHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image handler not available"})
		return
	}

	progress, err := imageHandler.GetImagePullProgress(imageName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_name": imageName,
		"progress":   progress,
	})
}

func (s *APIServer) getImageInfo(c *gin.Context) {
	imageName := c.Param("name")
	if imageName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image name is required"})
		return
	}

	imageHandler := s.handlerManager.GetImageHandler()
	if imageHandler == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image handler not available"})
		return
	}

	imageInfo, err := imageHandler.GetImageInfo(imageName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"image_name": imageName,
		"info":       imageInfo,
	})
}

func (s *APIServer) getAppLogs(c *gin.Context) {
	logs := s.logger.GetLogs()
	c.JSON(http.StatusOK, gin.H{
		"logs":       logs,
		"total_logs": len(logs),
	})
}

func (s *APIServer) getAppLogEntries(c *gin.Context) {
	entries := s.logger.GetLogEntries()
	c.JSON(http.StatusOK, gin.H{
		"log_entries": entries,
		"total_logs":  len(entries),
	})
}

func (s *APIServer) getAppLogsSince(c *gin.Context) {
	timestampStr := c.Param("timestamp")
	if timestampStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "timestamp is required"})
		return
	}

	// Parse timestamp
	since, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid timestamp format: %v", err)})
		return
	}

	entries := s.logger.GetLogsSince(since)
	c.JSON(http.StatusOK, gin.H{
		"log_entries": entries,
		"total_logs":  len(entries),
		"since":       timestampStr,
	})
}

func (s *APIServer) getRuntimeStatus(c *gin.Context) {
	// Use the comprehensive runtime status check
	detailedStatus := container.GetDetailedRuntimeStatus()

	// Return the detailed status - this works even if no runtime is currently selected
	c.JSON(http.StatusOK, detailedStatus)
}

func (s *APIServer) getCurrentRuntime(c *gin.Context) {
	if s.app.runtime == nil {
		c.JSON(http.StatusOK, gin.H{
			"runtime": nil,
			"name":    nil,
			"message": "No container runtime currently selected",
		})
		return
	}

	runtimeName := s.app.runtime.Name()
	c.JSON(http.StatusOK, gin.H{
		"runtime": runtimeName,
		"name":    runtimeName,
	})
}

func (s *APIServer) startRuntime(c *gin.Context) {
	var reqBody struct {
		Runtime string `json:"runtime"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if reqBody.Runtime == "" {
		if s.app.runtime != nil {
			reqBody.Runtime = s.app.runtime.Name() // Use current runtime if not specified and available
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "runtime must be specified when no runtime is currently selected"})
			return
		}
	}

	// Broadcast that runtime is starting
	s.BroadcastAppLogs(fmt.Sprintf("Starting %s runtime...", reqBody.Runtime), "info")

	// Use runtime manager to attempt starting the runtime
	runtimeManager := &container.RuntimeManager{}
	result := runtimeManager.AttemptStartRuntime(reqBody.Runtime)

	if result.Success {
		s.BroadcastAppLogs(fmt.Sprintf("Successfully started %s runtime", reqBody.Runtime), "info")
		c.JSON(http.StatusOK, gin.H{
			"runtime": reqBody.Runtime,
			"action":  "start",
			"status":  "success",
			"message": result.Message,
		})
	} else {
		s.BroadcastAppLogs(fmt.Sprintf("Failed to start %s runtime: %s", reqBody.Runtime, result.Error), "error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"runtime":                reqBody.Runtime,
			"action":                 "start",
			"status":                 "failed",
			"error":                  result.Error,
			"requires_manual_action": result.RequiresManualAction,
		})
	}
}

func (s *APIServer) reinitializeRuntime(c *gin.Context) {
	s.BroadcastAppLogs("Reinitializing container runtime...", "info")

	// Reinitialize the handler manager which will detect and setup the runtime again
	err := s.handlerManager.ReinitializeRuntime(s.app.instaDir, s.ctx)
	if err != nil {
		s.BroadcastAppLogs(fmt.Sprintf("Failed to reinitialize runtime: %v", err), "error")
		c.JSON(http.StatusInternalServerError, gin.H{
			"action": "reinitialize",
			"status": "failed",
			"error":  err.Error(),
		})
		return
	}

	runtime := s.handlerManager.GetContainerRuntime()
	if runtime == nil {
		s.BroadcastAppLogs("Reinitialization completed but no container runtime is available", "warning")
		c.JSON(http.StatusOK, gin.H{
			"action":  "reinitialize",
			"status":  "success",
			"runtime": nil,
			"message": "Reinitialized but no container runtime available",
		})
		return
	}

	s.BroadcastAppLogs("Successfully reinitialized container runtime", "info")
	c.JSON(http.StatusOK, gin.H{
		"action":  "reinitialize",
		"status":  "success",
		"runtime": runtime.Name(),
	})
}

func (s *APIServer) setCustomDockerPath(c *gin.Context) {
	var reqBody struct {
		Path string `json:"path"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if reqBody.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	// Set the environment variable for custom Docker path
	os.Setenv("INSTA_DOCKER_PATH", reqBody.Path)

	s.BroadcastAppLogs(fmt.Sprintf("Set custom Docker path to: %s", reqBody.Path), "info")

	c.JSON(http.StatusOK, gin.H{
		"action":      "set_docker_path",
		"docker_path": reqBody.Path,
		"status":      "success",
	})
}

func (s *APIServer) setCustomPodmanPath(c *gin.Context) {
	var reqBody struct {
		Path string `json:"path"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if reqBody.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "path is required"})
		return
	}

	// Set the environment variable for custom Podman path
	os.Setenv("INSTA_PODMAN_PATH", reqBody.Path)

	s.BroadcastAppLogs(fmt.Sprintf("Set custom Podman path to: %s", reqBody.Path), "info")

	c.JSON(http.StatusOK, gin.H{
		"action":      "set_podman_path",
		"podman_path": reqBody.Path,
		"status":      "success",
	})
}

func (s *APIServer) getCustomDockerPath(c *gin.Context) {
	dockerPath := os.Getenv("INSTA_DOCKER_PATH")

	c.JSON(http.StatusOK, gin.H{
		"docker_path": dockerPath,
		"is_custom":   dockerPath != "",
	})
}

func (s *APIServer) getCustomPodmanPath(c *gin.Context) {
	podmanPath := os.Getenv("INSTA_PODMAN_PATH")

	c.JSON(http.StatusOK, gin.H{
		"podman_path": podmanPath,
		"is_custom":   podmanPath != "",
	})
}

// WebSocket broadcasting is now handled by WebSocketBroadcaster

// startServiceStatusMonitor starts a background goroutine that monitors service status changes
// and broadcasts real-time updates to WebSocket clients
func (s *APIServer) startServiceStatusMonitor() {
	go func() {
		ticker := time.NewTicker(1 * time.Second) // Check every 1 second for faster test feedback
		defer ticker.Stop()

		// Keep track of last known statuses to detect changes
		lastKnownStatuses := make(map[string]string)

		for {
			select {
			case <-ticker.C:
				s.checkAndBroadcastStatusChanges(lastKnownStatuses)
			case <-s.ctx.Done():
				s.logger.Log("Service status monitor stopped")
				return
			}
		}
	}()

	s.logger.Log("Service status monitor started")
}

// checkAndBroadcastStatusChanges checks current service statuses and broadcasts changes
func (s *APIServer) checkAndBroadcastStatusChanges(lastKnownStatuses map[string]string) {
	serviceHandler := s.handlerManager.GetServiceHandler()
	if serviceHandler == nil {
		return
	}

	// Get current main service statuses
	mainServices, err := serviceHandler.GetAllRunningServices()
	if err != nil {
		s.logger.Log(fmt.Sprintf("Error getting main service statuses: %v", err))
		return
	}

	// Get current dependency container statuses
	dependencyServices, err := serviceHandler.GetAllDependencyStatuses()
	if err != nil {
		s.logger.Log(fmt.Sprintf("Error getting dependency statuses: %v", err))
		return
	}

	// Merge all statuses (main services + dependency containers)
	allCurrentStatuses := make(map[string]string)

	// Add main services
	for containerName, service := range mainServices {
		allCurrentStatuses[containerName] = service.Status
	}

	// Add dependency containers (avoid duplicates)
	for containerName, service := range dependencyServices {
		if _, exists := allCurrentStatuses[containerName]; !exists {
			allCurrentStatuses[containerName] = service.Status
		}
	}

	// Check for status changes
	for containerName, currentStatus := range allCurrentStatuses {
		lastStatus, existed := lastKnownStatuses[containerName]

		if !existed || lastStatus != currentStatus {
			// Status changed, broadcast update
			s.BroadcastServiceStatusUpdate(containerName, currentStatus)
			lastKnownStatuses[containerName] = currentStatus
		}
	}

	// Check for containers that are no longer running (stopped)
	// However, don't override transition states like "starting"
	for containerName, lastStatus := range lastKnownStatuses {
		if _, exists := allCurrentStatuses[containerName]; !exists && lastStatus != "stopped" {
			// Only broadcast stopped if the container wasn't in a starting state
			// This prevents the monitor from overriding "starting" status during service startup
			if lastStatus != "starting" {
				// Container is no longer running, broadcast stopped status
				s.BroadcastServiceStatusUpdate(containerName, "stopped")
				lastKnownStatuses[containerName] = "stopped"
			}
		}
	}
}

func (s *APIServer) handleWebSocket(c *gin.Context) {
	// Delegate to WebSocket broadcaster
	s.wsBroadcaster.HandleWebSocket(c.Writer, c.Request)
}

// handleWebSocketMessage is now handled by WebSocketBroadcaster

// BroadcastServiceStatusUpdate sends a service status update to all connected clients
func (s *APIServer) BroadcastServiceStatusUpdate(serviceName, status string, errorMsg ...string) {
	var errorString string
	if len(errorMsg) > 0 {
		errorString = errorMsg[0]
	}

	// Delegate to WebSocket broadcaster
	s.wsBroadcaster.BroadcastServiceStatusUpdate(serviceName, status, errorString)
}

// BroadcastServiceLogs sends service logs to all connected clients
func (s *APIServer) BroadcastServiceLogs(serviceName, logMessage string) {
	// Delegate to WebSocket broadcaster
	s.wsBroadcaster.BroadcastServiceLogs(serviceName, logMessage)
}

// BroadcastImagePullProgress sends image pull progress to all connected clients
func (s *APIServer) BroadcastImagePullProgress(serviceName, imageName string, progress float64, status string) {
	// Delegate to WebSocket broadcaster
	s.wsBroadcaster.BroadcastImagePullProgress(serviceName, imageName, progress, status)
}

// BroadcastAppLogs sends application logs to all connected clients
func (s *APIServer) BroadcastAppLogs(message string, level string) {
	// Delegate to WebSocket broadcaster
	s.wsBroadcaster.BroadcastAppLogs(message, level)
}
