import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import ErrorMessage, { ErrorToast, useErrorHandler } from '../ErrorMessage';

// Mock lucide-react icons
vi.mock('lucide-react', () => ({
  X: ({ size, ...props }) => <div data-testid="x-icon" {...props} />,
  AlertCircle: ({ size, ...props }) => <div data-testid="alert-circle-icon" {...props} />,
  AlertTriangle: ({ size, ...props }) => <div data-testid="alert-triangle-icon" {...props} />,
  Info: ({ size, ...props }) => <div data-testid="info-icon" {...props} />,
  CheckCircle: ({ size, ...props }) => <div data-testid="check-circle-icon" {...props} />
}));

describe('ErrorMessage Component', () => {
  const mockOnDismiss = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders error message with all props', () => {
    render(
      <ErrorMessage
        type="error"
        title="Test Error"
        message="This is a test error message"
        details="Additional error details"
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByText('Test Error')).toBeInTheDocument();
    expect(screen.getByText('This is a test error message')).toBeInTheDocument();
    expect(screen.getByText('Show Details')).toBeInTheDocument();
    expect(screen.getByTestId('alert-circle-icon')).toBeInTheDocument();
  });

  it('renders different types correctly', () => {
    const { rerender } = render(
      <ErrorMessage type="warning" title="Warning" message="Warning message" />
    );
    expect(screen.getByTestId('alert-triangle-icon')).toBeInTheDocument();

    rerender(<ErrorMessage type="info" title="Info" message="Info message" />);
    expect(screen.getByTestId('info-icon')).toBeInTheDocument();

    rerender(<ErrorMessage type="success" title="Success" message="Success message" />);
    expect(screen.getByTestId('check-circle-icon')).toBeInTheDocument();
  });

  it('calls onDismiss when close button is clicked', () => {
    render(
      <ErrorMessage
        title="Test Error"
        message="Test message"
        onDismiss={mockOnDismiss}
      />
    );

    const closeButton = screen.getByTestId('x-icon').closest('button');
    fireEvent.click(closeButton);

    expect(mockOnDismiss).toHaveBeenCalledTimes(1);
  });

  it('renders action buttons', () => {
    const mockAction = vi.fn();
    const actions = [
      { label: 'Retry', onClick: mockAction, variant: 'primary' },
      { label: 'Cancel', onClick: mockAction, variant: 'secondary' }
    ];

    render(
      <ErrorMessage
        title="Test Error"
        message="Test message"
        actions={actions}
      />
    );

    expect(screen.getByText('Retry')).toBeInTheDocument();
    expect(screen.getByText('Cancel')).toBeInTheDocument();

    fireEvent.click(screen.getByText('Retry'));
    expect(mockAction).toHaveBeenCalledTimes(1);
  });

  it('expands details when clicked', () => {
    render(
      <ErrorMessage
        title="Test Error"
        message="Test message"
        details="Detailed error information"
      />
    );

    const detailsButton = screen.getByText('Show Details');
    fireEvent.click(detailsButton);

    expect(screen.getByText('Detailed error information')).toBeInTheDocument();
  });

  it('auto-hides when autoHide is true', async () => {
    render(
      <ErrorMessage
        title="Test Error"
        message="Test message"
        autoHide={true}
        autoHideDelay={100}
        onDismiss={mockOnDismiss}
      />
    );

    expect(screen.getByText('Test Error')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockOnDismiss).toHaveBeenCalledTimes(1);
    }, { timeout: 200 });
  });

  it('does not auto-hide after viewing details', async () => {
    const onDismiss = vi.fn();
    render(
      <ErrorMessage
        message="Test message"
        details="Test details"
        autoHide={true}
        autoHideDelay={100}
        onDismiss={onDismiss}
      />
    );

    expect(screen.getByText('Test message')).toBeInTheDocument();

    // Click to view details
    const detailsButton = screen.getByText('Show Details');
    fireEvent.click(detailsButton);

    // Wait for the auto-hide delay
    await waitFor(() => {
      expect(screen.getByText('Test message')).toBeInTheDocument();
    }, { timeout: 200 });

    // Should not have called onDismiss
    expect(onDismiss).not.toHaveBeenCalled();
  });

  it('icon and title are on same line, description underneath', () => {
    render(
      <ErrorMessage
        type="error"
        title="Error Title"
        message="Error description"
      />
    );

    const header = screen.getByText('Error Title').closest('.error-message-header');
    const icon = header.querySelector('.error-message-icon');
    const title = header.querySelector('.error-message-title');

    expect(header).toBeInTheDocument();
    expect(icon).toBeInTheDocument();
    expect(title).toBeInTheDocument();

    // Description should be in the text section, not the header
    const description = screen.getByText('Error description');
    expect(description.closest('.error-message-text')).toBeInTheDocument();
    expect(description.closest('.error-message-header')).not.toBeInTheDocument();
  });
});

describe('ErrorToast Component', () => {
  const mockOnClose = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders toast message', () => {
    render(
      <ErrorToast
        type="error"
        message="Toast error message"
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText('Toast error message')).toBeInTheDocument();
    expect(screen.getByTestId('alert-circle-icon')).toBeInTheDocument();
  });

  it('auto-closes after duration', async () => {
    render(
      <ErrorToast
        message="Toast message"
        duration={100}
        onClose={mockOnClose}
      />
    );

    expect(screen.getByText('Toast message')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockOnClose).toHaveBeenCalledTimes(1);
    }, { timeout: 200 });
  });

  it('closes when close button is clicked', () => {
    render(
      <ErrorToast
        message="Toast message"
        onClose={mockOnClose}
      />
    );

    const closeButton = screen.getByTestId('x-icon').closest('button');
    fireEvent.click(closeButton);

    expect(screen.queryByText('Toast message')).not.toBeInTheDocument();
  });
});

describe('useErrorHandler Hook', () => {
  const TestComponent = () => {
    const { errors, addError, removeError, clearAllErrors } = useErrorHandler();

    return (
      <div>
        <div data-testid="error-count">{errors.length}</div>
        <button onClick={() => addError({ title: 'Test Error', message: 'Test message' })}>
          Add Error
        </button>
        <button onClick={() => errors.length > 0 && removeError(errors[0].id)}>
          Remove Error
        </button>
        <button onClick={clearAllErrors}>Clear All</button>
        {errors.map((error) => (
          <div key={error.id} data-testid="error-item">
            {error.title}: {error.message}
          </div>
        ))}
      </div>
    );
  };

  it('manages error state correctly', () => {
    render(<TestComponent />);

    expect(screen.getByTestId('error-count')).toHaveTextContent('0');

    // Add error
    fireEvent.click(screen.getByText('Add Error'));
    expect(screen.getByTestId('error-count')).toHaveTextContent('1');
    expect(screen.getByTestId('error-item')).toHaveTextContent('Test Error: Test message');

    // Remove error
    fireEvent.click(screen.getByText('Remove Error'));
    expect(screen.getByTestId('error-count')).toHaveTextContent('0');
    expect(screen.queryByTestId('error-item')).not.toBeInTheDocument();
  });

  it('clears all errors', () => {
    render(<TestComponent />);

    // Add multiple errors
    fireEvent.click(screen.getByText('Add Error'));
    fireEvent.click(screen.getByText('Add Error'));
    expect(screen.getByTestId('error-count')).toHaveTextContent('2');

    // Clear all
    fireEvent.click(screen.getByText('Clear All'));
    expect(screen.getByTestId('error-count')).toHaveTextContent('0');
  });
}); 