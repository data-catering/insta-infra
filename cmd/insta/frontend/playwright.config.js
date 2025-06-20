import { defineConfig, devices } from '@playwright/test';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

/**
 * Playwright configuration for insta-infra Web UI E2E testing
 * @see https://playwright.dev/docs/test-configuration
 */
export default defineConfig({
  testDir: './e2e/tests',
  
  // Run tests in files in parallel
  fullyParallel: false, // Disable to avoid container/service conflicts
  
  // Fail the build on CI if you accidentally left test.only in the source code
  forbidOnly: !!process.env.CI,
  
  // Retry on CI only
  retries: process.env.CI ? 2 : 0,
  
  // Opt out of parallel tests on CI
  workers: 1, // Force sequential execution to avoid container conflicts
  
  // Reporter to use
  reporter: [
    ['html'],
    ['json', { outputFile: 'test-results/e2e-results.json' }],
    ['junit', { outputFile: 'test-results/e2e-results.xml' }],
    ['list'] // Add list reporter for development
  ],
  
  // Shared settings for all the projects below
  use: {
    // Base URL for the web UI server
    baseURL: 'http://localhost:9310',
    
    // Collect trace when retrying the failed test
    trace: 'on-first-retry',
    
    // Capture screenshot on failure
    screenshot: 'only-on-failure',
    
    // Record video on failure
    video: 'retain-on-failure',
    
    // Global test timeout - increased for container operations
    actionTimeout: 15000,
    navigationTimeout: 30000,
  },

  // Configure projects for major browsers
  projects: [
    {
      name: 'chromium-desktop',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1280, height: 720 },
        // Web UI specific settings
        ignoreHTTPSErrors: true,
        permissions: ['clipboard-read', 'clipboard-write'],
      },
    },
    
    // Test responsive design
    {
      name: 'chromium-mobile',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 800, height: 600 },
        ignoreHTTPSErrors: true,
      },
    },
  ],

  // Run your local dev server before starting the tests
  webServer: {
    command: 'cd ../../.. && make dev-web',
    url: 'http://localhost:9310',
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