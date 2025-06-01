import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Playwright configuration for insta-infra Wails UI E2E testing
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './e2e',
  
  // Run tests in files in parallel
  fullyParallel: false, // Disable for Wails app testing to avoid port conflicts
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // Retry on CI only
  retries: process.env.CI ? 2 : 0,
  
  // Opt out of parallel tests on CI
  workers: process.env.CI ? 1 : undefined,
  
  // Reporter to use
  reporter: [
    ['html'],
    ['json', { outputFile: 'test-results/e2e-results.json' }],
    ['junit', { outputFile: 'test-results/e2e-results.xml' }]
  ],
  
  // Shared settings for all the projects below
  use: {
    // Base URL for your Wails dev server
    baseURL: 'http://localhost:34115',
    
    // Collect trace when retrying the failed test
    trace: 'on-first-retry',
    
    // Capture screenshot on failure
    screenshot: 'only-on-failure',
    
    // Record video on failure
    video: 'retain-on-failure',
    
    // Global test timeout
    actionTimeout: 10000,
    navigationTimeout: 30000,
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'wails-app',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 },
        // Wails-specific settings
        ignoreHTTPSErrors: true,
        permissions: ['clipboard-read', 'clipboard-write'],
      },
    },
    
    // Test responsive design within the Wails app
    {
      name: 'wails-app-small',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 800, height: 600 },
        ignoreHTTPSErrors: true,
      },
    },
  ],

  // Run your local dev server before starting the tests
  webServer: {
    command: 'cd ../../.. && make dev-ui',
    url: 'http://localhost:34115',
    reuseExistingServer: !process.env.CI,
    timeout: 120 * 1000, // 2 minutes to start
    stdout: 'pipe',
    stderr: 'pipe',
  },

  // Global setup and teardown
  globalSetup: join(__dirname, './e2e/global-setup.js'),
  globalTeardown: join(__dirname, './e2e/global-teardown.js'),

  // Test timeout
  timeout: 60000, // 1 minute per test

  // Expect timeout for assertions
  expect: {
    timeout: 10000,
  },

  // Output directory for test artifacts
  outputDir: 'test-results/',
}); 