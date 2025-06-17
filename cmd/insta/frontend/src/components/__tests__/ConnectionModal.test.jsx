import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import ConnectionModal from '../ConnectionModal'

// Mock createPortal to render directly
vi.mock('react-dom', () => ({
  createPortal: (element) => element
}))

describe('ConnectionModal', () => {
  const mockConnectionInfo = {
    serviceName: 'postgres',
    available: true,
    hasWebUI: true,
    webURL: 'http://localhost:8080',
    username: 'postgres',
    password: 'password',
    connectionCommand: 'psql -h localhost -p 5432 -U postgres'
  }

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    connectionInfo: mockConnectionInfo,
    copyFeedback: null,
    onCopyToClipboard: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  test('does not render when isOpen is false', () => {
    const { container } = render(
      <ConnectionModal {...defaultProps} isOpen={false} />
    )
    
    expect(container.firstChild).toBeNull()
  })

  test('does not render when connectionInfo is null', () => {
    const { container } = render(
      <ConnectionModal {...defaultProps} connectionInfo={null} />
    )
    
    expect(container.firstChild).toBeNull()
  })

  test('renders modal with service name when open', () => {
    render(<ConnectionModal {...defaultProps} />)
    
    expect(screen.getByText(/connection info/i)).toBeInTheDocument()
    expect(document.querySelector('.connection-modal-close')).toBeInTheDocument() // Close button
  })

  test('calls onClose when close button is clicked', async () => {
    const user = userEvent.setup()
    const onCloseMock = vi.fn()
    
    render(<ConnectionModal {...defaultProps} onClose={onCloseMock} />)
    
    const closeButton = document.querySelector('.connection-modal-close')
    await user.click(closeButton)
    
    expect(onCloseMock).toHaveBeenCalledTimes(1)
  })

  test('calls onClose when clicking outside modal', async () => {
    const user = userEvent.setup()
    const onCloseMock = vi.fn()
    
    render(<ConnectionModal {...defaultProps} onClose={onCloseMock} />)
    
    const overlay = document.querySelector('.connection-modal-overlay')
    await user.click(overlay)
    
    expect(onCloseMock).toHaveBeenCalledTimes(1)
  })

  test('displays web URL when available', () => {
    render(<ConnectionModal {...defaultProps} />)
    
    expect(screen.getByText('URL')).toBeInTheDocument()
    expect(screen.getByText('http://localhost:8080')).toBeInTheDocument()
  })

  test('displays credentials when available', () => {
    render(<ConnectionModal {...defaultProps} />)
    
    expect(screen.getByText('Default Credentials')).toBeInTheDocument()
    expect(screen.getByText('Username')).toBeInTheDocument()
    expect(screen.getByText('postgres')).toBeInTheDocument()
    expect(screen.getByText('Password')).toBeInTheDocument()
    // Password should be shown as 'password' in this test
    expect(screen.getByText('password')).toBeInTheDocument()
  })

  test('displays connection command when available', () => {
    render(<ConnectionModal {...defaultProps} />)
    
    expect(screen.getByText('Connection Command')).toBeInTheDocument()
    expect(screen.getByText('psql -h localhost -p 5432 -U postgres')).toBeInTheDocument()
  })

  test('calls onCopyToClipboard when copy buttons are clicked', async () => {
    const user = userEvent.setup()
    const onCopyMock = vi.fn()
    
    render(<ConnectionModal {...defaultProps} onCopyToClipboard={onCopyMock} />)
    
    const copyButtons = screen.getAllByTitle(/copy/i)
    expect(copyButtons.length).toBeGreaterThan(0)
    
    await user.click(copyButtons[0])
    
    expect(onCopyMock).toHaveBeenCalledTimes(1)
  })

  test('displays error when connection is not available', () => {
    const errorConnectionInfo = {
      serviceName: 'postgres',
      available: false,
      error: 'Service is not running'
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={errorConnectionInfo} />)
    
    expect(screen.getByText(/service connection information unavailable/i)).toBeInTheDocument()
    expect(screen.getByText('Service is not running')).toBeInTheDocument()
  })

  test('displays copy feedback when provided', () => {
    render(<ConnectionModal {...defaultProps} copyFeedback="Copied!" />)
    
    expect(screen.getByText('Copied!')).toBeInTheDocument()
  })

  test('does not display web URL section when not available', () => {
    const noWebConnectionInfo = {
      ...mockConnectionInfo,
      hasWebUI: false,
      webURL: null
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={noWebConnectionInfo} />)
    
    expect(screen.queryByText('Web Interface')).not.toBeInTheDocument()
  })

  test('does not display credentials section when not available', () => {
    const noCredsConnectionInfo = {
      ...mockConnectionInfo,
      username: null,
      password: null
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={noCredsConnectionInfo} />)
    
    expect(screen.queryByText('Default Credentials')).not.toBeInTheDocument()
  })

  test('does not display connection command when not available', () => {
    const noCommandConnectionInfo = {
      ...mockConnectionInfo,
      connectionCommand: null
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={noCommandConnectionInfo} />)
    
    expect(screen.queryByText('Connection Command')).not.toBeInTheDocument()
  })
}) 