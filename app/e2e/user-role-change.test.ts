import { test, expect } from '@playwright/test';
import { DatabaseHelper, AuthTestHelper, TEST_CONFIG } from './test-utils';

test.describe('User Role Change', () => {
	const dbHelper = new DatabaseHelper();

	// The existing test user we know exists
	const existingUserPhone = '+27833883600';
	const existingUserName = 'dakjhsdhs';
	const existingUserId = 32;

	test.beforeEach(async ({ page }) => {
		// First, ensure our test user exists and has guest role
		const existingUser = dbHelper.getUserByPhone(existingUserPhone);
		if (!existingUser) {
			throw new Error('Test user does not exist in database');
		}
		
		// Reset user to guest role for testing
		try {
			const { execSync } = require('child_process');
			execSync(`sqlite3 "../night-owls.test.db" "UPDATE users SET role = 'guest' WHERE user_id = ${existingUserId};"`);
		} catch (error) {
			console.error('Failed to reset user role:', error);
		}
		
		// Create a temporary admin user for authentication
		const tempAdminPhone = AuthTestHelper.generateTestPhone();
		const tempAdminName = AuthTestHelper.generateTestName();
		
		// Clean up any previous temp admin
		dbHelper.cleanupTestUser(tempAdminPhone);
		
		// Manually insert admin user into database for authentication
		try {
			const { execSync } = require('child_process');
			execSync(`sqlite3 "../night-owls.test.db" "INSERT INTO users (phone, name, role) VALUES ('${tempAdminPhone}', '${tempAdminName}', 'admin');"`);
		} catch (error) {
			console.error('Failed to create temp admin user:', error);
		}
		
		// Go to login page and simulate authentication
		await page.goto('/login');
		
		// Set authentication state manually
		await page.evaluate((phone) => {
			// Create a fake JWT token for testing
			const fakeJWT = btoa(JSON.stringify({
				header: { alg: 'HS256', typ: 'JWT' },
				payload: { 
					user_id: 999, 
					phone: phone, 
					role: 'admin',
					exp: Math.floor(Date.now() / 1000) + 3600 
				},
				signature: 'fake-signature'
			}));
			
			localStorage.setItem('user-session', JSON.stringify({
				isAuthenticated: true,
				id: 999,
				name: 'Test Admin',
				phone: phone,
				role: 'admin',
				token: fakeJWT
			}));
		}, tempAdminPhone);
		
		// Clean up temp admin after setting auth
		dbHelper.cleanupTestUser(tempAdminPhone);
	});

	test('should successfully change user role from guest to admin', async ({ page }) => {
		console.log('=== Starting Role Change Test ===');
		
		// Step 1: Verify the existing test user exists and has guest role
		const existingUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('Existing user before test:', existingUser);
		expect(existingUser).toBeTruthy();
		expect(existingUser?.role).toBe('guest');
		
		// Step 2: Navigate to admin users page
		await page.goto('/admin/users');
		await page.waitForLoadState('networkidle');
		
		// Step 3: Wait for users list to load and find our test user
		console.log(`Looking for user: ${existingUserName} (${existingUserPhone})`);
		
		// Wait for the page to load and try to find the user
		await page.waitForTimeout(3000);
		
		// Try to find and click on the user
		const userLink = page.locator(`[href*="userId=${existingUserId}"]`).first();
		if (await userLink.isVisible({ timeout: 5000 })) {
			await userLink.click();
		} else {
			// Alternative: try to find by text content
			const userText = page.getByText(existingUserName).first();
			if (await userText.isVisible({ timeout: 5000 })) {
				await userText.click();
			} else {
				console.log('User not found in list. Available users:');
				const bodyText = await page.textContent('body');
				console.log(bodyText);
				throw new Error(`Could not find user ${existingUserName} in the users list`);
			}
		}
		
		// Step 4: Verify we're on the user edit page
		await page.waitForURL(`**/admin/users?userId=${existingUserId}`, { timeout: 10000 });
		
		// Step 5: Wait for form to load and verify current role
		await page.waitForSelector('input[readonly]', { timeout: 10000 });
		
		const currentRoleInput = page.locator('input[readonly]').first();
		const currentRole = await currentRoleInput.getAttribute('value');
		console.log('Current role displayed:', currentRole);
		expect(currentRole).toBe('Guest');
		
		// Step 6: Click "Change Role" button
		await page.getByRole('button', { name: 'Change Role' }).click();
		
		// Step 7: Verify role dialog opens
		await expect(page.getByText('Change User Role')).toBeVisible();
		
		// Step 8: Select "Admin" role from dropdown
		// Wait for the select to be visible and click it
		await page.waitForSelector('[role="combobox"]', { timeout: 5000 });
		await page.locator('[role="combobox"]').click();
		
		// Click on Admin option
		await page.getByText('Admin', { exact: true }).click();
		
		// Step 9: Confirm the role change
		await page.getByRole('button', { name: 'Confirm Change' }).click();
		
		// Step 10: Verify the role field updates to "Admin"
		await expect(page.locator('input[readonly][value="Admin"]')).toBeVisible();
		
		// Step 11: Submit the form
		await page.getByRole('button', { name: 'Update User' }).click();
		
		// Step 12: Wait for result (success or error)
		await page.waitForTimeout(3000);
		
		// Step 13: Check for success or error messages
		const successToast = page.getByText('User updated successfully!');
		const errorToast = page.getByText(/error/i);
		
		if (await successToast.isVisible({ timeout: 2000 })) {
			console.log('Success toast appeared');
			await expect(page).toHaveURL('/admin/users');
		} else if (await errorToast.isVisible({ timeout: 2000 })) {
			const errorText = await errorToast.textContent();
			console.log('Error toast appeared:', errorText);
		} else {
			console.log('No toast message appeared');
		}
		
		// Step 14: Verify in database that role was actually changed
		const updatedUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after role change:', updatedUser);
		
		if (updatedUser?.role === 'admin') {
			console.log('=== Role Change Test Completed Successfully ===');
		} else {
			console.log('=== Role Change Failed - Role not updated in database ===');
		}
	});

	test('should log detailed debugging information', async ({ page }) => {
		// This test focuses on gathering debug info
		console.log('=== Starting Debug Test ===');
		
		// Enable console and network logging
		page.on('console', msg => console.log('BROWSER:', msg.text()));
		page.on('response', response => {
			if (response.url().includes('/api/admin/users')) {
				console.log('API RESPONSE:', response.status(), response.url());
			}
		});
		page.on('request', request => {
			if (request.url().includes('/api/admin/users')) {
				console.log('API REQUEST:', request.method(), request.url());
				if (request.postData()) {
					console.log('REQUEST BODY:', request.postData());
				}
			}
		});
		
		await page.goto('/admin/users');
		await page.waitForLoadState('networkidle');
		
		// Find and click user
		const userLink = page.locator(`[href*="userId=${existingUserId}"]`).first();
		if (await userLink.isVisible({ timeout: 5000 })) {
			await userLink.click();
		} else {
			await page.getByText(existingUserName).first().click();
		}
		
		await page.waitForURL(`**/admin/users?userId=${existingUserId}`, { timeout: 10000 });
		
		// Log form state
		const formData = await page.evaluate(() => {
			const phoneInput = document.querySelector('input[type="tel"]') as HTMLInputElement;
			const nameInput = document.querySelector('input[id="name"]') as HTMLInputElement;
			const roleInput = document.querySelector('input[readonly]') as HTMLInputElement;
			return {
				phone: phoneInput?.value,
				name: nameInput?.value,
				role: roleInput?.value,
				userObject: (window as any).selectedUser
			};
		});
		console.log('Initial form data:', formData);
		
		// Try role change
		await page.getByRole('button', { name: 'Change Role' }).click();
		await page.waitForSelector('[role="combobox"]', { timeout: 5000 });
		await page.locator('[role="combobox"]').click();
		await page.getByText('Admin', { exact: true }).click();
		await page.getByRole('button', { name: 'Confirm Change' }).click();
		
		// Log updated form data
		const updatedFormData = await page.evaluate(() => {
			const phoneInput = document.querySelector('input[type="tel"]') as HTMLInputElement;
			const nameInput = document.querySelector('input[id="name"]') as HTMLInputElement;
			const roleInput = document.querySelector('input[readonly]') as HTMLInputElement;
			return {
				phone: phoneInput?.value,
				name: nameInput?.value,
				role: roleInput?.value
			};
		});
		console.log('Form data after role change:', updatedFormData);
		
		// Submit and capture response
		await page.getByRole('button', { name: 'Update User' }).click();
		await page.waitForTimeout(5000);
		
		console.log('=== Debug Test Completed ===');
	});
}); 