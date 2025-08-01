package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/data-catering/insta-infra/v2/cmd/insta/models"
	"github.com/gorilla/websocket"
)

// Enhanced service WebSocket message types (new additions)
const (
	WSMsgEnhancedServiceUpdate   = "enhanced_service_update"
	WSMsgServiceDependencyUpdate = "service_dependency_update"
)

// WebSocketMessage represents a structured message sent over WebSocket
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp time.Time   `json:"timestamp"`
}

// ServiceStatusUpdate represents a service status change
type ServiceStatusUpdate struct {
	ServiceName string `json:"service_name"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	Message     string `json:"message,omitempty"`
}

// LogEntry represents a single log entry
type LogEntry struct {
	ServiceName string    `json:"service_name,omitempty"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level,omitempty"`
}

// ImagePullProgress represents image download progress
type ImagePullProgress struct {
	ImageName       string  `json:"image_name"`
	ServiceName     string  `json:"service_name,omitempty"`
	Progress        float64 `json:"progress"`
	Status          string  `json:"status"`
	BytesDownloaded int64   `json:"bytes_downloaded,omitempty"`
	TotalBytes      int64   `json:"total_bytes,omitempty"`
	Error           string  `json:"error,omitempty"`
}

// EnhancedServiceUpdate represents an enhanced service update with full data
type EnhancedServiceUpdate struct {
	ServiceName string                  `json:"service_name"`
	Service     *models.EnhancedService `json:"service"`
	UpdateType  string                  `json:"update_type"` // "status", "dependency", "config"
}

// WebSocketBroadcaster handles WebSocket connections and message broadcasting
type WebSocketBroadcaster struct {
	upgrader   websocket.Upgrader
	clients    map[*websocket.Conn]bool
	broadcast  chan WebSocketMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	logger     Logger
}

// Logger interface for WebSocket broadcaster
type Logger interface {
	Log(message string)
}

// NewWebSocketBroadcaster creates a new WebSocket broadcaster
func NewWebSocketBroadcaster(logger Logger) *WebSocketBroadcaster {
	ctx, cancel := context.WithCancel(context.Background())

	return &WebSocketBroadcaster{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for development
			},
		},
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WebSocketMessage, 256), // Buffered channel
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		ctx:        ctx,
		cancel:     cancel,
		logger:     logger,
	}
}

// Start begins the WebSocket broadcaster goroutines
func (wb *WebSocketBroadcaster) Start() {
	go wb.handleConnections()
	go wb.handleMessages()
	wb.logger.Log("WebSocket broadcaster started")
}

// Stop stops the WebSocket broadcaster
func (wb *WebSocketBroadcaster) Stop() {
	wb.cancel()
	close(wb.broadcast)
	close(wb.register)
	close(wb.unregister)
	wb.logger.Log("WebSocket broadcaster stopped")
}

// handleConnections manages client connections and disconnections
func (wb *WebSocketBroadcaster) handleConnections() {
	for {
		select {
		case client := <-wb.register:
			wb.mutex.Lock()
			wb.clients[client] = true
			wb.mutex.Unlock()
			wb.logger.Log("WebSocket client connected")

			// Send welcome message (using constant from webserver.go)
			welcomeMsg := WebSocketMessage{
				Type:      WSMsgClientConnected,
				Payload:   map[string]interface{}{"message": "Connected to insta-infra WebSocket"},
				Timestamp: time.Now(),
			}
			wb.sendToClient(client, welcomeMsg)

		case client := <-wb.unregister:
			wb.mutex.Lock()
			if _, ok := wb.clients[client]; ok {
				delete(wb.clients, client)
				client.Close()
			}
			wb.mutex.Unlock()
			wb.logger.Log("WebSocket client disconnected")

		case <-wb.ctx.Done():
			return
		}
	}
}

// handleMessages processes and broadcasts messages to all clients
func (wb *WebSocketBroadcaster) handleMessages() {
	for {
		select {
		case message := <-wb.broadcast:
			wb.broadcastToAllClients(message)

		case <-wb.ctx.Done():
			return
		}
	}
}

// HandleWebSocket handles HTTP upgrade to WebSocket
func (wb *WebSocketBroadcaster) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wb.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Register client
	wb.register <- conn

	// Handle client messages
	go wb.handleClientMessages(conn)
}

// handleClientMessages processes incoming messages from a client
func (wb *WebSocketBroadcaster) handleClientMessages(conn *websocket.Conn) {
	defer func() {
		wb.unregister <- conn
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle text messages
		if messageType == websocket.TextMessage {
			wb.processClientMessage(conn, message)
		}
	}
}

// processClientMessage processes a message from a client
func (wb *WebSocketBroadcaster) processClientMessage(conn *websocket.Conn, message []byte) {
	msgStr := string(message)

	// Handle simple ping/pong
	if msgStr == "ping" {
		pongMsg := WebSocketMessage{
			Type:      WSMsgPong,
			Payload:   "pong",
			Timestamp: time.Now(),
		}
		wb.sendToClient(conn, pongMsg)
		return
	}

	// Try to parse as structured message
	var incomingMsg WebSocketMessage
	if err := json.Unmarshal(message, &incomingMsg); err == nil {
		wb.handleStructuredMessage(conn, incomingMsg)
	}
}

