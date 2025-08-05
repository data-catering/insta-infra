import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import MainContent from '../MainContent'

// Mock the child components
vi.mock('../ServiceList', () => ({
  default: ({ services }) => (
    <div data-testid="service-list">
      Service List: {services.length} services
    </div>
  )
}))

vi.mock('../RunningServices', () => ({
  default: ({ services }) => (
    <div data-testid="running-services">
      Running Services: {services.length} services
    </div>
  )
}))

vi.mock('../CustomServicesList', () => ({
  default: () => <div data-testid="custom-services">Custom Services</div>
}))

vi.mock('../ServiceItem', () => ({
  ImageStatusProvider: ({ children }) => children,
  ImageLoader: () => <div>Image Loader</div>
}))

describe('MainContent', () => {
  const mockServices = [
    { Name: 'postgres', Type: 'builtin', Dependencies: [] },
    { Name: 'redis', Type: 'builtin', Dependencies: [] }
  ]

  const mockProps = {
    error: null,
    setError: vi.fn(),
    isLoading: false,
    services: mockServices,
    runningServices: [],
    statuses: {},
    dependencyStatuses: {},
    currentRuntime: 'docker',
    fetchAllServices: vi.fn(),
    isAnimating: false
  }

  it('should render service list when services are available', () => {
    render(<MainContent {...mockProps} />)
    
    expect(screen.getByTestId('service-list')).toBeInTheDocument()
    expect(screen.getByText('Service List: 2 services')).toBeInTheDocument()
  })

  it('should render running services when available', () => {
    const runningServices = [mockServices[0]]
    render(<MainContent {...mockProps} runningServices={runningServices} />)
    
    expect(screen.getByTestId('running-services')).toBeInTheDocument()
    expect(screen.getByText('Running Services: 1 services')).toBeInTheDocument()
  })

  it('should show error alert when error exists', () => {
    render(<MainContent {...mockProps} error="Connection failed" />)
    
    expect(screen.getByText('Connection Error')).toBeInTheDocument()
    expect(screen.getByText('Connection failed')).toBeInTheDocument()
  })

  it('should close error when close button is clicked', () => {
    render(<MainContent {...mockProps} error="Connection failed" />)
    
    const closeButton = screen.getByRole('button')
    fireEvent.click(closeButton)
    
    expect(mockProps.setError).toHaveBeenCalledWith(null)
  })

  it('should show loading state when loading and no services', () => {
    render(<MainContent {...mockProps} isLoading={true} services={[]} />)
    
    expect(screen.getByText('Loading infrastructure services...')).toBeInTheDocument()
    expect(screen.getByText('Discovering available services and their status')).toBeInTheDocument()
  })

  it('should show empty state when no services and not loading', () => {
    render(<MainContent {...mockProps} isLoading={false} services={[]} />)
    
    expect(screen.getByText('No services found')).toBeInTheDocument()
    expect(screen.getByText('There may be a configuration issue or services are not properly set up')).toBeInTheDocument()
  })

  it('should call fetchAllServices when retry button is clicked', () => {
    render(<MainContent {...mockProps} isLoading={false} services={[]} />)
    
    const retryButton = screen.getByText('Retry')
    fireEvent.click(retryButton)
    
    expect(mockProps.fetchAllServices).toHaveBeenCalled()
  })

  it('should render custom services list', () => {
    render(<MainContent {...mockProps} />)
    
    expect(screen.getByTestId('custom-services')).toBeInTheDocument()
  })
})