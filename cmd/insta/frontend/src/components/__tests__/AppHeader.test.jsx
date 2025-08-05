import { render, screen, fireEvent } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import AppHeader from '../AppHeader'

describe('AppHeader', () => {
  const mockProps = {
    lastUpdated: new Date('2024-01-01 12:00:00'),
    runningServicesCount: 2,
    isLoading: false,
    hasRunningServices: true,
    isStoppingAll: false,
    handleStopAllServices: vi.fn(),
    showActionsDropdown: false,
    setShowActionsDropdown: vi.fn(),
    setShowAbout: vi.fn(),
    handleRestartRuntime: vi.fn(),
    handleShutdownApplication: vi.fn(),
    currentRuntime: 'docker'
  }

  it('should render logo and title', () => {
    render(<AppHeader {...mockProps} />)
    
    expect(screen.getByText('Insta')).toBeInTheDocument()
    expect(screen.getByText('Infra')).toBeInTheDocument()
  })

  it('should display running services count', () => {
    render(<AppHeader {...mockProps} />)
    
    expect(screen.getByText('2 services running')).toBeInTheDocument()
  })

  it('should show stop all button when there are running services', () => {
    render(<AppHeader {...mockProps} />)
    
    expect(screen.getByTitle('Stop all running services')).toBeInTheDocument()
  })

  it('should not show stop all button when no running services', () => {
    render(<AppHeader {...mockProps} hasRunningServices={false} />)
    
    expect(screen.queryByTitle('Stop all running services')).not.toBeInTheDocument()
  })

  it('should call handleStopAllServices when stop all button is clicked', () => {
    render(<AppHeader {...mockProps} />)
    
    const stopAllButton = screen.getByTitle('Stop all running services')
    fireEvent.click(stopAllButton)
    
    expect(mockProps.handleStopAllServices).toHaveBeenCalled()
  })

  it('should show loading state when stopping all services', () => {
    render(<AppHeader {...mockProps} isStoppingAll={true} />)
    
    expect(screen.getByText('Stopping All...')).toBeInTheDocument()
  })

  it('should show syncing indicator when loading', () => {
    render(<AppHeader {...mockProps} isLoading={true} />)
    
    expect(screen.getByText('Syncing...')).toBeInTheDocument()
  })

  it('should show actions dropdown when clicked', () => {
    render(<AppHeader {...mockProps} showActionsDropdown={true} />)
    
    expect(screen.getByText('About')).toBeInTheDocument()
    expect(screen.getByText('Restart')).toBeInTheDocument()
    expect(screen.getByText('Shutdown')).toBeInTheDocument()
  })

  it('should toggle actions dropdown', () => {
    render(<AppHeader {...mockProps} />)
    
    const actionsButton = screen.getByTitle('More actions')
    fireEvent.click(actionsButton)
    
    expect(mockProps.setShowActionsDropdown).toHaveBeenCalledWith(true)
  })
})