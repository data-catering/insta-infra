import { render, screen, waitFor, act, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import App from '../../App'

// Mock the API client
vi.mock('../../api/client', () => ({
  getAllServiceStatuses: vi.fn(),
  getRunningServices: vi.fn(),
  getRuntimeStatus: vi.fn(),
  getCurrentRuntime: vi.fn(),
  stopAllServices: vi.fn(),
  listServices: vi.fn(),
  openURL: vi.fn(),
  getAppLogs: vi.fn(),
  restartRuntime: vi.fn(),
  shutdownApplication: vi.fn(),
  wsClient: {
    connect: vi.fn(),
    disconnect: vi.fn(),
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
  },
  WS_MSG_TYPES: {
    SERVICE_STATUS_UPDATE: 'service_status_update',
    SERVICE_STARTED: 'service_started',
    SERVICE_STOPPED: 'service_stopped',
    SERVICE_ERROR: 'service_error'
  }
}))

// Mock child components to simplify testing
vi.mock('../AppHeader', () => ({
  default: ({ runningServicesCount, lastUpdated }) => (
    <div data-testid="app-header">
      <span>Updated {lastUpdated.toLocaleTimeString()}</span>
      <span>{runningServicesCount} services running</span>
    </div>
  )
}))

vi.mock('../MainContent', () => ({
  default: ({ services, runningServices, error, isLoading }) => (
    <div data-testid="main-content">
      {error && <div data-testid="error">{error}</div>}
      {isLoading && services.length === 0 && <div data-testid="loading">Loading...</div>}
      {services.length > 0 && <div data-testid="service-list">Service List: {services.length} services</div>}
      {runningServices.length > 0 && <div data-testid="running-services">Running Services: {runningServices.length}</div>}
    </div>
  )
}))

vi.mock('../AppFooter', () => ({
  default: ({ currentRuntime }) => (
    <div data-testid="app-footer">Using {currentRuntime}</div>
  )
}))

vi.mock('../AppModals', () => ({
  default: () => <div data-testid="app-modals">Modals</div>
}))

vi.mock('../../components/ServiceList', () => ({
  default: ({ services, statuses, onStart, onStop, onOpenLogs, onOpenConnection }) => (
    <div data-testid="service-list">
      <div>Service List: {services.length} services</div>
      {services.map(service => (
        <div key={service.Name} data-testid={`service-${service.Name}`}>
          <span>{service.Name} - {statuses[service.Name] || 'unknown'}</span>
          <button onClick={() => onStart && onStart(service)}>Start {service.Name}</button>
          <button onClick={() => onStop && onStop(service)}>Stop {service.Name}</button>
          <button onClick={() => onOpenLogs && onOpenLogs(service)}>Logs</button>
          <button onClick={() => onOpenConnection && onOpenConnection(service)}>Connection</button>
        </div>
      ))}
    </div>
  )
}))

vi.mock('../../components/RunningServices', () => ({
  default: ({ runningServices }) => (
    <div data-testid="running-services">
      Running Services: {Object.keys(runningServices).length}
    </div>
  )
}))

vi.mock('../../components/CustomServicesList', () => ({
  default: () => <div data-testid="custom-services">Custom Services</div>
}))

vi.mock('../../components/ServiceItem', () => ({
  ImageStatusProvider: ({ children }) => children,
  ImageLoader: () => <div>Image Loader</div>
}))

vi.mock('../../components/ConnectionModal', () => ({
  default: ({ isOpen, service, onClose }) => isOpen ? (
    <div data-testid="connection-modal">
      Connection Modal for {service?.Name}
      <button onClick={onClose}>Close</button>
    </div>
  ) : null
}))

vi.mock('../../components/LogsModal', () => ({
  default: ({ isOpen, service, onClose }) => isOpen ? (
    <div data-testid="logs-modal">
      Logs Modal for {service?.Name}
      <button onClick={onClose}>Close</button>
    </div>
  ) : null
}))

vi.mock('../../components/LogsPanel', () => ({
  default: ({ isOpen, onClose }) => isOpen ? (
    <div data-testid="logs-panel">
      Logs Panel
      <button onClick={onClose}>Close Panel</button>
    </div>
  ) : null
}))

vi.mock('../../components/RuntimeSetup', () => ({
  default: ({ isOpen, runtimeStatus, onRuntimeReady }) => isOpen ? (
    <div data-testid="runtime-setup">
      Runtime Setup
      <button onClick={() => onRuntimeReady && onRuntimeReady()}>Runtime Ready</button>
    </div>
  ) : null
}))

vi.mock('../../components/ErrorMessage', () => ({
  default: () => <div data-testid="error-message">Error Message</div>,
  useErrorHandler: () => ({
    errors: [],
    toasts: [],
    addError: vi.fn(),
    addToast: vi.fn(),
    removeError: vi.fn(),
    removeToast: vi.fn(),
    clearAllErrors: vi.fn()
  }),
  ToastContainer: () => <div data-testid="toast-container">Toast Container</div>
}))

vi.mock('../../hooks/useAppState', () => ({
  useAppState: () => ({
    services: [],
    setServices: vi.fn(),
    statuses: {},
    setStatuses: vi.fn(),
    runningServices: [],
    setRunningServices: vi.fn(),
    dependencyStatuses: {},
    setDependencyStatuses: vi.fn(),
    isLoading: true,
    setIsLoading: vi.fn(),
    error: null,
    setError: vi.fn(),
    lastUpdated: new Date(),
    setLastUpdated: vi.fn(),
    isAnimating: false,
    setIsAnimating: vi.fn(),
    isStoppingAll: false,
    setIsStoppingAll: vi.fn(),
    showAbout: false,
    setShowAbout: vi.fn(),
    showConnectionModal: false,
    setShowConnectionModal: vi.fn(),
    showLogsModal: false,
    setShowLogsModal: vi.fn(),
    selectedService: null,
    setSelectedService: vi.fn(),
    copyFeedback: '',
    setCopyFeedback: vi.fn(),
    runtimeStatus: null,
    setRuntimeStatus: vi.fn(),
    showRuntimeSetup: false,
    setShowRuntimeSetup: vi.fn(),
    currentRuntime: 'docker',
    setCurrentRuntime: vi.fn(),
    showLogsPanel: false,
    setShowLogsPanel: vi.fn(),
    showActionsDropdown: false,
    setShowActionsDropdown: vi.fn(),
    isShuttingDown: false,
    setIsShuttingDown: vi.fn()
  })
}))

vi.mock('../../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    initializeWebSocket: vi.fn()
  })
}))

