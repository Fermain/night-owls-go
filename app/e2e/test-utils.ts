import { execSync } from 'child_process';
import { expect, type Page } from '@playwright/test';

export class DatabaseHelper {
	private dbPath = '../night-owls.test.db';

	/**
	 * Get the latest OTP for a phone number from the outbox
	 */
	getLatestOTP(phone: string): string | null {
		try {
			const result = execSync(
				`sqlite3 "${this.dbPath}" "SELECT payload FROM outbox WHERE recipient = '${phone}' AND message_type = 'OTP_VERIFICATION' ORDER BY outbox_id DESC LIMIT 1;"`,
				{ encoding: 'utf8' }
			);

			if (!result.trim()) {
				return null;
			}

			// Parse the JSON payload to extract the OTP
			const payload = JSON.parse(result.trim());
			return payload.otp || null;
		} catch (error) {
			console.error('Failed to get OTP from database:', error);
			return null;
		}
	}

	/**
	 * Get all OTPs for a phone number (useful for debugging)
	 */
	getAllOTPs(phone: string): Array<{ otp: string; outboxId: number; status: string }> {
		try {
			const result = execSync(
				`sqlite3 "${this.dbPath}" "SELECT outbox_id, payload, status FROM outbox WHERE recipient = '${phone}' AND message_type = 'OTP_VERIFICATION' ORDER BY outbox_id DESC;"`,
				{ encoding: 'utf8' }
			);

			if (!result.trim()) {
				return [];
			}

			return result
				.trim()
				.split('\n')
				.map((line) => {
					const [outboxId, payload, status] = line.split('|');
					const parsedPayload = JSON.parse(payload);
					return {
						outboxId: parseInt(outboxId),
						otp: parsedPayload.otp,
						status
					};
				});
		} catch (error) {
			console.error('Failed to get all OTPs from database:', error);
			return [];
		}
	}

	/**
	 * Clean up test data
	 */
	cleanupTestUser(phone: string): void {
		try {
			// Clean up outbox entries
			execSync(`sqlite3 "${this.dbPath}" "DELETE FROM outbox WHERE recipient = '${phone}';"`, {
				encoding: 'utf8'
			});

			// Clean up user and related data
			execSync(`sqlite3 "${this.dbPath}" "DELETE FROM users WHERE phone = '${phone}';"`, {
				encoding: 'utf8'
			});

			// Clean up any bookings for this user (if user_id foreign key exists)
			execSync(
				`sqlite3 "${this.dbPath}" "DELETE FROM bookings WHERE user_id IN (SELECT user_id FROM users WHERE phone = '${phone}');"`,
				{ encoding: 'utf8' }
			);
		} catch (error) {
			console.error('Failed to cleanup test data:', error);
		}
	}

	/**
	 * Check if user exists
	 */
	userExists(phone: string): boolean {
		try {
			const result = execSync(
				`sqlite3 "${this.dbPath}" "SELECT COUNT(*) FROM users WHERE phone = '${phone}';"`,
				{ encoding: 'utf8' }
			);
			return parseInt(result.trim()) > 0;
		} catch (error) {
			console.error('Failed to check if user exists:', error);
			return false;
		}
	}

	/**
	 * Get user details by phone
	 */
	getUserByPhone(
		phone: string
	): { id: number; phone: string; name: string | null; role: string } | null {
		try {
			const result = execSync(
				`sqlite3 "${this.dbPath}" "SELECT user_id, phone, name, role FROM users WHERE phone = '${phone}' LIMIT 1;"`,
				{ encoding: 'utf8' }
			);

			if (!result.trim()) {
				return null;
			}

			const [id, userPhone, name, role] = result.trim().split('|');
			return {
				id: parseInt(id),
				phone: userPhone,
				name: name || null,
				role
			};
		} catch (error) {
			console.error('Failed to get user by phone:', error);
			return null;
		}
	}

