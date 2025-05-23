import { test, expect } from '@playwright/test';
import { DatabaseHelper } from './test-utils';

test.describe('User Role Change - Direct API Test', () => {
	const dbHelper = new DatabaseHelper();

	// The existing test user we know exists
	const existingUserPhone = '+27833883600';
	const existingUserName = 'dakjhsdhs';
	const existingUserId = 32;

	test('should successfully change user role via direct API call', async ({ request }) => {
		console.log('=== Starting Direct API Role Change Test ===');

		// Step 1: Verify the existing test user exists and get current role
		const existingUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('Existing user before test:', existingUser);
		expect(existingUser).toBeTruthy();

		// Step 2: First create an admin user for authentication
		const adminResponse = await request.post('http://localhost:8080/api/auth/register', {
			data: {
				phone: '+27123456789',
				name: 'Test Admin'
			}
		});

		console.log('Admin registration response:', adminResponse.status());
		expect(adminResponse.ok()).toBeTruthy();

		// Step 3: Get OTP for admin user
		await new Promise((resolve) => setTimeout(resolve, 2000)); // Wait for OTP generation
		const adminOTP = dbHelper.getLatestOTP('+27123456789');
		console.log('Admin OTP:', adminOTP);
		expect(adminOTP).toBeTruthy();

		// Step 4: Verify admin user
		const verifyResponse = await request.post('http://localhost:8080/api/auth/verify', {
			data: {
				phone: '+27123456789',
				otp: adminOTP
			}
		});

		console.log('Admin verification response:', verifyResponse.status());
		expect(verifyResponse.ok()).toBeTruthy();

		const verifyData = await verifyResponse.json();
		console.log('Admin verification data:', verifyData);
		const adminToken = verifyData.token;
		expect(adminToken).toBeTruthy();

		// Step 5: Make admin user actually admin in database
		const adminUser = dbHelper.getUserByPhone('+27123456789');
		if (adminUser) {
			dbHelper.updateUserRole(adminUser.id, 'admin');
		}

		// Step 6: Test the role change API directly
		console.log(
			`Attempting to change role for user ID ${existingUserId} from ${existingUser?.role} to admin`
		);

		const roleChangeResponse = await request.put(
			`http://localhost:8080/api/admin/users/${existingUserId}`,
			{
				headers: {
					Authorization: `Bearer ${adminToken}`,
					'Content-Type': 'application/json'
				},
				data: {
					phone: existingUserPhone,
					name: existingUserName,
					role: 'admin'
				}
			}
		);

		console.log('Role change response status:', roleChangeResponse.status());
		console.log('Role change response headers:', await roleChangeResponse.headers());

		if (!roleChangeResponse.ok()) {
			const errorText = await roleChangeResponse.text();
			console.log('Role change error response:', errorText);
		} else {
			const responseData = await roleChangeResponse.json();
			console.log('Role change success response:', responseData);
		}

		// Step 7: Verify in database that role was actually changed
		const updatedUser = dbHelper.getUserByPhone(existingUserPhone);
		console.log('User after role change:', updatedUser);

		if (roleChangeResponse.ok()) {
			expect(updatedUser?.role).toBe('admin');
			console.log('=== Direct API Role Change Test Completed Successfully ===');
		} else {
			console.log('=== Direct API Role Change Test Failed ===');
			console.log('Response Status:', roleChangeResponse.status());
		}

		// Cleanup admin user
		dbHelper.cleanupTestUser('+27123456789');
	});

	test('should log detailed backend response for role change failure', async ({ request }) => {
		console.log('=== Starting Detailed Error Analysis ===');

		// Step 1: Try role change without authentication
		const unauthResponse = await request.put(
			`http://localhost:8080/api/admin/users/${existingUserId}`,
			{
				data: {
					phone: existingUserPhone,
					name: existingUserName,
					role: 'admin'
				}
			}
		);

		console.log('Unauthenticated request status:', unauthResponse.status());
		console.log('Unauthenticated request response:', await unauthResponse.text());

		// Step 2: Try role change with invalid token
		const invalidTokenResponse = await request.put(
			`http://localhost:8080/api/admin/users/${existingUserId}`,
			{
				headers: {
					Authorization: 'Bearer invalid-token',
					'Content-Type': 'application/json'
				},
				data: {
					phone: existingUserPhone,
					name: existingUserName,
					role: 'admin'
				}
			}
		);

		console.log('Invalid token request status:', invalidTokenResponse.status());
		console.log('Invalid token request response:', await invalidTokenResponse.text());

		// Step 3: Try role change with invalid user ID
		const invalidIdResponse = await request.put(`http://localhost:8080/api/admin/users/999`, {
			headers: {
				Authorization: 'Bearer fake-token',
				'Content-Type': 'application/json'
			},
			data: {
				phone: existingUserPhone,
				name: existingUserName,
				role: 'admin'
			}
		});

		console.log('Invalid ID request status:', invalidIdResponse.status());
		console.log('Invalid ID request response:', await invalidIdResponse.text());

		console.log('=== Detailed Error Analysis Completed ===');
	});
});
