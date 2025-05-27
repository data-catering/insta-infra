# Frontend Testing Guide

## Overview

The insta-infra frontend uses **Vitest** with **React Testing Library** for comprehensive component testing. The testing setup provides a complete mock environment for Wails functions and focuses on user-centric testing approaches.

## Current Test Status

### ✅ Working Tests
- **ServiceItem.test.jsx** - 8 comprehensive tests covering:
  - Component rendering and display
  - User interactions (start/stop buttons, expand/collapse)
  - Service status styling
  - Dependency handling
  - Error states
  - Accessibility features

- **simple.test.jsx** - Basic React component test

### 🔧 In Progress
- Additional component tests (ConnectionModal, ServiceList, etc.)
- Integration tests
- Coverage reporting

## Running Tests

```bash
# Run all tests (non-interactive)
make test-ui

# Run tests in watch mode (for development)
cd cmd/instaui/frontend && npm run test:watch

# Run with coverage
make test-ui-coverage
```

## Test Structure

### Mock System
All Wails functions are automatically mocked via the test setup file (`src/test/setup.js`). This includes:
- Service management functions (StartService, StopService, etc.)
- Connection and status functions
- Runtime management
- Event system mocking

### Testing Patterns

#### Component Testing
```javascript
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

test('user interaction example', async () => {
  const user = userEvent.setup()
  render(<MyComponent />)
  
  const button = screen.getByRole('button', { name: /start/i })
  await user.click(button)
  
  expect(mockFunction).toHaveBeenCalledWith('expected-args')
})
```

#### Async Operations
```javascript
test('async operation', async () => {
  mockFunction.mockResolvedValue({ status: 'success' })
  
  render(<MyComponent />)
  
  await waitFor(() => {
    expect(screen.getByText('Success')).toBeInTheDocument()
  })
})
```

#### Error Handling
```javascript
test('error handling', async () => {
  mockFunction.mockRejectedValue(new Error('Test error'))
  
  render(<MyComponent />)
  
  await waitFor(() => {
    expect(screen.getByText('Error occurred')).toBeInTheDocument()
  })
})
```

## Test Configuration

### Vitest Config (`vite.config.js`)
- **Environment**: jsdom for DOM simulation
- **Setup Files**: Automatic Wails mocking
- **JSX Support**: esbuild with automatic JSX runtime
- **Globals**: Vitest globals enabled

### Dependencies
- `vitest` - Fast test runner
- `@testing-library/react` - React component testing
- `@testing-library/jest-dom` - Enhanced assertions
- `@testing-library/user-event` - Realistic user interactions
- `jsdom` - DOM environment simulation

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
  mockStartService.mockRejectedValue(new Error('Service failed'))
  // ... test error display and recovery
})
```

### 4. Loading States
Test async operations and loading indicators:
```javascript
test('shows loading state during service start', async () => {
  mockStartService.mockImplementation(() => new Promise(() => {})) // Never resolves
  
  render(<ServiceItem service={mockService} />)
  
  await user.click(screen.getByRole('button', { name: /start/i }))
  expect(screen.getByText('Starting...')).toBeInTheDocument()
})
```

## Coverage Goals

- **Components**: 80%+ coverage for all UI components
- **User Flows**: Complete user journey testing
- **Error Scenarios**: Comprehensive error handling tests
- **Accessibility**: WCAG compliance testing

## Adding New Tests

1. Create test file in `src/components/__tests__/`
2. Import required testing utilities
3. Mock any additional Wails functions if needed
4. Follow user-centric testing patterns
5. Include accessibility and error handling tests

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure Wails functions are properly mocked in `src/test/setup.js`
2. **Act Warnings**: Use `waitFor` for async operations
3. **Mock Issues**: Clear mocks between tests with `vi.clearAllMocks()`

### Debug Tips
```javascript
// Debug rendered output
screen.debug()

// Find elements
screen.logTestingPlaygroundURL()
```

## Future Enhancements

- [ ] Integration tests with full app rendering
- [ ] Visual regression testing
- [ ] Performance testing
- [ ] E2E testing with Playwright
- [ ] Automated accessibility auditing 