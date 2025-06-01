import { type Page } from '@playwright/test';

export interface MockUser {
	id: string;
	name: string;
	phone: string;
	role: 'admin' | 'owl' | 'guest';
	token: string;
}

export const mockUsers = {
	admin: {
		id: '1',
		name: 'Alice Admin',
		phone: '+27821234567',
		role: 'admin' as const,
		token: 'mock-admin-token'
	},
	volunteer: {
		id: '2', 
		name: 'Bob Volunteer',
		phone: '+27821234568',
		role: 'owl' as const,
		token: 'mock-volunteer-token'
	},
	guest: {
		id: '3',
		name: 'Charlie Guest', 
		phone: '+27821234569',
		role: 'guest' as const,
		token: 'mock-guest-token'
	}
};

/**
 * Set authentication state in localStorage to bypass login for tests
 * This must be called after navigating to a page to avoid security errors
 */
export async function setAuthState(page: Page, user: MockUser) {
	// Navigate to homepage first to establish proper origin
	await page.goto('/');
	
	await page.evaluate((userData) => {
		const userSessionData = {
			isAuthenticated: true,
			id: userData.id,
			name: userData.name, 
			phone: userData.phone,
			role: userData.role,
			token: userData.token
		};
		localStorage.setItem('user-session', JSON.stringify(userSessionData));
	}, user);
	
	// Wait for state to be set
	await page.waitForTimeout(200);
}

/**
 * Clear authentication state 
 */
export async function clearAuthState(page: Page) {
	await page.evaluate(() => {
		localStorage.removeItem('user-session');
	});
}

/**
 * Login as admin and navigate to admin dashboard
 */
export async function loginAsAdmin(page: Page) {
	await setAuthState(page, mockUsers.admin);
	await page.goto('/admin');
	await page.waitForLoadState('networkidle');
}

/**
 * Login as volunteer and navigate to shifts page
 */
export async function loginAsVolunteer(page: Page) {
	await setAuthState(page, mockUsers.volunteer);
	await page.goto('/shifts');
	await page.waitForLoadState('networkidle');
}

/**
 * Login as guest 
 */
export async function loginAsGuest(page: Page) {
	await setAuthState(page, mockUsers.guest);
	await page.goto('/');
	await page.waitForLoadState('networkidle');
}

/**
 * Verify current authentication state
 */
export async function getAuthState(page: Page) {
	return await page.evaluate(() => {
		const userSession = localStorage.getItem('user-session');
		return userSession ? JSON.parse(userSession) : null;
	});
}

/**
 * Set up API mocks for authentication routes
 */
export async function setupAuthMocks(page: Page) {
	// Mock authentication verification endpoint
	await page.route('**/api/auth/verify', async (route) => {
		const authState = await getAuthState(page);
		
		if (authState && authState.isAuthenticated) {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify({
					success: true,
					user: authState,
					token: authState.token
				})
			});
		} else {
			await route.fulfill({
				status: 401,
				contentType: 'application/json', 
				body: JSON.stringify({ error: 'Unauthorized' })
			});
		}
	});
	
	// Mock user profile endpoint
	await page.route('**/api/user/profile', async (route) => {
		const authState = await getAuthState(page);
		
		if (authState && authState.isAuthenticated) {
			await route.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(authState)
			});
		} else {
			await route.fulfill({
				status: 401,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'Unauthorized' })
			});
		}
	});
} 