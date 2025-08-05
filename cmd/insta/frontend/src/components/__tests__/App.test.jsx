import { render, screen, waitFor } from '@testing-library/react'
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

vi.mock('../RuntimeSetup', () => ({
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

vi.mock('../../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    initializeWebSocket: vi.fn()
  })
}))

describe('App Component - Basic Tests', () => {
  const mockServices = [
    { name: 'postgres', Name: 'postgres', type: 'builtin', Type: 'builtin', Dependencies: [] },
    { name: 'redis', Name: 'redis', type: 'builtin', Type: 'builtin', Dependencies: [] }
  ]

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
    getAllServiceStatuses.mockResolvedValue({})
    getRunningServices.mockResolvedValue({})
    
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

  it('should render main app structure', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByTestId('app-header')).toBeInTheDocument()
      expect(screen.getByTestId('main-content')).toBeInTheDocument()
      expect(screen.getByTestId('app-footer')).toBeInTheDocument()
      expect(screen.getByTestId('app-modals')).toBeInTheDocument()
    })
  })

  it('should call runtime status check on mount', async () => {
    const { getRuntimeStatus } = await import('../../api/client')
    
    render(<App />)
    
    await waitFor(() => {
      expect(getRuntimeStatus).toHaveBeenCalled()
    })
  })

  it('should render footer with runtime info', async () => {
    render(<App />)
    
    await waitFor(() => {
      expect(screen.getByText('Using docker')).toBeInTheDocument()
    })
  })
})