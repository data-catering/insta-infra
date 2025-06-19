// HTTP API client for insta-infra backend
// HTTP API client for browser-based web UI

const API_BASE_URL = window.location.origin + '/api/v1';

// Helper function to handle API responses
async function handleResponse(response) {
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
    throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
  }
  return response.json();
}

// Helper function to make API requests
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  };

  try {
    const response = await fetch(url, config);
    return await handleResponse(response);
  } catch (error) {
    console.error(`API request failed: ${endpoint}`, error);
    throw error;
  }
}

// ==================== SERVICE MANAGEMENT ====================

export async function listServices() {
  return apiRequest('/services');
}

export async function getAllServiceStatuses() {
  return apiRequest('/services/all/status');
}

export async function startService(serviceName, persist = false) {
  return apiRequest(`/services/${serviceName}/start?persist=${persist}`, {
    method: 'POST',
  });
}

export async function stopService(serviceName) {
  return apiRequest(`/services/${serviceName}/stop`, {
    method: 'POST',
  });
}

export async function stopAllServices() {
  return apiRequest('/services/all/stop', {
    method: 'POST',
  });
}

export async function refreshServiceStatus() {
  return apiRequest('/services/refresh-status', {
    method: 'POST',
  });
}

export async function getRunningServices() {
  return apiRequest('/services/running');
}

export async function getServiceConnection(serviceName) {
  return apiRequest(`/services/${serviceName}/connection`);
}

export async function openServiceConnection(serviceName) {
  return apiRequest(`/services/${serviceName}/open`, {
    method: 'POST',
  });
}

// ==================== IMAGE MANAGEMENT ====================

export async function getAllImageStatuses() {
  return apiRequest('/images/all/status');
}

export async function checkImageExists(imageName) {
  return apiRequest(`/images/${imageName}/exists`);
}

export async function startImagePull(imageName) {
  return apiRequest(`/images/${imageName}/pull`, {
    method: 'POST',
  });
}

export async function stopImagePull(imageName) {
  return apiRequest(`/images/${imageName}/pull`, {
    method: 'DELETE',
  });
}

export async function getImagePullProgress(imageName) {
  return apiRequest(`/images/${imageName}/progress`);
}

// ==================== LOGGING ====================

export async function getServiceLogs(serviceName) {
  return apiRequest(`/services/${serviceName}/logs`);
}

export async function getAppLogs() {
  return apiRequest('/logs/app');
}

export async function getAppLogEntries() {
  return apiRequest('/logs/app/entries');
}

export async function getAppLogsSince(timestamp) {
  return apiRequest(`/logs/app/since/${timestamp}`);
}

// ==================== RUNTIME MANAGEMENT ====================

export async function getRuntimeStatus() {
  return apiRequest('/runtime/status');
}

export async function getCurrentRuntime() {
  return apiRequest('/runtime/current');
}

export async function startRuntime(runtime) {
  return apiRequest('/runtime/start', {
    method: 'POST',
    body: JSON.stringify({ runtime }),
  });
}

export async function restartRuntime(runtime) {
  return apiRequest('/runtime/restart', {
    method: 'POST',
    body: JSON.stringify({ runtime }),
  });
}

export async function reinitializeRuntime() {
  return apiRequest('/runtime/reinitialize', {
    method: 'POST',
  });
}

export async function setCustomDockerPath(path) {
  return apiRequest('/runtime/docker-path', {
    method: 'PUT',
    body: JSON.stringify({ path }),
  });
}

export async function setCustomPodmanPath(path) {
  return apiRequest('/runtime/podman-path', {
    method: 'PUT',
    body: JSON.stringify({ path }),
  });
}

export async function getCustomDockerPath() {
  return apiRequest('/runtime/docker-path');
}

export async function getCustomPodmanPath() {
  return apiRequest('/runtime/podman-path');
}

// ==================== HEALTH & INFO ====================

export async function getHealth() {
  return apiRequest('/health');
}

export async function getApiInfo() {
  return apiRequest('/info');
}

export async function shutdownApplication() {
  return apiRequest('/app/shutdown', {
    method: 'POST',
  });
}

// ==================== WEBSOCKET SUPPORT ====================

// WebSocket message types - must match server constants
export const WS_MSG_TYPES = {
  // Service-related events
  SERVICE_STATUS_UPDATE: 'service_status_update',
  SERVICE_STARTED: 'service_started',
  SERVICE_STOPPED: 'service_stopped',
  SERVICE_ERROR: 'service_error',

  // Logging events  
  SERVICE_LOGS: 'service_logs',
  APP_LOGS: 'app_logs',
  LOG_STREAM: 'log_stream',

  // Image/Container events
  IMAGE_PULL_PROGRESS: 'image_pull_progress',
  IMAGE_PULL_COMPLETE: 'image_pull_complete',
  IMAGE_PULL_ERROR: 'image_pull_error',

  // Runtime events
  RUNTIME_STATUS_UPDATE: 'runtime_status_update',

  // Connection events
  CLIENT_CONNECTED: 'client_connected',
  CLIENT_DISCONNECTED: 'client_disconnected',

  // Keep-alive
  PING: 'ping',
  PONG: 'pong'
};

export class WebSocketClient {
  constructor() {
    this.ws = null;
    this.listeners = new Map();
    this.reconnectAttempts = 0;
    this.maxReconnectAttempts = 3;
    this.reconnectDelay = 2000;
    this.isConnecting = false;
    this.shouldReconnect = true;
    this.pingInterval = null;
  }

  connect() {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return;
    }

