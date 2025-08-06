import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CustomServicesList from '../CustomServicesList'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

// Mock CustomServiceModal
vi.mock('../CustomServiceModal', () => ({
  default: ({ isOpen, onClose }) => isOpen ? <div data-testid="custom-service-modal">Modal Open</div> : null
}))

describe('CustomServicesList Component', () => {
  const user = userEvent.setup()

  const defaultProps = {
    onServiceAdded: vi.fn(),
    onServiceUpdated: vi.fn(),
    onServiceDeleted: vi.fn(),
    isVisible: true
  }

  beforeEach(() => {
    vi.clearAllMocks()
    
    // Reset all mock functions
    Object.keys(mockApiClient).forEach(key => {
      if (typeof mockApiClient[key] === 'function') {
        mockApiClient[key].mockClear();
      }
    });

    vi.spyOn(console, 'log').mockImplementation(() => {})
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  describe('Basic Rendering', () => {
    it('should render the component when visible', async () => {
      renderWithProviders(<CustomServicesList {...defaultProps} />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText('Custom Services')).toBeInTheDocument()
      })
    })

    it('should not render when not visible', () => {
      renderWithProviders(<CustomServicesList {...defaultProps} isVisible={false} />, { apiClient: mockApiClient })
      
      expect(screen.queryByText('Custom Services')).not.toBeInTheDocument()
    })

    it('should show empty state when no services exist', async () => {
      renderWithProviders(<CustomServicesList {...defaultProps} />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText('No custom services configured')).toBeInTheDocument()
      })
    })
  })

  describe('User Interactions', () => {
    it('should show modal when add button is clicked', async () => {
      renderWithProviders(<CustomServicesList {...defaultProps} />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        const addButton = screen.getByRole('button', { name: /add custom/i })
        return user.click(addButton)
      })
      
      expect(screen.getByTestId('custom-service-modal')).toBeInTheDocument()
    })

    it('should have add custom button', async () => {
      renderWithProviders(<CustomServicesList {...defaultProps} />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /add custom/i })).toBeInTheDocument()
      })
    })
  })

  describe('Services Display', () => {
    it('should display services when they exist', async () => {
      const mockServices = [
        {
          id: 'service-1',
          name: 'Test Service',
          description: 'Test description',
          filename: 'test.yaml',
          services: ['web'],
          created_at: '2024-01-01T10:00:00Z',
          updated_at: '2024-01-01T10:00:00Z',
          metadata: { warnings: [] }
        }
      ]
      
      mockApiClient.listCustomServices.mockResolvedValue(mockServices)
      
      renderWithProviders(<CustomServicesList {...defaultProps} />, { apiClient: mockApiClient })
      
      await waitFor(() => {
        expect(screen.getByText('Test Service')).toBeInTheDocument()
      })
    })
  })
})