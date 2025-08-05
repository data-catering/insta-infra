import { test, expect } from '@playwright/test';

test.describe('Runtime Startup', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
    
    // Wait for the page to load
    await page.waitForSelector('body', { timeout: 10000 });
  });

  test('should successfully start Docker runtime without showing error', async ({ page }) => {
    console.log('🧪 Testing runtime startup functionality...');
    
    // Wait for runtime setup UI to appear (when no runtime is available)
    const setupSection = page.locator('text=Container Runtime Setup');
    await expect(setupSection).toBeVisible({ timeout: 15000 });
    
    // Look for the "Start Docker" button
    const startDockerButton = page.locator('button:has-text("Start Docker")');
    await expect(startDockerButton).toBeVisible();
    
    // Intercept the runtime start API call to verify the request/response
    const apiResponsePromise = page.waitForResponse(response => 
      response.url().includes('/api/v1/runtime/start') && response.request().method() === 'POST'
    );
    
    console.log('🔄 Clicking Start Docker button...');
    await startDockerButton.click();
    
    // Wait for the API response
    const apiResponse = await apiResponsePromise;
    console.log('📡 API Response Status:', apiResponse.status());
    
    // Verify API returns 200
    expect(apiResponse.status()).toBe(200);
    
    // Get the response body
    const responseBody = await apiResponse.json();
    console.log('📋 API Response Body:', JSON.stringify(responseBody, null, 2));
    
    // Verify API response structure
    expect(responseBody).toHaveProperty('action', 'start');
    expect(responseBody).toHaveProperty('runtime', 'docker');
    expect(responseBody).toHaveProperty('status', 'success');
    expect(responseBody).toHaveProperty('message');
    
    // Wait a moment for UI to process the response
    await page.waitForTimeout(2000);
    
    // Check for error messages in the UI
    const setupError = page.locator('text=Setup Error').or(page.locator('text=Failed to start runtime'));
    
    // Take a screenshot for debugging
    await page.screenshot({
      path: 'test-results/runtime-startup-after-click.png',
      fullPage: true
    });
    
    // The main test: UI should NOT show error when API returns success
    const hasError = await setupError.isVisible();
    if (hasError) {
      const errorText = await setupError.textContent();
      console.log('❌ Error found in UI:', errorText);
      
      // Log all visible text for debugging
      const allText = await page.locator('body').textContent();
      console.log('🔍 All visible text:', allText);
      
      // This should fail the test
      expect(hasError).toBe(false);
    } else {
      console.log('✅ No error message found in UI - test passed!');
    }
    
    // Verify that startup progress is shown instead of error
    const startupProgress = page.locator('text=Starting docker').or(
      page.locator('text=Docker Desktop is starting')
    );
    
    // Should show progress message, not error
    await expect(startupProgress).toBeVisible({ timeout: 5000 });
    
    console.log('🎉 Runtime startup test completed successfully!');
  });
  
  test('should handle API errors gracefully', async ({ page }) => {
    console.log('🧪 Testing API error handling...');
    
    // Wait for runtime setup UI
    const setupSection = page.locator('text=Container Runtime Setup');
    await expect(setupSection).toBeVisible({ timeout: 15000 });
    
    // Mock a failing API response
    await page.route('**/api/v1/runtime/start', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          action: 'start',
          runtime: 'docker',
          status: 'failed',
          error: 'Docker Desktop not found'
        })
      });
    });
    
    const startDockerButton = page.locator('button:has-text("Start Docker")');
    await startDockerButton.click();
    
    // Should show error when API actually fails
    const setupError = page.locator('text=Setup Error').or(page.locator('text=Failed to start docker'));
    await expect(setupError).toBeVisible({ timeout: 5000 });
    
    console.log('✅ Error handling test passed!');
  });
}); 