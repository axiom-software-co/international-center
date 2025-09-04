/**
 * Newsletter Navigation E2E Tests
 * Verify that pages with newsletter CTAs are accessible and rendered correctly
 */

import { test, expect } from '@playwright/test';

test.describe('Newsletter CTA Page Navigation', () => {
  test('should navigate to news hub and find newsletter CTA', async ({ page }) => {
    console.log('🧪 Testing news hub page navigation and newsletter CTA presence');
    
    await page.goto('/company/news');
    
    // Wait for page to load
    await expect(page.locator('h1')).toContainText(/news/i);
    
    // Check for newsletter CTA presence
    const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i });
    await expect(ctaSection).toBeVisible();
    
    // Check for email input
    const emailInput = ctaSection.locator('input[type="email"]');
    await expect(emailInput).toBeVisible();
    
    // Check for submit button  
    const submitButton = ctaSection.locator('button[type="submit"]');
    await expect(submitButton).toBeVisible();
    
    console.log('✅ News hub newsletter CTA found and accessible');
  });

  test('should navigate to research hub and find newsletter CTA', async ({ page }) => {
    console.log('🧪 Testing research hub page navigation and newsletter CTA presence');
    
    await page.goto('/community/research');
    
    // Wait for page to load
    await expect(page.locator('h1')).toContainText(/research/i);
    
    // Check for newsletter CTA presence
    const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i });
    await expect(ctaSection).toBeVisible();
    
    // Check for email input
    const emailInput = ctaSection.locator('input[type="email"]');
    await expect(emailInput).toBeVisible();
    
    // Check for submit button
    const submitButton = ctaSection.locator('button[type="submit"]');
    await expect(submitButton).toBeVisible();
    
    console.log('✅ Research hub newsletter CTA found and accessible');
  });

  test('should navigate to a news article and find newsletter CTA', async ({ page }) => {
    console.log('🧪 Testing news article navigation and newsletter CTA presence');
    
    // First go to news hub
    await page.goto('/company/news');
    await expect(page.locator('h1')).toContainText(/news/i);
    
    // Find and click on first article link
    const firstArticleLink = page.locator('a[href*="/company/news/"]').first();
    
    if (await firstArticleLink.isVisible()) {
      await firstArticleLink.click();
      
      // Verify we're on article page
      await expect(page.locator('article, .article-content, [data-article]')).toBeVisible();
      
      // Check for newsletter CTA presence
      const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i });
      await expect(ctaSection).toBeVisible();
      
      console.log('✅ News article newsletter CTA found and accessible');
    } else {
      console.log('⚠️ No news articles found to test');
    }
  });

  test('should verify newsletter CTA form elements are properly configured', async ({ page }) => {
    console.log('🧪 Testing newsletter CTA form accessibility and configuration');
    
    await page.goto('/company/news');
    
    const ctaSection = page.locator('section').filter({ hasText: /stay|newsletter|subscribe|updates/i }).first();
    await expect(ctaSection).toBeVisible();
    
    const emailInput = ctaSection.locator('input[type="email"]');
    const submitButton = ctaSection.locator('button[type="submit"]');
    
    // Check for proper email input attributes
    await expect(emailInput).toHaveAttribute('type', 'email');
    await expect(emailInput).toHaveAttribute('autocomplete', 'email');
    
    // Check for proper button attributes
    await expect(submitButton).toHaveAttribute('type', 'submit');
    
    // Check that form is within a proper form element
    const formElement = ctaSection.locator('form');
    await expect(formElement).toBeVisible();
    
    console.log('✅ Newsletter CTA form is properly configured for accessibility');
  });
});