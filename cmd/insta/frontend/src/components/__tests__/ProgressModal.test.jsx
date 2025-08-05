import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ProgressModal from '../ProgressModal'

// Mock the API client
vi.mock('../../api/client', () => ({
  wsClient: {
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
  },
  WS_MSG_TYPES: {
    IMAGE_PULL_PROGRESS: 'image_pull_progress',
    IMAGE_PULL_COMPLETE: 'image_pull_complete',
    IMAGE_PULL_ERROR: 'image_pull_error',
  }
}))

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
      render(<ProgressModal {...defaultProps} />)
      
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText('Downloading Image - postgres')).toBeInTheDocument()
      expect(screen.getByText('postgres:15')).toBeInTheDocument()
    })

    it('should not render modal when isOpen is false', () => {
      render(<ProgressModal {...defaultProps} isOpen={false} />)
      
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('should show service information', () => {
      render(<ProgressModal {...defaultProps} />)
      
      expect(screen.getByText('Service:')).toBeInTheDocument()
      expect(screen.getByText('Image:')).toBeInTheDocument()
      expect(screen.getByText('Status:')).toBeInTheDocument()
      expect(screen.getByText('postgres')).toBeInTheDocument()
      expect(screen.getByText('postgres:15')).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels and roles', () => {
      render(<ProgressModal {...defaultProps} />)
      
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
      render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
      
      // The close button doesn't have a label, so we'll use the X button class
      const closeButton = screen.getByRole('button', { name: '' })
      await user.click(closeButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when cancel button is clicked', async () => {
      const onCloseMock = vi.fn()
      render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
      
      const cancelButton = screen.getByRole('button', { name: /cancel download/i })
      await user.click(cancelButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when clicking outside modal', async () => {
      const onCloseMock = vi.fn()
      render(<ProgressModal {...defaultProps} onClose={onCloseMock} />)
      
      const overlay = screen.getByRole('dialog').parentElement
      await user.click(overlay)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })
  })

  describe('Initial State', () => {
    it('should show initial idle state with 0% progress', () => {
      render(<ProgressModal {...defaultProps} />)
      
      expect(screen.getByText('Idle')).toBeInTheDocument()
      expect(screen.getByText('0.0%')).toBeInTheDocument()
      expect(screen.getByText('0s')).toBeInTheDocument()
    })

    it('should show progress bar with 0% width initially', () => {
      render(<ProgressModal {...defaultProps} />)
      
      const progressBar = screen.getByRole('progressbar')
      expect(progressBar).toHaveStyle('width: 0%')
    })
  })

  describe('Display Information', () => {
    it('should display service name in title and info section', () => {
      render(<ProgressModal {...defaultProps} serviceName="mysql" />)
      
      expect(screen.getByText('Downloading Image - mysql')).toBeInTheDocument()
      expect(screen.getByText('mysql')).toBeInTheDocument()
    })

    it('should handle missing image name gracefully', () => {
      render(<ProgressModal {...defaultProps} imageName="" />)
      
      expect(screen.getByText('Service:')).toBeInTheDocument()
      expect(screen.queryByText('Image:')).not.toBeInTheDocument()
    })
  })
})