import { test, expect } from '@playwright/test';

/**
 * Setup verification test for insta-infra Web UI
 * This test ensures the basic test infrastructure is working correctly
 */

test.describe('Setup Verification', () => {
  
  test.beforeEach(async ({ page }) => {
    // Navigate to the web UI
    await page.goto('/');
    
    // Wait for the main UI to be ready
    await expect(page.locator('body')).toBeVisible();
    await page.waitForTimeout(2000); // Give the app time to initialize
  });

  test('should load the application successfully', async ({ page }) => {
    // Check that the main app loaded
    await expect(page.locator('body')).toBeVisible();
    
    // Check for the page title or main container
    const pageTitle = page.locator('title');
    if (await pageTitle.count() > 0) {
      const title = await pageTitle.textContent();
      expect(title).toBeTruthy();
    }
    
    // Verify the app is interactive (not just a blank page)
    const hasContent = await page.locator('*').count() > 1;
    expect(hasContent).toBe(true);
  });

  test('should respond to basic interactions', async ({ page }) => {
    // Wait for the page to load
    await page.waitForTimeout(1000);
    
    // Try to find any clickable elements
    const clickableElements = await page.locator('button, [role="button"], a').count();
    
    // The app should have some interactive elements
    expect(clickableElements).toBeGreaterThan(0);
    
    // Verify the page doesn't crash when we try to interact
    if (clickableElements > 0) {
      const firstButton = page.locator('button, [role="button"], a').first();
      await firstButton.focus();
      
      // App should still be responsive after interaction
      await expect(page.locator('body')).toBeVisible();
    }
  });

  test('should handle window resize gracefully', async ({ page }) => {
    // Start with a standard size
    await page.setViewportSize({ width: 1280, height: 720 });
    await page.waitForTimeout(500);
    
    // Verify app is still visible
    await expect(page.locator('body')).toBeVisible();
    
    // Resize to smaller (mobile-like)
    await page.setViewportSize({ width: 375, height: 667 });
    await page.waitForTimeout(500);
    
    // App should still be functional
    await expect(page.locator('body')).toBeVisible();
    
    // Resize back to larger
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    
    // App should still be functional
    await expect(page.locator('body')).toBeVisible();
  });
}); 