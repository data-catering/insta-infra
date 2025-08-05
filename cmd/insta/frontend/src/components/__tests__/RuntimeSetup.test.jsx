import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import RuntimeSetup from '../RuntimeSetup'

// Mock the API client
vi.mock('../../api/client', () => ({
  getRuntimeStatus: vi.fn(),
  startRuntime: vi.fn(),
  reinitializeRuntime: vi.fn(),
  setCustomDockerPath: vi.fn(),
  setCustomPodmanPath: vi.fn(),
  getCustomDockerPath: vi.fn(),
  getCustomPodmanPath: vi.fn(),
  getAppLogs: vi.fn(),
  WaitForRuntimeReady: vi.fn(),
  ReinitializeRuntime: vi.fn(),
}))

import * as client from '../../api/client'

describe('RuntimeSetup Component', () => {
  const user = userEvent.setup()

  const mockRuntimeStatusReady = {
    canProceed: true,
    preferredRuntime: 'docker',
    platform: 'darwin',
    runtimeStatuses: [
      {
        name: 'docker',
        isInstalled: true,
        isRunning: true,
        version: '20.10.8'
      }
    ]
  }

  const mockRuntimeStatusNotReady = {
    canProceed: false,
    preferredRuntime: 'docker',
    platform: 'darwin',
    runtimeStatuses: [
      {
        name: 'docker',
        isInstalled: true,
        isRunning: false,
        version: '20.10.8'
      }
    ]
  }

  const defaultProps = {
    onRuntimeReady: vi.fn()
  }

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Setup default mocked client functions
    client.getRuntimeStatus.mockResolvedValue(mockRuntimeStatusNotReady)
    client.getCustomDockerPath.mockResolvedValue('')
    client.getCustomPodmanPath.mockResolvedValue('')
    client.getAppLogs.mockResolvedValue([])
    
    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  describe('Basic Rendering', () => {
    it('should render main setup interface', async () => {
      render(<RuntimeSetup {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByText(/Container Runtime Setup/i)).toBeInTheDocument()
      })
    })

    it('should show quick start section', async () => {
      render(<RuntimeSetup {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByText(/Quick Start/i)).toBeInTheDocument()
      })
    })
  })

  describe('Ready State', () => {
    it('should show ready state when runtime is available', async () => {
      client.getRuntimeStatus.mockResolvedValue(mockRuntimeStatusReady)
      
      render(<RuntimeSetup {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByText(/Container Runtime Ready/i)).toBeInTheDocument()
      })
    })

    it('should show continue button when runtime is ready', async () => {
      client.getRuntimeStatus.mockResolvedValue(mockRuntimeStatusReady)
      
      render(<RuntimeSetup {...defaultProps} />)
      
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /continue to services/i })).toBeInTheDocument()
      })
    })

    it('should call onRuntimeReady when continue button is clicked', async () => {
      client.getRuntimeStatus.mockResolvedValue(mockRuntimeStatusReady)
      const onRuntimeReadySpy = vi.fn()
      
      render(<RuntimeSetup {...defaultProps} onRuntimeReady={onRuntimeReadySpy} />)
      
      await waitFor(() => {
        const continueButton = screen.getByRole('button', { name: /continue to services/i })
        return user.click(continueButton)
      })
      
      expect(onRuntimeReadySpy).toHaveBeenCalledTimes(1)
    })
  })

  describe('Advanced Options', () => {
    it('should toggle advanced options when button is clicked', async () => {
      render(<RuntimeSetup {...defaultProps} />)
      
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      expect(screen.getByText(/Choose Your Container Runtime/i)).toBeInTheDocument()
    })

    it('should show runtime status cards in advanced options', async () => {
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      expect(screen.getByText('Docker')).toBeInTheDocument()
    })
  })

  describe('Application Logs', () => {
    it('should toggle logs visibility', async () => {
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options first
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Toggle logs using the Show button
      const logsButton = screen.getByRole('button', { name: /show/i })
      await user.click(logsButton)
      
      expect(screen.getByText(/No logs available/i)).toBeInTheDocument()
    })
  })

  describe('Help Links', () => {
    it('should show help links in advanced options', async () => {
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      expect(screen.getByText(/Need Help/i)).toBeInTheDocument()
    })
  })

  describe('Trim Error Reproduction', () => {
    it('should handle null custom paths without throwing trim error', async () => {
      // Mock API to return null values (which might cause the trim error)
      client.getCustomDockerPath.mockResolvedValue(null)
      client.getCustomPodmanPath.mockResolvedValue(null)
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options - this should not throw an error
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Component should render without errors
      expect(screen.getByText(/Choose Your Container Runtime/i)).toBeInTheDocument()
    })

    it('should handle undefined custom paths without throwing trim error', async () => {
      // Mock API to return undefined values
      client.getCustomDockerPath.mockResolvedValue(undefined)
      client.getCustomPodmanPath.mockResolvedValue(undefined)
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options - this should not throw an error
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Component should render without errors
      expect(screen.getByText(/Choose Your Container Runtime/i)).toBeInTheDocument()
    })

    it('should handle object custom paths without throwing trim error', async () => {
      // Mock API to return object values (potential cause of trim error)
      client.getCustomDockerPath.mockResolvedValue({ path: '/usr/bin/docker' })
      client.getCustomPodmanPath.mockResolvedValue({ path: '/usr/bin/podman' })
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options - this should not throw an error
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Component should render without errors
      expect(screen.getByText(/Choose Your Container Runtime/i)).toBeInTheDocument()
    })

    it('should handle number custom paths without throwing trim error', async () => {
      // Mock API to return numeric values
      client.getCustomDockerPath.mockResolvedValue(123)
      client.getCustomPodmanPath.mockResolvedValue(456)
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options - this should not throw an error
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Component should render without errors
      expect(screen.getByText(/Choose Your Container Runtime/i)).toBeInTheDocument()
    })
  })

  describe('Application Logs Map Error Reproduction', () => {
    it('should handle backend logs response object without throwing map error', async () => {
      // Mock API to return backend object structure (which causes the map error)
      client.getAppLogs.mockResolvedValue({
        logs: ['Log message 1', 'Log message 2', 'Log message 3'],
        total_logs: 3
      })
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Click the logs show button
      const showLogsButton = screen.getByRole('button', { name: /show/i })
      await user.click(showLogsButton)
      
      // Find the refresh button specifically in the application logs section
      const logsSection = screen.getByText('Application Logs')
      const refreshButtons = screen.getAllByRole('button', { name: /refresh/i })
      // The logs refresh button is smaller (first one found in the logs section)
      const logsRefreshButton = refreshButtons.find(button => 
        button.style.fontSize === '0.75rem' && button.textContent === 'Refresh'
      )
      
      await user.click(logsRefreshButton)
      
      // Component should handle this without throwing map error
      await waitFor(() => {
        expect(screen.getByText(/Log message 1/i)).toBeInTheDocument()
      })
    })

    it('should handle null logs response without throwing map error', async () => {
      // Mock API to return null
      client.getAppLogs.mockResolvedValue(null)
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Click the logs show button
      const showLogsButton = screen.getByRole('button', { name: /show/i })
      await user.click(showLogsButton)
      
      // Find the refresh button specifically in the application logs section
      const refreshButtons = screen.getAllByRole('button', { name: /refresh/i })
      const logsRefreshButton = refreshButtons.find(button => 
        button.style.fontSize === '0.75rem' && button.textContent === 'Refresh'
      )
      
      await user.click(logsRefreshButton)
      
      // Should show "No logs available" message
      expect(screen.getByText(/No logs available/i)).toBeInTheDocument()
    })

    it('should handle string logs response without throwing map error', async () => {
      // Mock API to return a string instead of array
      client.getAppLogs.mockResolvedValue('This is a log string')
      
      render(<RuntimeSetup {...defaultProps} />)
      
      // Open advanced options
      await waitFor(() => {
        const advancedButton = screen.getByRole('button', { name: /advanced options/i })
        return user.click(advancedButton)
      })
      
      // Click the logs show button
      const showLogsButton = screen.getByRole('button', { name: /show/i })
      await user.click(showLogsButton)
      
      // Find the refresh button specifically in the application logs section
      const refreshButtons = screen.getAllByRole('button', { name: /refresh/i })
      const logsRefreshButton = refreshButtons.find(button => 
        button.style.fontSize === '0.75rem' && button.textContent === 'Refresh'
      )
      
      await user.click(logsRefreshButton)
      
      // Should handle gracefully
      expect(screen.getByText(/No logs available/i)).toBeInTheDocument()
    })
  })
})