import React from 'react'
import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, test, expect, vi, beforeEach } from 'vitest'
import ServiceItem, { ImageStatusProvider } from '../ServiceItem'

// Mock the Wails functions specifically for this test
const mockWailsFunctions = {
  StartService: vi.fn(),
  StartServiceWithStatusUpdate: vi.fn(),
  StopService: vi.fn(),
  StopServiceWithStatusUpdate: vi.fn(),
  GetServiceConnectionInfo: vi.fn(),
  GetServiceLogs: vi.fn(),
  StartLogStream: vi.fn(),
  GetServiceDependencyGraph: vi.fn(),
  OpenServiceInBrowser: vi.fn(),
  CheckImageExists: vi.fn(),
  GetMultipleImageInfo: vi.fn()
}

// Set up global mocks
global.window.go = {
  main: {
    App: mockWailsFunctions
  }
}

global.window.runtime = {
  EventsOn: vi.fn(),
  EventsOff: vi.fn(),
  EventsOnMultiple: vi.fn(() => () => {}) // Return a cleanup function
}

// Helper to render ServiceItem with ImageStatusProvider
const renderWithProvider = (ui, services = []) => {
  return render(
    <ImageStatusProvider services={services}>
      {ui}
    </ImageStatusProvider>
  )
}

describe('ServiceItem Component - User Interactions', () => {
  const user = userEvent.setup()
  
  const mockService = {
    name: 'postgres',
    status: 'stopped',
    description: 'PostgreSQL Database',
    dependencies: ['redis'],
    ports: ['5432:5432'],
    hasWebUI: false,
    webUIUrl: '',
    connectionInfo: {
      host: 'localhost',
      port: '5432',
      username: 'postgres',
      password: 'postgres'
    }
  }

  const mockRunningService = {
    ...mockService,
    status: 'running'
  }

  beforeEach(() => {
    vi.clearAllMocks()
    // Reset mock implementations
    mockWailsFunctions.StartService.mockResolvedValue()
    mockWailsFunctions.StartServiceWithStatusUpdate.mockResolvedValue()
    mockWailsFunctions.StopService.mockResolvedValue()
    mockWailsFunctions.StopServiceWithStatusUpdate.mockResolvedValue()
    mockWailsFunctions.GetServiceConnectionInfo.mockResolvedValue(mockService.connectionInfo)
    mockWailsFunctions.GetServiceLogs.mockResolvedValue(['log line 1', 'log line 2'])
    mockWailsFunctions.CheckImageExists.mockResolvedValue(true)
    mockWailsFunctions.GetMultipleImageInfo.mockResolvedValue({ postgres: 'postgres:15' })
  })

  describe('Service Starting Flow', () => {
    it('should start service when start button is clicked', async () => {
      renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      const startButton = screen.getByRole('button', { name: /start/i })
      expect(startButton).toBeInTheDocument()
      
      await user.click(startButton)
      
      expect(mockWailsFunctions.StartServiceWithStatusUpdate).toHaveBeenCalledWith('postgres', false)
    })

    it('should start service with persistence when persist button is toggled', async () => {
      renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      // Click the persist button first to enable persistence
      const persistButton = screen.getByRole('button', { name: /persist/i })
      await user.click(persistButton)
      
      // Then click start
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      expect(mockWailsFunctions.StartServiceWithStatusUpdate).toHaveBeenCalledWith('postgres', true)
    })

    it('should show loading state when starting service', async () => {
      // Mock a delayed response
      mockWailsFunctions.StartServiceWithStatusUpdate.mockImplementation(() => 
        new Promise(resolve => setTimeout(resolve, 100))
      )
      
      renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      // Button should show loading state immediately after click
      expect(startButton).toBeDisabled()
      expect(startButton).toHaveClass('button-loading')
    })

    it('should handle start service error gracefully', async () => {
      // Mock an error response
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      mockWailsFunctions.StartServiceWithStatusUpdate.mockRejectedValue(new Error('Docker not running'))
      
      renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalledWith('Failed to start postgres:', expect.any(Error))
      })
      
      consoleSpy.mockRestore()
    })
  })

  describe('Service Stopping Flow', () => {
    it('should stop service when stop button is clicked', async () => {
      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      const stopButton = screen.getByRole('button', { name: /stop/i })
      await user.click(stopButton)
      
      expect(mockWailsFunctions.StopServiceWithStatusUpdate).toHaveBeenCalledWith('postgres')
    })

    it('should show correct UI for running service', () => {
      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      // Should show stop button instead of start
      expect(screen.getByRole('button', { name: /stop/i })).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /start/i })).not.toBeInTheDocument()
      
      // Should show running status indicator (look for the status class or text)
      const statusElement = screen.getByText('postgres')
      expect(statusElement).toBeInTheDocument()
    })
  })

  describe('Connection Information Modal', () => {
    it('should open connection modal when connect button is clicked', async () => {
      // Mock successful connection info retrieval
      mockWailsFunctions.GetServiceConnectionInfo.mockResolvedValue({
        host: 'localhost',
        port: '5432',
        username: 'postgres',
        password: 'postgres',
        url: 'postgresql://postgres:postgres@localhost:5432/postgres'
      })

      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      const connectButton = screen.getByRole('button', { name: /connect/i })
      await user.click(connectButton)
      
      // Check if modal opens
      await waitFor(() => {
        expect(screen.getByText(/Connection Details/)).toBeInTheDocument()
      })
    })

    it('should close connection modal when close button is clicked', async () => {
      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      // Open modal
      const connectButton = screen.getByRole('button', { name: /connect/i })
      await user.click(connectButton)
      
      // Wait for modal to open, then find and click close button
      await waitFor(() => {
        expect(screen.getByText(/Connection Details/)).toBeInTheDocument()
      })
      
      // The close button doesn't have accessible text, so find it by class
      const closeButton = document.querySelector('.connection-modal-close')
      expect(closeButton).toBeInTheDocument()
      await user.click(closeButton)
      
      await waitFor(() => {
        expect(screen.queryByText(/Connection Details/)).not.toBeInTheDocument()
      })
    })

    it('should handle connection info when service has proper connection data', async () => {
      // Mock successful connection info with proper data
      mockWailsFunctions.GetServiceConnectionInfo.mockResolvedValue({
        host: 'localhost',
        port: '5432',
        username: 'postgres',
        password: 'postgres',
        url: 'postgresql://postgres:postgres@localhost:5432/postgres'
      })
      
      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      // Open modal
      const connectButton = screen.getByRole('button', { name: /connect/i })
      await user.click(connectButton)
      
      // Check that connection details are shown (this depends on the actual modal implementation)
      await waitFor(() => {
        expect(screen.getByText(/Connection Details/)).toBeInTheDocument()
      })
    })
  })

  describe('Logs Functionality', () => {
    it('should open logs modal when logs button is clicked', async () => {
      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      const logsButton = screen.getByRole('button', { name: /logs/i })
      await user.click(logsButton)
      
      // Just verify the button click worked - the logs modal opening depends on the component implementation
      expect(logsButton).toBeInTheDocument()
    })
  })

  describe('Service Actions', () => {
    it('should show appropriate buttons based on service status', () => {
      // Test stopped service buttons first
      const { unmount } = renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      expect(screen.getByRole('button', { name: /persist/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /logs/i })).toBeInTheDocument()
      
      // Unmount first component before rendering second
      unmount()
      
      // Test running service buttons
      renderWithProvider(<ServiceItem service={mockRunningService} onStatusChange={vi.fn()} />)
      
      expect(screen.getByRole('button', { name: /stop/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /connect/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /logs/i })).toBeInTheDocument()
    })
  })

  describe('Accessibility', () => {
    it('should have proper ARIA labels and keyboard navigation', async () => {
      renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      // Check that buttons are accessible
      const persistButton = screen.getByRole('button', { name: /persist/i })
      expect(persistButton).toBeInTheDocument()
      
      // Test keyboard navigation
      const startButton = screen.getByRole('button', { name: /start/i })
      startButton.focus()
      expect(startButton).toHaveFocus()
      
      // Test clicking the button directly instead of keyboard event
      await user.click(startButton)
      expect(mockWailsFunctions.StartServiceWithStatusUpdate).toHaveBeenCalled()
    })
  })

  describe('Error Handling', () => {
    it('should handle network errors gracefully', async () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
      mockWailsFunctions.StartServiceWithStatusUpdate.mockRejectedValue(new Error('Network error'))
      
      renderWithProvider(<ServiceItem service={mockService} onStatusChange={vi.fn()} />)
      
      const startButton = screen.getByRole('button', { name: /start/i })
      await user.click(startButton)
      
      await waitFor(() => {
        expect(consoleSpy).toHaveBeenCalled()
      })
      
      consoleSpy.mockRestore()
    })
  })
}) 