import { screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import CustomServiceModal from '../CustomServiceModal'
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils'

// Mock createPortal to render in the current container instead of document.body
vi.mock('react-dom', async () => {
  const actual = await vi.importActual('react-dom')
  return {
    ...actual,
    createPortal: (element) => element
  }
})

describe('CustomServiceModal Component', () => {
  const user = userEvent.setup()

  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    onServiceAdded: vi.fn(),
    onServiceUpdated: vi.fn(),
    editingService: null,
    isEditing: false
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
    it('should render modal when isOpen is true', () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Add Custom Service')).toBeInTheDocument()
    })

    it('should not render modal when isOpen is false', () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} isOpen={false} />, { apiClient: mockApiClient })
      
      expect(screen.queryByText('Add Custom Service')).not.toBeInTheDocument()
    })

    it('should show edit title when editing', () => {
      const editingService = {
        id: 'test-id',
        name: 'Test Service',
        description: 'Test description',
        content: 'services:\n  test:\n    image: nginx'
      }

      renderWithProviders(<CustomServiceModal {...defaultProps} editingService={editingService} isEditing={true} />, { apiClient: mockApiClient })
      
      expect(screen.getByText('Edit Custom Service')).toBeInTheDocument()
    })
  })

  describe('Form Elements', () => {
    it('should have required form fields', () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByLabelText(/service name/i)).toBeInTheDocument()
      expect(screen.getByLabelText(/description/i)).toBeInTheDocument()
      expect(screen.getByPlaceholderText(/paste your docker-compose/i)).toBeInTheDocument()
    })

    it('should have form action buttons', () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('button', { name: /cancel/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /add service/i })).toBeInTheDocument()
    })

    it('should have mode toggle buttons', () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('button', { name: /text editor/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /file upload/i })).toBeInTheDocument()
    })
  })

  describe('User Interactions', () => {
    it('should call onClose when cancel button is clicked', async () => {
      const onCloseMock = vi.fn()
      renderWithProviders(<CustomServiceModal {...defaultProps} onClose={onCloseMock} />, { apiClient: mockApiClient })
      
      const cancelButton = screen.getByRole('button', { name: /cancel/i })
      await user.click(cancelButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when close button is clicked', async () => {
      const onCloseMock = vi.fn()
      renderWithProviders(<CustomServiceModal {...defaultProps} onClose={onCloseMock} />, { apiClient: mockApiClient })
      
      const closeButton = screen.getByRole('button', { name: /close modal/i })
      await user.click(closeButton)
      
      expect(onCloseMock).toHaveBeenCalledTimes(1)
    })

    it('should allow typing in service name field', async () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} />, { apiClient: mockApiClient })
      
      const nameInput = screen.getByLabelText(/service name/i)
      await user.type(nameInput, 'Test Service')
      
      expect(nameInput).toHaveValue('Test Service')
    })

    it('should allow typing in description field', async () => {
      renderWithProviders(<CustomServiceModal {...defaultProps} />, { apiClient: mockApiClient })
      
      const descInput = screen.getByLabelText(/description/i)
      await user.type(descInput, 'Test description')
      
      expect(descInput).toHaveValue('Test description')
    })
  })

  describe('Editing Mode', () => {
    it('should populate form when editing a service', () => {
      const editingService = {
        id: 'test-id',
        name: 'Test Service',
        description: 'Test description',
        content: 'services:\n  test:\n    image: nginx'
      }

      renderWithProviders(<CustomServiceModal {...defaultProps} editingService={editingService} isEditing={true} />, { apiClient: mockApiClient })
      
      expect(screen.getByDisplayValue('Test Service')).toBeInTheDocument()
      expect(screen.getByDisplayValue('Test description')).toBeInTheDocument()
      expect(screen.getByDisplayValue(/services:/)).toBeInTheDocument()
    })

    it('should show update button when editing', () => {
      const editingService = {
        id: 'test-id',
        name: 'Test Service',
        description: 'Test description',
        content: 'services:\n  test:\n    image: nginx'
      }

      renderWithProviders(<CustomServiceModal {...defaultProps} editingService={editingService} isEditing={true} />, { apiClient: mockApiClient })
      
      expect(screen.getByRole('button', { name: /update service/i })).toBeInTheDocument()
    })
  })
})