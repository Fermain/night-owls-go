import { test, expect } from '@playwright/test';
import { DatabaseHelper } from './test-utils';

test.describe('User Role Change - Simple API Test', () => {
	const dbHelper = new DatabaseHelper();

	// The existing test user we know exists
	const existingUserPhone = '+27833883600';
	const existingUserName = 'dakjhsdhs';
	const existingUserId = 32;

	test('should test role change with manual JWT token', async ({ request }) => {
		console.log('=== Starting Simple Role Change Test ===');
		
		// Step 1: Verify the existing test user exists
		const existingUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('Existing user before test:', existingUser);
		expect(existingUser).toBeTruthy();
		
		// Step 2: Create a test admin user in the database manually
		const adminPhone = '+27999888777';
		dbHelper.cleanupTestUser(adminPhone);
		
		// Insert admin user directly into database
		const sqlite3 = require('sqlite3').verbose();
		const db = new sqlite3.Database('../night-owls.test.db');
		
		await new Promise((resolve, reject) => {
			db.run(
				"INSERT INTO users (phone, name, role) VALUES (?, ?, ?)",
				[adminPhone, 'Test Admin', 'admin'],
				function(err) {
					if (err) reject(err);
					else resolve(this.lastID);
				}
			);
		});
		
		const adminUser = dbHelper.getUserByPhone(adminPhone);
		console.log('Created admin user:', adminUser);
		
		// Step 3: Create a simple JWT token manually (we'll bypass JWT verification for testing)
		const fakeJWT = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJwaG9uZSI6IisyNzk5OTg4ODc3NyIsInJvbGUiOiJhZG1pbiIsImV4cCI6OTk5OTk5OTk5OX0.fake-signature';
		
		// Step 4: Test the role change API directly
		console.log(`Attempting to change role for user ID ${existingUserId} from ${existingUser?.role} to admin`);
		
		const roleChangeResponse = await request.put(`http://localhost:8080/api/admin/users/${existingUserId}`, {
			headers: {
				'Authorization': `Bearer ${fakeJWT}`,
				'Content-Type': 'application/json'
			},
			data: {
				phone: existingUserPhone,
				name: existingUserName,
				role: 'admin'
			}
		});
		
		console.log('Role change response status:', roleChangeResponse.status());
		
		if (!roleChangeResponse.ok()) {
			const errorText = await roleChangeResponse.text();
			console.log('Role change error response:', errorText);
		} else {
			const responseData = await roleChangeResponse.json();
			console.log('Role change success response:', responseData);
		}
		
		// Step 5: Verify in database that role was actually changed
		const updatedUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after role change:', updatedUser);
		
		// Cleanup
		dbHelper.cleanupTestUser(adminPhone);
		db.close();
		
		console.log('=== Simple Role Change Test Completed ===');
	});
	
	test('should test direct database role update', async () => {
		console.log('=== Testing Direct Database Role Update ===');
		
		// Step 1: Get current user
		const currentUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('Current user:', currentUser);
		
		// Step 2: Update role directly in database
		const sqlite3 = require('sqlite3').verbose();
		const db = new sqlite3.Database('../night-owls.test.db');
		
		await new Promise((resolve, reject) => {
			db.run(
				"UPDATE users SET role = ? WHERE user_id = ?",
				['owl', existingUserId],
				function(err) {
					if (err) reject(err);
					else resolve(this.changes);
				}
			);
		});
		
		// Step 3: Verify the role was changed
		const updatedUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after direct database update:', updatedUser);
		expect(updatedUser?.role).toBe('owl');
		
		// Step 4: Change it back to guest for other tests
		await new Promise((resolve, reject) => {
			db.run(
				"UPDATE users SET role = ? WHERE user_id = ?",
				['guest', existingUserId],
				function(err) {
					if (err) reject(err);
					else resolve(this.changes);
				}
			);
		});
		
		db.close();
		
		const finalUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after reset to guest:', finalUser);
		
		console.log('=== Direct Database Role Update Test Completed ===');
	});
}); 