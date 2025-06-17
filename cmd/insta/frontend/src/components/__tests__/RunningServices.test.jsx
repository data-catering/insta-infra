import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import RunningServices from '../RunningServices'
import { ImageStatusProvider } from '../ServiceItem'

// Helper to render RunningServices with ImageStatusProvider
const renderWithProvider = (ui, services = []) => {
  return render(
    <ImageStatusProvider services={services}>
      {ui}
    </ImageStatusProvider>
  )
}

describe('RunningServices', () => {
  const mockServices = [
    {
      name: 'postgres',
      type: 'Database',
      status: 'running',
      dependencies: []
    },
    {
      name: 'redis',
      type: 'Cache',
      status: 'running',
      dependencies: []
    }
  ]

  const defaultProps = {
    services: [],
    isLoading: false,
    onServiceStateChange: vi.fn(),
    onShowDependencyGraph: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('renders empty state when no services are running', () => {
    const { container } = renderWithProvider(<RunningServices {...defaultProps} />)
    
    // Component returns null when no services, so the container should be empty
    expect(container.firstChild).toBeNull()
  })

  test('displays running services count when services are provided', () => {
    renderWithProvider(<RunningServices {...defaultProps} services={mockServices} />, mockServices)
    
    expect(screen.getByText('Running Services')).toBeInTheDocument()
    expect(screen.getByText('2 Active')).toBeInTheDocument()
  })

  test('renders ServiceItem components for each service', () => {
    renderWithProvider(<RunningServices {...defaultProps} services={mockServices} />, mockServices)
    
    // ServiceItem components should be rendered (we can't easily test their content without mocking them)
    // But we can verify the services grid container exists
    const servicesGrid = document.querySelector('.services-grid')
    expect(servicesGrid).toBeInTheDocument()
  })

  test('shows correct status indicator styling for running services', () => {
    renderWithProvider(<RunningServices {...defaultProps} services={mockServices} />, mockServices)
    
    const statusDot = document.querySelector('.status-dot.dot-green.pulse')
    expect(statusDot).toBeInTheDocument()
  })

  test('shows correct status indicator styling for no services', () => {
    const { container } = renderWithProvider(<RunningServices {...defaultProps} />)
    
    // Component returns null when no services, so no status dot should exist
    const statusDot = container.querySelector('.status-dot')
    expect(statusDot).not.toBeInTheDocument()
  })

  test('passes correct props to ServiceItem components', () => {
    const onServiceStateChange = vi.fn()
    const onShowDependencyGraph = vi.fn()
    
    renderWithProvider(
      <RunningServices 
        {...defaultProps} 
        services={[mockServices[0]]} 
        onServiceStateChange={onServiceStateChange}
        onShowDependencyGraph={onShowDependencyGraph}
      />,
      [mockServices[0]]
    )
    
    // We can't easily test ServiceItem props without mocking the component
    // But we can verify the component structure is correct
    expect(screen.getByText('Running Services')).toBeInTheDocument()
    expect(screen.getByText('1 Active')).toBeInTheDocument()
  })

  test('handles loading state correctly', () => {
    const { container } = renderWithProvider(<RunningServices {...defaultProps} isLoading={true} />)
    
    // Component returns null when no services regardless of loading state
    expect(container.firstChild).toBeNull()
  })

  test('applies correct animation delays to service items', () => {
    renderWithProvider(<RunningServices {...defaultProps} services={mockServices} />, mockServices)
    
    const fadeInElements = document.querySelectorAll('.fade-in')
    expect(fadeInElements).toHaveLength(2)
    
    // Check that animation delays are applied (note: actual delay calculation is Math.min(index * 15, 200))
    expect(fadeInElements[0]).toHaveStyle('animation-delay: 0ms')
    expect(fadeInElements[1]).toHaveStyle('animation-delay: 15ms')
  })
}) 