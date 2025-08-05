import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import LogsPanel from '../LogsPanel'

// Mock the API client
vi.mock('../../api/client', () => ({
  getAppLogEntries: vi.fn(),
  getAppLogsSince: vi.fn(),
}))

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
      render(<LogsPanel {...defaultProps} />)
      
      expect(screen.getByText('Application Logs')).toBeInTheDocument()
    })

    it('should not render panel when not visible', () => {
      render(<LogsPanel {...defaultProps} isVisible={false} />)
      
      expect(screen.queryByText('Application Logs')).not.toBeInTheDocument()
    })
  })

  describe('User Interactions', () => {
    it('should call onToggle when close button is clicked', async () => {
      const onToggleMock = vi.fn()
      render(<LogsPanel {...defaultProps} onToggle={onToggleMock} />)
      
      const closeButton = screen.getByRole('button', { name: '✕' })
      await user.click(closeButton)
      
      expect(onToggleMock).toHaveBeenCalledTimes(1)
    })

    it('should have control buttons', () => {
      render(<LogsPanel {...defaultProps} />)
      
      expect(screen.getByRole('button', { name: '⏸️' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '🗑️' })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: '✕' })).toBeInTheDocument()
    })
  })

  describe('Panel Visibility', () => {
    it('should show loading state initially', () => {
      render(<LogsPanel {...defaultProps} />)
      
      expect(screen.getByText('Loading logs...')).toBeInTheDocument()
    })
  })
})