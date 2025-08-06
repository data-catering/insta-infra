import { screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ProgressModal from '../ProgressModal'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

// Mock createPortal to render in the current container instead of document.body
vi.mock('react-dom', async () => {
  const actual = await vi.importActual('react-dom')
  return {
    ...actual,
    createPortal: (element) => element
  }
})

describe('ProgressModal Component', () => {
  const user = userEvent.setup()

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    serviceName: 'postgres',
    imageName: 'postgres:15'
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Basic Rendering', () => {
    it('should render modal when isOpen is true', () => {
      renderWithProviders(<ProgressModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText('Downloading Image - postgres')).toBeInTheDocument()
      expect(screen.getByText('postgres:15')).toBeInTheDocument()
    })

    it('should not render modal when isOpen is false', () => {
      renderWithProviders(<ProgressModal {...defaultProps} isOpen={false} />, { apiClient: mockApiClient })
      
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('should show service information', () => {
      renderWithProviders(<ProgressModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Service:')).toBeInTheDocument()
      expect(screen.getByText('Image:')).toBeInTheDocument()
      expect(screen.getByText('Status:')).toBeInTheDocument()
      expect(screen.getByText('postgres')).toBeInTheDocument()
      expect(screen.getByText('postgres:15')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels and roles', () => {
      renderWithProviders(<ProgressModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByRole('progressbar')).toBeInTheDocument()
      expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '0')
      expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemin', '0')
      expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuemax', '100')
    })
  })

  describe('User Interactions', () => {
    it('should call onClose when close button is clicked', async () => {
      const onCloseMock = vi.fn()
      renderWithProviders(<ProgressModal {...defaultProps} onClose={onCloseMock} />, { apiClient: mockApiClient })
      
      // The close button doesn't have a label, so we'll use the X button class
      const closeButton = screen.getByRole('button', { name: '' })
      await user.click(closeButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when cancel button is clicked', async () => {
      const onCloseMock = vi.fn()
      renderWithProviders(<ProgressModal {...defaultProps} onClose={onCloseMock} />, { apiClient: mockApiClient })
      
      const cancelButton = screen.getByRole('button', { name: /cancel download/i })
      await user.click(cancelButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when clicking outside modal', async () => {
      const onCloseMock = vi.fn()
      renderWithProviders(<ProgressModal {...defaultProps} onClose={onCloseMock} />, { apiClient: mockApiClient })
      
      const overlay = screen.getByRole('dialog').parentElement
      await user.click(overlay)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })
  })

  describe('Initial State', () => {
    it('should show initial idle state with 0% progress', () => {
      renderWithProviders(<ProgressModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Idle')).toBeInTheDocument()
      expect(screen.getByText('0.0%')).toBeInTheDocument()
      expect(screen.getByText('0s')).toBeInTheDocument()
    })

    it('should show progress bar with 0% width initially', () => {
      renderWithProviders(<ProgressModal {...defaultProps} />, { apiClient: mockApiClient })
      
      const progressBar = screen.getByRole('progressbar')
      expect(progressBar).toHaveStyle('width: 0%')
    })
  })

  describe('Display Information', () => {
    it('should display service name in title and info section', () => {
      renderWithProviders(<ProgressModal {...defaultProps} serviceName="mysql" />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Downloading Image - mysql')).toBeInTheDocument()
      expect(screen.getByText('mysql')).toBeInTheDocument()
    })

    it('should handle missing image name gracefully', () => {
      renderWithProviders(<ProgressModal {...defaultProps} imageName="" />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Service:')).toBeInTheDocument()
      expect(screen.queryByText('Image:')).not.toBeInTheDocument()
    })
  })
})