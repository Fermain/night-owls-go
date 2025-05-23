import { login, logout, userStore } from '$lib/stores/authStore';
import type { UserSessionData } from '$lib/types';

// Utility functions for testing and development

/**
 * Login as an admin user for testing
 */
export function loginAsAdmin() {
	login({
		id: 'admin-123',
		name: 'Admin User',
		phone: '+27123456789',
		role: 'admin',
		token: 'fake-admin-token'
	});
}

/**
 * Login as a night owl user for testing
 */
export function loginAsOwl() {
	login({
		id: 'owl-456',
		name: 'Night Owl',
		phone: '+27987654321',
		role: 'owl',
		token: 'fake-owl-token'
	});
}

/**
 * Login as a guest user for testing
 */
export function loginAsGuest() {
	login({
		id: 'guest-789',
		name: 'Guest User',
		phone: '+27555000111',
		role: 'guest',
		token: 'fake-guest-token'
	});
}

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
	let authenticated = false;
	userStore.subscribe(user => {
		authenticated = user.isAuthenticated;
	})();
	return authenticated;
}

/**
 * Get current user data
 */
export function getCurrentUser(): UserSessionData | null {
	let currentUser: UserSessionData | null = null;
	userStore.subscribe(user => {
		currentUser = user.isAuthenticated ? user : null;
	})();
	return currentUser;
}

/**
 * Check if current user has admin role
 */
export function isAdmin(): boolean {
	const user = getCurrentUser();
	return user?.role === 'admin';
}

/**
 * Check if current user has owl role
 */
export function isOwl(): boolean {
	const user = getCurrentUser();
	return user?.role === 'owl';
}

// Export the logout function for convenience
export { logout }; 