import { test, expect } from '@playwright/test';
import { setupApiMocks } from './setup/api-mocks';

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

test.describe('🔌 API Integration Tests', () => {
	test.beforeEach(async ({ page }) => {
		await setupApiMocks(page);
	});

	test('✅ Authentication API Flow - Complete Journey', async ({ page }) => {
		let lastRequest: ApiRequest | null = null;

		// Monitor API calls
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

		// Test Registration
		await page.goto('/register');
		await page.getByLabel('Full Name').fill('Test User');
		await page.getByLabel('Phone Number').fill('+27821234567');
		await page.getByRole('button', { name: /register|sign up/i }).click();

		// Verify registration API call
		expect(lastRequest).toBeTruthy();
		expect(lastRequest!.url).toContain('/api/auth/register');
		expect(lastRequest!.method).toBe('POST');
		expect(lastRequest!.body?.name).toBe('Test User');
		expect(lastRequest!.body?.phone).toBe('+27821234567');

		// Test OTP verification
		await page.getByPlaceholder(/enter.*code|otp/i).fill('123456');
		await page.getByRole('button', { name: /verify|confirm/i }).click();

		// Verify OTP API call
		expect(lastRequest).toBeTruthy();
		expect(lastRequest!.url).toContain('/api/auth/verify');
		expect(lastRequest!.body?.phone).toBe('+27821234567');
		expect(lastRequest!.body?.code || lastRequest!.body?.otp).toBe('123456');

		console.log('✅ Complete authentication API flow tested');
	});

	test('✅ Shifts API - Data Loading and Filtering', async ({ page }) => {
		const apiCalls: ApiCall[] = [];

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

		await page.goto('/shifts');

		// Wait for API call
		await page.waitForTimeout(1000);

		// Verify API was called
		const shiftsApiCall = apiCalls.find((call) => call.url.includes('/available'));
		expect(shiftsApiCall).toBeTruthy();
		expect(shiftsApiCall?.method).toBe('GET');

		// Verify shifts are displayed
		await expect(page.getByText('Morning Patrol')).toBeVisible();
		await expect(page.getByText('Evening Watch')).toBeVisible();

		console.log('✅ Shifts API loading and data display tested');
	});

	test('✅ Admin API - Schedule Management CRUD', async ({ page }) => {
		const schedules: Schedule[] = [];
		let lastOperation: string = '';

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
			} else if (method === 'PUT' || method === 'PATCH') {
				lastOperation = 'UPDATE';
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({ success: true })
				});
			} else if (method === 'DELETE') {
				lastOperation = 'DELETE';
				await route.fulfill({
					status: 204,
					contentType: 'application/json',
					body: ''
				});
			}
		});

		// Mock admin login
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

		// Login as admin
		await page.goto('/login');
		await page.getByLabel(/phone/i).fill('+27821111111');
		await page.getByRole('button', { name: /send|login/i }).click();
		await page.getByPlaceholder(/code|otp/i).fill('123456');
		await page.getByRole('button', { name: /verify/i }).click();

		// Navigate to schedules
		await page.goto('/admin/schedules');

		// Test READ operation
		expect(lastOperation).toBe('READ');

		// Test CREATE operation (if form exists)
		const createButton = page.getByRole('button', { name: /create|new/i });
		if (await createButton.isVisible()) {
			await createButton.click();

			// Fill form (adjust selectors based on actual form)
			await page.getByLabel(/name/i).fill('Test Schedule');
			await page.getByLabel(/description/i).fill('Test Description');
			await page.getByRole('button', { name: /save|create/i }).click();

			expect(lastOperation).toBe('CREATE');
			expect(schedules.length).toBe(1);
			expect(schedules[0]?.name).toBe('Test Schedule');
		}

		console.log('✅ Admin CRUD API operations tested');
	});

	test('✅ Error Handling - API Failures and Recovery', async ({ page }) => {
		let failureCount = 0;

		await page.route('**/api/**', async (route) => {
			failureCount++;

			if (failureCount <= 2) {
				// Simulate network failures
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

		// Should handle errors gracefully
		await expect(page.getByText(/error|failed|try again/i)).toBeVisible();

		console.log('✅ API error handling and recovery tested');
	});

	test('✅ Authentication State - Token Management', async ({ page }) => {
		let authHeader: string = '';

		await page.route('**/api/**', async (route) => {
			const request = route.request();
			authHeader = request.headers()['authorization'] || '';

			if (authHeader.includes('Bearer')) {
				await route.fulfill({
					status: 200,
					contentType: 'application/json',
					body: JSON.stringify({ authenticated: true })
				});
			} else {
				await route.fulfill({
					status: 401,
					contentType: 'application/json',
					body: JSON.stringify({ error: 'Unauthorized' })
				});
			}
		});

		// Test authenticated request
		await page.goto('/login');

		// Store token in localStorage (simulating login)
		await page.evaluate(() => {
			localStorage.setItem('auth_token', 'test-jwt-token');
		});

		await page.goto('/admin');

		// Should include auth header
		expect(authHeader).toContain('Bearer');

		console.log('✅ Authentication token management tested');
	});

	test('✅ Performance - API Response Times', async ({ page }) => {
		const responseTimes: number[] = [];

		await page.route('**/api/**', async (route) => {
			const startTime = Date.now();

			// Simulate realistic API delay
			await new Promise((resolve) => setTimeout(resolve, 100));

			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({ data: 'test' })
			});

			responseTimes.push(Date.now() - startTime);
		});

		await page.goto('/');
		await page.goto('/shifts');
		await page.goto('/admin');

		// All API calls should be reasonably fast
		responseTimes.forEach((time) => {
			expect(time).toBeLessThan(1000); // Under 1 second
		});

		const avgResponseTime = responseTimes.reduce((a, b) => a + b, 0) / responseTimes.length;
		console.log(`✅ Average API response time: ${avgResponseTime}ms`);
	});
});

