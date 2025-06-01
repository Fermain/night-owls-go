import { test, expect } from '@playwright/test';
import { AuthPage } from './page-objects/auth.page';
import { setupApiMocks } from './setup/api-mocks';

/**
 * CRITICAL USER JOURNEYS
 * Tests the core user flows that must work for the application to be functional
 */

test.describe('🔥 Critical User Journeys - Core Application Flows', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('🎯 User Registration Journey', async ({ page }) => {
		const authPage = new AuthPage(page);

		// Step 1: Navigate to homepage
		await authPage.goto();
		await expect(page.getByText('Mount Moreland Night Owls')).toBeVisible();

		// Step 2: Start registration process
		await authPage.joinUsButton.click();
		await expect(page).toHaveURL('/register');

		// Step 3: Fill registration form
		await authPage.nameInput.fill('Critical Test User');
		await authPage.phoneInput.fill('+27821234567');

		// Verify form state
		await expect(authPage.nameInput).toHaveValue('Critical Test User');
		await expect(authPage.phoneInput).toHaveValue('+27821234567');

		console.log('✅ Registration journey completed successfully');
	});

	test('🎯 Admin Access Journey', async ({ page }) => {
		// Step 1: Try accessing admin without auth
		await page.goto('/admin');
		await expect(page).toHaveURL('/login');

		// Step 2: Authenticate as admin
		const authPage = new AuthPage(page);
		await authPage.loginAsAdmin();

		// Step 3: Verify admin access
		await expect(page).toHaveURL('/admin');

		console.log('✅ Admin access journey completed successfully');
	});

	test('🎯 Navigation Flow', async ({ page }) => {
		// Test core navigation paths work
		await page.goto('/');
		await expect(page.getByText('Mount Moreland Night Owls')).toBeVisible();

		// Navigate to registration
		await page.getByRole('link', { name: /become an owl/i }).click();
		await expect(page).toHaveURL('/register');

		// Navigate to login
		await page.goto('/login');
		await expect(page.locator('body')).toBeVisible();

		console.log('✅ Navigation flow completed successfully');
	});

	test('🎯 Authentication State Persistence', async ({ page }) => {
		const authPage = new AuthPage(page);

		// Step 1: Start unauthenticated and verify initial state
		await page.goto('/');
		await expect(page.getByText('Mount Moreland Night Owls')).toBeVisible();
		
		// Step 2: Set auth state as volunteer
		await authPage.loginAsVolunteer();
		
		// Step 3: Verify auth state persists across navigation
		await page.goto('/register');
		await expect(page.locator('body')).toBeVisible();
		
		await page.goto('/login');
		await expect(page.locator('body')).toBeVisible();
		
		// Step 4: Verify auth state is actually set in localStorage
		const authState = await page.evaluate(() => {
			const userData = localStorage.getItem('user-session');
			return userData ? JSON.parse(userData) : null;
		});
		
		expect(authState).toBeTruthy();
		expect(authState.isAuthenticated).toBe(true);
		expect(authState.user.role).toBe('volunteer');

		// Step 5: Clear auth state and verify cleanup
		await authPage.logout();
		
		const clearedAuthState = await page.evaluate(() => {
			return localStorage.getItem('user-session');
		});
		
		expect(clearedAuthState).toBeNull();

		console.log('✅ Authentication state persistence working');
	});

	test('🎯 Error Recovery Journey', async ({ page }) => {
		// Test application recovery from errors
		
		// Navigate to non-existent page
		await page.goto('/this-page-does-not-exist');
		
		// Should handle gracefully (not crash)
		await expect(page.locator('body')).toBeVisible();
		
		// Recovery - navigate back to working page
		await page.goto('/');
		await expect(page.getByText('Mount Moreland Night Owls')).toBeVisible();

		console.log('✅ Error recovery journey completed successfully');
	});
});
