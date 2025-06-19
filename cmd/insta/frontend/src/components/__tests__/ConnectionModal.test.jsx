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
    webUrls: [
      {
        name: 'Web Interface',
        url: 'http://localhost:8080',
        description: 'Web Interface'
      }
    ],
    credentials: [
      {
        description: 'Username',
        value: 'postgres'
      },
      {
        description: 'Password',
        value: 'password'
      }
    ],
    connectionStrings: [
      {
        description: 'Connection Command',
        connectionString: 'psql -h localhost -p 5432 -U postgres'
      }
    ],
    exposedPorts: [
      {
        host_port: '5432',
        container_port: '5432',
        protocol: 'tcp',
        type: 'DATABASE',
        description: 'PostgreSQL database port'
      }
    ]
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
    
    expect(screen.getByText('Web URLs')).toBeInTheDocument()
    expect(screen.getByText('http://localhost:8080')).toBeInTheDocument()
  })

  test('displays credentials when available', () => {
    render(<ConnectionModal {...defaultProps} />)
    
    expect(screen.getByText('Credentials')).toBeInTheDocument()
    expect(screen.getByText('Username')).toBeInTheDocument()
    expect(screen.getByText('postgres')).toBeInTheDocument()
    expect(screen.getByText('Password')).toBeInTheDocument()
    // Password should be masked by default
    expect(screen.getByText('••••••••')).toBeInTheDocument()
  })

  test('toggles password visibility when eye button is clicked', async () => {
    const user = userEvent.setup()
    
    render(<ConnectionModal {...defaultProps} />)
    
    // Initially password should be masked
    expect(screen.getByText('••••••••')).toBeInTheDocument()
    expect(screen.queryByText('password')).not.toBeInTheDocument()
    
    // Find and click the eye button
    const eyeButton = screen.getByTitle('Show password')
    await user.click(eyeButton)
    
    // Password should now be visible
    expect(screen.getByText('password')).toBeInTheDocument()
    expect(screen.queryByText('••••••••')).not.toBeInTheDocument()
    
    // Button should now show "Hide password"
    expect(screen.getByTitle('Hide password')).toBeInTheDocument()
    
    // Click again to hide
    const hideButton = screen.getByTitle('Hide password')
    await user.click(hideButton)
    
    // Password should be masked again
    expect(screen.getByText('••••••••')).toBeInTheDocument()
    expect(screen.queryByText('password')).not.toBeInTheDocument()
  })

  test('displays connection command when available', () => {
    render(<ConnectionModal {...defaultProps} />)
    
    expect(screen.getByText('Connection Strings')).toBeInTheDocument()
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
      webUrls: []
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={noWebConnectionInfo} />)
    
    expect(screen.queryByText('Web URLs')).not.toBeInTheDocument()
  })

  test('does not display credentials section when not available', () => {
    const noCredsConnectionInfo = {
      ...mockConnectionInfo,
      credentials: []
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={noCredsConnectionInfo} />)
    
    expect(screen.queryByText('Credentials')).not.toBeInTheDocument()
  })

  test('does not display connection command when not available', () => {
    const noCommandConnectionInfo = {
      ...mockConnectionInfo,
      connectionStrings: []
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={noCommandConnectionInfo} />)
    
    expect(screen.queryByText('Connection Strings')).not.toBeInTheDocument()
  })

  test('displays port mappings with enhanced format', () => {
    const enhancedConnectionInfo = {
      serviceName: 'postgres',
      available: true,
      exposedPorts: [
        {
          host_port: '5432',
          container_port: '5432',
          protocol: 'tcp',
          type: 'DATABASE',
          description: 'PostgreSQL database port'
        }
      ],
      connectionStrings: [
        {
          description: 'Container Connection',
          connectionString: 'docker exec -it postgres sh -c "PGPASSWORD=postgres psql -Upostgres"',
          note: 'Run this command in your terminal'
        }
      ],
      credentials: [
        {
          description: 'Default Username',
          value: 'postgres',
          note: 'Default username for authentication'
        },
        {
          description: 'Default Password',
          value: 'postgres',
          note: 'Default password for authentication'
        }
      ]
    }
    
    render(<ConnectionModal {...defaultProps} connectionInfo={enhancedConnectionInfo} />)
    
         // Check port mappings table
     expect(screen.getByText('Port Mappings')).toBeInTheDocument()
     expect(screen.getByText('Container Port')).toBeInTheDocument()
     expect(screen.getByText('Host Port')).toBeInTheDocument()
     expect(screen.getAllByText('5432')).toHaveLength(2) // Should appear twice (container and host port)
     expect(screen.getByText('PostgreSQL database port')).toBeInTheDocument()
    
    // Check connection strings
    expect(screen.getByText('Connection Strings')).toBeInTheDocument()
    expect(screen.getByText('Container Connection')).toBeInTheDocument()
    
    // Check credentials with masking
    expect(screen.getByText('Default Username')).toBeInTheDocument()
    expect(screen.getByText('Default Password')).toBeInTheDocument()
    expect(screen.getAllByText('postgres')).toHaveLength(1) // Username should be visible
    expect(screen.getByText('••••••••')).toBeInTheDocument() // Password should be masked
  })
}) 