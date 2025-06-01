import { persisted } from '$lib/utils/persisted';
import type { UserSessionData } from '../types'; // Updated import path

// Define the shape of our user session data
// export interface UserSessionData { ... }

// Define the initial state for an unauthenticated user
const initialSession: UserSessionData = {
	isAuthenticated: false,
	id: null,
	name: null,
	phone: null,
	role: null,
	token: null
};

// Create the persisted store
// The first parameter is the localStorage key.
// The second parameter is the initial value.
export const userSession = persisted<UserSessionData>('user-session', initialSession);

// Export alias for consistency across components
export const userStore = userSession;

// Helper function to log in a user
export function login(userData: Partial<UserSessionData>) {
	userSession.set({
		isAuthenticated: true,
		id: userData.id || null,
		name: userData.name || null,
		phone: userData.phone || null,
		role: userData.role || null,
		token: userData.token || null
	});
}

// Helper function for fake login (backward compatibility)
export function fakeLogin(phone: string, token: string) {
	login({
		id: 'dummy-user-id-123',
		name: 'Admin User',
		phone: phone,
		role: 'admin',
		token: token
	});
}

// Helper function to log out a user
export function logout() {
	userSession.set(initialSession);
	// Navigate to login page
	if (typeof window !== 'undefined') {
		window.location.href = '/login';
	}
}
