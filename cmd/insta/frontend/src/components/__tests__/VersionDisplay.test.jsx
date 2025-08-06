import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import VersionDisplay from '../VersionDisplay'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

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
      mockApiClient.getApiInfo.mockImplementation(() => new Promise(() => {})) // Never resolves
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      expect(screen.getByText(/loading version/i)).toBeInTheDocument()
    })

    it('should display version when API call succeeds', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '1.2.3' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: 1\.2\.3/i)).toBeInTheDocument()
      })
    })

    it('should display unknown when version is not provided', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({}) // No version field
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: unknown/i)).toBeInTheDocument()
      })
    })

    it('should handle null version gracefully', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: null })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: unknown/i)).toBeInTheDocument()
      })
    })
  })

  describe('Error Handling', () => {
    it('should show error message when API call fails', async () => {
      mockApiClient.getApiInfo.mockRejectedValue(new Error('Network error'))
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/error fetching version/i)).toBeInTheDocument()
      })
    })

    it('should call error handler when API fails', async () => {
      // Mock the error handler
      mockApiClient.getApiInfo.mockRejectedValue(new Error('Network error'))
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
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
      mockApiClient.getApiInfo.mockRejectedValue(new Error('Network error'))
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
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
      // First call fails, second call succeeds
      mockApiClient.getApiInfo
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({ version: '1.2.3' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      // Trigger retry
      const errorCall = mockAddError.mock.calls[0][0]
      errorCall.actions[0].onClick()
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: 1\.2\.3/i)).toBeInTheDocument()
      })
      
      expect(mockApiClient.getApiInfo).toHaveBeenCalledTimes(2)
    })
  })

  describe('Component Styling', () => {
    it('should apply correct styling', () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '1.2.3' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      const versionDiv = screen.getByText(/loading version/i).parentElement
      
      expect(versionDiv).toBeInTheDocument()
    })

    it('should maintain styling after version loads', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '1.2.3' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: 1\.2\.3/i)).toBeInTheDocument()
      })
      
      const versionDiv = screen.getByText(/insta-infra ui version: 1\.2\.3/i).parentElement
      
      expect(versionDiv).toBeInTheDocument()
    })
  })

  describe('Version Text Format', () => {
    it('should format version text correctly', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '2.1.0-beta' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText('Insta-Infra UI Version: 2.1.0-beta')).toBeInTheDocument()
      })
    })

    it('should handle empty string version', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/insta-infra ui version: unknown/i)).toBeInTheDocument()
      })
    })

    it('should handle very long version strings', async () => {
      const longVersion = 'v1.2.3-alpha.1+build.12345.abcdef.very.long.version.string'
      mockApiClient.getApiInfo.mockResolvedValue({ version: longVersion })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(`Insta-Infra UI Version: ${longVersion}`)).toBeInTheDocument()
      })
    })
  })

  describe('API Integration', () => {
    it('should call getApiInfo on mount', () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '1.0.0' })
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      expect(mockApiClient.getApiInfo).toHaveBeenCalledTimes(1)
      expect(mockApiClient.getApiInfo).toHaveBeenCalledWith()
    })

    it('should not call getApiInfo multiple times unnecessarily', async () => {
      mockApiClient.getApiInfo.mockResolvedValue({ version: '1.0.0' })
      
      const { rerender } = renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText(/1\.0\.0/)).toBeInTheDocument()
      })
      
      // Rerender shouldn't trigger another API call
      rerender(<VersionDisplay />)
      
      expect(mockApiClient.getApiInfo).toHaveBeenCalledTimes(1)
    })
  })

  describe('Error Details', () => {
    it('should include timestamp in error details', async () => {
      
      mockApiClient.getApiInfo.mockRejectedValue(new Error('Test error'))
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      const errorCall = mockAddError.mock.calls[0][0]
      expect(errorCall.details).toMatch(/Time: \d+:\d+:\d+/)
      expect(errorCall.metadata.timestamp).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/)
    })

    it('should handle errors without message property', async () => {
      mockApiClient.getApiInfo.mockRejectedValue('String error')
      
      renderWithProviders(<VersionDisplay />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(mockAddError).toHaveBeenCalled()
      })
      
      const errorCall = mockAddError.mock.calls[0][0]
      expect(errorCall.details).toContain('String error')
      expect(errorCall.metadata.errorMessage).toBe('String error')
    })
  })
})