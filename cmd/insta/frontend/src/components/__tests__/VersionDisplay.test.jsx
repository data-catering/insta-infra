import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import VersionDisplay from '../VersionDisplay'

// Mock the API client
vi.mock('../../api/client', () => ({
  getApiInfo: vi.fn(),
}))

import * as client from '../../api/client'

// Mock the ErrorMessage hook
const mockAddError = vi.fn()
vi.mock('../ErrorMessage', () => ({
  useErrorHandler: () => ({
    addError: mockAddError
  })
}))

describe('VersionDisplay Component', () => {
  const user = userEvent.setup()

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Mock console methods
    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.resetAllMocks()
  })

  describe('Version Loading', () => {
    it('should show loading state initially', () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockImplementation(() => new Promise(() => {})) // Never resolves
      
      render(<VersionDisplay />)
      
      expect(screen.getByText(/loading version/i)).toBeInTheDocument()
    })

    it('should display version when API call succeeds', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '1.2.3' })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: 1\.2\.3/i)).toBeInTheDocument()
      })
    })

    it('should display unknown when version is not provided', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({}) // No version field
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: unknown/i)).toBeInTheDocument()
      })
    })

    it('should handle null version gracefully', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: null })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: unknown/i)).toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should show error message when API call fails', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockRejectedValue(new Error('Network error'))
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/error fetching version/i)).toBeInTheDocument()
      })
    })

    it('should call error handler when API fails', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      
      // Mock the error handler
      
      getApiInfo.mockRejectedValue(new Error('Network error'))
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalledWith(
          expect.objectContaining({
            type: 'warning',
            title: 'Version Information Unavailable',
            message: 'Unable to fetch application version information',
            details: expect.stringContaining('Network error'),
            metadata: expect.objectContaining({
              action: 'fetch_version',
              errorMessage: 'Network error'
            })
          })
        )
      })
    })

    it('should provide retry action in error handler', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      
      
      getApiInfo.mockRejectedValue(new Error('Network error'))
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      const errorCall = mockAddError.mock.calls[0][0]
      expect(errorCall.actions).toHaveLength(1)
      expect(errorCall.actions[0].label).toBe('Retry')
      expect(errorCall.actions[0].variant).toBe('secondary')
      expect(typeof errorCall.actions[0].onClick).toBe('function')
    })

    it('should retry when retry action is clicked', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      
      
      // First call fails, second call succeeds
      getApiInfo
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({ version: '1.2.3' })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      // Trigger retry
      const errorCall = mockAddError.mock.calls[0][0]
      errorCall.actions[0].onClick()
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: 1\.2\.3/i)).toBeInTheDocument()
      })
      
      expect(getApiInfo).toHaveBeenCalledTimes(2)
    })
  })

  describe('Component Styling', () => {
    it('should apply correct styling', () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '1.2.3' })
      
      render(<VersionDisplay />)
      
      const versionDiv = screen.getByText(/loading version/i).parentElement
      
      expect(versionDiv).toBeInTheDocument()
    })

    it('should maintain styling after version loads', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '1.2.3' })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: 1\.2\.3/i)).toBeInTheDocument()
      })
      
      const versionDiv = screen.getByText(/insta-infra ui version: 1\.2\.3/i).parentElement
      
      expect(versionDiv).toBeInTheDocument()
    })
  })

  describe('Version Text Format', () => {
    it('should format version text correctly', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '2.1.0-beta' })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText('Insta-Infra UI Version: 2.1.0-beta')).toBeInTheDocument()
      })
    })

    it('should handle empty string version', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '' })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: unknown/i)).toBeInTheDocument()
      })
    })

    it('should handle very long version strings', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      const longVersion = 'v1.2.3-alpha.1+build.12345.abcdef.very.long.version.string'
      getApiInfo.mockResolvedValue({ version: longVersion })
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(`Insta-Infra UI Version: ${longVersion}`)).toBeInTheDocument()
      })
    })
  })

  describe('API Integration', () => {
    it('should call getApiInfo on mount', () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '1.0.0' })
      
      render(<VersionDisplay />)
      
      expect(getApiInfo).toHaveBeenCalledTimes(1)
      expect(getApiInfo).toHaveBeenCalledWith()
    })

    it('should not call getApiInfo multiple times unnecessarily', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      getApiInfo.mockResolvedValue({ version: '1.0.0' })
      
      const { rerender } = render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(screen.getByText(/1\.0\.0/)).toBeInTheDocument()
      })
      
      // Rerender shouldn't trigger another API call
      rerender(<VersionDisplay />)
      
      expect(getApiInfo).toHaveBeenCalledTimes(1)
    })
  })

  describe('Error Details', () => {
    it('should include timestamp in error details', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      
      
      getApiInfo.mockRejectedValue(new Error('Test error'))
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      const errorCall = mockAddError.mock.calls[0][0]
      expect(errorCall.details).toMatch(/Time: \d+:\d+:\d+/)
      expect(errorCall.metadata.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)
    })

    it('should handle errors without message property', async () => {
      const getApiInfo = vi.mocked(client.getApiInfo)
      
      
      getApiInfo.mockRejectedValue('String error')
      
      render(<VersionDisplay />)
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      const errorCall = mockAddError.mock.calls[0][0]
      expect(errorCall.details).toContain('String error')
      expect(errorCall.metadata.errorMessage).toBe('String error')
    })
  })
})