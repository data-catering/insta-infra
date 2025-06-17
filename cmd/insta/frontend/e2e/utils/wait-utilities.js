/**
 * Wait utilities for insta-infra E2E tests
 * Provides smart waiting functions for service states and UI transitions
 */

/**
 * Wait for a service to reach a specific status
 */
export async function waitForServiceStatus(page, serviceName, expectedStatus, timeout = 60000) {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    try {
      // Look for the service card/item
      const serviceElement = page.locator(`[data-testid="service-${serviceName}"], .service-item:has-text("${serviceName}"), .service-card:has-text("${serviceName}")`).first();
      
      if (await serviceElement.isVisible()) {
        // Check for status indicators
        const statusIndicators = {
          'running': ['text=running', 'text=active', '[class*="running"]', '[class*="active"]', '.status-running'],
          'stopped': ['text=stopped', 'text=inactive', '[class*="stopped"]', '[class*="inactive"]', '.status-stopped'],
          'starting': ['text=starting', '[class*="starting"]', '.status-starting'],
          'stopping': ['text=stopping', '[class*="stopping"]', '.status-stopping'],
          'failed': ['text=failed', 'text=error', '[class*="failed"]', '[class*="error"]', '.status-failed']
        };
        
        const indicators = statusIndicators[expectedStatus] || [];
        
        for (const indicator of indicators) {
          if (await serviceElement.locator(indicator).isVisible()) {
            return true;
          }
        }
      }
      
      // Wait before checking again
      await page.waitForTimeout(1000);
    } catch (error) {
      // Continue checking
      await page.waitForTimeout(1000);
    }
  }
  
  throw new Error(`Service ${serviceName} did not reach status ${expectedStatus} within ${timeout}ms`);
}

/**
 * Wait for multiple services to reach specific statuses
 */
export async function waitForMultipleServiceStatuses(page, serviceStatuses, timeout = 60000) {
  const promises = Object.entries(serviceStatuses).map(([serviceName, expectedStatus]) =>
    waitForServiceStatus(page, serviceName, expectedStatus, timeout)
  );
  
  await Promise.all(promises);
}

/**
 * Wait for UI element to be visible and stable
 */
export async function waitForStableElement(page, selector, timeout = 10000) {
  // First wait for element to be visible
  await page.waitForSelector(selector, { timeout });
  
  // Then wait for it to be stable (not changing position/size)
  const element = page.locator(selector);
  let previousBox = null;
  let stableCount = 0;
  const requiredStableChecks = 3;
  
  while (stableCount < requiredStableChecks) {
    await page.waitForTimeout(100);
    
    try {
      const currentBox = await element.boundingBox();
      
      if (previousBox && 
          currentBox.x === previousBox.x && 
          currentBox.y === previousBox.y && 
          currentBox.width === previousBox.width && 
          currentBox.height === previousBox.height) {
        stableCount++;
      } else {
        stableCount = 0;
      }
      
      previousBox = currentBox;
    } catch (error) {
      // Element might not be visible yet, reset counter
      stableCount = 0;
    }
  }
  
  return element;
}

/**
 * Wait for modal to appear and be ready for interaction
 */
export async function waitForModal(page, modalSelector, timeout = 10000) {
  // Wait for modal to appear
  await page.waitForSelector(modalSelector, { timeout });
  
  // Wait for modal animation/transition to complete
  await page.waitForTimeout(500);
  
  // Ensure modal is visible and interactive
  const modal = page.locator(modalSelector);
  await modal.waitFor({ state: 'visible' });
  
  return modal;
}

/**
 * Wait for progress to update (for download/loading scenarios)
 */
export async function waitForProgressUpdate(page, progressSelector, timeout = 30000) {
  const startTime = Date.now();
  let lastProgressValue = null;
  
  while (Date.now() - startTime < timeout) {
    try {
      const progressElement = page.locator(progressSelector);
      
      if (await progressElement.isVisible()) {
        // Try to get progress value from various possible attributes/text
        let currentProgress = null;
        
        // Check for value attribute (progress element)
        try {
          currentProgress = await progressElement.getAttribute('value');
        } catch {}
        
        // Check for text content (percentage)
        if (!currentProgress) {
          try {
            const text = await progressElement.textContent();
            const match = text.match(/(\d+)%/);
            currentProgress = match ? match[1] : null;
          } catch {}
        }
        
        // Check for style width (progress bar)
        if (!currentProgress) {
          try {
            const style = await progressElement.getAttribute('style');
            const match = style.match(/width:\s*(\d+)%/);
            currentProgress = match ? match[1] : null;
          } catch {}
        }
        
        if (currentProgress && currentProgress !== lastProgressValue) {
          // Progress has updated
          return true;
        }
        
        lastProgressValue = currentProgress;
      }
      
      await page.waitForTimeout(1000);
    } catch (error) {
      await page.waitForTimeout(1000);
    }
  }
  
  throw new Error(`Progress did not update within ${timeout}ms`);
}

/**
 * Wait for service logs to appear with content
 */
export async function waitForServiceLogs(page, serviceName, timeout = 15000) {
  const startTime = Date.now();
  
  while (Date.now() - startTime < timeout) {
    try {
      // Look for logs modal or logs content
      const logsSelectors = [
        '[data-testid="logs-modal"]',
        '[class*="logs-modal"]',
        '[class*="logs-content"]',
        '.logs-panel',
        '.logs-container'
      ];
      
      for (const selector of logsSelectors) {
        const logsElement = page.locator(selector);
        
        if (await logsElement.isVisible()) {
          // Check if logs have content (not just empty)
          const logText = await logsElement.textContent();
          
          if (logText && logText.trim().length > 0) {
            // Found logs with content
            return logsElement;
          }
        }
      }
      
      await page.waitForTimeout(1000);
    } catch (error) {
      await page.waitForTimeout(1000);
    }
  }
  
  throw new Error(`Service logs for ${serviceName} did not appear with content within ${timeout}ms`);
}

/**
 * Wait for network idle (useful after service operations)
 */
export async function waitForNetworkIdle(page, timeout = 5000) {
  await page.waitForLoadState('networkidle', { timeout });
}

/**
 * Smart wait that combines multiple waiting strategies
 */
export async function smartWait(page, options = {}) {
  const {
    selector,
    state = 'visible',
    timeout = 10000,
    stable = false,
    networkIdle = false
  } = options;
  
  if (selector) {
    await page.waitForSelector(selector, { state, timeout });
    
    if (stable) {
      await waitForStableElement(page, selector, timeout);
    }
  }
  
  if (networkIdle) {
    await waitForNetworkIdle(page, timeout);
  }
} 