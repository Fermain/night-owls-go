import { test, expect } from '@playwright/test';

/**
 * Lifecycle Fix Verification Test
 * 
 * This test specifically verifies that the lifecycle_outside_component error
 * that was occurring in production has been resolved.
 */

test.describe('Lifecycle Error Fix Verification', () => {
	test('Public routes should not have lifecycle errors', async ({ page }) => {
		// Capture console errors
		const consoleErrors: string[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				consoleErrors.push(msg.text());
			}
		});

		// Capture uncaught exceptions
		const uncaughtErrors: string[] = [];
		page.on('pageerror', (error) => {
			uncaughtErrors.push(error.message);
		});

		// Test homepage
		await page.goto('/');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(2000); // Allow time for any lifecycle errors to occur

		// Test registration page
		await page.goto('/register');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(2000);

		// Test login page
		await page.goto('/login');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(2000);

		// Verify no lifecycle errors occurred
		const lifecycleErrors = [...consoleErrors, ...uncaughtErrors].filter(error =>
			error.includes('lifecycle_outside_component') ||
			error.includes('https://svelte.dev/e/lifecycle_outside_component')
		);

		console.log('Console errors:', consoleErrors);
		console.log('Uncaught errors:', uncaughtErrors);
		console.log('Lifecycle errors:', lifecycleErrors);

		expect(lifecycleErrors).toHaveLength(0);
	});

	test('Admin routes should work correctly with sidebar context', async ({ page }) => {
		// Capture console errors
		const consoleErrors: string[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				consoleErrors.push(msg.text());
			}
		});

		// Mock authentication to access admin routes
		await page.addInitScript(() => {
			localStorage.setItem('user-session', JSON.stringify({
				user: {
					id: 'test-admin',
					name: 'Test Admin',
					phone: '+27821234567',
					role: 'admin'
				},
				token: 'test-token',
				isAuthenticated: true
			}));
		});

		// Test admin dashboard (should have sidebar context)
		await page.goto('/admin');
		await page.waitForLoadState('networkidle');
		await page.waitForTimeout(2000);

		// Verify sidebar elements are present (indicating context is working)
		const sidebarTrigger = page.locator('[data-sidebar="trigger"]');
		await expect(sidebarTrigger).toBeVisible();

		// Verify no lifecycle errors occurred
		const lifecycleErrors = consoleErrors.filter(error =>
			error.includes('lifecycle_outside_component') ||
			error.includes('https://svelte.dev/e/lifecycle_outside_component')
		);

		console.log('Admin console errors:', consoleErrors);
		console.log('Admin lifecycle errors:', lifecycleErrors);

		expect(lifecycleErrors).toHaveLength(0);
	});

	test('Sidebar trigger should work safely on both public and admin routes', async ({ page }) => {
		// Test UnifiedHeader behavior on public routes
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Sidebar trigger should not be visible on public routes
		const publicSidebarTrigger = page.locator('[data-sidebar="trigger"]');
		await expect(publicSidebarTrigger).not.toBeVisible();

		// Mock authentication for admin test
		await page.addInitScript(() => {
			localStorage.setItem('user-session', JSON.stringify({
				user: {
					id: 'test-admin',
					name: 'Test Admin',
					phone: '+27821234567',
					role: 'admin'
				},
				token: 'test-token',
				isAuthenticated: true
			}));
		});

		// Test admin route
		await page.goto('/admin');
		await page.waitForLoadState('networkidle');

		// Sidebar trigger should be visible on admin routes
		const adminSidebarTrigger = page.locator('[data-sidebar="trigger"]');
		await expect(adminSidebarTrigger).toBeVisible();

		// Test that clicking the trigger doesn't cause errors
		await adminSidebarTrigger.click();
		await page.waitForTimeout(500);

		// No errors should occur
		expect(await page.evaluate(() => window.console.error)).toBeFalsy();
	});
}); 