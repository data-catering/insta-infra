import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import LogsModal from '../LogsModal'

// Mock the API client
vi.mock('../../api/client', () => ({
  getServiceLogs: vi.fn(),
  wsClient: {
    subscribe: vi.fn(),
    unsubscribe: vi.fn(),
  },
  WS_MSG_TYPES: {
    SERVICE_LOGS: 'service_logs',
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
      render(<LogsModal {...defaultProps} />)
      
      expect(screen.getByRole('dialog')).toBeInTheDocument()
      expect(screen.getByText('Logs - postgres')).toBeInTheDocument()
    })

    it('should not render modal when isOpen is false', () => {
      render(<LogsModal {...defaultProps} isOpen={false} />)
      
      expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    })

    it('should show service name in title', () => {
      render(<LogsModal {...defaultProps} serviceName="mysql" />)
      
      expect(screen.getByText('Logs - mysql')).toBeInTheDocument()
    })
  })

  describe('User Interactions', () => {
    it('should call onClose when close button is clicked', async () => {
      const onCloseMock = vi.fn()
      render(<LogsModal {...defaultProps} onClose={onCloseMock} />)
      
      const closeButton = screen.getByRole('button', { name: '' })
      await user.click(closeButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should have search input', () => {
      render(<LogsModal {...defaultProps} />)
      
      expect(screen.getByPlaceholderText('Search logs...')).toBeInTheDocument()
    })

    it('should have level filter dropdown', () => {
      render(<LogsModal {...defaultProps} />)
      
      expect(screen.getByDisplayValue('All Levels')).toBeInTheDocument()
    })

    it('should have clear button', () => {
      render(<LogsModal {...defaultProps} />)
      
      expect(screen.getByRole('button', { name: /clear/i })).toBeInTheDocument()
    })
  })

  describe('Controls', () => {
    it('should show auto-scroll checkbox', () => {
      render(<LogsModal {...defaultProps} />)
      
      const autoScrollCheckbox = screen.getByRole('checkbox')
      expect(autoScrollCheckbox).toBeInTheDocument()
      expect(autoScrollCheckbox).toBeChecked()
    })
  })
})