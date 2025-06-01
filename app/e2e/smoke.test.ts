import { test, expect } from '@playwright/test';
import { setupApiMocks } from './setup/api-mocks';

test.describe('Smoke Tests', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

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

		// MSW should intercept this and return our mock response (501 indicates interception)
		expect(response.status()).toBe(501);
		
		const responseData = await response.json();
		expect(responseData.intercepted).toBe(true);
		expect(responseData.message).toContain('MSW intercepted');
	});
});
