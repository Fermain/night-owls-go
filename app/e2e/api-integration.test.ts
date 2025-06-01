import { test, expect } from '@playwright/test';
import { setupApiMocks } from './setup/api-mocks';
import { fillRegistrationForm, waitForSubmitButton } from './utils/form-helpers';
import { loginAsAdmin, setupAuthMocks } from './utils/auth-helpers';

/**
 * E2E User Journey Tests - Updated for Current Architecture
 * 
 * These tests focus on complete user journeys through the application
 * Purpose: Test user flows, UI interactions, and frontend logic
 * Strategy: Mock all external APIs for predictable, fast testing
 */

interface ApiRequest {
	url: string;
	method: string;
	headers: Record<string, string>;
	body: Record<string, unknown> | null;
}

interface ApiCall {
	url: string;
	method: string;
	timestamp: number;
}

interface Schedule {
	id: number;
	name: string;
	description: string;
	created_at: string;
	[key: string]: unknown;
}

test.describe('🚀 E2E User Journey Tests - Current Architecture', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('✅ Modern Registration Journey - "Become an Owl"', async ({ page }) => {
		let lastRequest: ApiRequest | null = null;

		// Monitor API calls for verification
		await page.route('**/api/auth/**', async (route) => {
			const request = route.request();
			lastRequest = {
				url: request.url(),
				method: request.method(),
				headers: request.headers(),
				body: request.postData() ? JSON.parse(request.postData()!) : null
			};

			if (request.url().includes('/register')) {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({
						success: true,
						message: 'Registration successful!',
						user: { id: 123, name: 'Test User', phone: '+27821234567', role: 'guest' }
					})
				});
			} else if (request.url().includes('/verify')) {
				const body = JSON.parse(request.postData() || '{}') as Record<string, unknown>;
				const code = (body.code || body.otp || '') as string;

				if (code === '123456') {
					await route.fulfill({
						status: 200,
						contentType: 'application/json',
						body: JSON.stringify({
							success: true,
							message: 'Login successful!',
							token: 'mock-jwt-token-12345',
							user: { id: 123, name: 'Test User', phone: '+27821234567', role: 'guest' }
						})
					});
				} else {
					await route.fulfill({
						status: 400,
						contentType: 'application/json',
						body: JSON.stringify({ error: 'Invalid OTP' })
					});
				}
			}
		});

		// Start user journey: Visit home page and click "Become an Owl"
		await page.goto('/');
		
		// Modern button text: "Become an Owl" (it's actually a link)
		const becomeOwlLink = page.getByRole('link', { name: /become an owl/i });
		await expect(becomeOwlLink).toBeVisible({ timeout: 10000 });
		await becomeOwlLink.click();

		// Should navigate to registration page
		await expect(page).toHaveURL('/register');

		// Fill registration form using helper utilities
		await fillRegistrationForm(page, 'Test User', '0821234567');
		
		// Wait for the Create account button to become enabled and click it
		const createAccountButton = await waitForSubmitButton(page, /create account/i, 5000);
		await createAccountButton.click();

		// Verify registration API call was made
		await page.waitForTimeout(1000);
		expect(lastRequest).toBeTruthy();
		expect(lastRequest!.url).toContain('/api/auth/register');
		expect(lastRequest!.method).toBe('POST');
		expect(lastRequest!.body?.name).toBe('Test User');

		console.log('✅ Modern registration journey tested successfully');
	});

	test('✅ Home Page Shift Browsing - No Deprecated /shifts Route', async ({ page }) => {
		const apiCalls: ApiCall[] = [];

		// Mock home page shift data (updated endpoint)
		await page.route('**/shifts/available', async (route) => {
			const request = route.request();
			apiCalls.push({
				url: request.url(),
				method: request.method(),
				timestamp: Date.now()
			});

			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify([
					{
						schedule_id: 1,
						schedule_name: 'Morning Patrol',
						start_time: '2024-12-25T08:00:00Z',
						end_time: '2024-12-25T12:00:00Z',
						timezone: 'Africa/Johannesburg',
						is_booked: false,
						positions_available: 2,
						positions_filled: 0
					}
				])
			});
		});

		// User journey: Use home page for shift functionality (not /shifts)
		await page.goto('/');

		// Wait for shifts to load on home page
		await page.waitForTimeout(2000);

		// Verify shifts are displayed on home page
		await expect(page.getByText('Morning Patrol')).toBeVisible({ timeout: 10000 });

		console.log('✅ Home page shift browsing tested successfully');
	});

	test('✅ Broadcasts Page Journey', async ({ page }) => {
		// User journey: Access broadcasts from modern navigation
		await page.goto('/broadcasts');

		// Wait for broadcasts to load
		await page.waitForTimeout(1000);

		// Check for broadcast content (should be handled by our mocks)
		const broadcastsSection = page.locator('[data-testid="broadcasts"], .broadcasts, main');
		await expect(broadcastsSection).toBeVisible({ timeout: 10000 });

		console.log('✅ Broadcasts page journey tested successfully');
	});

	test('✅ Admin Dashboard - Modern Layout', async ({ page }) => {
		// Set up authentication mocks
		await setupAuthMocks(page);
		
		// Login as admin user (sets auth state and navigates)
		await loginAsAdmin(page);

		// Should load admin dashboard (not redirect to login)
		await expect(page).toHaveURL('/admin');

		console.log('✅ Modern admin dashboard tested successfully');
	});

	test('✅ Mobile Navigation - "Join Community" Text', async ({ page }) => {
		// Set mobile viewport
		await page.setViewportSize({ width: 375, height: 667 });
		
		await page.goto('/');

		// Look for mobile navigation menu
		const mobileMenuButton = page.getByRole('button', { name: /menu/i });
		if (await mobileMenuButton.isVisible()) {
			await mobileMenuButton.click();
			
			// Check for mobile-specific text: "Join Community"
			await expect(page.getByText(/join community/i)).toBeVisible({ timeout: 5000 });
		}

		console.log('✅ Mobile navigation text tested successfully');
	});

	test('✅ Error Handling - Network Resilience', async ({ page }) => {
		let failureCount = 0;

		// Simulate API failures
		await page.route('**/api/**', async (route) => {
			failureCount++;

			if (failureCount <= 2) {
				await route.fulfill({
					status: 500,
					contentType: 'application/json',
					body: JSON.stringify({ error: 'Internal Server Error' })
				});
			} else {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({ success: true, data: [] })
				});
			}
		});

		await page.goto('/');

		// Should handle errors gracefully
		const errorElement = page.getByText(/error|failed|try again|something went wrong/i);
		// Allow longer timeout for error handling
		await expect(errorElement).toBeVisible({ timeout: 15000 });

		console.log('✅ Error handling tested successfully');
	});
});
