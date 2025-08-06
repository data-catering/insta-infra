import { screen, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import App from '../App';
import { renderWithProviders, mockApiClient } from '../test-utils/test-utils';

describe('App Integration Tests - Simplified', () => {
  beforeEach(async () => {
    vi.clearAllMocks();

    // Reset all mock functions
    Object.keys(mockApiClient).forEach(key => {
      if (typeof mockApiClient[key] === 'function') {
        mockApiClient[key].mockClear();
      }
    });
  });

  describe('Application Rendering', () => {
    it('should render without crashing', () => {
      renderWithProviders(<App />, { apiClient: mockApiClient });
      expect(document.body).toBeInTheDocument();
    });

    it('should display header with logo', async () => {
      renderWithProviders(<App />, { apiClient: mockApiClient });
      
      await waitFor(() => {
        expect(screen.getByText('Insta')).toBeInTheDocument();
        expect(screen.getByText('Infra')).toBeInTheDocument();
      }, { timeout: 3000 });
    });

    it('should display footer with infrastructure text', async () => {
      renderWithProviders(<App />, { apiClient: mockApiClient });
      
      await waitFor(() => {
        expect(screen.getByText('Infrastructure Control Panel')).toBeInTheDocument();
      }, { timeout: 3000 });
    });
  });
});