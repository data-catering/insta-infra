import { test, expect } from '@playwright/test';
import { performCleanup } from '../utils/test-cleanup.js';

// Enhanced utility function for detailed status monitoring
async function captureStatusTransitions(page, serviceName, actionFn, expectedFinalStatus = 'running', maxWaitTime = 15000) {
  const transitions = [];
  const screenshots = [];
  const startTime = Date.now();
  
  // Find the service item
  const serviceItem = page.locator('.service-item').filter({ 
    has: page.locator(`.service-name:has-text("${serviceName}")`) 
  }).first();
  
  console.log(`🔍 Starting status capture for ${serviceName}...`);
  
  // Capture initial state
  const initialClasses = await serviceItem.getAttribute('class');
  const initialStatusText = await serviceItem.locator('.service-status').textContent();
  transitions.push({
    timestamp: 0,
    classes: initialClasses,
    statusText: initialStatusText,
    description: 'Initial state'
  });
  
  // Take screenshot of initial state
  await page.screenshot({
    path: `test-results/status-${serviceName}-00-initial.png`,
    fullPage: true
  });
  screenshots.push('initial');
  
  // Execute the action (start/stop)
  console.log(`⚡ Executing action for ${serviceName}...`);
  await actionFn();
  
  // Monitor status changes
  let lastClasses = initialClasses;
  let lastStatusText = initialStatusText;
  let screenshotCounter = 1;
  
  while (Date.now() - startTime < maxWaitTime) {
    await page.waitForTimeout(100); // Check every 100ms for fine-grained monitoring
    
    try {
      const currentClasses = await serviceItem.getAttribute('class');
      const currentStatusText = await serviceItem.locator('.service-status').textContent();
      
      // Check if state changed
      if (currentClasses !== lastClasses || currentStatusText !== lastStatusText) {
        const timeSinceStart = Date.now() - startTime;
        
        transitions.push({
          timestamp: timeSinceStart,
          classes: currentClasses,
          statusText: currentStatusText,
          description: `Status change at ${timeSinceStart}ms`
        });
        
        // Take screenshot of state change
        const screenshotName = `status-${serviceName}-${screenshotCounter.toString().padStart(2, '0')}-${timeSinceStart}ms`;
        await page.screenshot({
          path: `test-results/${screenshotName}.png`,
          fullPage: true
        });
        screenshots.push(screenshotName);
        
        console.log(`📸 Captured ${screenshotName}: ${currentStatusText} (classes: ${currentClasses})`);
        
        lastClasses = currentClasses;
        lastStatusText = currentStatusText;
        screenshotCounter++;
      }
      
      // Check if we've reached the expected final status
      if (currentClasses?.includes(`service-item-${expectedFinalStatus}`) || 
          currentStatusText?.toLowerCase().includes(expectedFinalStatus)) {
        console.log(`✅ Reached expected final status: ${expectedFinalStatus}`);
        break;
      }
      
    } catch (error) {
      console.log(`⚠️  Error during monitoring: ${error.message}`);
    }
  }
  
  // Take final screenshot
  await page.screenshot({
    path: `test-results/status-${serviceName}-${screenshotCounter.toString().padStart(2, '0')}-final.png`,
    fullPage: true
  });
  screenshots.push('final');
  
  return { transitions, screenshots };
}

// Utility to check for flashing behavior
function analyzeFlashingBehavior(transitions) {
  const issues = [];
  
  // Look for rapid back-and-forth status changes
  for (let i = 1; i < transitions.length - 1; i++) {
    const prev = transitions[i - 1];
    const current = transitions[i];
    const next = transitions[i + 1];
    
    // Check for status reverting back to previous state
    if (prev.statusText === next.statusText && prev.statusText !== current.statusText) {
      issues.push({
        type: 'flash',
        description: `Status flashed from "${prev.statusText}" to "${current.statusText}" and back to "${next.statusText}"`,
        timestamps: [prev.timestamp, current.timestamp, next.timestamp]
      });
    }
    
    // Check for rapid consecutive changes (within 500ms)
    if (current.timestamp - prev.timestamp < 500) {
      issues.push({
        type: 'rapid_change',
        description: `Rapid status change from "${prev.statusText}" to "${current.statusText}" in ${current.timestamp - prev.timestamp}ms`,
        timestamps: [prev.timestamp, current.timestamp]
      });
    }
  }
  
  return issues;
}

