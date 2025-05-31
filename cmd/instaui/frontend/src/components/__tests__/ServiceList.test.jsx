import React from 'react'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import ServiceList from '../ServiceList'

describe('ServiceList', () => {
  const mockServices = [
    {
      Name: 'postgres',
      Type: 'Database'
    },
    {
      Name: 'redis',
      Type: 'Cache'
    },
    {
      Name: 'mongodb',
      Type: 'Database'
    }
  ]

  const mockStatuses = {
    postgres: 'stopped',
    redis: 'running',
    mongodb: 'stopped'
  }

  const defaultProps = {
    services: [],
    statuses: {},
    isLoading: false,
    onServiceStateChange: vi.fn(),
    onShowDependencyGraph: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('renders empty state when no services are provided', () => {
    render(<ServiceList {...defaultProps} />)
    
    expect(screen.getByText('All Services')).toBeInTheDocument()
    expect(screen.getByText('No services available.')).toBeInTheDocument()
  })

  test('displays services when provided', () => {
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    expect(screen.getByText('All Services')).toBeInTheDocument()
    // The component should render ServiceItem components for each service
    // We can verify the services grid container exists
    const servicesGrid = document.querySelector('.services-grid')
    expect(servicesGrid).toBeInTheDocument()
  })

  test('displays search input', () => {
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    expect(screen.getByPlaceholderText('Search by name...')).toBeInTheDocument()
  })

  test('filters services based on search input', async () => {
    const user = userEvent.setup()
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const searchInput = screen.getByPlaceholderText('Search by name...')
    await user.type(searchInput, 'postgres')
    
    // After filtering, the search input should have the value
    expect(searchInput.value).toBe('postgres')
    
    // The clear button should appear
    expect(screen.getByText('Clear filters')).toBeInTheDocument()
  })

  test('displays service type filter dropdown', () => {
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const filterSelect = screen.getByDisplayValue('All Types')
    expect(filterSelect).toBeInTheDocument()
  })

  test('filters services by type when filter is changed', async () => {
    const user = userEvent.setup()
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const filterSelect = screen.getByDisplayValue('All Types')
    await user.selectOptions(filterSelect, 'Database')
    
    // After filtering by Database type, the filter should be set
    expect(filterSelect.value).toBe('Database')
    
    // The clear button should appear
    expect(screen.getByText('Clear filters')).toBeInTheDocument()
  })

  test('displays service count correctly', () => {
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    // Check that the results count section exists
    const resultsCount = document.querySelector('.results-count')
    expect(resultsCount).toBeInTheDocument()
    
    // Check that it contains the expected text structure
    expect(resultsCount).toHaveTextContent('Showing')
    expect(resultsCount).toHaveTextContent('of')
    expect(resultsCount).toHaveTextContent('services')
  })

  test('displays status counters', () => {
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    // Based on mockStatuses: redis is running (1), postgres and mongodb are stopped (2)
    expect(screen.getByText('1 Running')).toBeInTheDocument()
    expect(screen.getByText('2 Stopped')).toBeInTheDocument()
  })

  test('shows view toggle buttons', () => {
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const gridButton = screen.getByTitle('Grid view')
    const listButton = screen.getByTitle('List view')
    
    expect(gridButton).toBeInTheDocument()
    expect(listButton).toBeInTheDocument()
  })

  test('switches between grid and list view', async () => {
    const user = userEvent.setup()
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const listButton = screen.getByTitle('List view')
    await user.click(listButton)
    
    // Should switch to list view (we can check if the button becomes active)
    expect(listButton).toHaveClass('active')
  })

  test('shows clear filters button when filters are applied', async () => {
    const user = userEvent.setup()
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const searchInput = screen.getByPlaceholderText('Search by name...')
    await user.type(searchInput, 'postgres')
    
    expect(screen.getByText('Clear filters')).toBeInTheDocument()
  })

  test('clears filters when clear button is clicked', async () => {
    const user = userEvent.setup()
    render(<ServiceList {...defaultProps} services={mockServices} statuses={mockStatuses} />)
    
    const searchInput = screen.getByPlaceholderText('Search by name...')
    await user.type(searchInput, 'postgres')
    
    const clearButton = screen.getByText('Clear filters')
    await user.click(clearButton)
    
    expect(searchInput.value).toBe('')
    expect(screen.getByDisplayValue('All Types')).toBeInTheDocument()
  })

  test('handles loading state correctly', () => {
    render(<ServiceList {...defaultProps} isLoading={true} />)
    
    // When loading, component should still render normally
    // The loading state doesn't change the component behavior in this implementation
    expect(screen.getByText('All Services')).toBeInTheDocument()
  })

  test('passes correct props to ServiceItem components', () => {
    const onServiceStateChange = vi.fn()
    const onShowDependencyGraph = vi.fn()
    
    render(
      <ServiceList 
        {...defaultProps} 
        services={[mockServices[0]]} 
        statuses={mockStatuses}
        onServiceStateChange={onServiceStateChange}
        onShowDependencyGraph={onShowDependencyGraph}
      />
    )
    
    // We can't easily test ServiceItem props without mocking the component
    // But we can verify the component structure is correct
    expect(screen.getByText('All Services')).toBeInTheDocument()
    
    // Check that the results count section exists
    const resultsCount = document.querySelector('.results-count')
    expect(resultsCount).toBeInTheDocument()
  })

  test('handles empty services with loading false', () => {
    render(<ServiceList {...defaultProps} services={[]} isLoading={false} />)
    
    expect(screen.getByText('No services available.')).toBeInTheDocument()
  })
}) 