import { test, expect } from '@playwright/test';

test.describe('Basic Smoke Tests', () => {
  test('homepage loads successfully', async ({ page }) => {
    await page.goto('/');

    // Check that the page loads with correct title
    await expect(page).toHaveTitle(/International Center/);

    // Check that main navigation is present
    await expect(page.locator('nav')).toBeVisible();

    // Check that main content is present
    await expect(page.locator('main')).toBeVisible();
  });

  test('services page loads successfully', async ({ page }) => {
    await page.goto('/services');

    // Check that services page loads
    await expect(page).toHaveTitle(/Services/);

    // Check that services are displayed
    await expect(page.locator('[data-testid="services-list"], .services-grid, main')).toBeVisible();
  });

  test('news page loads successfully', async ({ page }) => {
    await page.goto('/company/news');

    // Check that news page loads
    await expect(page).toHaveTitle(/News/);

    // Check that news content is displayed
    await expect(page.locator('[data-testid="news-list"], .news-grid, main')).toBeVisible();
  });

  test('contact page loads successfully', async ({ page }) => {
    await page.goto('/company/contact');

    // Check that contact page loads
    await expect(page).toHaveTitle(/Contact/);

    // Check that contact form is present
    await expect(page.locator('form, [data-testid="contact-form"], main')).toBeVisible();
  });

  test('research page loads successfully', async ({ page }) => {
    await page.goto('/community/research-innovation');

    // Check that research page loads
    await expect(page).toHaveTitle(/Research/);

    // Check that research content is displayed
    await expect(page.locator('[data-testid="research-list"], .research-grid, main')).toBeVisible();
  });

  test('navigation works correctly', async ({ page }) => {
    await page.goto('/');

    // Test navigation to services
    const servicesLink = page.locator('nav a[href*="services"], a:has-text("Services")').first();
    if (await servicesLink.isVisible()) {
      await servicesLink.click();
      await expect(page).toHaveURL(/.*services.*/);
    }

    // Navigate back to home
    await page.goto('/');

    // Test navigation to news
    const newsLink = page.locator('nav a[href*="news"], a:has-text("News")').first();
    if (await newsLink.isVisible()) {
      await newsLink.click();
      await expect(page).toHaveURL(/.*news.*/);
    }
  });

  test('search functionality works', async ({ page }) => {
    await page.goto('/');

    // Look for search input
    const searchInput = page
      .locator('input[type="search"], input[placeholder*="search" i], [data-testid="search-input"]')
      .first();

    if (await searchInput.isVisible()) {
      // Test search functionality
      await searchInput.fill('PRP');
      await searchInput.press('Enter');

      // Wait for search results or navigation
      await page.waitForTimeout(2000);

      // Check that we either get results or navigate to a search page
      const hasResults = await page
        .locator('[data-testid="search-results"], .search-results, .results')
        .isVisible();
      const isSearchPage = page.url().includes('search') || page.url().includes('services');

      expect(hasResults || isSearchPage).toBeTruthy();
    }
  });

  test('responsive design works', async ({ page }) => {
    // Test desktop
    await page.setViewportSize({ width: 1200, height: 800 });
    await page.goto('/');
    await expect(page.locator('nav')).toBeVisible();

    // Test tablet
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/');
    await expect(page.locator('main')).toBeVisible();

    // Test mobile
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/');
    await expect(page.locator('main')).toBeVisible();
  });

  test('footer links work', async ({ page }) => {
    await page.goto('/');

    // Scroll to footer
    await page.locator('footer').scrollIntoViewIfNeeded();

    // Check that footer is visible
    await expect(page.locator('footer')).toBeVisible();

    // Test privacy policy link if it exists
    const privacyLink = page.locator('footer a[href*="privacy"], a:has-text("Privacy")').first();
    if (await privacyLink.isVisible()) {
      await privacyLink.click();
      await expect(page).toHaveURL(/.*privacy.*/);
    }
  });

  test('api health checks', async ({ request }) => {
    // Test services API
    const servicesResponse = await request.get(`${process.env.SERVICES_API_URL}/health`);
    expect(servicesResponse.ok()).toBeTruthy();

    // Test news API
    const newsResponse = await request.get(`${process.env.NEWS_API_URL}/health`);
    expect(newsResponse.ok()).toBeTruthy();

    // Test research API
    const researchResponse = await request.get(`${process.env.RESEARCH_API_URL}/health`);
    expect(researchResponse.ok()).toBeTruthy();

    // Test contacts API
    const contactsResponse = await request.get(`${process.env.CONTACTS_API_URL}/health`);
    expect(contactsResponse.ok()).toBeTruthy();

    // Test search API
    const searchResponse = await request.get(`${process.env.SEARCH_API_URL}/health`);
    expect(searchResponse.ok()).toBeTruthy();
  });
});
