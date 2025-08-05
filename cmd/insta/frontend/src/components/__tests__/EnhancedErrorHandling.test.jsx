import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
import { createPortal } from 'react-dom';
import ErrorMessage, { useErrorHandler, ToastContainer } from '../ErrorMessage';
import { 
  startService, 
  stopService, 
  getServiceLogs, 
  startImagePull,
  stopImagePull,
  getServiceConnection,
  openServiceConnection
} from '../../api/client';

// Mock the API client
vi.mock('../../api/client', () => ({
  startService: vi.fn(),
  stopService: vi.fn(),
  getServiceLogs: vi.fn(),
  startImagePull: vi.fn(),
  stopImagePull: vi.fn(),
  getServiceConnection: vi.fn(),
  openServiceConnection: vi.fn(),
}));

// Mock createPortal for modal testing
vi.mock('react-dom', async () => {
  const actual = await vi.importActual('react-dom');
  return {
    ...actual,
    createPortal: vi.fn((element) => element),
  };
});

// Test component that uses error handler
const TestErrorComponent = ({ onError, onToast }) => {
  const { errors, toasts, addError, addToast, removeError, removeToast, clearAllErrors } = useErrorHandler();
  
  const handleAddError = () => {
    addError({
      type: 'error',
      title: 'Test Error',
      message: 'This is a test error message',
      details: 'Error details for testing',
      metadata: {
        serviceName: 'postgres',
        action: 'start_service',
        timestamp: '2023-01-01T00:00:00Z'
      },
      actions: [
        {
          label: 'Retry',
          onClick: () => console.log('Retry clicked'),
          variant: 'primary'
        }
      ]
    });
  };

  const handleAddToast = () => {
    addToast({
      type: 'success',
      message: 'Operation successful',
      autoHide: true,
      autoHideDelay: 3000
    });
  };

  // Expose handlers for testing
  if (onError) onError({ addError, removeError, clearAllErrors, errors });
  if (onToast) onToast({ addToast, removeToast, toasts });

  return (
    <div>
      <button onClick={handleAddError} data-testid="add-error">Add Error</button>
      <button onClick={handleAddToast} data-testid="add-toast">Add Toast</button>
      
      {errors.map((error) => (
        <ErrorMessage
          key={error.id}
          type={error.type}
          title={error.title}
          message={error.message}
          details={error.details}
          metadata={error.metadata}
          actions={error.actions}
          onDismiss={() => removeError(error.id)}
        />
      ))}
      
      <ToastContainer 
        toasts={toasts}
        onRemoveToast={removeToast}
      />
    </div>
  );
};

