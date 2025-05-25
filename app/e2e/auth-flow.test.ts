import { test, expect } from '@playwright/test';
import { DatabaseHelper, AuthTestHelper, TEST_CONFIG } from './test-utils';

test.describe('Authentication Flow', () => {
	const dbHelper = new DatabaseHelper();
	let testPhone: string;
	let testName: string;

	// Generate unique test data for each test run
	test.beforeEach(async () => {
		testPhone = AuthTestHelper.generateTestPhone();
		testName = AuthTestHelper.generateTestName();
		dbHelper.cleanupTestUser(testPhone);
	});

	test.afterEach(async () => {
		dbHelper.cleanupTestUser(testPhone);
	});

	test('should complete full registration and login flow', async ({ page }) => {
		// Step 1: Visit homepage
		await page.goto('/');

		// Verify homepage loads
		await expect(
			page.getByRole('heading', { name: 'Protecting Our Community Together' })
		).toBeVisible();

		// Step 2: Click "Join Us" to go to registration
		await page.getByRole('button', { name: 'Join Us' }).click();
		await expect(page).toHaveURL('/register');

		// Step 3: Fill out registration form
		await expect(page.getByRole('heading', { name: 'Join the Community Watch' })).toBeVisible();

		await page.getByLabel('Full Name').fill(testName);
		await page.getByLabel('Phone Number').fill(testPhone);

		// Step 4: Submit registration
		await page.getByRole('button', { name: 'Create account' }).click();

		// Step 5: Verify success message and redirection to login
		await expect(page.getByText('Registration successful!')).toBeVisible();
		await expect(page).toHaveURL(
			`/login?phone=${encodeURIComponent(testPhone)}&name=${encodeURIComponent(testName)}`
		);

		// Step 6: Verify login page loads with pre-filled data
		await expect(page.getByRole('heading', { name: 'Enter verification code' })).toBeVisible();
		await expect(page.getByText(`We sent a verification code to ${testPhone}`)).toBeVisible();

		// Step 7: Wait for OTP to be generated and processed
		const otp = await dbHelper.waitForOutboxProcessing(testPhone, TEST_CONFIG.MAX_OTP_WAIT_TIME);

		expect(otp).toBeTruthy();
		expect(otp).toHaveLength(6);

		console.log(`Retrieved OTP for ${testPhone}: ${otp}`);

		// Step 8: Enter OTP
		const otpInput = page.locator('[data-input-otp-root]');
		await expect(otpInput).toBeVisible();

		// Fill OTP character by character
		for (let i = 0; i < otp!.length; i++) {
			await page.keyboard.type(otp![i]);
		}

		// Step 9: Submit verification
		await page.getByRole('button', { name: 'Verify & Continue' }).click();

		// Step 10: Verify successful login and redirection to admin
		await expect(page.getByText('Login successful!')).toBeVisible();
		await expect(page).toHaveURL('/admin');

		// Step 11: Verify user is authenticated and admin page loads
		await expect(page.getByText('Admin Dashboard').or(page.getByText('Dashboard'))).toBeVisible({
			timeout: 300
		});

		// Verify user exists in database
		const user = dbHelper.getUserByPhone(testPhone);
		expect(user).toBeTruthy();
		expect(user?.phone).toBe(testPhone);
		expect(user?.name).toBe(testName);
	});

	test('should handle existing user login flow', async ({ page }) => {
		// Pre-create the user by going through registration first
		await page.goto('/register');
		await page.getByLabel('Full Name').fill(testName);
		await page.getByLabel('Phone Number').fill(testPhone);
		await page.getByRole('button', { name: 'Create account' }).click();

		// Wait for user to be created
		await expect(page).toHaveURL(/\/login/);
		await page.waitForTimeout(2000);

		// Clear any existing session
		await page.evaluate(() => localStorage.clear());

		// Now test the login flow for existing user
		await page.goto('/login');

		// Step 1: Fill in phone number for existing user
		await page.getByLabel('Phone Number').fill(testPhone);
		await page.getByRole('button', { name: 'Send verification code' }).click();

		// Step 2: Verify we move to verification step
		await expect(page.getByRole('heading', { name: 'Enter verification code' })).toBeVisible();

		// Step 3: Get OTP from database
		const otp = await dbHelper.waitForOutboxProcessing(testPhone, TEST_CONFIG.MAX_OTP_WAIT_TIME);
		expect(otp).toBeTruthy();

		// Step 4: Enter OTP and verify
		for (let i = 0; i < otp!.length; i++) {
			await page.keyboard.type(otp![i]);
		}

		await page.getByRole('button', { name: 'Verify & Continue' }).click();

		// Step 5: Verify successful login
		await expect(page.getByText('Login successful!')).toBeVisible();
		await expect(page).toHaveURL('/admin');
	});

	test('should handle invalid OTP gracefully', async ({ page }) => {
		// Go through registration
		await page.goto('/register');
		await page.getByLabel('Full Name').fill(testName);
		await page.getByLabel('Phone Number').fill(testPhone);
		await page.getByRole('button', { name: 'Create account' }).click();

		await expect(page).toHaveURL(/\/login/);

		// Try to enter invalid OTP
		const invalidOTP = '123456';
		for (let i = 0; i < invalidOTP.length; i++) {
			await page.keyboard.type(invalidOTP[i]);
		}

		await page.getByRole('button', { name: 'Verify & Continue' }).click();

		// Should show error message
		await expect(page.getByText(/verification failed/i)).toBeVisible();

		// OTP input should be cleared
		const otpInput = page.locator('[data-input-otp-root] input').first();
		await expect(otpInput).toHaveValue('');
	});

	test('should redirect authenticated users away from auth pages', async ({ page }) => {
		// First complete the authentication flow
		await page.goto('/register');
		await page.getByLabel('Full Name').fill(testName);
		await page.getByLabel('Phone Number').fill(testPhone);
		await page.getByRole('button', { name: 'Create account' }).click();

		await expect(page).toHaveURL(/\/login/);

		// Get OTP and complete verification
		const otp = await dbHelper.waitForOutboxProcessing(testPhone, TEST_CONFIG.MAX_OTP_WAIT_TIME);
		expect(otp).toBeTruthy();

		for (let i = 0; i < otp!.length; i++) {
			await page.keyboard.type(otp![i]);
		}

		await page.getByRole('button', { name: 'Verify & Continue' }).click();
		await expect(page).toHaveURL('/admin');

		// Now try to visit auth pages - should redirect to admin
		await page.goto('/register');
		await expect(page).toHaveURL('/admin');

		await page.goto('/login');
		await expect(page).toHaveURL('/admin');

		await page.goto('/');
		await expect(page).toHaveURL('/admin');
	});

	test('should handle back navigation in auth flow', async ({ page }) => {
		// Start registration
		await page.goto('/register');
		await page.getByLabel('Full Name').fill(testName);
		await page.getByLabel('Phone Number').fill(testPhone);
		await page.getByRole('button', { name: 'Create account' }).click();

		await expect(page).toHaveURL(/\/login/);
		await expect(page.getByRole('heading', { name: 'Enter verification code' })).toBeVisible();

		// Click back to registration
		await page.getByRole('button', { name: 'Wrong phone number? Go back' }).click();

		// Should return to registration step
		await expect(page.getByRole('heading', { name: 'Welcome to Community Watch' })).toBeVisible();
		await expect(page.getByLabel('Phone Number')).toHaveValue(testPhone);
		await expect(page.getByLabel('Name (optional)')).toHaveValue(testName);
	});

	test('should handle outbox processing delays', async ({ page }) => {
		// This test specifically checks that we can handle cases where outbox processing is slow
		await page.goto('/register');
		await page.getByLabel('Full Name').fill(testName);
		await page.getByLabel('Phone Number').fill(testPhone);
		await page.getByRole('button', { name: 'Create account' }).click();

		await expect(page).toHaveURL(/\/login/);

		// Check that we have pending items in outbox before they get processed
		const allOTPs = dbHelper.getAllOTPs(testPhone);
		console.log(`OTPs for ${testPhone}:`, allOTPs);

		// Wait for processing with extended timeout
		const otp = await dbHelper.waitForOutboxProcessing(testPhone, 30000);
		expect(otp).toBeTruthy();

		// Verify the OTP works
		for (let i = 0; i < otp!.length; i++) {
			await page.keyboard.type(otp![i]);
		}

		await page.getByRole('button', { name: 'Verify & Continue' }).click();
		await expect(page.getByText('Login successful!')).toBeVisible();
	});
});
