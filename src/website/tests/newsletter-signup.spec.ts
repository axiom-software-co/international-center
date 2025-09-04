/**
 * Newsletter Signup E2E Tests
 * End-to-end testing of newsletter subscription flow across the website
 */

import { test, expect, Page } from '@playwright/test';

// Test configuration
const NEWSLETTER_API_URL = 'http://localhost:8086';
const TEST_EMAIL_DOMAIN = '@e2e-test.com';

// Utility to generate unique test emails
const generateTestEmail = (prefix: string = 'newsletter-e2e'): string => {
  const timestamp = Date.now();
  const random = Math.random().toString(36).substring(7);
  return `${prefix}-${timestamp}-${random}${TEST_EMAIL_DOMAIN}`;
};

// Check if newsletter service is running
const checkNewsletterService = async (): Promise<boolean> => {
  try {
    const response = await fetch(`${NEWSLETTER_API_URL}/api/newsletter/confirm/test`);
    return response.status === 400; // Expected response for invalid token
  } catch {
    return false;
  }
};

// Cleanup function to unsubscribe test emails
const cleanupTestEmail = async (email: string) => {
  try {
    await fetch(`${NEWSLETTER_API_URL}/api/newsletter/unsubscribe`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email,
        reason: 'E2E test cleanup',
      }),
    });
  } catch (error) {
    console.warn(`Cleanup failed for ${email}:`, error);
  }
};

// Helper function to wait for and interact with newsletter CTA
const interactWithNewsletterCTA = async (page: Page, email: string) => {
  // Find the newsletter CTA form
  const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i }).first();
  
  // Wait for the section to be visible
  await expect(ctaSection).toBeVisible();
  
  // Find email input and submit button within the CTA
  const emailInput = ctaSection.locator('input[type="email"]');
  const submitButton = ctaSection.locator('button[type="submit"]');
  
  // Verify elements exist
  await expect(emailInput).toBeVisible();
  await expect(submitButton).toBeVisible();
  
  // Fill and submit the form
  await emailInput.fill(email);
  await submitButton.click();
  
  return { ctaSection, emailInput, submitButton };
};