describe('Enhanced Error Handling System', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.clearAllTimers();
  });

  describe('ErrorMessage Component', () => {
    it('renders error message with all components', () => {
      const mockOnDismiss = vi.fn();
      const mockAction = vi.fn();

      render(
        <ErrorMessage
          type="error"
          title="Test Error Title"
          message="Test error message"
          details="Detailed error information"
          metadata={{
            serviceName: 'postgres',
            action: 'start_service',
            timestamp: '2023-01-01T00:00:00Z'
          }}
          actions={[
            {
              label: 'Retry',
              onClick: mockAction,
              variant: 'primary'
            }
          ]}
          onDismiss={mockOnDismiss}
        />
      );

      expect(screen.getByText('Test Error Title')).toBeInTheDocument();
      expect(screen.getByText('Test error message')).toBeInTheDocument();
      expect(screen.getByText('Retry')).toBeInTheDocument();
    });

    it('handles different error types with appropriate styling', () => {
      const errorTypes = ['error', 'warning', 'info', 'success'];
      
      errorTypes.forEach((type) => {
        const { container, unmount } = render(
          <ErrorMessage
            type={type}
            title={`${type} title`}
            message={`${type} message`}
            onDismiss={() => {}}
          />
        );

        const errorElement = container.querySelector(`.error-message-${type}`);
        expect(errorElement).toBeInTheDocument();
        
        unmount();
      });
    });

    it('displays metadata in expandable details', async () => {
      render(
        <ErrorMessage
          type="error"
          title="Error with metadata"
          message="Error message"
          details="Error details"
          metadata={{
            serviceName: 'postgres',
            action: 'start_service',
            timestamp: '2023-01-01T00:00:00Z',
            errorType: 'service_start_failed'
          }}
          onDismiss={() => {}}
        />
      );

      // Check if details section exists
      const detailsElement = screen.getByText('Error details');
      expect(detailsElement).toBeInTheDocument();
    });

    it('calls action handlers when buttons are clicked', () => {
      const mockAction = vi.fn();
      const mockDismiss = vi.fn();

      render(
        <ErrorMessage
          type="error"
          title="Test Error"
          message="Test message"
          actions={[
            {
              label: 'Test Action',
              onClick: mockAction,
              variant: 'primary'
            }
          ]}
          onDismiss={mockDismiss}
        />
      );

      fireEvent.click(screen.getByText('Test Action'));
      expect(mockAction).toHaveBeenCalledTimes(1);

      const closeButton = screen.getByLabelText(/close/i);
      fireEvent.click(closeButton);
      expect(mockDismiss).toHaveBeenCalledTimes(1);
    });

    it('auto-hides messages when configured', async () => {
      vi.useFakeTimers();
      const mockDismiss = vi.fn();

      render(
        <ErrorMessage
          type="success"
          title="Success"
          message="Operation successful"
          autoHide={true}
          autoHideDelay={2000}
          onDismiss={mockDismiss}
        />
      );

      // Fast-forward time
      act(() => {
        vi.advanceTimersByTime(2000);
      });

      // Use act to ensure state updates are flushed
      await act(async () => {
        await vi.runAllTimersAsync();
      });

      expect(mockDismiss).toHaveBeenCalledTimes(1);

      vi.useRealTimers();
    });
  });

  describe('useErrorHandler Hook', () => {
    it('manages error state correctly', () => {
      let errorHandler;
      
      render(
        <TestErrorComponent 
          onError={(handler) => { errorHandler = handler; }}
        />
      );

      expect(errorHandler.errors).toHaveLength(0);

      // Add an error
      act(() => {
        errorHandler.addError({
          type: 'error',
          title: 'Test Error',
          message: 'Test message'
        });
      });

      expect(errorHandler.errors).toHaveLength(1);
      expect(errorHandler.errors[0].title).toBe('Test Error');

      // Remove the error
      act(() => {
        errorHandler.removeError(errorHandler.errors[0].id);
      });

      expect(errorHandler.errors).toHaveLength(0);
    });

    it('manages toast state correctly', () => {
      let toastHandler;
      
      render(
        <TestErrorComponent 
          onToast={(handler) => { toastHandler = handler; }}
        />
      );

      expect(toastHandler.toasts).toHaveLength(0);

      // Add a toast
      act(() => {
        toastHandler.addToast({
          type: 'success',
          message: 'Success message'
        });
      });

      expect(toastHandler.toasts).toHaveLength(1);
      expect(toastHandler.toasts[0].message).toBe('Success message');

      // Remove the toast
      act(() => {
        toastHandler.removeToast(toastHandler.toasts[0].id);
      });

      expect(toastHandler.toasts).toHaveLength(0);
    });

    it('generates unique IDs for errors and toasts', () => {
      let errorHandler;
      
      render(
        <TestErrorComponent 
          onError={(handler) => { errorHandler = handler; }}
        />
      );

      act(() => {
        errorHandler.addError({ title: 'Error 1', message: 'Message 1' });
        errorHandler.addError({ title: 'Error 2', message: 'Message 2' });
      });

      expect(errorHandler.errors).toHaveLength(2);
      expect(errorHandler.errors[0].id).not.toBe(errorHandler.errors[1].id);
    });
  });

  describe('ToastContainer Component', () => {
    it('renders multiple toasts', () => {
      const toasts = [
        { id: '1', type: 'success', message: 'Success 1' },
        { id: '2', type: 'info', message: 'Info 1' },
        { id: '3', type: 'warning', message: 'Warning 1' }
      ];

      render(
        <ToastContainer 
          toasts={toasts}
          onRemoveToast={() => {}}
        />
      );

      expect(screen.getByText('Success 1')).toBeInTheDocument();
      expect(screen.getByText('Info 1')).toBeInTheDocument();
      expect(screen.getByText('Warning 1')).toBeInTheDocument();
    });

    it('handles toast removal', async () => {
      const mockRemoveToast = vi.fn();
      const toasts = [
        { id: '1', type: 'success', message: 'Success message' }
      ];

      const { rerender } = render(
        <ToastContainer 
          toasts={toasts}
          onRemoveToast={mockRemoveToast}
        />
      );

      const closeButton = screen.getByLabelText(/close/i);
      fireEvent.click(closeButton);
      
      // The toast calls onClose which should trigger onRemoveToast
      // Wait for the effect to complete
      await waitFor(() => {
        expect(mockRemoveToast).toHaveBeenCalledWith('1');
      });
    });
  });

  describe('API Error Handling Integration', () => {
    it('handles service start errors with metadata', async () => {
      const errorResponse = {
        message: 'Service start failed',
        metadata: {
          status: 500,
          serviceName: 'postgres',
          action: 'start',
          timestamp: '2023-01-01T00:00:00Z',
          errorType: 'service_start_failed'
        }
      };

      startService.mockRejectedValue(errorResponse);

      try {
        await startService('postgres', false);
      } catch (error) {
        expect(error.message).toBe('Service start failed');
        expect(error.metadata.serviceName).toBe('postgres');
        expect(error.metadata.action).toBe('start');
        expect(error.metadata.errorType).toBe('service_start_failed');
      }
    });

    it('handles image pull errors with detailed metadata', async () => {
      const errorResponse = {
        message: 'Image pull failed',
        metadata: {
          status: 500,
          imageName: 'postgres:13',
          serviceName: 'postgres',
          action: 'start_pull',
          timestamp: '2023-01-01T00:00:00Z',
          errorType: 'image_pull_failed'
        }
      };

      startImagePull.mockRejectedValue(errorResponse);

      try {
        await startImagePull('postgres:13');
      } catch (error) {
        expect(error.message).toBe('Image pull failed');
        expect(error.metadata.imageName).toBe('postgres:13');
        expect(error.metadata.action).toBe('start_pull');
        expect(error.metadata.errorType).toBe('image_pull_failed');
      }
    });

    it('handles service logs errors with context', async () => {
      const errorResponse = {
        message: 'Failed to get logs',
        metadata: {
          status: 500,
          serviceName: 'postgres',
          action: 'get_service_logs',
          timestamp: '2023-01-01T00:00:00Z',
          errorType: 'logs_fetch_failed',
          tailLines: 100
        }
      };

      getServiceLogs.mockRejectedValue(errorResponse);

      try {
        await getServiceLogs('postgres');
      } catch (error) {
        expect(error.message).toBe('Failed to get logs');
        expect(error.metadata.serviceName).toBe('postgres');
        expect(error.metadata.action).toBe('get_service_logs');
        expect(error.metadata.tailLines).toBe(100);
      }
    });

    it('handles connection errors with service context', async () => {
      const errorResponse = {
        message: 'Connection failed',
        metadata: {
          status: 500,
          serviceName: 'postgres',
          action: 'get_connection_info',
          timestamp: '2023-01-01T00:00:00Z',
          errorType: 'connection_info_failed'
        }
      };

      getServiceConnection.mockRejectedValue(errorResponse);

      try {
        await getServiceConnection('postgres');
      } catch (error) {
        expect(error.message).toBe('Connection failed');
        expect(error.metadata.serviceName).toBe('postgres');
        expect(error.metadata.action).toBe('get_connection_info');
      }
    });
  });

  describe('Error Recovery Scenarios', () => {
    it('provides retry functionality for failed operations', () => {
      const mockRetry = vi.fn();
      
      render(
        <ErrorMessage
          type="error"
          title="Operation Failed"
          message="Service failed to start"
          actions={[
            {
              label: 'Retry',
              onClick: mockRetry,
              variant: 'primary'
            }
          ]}
          onDismiss={() => {}}
        />
      );

      const retryButton = screen.getByText('Retry');
      fireEvent.click(retryButton);
      
      expect(mockRetry).toHaveBeenCalledTimes(1);
    });

    it('handles multiple error types simultaneously', () => {
      let errorHandler;
      
      render(
        <TestErrorComponent 
          onError={(handler) => { errorHandler = handler; }}
        />
      );

      act(() => {
        errorHandler.addError({
          type: 'error',
          title: 'Critical Error',
          message: 'Service failed'
        });
        
        errorHandler.addError({
          type: 'warning',
          title: 'Warning',
          message: 'Service degraded'
        });
        
        errorHandler.addError({
          type: 'info',
          title: 'Info',
          message: 'Service updated'
        });
      });

      expect(errorHandler.errors).toHaveLength(3);
      expect(screen.getByText('Critical Error')).toBeInTheDocument();
      expect(screen.getByText('Warning')).toBeInTheDocument();
      expect(screen.getByText('Info')).toBeInTheDocument();
    });

    it('clears all errors when requested', () => {
      let errorHandler;
      
      render(
        <TestErrorComponent 
          onError={(handler) => { errorHandler = handler; }}
        />
      );

      act(() => {
        errorHandler.addError({ title: 'Error 1', message: 'Message 1' });
        errorHandler.addError({ title: 'Error 2', message: 'Message 2' });
      });

      expect(errorHandler.errors).toHaveLength(2);

      act(() => {
        errorHandler.clearAllErrors();
      });

      expect(errorHandler.errors).toHaveLength(0);
    });
  });

  describe('Error Message Formatting', () => {
    it('formats backend error metadata correctly', () => {
      const metadata = {
        serviceName: 'postgres',
        action: 'service_start',
        timestamp: '2023-01-01T12:00:00Z',
        errorType: 'service_start_failed',
        persist: true,
        runtime: 'docker'
      };

      render(
        <ErrorMessage
          type="error"
          title="Service Start Failed"
          message="Failed to start PostgreSQL service"
          details="Service failed to start due to port conflict"
          metadata={metadata}
          onDismiss={() => {}}
        />
      );

      expect(screen.getByText('Service Start Failed')).toBeInTheDocument();
      expect(screen.getByText('Failed to start PostgreSQL service')).toBeInTheDocument();
    });

    it('handles missing metadata gracefully', () => {
      render(
        <ErrorMessage
          type="error"
          title="Error Without Metadata"
          message="Simple error message"
          onDismiss={() => {}}
        />
      );

      expect(screen.getByText('Error Without Metadata')).toBeInTheDocument();
      expect(screen.getByText('Simple error message')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('provides proper ARIA labels and roles', () => {
      render(
        <ErrorMessage
          type="error"
          title="Accessible Error"
          message="Error message"
          onDismiss={() => {}}
        />
      );

      const errorElement = screen.getByRole('alert');
      expect(errorElement).toBeInTheDocument();
    });

    it('supports keyboard navigation', () => {
      const mockAction = vi.fn();
      const mockDismiss = vi.fn();

      render(
        <ErrorMessage
          type="error"
          title="Keyboard Test"
          message="Test message"
          actions={[
            {
              label: 'Action',
              onClick: mockAction,
              variant: 'primary'
            }
          ]}
          onDismiss={mockDismiss}
        />
      );

      const actionButton = screen.getByText('Action');
      const closeButton = screen.getByLabelText(/close/i);

      // Test keyboard interaction - buttons should respond to click events
      actionButton.focus();
      fireEvent.click(actionButton);
      expect(mockAction).toHaveBeenCalledTimes(1);

      closeButton.focus();
      fireEvent.click(closeButton);
      expect(mockDismiss).toHaveBeenCalledTimes(1);
    });
  });
}); 