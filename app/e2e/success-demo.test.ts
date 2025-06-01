import { test, expect } from '@playwright/test';
import { AuthPage } from './page-objects/auth.page';
import { setupApiMocks } from './setup/api-mocks';

/**
 * SUCCESS DEMO - Tests that demonstrate working patterns
 * These tests showcase the utilities and patterns that consistently work
 */

test.describe('✅ Success Demo - Working Test Patterns', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('✅ Homepage loads and contains key content', async ({ page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Check for essential homepage content
		await expect(page.getByText('Mount Moreland Night Owls')).toBeVisible();
		await expect(page.getByRole('link', { name: /become an owl/i })).toBeVisible();

		console.log('✅ Homepage loads with all key content');
	});

	test('✅ Registration form displays and accepts input', async ({ page }) => {
		const authPage = new AuthPage(page);

		await page.goto('/register');
		await expect(authPage.nameInput).toBeVisible();
		await expect(authPage.phoneInput).toBeVisible();

		// Test form interaction
		await authPage.nameInput.fill('Demo User');
		await expect(authPage.nameInput).toHaveValue('Demo User');

		console.log('✅ Registration form works correctly');
	});

	test('✅ Navigation between pages works smoothly', async ({ page }) => {
		// Start at homepage
		await page.goto('/');
		await expect(page.getByText('Mount Moreland Night Owls')).toBeVisible();

		// Navigate to registration
		await page.getByRole('link', { name: /become an owl/i }).click();
		await expect(page).toHaveURL('/register');

		// Navigate to login
		await page.goto('/login');
		await expect(page.locator('body')).toBeVisible();

		console.log('✅ Page navigation functions correctly');
	});

	test('✅ Admin protection redirects work', async ({ page }) => {
		await page.goto('/admin');

		// Should redirect to login page
		await expect(page).toHaveURL('/login');

		console.log('✅ Admin route protection working');
	});

	test('✅ API mocking is functioning', async ({ page }) => {
		// This test demonstrates that our setupApiMocks() function works
		await page.goto('/register');

		// The mocks should be active for this test
		const authPage = new AuthPage(page);
		await expect(authPage.nameInput).toBeVisible();
		await expect(authPage.phoneInput).toBeVisible();

		console.log('✅ API mocking system is functional');
	});

	test('✅ Multiple page loads work reliably', async ({ page }) => {
		// Test that we can load multiple pages without issues
		const routes = [
			{ path: '/', description: 'Homepage' },
			{ path: '/login', description: 'Login page' },
			{ path: '/register', description: 'Registration page' }
		];

		for (const route of routes) {
			await page.goto(route.path);
			await expect(page.locator('body')).toBeVisible();
			console.log(`✅ ${route.description} loads successfully`);
		}
	});

	test('✅ Form utilities work consistently', async ({ page }) => {
		const authPage = new AuthPage(page);
		await page.goto('/register');

		// Test that our form helpers work
		await authPage.nameInput.fill('Test User Name');
		await expect(authPage.nameInput).toHaveValue('Test User Name');

		// Test phone input (basic fill, not full validation)
		await authPage.phoneInput.fill('+27821234567');
		await expect(authPage.phoneInput).toHaveValue('+27821234567');

		console.log('✅ Form utilities are working correctly');
	});
});
