/**
 * Newsletter Health Check Test
 * Verify that the newsletter service and website are ready for E2E testing
 */

import { test, expect } from '@playwright/test';

const NEWSLETTER_API_URL = 'http://localhost:8086';
const WEBSITE_URL = 'http://localhost:4321';

test.describe('Newsletter E2E Health Check', () => {
  test('should verify newsletter service is running', async () => {
    const response = await fetch(`${NEWSLETTER_API_URL}/api/newsletter/confirm/test`);
    
    // Should return 400 for invalid token (as we fixed)
    expect(response.status).toBe(400);
    
    const data = await response.json();
    expect(data).toHaveProperty('success', false);
    expect(data).toHaveProperty('error', 'INVALID_TOKEN');
    
    console.log('✅ Newsletter service health check passed');
  });

  test('should verify website is accessible', async () => {
    const response = await fetch(WEBSITE_URL);
    
    expect(response.status).toBe(200);
    expect(response.headers.get('content-type')).toContain('text/html');
    
    console.log('✅ Website accessibility check passed');
  });

  test('should verify newsletter subscription API endpoint', async () => {
    const timestamp = Date.now();
    const testData = {
      email: `health-check-${timestamp}@e2e-test.com`,
      source: 'e2e_health_check',
      contentType: 'all'
    };

    const response = await fetch(`${NEWSLETTER_API_URL}/api/newsletter/subscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(testData),
    });

    // Should create subscription successfully (201) or already exists (200)
    expect([200, 201]).toContain(response.status);
    
    const data = await response.json();
    expect(data).toHaveProperty('success', true);
    expect(data).toHaveProperty('subscription_id');
    
    console.log('✅ Newsletter subscription API check passed');

    // Cleanup - unsubscribe the test email
    await fetch(`${NEWSLETTER_API_URL}/api/newsletter/unsubscribe`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        email: testData.email,
        reason: 'E2E health check cleanup',
      }),
    });
    
    console.log('✅ Test email cleaned up');
  });
});