// handleStructuredMessage processes structured WebSocket messages
func (wb *WebSocketBroadcaster) handleStructuredMessage(conn *websocket.Conn, message WebSocketMessage) {
	switch message.Type {
	case WSMsgPing:
		pongMsg := WebSocketMessage{
			Type:      WSMsgPong,
			Payload:   "pong",
			Timestamp: time.Now(),
		}
		wb.sendToClient(conn, pongMsg)

	default:
		// Unknown message type - log for debugging
		log.Printf("Unknown WebSocket message type: %s", message.Type)
	}
}

// sendToClient sends a message to a specific client
func (wb *WebSocketBroadcaster) sendToClient(conn *websocket.Conn, message WebSocketMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
		log.Printf("WebSocket write error: %v", err)
		wb.unregister <- conn
	}
}

// broadcastToAllClients sends a message to all connected clients
func (wb *WebSocketBroadcaster) broadcastToAllClients(message WebSocketMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal WebSocket message: %v", err)
		return
	}

	wb.mutex.RLock()
	defer wb.mutex.RUnlock()

	for client := range wb.clients {
		if err := client.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("WebSocket write error: %v", err)
			client.Close()
			delete(wb.clients, client)
		}
	}
}

// BroadcastServiceStatusUpdate sends a service status update to all connected clients
func (wb *WebSocketBroadcaster) BroadcastServiceStatusUpdate(serviceName, status string, errorMsg ...string) {
	var errorString string
	if len(errorMsg) > 0 {
		errorString = errorMsg[0]
	}

	update := ServiceStatusUpdate{
		ServiceName: serviceName,
		Status:      status,
		Error:       errorString,
	}

	message := WebSocketMessage{
		Type:      WSMsgServiceStatusUpdate,
		Payload:   update,
		Timestamp: time.Now(),
	}

	select {
	case wb.broadcast <- message:
	default:
		log.Printf("WebSocket broadcast channel full, dropped service status update for %s", serviceName)
	}
}

// BroadcastEnhancedServiceUpdate sends an enhanced service update to all connected clients
func (wb *WebSocketBroadcaster) BroadcastEnhancedServiceUpdate(serviceName string, service *models.EnhancedService, updateType string) {
	update := EnhancedServiceUpdate{
		ServiceName: serviceName,
		Service:     service,
		UpdateType:  updateType,
	}

	message := WebSocketMessage{
		Type:      WSMsgEnhancedServiceUpdate,
		Payload:   update,
		Timestamp: time.Now(),
	}

	select {
	case wb.broadcast <- message:
	default:
		log.Printf("WebSocket broadcast channel full, dropped enhanced service update for %s", serviceName)
	}
}

// BroadcastServiceLogs sends service logs to all connected clients
func (wb *WebSocketBroadcaster) BroadcastServiceLogs(serviceName, logMessage string) {
	logEntry := LogEntry{
		ServiceName: serviceName,
		Message:     logMessage,
		Timestamp:   time.Now(),
	}

	message := WebSocketMessage{
		Type:      WSMsgServiceLogs,
		Payload:   logEntry,
		Timestamp: time.Now(),
	}

	select {
	case wb.broadcast <- message:
	default:
		log.Printf("WebSocket broadcast channel full, dropped log message for %s", serviceName)
	}
}

// BroadcastImagePullProgress sends image pull progress to all connected clients
func (wb *WebSocketBroadcaster) BroadcastImagePullProgress(serviceName, imageName string, progress float64, status string) {
	progressUpdate := ImagePullProgress{
		ImageName:   imageName,
		ServiceName: serviceName,
		Progress:    progress,
		Status:      status,
	}

	message := WebSocketMessage{
		Type:      WSMsgImagePullProgress,
		Payload:   progressUpdate,
		Timestamp: time.Now(),
	}

	select {
	case wb.broadcast <- message:
	default:
		log.Printf("WebSocket broadcast channel full, dropped image progress for %s", serviceName)
	}
}

// BroadcastAppLogs sends application logs to all connected clients
func (wb *WebSocketBroadcaster) BroadcastAppLogs(message string, level string) {
	logEntry := LogEntry{
		Message:   message,
		Timestamp: time.Now(),
		Level:     level,
	}

	wsMessage := WebSocketMessage{
		Type:      WSMsgAppLogs,
		Payload:   logEntry,
		Timestamp: time.Now(),
	}

	select {
	case wb.broadcast <- wsMessage:
	default:
		// Don't log here to avoid infinite loop
	}
}

// GetClientCount returns the number of connected WebSocket clients
func (wb *WebSocketBroadcaster) GetClientCount() int {
	wb.mutex.RLock()
	defer wb.mutex.RUnlock()
	return len(wb.clients)
}