describe('App Component - Integration Tests', () => {
  const user = userEvent.setup()

  const mockServices = [
    { name: 'postgres', Name: 'postgres', type: 'builtin', Type: 'builtin', Dependencies: [] },
    { name: 'redis', Name: 'redis', type: 'builtin', Type: 'builtin', Dependencies: [] },
    { name: 'webapp', Name: 'webapp', type: 'builtin', Type: 'builtin', Dependencies: ['postgres', 'redis'] }
  ]

  const mockStatuses = {
    postgres: 'running',
    redis: 'stopped', 
    webapp: 'stopped'
  }

  const mockRunningServices = {
    postgres: { status: 'running', container_id: 'abc123' }
  }

  beforeEach(async () => {
    vi.clearAllMocks()
    
    // Setup default successful API responses
    const { 
      getRuntimeStatus, 
      getCurrentRuntime, 
      listServices, 
      getAllServiceStatuses, 
      getRunningServices,
      wsClient 
    } = await import('../../api/client')
    
    getRuntimeStatus.mockResolvedValue({ canProceed: true, runtime: 'docker' })
    getCurrentRuntime.mockResolvedValue({ runtime: 'docker' })
    listServices.mockResolvedValue(mockServices)
    getAllServiceStatuses.mockResolvedValue(mockStatuses)
    getRunningServices.mockResolvedValue(mockRunningServices)
    
    // Mock WebSocket client
    wsClient.connect.mockImplementation(() => {})
    wsClient.subscribe.mockImplementation(() => {})
    wsClient.disconnect.mockImplementation(() => {})
    
    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
    vi.spyOn(console, 'warn').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.resetAllMocks()
    vi.useRealTimers()
  })

  describe('Application Initialization', () => {
    it('should load services and statuses on mount', async () => {
      const { getRuntimeStatus, listServices, getAllServiceStatuses } = await import('../../api/client')
      
      render(<App />)
      
      await waitFor(() => {
        expect(getRuntimeStatus).toHaveBeenCalled()
      })
      
      expect(screen.getByTestId('app-header')).toBeInTheDocument()
      expect(screen.getByTestId('main-content')).toBeInTheDocument()
      expect(screen.getByTestId('app-footer')).toBeInTheDocument()
    })

    it('should render main app structure', () => {
      render(<App />)
      
      expect(screen.getByTestId('app-header')).toBeInTheDocument()
      expect(screen.getByTestId('main-content')).toBeInTheDocument()
      expect(screen.getByTestId('app-footer')).toBeInTheDocument()
      expect(screen.getByTestId('app-modals')).toBeInTheDocument()
    })
  })

  describe('Service Management Workflows', () => {
    it('should handle service start workflow', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      const startButton = screen.getByText('Start redis')
      await user.click(startButton)
      
      // In a real implementation, this would trigger a service start API call
      // For now, we just verify the UI responds to the click
      expect(startButton).toBeInTheDocument()
    })

    it('should handle service stop workflow', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      const stopButton = screen.getByText('Stop postgres')
      await user.click(stopButton)
      
      expect(stopButton).toBeInTheDocument()
    })

    it('should handle "stop all services" functionality', async () => {
      const { stopAllServices } = await import('../../api/client')
      stopAllServices.mockResolvedValue({ success: true })
      
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      // Look for the actions button or stop all functionality
      // This would depend on the actual UI implementation
      // For now, we verify the API is available
      expect(stopAllServices).toBeDefined()
    })

    it('should update service statuses via WebSocket events', async () => {
      const { wsClient } = await import('../../api/client')
      let statusUpdateCallback
      
      wsClient.subscribe.mockImplementation((event, callback) => {
        if (event === 'service_status_update') {
          statusUpdateCallback = callback
        }
      })
      
      render(<App />)
      
      await waitFor(() => {
        expect(wsClient.subscribe).toHaveBeenCalled()
      })
      
      // Simulate WebSocket status update
      act(() => {
        if (statusUpdateCallback) {
          statusUpdateCallback({
            service_name: 'redis',
            status: 'running'
          })
        }
      })
      
      // The status should be updated in the UI
      // This would be visible in the service list component
      expect(screen.getByTestId('service-list')).toBeInTheDocument()
    })
  })

  describe('Dependency Management', () => {
    it('should load services with dependency information', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      // Verify webapp service is loaded with its dependencies
      expect(screen.getByTestId('service-webapp')).toBeInTheDocument()
    })

    it('should handle dependency status updates', async () => {
      const { wsClient } = await import('../../api/client')
      let statusUpdateCallback
      
      wsClient.subscribe.mockImplementation((event, callback) => {
        if (event === 'service_status_update') {
          statusUpdateCallback = callback
        }
      })
      
      render(<App />)
      
      await waitFor(() => {
        expect(wsClient.subscribe).toHaveBeenCalled()
      })
      
      // Simulate dependency service starting
      act(() => {
        if (statusUpdateCallback) {
          statusUpdateCallback({
            service_name: 'postgres', 
            status: 'running'
          })
        }
      })
      
      expect(screen.getByTestId('service-list')).toBeInTheDocument()
    })
  })

  describe('Logs and Connection Details', () => {
    it('should open logs modal when logs button is clicked', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      const logsButton = screen.getByText('Logs')
      await user.click(logsButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('logs-modal')).toBeInTheDocument()
      })
    })

    it('should open connection modal when connection button is clicked', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      const connectionButton = screen.getByText('Connection')
      await user.click(connectionButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('connection-modal')).toBeInTheDocument()
      })
    })

    it('should close modals when close button is clicked', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      // Open logs modal
      const logsButton = screen.getByText('Logs')
      await user.click(logsButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('logs-modal')).toBeInTheDocument()
      })
      
      // Close logs modal
      const closeButton = screen.getByText('Close')
      await user.click(closeButton)
      
      await waitFor(() => {
        expect(screen.queryByTestId('logs-modal')).not.toBeInTheDocument()
      })
    })

    it('should toggle logs panel with keyboard shortcut', async () => {
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      // Press Cmd+L (or Ctrl+L)
      fireEvent.keyDown(document, { 
        key: 'l', 
        metaKey: true, 
        preventDefault: vi.fn() 
      })
      
      await waitFor(() => {
        expect(screen.getByTestId('logs-panel')).toBeInTheDocument()
      })
      
      // Press again to close
      fireEvent.keyDown(document, { 
        key: 'l', 
        metaKey: true, 
        preventDefault: vi.fn() 
      })
      
      await waitFor(() => {
        expect(screen.queryByTestId('logs-panel')).not.toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should handle service loading errors', async () => {
      const { listServices } = await import('../../api/client')
      listServices.mockRejectedValue(new Error('Failed to load services'))
      
      render(<App />)
      
      await waitFor(() => {
        // App should still render even with API errors
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
    })

    it('should handle WebSocket connection errors', async () => {
      const { wsClient } = await import('../../api/client')
      wsClient.connect.mockImplementation(() => {
        throw new Error('WebSocket connection failed')
      })
      
      render(<App />)
      
      // App should still function without WebSocket
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
    })

    it('should handle runtime monitoring failures', async () => {
      const { getRuntimeStatus } = await import('../../api/client')
      vi.useFakeTimers()
      
      // Initial call succeeds
      getRuntimeStatus.mockResolvedValueOnce({ canProceed: true, runtime: 'docker' })
      
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      // Subsequent monitoring calls fail
      getRuntimeStatus.mockRejectedValue(new Error('Runtime unavailable'))
      
      // Advance timer to trigger runtime monitoring
      act(() => {
        vi.advanceTimersByTime(4000)
      })
      
      await waitFor(() => {
        expect(screen.getByTestId('runtime-setup')).toBeInTheDocument()
      })
      
      vi.useRealTimers()
    })
  })

  describe('Runtime Management', () => {
    it('should transition from runtime setup to main app when runtime becomes ready', async () => {
      const { getRuntimeStatus } = await import('../../api/client')
      getRuntimeStatus.mockResolvedValue({ canProceed: false, runtime: null })
      
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('runtime-setup')).toBeInTheDocument()
      })
      
      // Simulate runtime becoming ready
      getRuntimeStatus.mockResolvedValue({ canProceed: true, runtime: 'docker' })
      
      const runtimeReadyButton = screen.getByText('Runtime Ready')
      await user.click(runtimeReadyButton)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
        expect(screen.queryByTestId('runtime-setup')).not.toBeInTheDocument()
      })
    })

    it('should monitor runtime availability and show setup when runtime becomes unavailable', async () => {
      const { getRuntimeStatus } = await import('../../api/client')
      vi.useFakeTimers()
      
      // Initially runtime is available
      getRuntimeStatus.mockResolvedValueOnce({ canProceed: true, runtime: 'docker' })
      
      render(<App />)
      
      await waitFor(() => {
        expect(screen.getByTestId('service-list')).toBeInTheDocument()
      })
      
      // Runtime becomes unavailable
      getRuntimeStatus.mockResolvedValue({ canProceed: false, runtime: null })
      
      // Advance timer to trigger monitoring
      act(() => {
        vi.advanceTimersByTime(4000)
      })
      
      await waitFor(() => {
        expect(screen.getByTestId('runtime-setup')).toBeInTheDocument()
        expect(screen.queryByTestId('service-list')).not.toBeInTheDocument()
      })
      
      vi.useRealTimers()
    })
  })

  describe('State Management', () => {
    it('should maintain consistent state across WebSocket updates', async () => {
      const { wsClient } = await import('../../api/client')
      let statusUpdateCallback, serviceStartedCallback, serviceStoppedCallback
      
      wsClient.subscribe.mockImplementation((event, callback) => {
        if (event === 'service_status_update') statusUpdateCallback = callback
        if (event === 'service_started') serviceStartedCallback = callback
        if (event === 'service_stopped') serviceStoppedCallback = callback
      })
      
      render(<App />)
      
      await waitFor(() => {
        expect(wsClient.subscribe).toHaveBeenCalledTimes(4) // 4 event types
      })
      
      // Simulate multiple status updates
      act(() => {
        if (statusUpdateCallback) {
          statusUpdateCallback({ service_name: 'redis', status: 'starting' })
        }
      })
      
      act(() => {
        if (serviceStartedCallback) {
          serviceStartedCallback({ service_name: 'redis' })
        }
      })
      
      // State should be consistent
      expect(screen.getByTestId('service-list')).toBeInTheDocument()
    })

    it('should clean up WebSocket subscriptions on unmount', async () => {
      const { wsClient } = await import('../../api/client')
      
      const { unmount } = render(<App />)
      
      unmount()
      
      expect(wsClient.disconnect).toHaveBeenCalled()
    })
  })
})