	/**
	 * Force trigger outbox processing (useful when OTP is stuck in pending)
	 */
	async waitForOutboxProcessing(
		phone: string,
		maxWaitTimeMs: number = 20000
	): Promise<string | null> {
		const startTime = Date.now();
		let otp: string | null = null;

		while (!otp && Date.now() - startTime < maxWaitTimeMs) {
			// Wait a bit
			await new Promise((resolve) => setTimeout(resolve, 2000));

			// Check for OTP
			otp = this.getLatestOTP(phone);

			if (!otp) {
				// Check if there are any pending items that might need processing
				try {
					const pendingResult = execSync(
						`sqlite3 "${this.dbPath}" "SELECT COUNT(*) FROM outbox WHERE recipient = '${phone}' AND status = 'pending';"`,
						{ encoding: 'utf8' }
					);

					const pendingCount = parseInt(pendingResult.trim());
					if (pendingCount > 0) {
						console.log(
							`Waiting for ${pendingCount} pending outbox items to be processed for ${phone}`
						);
					}
				} catch (error) {
					console.error('Failed to check pending outbox items:', error);
				}
			}
		}

		return otp;
	}

	/**
	 * Update user role by user ID
	 */
	updateUserRole(userId: number, role: string): void {
		try {
			execSync(
				`sqlite3 "${this.dbPath}" "UPDATE users SET role = '${role}' WHERE user_id = ${userId};"`,
				{ encoding: 'utf8' }
			);
		} catch (error) {
			console.error('Failed to update user role:', error);
		}
	}

	/**
	 * Create a user directly in the database
	 */
	createUser(phone: string, name: string, role: string): void {
		try {
			execSync(
				`sqlite3 "${this.dbPath}" "INSERT INTO users (phone, name, role) VALUES ('${phone}', '${name}', '${role}');"`,
				{ encoding: 'utf8' }
			);
		} catch (error) {
			console.error('Failed to create user:', error);
		}
	}
}

export class AuthTestHelper {
	/**
	 * Generate a unique test phone number
	 */
	static generateTestPhone(): string {
		const timestamp = Date.now().toString().slice(-8);
		return `+1555${timestamp}`;
	}

	/**
	 * Generate a unique test name
	 */
	static generateTestName(): string {
		const timestamp = Date.now().toString().slice(-8);
		return `E2E User ${timestamp}`;
	}
}

export const TEST_CONFIG = {
	// Maximum time to wait for OTP generation
	MAX_OTP_WAIT_TIME: 30000,

	// Time between OTP check attempts
	OTP_CHECK_INTERVAL: 2000,

	// Default test timeout for auth operations
	AUTH_TIMEOUT: 15000,

	// Default test phone and name
	DEFAULT_TEST_PHONE: '+1555000E2E',
	DEFAULT_TEST_NAME: 'E2E Test User'
} as const;

// Shared authentication configuration for e2e tests
export const AUTH_CONFIG = {
	ADMIN_PHONE: '+27821234567', // Alice Admin from seeded data
	OTP: '123456' // Dev mode OTP
} as const;

/**
 * Simple admin login using dev mode authentication endpoint
 * This bypasses OTP entirely for reliable e2e testing
 */
export async function loginAsAdmin(page: Page): Promise<void> {
	// First navigate to the app to establish a proper origin for localStorage
	await page.goto('/');
	
	// Call the dev login endpoint directly
	const response = await page.request.post('/api/auth/dev-login', {
		data: {
			phone: AUTH_CONFIG.ADMIN_PHONE
		}
	});

	if (!response.ok()) {
		throw new Error(`Dev login failed: ${response.status()} ${await response.text()}`);
	}

	const result = await response.json();
	const token = result.token;

	if (!token) {
		throw new Error('No token received from dev login endpoint');
	}

	// Set the authentication token in browser storage (now that we have a proper origin)
	await page.evaluate((data) => {
		const userSessionData = {
			isAuthenticated: true,
			id: data.user.id.toString(), // Convert to string as expected by UserSessionData
			name: data.user.name,
			phone: data.user.phone,
			role: data.user.role, // 'admin' | 'owl' | 'guest'
			token: data.token
		};
		localStorage.setItem('user-session', JSON.stringify(userSessionData));
	}, result);

	// Navigate to admin area to verify login worked
	await page.goto('/admin');
	
	// Wait for the page to load and verify we're authenticated
	await page.waitForLoadState('networkidle');
	
	// Verify we didn't get redirected to login
	expect(page.url()).toContain('/admin');
}