test.describe('Service Status Transitions and Health Checks', () => {
  test.beforeEach(async ({ page }) => {

    
    // Navigate to the application
    await page.goto('/');
    
    // Wait for the page to load and services to be visible
    await page.waitForSelector('.services-grid, .services-list', { timeout: 10000 });
    
    // Ensure clean state
    await performCleanup(page);
  });

  test.afterEach(async ({ page }) => {
    // Clean up after test
    await performCleanup(page);
  });

  test('Investigate UI flashing during Jaeger service start', async ({ page }) => {
    test.setTimeout(60000); // 1 minute for comprehensive testing
    
    console.log('🧪 Starting Jaeger UI flashing investigation...');

    // Capture Jaeger service transitions
    const result = await captureStatusTransitions(
      page,
      'jaeger',
      async () => {
        const jaegerServiceItem = page.locator('.service-item').filter({ 
          has: page.locator('.service-name:has-text("jaeger")') 
        }).first();
        
        const startButton = jaegerServiceItem.locator('.action-button.start-button');
        await expect(startButton).toBeVisible();
        await startButton.click();
      },
      'running',
      20000 // 20 seconds timeout
    );
    
    console.log(`📊 Captured ${result.transitions.length} status transitions`);
    console.log(`📸 Generated ${result.screenshots.length} screenshots`);
    
    // Analyze for flashing behavior
    const flashingIssues = analyzeFlashingBehavior(result.transitions);
    
    if (flashingIssues.length > 0) {
      console.log('🚨 FLASHING DETECTED:');
      flashingIssues.forEach((issue, index) => {
        console.log(`  ${index + 1}. ${issue.type}: ${issue.description}`);
      });
      
      // Fail the test if flashing is detected
      throw new Error(`UI flashing detected! Found ${flashingIssues.length} issues. Check screenshots in test-results/`);
    } else {
      console.log('✅ No UI flashing detected');
    }
    
    // Print transition timeline
    console.log('\n📈 Status Transition Timeline:');
    result.transitions.forEach((transition, index) => {
      console.log(`  ${index + 1}. [${transition.timestamp}ms] ${transition.statusText} (${transition.description})`);
    });
  });

  test('Test different service health check scenarios', async ({ page }) => {
    test.setTimeout(90000); // 90 seconds for health check testing
    
    console.log('🏥 Testing health check scenarios...');

    // Test services with different health check configurations
    const servicesToTest = [
      { name: 'postgres', expectsHealthCheck: true },
      { name: 'redis', expectsHealthCheck: false },
      { name: 'jaeger', expectsHealthCheck: true }
    ];
    
    for (const serviceConfig of servicesToTest) {
      console.log(`\n🔬 Testing ${serviceConfig.name} (health check: ${serviceConfig.expectsHealthCheck})`);
      
      const result = await captureStatusTransitions(
        page,
        serviceConfig.name,
        async () => {
          const serviceItem = page.locator('.service-item').filter({ 
            has: page.locator(`.service-name:has-text("${serviceConfig.name}")`) 
          }).first();
          
          const startButton = serviceItem.locator('.action-button.start-button');
          await expect(startButton).toBeVisible();
          await startButton.click();
        },
        'running',
        25000 // 25 seconds for health checks to pass
      );
      
      // Analyze health check states
      const healthStates = result.transitions.map(t => ({
        timestamp: t.timestamp,
        status: t.statusText,
        isStarting: t.classes?.includes('starting') || t.statusText?.includes('starting'),
        isRunning: t.classes?.includes('running') || t.statusText?.includes('running'),
        isUnhealthy: t.classes?.includes('unhealthy') || t.statusText?.includes('unhealthy'),
        isHealthy: t.classes?.includes('healthy') || t.statusText?.includes('healthy')
      }));
      
      console.log(`📋 Health states for ${serviceConfig.name}:`);
      healthStates.forEach((state, index) => {
        console.log(`  ${index + 1}. [${state.timestamp}ms] ${state.status} (starting: ${state.isStarting}, running: ${state.isRunning}, unhealthy: ${state.isUnhealthy}, healthy: ${state.isHealthy})`);
      });
      
      // Stop the service before testing the next one
      const serviceItem = page.locator('.service-item').filter({ 
        has: page.locator(`.service-name:has-text("${serviceConfig.name}")`) 
      }).first();
      
      const stopButton = serviceItem.locator('.action-button.stop-button, .action-button').filter({ hasText: /stop/i }).first();
      const stopButtonCount = await stopButton.count();
      if (stopButtonCount > 0) {
        await stopButton.click();
        await page.waitForTimeout(2000); // Wait for stop to complete
      } else {
        await performCleanup(page);
      }
    }
  });

  test('Test rapid start/stop cycles for race conditions', async ({ page }) => {
    test.setTimeout(60000);
    
    console.log('⚡ Testing rapid start/stop cycles...');
    
    const serviceName = 'redis'; // Use a fast-starting service
    const serviceItem = page.locator('.service-item').filter({ 
      has: page.locator(`.service-name:has-text("${serviceName}")`) 
    }).first();
    
    // Perform rapid start/stop cycles
    for (let cycle = 1; cycle <= 3; cycle++) {
      console.log(`🔄 Cycle ${cycle}: Starting ${serviceName}...`);
      
      const startButton = serviceItem.locator('.action-button.start-button');
      await expect(startButton).toBeVisible();
      await startButton.click();
      
      // Wait briefly
      await page.waitForTimeout(1000);
      
      console.log(`🔄 Cycle ${cycle}: Stopping ${serviceName}...`);
      
      const stopButton = serviceItem.locator('.action-button.stop-button, .action-button').filter({ hasText: /stop/i }).first();
      const stopButtonCount = await stopButton.count();
      if (stopButtonCount > 0) {
        await stopButton.click();
      } else {
        console.log('⚠️  Using Stop All as fallback');
        await performCleanup(page);
      }
      
      // Wait for cycle to complete
      await page.waitForTimeout(2000);
      
      // Take screenshot of current state
      await page.screenshot({
        path: `test-results/rapid-cycle-${cycle}.png`,
        fullPage: true
      });
    }
    
    console.log('✅ Rapid cycle testing completed');
  });
}); 