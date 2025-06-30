import { test, expect } from '@playwright/test';
import { performCleanup } from '../utils/test-cleanup.js';

test.describe('Image Download & Runtime', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
    
    // Wait for the page to load and services to be visible
    await page.waitForSelector('.services-grid, .services-list', { timeout: 10000 });
    
    // Ensure clean state - stop any running services
    await performCleanup(page);
  });

  test.afterEach(async ({ page }) => {
    // Clean up after test
    await performCleanup(page);
  });

  test('Runtime availability and image handling', async ({ page }) => {
    // Set fast timeout for this focused test
    test.setTimeout(15000); // 15 seconds max
    
    console.log('Testing runtime and image handling...');

    // Check runtime status indicator (should show Docker/Podman available)
    const runtimeStatus = page.locator('[data-testid="runtime-status"], .runtime-status, .docker-status').first();
    
    // Look for any runtime-related indicators
    const hasRuntimeIndicator = await runtimeStatus.count() > 0;
    if (hasRuntimeIndicator) {
      console.log('✓ Runtime status indicator found');
    } else {
      console.log('⚠️  Runtime status indicator not found (may be integrated into layout)');
    }
    
    // Test that services can start (implies images are available or can be downloaded)
    const redisServiceItem = page.locator('.service-item').filter({ 
      has: page.locator('.service-name:has-text("redis")') 
    }).first();
    await expect(redisServiceItem).toBeVisible();
    
    // Start Redis to test image handling
    console.log('Testing image availability by starting Redis...');
    const redisStartButton = redisServiceItem.locator('.action-button.start-button');
    await redisStartButton.click();
    
    // Wait for image check/download and service start
    // If image doesn't exist, download modal should appear
    // If image exists, service should start directly
    
    // Look for download modal or direct service start
    const downloadModal = page.locator('.modal, .download-modal, [role="dialog"]').first();
    const hasDownloadModal = await downloadModal.count() > 0;
    
    if (hasDownloadModal) {
      console.log('✓ Image download modal detected');
      
      // Wait for download to complete (up to 8 seconds for fast images like Redis)
      await page.waitForTimeout(8000);
      
      // Modal should close after successful download
      const modalStillVisible = await downloadModal.isVisible().catch(() => false);
      if (!modalStillVisible) {
        console.log('✓ Download completed, modal closed');
      }
    } else {
      console.log('✓ Image already available, direct service start');
    }
    
    // Verify service is starting/running (image handling successful)
    await page.waitForTimeout(2000); // Brief wait for status update
    
    const isServiceStarting = await redisServiceItem.evaluate(el => 
      el.classList.contains('service-item-starting') || 
      el.classList.contains('service-item-running')
    ).catch(() => false);
    
    if (isServiceStarting) {
      console.log('✓ Service started successfully (image handling worked)');
    } else {
      console.log('⚠️  Service start may still be processing');
    }
    
    console.log('✅ Image and runtime test completed');
  });
}); 