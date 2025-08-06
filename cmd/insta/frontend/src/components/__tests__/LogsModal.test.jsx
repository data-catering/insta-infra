import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import LogsModal from '../LogsModal'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

// Mock createPortal to render in the current container instead of document.body
vi.mock('react-dom', async () => {
  const actual = await vi.importActual('react-dom')
  return {
    ...actual,
    createPortal: (element) => element
  }
})

describe('LogsModal Component', () => {
  const user = userEvent.setup()

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    serviceName: 'postgres'
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Basic Rendering', () => {
    it('should render modal when isOpen is true', () => {
      renderWithProviders(<LogsModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText('Logs - postgres')).toBeInTheDocument()
    })

    it('should not render modal when isOpen is false', () => {
      renderWithProviders(<LogsModal {...defaultProps} isOpen={false} />, { apiClient: mockApiClient })
      
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('should show service name in title', () => {
      renderWithProviders(<LogsModal {...defaultProps} serviceName="mysql" />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Logs - mysql')).toBeInTheDocument()
    })
  })

  describe('User Interactions', () => {
    it('should call onClose when close button is clicked', async () => {
      const onCloseMock = vi.fn()
      renderWithProviders(<LogsModal {...defaultProps} onClose={onCloseMock} />, { apiClient: mockApiClient })
      
      const closeButton = screen.getByRole('button', { name: '' })
      await user.click(closeButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should have search input', () => {
      renderWithProviders(<LogsModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByPlaceholderText('Search logs...')).toBeInTheDocument()
    })

    it('should have level filter dropdown', () => {
      renderWithProviders(<LogsModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByDisplayValue('All Levels')).toBeInTheDocument()
    })

    it('should have clear button', () => {
      renderWithProviders(<LogsModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
    })
  })

  describe('Controls', () => {
    it('should show auto-scroll checkbox', () => {
      renderWithProviders(<LogsModal {...defaultProps} />, { apiClient: mockApiClient })
      
      const autoScrollCheckbox = screen.getByRole('checkbox')
      expect(autoScrollCheckbox).toBeInTheDocument()
      expect(autoScrollCheckbox).toBeChecked()
    })
  })
})