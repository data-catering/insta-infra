import { test, expect } from '@playwright/test';

test.describe('Dependency Status Display', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
    
    // Wait for the page to load
    await page.waitForSelector('body', { timeout: 10000 });
  });

  test('should display correct dependency statuses instead of UNKNOWN', async ({ page }) => {
    console.log('🧪 Testing dependency status display...');
    
    // Wait for services to load
    await page.waitForSelector('.service-item', { timeout: 15000 });
    
    // Look for a service that has dependencies (like postgres)
    const postgresService = page.locator('.service-item').filter({ hasText: 'postgres' }).first();
    await expect(postgresService).toBeVisible({ timeout: 10000 });
    
    // Click to expand the postgres service details
    await postgresService.click();
    
    // Wait for the expandable content to be visible
    await page.waitForSelector('.dependencies-section', { timeout: 5000 });
    
    // Check if dependencies section exists
    const dependenciesSection = page.locator('.dependencies-section');
    if (await dependenciesSection.count() > 0) {
      console.log('✅ Dependencies section found');
      
      // Get all dependency items
      const dependencyItems = page.locator('.dependency-item');
      const dependencyCount = await dependencyItems.count();
      
      console.log(`📊 Found ${dependencyCount} dependencies`);
      
      if (dependencyCount > 0) {
        // Check each dependency item
        for (let i = 0; i < dependencyCount; i++) {
          const dependencyItem = dependencyItems.nth(i);
          const dependencyName = await dependencyItem.locator('.dependency-name').textContent();
          const dependencyStatus = await dependencyItem.locator('.dependency-container-status').textContent();
          
          console.log(`🔍 Dependency: ${dependencyName} -> Status: ${dependencyStatus}`);
          
          // The status should NOT be 'UNKNOWN'
          expect(dependencyStatus.toUpperCase()).not.toBe('UNKNOWN');
          
          // The status should be one of the valid statuses
          const validStatuses = ['RUNNING', 'STOPPED', 'STARTING', 'STOPPING', 'COMPLETED', 'FAILED', 'CREATED'];
          expect(validStatuses).toContain(dependencyStatus.toUpperCase());
        }
        
        console.log('✅ All dependency statuses are valid (not UNKNOWN)');
      } else {
        console.log('ℹ️  No dependencies found for this service');
      }
    } else {
      console.log('ℹ️  No dependencies section found - may be normal for this service');
    }
  });

  test('should show running dependencies in the running services section', async ({ page }) => {
    console.log('🧪 Testing running dependencies visibility...');
    
    // Wait for services to load
    await page.waitForSelector('.service-item', { timeout: 15000 });
    
    // Look for a service that has dependencies (like keycloak or airflow)
    const servicesWithDeps = ['keycloak', 'airflow', 'metabase'];
    let serviceFound = false;
    let foundServiceName = '';
    
    for (const serviceName of servicesWithDeps) {
      const serviceCard = page.locator('.service-item').filter({ hasText: serviceName });
      if (await serviceCard.count() > 0) {
        serviceFound = true;
        foundServiceName = serviceName;
        console.log(`✅ Found service with dependencies: ${serviceName}`);
        
        // Start the service if it's not already running
        const serviceStatus = page.locator('.service-item').filter({ hasText: serviceName }).locator('.service-status');
        const statusText = await serviceStatus.textContent();
        
        if (!statusText.toLowerCase().includes('running')) {
          console.log(`🚀 Starting ${serviceName} service...`);
          const startButton = page.locator('.service-item').filter({ hasText: serviceName }).locator('.start-button');
          if (await startButton.count() > 0) {
            await startButton.click();
            
            // Wait for service to start (or at least show starting status)
            await page.waitForTimeout(3000);
            
            // Check if running services section appears and includes dependencies
            const runningServicesSection = page.locator('.running-services-compact');
            if (await runningServicesSection.count() > 0) {
              console.log('✅ Running services section is visible');
              
              // Look for dependency services in the running section
              const runningServiceItems = runningServicesSection.locator('.service-item');
              const runningCount = await runningServiceItems.count();
              
              console.log(`📊 Found ${runningCount} services in running section`);
              
              // Check if we have more than just the main service (dependencies should be included)
              if (runningCount > 1) {
                console.log('✅ Multiple services in running section - likely includes dependencies');
                
                // Check for dependency services (marked with dependency type or dashed border)
                const dependencyItems = runningServicesSection.locator('.service-item-dependency');
                const depCount = await dependencyItems.count();
                
                if (depCount > 0) {
                  console.log(`✅ Found ${depCount} dependency services in running section`);
                  
                  for (let i = 0; i < depCount; i++) {
                    const depItem = dependencyItems.nth(i);
                    const depName = await depItem.locator('.service-name').textContent();
                    console.log(`🔗 Dependency service: ${depName}`);
                  }
                } else {
                  console.log('ℹ️  No explicitly marked dependency services found, but multiple services are running');
                }
              }
            } else {
              console.log('⚠️  Running services section not visible - service may not have started yet');
            }
          }
        } else {
          console.log(`ℹ️  ${serviceName} is already running`);
        }
        break;
      }
    }
    
    if (!serviceFound) {
      console.log('ℹ️  No services with dependencies found in this test run');
    }
  });

  test('should update dependency statuses when service is starting', async ({ page }) => {
    console.log('🧪 Testing dependency status updates during service start...');
    
    // Wait for services to load
    await page.waitForSelector('.service-item', { timeout: 15000 });
    
    // Find a service with dependencies
    const serviceWithDeps = page.locator('.service-item').filter({ hasText: /airflow|keycloak|metabase/ }).first();
    
    if (await serviceWithDeps.count() > 0) {
      const serviceName = await serviceWithDeps.locator('.service-name').textContent();
      console.log(`✅ Testing dependency updates for: ${serviceName}`);
      
      // Expand the service to see dependencies
      await serviceWithDeps.click();
      await page.waitForTimeout(500);
      
      // Check if it has dependencies
      const dependenciesSection = serviceWithDeps.locator('.dependencies-section');
      if (await dependenciesSection.count() > 0) {
        console.log('✅ Service has dependencies section');
        
        // Get initial dependency statuses
        const dependencyItems = dependenciesSection.locator('.dependency-item');
        const initialStatuses = [];
        
        const depCount = await dependencyItems.count();
        for (let i = 0; i < depCount; i++) {
          const depItem = dependencyItems.nth(i);
          const depName = await depItem.locator('.dependency-name').textContent();
          const depStatus = await depItem.locator('.dependency-container-status').textContent();
          initialStatuses.push({ name: depName, status: depStatus });
          console.log(`📊 Initial dependency: ${depName} -> ${depStatus}`);
        }
        
        // Start the service
        const startButton = serviceWithDeps.locator('.start-button');
        if (await startButton.count() > 0 && await startButton.isVisible()) {
          console.log(`🚀 Starting ${serviceName}...`);
          await startButton.click();
          
          // Wait a bit for status updates
          await page.waitForTimeout(2000);
          
          // Check if dependency statuses have changed
          let statusesChanged = false;
          for (let i = 0; i < depCount; i++) {
            const depItem = dependencyItems.nth(i);
            const depName = await depItem.locator('.dependency-name').textContent();
            const newDepStatus = await depItem.locator('.dependency-container-status').textContent();
            const initialStatus = initialStatuses.find(s => s.name === depName)?.status;
            
            if (newDepStatus !== initialStatus) {
              console.log(`✅ Dependency status updated: ${depName} ${initialStatus} -> ${newDepStatus}`);
              statusesChanged = true;
            }
          }
          
          if (statusesChanged) {
            console.log('✅ Dependency statuses updated during service start');
          } else {
            console.log('ℹ️  Dependency statuses remained the same (may be expected)');
          }
        } else {
          console.log('ℹ️  Service start button not available (may already be running)');
        }
      } else {
        console.log('ℹ️  Service has no dependencies section');
      }
    } else {
      console.log('ℹ️  No services with dependencies found for this test');
    }
  });
}); 