    this.isConnecting = true;
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws`;

    try {
      this.ws = new WebSocket(wsUrl);

      this.ws.onopen = () => {
        this.isConnecting = false;
        this.reconnectAttempts = 0;
        this.startPingInterval();
      };

      this.ws.onmessage = (event) => {
        try {
          // Try to parse as JSON for structured messages
          const data = JSON.parse(event.data);
          
          // Validate message structure
          if (data && typeof data === 'object' && data.type && data.payload !== undefined) {
            this.handleMessage(data);
          } else {
            console.warn('Received invalid WebSocket message structure:', data);
          }
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error, 'Data:', event.data);
        }
      };

      this.ws.onclose = (event) => {
        this.isConnecting = false;
        this.stopPingInterval();
        
        // Only attempt to reconnect if it's not a normal closure or if we were previously connected
        if (this.shouldReconnect && event.code !== 1000) {
          this.attemptReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        this.isConnecting = false;
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      this.isConnecting = false;
    }
  }

  disconnect() {
    this.shouldReconnect = false;
    this.stopPingInterval();
    if (this.ws) {
      this.ws.close(1000, 'Normal closure'); // Normal closure
      this.ws = null;
    }
  }

  startPingInterval() {
    this.stopPingInterval(); // Clear any existing interval
    this.pingInterval = setInterval(() => {
      this.sendPing();
    }, 30000); // Ping every 30 seconds
  }

  stopPingInterval() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  sendPing() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      // Try structured ping first, fallback to plain text for compatibility
      try {
        this.sendMessage(WS_MSG_TYPES.PING, 'ping');
      } catch (error) {
        // Fallback to plain text ping
        this.ws.send('ping');
      }
    }
  }

  attemptReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts && this.shouldReconnect) {
      this.reconnectAttempts++;
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1); // Exponential backoff
      
      setTimeout(() => {
        if (this.shouldReconnect) {
          this.connect();
        }
      }, delay);
    } else {
      console.error('Max reconnection attempts reached or reconnection disabled');
    }
  }

  handleMessage(data) {
    const { type, payload, timestamp } = data;
    
    // Always notify subscribers
    if (this.listeners.has(type)) {
      this.listeners.get(type).forEach(callback => {
        try {
          callback(payload, timestamp);
        } catch (error) {
          console.error(`Error in WebSocket callback for ${type}:`, error);
        }
      });
    }
  }

  subscribe(eventType, callback) {
    if (!this.listeners.has(eventType)) {
      this.listeners.set(eventType, new Set());
    }
    
    // Prevent duplicate subscriptions
    if (this.listeners.get(eventType).has(callback)) {
      return; // Already subscribed
    }
    
    this.listeners.get(eventType).add(callback);
  }

  unsubscribe(eventType, callback) {
    if (this.listeners.has(eventType)) {
      const wasDeleted = this.listeners.get(eventType).delete(callback);
    }
  }

  // Subscribe to service status updates for a specific service
  subscribeToServiceStatus(serviceName, callback) {
    const wrappedCallback = (payload) => {
      if (payload.service_name === serviceName) {
        callback(payload);
      }
    };
    this.subscribe(WS_MSG_TYPES.SERVICE_STATUS_UPDATE, wrappedCallback);
    return wrappedCallback; // Return for unsubscribing
  }

  // Subscribe to service logs for a specific service
  subscribeToServiceLogs(serviceName, callback) {
    const wrappedCallback = (payload) => {
      if (payload.service_name === serviceName) {
        callback(payload);
      }
    };
    this.subscribe(WS_MSG_TYPES.SERVICE_LOGS, wrappedCallback);
    return wrappedCallback; // Return for unsubscribing
  }

  // Send structured message to server
  sendMessage(type, payload) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message = {
        type,
        payload,
        timestamp: new Date().toISOString()
      };
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket not connected, cannot send message');
    }
  }

  // Get connection state
  getState() {
    if (!this.ws) return 'DISCONNECTED';
    switch (this.ws.readyState) {
      case WebSocket.CONNECTING: return 'CONNECTING';
      case WebSocket.OPEN: return 'CONNECTED';
      case WebSocket.CLOSING: return 'CLOSING';
      case WebSocket.CLOSED: return 'DISCONNECTED';
      default: return 'UNKNOWN';
    }
  }
}

// Create a singleton WebSocket client
export const wsClient = new WebSocketClient();

// Auto-connect when module is imported
wsClient.connect();

// Browser URL opening utility
export function openURL(url) {
  window.open(url, '_blank');
}

// These functions are just aliases to match the frontend usage
export async function AttemptStartRuntime(runtime) {
  return startRuntime(runtime);
}

export async function WaitForRuntimeReady(runtimeName, maxWaitSeconds) {
  // Simple wait implementation - check runtime status in a loop
  for (let i = 0; i < maxWaitSeconds; i++) {
    try {
      const status = await getRuntimeStatus();
      const runtimeStatus = status.runtimeStatuses?.find(rt => rt.name === runtimeName);
      if (runtimeStatus && runtimeStatus.isAvailable) {
        return {
          success: true,
          message: `${runtimeName} is ready`
        };
      }
      // Wait 1 second before next check
      await new Promise(resolve => setTimeout(resolve, 1000));
    } catch (error) {
      // Continue checking even if there's an error
    }
  }
  
  return {
    success: false,
    error: `${runtimeName} did not become ready within ${maxWaitSeconds} seconds`
  };
}

export async function ReinitializeRuntime() {
  return reinitializeRuntime();
}

export async function SetCustomDockerPath(path) {
  return setCustomDockerPath(path);
}

export async function SetCustomPodmanPath(path) {
  return setCustomPodmanPath(path);
} 