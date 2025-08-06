import React from 'react';
import { render } from '@testing-library/react';
import { vi } from 'vitest';
import { ApiProvider } from '../contexts/ApiContext';
import * as demoApiClient from '../api/demo-client';

// Custom render function that wraps components with necessary providers
export function renderWithProviders(ui, { apiClient = demoApiClient, ...options } = {}) {
  function Wrapper({ children }) {
    return (
      <ApiProvider apiClient={apiClient}>
        {children}
      </ApiProvider>
    );
  }
  return render(ui, { wrapper: Wrapper, ...options });
}

// Mock API client for tests that returns empty/default values
export const mockApiClient = {
  // Service management
  listServices: vi.fn().mockResolvedValue([]),
  getAllServiceStatuses: vi.fn().mockResolvedValue({}),
  startService: vi.fn().mockResolvedValue({ status: 'success' }),
  stopService: vi.fn().mockResolvedValue({ status: 'success' }),
  stopAllServices: vi.fn().mockResolvedValue({ status: 'success' }),
  refreshServiceStatus: vi.fn().mockResolvedValue({}),
  getRunningServices: vi.fn().mockResolvedValue({ running_services: {} }),
  getServiceConnection: vi.fn().mockResolvedValue({}),
  openServiceConnection: vi.fn().mockResolvedValue({ status: 'success' }),

  // Custom services
  listCustomServices: vi.fn().mockResolvedValue([]),
  getCustomService: vi.fn().mockResolvedValue({}),
  uploadCustomService: vi.fn().mockResolvedValue({ status: 'success' }),
  updateCustomService: vi.fn().mockResolvedValue({ status: 'success' }),
  deleteCustomService: vi.fn().mockResolvedValue({ status: 'success' }),
  validateCustomCompose: vi.fn().mockResolvedValue({ valid: true }),
  getCustomServiceStats: vi.fn().mockResolvedValue({ total: 0, active: 0 }),

  // Image management
  getAllImageStatuses: vi.fn().mockResolvedValue({ image_statuses: {}, total_services: 0 }),
  checkImageExists: vi.fn().mockResolvedValue({ exists: true }),
  startImagePull: vi.fn().mockResolvedValue({ status: 'success' }),
  stopImagePull: vi.fn().mockResolvedValue({ status: 'success' }),
  getImagePullProgress: vi.fn().mockResolvedValue({ progress: 0, status: 'pending' }),

  // Logging
  getServiceLogs: vi.fn().mockResolvedValue({ logs: [] }),
  startServiceLogStream: vi.fn().mockResolvedValue({ status: 'success' }),
  stopServiceLogStream: vi.fn().mockResolvedValue({ status: 'success' }),
  getActiveLogStreams: vi.fn().mockResolvedValue({ streams: [] }),
  getAppLogs: vi.fn().mockResolvedValue({ logs: [] }),
  getAppLogEntries: vi.fn().mockResolvedValue({ logs: [] }),
  getAppLogsSince: vi.fn().mockResolvedValue({ logs: [] }),

  // Runtime management
  getRuntimeStatus: vi.fn().mockResolvedValue({
    hasAnyRuntime: true,
    preferredRuntime: 'docker',
    availableRuntimes: ['docker'],
    canProceed: true,
    recommendedAction: 'Ready to use docker',
    runtimeStatuses: [
      {
        name: 'docker',
        isInstalled: true,
        isRunning: true,
        isAvailable: true,
        hasCompose: true,
        version: '24.0.0',
        error: '',
        canAutoStart: true,
        installationGuide: '',
        startupCommand: '',
        requiresMachine: false,
        machineStatus: ''
      }
    ]
  }),
  getCurrentRuntime: vi.fn().mockResolvedValue({ runtime: 'docker' }),
  startRuntime: vi.fn().mockResolvedValue({ status: 'success' }),
  restartRuntime: vi.fn().mockResolvedValue({ status: 'success' }),
  reinitializeRuntime: vi.fn().mockResolvedValue({ status: 'success' }),
  setCustomDockerPath: vi.fn().mockResolvedValue({ status: 'success' }),
  setCustomPodmanPath: vi.fn().mockResolvedValue({ status: 'success' }),
  getCustomDockerPath: vi.fn().mockResolvedValue({ path: '' }),
  getCustomPodmanPath: vi.fn().mockResolvedValue({ path: '' }),

  // Health & info
  getHealth: vi.fn().mockResolvedValue({ status: 'healthy' }),
  getApiInfo: vi.fn().mockResolvedValue({ version: '1.0.0' }),
  shutdownApplication: vi.fn().mockResolvedValue({ status: 'success' }),

  // WebSocket
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
    subscribeToServiceStatus: vi.fn(() => vi.fn()), // Returns an unsubscribe function
    getState: vi.fn().mockReturnValue('CONNECTED')
  },

  // URL opening
  openURL: vi.fn(),

  // Message types
  WS_MSG_TYPES: {
    SERVICE_STATUS_UPDATE: 'service_status_update',
    SERVICE_STARTED: 'service_started',
    SERVICE_STOPPED: 'service_stopped',
    SERVICE_ERROR: 'service_error',
    SERVICE_LOGS: 'service_logs',
    APP_LOGS: 'app_logs',
    LOG_STREAM: 'log_stream',
    IMAGE_PULL_PROGRESS: 'image_pull_progress',
    IMAGE_PULL_COMPLETE: 'image_pull_complete',
    IMAGE_PULL_ERROR: 'image_pull_error',
    RUNTIME_STATUS_UPDATE: 'runtime_status_update',
    CLIENT_CONNECTED: 'client_connected',
    CLIENT_DISCONNECTED: 'client_disconnected',
    PING: 'ping',
    PONG: 'pong'
  }
};