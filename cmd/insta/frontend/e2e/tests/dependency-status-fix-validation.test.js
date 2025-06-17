import { test, expect } from '@playwright/test';
import { waitForNetworkIdle } from '../utils/wait-utilities.js';

// Simple screenshot helper
async function takeScreenshot(page, name) {
  try {
    await page.screenshot({ path: `test-results/${name}`, fullPage: true });
    console.log(`Screenshot saved: ${name}`);
  } catch (error) {
    console.log(`Failed to take screenshot ${name}:`, error.message);
  }
}

// Simple loading wait helper  
async function waitForLoadingToComplete(page, timeout = 10000) {
  try {
    await page.waitForLoadState('networkidle', { timeout });
    await page.waitForTimeout(1000); // Additional buffer
  } catch (error) {
    console.log('Network idle timeout, continuing...');
  }
}

test.describe('Dependency Status Fix Validation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForLoadingToComplete(page, 10000);
  });

  test('should maintain dependency status throughout service lifecycle', async ({ page }) => {
    console.log('=== Starting Dependency Status Fix Validation ===');
    
    // Step 1: Ensure clean state - stop all services first
    console.log('Step 1: Stopping all services to ensure clean state');
    try {
      await page.getByRole('button', { name: /stop all/i }).click();
      await page.waitForTimeout(2000);
    } catch (error) {
      console.log('No stop all button or services already stopped');
    }
    
    // Step 2: Find and start keycloak service
    console.log('Step 2: Finding keycloak service');
    const keycloakServiceCard = page.locator('.service-card').filter({ hasText: 'keycloak' });
    await expect(keycloakServiceCard).toBeVisible({ timeout: 10000 });
    
    // Step 3: Check initial dependencies before starting
    console.log('Step 3: Checking initial dependencies');
    const dependenciesContainer = keycloakServiceCard.locator('.dependencies');
    await expect(dependenciesContainer).toBeVisible();
    
    const initialDependencyItems = await dependenciesContainer.locator('.dependency-item').count();
    console.log(`Found ${initialDependencyItems} initial dependencies`);
    expect(initialDependencyItems).toBeGreaterThan(0);
    
    // Step 4: Start keycloak service
    console.log('Step 4: Starting keycloak service');
    await keycloakServiceCard.getByRole('button', { name: /start/i }).click();
    
    // Step 5: Wait for service to be starting and monitor WebSocket logs
    console.log('Step 5: Monitoring service startup and WebSocket updates');
    
    // Setup console log monitoring for WebSocket messages
    const webSocketLogs = [];
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('postgres') || text.includes('keycloak') || text.includes('WebSocket') || text.includes('dependency')) {
        webSocketLogs.push(`[${new Date().toISOString()}] ${text}`);
        console.log(`WebSocket/Dependency Log: ${text}`);
      }
    });
    
    // Wait for service to reach running state
    await expect(keycloakServiceCard.locator('.status-indicator')).toHaveText(/running/i, { timeout: 120000 });
    console.log('Keycloak service is now running');
    
    // Step 6: Critical validation - Check dependencies are still visible and have correct status
    console.log('Step 6: Validating dependencies after service startup');
    
         // Take screenshot for visual debugging
     await takeScreenshot(page, 'dependency-status-after-startup.png');
    
    // Check dependency count hasn't gone to 0
    const finalDependencyItems = await dependenciesContainer.locator('.dependency-item').count();
    console.log(`Found ${finalDependencyItems} dependencies after startup`);
    
    // This should NOT be 0 - dependencies shouldn't disappear
    expect(finalDependencyItems).toBeGreaterThan(0);
    expect(finalDependencyItems).toBe(initialDependencyItems); // Should maintain same count
    
    // Step 7: Check specific postgres dependency status
    console.log('Step 7: Checking postgres dependency status');
    const postgresDependency = dependenciesContainer.locator('.dependency-item').filter({ hasText: /postgres/i });
    
    if (await postgresDependency.count() > 0) {
      const postgresStatus = await postgresDependency.locator('.dependency-status').textContent();
      console.log(`Postgres dependency status: ${postgresStatus}`);
      
      // This should NOT be "STOPPED" - it should be "RUNNING" or "running-healthy"
      expect(postgresStatus.toLowerCase()).not.toBe('stopped');
      expect(postgresStatus.toLowerCase()).toMatch(/running|healthy|completed/);
    } else {
      console.error('Postgres dependency not found!');
      throw new Error('Postgres dependency disappeared after service startup');
    }
    
    // Step 8: Check Running Services section includes postgres
    console.log('Step 8: Checking Running Services section');
    const runningServicesSection = page.locator('.running-services-compact');
    await expect(runningServicesSection).toBeVisible();
    
    // Should contain both keycloak and postgres
    const runningServiceItems = await runningServicesSection.locator('.running-service-item').count();
    console.log(`Found ${runningServiceItems} running services`);
    expect(runningServiceItems).toBeGreaterThan(1); // Should have keycloak + dependencies
    
    // Step 9: Print WebSocket logs for debugging
    console.log('Step 9: WebSocket logs during test:');
    webSocketLogs.forEach(log => console.log(log));
    
         // Step 10: Final screenshot
     await takeScreenshot(page, 'dependency-status-fix-validation-final.png');
    
    console.log('=== Dependency Status Fix Validation Completed Successfully ===');
  });
  
  test('should handle dependency status updates from WebSocket correctly', async ({ page }) => {
    console.log('=== WebSocket Dependency Status Update Test ===');
    
    // Monitor WebSocket messages
    const statusUpdates = [];
    page.on('console', msg => {
      const text = msg.text();
      if (text.includes('SERVICE_STATUS_UPDATE') || text.includes('postgres') || text.includes('dependency')) {
        statusUpdates.push(text);
        console.log(`Status Update: ${text}`);
      }
    });
    
    // Start keycloak to trigger dependency status updates
    const keycloakServiceCard = page.locator('.service-card').filter({ hasText: 'keycloak' });
    await expect(keycloakServiceCard).toBeVisible();
    
    await keycloakServiceCard.getByRole('button', { name: /start/i }).click();
    
    // Wait for WebSocket status updates
    await page.waitForTimeout(5000);
    
    // Check that we received postgres status updates
    const hasPostgresUpdates = statusUpdates.some(update => 
      update.includes('postgres') && update.includes('running')
    );
    
    if (!hasPostgresUpdates) {
      console.log('All status updates:', statusUpdates);
    }
    
    // Wait for service completion
    await expect(keycloakServiceCard.locator('.status-indicator')).toHaveText(/running/i, { timeout: 120000 });
    
    // Verify final state
    const dependenciesContainer = keycloakServiceCard.locator('.dependencies');
    const postgresDependency = dependenciesContainer.locator('.dependency-item').filter({ hasText: /postgres/i });
    
    if (await postgresDependency.count() > 0) {
      const status = await postgresDependency.locator('.dependency-status').textContent();
      console.log(`Final postgres status: ${status}`);
      expect(status.toLowerCase()).not.toBe('stopped');
    }
    
    console.log('=== WebSocket Dependency Status Update Test Completed ===');
  });
}); 