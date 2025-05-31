import React from 'react'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import RuntimeSetup from '../RuntimeSetup'

// Mock the Wails functions
vi.mock('../../../wailsjs/go/main/App', () => ({
  GetRuntimeStatus: vi.fn(),
  AttemptStartRuntime: vi.fn(),
  WaitForRuntimeReady: vi.fn(),
  ReinitializeRuntime: vi.fn(),
}))

import { 
  GetRuntimeStatus, 
  AttemptStartRuntime, 
  WaitForRuntimeReady, 
  ReinitializeRuntime 
} from '../../../wailsjs/go/main/App'

describe('RuntimeSetup', () => {
  const mockRuntimeStatus = {
    platform: 'darwin',
    runtimeStatuses: [
      {
        name: 'docker',
        isInstalled: false,
        isRunning: false,
        isAvailable: false,
        canAutoStart: false,
        version: '',
        error: ''
      },
      {
        name: 'podman',
        isInstalled: false,
        isRunning: false,
        isAvailable: false,
        canAutoStart: false,
        version: '',
        error: ''
      }
    ]
  }

  const defaultProps = {
    onRuntimeReady: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
    GetRuntimeStatus.mockResolvedValue(mockRuntimeStatus)
    // Mock window.open for installation links
    global.window.open = vi.fn()
  })

  test('renders loading state initially', () => {
    GetRuntimeStatus.mockImplementation(() => new Promise(() => {})) // Never resolves
    
    render(<RuntimeSetup {...defaultProps} />)
    
    expect(screen.getByText('Checking container runtime...')).toBeInTheDocument()
  })

  test('displays runtime setup interface when no runtime is available', async () => {
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Container Runtime Setup')).toBeInTheDocument()
    })
    
    expect(screen.getByText(/insta-infra requires Docker or Podman/)).toBeInTheDocument()
    expect(screen.getByText('Quick Start')).toBeInTheDocument()
    expect(screen.getByText('Install Docker')).toBeInTheDocument()
  })

  test('shows install Docker button when Docker is not installed', async () => {
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Install Docker')).toBeInTheDocument()
    })
    
    expect(screen.getByText('We recommend Docker for the best experience. Click below to get started quickly.')).toBeInTheDocument()
  })

  test('opens installation link when install button is clicked', async () => {
    const user = userEvent.setup()
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Install Docker')).toBeInTheDocument()
    })
    
    const installButton = screen.getByText('Install Docker')
    await user.click(installButton)
    
    expect(global.window.open).toHaveBeenCalledWith(
      expect.stringContaining('docker.com'),
      '_blank'
    )
  })

  test('shows start button when Docker is installed but not running', async () => {
    const installedButStoppedStatus = {
      ...mockRuntimeStatus,
      runtimeStatuses: [
        {
          name: 'docker',
          isInstalled: true,
          isRunning: false,
          isAvailable: false,
          canAutoStart: true,
          version: '20.10.0',
          error: ''
        }
      ]
    }
    
    GetRuntimeStatus.mockResolvedValue(installedButStoppedStatus)
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Start Docker')).toBeInTheDocument()
    })
  })

  test('attempts to start runtime when start button is clicked', async () => {
    const user = userEvent.setup()
    const installedButStoppedStatus = {
      ...mockRuntimeStatus,
      runtimeStatuses: [
        {
          name: 'docker',
          isInstalled: true,
          isRunning: false,
          isAvailable: false,
          canAutoStart: true,
          version: '20.10.0',
          error: ''
        }
      ]
    }
    
    GetRuntimeStatus.mockResolvedValue(installedButStoppedStatus)
    
    // Create a promise that we can control
    let resolveAttemptStart
    const attemptStartPromise = new Promise((resolve) => {
      resolveAttemptStart = resolve
    })
    AttemptStartRuntime.mockReturnValue(attemptStartPromise)
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Start Docker')).toBeInTheDocument()
    })
    
    const startButton = screen.getByText('Start Docker')
    await user.click(startButton)
    
    expect(AttemptStartRuntime).toHaveBeenCalledWith('docker')
    
    // The component should show starting state immediately
    expect(screen.getByText('Starting docker...')).toBeInTheDocument()
    
    // Now resolve the promise to complete the test
    resolveAttemptStart({ success: true, message: 'Docker started successfully' })
  })

  test('shows success message when runtime starts successfully', async () => {
    const user = userEvent.setup()
    const installedButStoppedStatus = {
      ...mockRuntimeStatus,
      runtimeStatuses: [
        {
          name: 'docker',
          isInstalled: true,
          isRunning: false,
          isAvailable: false,
          canAutoStart: true,
          version: '20.10.0',
          error: ''
        }
      ]
    }
    
    GetRuntimeStatus.mockResolvedValue(installedButStoppedStatus)
    AttemptStartRuntime.mockResolvedValue({ success: true, message: 'Docker started successfully' })
    WaitForRuntimeReady.mockResolvedValue({ success: true })
    ReinitializeRuntime.mockResolvedValue()
    
    const onRuntimeReady = vi.fn()
    render(<RuntimeSetup {...defaultProps} onRuntimeReady={onRuntimeReady} />)
    
    await waitFor(() => {
      expect(screen.getByText('Start Docker')).toBeInTheDocument()
    })
    
    const startButton = screen.getByText('Start Docker')
    await user.click(startButton)
    
    await waitFor(() => {
      expect(screen.getByText('Success! insta-infra is ready to use.')).toBeInTheDocument()
    })
    
    expect(ReinitializeRuntime).toHaveBeenCalled()
    // onRuntimeReady is called with a timeout, so we need to wait
    await waitFor(() => {
      expect(onRuntimeReady).toHaveBeenCalled()
    }, { timeout: 2000 })
  })

  test('shows error message when runtime fails to start', async () => {
    const user = userEvent.setup()
    const installedButStoppedStatus = {
      ...mockRuntimeStatus,
      runtimeStatuses: [
        {
          name: 'docker',
          isInstalled: true,
          isRunning: false,
          isAvailable: false,
          canAutoStart: true,
          version: '20.10.0',
          error: ''
        }
      ]
    }
    
    GetRuntimeStatus.mockResolvedValue(installedButStoppedStatus)
    AttemptStartRuntime.mockResolvedValue({ 
      success: false, 
      error: 'Failed to start Docker',
      requiresManualAction: true,
      message: 'Please start Docker Desktop manually'
    })
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Start Docker')).toBeInTheDocument()
    })
    
    const startButton = screen.getByText('Start Docker')
    await user.click(startButton)
    
    await waitFor(() => {
      expect(screen.getByText('Manual action required: Please start Docker Desktop manually')).toBeInTheDocument()
    })
  })

  test('shows refresh button and allows retrying', async () => {
    const user = userEvent.setup()
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Refresh Status')).toBeInTheDocument()
    })
    
    const refreshButton = screen.getByText('Refresh Status')
    await user.click(refreshButton)
    
    expect(GetRuntimeStatus).toHaveBeenCalledTimes(2)
  })

  test('handles runtime status fetch error', async () => {
    GetRuntimeStatus.mockRejectedValue(new Error('Failed to get runtime status'))
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Runtime Check Failed')).toBeInTheDocument()
    })
    
    expect(screen.getByText('Failed to determine container runtime availability.')).toBeInTheDocument()
  })

  test('shows advanced options toggle', async () => {
    const user = userEvent.setup()
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Advanced Options')).toBeInTheDocument()
    })
    
    const advancedButton = screen.getByText('Advanced Options')
    await user.click(advancedButton)
    
    // Should expand to show more options
    await waitFor(() => {
      expect(screen.getByText('Docker')).toBeInTheDocument()
    })
  })

  test('shows manual start option when Docker cannot auto-start', async () => {
    const manualStartStatus = {
      ...mockRuntimeStatus,
      runtimeStatuses: [
        {
          name: 'docker',
          isInstalled: true,
          isRunning: false,
          isAvailable: false,
          canAutoStart: false,
          version: '20.10.0',
          error: ''
        }
      ]
    }
    
    GetRuntimeStatus.mockResolvedValue(manualStartStatus)
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Start Docker Manually')).toBeInTheDocument()
    })
  })

  test('shows ready state when Docker is available', async () => {
    const readyStatus = {
      ...mockRuntimeStatus,
      runtimeStatuses: [
        {
          name: 'docker',
          isInstalled: true,
          isRunning: true,
          isAvailable: true,
          canAutoStart: true,
          version: '20.10.0',
          error: ''
        }
      ]
    }
    
    GetRuntimeStatus.mockResolvedValue(readyStatus)
    
    render(<RuntimeSetup {...defaultProps} />)
    
    await waitFor(() => {
      expect(screen.getByText('Docker Ready')).toBeInTheDocument()
    })
  })
}) 