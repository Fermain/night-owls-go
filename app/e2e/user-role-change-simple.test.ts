import { test, expect } from '@playwright/test';
import { DatabaseHelper } from './test-utils';

test.describe('User Role Change - Simple Test', () => {
	const dbHelper = new DatabaseHelper();

	test('should test direct database role update', async () => {
		console.log('=== Starting Simple Database Role Update Test ===');

		// Get existing user
		const existingUserPhone = '+27833883600';
		const existingUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('Existing user before test:', existingUser);

		if (!existingUser) {
			console.log('Test user does not exist, skipping test');
			return;
		}

		const originalRole = existingUser.role;
		const newRole = originalRole === 'admin' ? 'guest' : 'admin';

		console.log(`Changing role from ${originalRole} to ${newRole}`);

		// Update role using DatabaseHelper
		dbHelper.updateUserRole(existingUser.id, newRole);

		// Verify the change
		const updatedUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after role change:', updatedUser);

		expect(updatedUser?.role).toBe(newRole);

		// Restore original role
		dbHelper.updateUserRole(existingUser.id, originalRole);

		// Verify restoration
		const restoredUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after role restoration:', restoredUser);

		expect(restoredUser?.role).toBe(originalRole);

		console.log('=== Simple Database Role Update Test Completed Successfully ===');
	});

	test('should test role validation', async () => {
		console.log('=== Starting Role Validation Test ===');

		const existingUserPhone = '+27833883600';
		const existingUser = dbHelper.getUserByPhone(existingUserPhone);

		if (!existingUser) {
			console.log('Test user does not exist, skipping test');
			return;
		}

		const originalRole = existingUser.role;

		// Test valid roles
		const validRoles = ['admin', 'owl', 'guest'];
		for (const role of validRoles) {
			console.log(`Testing role: ${role}`);
			dbHelper.updateUserRole(existingUser.id, role);

			const updatedUser = dbHelper.getUserByPhone(existingUserPhone);
			expect(updatedUser?.role).toBe(role);
		}

		// Restore original role
		dbHelper.updateUserRole(existingUser.id, originalRole);

		console.log('=== Role Validation Test Completed Successfully ===');
	});
});
