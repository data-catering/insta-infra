/**
 * Test cleanup utilities for insta-infra E2E tests
 * Provides functions to clean up services and test state between tests
 */

import { request } from '@playwright/test';

/**
 * Stop all running services to ensure clean test state
 */
export async function stopAllServices() {
  try {
    const context = await request.newContext({
      baseURL: 'http://localhost:9310'
    });
    
    const response = await context.post('/api/v1/services/all/stop');
    
    if (!response.ok()) {
      console.warn(`Failed to stop all services: ${response.status()}`);
    }
    
    await context.dispose();
    
    // Wait for services to stop (reduced from 5s to 2s with 1-second polling)
    await new Promise(resolve => setTimeout(resolve, 2000));
  } catch (error) {
    console.warn('Error stopping all services:', error.message);
  }
}

/**
 * Stop a specific service
 */
export async function stopService(serviceName) {
  try {
    const context = await request.newContext({
      baseURL: 'http://localhost:9310'
    });
    
    const response = await context.post(`/api/v1/services/${serviceName}/stop`);
    
    if (!response.ok()) {
      console.warn(`Failed to stop service ${serviceName}: ${response.status()}`);
    }
    
    await context.dispose();
    
    // Wait for service to stop (reduced timing)
    await new Promise(resolve => setTimeout(resolve, 1500));
  } catch (error) {
    console.warn(`Error stopping service ${serviceName}:`, error.message);
  }
}

/**
 * Wait for all services to be in stopped state
 */
export async function waitForAllServicesStopped(timeout = 30000) {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    try {
      const context = await request.newContext({
        baseURL: 'http://localhost:9310'
      });
      
      const response = await context.get('/api/v1/services/running');
      const runningServices = await response.json();
      
      await context.dispose();
      
      // Check if any services are still running
      const stillRunning = Object.values(runningServices).some(
        service => service.status === 'running'
      );
      
      if (!stillRunning) {
        return true;
      }
      
      // Wait before checking again
      await new Promise(resolve => setTimeout(resolve, 1000));
    } catch (error) {
      console.warn('Error checking service status:', error.message);
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
  }
  
  return false;
}

/**
 * Clean up test environment by stopping all services and clearing any temp data
 */
export async function cleanupTestEnvironment() {
  console.log('Cleaning up test environment...');
  
  // Stop all services
  await stopAllServices();
  
  // Wait for services to be stopped
  const stopped = await waitForAllServicesStopped();
  
  if (!stopped) {
    console.warn('Some services may still be running after cleanup');
  } else {
    console.log('All services stopped successfully');
  }
}

/**
 * Perform cleanup using page interactions (for use in beforeEach/afterEach)
 */
export async function performCleanup(page) {
  try {
    console.log('Performing page-based cleanup...');
    
    // Wait for page to be ready
    await page.waitForSelector('.services-grid, .services-list', { timeout: 5000 });
    
    // Look for "Stop All" button and click it if visible
    const stopAllButton = page.locator('button').filter({ hasText: /stop all/i }).or(
      page.locator('button[title*="Stop all"]')
    ).first();
    
    // Check if Stop All button exists and is visible
    const buttonExists = await stopAllButton.count() > 0;
    if (buttonExists) {
      try {
        const isVisible = await stopAllButton.isVisible();
        if (isVisible) {
          await stopAllButton.click();
          console.log('Clicked Stop All button');
          
          // Wait for services to stop (faster with 1-second polling)
          await page.waitForTimeout(1500);
        }
      } catch (error) {
        console.log('Could not click Stop All button:', error.message);
      }
    }
    
    // Fallback: Use API cleanup
    await cleanupTestEnvironment();
    
    console.log('Cleanup completed');
  } catch (error) {
    console.warn('Error during page cleanup:', error.message);
    // Fallback to API cleanup
    await cleanupTestEnvironment();
  }
}

// Keep the original alias for backwards compatibility
export const cleanupTestEnvironment2 = cleanupTestEnvironment;

/**
 * Prepare clean test environment
 */
export async function prepareTestEnvironment() {
  console.log('Preparing test environment...');
  
  // Ensure all services are stopped
  await cleanupTestEnvironment();
  
  // Additional setup can be added here
  console.log('Test environment ready');
} 