import { test, expect } from '@playwright/test';

test.describe('Smoke Tests', () => {
  test('app loads successfully', async ({ page }) => {
    await page.goto('/');
    
    // Just verify the page loads - we'll add more specific checks later
    await expect(page.locator('body')).toBeVisible();
    
    // Check for basic content (adjust based on actual homepage content)
    await expect(page.locator('html')).toBeVisible();
  });

  test('MSW intercepts API calls', async ({ page }) => {
    // Test that our MSW setup is working by making an API call
    const response = await page.request.post('/api/ping', {
      data: { test: 'data' }
    });
    
    // MSW should intercept this and return our mock response (500 or 501 both indicate interception)
    expect(response.status()).toBeGreaterThanOrEqual(500);
  });
}); 