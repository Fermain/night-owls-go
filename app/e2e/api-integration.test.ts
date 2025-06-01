import { test, expect } from '@playwright/test';
import { setupApiMocks } from './setup/api-mocks';

/**
 * E2E User Journey Tests - Mocked APIs
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

test.describe('🚀 E2E User Journey Tests', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('✅ Complete User Registration Journey', async ({ page }) => {
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

		// Start user journey: Registration
		await page.goto('/register');
		
		// Check if the registration form elements exist before interaction
		const nameField = page.getByLabel('Full Name');
		const phoneField = page.getByLabel('Phone Number');
		
		await expect(nameField).toBeVisible({ timeout: 10000 });
		await expect(phoneField).toBeVisible({ timeout: 10000 });
		
		await nameField.fill('Test User');
		await phoneField.fill('+27821234567');
		
		const registerButton = page.getByRole('button', { name: /register|sign up/i });
		await expect(registerButton).toBeVisible();
		await registerButton.click();

		// Verify registration API call was made
		await page.waitForTimeout(1000); // Allow time for API call
		expect(lastRequest).toBeTruthy();
		expect(lastRequest!.url).toContain('/api/auth/register');
		expect(lastRequest!.method).toBe('POST');
		expect(lastRequest!.body?.name).toBe('Test User');
		expect(lastRequest!.body?.phone).toBe('+27821234567');

		// Continue journey: OTP verification
		const otpField = page.getByPlaceholder(/enter.*code|otp/i);
		await expect(otpField).toBeVisible({ timeout: 10000 });
		await otpField.fill('123456');
		
		const verifyButton = page.getByRole('button', { name: /verify|confirm/i });
		await expect(verifyButton).toBeVisible();
		await verifyButton.click();

		// Verify OTP API call was made
		await page.waitForTimeout(1000);
		expect(lastRequest).toBeTruthy();
		expect(lastRequest!.url).toContain('/api/auth/verify');

		console.log('✅ Complete user registration journey tested successfully');
	});

	test('✅ Shifts Browsing and Booking Journey', async ({ page }) => {
		const apiCalls: ApiCall[] = [];

		// Mock shifts endpoint with realistic data
		await page.route('**/shifts/**', async (route) => {
			const request = route.request();
			apiCalls.push({
				url: request.url(),
				method: request.method(),
				timestamp: Date.now()
			});

			if (request.url().includes('/available')) {
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
						},
						{
							schedule_id: 2,
							schedule_name: 'Evening Watch',
							start_time: '2024-12-25T18:00:00Z',
							end_time: '2024-12-25T22:00:00Z',
							timezone: 'Africa/Johannesburg',
							is_booked: false,
							positions_available: 1,
							positions_filled: 1
						}
					])
				});
			}
		});

		// User journey: Browse available shifts
		await page.goto('/shifts');

		// Wait for shifts to load
		await page.waitForTimeout(2000);

		// Verify API was called
		const shiftsApiCall = apiCalls.find((call) => call.url.includes('/available'));
		expect(shiftsApiCall).toBeTruthy();
		expect(shiftsApiCall?.method).toBe('GET');

		// Verify shifts are displayed in UI
		await expect(page.getByText('Morning Patrol')).toBeVisible({ timeout: 10000 });
		await expect(page.getByText('Evening Watch')).toBeVisible({ timeout: 10000 });

		console.log('✅ Shifts browsing journey tested successfully');
	});

	test('✅ Error Handling User Experience', async ({ page }) => {
		let failureCount = 0;

		// Simulate intermittent API failures
		await page.route('**/api/**', async (route) => {
			failureCount++;

			if (failureCount <= 2) {
				// Simulate network failures for first 2 requests
				await route.fulfill({
					status: 500,
					contentType: 'application/json',
					body: JSON.stringify({ error: 'Internal Server Error' })
				});
			} else {
				// Success after retries
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({ success: true, data: [] })
				});
			}
		});

		await page.goto('/shifts');

		// Should handle errors gracefully and show user-friendly messages
		const errorElement = page.getByText(/error|failed|try again/i);
		await expect(errorElement).toBeVisible({ timeout: 15000 });

		console.log('✅ Error handling user experience tested');
	});

	test('✅ Admin Dashboard User Journey', async ({ page }) => {
		const schedules: Schedule[] = [];
		let lastOperation: string = '';

		// Mock admin endpoints
		await page.route('**/api/admin/schedules**', async (route) => {
			const request = route.request();
			const method = request.method();

			if (method === 'GET') {
				lastOperation = 'READ';
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify(schedules)
				});
			} else if (method === 'POST') {
				const body = JSON.parse(request.postData() || '{}') as Record<string, unknown>;
				const newSchedule: Schedule = {
					id: Date.now(),
					name: body.name as string,
					description: body.description as string,
					created_at: new Date().toISOString(),
					...body
				};
				schedules.push(newSchedule);
				lastOperation = 'CREATE';

				await route.fulfill({
					status: 201,
					contentType: 'application/json',
					body: JSON.stringify(newSchedule)
				});
			}
		});

		// Mock admin authentication
		await page.route('**/api/auth/verify', async (route) => {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					success: true,
					token: 'admin-token',
					user: { id: 1, name: 'Admin User', role: 'admin' }
				})
			});
		});

		// Admin user journey: Login and manage schedules
		await page.goto('/login');
		
		// Simulate admin login
		const phoneField = page.getByLabel(/phone/i);
		if (await phoneField.isVisible()) {
			await phoneField.fill('+27821111111');
			const loginButton = page.getByRole('button', { name: /send|login/i });
			await loginButton.click();
			
			const otpField = page.getByPlaceholder(/code|otp/i);
			await otpField.fill('123456');
			const verifyButton = page.getByRole('button', { name: /verify/i });
			await verifyButton.click();
		}

		// Navigate to admin area
		await page.goto('/admin/schedules');
		await page.waitForTimeout(1000);

		// Verify schedules were loaded
		expect(lastOperation).toBe('READ');

		console.log('✅ Admin dashboard user journey tested');
	});
});
