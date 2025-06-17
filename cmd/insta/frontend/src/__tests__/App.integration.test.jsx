import { render, screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from '../App';

// Mock the API client to prevent actual HTTP requests
vi.mock('../api/client', () => ({
  listServices: vi.fn(),
  fetchServices: vi.fn(),
  fetchServiceStatuses: vi.fn(),
  fetchRunningServices: vi.fn(),
  checkRuntimeStatus: vi.fn(),
  getCurrentRuntime: vi.fn(),
  checkHealth: vi.fn(),
  getAllServiceStatuses: vi.fn(),
  startService: vi.fn(),
  stopService: vi.fn(),
  getServiceConnection: vi.fn(),
  getServiceLogs: vi.fn(),
  stopAllServices: vi.fn(),
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
    sendMessage: vi.fn(),
    getState: vi.fn(() => ({ connected: false, reconnecting: false })),
    ws: { close: vi.fn() }
  },
  WS_MSG_TYPES: {
    SERVICE_STATUS_UPDATE: 'service_status_update',
    SERVICE_STARTED: 'service_started',
    SERVICE_STOPPED: 'service_stopped',
    SERVICE_ERROR: 'service_error',
  }
}));

describe('App Integration Tests - Simplified', () => {
  beforeEach(async () => {
    vi.clearAllMocks();

    // Mock all API functions to return appropriate data
    const { 
      listServices,
      fetchServices, 
      fetchServiceStatuses, 
      fetchRunningServices, 
      checkRuntimeStatus, 
      getCurrentRuntime, 
      checkHealth,
      getAllServiceStatuses,
      startService,
      stopService,
      getServiceConnection,
      getServiceLogs,
      stopAllServices
    } = await import('../api/client');

    // Mock successful responses to prevent error states
    listServices.mockResolvedValue([]);
    fetchServices.mockResolvedValue([]);
    fetchServiceStatuses.mockResolvedValue({});
    getAllServiceStatuses.mockResolvedValue({});
    fetchRunningServices.mockResolvedValue([]);
    checkRuntimeStatus.mockResolvedValue({ canProceed: true, status: 'ready', runtime: 'docker' });
    getCurrentRuntime.mockResolvedValue({ runtime: 'docker' });
    checkHealth.mockResolvedValue({ status: 'ok' });
    startService.mockResolvedValue({ success: true });
    stopService.mockResolvedValue({ success: true });
    stopAllServices.mockResolvedValue({ success: true });
    getServiceConnection.mockResolvedValue({
      host: 'localhost',
      port: '5432',
      url: 'http://localhost:5432'
    });
    getServiceLogs.mockResolvedValue([
      'Service started successfully',
      'Ready to accept connections'
    ]);
  });

  describe('Application Rendering', () => {
    it('should render without crashing', () => {
      render(<App />);
      expect(document.body).toBeInTheDocument();
    });

    it('should display header with logo', async () => {
      render(<App />);
      
      await waitFor(() => {
        expect(screen.getByText('Insta')).toBeInTheDocument();
        expect(screen.getByText('Infra')).toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should display footer with infrastructure text', async () => {
      render(<App />);
      
      await waitFor(() => {
        expect(screen.getByText('Infrastructure Control Panel')).toBeInTheDocument();
      }, { timeout: 3000 });
    });
  });
});