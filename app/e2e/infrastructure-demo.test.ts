import { test, expect } from '@playwright/test';

test.describe('E2E Infrastructure Demo', () => {
	test('Homepage loads with correct content', async ({ page }) => {
		await page.goto('/');

		// Verify the main heading exists (from the homepage)
		await expect(
			page.getByRole('heading', { name: 'Protecting Our Community Together' })
		).toBeVisible();

		// Verify we can see auth links (there are multiple, so get the first one)
		await expect(page.getByRole('link', { name: 'Join Us' }).first()).toBeVisible();
		await expect(page.getByRole('link', { name: 'Sign In' }).first()).toBeVisible();
	});

	test('MSW intercepts and mocks authentication registration', async ({ page }) => {
		// Navigate to registration page
		await page.goto('/register');

		// Mock a registration API call directly to verify MSW is working
		const response = await page.request.post('/api/auth/register', {
			data: {
				name: 'Test User',
				phone: '+27821234567'
			}
		});

		// Verify MSW intercepted and returned our mock response
		expect(response.status()).toBe(200);

		const responseData = await response.json();
		expect(responseData.success).toBe(true);
		expect(responseData.message).toBe('Registration successful!');
		expect(responseData.user.name).toBe('Test User');
		expect(responseData.user.phone).toBe('+27821234567');
	});

	test('MSW intercepts and mocks OTP verification', async ({ page }) => {
		// Test OTP verification endpoint
		const response = await page.request.post('/api/auth/verify', {
			data: {
				phone: '+27821234567',
				otp: '123456'
			}
		});

		expect(response.status()).toBe(200);

		const responseData = await response.json();
		expect(responseData.success).toBe(true);
		expect(responseData.message).toBe('Login successful!');
		expect(responseData.token).toBe('mock-jwt-token');
	});

	test('MSW handles invalid OTP correctly', async ({ page }) => {
		// Test invalid OTP
		const response = await page.request.post('/api/auth/verify', {
			data: {
				phone: '+27821234567',
				otp: 'invalid'
			}
		});

		expect(response.status()).toBe(400);

		const responseData = await response.json();
		expect(responseData.error).toBe('Invalid OTP format');
	});

	test('MSW mocks admin schedule endpoints', async ({ page }) => {
		// Test getting schedules
		const getResponse = await page.request.get('/api/admin/schedules');
		expect(getResponse.status()).toBe(200);

		const schedules = await getResponse.json();
		expect(Array.isArray(schedules)).toBe(true);
		expect(schedules.length).toBeGreaterThan(0);
		expect(schedules[0].name).toBe('Evening Patrol');

		// Test creating a schedule
		const createResponse = await page.request.post('/api/admin/schedules', {
			data: {
				name: 'Test Schedule',
				description: 'Test description',
				cron_expression: '0 12 * * *',
				duration_minutes: 120,
				timezone: 'Africa/Johannesburg',
				positions_available: 2
			}
		});

		expect(createResponse.status()).toBe(200);

		const newSchedule = await createResponse.json();
		expect(newSchedule.name).toBe('Test Schedule');
		expect(newSchedule.description).toBe('Test description');
	});

	test('MSW mocks shift booking endpoints', async ({ page }) => {
		// Test getting available shifts
		const shiftsResponse = await page.request.get('/shifts/available');
		expect(shiftsResponse.status()).toBe(200);

		const shifts = await shiftsResponse.json();
		expect(Array.isArray(shifts)).toBe(true);

		// Test booking a shift
		const bookingResponse = await page.request.post('/bookings', {
			data: {
				shift_id: 1,
				buddy_name: 'Test Buddy'
			}
		});

		expect(bookingResponse.status()).toBe(200);

		const bookingData = await bookingResponse.json();
		expect(bookingData.success).toBe(true);
		expect(bookingData.message).toBe('Shift booked successfully!');
		expect(bookingData.booking.buddy_name).toBe('Test Buddy');
	});

	test('MSW handles shift booking conflicts', async ({ page }) => {
		// Try to book a shift that's already full (shift 2 in our mock data)
		const response = await page.request.post('/bookings', {
			data: {
				shift_id: 2 // This shift is full in our mock data
			}
		});

		expect(response.status()).toBe(400);

		const responseData = await response.json();
		expect(responseData.error).toBe('Shift is full');
	});

	test('Authentication state management works', async ({ page }) => {
		// Go to homepage (unauthenticated)
		await page.goto('/');
		await expect(
			page.getByRole('heading', { name: 'Protecting Our Community Together' })
		).toBeVisible();

		// Simulate login by setting localStorage
		await page.evaluate(() => {
			const userSessionData = {
				isAuthenticated: true,
				id: '1',
				name: 'Test User',
				phone: '+27821234567',
				role: 'admin',
				token: 'mock-jwt-token'
			};
			localStorage.setItem('user-session', JSON.stringify(userSessionData));
		});

		// Reload page to see authenticated view
		await page.reload();

		// Should now see authenticated dashboard
		await expect(page.getByText('Evening, Test')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Emergency' })).toBeVisible();
	});
});
