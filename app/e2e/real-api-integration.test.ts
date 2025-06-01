import { test, expect } from '@playwright/test';

test.describe('🔌 Real API Integration Tests', () => {
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
		expect(registerData.message).toContain('Verification code sent');
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

	test('should handle phone number normalization', async ({ page }) => {
		const phones = [
			'+27821234567', // Already E.164 format
			'0821234567', // South African local format
			'27821234567' // International without +
		];

		for (const phone of phones) {
			const registerResponse = await page.request.post('http://localhost:5888/api/auth/register', {
				data: {
					phone: phone,
					name: 'Normalization Test User'
				}
			});

			// Should either succeed (200) or give proper validation error (400)
			expect([200, 400]).toContain(registerResponse.status());

			if (registerResponse.status() === 200) {
				console.log(`✅ Phone ${phone} accepted and normalized`);
			} else {
				console.log(`✅ Phone ${phone} properly rejected with validation error`);
			}
		}
	});

	test('should test available shifts endpoint', async ({ page }) => {
		const shiftsResponse = await page.request.get('http://localhost:5888/shifts/available');

		expect(shiftsResponse.status()).toBe(200);
		const shiftsData = await shiftsResponse.json();

		expect(Array.isArray(shiftsData)).toBe(true);
		console.log(`✅ Available shifts endpoint returned ${shiftsData.length} shifts`);

		// If there are shifts, verify structure
		if (shiftsData.length > 0) {
			const firstShift = shiftsData[0];
			expect(firstShift).toHaveProperty('schedule_id');
			expect(firstShift).toHaveProperty('start_time');
			expect(firstShift).toHaveProperty('end_time');
			console.log('✅ Shift data structure validated');
		}
	});

	test('should test development login endpoint', async ({ page }) => {
		// First register a user
		const phone = '+27821111111';
		await page.request.post('http://localhost:5888/api/auth/register', {
			data: { phone, name: 'Dev Login Test User' }
		});

		// Then try dev login
		const devLoginResponse = await page.request.post('http://localhost:5888/api/auth/dev-login', {
			data: { phone }
		});

		expect(devLoginResponse.status()).toBe(200);
		const devLoginData = await devLoginResponse.json();

		expect(devLoginData.token).toBeDefined();
		expect(devLoginData.user).toBeDefined();
		expect(devLoginData.user.phone).toBe(phone);

		console.log('✅ Development login successful, role:', devLoginData.user.role);
	});

	test('should handle malformed requests gracefully', async ({ page }) => {
		// Test with invalid JSON
		const invalidJsonResponse = await page.request.post('http://localhost:5888/api/auth/register', {
			data: 'invalid json',
			headers: { 'Content-Type': 'application/json' }
		});
		expect(invalidJsonResponse.status()).toBe(400);

		// Test with missing required fields
		const missingFieldsResponse = await page.request.post(
			'http://localhost:5888/api/auth/register',
			{
				data: { name: 'Test' } // Missing phone
			}
		);
		expect(missingFieldsResponse.status()).toBe(400);

		console.log('✅ Malformed requests properly rejected with 400 status');
	});
});
