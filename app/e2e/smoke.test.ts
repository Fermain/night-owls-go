import { test, expect } from '@playwright/test';
import { setupApiMocks } from './setup/api-mocks';

test.describe('Smoke Tests', () => {
	test('app loads successfully', async ({ page }) => {
		await page.goto('/');

		// Just verify the page loads - we'll add more specific checks later
		await expect(page.locator('body')).toBeVisible();

		// Check for basic content (adjust based on actual homepage content)
		await expect(page.locator('html')).toBeVisible();
	});

	test('MSW intercepts API calls', async ({ page }) => {
		// Set up route interception directly in this test
		await page.route('**/api/ping', async (route) => {
			await route.fulfill({
				status: 501,
				contentType: 'application/json',
				body: JSON.stringify({
					message: 'MSW intercepted - ping endpoint mocked',
					intercepted: true
				})
			});
		});

		// Test that our route interception is working
		const response = await page.request.post('/api/ping', {
			data: { test: 'data' }
		});

		// Route should intercept this and return our mock response (501 indicates interception)
		expect(response.status()).toBe(501);
		
		const responseData = await response.json();
		expect(responseData.intercepted).toBe(true);
		expect(responseData.message).toContain('MSW intercepted');
	});

	test('API mocks work for e2e tests', async ({ page }) => {
		// Set up all API mocks for this test
		await setupApiMocks(page);
		
		// Navigate to a page that might make API calls
		await page.goto('/');
		
		// Test that emergency contacts API is mocked
		const emergencyResponse = await page.request.get('/api/emergency-contacts');
		expect(emergencyResponse.status()).toBe(200);
		
		const emergencyData = await emergencyResponse.json();
		expect(Array.isArray(emergencyData)).toBe(true);
		expect(emergencyData.length).toBeGreaterThan(0);
	});
});
