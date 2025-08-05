# Frontend Testing Guide

## Overview

The insta-infra frontend uses **Vitest** with **React Testing Library** for comprehensive component testing. The testing setup provides a complete mock environment for HTTP API calls and focuses on user-centric testing approaches.

## Current Test Status

### ✅ Working Tests
- **ServiceItem.test.jsx** - 13 comprehensive tests covering:
  - Component rendering and display
  - User interactions (start/stop buttons, connection modal)
  - Service status styling
  - Dependency handling
  - Error states
  - Accessibility features

- **App.integration.test.jsx** - Integration tests covering:
  - Service loading and display
  - Enhanced service data handling
  - API integration
  - Error handling

### 🔧 In Progress
- Additional component tests (ConnectionModal, ServiceList, etc.)
- E2E tests with Playwright
- Coverage reporting

## Running Tests

```bash
# Run all tests (non-interactive)
make test-ui

# Run tests in watch mode (for development)
cd cmd/insta/frontend && npm run test:watch

# Run with coverage
make test-ui-coverage

# Run E2E tests
cd cmd/insta/frontend && npm run test:e2e
```

## Test Structure

### Mock System
All HTTP API calls are automatically mocked via the test setup file (`src/test/setup.js`). This includes:
- Service management endpoints (start/stop services)
- Connection and status endpoints
- Runtime management APIs
- WebSocket connections

### Testing Patterns

#### Component Testing
```javascript
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

test('user interaction example', async () => {
  const user = userEvent.setup()
  
  // Mock API response
  global.fetch.mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve({ success: true })
  })
  
  render(<MyComponent />)
  
  const button = screen.getByRole('button', { name: /start/i })
  await user.click(button)
  
  expect(global.fetch).toHaveBeenCalledWith(
    expect.stringContaining('/start'),
    expect.objectContaining({ method: 'POST' })
  )
})
```

#### Async Operations
```javascript
test('async operation', async () => {
  global.fetch.mockResolvedValue({
    ok: true,
    status: 200,
    json: () => Promise.resolve({ status: 'success' })
  })
  
  render(<MyComponent />)
  
  await waitFor(() => {
    expect(screen.getByText('Success')).toBeInTheDocument()
  })
})
```

#### Error Handling
```javascript
test('error handling', async () => {
  global.fetch.mockResolvedValue({
    ok: false,
    status: 500,
    json: () => Promise.resolve({ error: 'Server error' })
  })
  
  render(<MyComponent />)
  
  await waitFor(() => {
    expect(screen.getByText('Error occurred')).toBeInTheDocument()
  })
})
```

## Test Configuration

### Vitest Config (`vite.config.js`)
- **Environment**: jsdom for DOM simulation
- **Setup Files**: Automatic HTTP API mocking
- **JSX Support**: esbuild with automatic JSX runtime
- **Globals**: Vitest globals enabled

### Dependencies
- `vitest` - Fast test runner
- `@testing-library/react` - React component testing
- `@testing-library/jest-dom` - Enhanced assertions
- `@testing-library/user-event` - Realistic user interactions
- `jsdom` - DOM environment simulation
- `@playwright/test` - E2E testing

## Best Practices

### 1. User-Centric Testing
Focus on testing what users see and do, not implementation details:
```javascript
// ✅ Good - tests user behavior
expect(screen.getByRole('button', { name: /start service/i })).toBeInTheDocument()

// ❌ Avoid - tests implementation
expect(component.state.isStarted).toBe(true)
```

### 2. Accessibility Testing
Include accessibility checks in tests:
```javascript
// Test for proper ARIA labels
expect(screen.getByLabelText('Service status')).toBeInTheDocument()

// Test keyboard navigation
await user.keyboard('{Tab}')
expect(screen.getByRole('button')).toHaveFocus()
```

### 3. Error Boundaries
Test error handling and recovery:
```javascript
test('handles service errors gracefully', async () => {
  global.fetch.mockRejectedValue(new Error('Network error'))
  // ... test error display and recovery
})
```

### 4. Loading States
Test async operations and loading indicators:
```javascript
test('shows loading state during service start', async () => {
  global.fetch.mockImplementation(() => 
    new Promise(() => {}) // Never resolves
  )
  
  render(<ServiceItem service={mockService} />)
  
  await user.click(screen.getByRole('button', { name: /start/i }))
  expect(screen.getByText('Starting...')).toBeInTheDocument()
})
```

## Test Types

### Unit Tests
- Component rendering and behavior
- User interactions
- Props handling
- Error states

### Integration Tests
- HTTP API integration
- Service data flow
- Enhanced service features
- WebSocket communication

### E2E Tests
- Full user workflows
- Real backend integration
- Cross-browser testing
- Performance validation

## Coverage Goals

- **Components**: 80%+ coverage for all UI components
- **User Flows**: Complete user journey testing
- **Error Scenarios**: Comprehensive error handling tests
- **Accessibility**: WCAG compliance testing

## Adding New Tests

1. Create test file in `src/components/__tests__/` or `src/__tests__/`
2. Import required testing utilities
3. Mock HTTP API responses as needed
4. Follow user-centric testing patterns
5. Include accessibility and error handling tests

## Troubleshooting

### Common Issues

1. **API Mock Issues**: Ensure fetch is properly mocked in test setup
2. **Act Warnings**: Use `waitFor` for async operations
3. **Mock Cleanup**: Clear mocks between tests with `vi.clearAllMocks()`
4. **WebSocket Mocks**: Mock WebSocket connections to prevent connection attempts

### Debug Tips
```javascript
// Debug rendered output
screen.debug()

// Find elements
screen.logTestingPlaygroundURL()

// Debug API calls
console.log(global.fetch.mock.calls)
```

## E2E Testing

E2E tests use Playwright to test the full application with a real backend:

```bash
# Run E2E tests
npm run test:e2e

# Run E2E tests with UI
npm run test:e2e:ui

# Run E2E tests in headed mode
npm run test:e2e:headed
```

E2E tests cover:
- Service management workflows
- Real HTTP API integration
- Browser interactions
- Performance validation

## Future Enhancements

- [ ] Visual regression testing
- [ ] Performance testing
- [ ] Enhanced E2E coverage
- [ ] Automated accessibility auditing
- [ ] API contract testing 