test.describe('Newsletter Signup E2E Tests', () => {
  test.beforeAll(async () => {
    // Verify newsletter service is running
    const isServiceRunning = await checkNewsletterService();
    if (!isServiceRunning) {
      throw new Error('Newsletter service is not running on localhost:8086');
    }
    console.log('✅ Newsletter service is available');
  });

  test.describe('Newsletter CTA on News Hub Page', () => {
    test('should successfully subscribe from news hub page', async ({ page }) => {
      const testEmail = generateTestEmail('news-hub');
      
      console.log(`🧪 Testing newsletter signup from news hub with email: ${testEmail}`);
      
      try {
        // Navigate to news hub
        await page.goto('/company/news');
        
        // Wait for page to load
        await expect(page.locator('h1')).toContainText(/news/i);
        
        // Scroll to newsletter CTA (usually at bottom)
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        // Interact with newsletter CTA
        const { ctaSection } = await interactWithNewsletterCTA(page, testEmail);
        
        // Wait for loading state
        await expect(ctaSection.locator('button')).toContainText(/subscribing/i);
        
        // Wait for success message
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout: 15000 });
        await expect(ctaSection.locator('button')).toContainText(/subscribed/i);
        
        console.log('✅ Newsletter signup successful from news hub');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
    
    test('should show validation error for invalid email on news hub', async ({ page }) => {
      console.log('🧪 Testing email validation on news hub page');
      
      await page.goto('/company/news');
      await expect(page.locator('h1')).toContainText(/news/i);
      
      // Scroll to newsletter CTA
      await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      
      // Try to submit invalid email
      await interactWithNewsletterCTA(page, 'invalid-email');
      
      // Check for validation error
      const errorMessage = page.locator('text=/valid email|email format|invalid/i');
      await expect(errorMessage).toBeVisible({ timeout: 5000 });
      
      console.log('✅ Email validation working on news hub');
    });
  });

  test.describe('Newsletter CTA on Research Hub Page', () => {
    test('should successfully subscribe from research hub page', async ({ page }) => {
      const testEmail = generateTestEmail('research-hub');
      
      console.log(`🧪 Testing newsletter signup from research hub with email: ${testEmail}`);
      
      try {
        // Navigate to research hub
        await page.goto('/community/research');
        
        // Wait for page to load
        await expect(page.locator('h1')).toContainText(/research/i);
        
        // Scroll to newsletter CTA
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        // Interact with newsletter CTA
        const { ctaSection } = await interactWithNewsletterCTA(page, testEmail);
        
        // Wait for success
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout: 15000 });
        
        console.log('✅ Newsletter signup successful from research hub');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
  });

  test.describe('Newsletter CTA on Dynamic Article Pages', () => {
    test('should successfully subscribe from a news article page', async ({ page }) => {
      const testEmail = generateTestEmail('news-article');
      
      console.log(`🧪 Testing newsletter signup from news article with email: ${testEmail}`);
      
      try {
        // First go to news hub to find an article
        await page.goto('/company/news');
        await expect(page.locator('h1')).toContainText(/news/i);
        
        // Wait for articles to load and click on first article
        const firstArticleLink = page.locator('a[href*="/company/news/"]').first();
        await expect(firstArticleLink).toBeVisible({ timeout: 10000 });
        await firstArticleLink.click();
        
        // Verify we're on article page
        await expect(page.locator('article, .article-content, [data-article]')).toBeVisible();
        
        // Scroll to newsletter CTA (should be above footer)
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        // Interact with newsletter CTA
        const { ctaSection } = await interactWithNewsletterCTA(page, testEmail);
        
        // Wait for success
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout: 15000 });
        
        console.log('✅ Newsletter signup successful from news article');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
    
    test('should successfully subscribe from a research article page', async ({ page }) => {
      const testEmail = generateTestEmail('research-article');
      
      console.log(`🧪 Testing newsletter signup from research article with email: ${testEmail}`);
      
      try {
        // First go to research hub to find an article
        await page.goto('/community/research');
        await expect(page.locator('h1')).toContainText(/research/i);
        
        // Wait for articles to load and click on first article
        const firstArticleLink = page.locator('a[href*="/community/research/"]').first();
        await expect(firstArticleLink).toBeVisible({ timeout: 10000 });
        await firstArticleLink.click();
        
        // Verify we're on research article page
        await expect(page.locator('article, .article-content, [data-article]')).toBeVisible();
        
        // Scroll to newsletter CTA
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        // Interact with newsletter CTA
        const { ctaSection } = await interactWithNewsletterCTA(page, testEmail);
        
        // Wait for success
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout: 15000 });
        
        console.log('✅ Newsletter signup successful from research article');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
  });

  test.describe('Mobile Newsletter Signup', () => {
    test('should work correctly on mobile devices', async ({ page, browserName }) => {
      // Skip webkit on mobile due to potential issues
      if (browserName === 'webkit') {
        test.skip();
      }
      
      const testEmail = generateTestEmail('mobile');
      
      console.log(`🧪 Testing mobile newsletter signup with email: ${testEmail}`);
      
      try {
        // Set mobile viewport
        await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE size
        
        await page.goto('/company/news');
        await expect(page.locator('h1')).toContainText(/news/i);
        
        // Scroll to newsletter CTA
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        // On mobile, the form might be stacked vertically
        const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i }).first();
        await expect(ctaSection).toBeVisible();
        
        const emailInput = ctaSection.locator('input[type="email"]');
        const submitButton = ctaSection.locator('button[type="submit"]');
        
        // Verify mobile responsiveness
        await expect(emailInput).toBeVisible();
        await expect(submitButton).toBeVisible();
        
        // Check that form is properly sized for mobile
        const emailInputBox = await emailInput.boundingBox();
        const submitButtonBox = await submitButton.boundingBox();
        
        if (emailInputBox && submitButtonBox) {
          expect(emailInputBox.width).toBeGreaterThan(200); // Reasonable width on mobile
          expect(submitButtonBox.width).toBeGreaterThan(80); // Button should be tappable
        }
        
        // Submit form
        await emailInput.fill(testEmail);
        await submitButton.click();
        
        // Wait for success
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout: 15000 });
        
        console.log('✅ Mobile newsletter signup successful');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
  });

  test.describe('Cross-Browser Compatibility', () => {
    test('should work consistently across browsers', async ({ page, browserName }) => {
      const testEmail = generateTestEmail(`browser-${browserName}`);
      
      console.log(`🧪 Testing newsletter signup in ${browserName} with email: ${testEmail}`);
      
      try {
        await page.goto('/company/news');
        await expect(page.locator('h1')).toContainText(/news/i);
        
        // Scroll to newsletter CTA
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        // Test form interaction
        const { ctaSection } = await interactWithNewsletterCTA(page, testEmail);
        
        // Wait for success (timeout might vary by browser)
        const timeout = browserName === 'webkit' ? 20000 : 15000;
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout });
        
        console.log(`✅ Newsletter signup successful in ${browserName}`);
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
  });

  test.describe('Performance and User Experience', () => {
    test('should handle newsletter signup without blocking UI', async ({ page }) => {
      const testEmail = generateTestEmail('performance');
      
      console.log(`🧪 Testing newsletter signup performance with email: ${testEmail}`);
      
      try {
        await page.goto('/company/news');
        await expect(page.locator('h1')).toContainText(/news/i);
        
        // Scroll to newsletter CTA
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        const { ctaSection, emailInput, submitButton } = await interactWithNewsletterCTA(page, testEmail);
        
        // Verify UI responds immediately to click
        await expect(submitButton).toHaveAttribute('disabled', '');
        await expect(submitButton).toContainText(/subscribing/i);
        
        // Verify email input is disabled during submission
        await expect(emailInput).toHaveAttribute('disabled', '');
        
        // Verify success state
        await expect(ctaSection.locator('text=/subscribed|success/i')).toBeVisible({ timeout: 15000 });
        await expect(submitButton).toContainText(/subscribed/i);
        
        // Verify form is reset to success state properly
        await expect(emailInput).toHaveAttribute('disabled', '');
        await expect(submitButton).toHaveAttribute('disabled', '');
        
        console.log('✅ Newsletter signup performance test passed');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });

    test('should show clear feedback for duplicate subscriptions', async ({ page }) => {
      const testEmail = generateTestEmail('duplicate-e2e');
      
      console.log(`🧪 Testing duplicate subscription handling with email: ${testEmail}`);
      
      try {
        // First, subscribe via API
        const subscribeResponse = await fetch(`${NEWSLETTER_API_URL}/api/newsletter/subscribe`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            email: testEmail,
            source: 'website_newsletter_signup',
            contentType: 'all',
          }),
        });
        
        expect(subscribeResponse.ok).toBe(true);
        console.log('📧 First subscription created via API');
        
        // Now try to subscribe via UI
        await page.goto('/company/news');
        await expect(page.locator('h1')).toContainText(/news/i);
        
        await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
        
        const { ctaSection } = await interactWithNewsletterCTA(page, testEmail);
        
        // Should still show success but mention already subscribed
        await expect(ctaSection.locator('text=/subscribed|already subscribed|success/i')).toBeVisible({ timeout: 15000 });
        
        console.log('✅ Duplicate subscription handled gracefully');
        
      } finally {
        // Cleanup
        await cleanupTestEmail(testEmail);
      }
    });
  });

  test.describe('Accessibility', () => {
    test('should be accessible via keyboard navigation', async ({ page }) => {
      console.log('🧪 Testing newsletter CTA keyboard accessibility');
      
      await page.goto('/company/news');
      await expect(page.locator('h1')).toContainText(/news/i);
      
      // Scroll to newsletter CTA
      await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      
      // Tab to the email input
      const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i }).first();
      const emailInput = ctaSection.locator('input[type="email"]');
      
      // Focus should work
      await emailInput.focus();
      await expect(emailInput).toBeFocused();
      
      // Should be able to type
      const testEmail = generateTestEmail('accessibility');
      await page.keyboard.type(testEmail);
      await expect(emailInput).toHaveValue(testEmail);
      
      // Tab to submit button and activate with keyboard
      await page.keyboard.press('Tab');
      const submitButton = ctaSection.locator('button[type="submit"]');
      await expect(submitButton).toBeFocused();
      
      // Submit with Enter key
      await page.keyboard.press('Enter');
      
      // Should show loading state
      await expect(submitButton).toContainText(/subscribing/i);
      
      console.log('✅ Newsletter CTA keyboard accessibility test passed');
      
      // Cleanup
      await cleanupTestEmail(testEmail);
    });

    test('should have proper ARIA attributes', async ({ page }) => {
      console.log('🧪 Testing newsletter CTA ARIA attributes');
      
      await page.goto('/company/news');
      await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
      
      const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i }).first();
      const emailInput = ctaSection.locator('input[type="email"]');
      const submitButton = ctaSection.locator('button[type="submit"]');
      
      // Check for proper ARIA attributes
      await expect(emailInput).toHaveAttribute('autocomplete', 'email');
      await expect(emailInput).toHaveAttribute('type', 'email');
      
      // Button should have proper type and not be link
      await expect(submitButton).toHaveAttribute('type', 'submit');
      
      console.log('✅ Newsletter CTA ARIA attributes test passed');
    });
  });
});