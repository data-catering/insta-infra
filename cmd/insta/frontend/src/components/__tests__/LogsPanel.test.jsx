import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import LogsPanel from '../LogsPanel'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

describe('LogsPanel Component', () => {
  const user = userEvent.setup()

  const defaultProps = {
    isVisible: true,
    onToggle: vi.fn(),
    serviceName: 'postgres'
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Basic Rendering', () => {
    it('should render panel when visible', () => {
      renderWithProviders(<LogsPanel {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Application Logs')).toBeInTheDocument()
    })

    it('should not render panel when not visible', () => {
      renderWithProviders(<LogsPanel {...defaultProps} isVisible={false} />, { apiClient: mockApiClient })
      
      expect(screen.queryByText('Application Logs')).not.toBeInTheDocument()
    })
  })

  describe('User Interactions', () => {
    it('should call onToggle when close button is clicked', async () => {
      const onToggleMock = vi.fn()
      renderWithProviders(<LogsPanel {...defaultProps} onToggle={onToggleMock} />, { apiClient: mockApiClient })
      
      const closeButton = screen.getByRole('button', { name: '✕' })
      await user.click(closeButton)
      
      expect(onToggleMock).toHaveBeenCalledTimes(1)
    })

    it('should have control buttons', () => {
      renderWithProviders(<LogsPanel {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('button', { name: '⏸️' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '🗑️' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '✕' })).toBeInTheDocument()
    })
  })

  describe('Panel Visibility', () => {
    it('should show loading state initially', () => {
      renderWithProviders(<LogsPanel {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Loading logs...')).toBeInTheDocument()
    })
  })
})