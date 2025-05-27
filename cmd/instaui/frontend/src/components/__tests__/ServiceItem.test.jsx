import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import ServiceItem from '../ServiceItem'

// Mock the Wails functions
vi.mock('../../wailsjs/go/main/App', () => ({
  StartService: vi.fn(),
  StopService: vi.fn(),
  GetServiceConnectionInfo: vi.fn(),
  CheckImageExists: vi.fn(),
  GetDependencyStatus: vi.fn(),
  GetImageInfo: vi.fn(),
}))

describe('ServiceItem', () => {
  const mockService = {
    name: 'postgres',
    type: 'Database',
    status: 'stopped',
    dependencies: []
  }

  const defaultProps = {
    service: mockService,
    onServiceStateChange: vi.fn(),
    onShowDependencyGraph: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('renders service name and type', () => {
    render(<ServiceItem {...defaultProps} />)
    
    expect(screen.getByText('postgres')).toBeInTheDocument()
    expect(screen.getByText('Database')).toBeInTheDocument()
  })

  test('shows start button when service is stopped', () => {
    render(<ServiceItem {...defaultProps} />)
    
    expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument()
  })

  test('shows stop button when service is running', () => {
    const runningService = { ...mockService, status: 'running' }
    render(<ServiceItem {...defaultProps} service={runningService} />)
    
    expect(screen.getByRole('button', { name: /stop/i })).toBeInTheDocument()
  })

  test('shows logs button', () => {
    render(<ServiceItem {...defaultProps} />)
    
    expect(screen.getByRole('button', { name: /logs/i })).toBeInTheDocument()
  })

  test('shows persist toggle button', () => {
    render(<ServiceItem {...defaultProps} />)
    
    expect(screen.getByRole('button', { name: /persist/i })).toBeInTheDocument()
  })

  test('can expand and collapse details', async () => {
    const user = userEvent.setup()
    render(<ServiceItem {...defaultProps} />)
    
    const expandButton = screen.getByRole('button', { name: /expand details/i })
    expect(expandButton).toHaveAttribute('aria-expanded', 'false')
    
    await user.click(expandButton)
    expect(expandButton).toHaveAttribute('aria-expanded', 'true')
  })

  test('displays correct status styling for stopped service', () => {
    render(<ServiceItem {...defaultProps} />)
    
    const serviceItem = screen.getByText('postgres').closest('.service-item')
    expect(serviceItem).toHaveClass('service-item-stopped')
  })

  test('displays correct status styling for running service', () => {
    const runningService = { ...mockService, status: 'running' }
    render(<ServiceItem {...defaultProps} service={runningService} />)
    
    const serviceItem = screen.getByText('postgres').closest('.service-item')
    expect(serviceItem).toHaveClass('service-item-running')
  })
}) 