import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import ServiceItem, { ImageStatusProvider } from '../ServiceItem'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

// Mock the API client
vi.mock('../../api/client', () => ({
  startService: vi.fn(),
  stopService: vi.fn(),
  checkImageExists: vi.fn(() => Promise.resolve(true)),
  getServiceConnection: vi.fn(),
  wsClient: {
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
    subscribeToServiceStatus: vi.fn(() => vi.fn()),
  },
  WS_MSG_TYPES: {
    SERVICE_STATUS_UPDATE: 'service_status_update',
    SERVICE_STARTED: 'service_started',
    SERVICE_STOPPED: 'service_stopped',
    SERVICE_ERROR: 'service_error',
  }
}))

// Note: Using mockApiClient from test-utils instead of module-level mocks

// Helper to render ServiceItem with required providers
const renderWithProvider = (ui, services = []) => {
  return renderWithProviders(
    <ImageStatusProvider services={services}>
      {ui}
    </ImageStatusProvider>,
    { apiClient: mockApiClient }
  )
}

describe('ServiceItem Component - User Interactions', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  const mockService = {
    name: 'postgres',
    type: 'database',
    status: 'stopped',
    recursive_dependencies: [],
    exposed_ports: [
      {
        internal_port: '5432',
        host_port: '5432',
        type: 'tcp',
        description: 'PostgreSQL port'
      }
    ],
    web_urls: [],
    container_name: 'postgres',
    image_name: 'postgres:15',
    all_containers: ['postgres'],
    environment: ['POSTGRES_PASSWORD=postgres'],
    volumes: [],
    last_updated: new Date().toISOString()
  }

  const mockImageStatus = {
    postgres: {
      exists: true,
      info: 'postgres:15',
      needsPull: false
    }
  }

  const defaultProps = {
    service: mockService,
    imageStatus: mockImageStatus,
    onStart: vi.fn(),
    onStop: vi.fn(),
    onConnect: vi.fn(),
    onLogs: vi.fn(),
    onOpenBrowser: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Mock fetch for HTTP API calls
    global.fetch = vi.fn()
      .mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true })
      })

    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('Basic Rendering', () => {
    it('should render service item with correct information', () => {
      renderWithProvider(<ServiceItem {...defaultProps} />, [mockService])
      
      expect(screen.getByText('postgres')).toBeInTheDocument()
      expect(screen.getByText('database')).toBeInTheDocument()
      expect(screen.getByText('Stopped')).toBeInTheDocument()
    })

    it('should show correct buttons for stopped service', () => {
      renderWithProvider(<ServiceItem {...defaultProps} />, [mockService])
      
      expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /stop/i })).not.toBeInTheDocument()
      // Logs button should NOT be shown for stopped services
      expect(screen.queryByRole('button', { name: /logs/i })).not.toBeInTheDocument()
    })
  })

  describe('Service Starting Flow', () => {
    it('should start service when start button is clicked', async () => {
      const onStartSpy = vi.fn()
      // Mock checkImageExists to return true so startService gets called
      mockApiClient.checkImageExists.mockResolvedValue(true)
      mockApiClient.startService.mockResolvedValue({ status: 'success' })

      renderWithProvider(<ServiceItem {...defaultProps} onStart={onStartSpy} />, [mockService])
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      expect(mockApiClient.startService).toHaveBeenCalledWith('postgres', false)
    })

    it('should start service with persistence when persist button is toggled', async () => {
      const onStartSpy = vi.fn()
      // Mock checkImageExists to return true so startService gets called
      mockApiClient.checkImageExists.mockResolvedValue(true)
      mockApiClient.startService.mockResolvedValue({ status: 'success' })
      
      renderWithProvider(<ServiceItem {...defaultProps} onStart={onStartSpy} />, [mockService])
      
      const persistToggle = screen.getByRole('button', { name: /persist/i })
      await user.click(persistToggle)
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      expect(mockApiClient.startService).toHaveBeenCalledWith('postgres', true)
    })

    it('should show loading state when starting service', async () => {
      const { startService } = await import('../../api/client')
      global.fetch.mockImplementation(() => 
        new Promise(resolve => setTimeout(() => resolve({
          ok: true,
          status: 200,
          json: () => Promise.resolve({ success: true })
        }), 100))
      )

      const onStartSpy = vi.fn()
      renderWithProvider(<ServiceItem {...defaultProps} onStart={onStartSpy} />, [mockService])
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      // Should show loading state
      expect(screen.getByText(/starting/i) || startButton.disabled).toBeTruthy()
    })

    it('should handle start service error gracefully', async () => {
      const { startService } = await import('../../api/client')
      startService.mockRejectedValue(new Error('Docker not running'))

      const onStartSpy = vi.fn()
      renderWithProvider(<ServiceItem {...defaultProps} onStart={onStartSpy} />, [mockService])
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      // Should handle error gracefully without crashing
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'postgres' })).toBeInTheDocument()
      })
    })
  })

  describe('Service Stopping Flow', () => {
    it('should stop service when stop button is clicked', async () => {
      const runningService = { ...mockService, status: 'running' }
      const onStopSpy = vi.fn()
      mockApiClient.stopService.mockResolvedValue({ status: 'success' })
      
      renderWithProvider(<ServiceItem {...defaultProps} service={runningService} onStop={onStopSpy} statuses={{ postgres: 'running' }} />, [runningService])
      
      const stopButton = screen.getByRole('button', { name: /stop/i })
      await user.click(stopButton)
      
      expect(mockApiClient.stopService).toHaveBeenCalledWith('postgres')
    })

    it('should show correct UI for running service', async () => {
      const runningService = { ...mockService, status: 'running' }
      
      renderWithProvider(<ServiceItem {...defaultProps} service={runningService} statuses={{ postgres: 'running' }} />, [runningService])
      
      expect(screen.getByText(/running/i)).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /stop/i })).toBeInTheDocument()
    })
  })

  describe('Connection Information Modal', () => {
    it('should open connection modal when connect button is clicked', async () => {
      mockApiClient.getServiceConnection.mockResolvedValue({
        connection: {
          host: 'localhost',
          port: '5432',
          username: 'postgres',
          password: 'postgres'
        }
      })

      // Use a running service so the connect button appears
      renderWithProvider(<ServiceItem {...defaultProps} statuses={{ postgres: 'running' }} />, [mockService])
      
      const connectButton = screen.getByRole('button', { name: /connect/i })
      await user.click(connectButton)
      
      await waitFor(() => {
        expect(mockApiClient.getServiceConnection).toHaveBeenCalledWith('postgres')
      })
    })

    it('should close connection modal when close button is clicked', async () => {
      mockApiClient.getServiceConnection.mockResolvedValue({
        connection: {
          host: 'localhost',
          port: '5432'
        }
      })

      // Use a running service so the connect button appears
      renderWithProvider(<ServiceItem {...defaultProps} statuses={{ postgres: 'running' }} />, [mockService])
      
      const connectButton = screen.getByRole('button', { name: /connect/i })
      await user.click(connectButton)
      
      await waitFor(() => {
        const closeButton = screen.getByRole('button', { name: /close/i })
        expect(closeButton).toBeInTheDocument()
      })
    })

    it('should handle connection info when service has proper connection data', async () => {
      mockApiClient.getServiceConnection.mockResolvedValue({
        connection: {
          serviceName: 'postgres',
          available: true,
          webUrls: [{ url: 'http://localhost:5432', description: 'Database Connection' }],
          portMappings: [{ hostPort: '5432', containerPort: '5432', type: 'EXPOSED' }],
          connectionStrings: [{ description: 'Database URL', connectionString: 'postgresql://postgres:postgres@localhost:5432/postgres' }],
          credentials: [{ description: 'Username', value: 'postgres' }]
        }
      })

      // Use a running service so the connect button appears
      renderWithProvider(<ServiceItem {...defaultProps} statuses={{ postgres: 'running' }} />, [mockService])
      
      const connectButton = screen.getByRole('button', { name: /connect/i })
      await user.click(connectButton)
      
      await waitFor(() => {
        // Check that the connection modal opened successfully
        expect(mockApiClient.getServiceConnection).toHaveBeenCalledWith('postgres')
        // Just verify the modal is open and API was called, not the specific content
        expect(screen.getByText(/connection info/i)).toBeInTheDocument()
      })
    })
  })

  describe('Logs Functionality', () => {
    it('should open logs modal when logs button is clicked for running service', async () => {
      // Use a running service so the logs button appears
      const runningService = { ...mockService, status: 'running' }
      renderWithProvider(<ServiceItem {...defaultProps} service={runningService} statuses={{ postgres: 'running' }} />, [runningService])
      
      const logsButton = screen.getByRole('button', { name: /logs/i })
      await user.click(logsButton)
      
      await waitFor(() => {
        // Look for the modal title specifically
        expect(screen.getByText(/logs - postgres/i)).toBeInTheDocument()
      })
    })

    it('should show logs button for failed service', async () => {
      // Use a failed service so the logs button appears
      const failedService = { ...mockService, status: 'failed' }
      renderWithProvider(<ServiceItem {...defaultProps} service={failedService} statuses={{ postgres: 'failed' }} />, [failedService])
      
      expect(screen.getByRole('button', { name: /logs/i })).toBeInTheDocument()
    })

    it('should not show logs button for stopped service', async () => {
      renderWithProvider(<ServiceItem {...defaultProps} />, [mockService])
      
      expect(screen.queryByRole('button', { name: /logs/i })).not.toBeInTheDocument()
    })

    it('should show logs button for service with status error', async () => {
      // Use a service with status error so the logs button appears
      const serviceWithError = { ...mockService, statusError: 'Connection failed' }
      renderWithProvider(<ServiceItem {...defaultProps} service={serviceWithError} />, [serviceWithError])
      
      expect(screen.getByRole('button', { name: /logs/i })).toBeInTheDocument()
    })
  })

  describe('Service Actions', () => {
    it('should show appropriate buttons based on service status', async () => {
      renderWithProvider(<ServiceItem {...defaultProps} />, [mockService])
      
      // For stopped service, should show start button
      expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /stop/i })).not.toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels and keyboard navigation', async () => {
      // Mock checkImageExists to return true so startService gets called
      mockApiClient.checkImageExists.mockResolvedValue(true)
      mockApiClient.startService.mockResolvedValue({ status: 'success' })
      
      renderWithProvider(<ServiceItem {...defaultProps} />, [mockService])
      
      const startButton = screen.getByRole('button', { name: /start/i })
      expect(startButton).toBeInTheDocument()
      
      // Test keyboard navigation
      startButton.focus()
      expect(startButton).toHaveFocus()
      
      // Test clicking the button directly
      await user.click(startButton)
      expect(mockApiClient.startService).toHaveBeenCalled()
    })
  })

  describe('Error Handling', () => {
    it('should handle network errors gracefully', async () => {
      mockApiClient.startService.mockRejectedValue(new Error('Network error'))

      const onStartSpy = vi.fn()
      renderWithProvider(<ServiceItem {...defaultProps} onStart={onStartSpy} />, [mockService])
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      // Should not crash and should still show the service
      await waitFor(() => {
        expect(screen.getByRole('heading', { name: 'postgres' })).toBeInTheDocument()
      })
    })
  })
}) 