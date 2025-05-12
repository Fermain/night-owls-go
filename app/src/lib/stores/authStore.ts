import { persisted } from 'svelte-persisted-store';
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

// Helper function to log in a user (for dummy flow)
export function fakeLogin(phone: string, token: string) {
	userSession.set({
		isAuthenticated: true,
		id: 'dummy-user-id-123', // Dummy user ID
		name: 'Dummy User', // Dummy name
		phone: phone, // Phone used for login
		role: 'admin', // Dummy role
		token: token // Fake token
	});
}

// Helper function to log out a user
export function logout() {
	userSession.set(initialSession);
	// Optionally, also explicitly remove the token from localStorage if it was set outside the store
	// localStorage.removeItem('your-separate-token-key'); // If you had one
}
