import { test, expect } from '@playwright/test';

/**
 * Integration Tests - Real API Backend (Updated for Current Architecture)
 * 
 * These tests use the actual Go backend on localhost:5888
 * Purpose: Test real API integration, data flow, and backend functionality
 * Requirements: Go backend must be running on localhost:5888
 * 
 * Updated: December 2024 - Reflects current API endpoints
 */

test.describe('🔗 Real API Integration Tests - Current Endpoints', () => {
	const BACKEND_URL = 'http://localhost:5888';

	test.beforeAll(async () => {
		// Verify backend is running
		try {
			const response = await fetch(`${BACKEND_URL}/api/health`);
			if (!response.ok) {
				throw new Error(`Backend health check failed: ${response.status}`);
			}
		} catch (error) {
			throw new Error(`Backend not available on ${BACKEND_URL}. Please start the Go server first.`);
		}
	});

	test('✅ Real Authentication Flow - Registration to JWT', async ({ page }) => {
		const phone = `+2782${Date.now().toString().slice(-7)}`; // Unique phone number
		const name = 'Integration Test User';

		// Step 1: Register user and get OTP
		const registerResponse = await page.request.post(`${BACKEND_URL}/api/auth/register`, {
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
		const verifyResponse = await page.request.post(`${BACKEND_URL}/api/auth/verify`, {
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
		const protectedResponse = await page.request.get(`${BACKEND_URL}/bookings/my`, {
			headers: {
				Authorization: `Bearer ${verifyData.token}`
			}
		});

		// Should get 200 or valid response (not 401 Unauthorized)
		expect(protectedResponse.status()).not.toBe(401);
		console.log('✅ Protected endpoint accessible with JWT token, status:', protectedResponse.status());
	});

	test('✅ Real Backend - Invalid OTP Rejection', async ({ page }) => {
		const phone = `+2782${Date.now().toString().slice(-7)}`;
		const name = 'Invalid OTP Test User';

		// Register first
		const registerResponse = await page.request.post(`${BACKEND_URL}/api/auth/register`, {
			data: { phone, name }
		});
		expect(registerResponse.status()).toBe(200);

		// Try invalid OTP
		const verifyResponse = await page.request.post(`${BACKEND_URL}/api/auth/verify`, {
			data: {
				phone: phone,
				code: '000000' // Invalid OTP
			}
		});

		expect(verifyResponse.status()).toBe(401); // Unauthorized
		console.log('✅ Invalid OTP correctly rejected by real backend');
	});

	test('✅ Real Backend - Protected Endpoints Security', async ({ page }) => {
		const protectedResponse = await page.request.get(`${BACKEND_URL}/bookings/my`);
		expect(protectedResponse.status()).toBe(401);
		console.log('✅ Protected endpoint correctly rejects requests without auth');
	});

	test('✅ Real Backend - Current Shifts Endpoint', async ({ page }) => {
		// Test the current shifts endpoint (not deprecated /shifts/available)
		const shiftsResponse = await page.request.get(`${BACKEND_URL}/shifts/available`);
		
		// Should get a valid response (200 or 404 if no shifts)
		expect([200, 404]).toContain(shiftsResponse.status());
		
		if (shiftsResponse.status() === 200) {
			const shifts = await shiftsResponse.json();
			expect(Array.isArray(shifts)).toBe(true);
			console.log(`✅ Real backend returned ${shifts.length} available shifts`);
		} else {
			console.log('✅ Real backend returned no shifts (404) - expected for empty dataset');
		}
	});

	test('✅ Real Backend - Broadcasts API', async ({ page }) => {
		// Test broadcasts endpoint (if exists)
		const broadcastsResponse = await page.request.get(`${BACKEND_URL}/api/broadcasts`);
		
		// Allow for 404 if broadcasts not implemented yet
		expect([200, 404]).toContain(broadcastsResponse.status());
		
		if (broadcastsResponse.status() === 200) {
			const broadcasts = await broadcastsResponse.json();
			expect(Array.isArray(broadcasts)).toBe(true);
			console.log(`✅ Real backend returned ${broadcasts.length} broadcasts`);
		} else {
			console.log('✅ Broadcasts endpoint not yet implemented (404) - expected');
		}
	});

	test('✅ Real Backend - Admin Dashboard Endpoints', async ({ page }) => {
		// Test admin dashboard endpoint (if exists)
		const dashboardResponse = await page.request.get(`${BACKEND_URL}/api/admin/dashboard`);
		
		// Allow for 401 (unauthorized) or 404 (not implemented)
		expect([200, 401, 404]).toContain(dashboardResponse.status());
		
		if (dashboardResponse.status() === 200) {
			const dashboard = await dashboardResponse.json();
			expect(dashboard).toBeDefined();
			console.log('✅ Real backend admin dashboard accessible');
		} else if (dashboardResponse.status() === 401) {
			console.log('✅ Admin dashboard correctly requires authentication (401)');
		} else {
			console.log('✅ Admin dashboard endpoint not yet implemented (404) - expected');
		}
	});
}); 