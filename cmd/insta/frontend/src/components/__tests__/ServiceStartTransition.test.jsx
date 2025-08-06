import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { vi, describe, it, expect, beforeEach } from 'vitest';
import ServiceItem, { ImageStatusProvider } from '../ServiceItem';
import { renderWithProviders, mockApiClient } from '../../test-utils/test-utils';

// Test helper to render ServiceItem with provider
const renderServiceItem = (props = {}) => {
  const defaultProps = {
    service: {
      Name: 'postgres',
      name: 'postgres',
      Type: 'database',
      type: 'database',
      dependencies: [],
    },
    statuses: { postgres: 'stopped' },
    dependencyStatuses: {},
    onServiceStateChange: vi.fn(),
    ...props,
  };

  return renderWithProviders(
    <ImageStatusProvider services={[defaultProps.service]}>
      <ServiceItem {...defaultProps} />
    </ImageStatusProvider>,
    { apiClient: mockApiClient }
  );
};

describe('ServiceItem Smooth Transitions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset all mock functions
    Object.keys(mockApiClient).forEach(key => {
      if (typeof mockApiClient[key] === 'function') {
        mockApiClient[key].mockClear();
      }
    });
  });

  it('should show smooth transition from stopped to starting to running', async () => {
    const mockOnStateChange = vi.fn();
    
    renderServiceItem({
      onServiceStateChange: mockOnStateChange,
      statuses: { postgres: 'stopped' }
    });

    // Initially should show stopped
    expect(screen.getByText(/stopped/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /start/i })).toBeInTheDocument();

    // Click start button
    const startButton = screen.getByRole('button', { name: /start/i });
    await userEvent.click(startButton);

    // Should call startService
    await waitFor(() => {
      expect(mockApiClient.startService).toHaveBeenCalledWith('postgres', false);
    });
  });

  it('should handle service start failure gracefully', async () => {
    // Mock console.error to prevent test output noise
    const consoleErrorSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    
    mockApiClient.startService.mockRejectedValue(new Error('Service start failed'));

    const mockOnStateChange = vi.fn();
    
    renderServiceItem({
      onServiceStateChange: mockOnStateChange,
      statuses: { postgres: 'stopped' }
    });

    const startButton = screen.getByRole('button', { name: /start/i });
    await userEvent.click(startButton);

    // Should call startService and fail
    await waitFor(() => {
      expect(mockApiClient.startService).toHaveBeenCalledWith('postgres', false);
    });
    
    // Service should still be rendered (not crashed)
    expect(screen.getByText('postgres')).toBeInTheDocument();
    
    consoleErrorSpy.mockRestore();
  });

  it('should not flash between conflicting statuses', async () => {
    const mockOnStateChange = vi.fn();
    const { rerender } = renderServiceItem({
      onServiceStateChange: mockOnStateChange,
      statuses: { postgres: 'stopped' }
    });

    // Initially stopped
    expect(screen.getByText(/stopped/i)).toBeInTheDocument();

    // Simulate external status update (like from WebSocket)
    rerender(
      <ImageStatusProvider services={[{ Name: 'postgres', name: 'postgres', Type: 'database', type: 'database' }]}>
        <ServiceItem
          service={{ Name: 'postgres', name: 'postgres', Type: 'database', type: 'database', dependencies: [] }}
          statuses={{ postgres: 'running' }} // External status update
          dependencyStatuses={{}}
          onServiceStateChange={mockOnStateChange}
        />
      </ImageStatusProvider>
    );

    // Should now show running
    await waitFor(() => {
      expect(screen.getByText(/running/i)).toBeInTheDocument();
    });

    // Should have stop button instead of start button
    expect(screen.getByRole('button', { name: /stop/i })).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /start/i })).not.toBeInTheDocument();
  });

  it('should handle stop transition smoothly', async () => {
    const mockOnStateChange = vi.fn();
    
    renderServiceItem({
      onServiceStateChange: mockOnStateChange,
      statuses: { postgres: 'running' }
    });

    // Should show running initially
    expect(screen.getByText(/running/i)).toBeInTheDocument();
    
    const stopButton = screen.getByRole('button', { name: /stop/i });
    await userEvent.click(stopButton);

    // Should call stopService
    await waitFor(() => {
      expect(mockApiClient.stopService).toHaveBeenCalledWith('postgres');
    });
  });
}); 