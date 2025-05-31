import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, test, expect, vi, beforeEach, afterEach } from 'vitest'
import ProgressModal from '../ProgressModal'

// Mock createPortal to render directly
vi.mock('react-dom', () => ({
  createPortal: (element) => element
}))

// Mock the runtime module for this test
vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  EventsOff: vi.fn()
}))

describe('ProgressModal', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    serviceName: 'postgres',
    imageName: 'postgres:13'
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('renders when open', () => {
    render(<ProgressModal {...defaultProps} />)
    
    expect(screen.getByText('Downloading Image - postgres')).toBeInTheDocument()
    expect(screen.getByText('postgres:13')).toBeInTheDocument()
    expect(screen.getByText('Overall Progress')).toBeInTheDocument()
  })

  test('does not render when closed', () => {
    render(<ProgressModal {...defaultProps} isOpen={false} />)
    
    expect(screen.queryByText('Downloading Image - postgres')).not.toBeInTheDocument()
  })

  test('displays service information correctly', () => {
    render(<ProgressModal {...defaultProps} />)
    
    expect(screen.getByText('Service:')).toBeInTheDocument()
    expect(screen.getByText('postgres')).toBeInTheDocument()
    expect(screen.getByText('Image:')).toBeInTheDocument()
    expect(screen.getByText('postgres:13')).toBeInTheDocument()
    expect(screen.getByText('Status:')).toBeInTheDocument()
    expect(screen.getByText('Idle')).toBeInTheDocument()
  })

  test('shows initial progress state', () => {
    render(<ProgressModal {...defaultProps} />)
    
    expect(screen.getByText('0.0%')).toBeInTheDocument()
    expect(screen.getByText('Overall Progress')).toBeInTheDocument()
    expect(screen.getByText('Elapsed')).toBeInTheDocument()
    expect(screen.getByText('0s')).toBeInTheDocument()
  })

  test('has cancel button', () => {
    render(<ProgressModal {...defaultProps} />)
    
    expect(screen.getByText('Cancel Download')).toBeInTheDocument()
  })

  test('calls onClose when cancel button is clicked', async () => {
    const user = userEvent.setup()
    const onCloseMock = vi.fn()
    
    render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
    
    const cancelButton = screen.getByText('Cancel Download')
    await user.click(cancelButton)
    
    expect(onCloseMock).toHaveBeenCalledTimes(1)
  })

  test('calls onClose when close button is clicked', async () => {
    const user = userEvent.setup()
    const onCloseMock = vi.fn()
    
    render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
    
    const closeButton = document.querySelector('.connection-modal-close')
    await user.click(closeButton)
    
    expect(onCloseMock).toHaveBeenCalledTimes(1)
  })

  test('calls onClose when clicking overlay', async () => {
    const user = userEvent.setup()
    const onCloseMock = vi.fn()
    
    render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
    
    const overlay = document.querySelector('.connection-modal-overlay')
    await user.click(overlay)
    
    expect(onCloseMock).toHaveBeenCalledTimes(1)
  })

  test('does not close when clicking modal content', async () => {
    const user = userEvent.setup()
    const onCloseMock = vi.fn()
    
    render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
    
    const modal = document.querySelector('.connection-modal')
    await user.click(modal)
    
    expect(onCloseMock).not.toHaveBeenCalled()
  })

  test('renders without image name', () => {
    render(<ProgressModal {...defaultProps} imageName="" />)
    
    expect(screen.getByText('Downloading Image - postgres')).toBeInTheDocument()
    expect(screen.getByText('Service:')).toBeInTheDocument()
    expect(screen.queryByText('Image:')).not.toBeInTheDocument()
  })
}) 