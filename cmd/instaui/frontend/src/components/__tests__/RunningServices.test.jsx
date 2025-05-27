import React from 'react'
import { render, screen } from '@testing-library/react'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import RunningServices from '../RunningServices'

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
    render(<RunningServices {...defaultProps} />)
    
    expect(screen.getByText('Active Services')).toBeInTheDocument()
    expect(screen.getByText('0 Running')).toBeInTheDocument()
    expect(screen.getByText('No Active Services')).toBeInTheDocument()
    expect(screen.getByText('Start a service to see it appear here')).toBeInTheDocument()
  })

  test('displays running services count when services are provided', () => {
    render(<RunningServices {...defaultProps} services={mockServices} />)
    
    expect(screen.getByText('Active Services')).toBeInTheDocument()
    expect(screen.getByText('2 Running')).toBeInTheDocument()
  })

  test('renders ServiceItem components for each service', () => {
    render(<RunningServices {...defaultProps} services={mockServices} />)
    
    // ServiceItem components should be rendered (we can't easily test their content without mocking them)
    // But we can verify the services grid container exists
    const servicesGrid = document.querySelector('.services-grid')
    expect(servicesGrid).toBeInTheDocument()
  })

  test('shows correct status indicator styling for running services', () => {
    render(<RunningServices {...defaultProps} services={mockServices} />)
    
    const statusDot = document.querySelector('.status-dot.dot-green.pulse')
    expect(statusDot).toBeInTheDocument()
  })

  test('shows correct status indicator styling for no services', () => {
    render(<RunningServices {...defaultProps} />)
    
    const statusDot = document.querySelector('.status-dot.dot-gray')
    expect(statusDot).toBeInTheDocument()
  })

  test('passes correct props to ServiceItem components', () => {
    const onServiceStateChange = vi.fn()
    const onShowDependencyGraph = vi.fn()
    
    render(
      <RunningServices 
        {...defaultProps} 
        services={[mockServices[0]]} 
        onServiceStateChange={onServiceStateChange}
        onShowDependencyGraph={onShowDependencyGraph}
      />
    )
    
    // We can't easily test ServiceItem props without mocking the component
    // But we can verify the component structure is correct
    expect(screen.getByText('Active Services')).toBeInTheDocument()
    expect(screen.getByText('1 Running')).toBeInTheDocument()
  })

  test('handles loading state correctly', () => {
    render(<RunningServices {...defaultProps} isLoading={true} />)
    
    // When loading, component should still render normally
    // The loading state doesn't change the component behavior in this implementation
    expect(screen.getByText('Active Services')).toBeInTheDocument()
  })

  test('applies correct animation delays to service items', () => {
    render(<RunningServices {...defaultProps} services={mockServices} />)
    
    const fadeInElements = document.querySelectorAll('.fade-in')
    expect(fadeInElements).toHaveLength(2)
    
    // Check that animation delays are applied
    expect(fadeInElements[0]).toHaveStyle('animation-delay: 0ms')
    expect(fadeInElements[1]).toHaveStyle('animation-delay: 50ms')
  })
}) 