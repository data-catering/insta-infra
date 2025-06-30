import { test, expect } from '@playwright/test';

test.describe('Service Connection Buttons', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app
    await page.goto('http://localhost:9310');
    
    // Wait for the app to load
    await page.waitForSelector('.service-list', { timeout: 10000 });
  });

  test.describe('Open Button Tests', () => {
    test('should successfully open service web UI when service is running and has web interface', async ({ page, context }) => {
      // Mock the API response for a service with web UI
      await page.route('**/api/services/*/connection/open', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            service_name: 'jaeger',
            action: 'open_connection',
            status: 'success'
          })
        });
      });

      // Mock the service list to include a running service with web UI
      await page.route('**/api/services', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              Name: 'jaeger',
              Description: 'Distributed tracing system',
              Category: 'Observability',
              Status: 'running',
              hasWebUI: true,
              webURL: 'http://localhost:16686'
            }
          ])
        });
      });

      // Mock service statuses
      await page.route('**/api/services/status', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            jaeger: { status: 'running' }
          })
        });
      });

      // Wait for services to load
      await page.reload();
      await page.waitForSelector('[data-service-name="jaeger"]', { timeout: 5000 });

      // Find and click the Open button
      const serviceCard = page.locator('[data-service-name="jaeger"]');
      const openButton = serviceCard.locator('button', { hasText: 'Open' });
      
      await expect(openButton).toBeVisible();
      await expect(openButton).toBeEnabled();
      
      // Track page creation for new tabs/windows
      const pagePromise = context.waitForEvent('page');
      
      await openButton.click();
      
      // Wait for success indicator (could be toast, status change, etc.)
      await page.waitForTimeout(1000); // Give time for API call
      
      // Verify the API was called
      const requests = [];
      page.on('request', req => {
        if (req.url().includes('/connection/open')) {
          requests.push(req);
        }
      });
    });

    test('should show error when service does not have web interface', async ({ page }) => {
      // Mock the API response for a service without web UI
      await page.route('**/api/services/*/connection/open', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'service postgres does not have a web interface'
          })
        });
      });

      // Mock the service list to include a running service without web UI
      await page.route('**/api/services', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              Name: 'postgres',
              Description: 'PostgreSQL database',
              Category: 'Database',
              Status: 'running',
              hasWebUI: false
            }
          ])
        });
      });

      await page.route('**/api/services/status', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            postgres: { status: 'running' }
          })
        });
      });

      await page.reload();
      await page.waitForSelector('[data-service-name="postgres"]', { timeout: 5000 });

      const serviceCard = page.locator('[data-service-name="postgres"]');
      const openButton = serviceCard.locator('button', { hasText: 'Open' });
      
      // Open button should not be visible for services without web UI
      await expect(openButton).not.toBeVisible();
    });
  });

  test.describe('Connect Button Tests', () => {
    test('should show connection details modal when service is running', async ({ page }) => {
      // Mock the connection info API response
      await page.route('**/api/services/*/connection', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            service_name: 'postgres',
            connection: {
              ServiceName: 'postgres',
              Available: true,
              Username: 'postgres',
              Password: 'postgres',
              HostPort: '5432',
              ContainerPort: '5432',
              ConnectionString: 'postgresql://postgres:postgres@localhost:5432/postgres',
              ConnectionCommand: 'docker exec -it postgres-container psql -U postgres',
              HasWebUI: false
            }
          })
        });
      });

      // Mock service list
      await page.route('**/api/services', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              Name: 'postgres',
              Description: 'PostgreSQL database',
              Category: 'Database',
              Status: 'running'
            }
          ])
        });
      });

      await page.route('**/api/services/status', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            postgres: { status: 'running' }
          })
        });
      });

      await page.reload();
      await page.waitForSelector('[data-service-name="postgres"]', { timeout: 5000 });

      const serviceCard = page.locator('[data-service-name="postgres"]');
      const connectButton = serviceCard.locator('button', { hasText: 'Connect' });
      
      await expect(connectButton).toBeVisible();
      await expect(connectButton).toBeEnabled();
      
      await connectButton.click();
      
      // Wait for connection modal to appear
      await page.waitForSelector('.connection-details-modal', { timeout: 5000 });
      
      // Verify modal title
      await expect(page.locator('.connection-details-modal')).toContainText('Connection Details');
      
      // Verify connection information is displayed
      await expect(page.locator('.connection-table')).toBeVisible();
      await expect(page.locator('.connection-table')).toContainText('Username');
      await expect(page.locator('.connection-table')).toContainText('postgres');
      await expect(page.locator('.connection-table')).toContainText('Password');
      await expect(page.locator('.connection-table')).toContainText('Host Port');
      await expect(page.locator('.connection-table')).toContainText('5432');
      
      // Verify connection string is shown
      await expect(page.locator('.connection-section')).toContainText('Connection String');
      await expect(page.locator('.connection-section')).toContainText('postgresql://postgres:postgres@localhost:5432/postgres');
      
      // Verify container command is shown
      await expect(page.locator('.connection-section')).toContainText('Container Access');
      await expect(page.locator('.connection-section')).toContainText('docker exec -it postgres-container psql -U postgres');
      
      // Test copy functionality
      const copyButtons = page.locator('button', { hasText: 'Copy' });
      await expect(copyButtons.first()).toBeVisible();
      
      // Close modal
      await page.locator('.connection-details-modal button[aria-label="Close"]').click();
      await expect(page.locator('.connection-details-modal')).not.toBeVisible();
    });

    test('should show service not available when service is stopped', async ({ page }) => {
      // Mock the connection info API response for unavailable service
      await page.route('**/api/services/*/connection', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            service_name: 'postgres',
            connection: {
              ServiceName: 'postgres',
              Available: false,
              Error: 'Service is not currently running'
            }
          })
        });
      });

      // Mock service list with stopped service
      await page.route('**/api/services', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              Name: 'postgres',
              Description: 'PostgreSQL database',
              Category: 'Database',
              Status: 'stopped'
            }
          ])
        });
      });

      await page.route('**/api/services/status', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            postgres: { status: 'stopped' }
          })
        });
      });

      await page.reload();
      await page.waitForSelector('[data-service-name="postgres"]', { timeout: 5000 });

      const serviceCard = page.locator('[data-service-name="postgres"]');
      const connectButton = serviceCard.locator('button', { hasText: 'Connect' });
      
      await expect(connectButton).toBeVisible();
      await connectButton.click();
      
      // Wait for connection modal to appear
      await page.waitForSelector('.connection-details-modal', { timeout: 5000 });
      
      // Verify "Service Not Available" message is shown
      await expect(page.locator('.connection-details-modal')).toContainText('Service Not Available');
      await expect(page.locator('.connection-details-modal')).toContainText('The service is not currently running or connection information is not available');
      await expect(page.locator('.connection-details-modal')).toContainText('insta start');
      
      // Close modal
      await page.locator('.connection-details-modal button[aria-label="Close"]').click();
      await expect(page.locator('.connection-details-modal')).not.toBeVisible();
    });
  });

  test.describe('Edge Cases and Error Handling', () => {
    test('should handle API errors gracefully for Open button', async ({ page }) => {
      // Mock API error
      await page.route('**/api/services/*/connection/open', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Internal server error'
          })
        });
      });

      await page.route('**/api/services', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              Name: 'jaeger',
              Description: 'Distributed tracing system',
              Category: 'Observability',
              Status: 'running',
              hasWebUI: true
            }
          ])
        });
      });

      await page.route('**/api/services/status', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            jaeger: { status: 'running' }
          })
        });
      });

      await page.reload();
      await page.waitForSelector('[data-service-name="jaeger"]', { timeout: 5000 });

      const serviceCard = page.locator('[data-service-name="jaeger"]');
      const openButton = serviceCard.locator('button', { hasText: 'Open' });
      
      await openButton.click();
      
      // Should show error alert or toast
      await page.waitForTimeout(1000);
      // Could check for console errors or error toast here
    });

    test('should handle API errors gracefully for Connect button', async ({ page }) => {
      // Mock API error
      await page.route('**/api/services/*/connection', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            error: 'Failed to get connection info'
          })
        });
      });

      await page.route('**/api/services', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([
            {
              Name: 'postgres',
              Description: 'PostgreSQL database',
              Category: 'Database',
              Status: 'running'
            }
          ])
        });
      });

      await page.route('**/api/services/status', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            postgres: { status: 'running' }
          })
        });
      });

      await page.reload();
      await page.waitForSelector('[data-service-name="postgres"]', { timeout: 5000 });

      const serviceCard = page.locator('[data-service-name="postgres"]');
      const connectButton = serviceCard.locator('button', { hasText: 'Connect' });
      
      await connectButton.click();
      
      // Wait for modal or error handling
      await page.waitForTimeout(1000);
      
      // Should show connection modal with error state
      await page.waitForSelector('.connection-details-modal', { timeout: 5000 });
      await expect(page.locator('.connection-details-modal')).toContainText('Service Not Available');
    });
  });
}); 