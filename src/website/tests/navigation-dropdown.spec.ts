import { test, expect } from '@playwright/test';

test.describe('Navigation Dropdown Functionality', () => {
  test.beforeEach(async ({ page }) => {
    // Enable detailed console logging
    page.on('console', msg => {
      console.log(`BROWSER: ${msg.type()}: ${msg.text()}`);
    });

    // Listen for errors
    page.on('pageerror', error => {
      console.log(`PAGE ERROR: ${error.message}`);
    });

    // Navigate to homepage
    await page.goto('/');

    // Wait for the page to fully load
    await page.waitForLoadState('domcontentloaded');
    await page.waitForLoadState('networkidle');
  });

  test('should render navigation dropdowns and handle clicks', async ({ page }) => {
    console.log('🎯 Starting dropdown functionality test');

    // Wait for the navigation component to be visible
    await page.waitForSelector('nav', { state: 'visible', timeout: 10000 });
    console.log('✅ Navigation element found');

    // Check if Services dropdown button exists
    const servicesButton = page.locator('nav button:has-text("Services")');
    await expect(servicesButton).toBeVisible();
    console.log('✅ Services dropdown button is visible');

    // Check if Patient Resources dropdown button exists
    const patientResourcesButton = page.locator('nav button:has-text("Patient Resources")');
    await expect(patientResourcesButton).toBeVisible();
    console.log('✅ Patient Resources dropdown button is visible');

    // Check if Community dropdown button exists
    const communityButton = page.locator('nav button:has-text("Community")');
    await expect(communityButton).toBeVisible();
    console.log('✅ Community dropdown button is visible');

    // Check if Company dropdown button exists
    const companyButton = page.locator('nav button:has-text("Company")');
    await expect(companyButton).toBeVisible();
    console.log('✅ Company dropdown button is visible');

    // Test clicking Services dropdown
    console.log('🖱️ Clicking Services dropdown...');
    await servicesButton.click();

    // Wait a moment for the dropdown to potentially appear
    await page.waitForTimeout(1000);

    // Check if any dropdown content appeared
    const dropdownContent = page.locator(
      '[data-testid="navigation-dropdown"], .dropdown-content, .absolute'
    );
    const dropdownCount = await dropdownContent.count();
    console.log(`📊 Found ${dropdownCount} potential dropdown elements`);

    if (dropdownCount > 0) {
      console.log('✅ Dropdown content detected');
      for (let i = 0; i < dropdownCount; i++) {
        const element = dropdownContent.nth(i);
        const isVisible = await element.isVisible();
        console.log(`  - Element ${i}: visible = ${isVisible}`);
      }
    } else {
      console.log('❌ No dropdown content detected');
    }

    // Test Patient Resources dropdown (should work even without APIs)
    console.log('🖱️ Clicking Patient Resources dropdown...');
    await patientResourcesButton.click();

    // Wait a moment for the dropdown to potentially appear
    await page.waitForTimeout(1000);

    // Check console logs for our debug messages
    await page.evaluate(() => {
      console.log('🔍 Manual check: activeDropdown state from browser');
    });

    // Test Community dropdown
    console.log('🖱️ Clicking Community dropdown...');
    await communityButton.click();

    await page.waitForTimeout(1000);

    // Test Company dropdown
    console.log('🖱️ Clicking Company dropdown...');
    await companyButton.click();

    await page.waitForTimeout(1000);

    // Get component mounting status
    const componentStatus = await page.evaluate(() => {
      // Check if React component debugging info is available
      return {
        hasReactDevtools: typeof window.__REACT_DEVTOOLS_GLOBAL_HOOK__ !== 'undefined',
        timestamp: new Date().toISOString(),
      };
    });

    console.log('🔧 Component status:', componentStatus);
  });

  test('should detect component hydration issues', async ({ page }) => {
    console.log('🔍 Testing component hydration and event handling');

    await page.waitForSelector('nav', { state: 'visible' });

    // Check if the component is properly hydrated by testing React event handlers
    const servicesButton = page.locator('nav button:has-text("Services")');

    // Get button properties to check if React event handlers are attached
    const buttonProps = await servicesButton.evaluate(button => {
      const events = [];
      for (const key in button) {
        if (key.startsWith('__reactInternalInstance') || key.startsWith('_reactInternalInstance')) {
          events.push(key);
        }
      }
      return {
        hasOnClick: button.onclick !== null,
        hasEventListeners: events.length > 0,
        className: button.className,
        hasReactFiber: Object.keys(button).some(
          key => key.startsWith('__reactInternalFiber') || key.startsWith('__reactFiber')
        ),
      };
    });

    console.log('🔧 Button React properties:', buttonProps);

    // Test if clicking triggers any events
    let clickEventTriggered = false;
    await page.exposeFunction('onButtonClick', () => {
      clickEventTriggered = true;
      console.log('✅ Click event was triggered');
    });

    // Add a click event listener to detect if clicks are being handled
    await page.evaluate(() => {
      const button = document.querySelector('nav button');
      if (button) {
        button.addEventListener('click', () => {
          (window as any).onButtonClick();
        });
      }
    });

    await servicesButton.click();
    await page.waitForTimeout(500);

    if (clickEventTriggered) {
      console.log('✅ DOM click events are working');
    } else {
      console.log('❌ DOM click events are not working');
    }
  });

  test('should check component mount state and API loading', async ({ page }) => {
    console.log('🔍 Checking component mount state and API loading status');

    await page.waitForSelector('nav', { state: 'visible' });

    // Check the browser console for our debug logs
    const logs: string[] = [];
    page.on('console', msg => {
      if (msg.text().includes('ModernNavigation')) {
        logs.push(msg.text());
      }
    });

    // Click a button to trigger our logging
    const servicesButton = page.locator('nav button:has-text("Services")');
    await servicesButton.click();

    // Wait for logs to accumulate
    await page.waitForTimeout(2000);

    console.log('📝 Captured ModernNavigation logs:');
    logs.forEach((log, index) => {
      console.log(`  ${index + 1}. ${log}`);
    });

    if (logs.length === 0) {
      console.log(
        '❌ No ModernNavigation debug logs captured - component may not be mounted correctly'
      );
    } else {
      console.log(`✅ Captured ${logs.length} debug logs from ModernNavigation component`);
    }

    // Check for any React errors
    const reactErrors = await page.evaluate(() => {
      const errors: string[] = [];
      // Check if there are any React error boundaries triggered
      if ((window as any).__REACT_ERROR_OVERLAY_GLOBAL_HOOK__) {
        errors.push('React error overlay detected');
      }
      return errors;
    });

    if (reactErrors.length > 0) {
      console.log('❌ React errors detected:', reactErrors);
    } else {
      console.log('✅ No React errors detected');
    }
  });
});