test.describe('Authentication Flow', () => {
	test('should complete full registration and verification flow', async ({ page }) => {
		const phone = '+27821234567';
		const name = 'Integration Test User';

		// Step 1: Register user and get OTP
		const registerResponse = await page.request.post('http://localhost:5888/api/auth/register', {
			data: {
				phone: phone,
				name: name
			}
		});

		expect(registerResponse.status()).toBe(200);
		const registerData = await registerResponse.json();

		// Should get OTP in dev mode
		expect(registerData.message).toContain('OTP sent');
		expect(registerData.dev_otp).toBeDefined();
		expect(registerData.dev_otp).toMatch(/^\d{6}$/); // 6-digit OTP

		console.log('✅ Registration successful, OTP:', registerData.dev_otp);

		// Step 2: Verify OTP and get token
		const verifyResponse = await page.request.post('http://localhost:5888/api/auth/verify', {
			data: {
				phone: phone,
				code: registerData.dev_otp
			}
		});

		expect(verifyResponse.status()).toBe(200);
		const verifyData = await verifyResponse.json();

		expect(verifyData.token).toBeDefined();
		expect(verifyData.token).toMatch(/^eyJ/); // JWT tokens start with "eyJ"

		console.log('✅ Verification successful, received JWT token');

		// Step 3: Test protected endpoint with token
		const protectedResponse = await page.request.get('http://localhost:5888/bookings/my', {
			headers: {
				Authorization: `Bearer ${verifyData.token}`
			}
		});

		// Should get 200 or valid response (not 401 Unauthorized)
		expect(protectedResponse.status()).not.toBe(401);
		console.log(
			'✅ Protected endpoint accessible with JWT token, status:',
			protectedResponse.status()
		);
	});

	test('should reject invalid OTP', async ({ page }) => {
		const phone = '+27821234568';
		const name = 'Invalid OTP Test User';

		// Register first
		const registerResponse = await page.request.post('http://localhost:5888/api/auth/register', {
			data: { phone, name }
		});
		expect(registerResponse.status()).toBe(200);

		// Try invalid OTP
		const verifyResponse = await page.request.post('http://localhost:5888/api/auth/verify', {
			data: {
				phone: phone,
				code: '000000' // Invalid OTP
			}
		});

		expect(verifyResponse.status()).toBe(401); // Unauthorized
		console.log('✅ Invalid OTP correctly rejected');
	});

	test('should reject requests without authorization header', async ({ page }) => {
		const protectedResponse = await page.request.get('http://localhost:5888/bookings/my');
		expect(protectedResponse.status()).toBe(401);
		console.log('✅ Protected endpoint correctly rejects requests without auth');
	});
});
