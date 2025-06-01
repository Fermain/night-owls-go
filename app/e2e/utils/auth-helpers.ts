import type { Page } from '@playwright/test';

export const mockUsers = {
	admin: {
		phone: '+27821234567',
		name: 'Admin User',
		role: 'admin',
		token: 'mock-admin-token-12345'
	},
	volunteer: {
		phone: '+27827654321',
		name: 'Volunteer User',
		role: 'volunteer',
		token: 'mock-volunteer-token-67890'
	}
};

export async function setAuthState(
	page: Page,
	user: typeof mockUsers.admin | typeof mockUsers.volunteer
) {
	await page.evaluate((userData) => {
		const authData = {
			user: {
				id: 1,
				name: userData.name,
				phone: userData.phone,
				role: userData.role
			},
			token: userData.token,
			isAuthenticated: true,
			lastLogin: new Date().toISOString()
		};

		localStorage.setItem('user-session', JSON.stringify(authData));
		localStorage.setItem('auth-token', userData.token);
	}, user);

	console.log(`✅ Auth state set for ${user.role}: ${user.name}`);
}

export async function clearAuthState(page: Page) {
	await page.evaluate(() => {
		localStorage.removeItem('user-session');
		localStorage.removeItem('auth-token');
	});

	console.log('✅ Auth state cleared');
}

export async function loginAsAdmin(page: Page) {
	await setAuthState(page, mockUsers.admin);
	await page.goto('/admin');
	console.log('✅ Logged in as admin');
}

export async function loginAsVolunteer(page: Page) {
	await setAuthState(page, mockUsers.volunteer);
	await page.goto('/');
	console.log('✅ Logged in as volunteer');
}

export async function setupAuthMocks(page: Page) {
	// Mock the registration API
	await page.route('**/api/auth/register', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({
				message: 'Verification code sent to your phone',
				dev_otp: '123456'
			})
		});
	});

	// Mock the verification API
	await page.route('**/api/auth/verify', async (route) => {
		await route.fulfill({
			status: 200,
			contentType: 'application/json',
			body: JSON.stringify({
				token: 'mock-jwt-token-12345',
				user: {
					id: 1,
					name: 'Test User',
					phone: '+27821234567',
					role: 'volunteer'
				}
			})
		});
	});

	console.log('✅ Auth API mocks set up');
}
