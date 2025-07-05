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

// Cookie-based authentication helpers
function getCookieValue(name: string): string | null {
	if (typeof document === 'undefined') return null;

	const cookies = document.cookie.split(';');
	for (const cookie of cookies) {
		const [cookieName, cookieValue] = cookie.split('=').map((c) => c.trim());
		if (cookieName === name) {
			return decodeURIComponent(cookieValue);
		}
	}
	return null;
}

function checkCookieAuthentication(): UserSessionData | null {
	const authToken = getCookieValue('auth_token');
	if (!authToken) return null;

	// JWT tokens have 3 parts: header.payload.signature
	try {
		const [, payloadB64] = authToken.split('.');
		if (!payloadB64) return null;

		// Decode JWT payload to extract user info
		const payload = JSON.parse(atob(payloadB64));

		// Check if token is expired
		if (payload.exp && payload.exp * 1000 < Date.now()) {
			return null;
		}

		return {
			isAuthenticated: true,
			id: payload.user_id?.toString() || null,
			name: payload.name || null,
			phone: payload.phone || null,
			role: payload.role || null,
			token: authToken
		};
	} catch (error) {
		console.warn('Failed to parse JWT from cookie:', error);
		return null;
	}
}

// Create the persisted store (fallback for localStorage compatibility)
export const userSession = persisted<UserSessionData>('user-session', initialSession);

// Enhanced store that prioritizes cookie-based auth
function createSecureUserStore() {
	const { subscribe, set, update } = userSession;

	return {
		subscribe,

		// Initialize store by checking cookies first, then localStorage
		init() {
			if (typeof window === 'undefined') return;

			// Check for cookie-based authentication first
			const cookieAuth = checkCookieAuthentication();
			if (cookieAuth) {
				set(cookieAuth);
				return;
			}

			// Fallback to localStorage for backward compatibility
			// The persisted store will automatically load from localStorage
		},

		// Updated login function that handles both cookie and localStorage auth
		login(userData: Partial<UserSessionData>) {
			const newSession: UserSessionData = {
				isAuthenticated: true,
				id: userData.id || null,
				name: userData.name || null,
				phone: userData.phone || null,
				role: userData.role || null,
				token: userData.token || null
			};

			set(newSession);
		},

		// Logout function that clears both cookies and localStorage
		async logout() {
			// Clear localStorage
			set(initialSession);

			// Call logout endpoint to clear HTTP-only cookie
			try {
				await fetch('/api/auth/logout', {
					method: 'POST',
					credentials: 'include', // Include cookies in request
					headers: {
						'Content-Type': 'application/json'
					}
				});
			} catch (error) {
				console.warn('Failed to call logout endpoint:', error);
			}

			// Navigate to login page
			if (typeof window !== 'undefined') {
				window.location.href = '/login';
			}
		},

		// Check if user is authenticated (either via cookie or localStorage)
		isAuthenticated(): boolean {
			// Check cookie first
			const cookieAuth = checkCookieAuthentication();
			if (cookieAuth) return true;

			// Fallback to localStorage
			let currentSession: UserSessionData = initialSession;
			const unsubscribe = userSession.subscribe((session) => (currentSession = session));
			unsubscribe();
			return currentSession.isAuthenticated && !!currentSession.token;
		},

		// Get current user data (prioritize cookie over localStorage)
		getCurrentUser(): UserSessionData {
			const cookieAuth = checkCookieAuthentication();
			if (cookieAuth) return cookieAuth;

			// Fallback to localStorage
			let currentSession: UserSessionData = initialSession;
			const unsubscribe = userSession.subscribe((session) => (currentSession = session));
			unsubscribe();
			return currentSession;
		}
	};
}

// Create the enhanced store
const secureUserStore = createSecureUserStore();

// Export the enhanced store as default
export default secureUserStore;

// Export alias for consistency across components
export const userStore = secureUserStore;

// Legacy helper functions for backward compatibility
export function login(userData: Partial<UserSessionData>) {
	secureUserStore.login(userData);
}

export function fakeLogin(phone: string, token: string) {
	login({
		id: 'dummy-user-id-123',
		name: 'Admin User',
		phone: phone,
		role: 'admin',
		token: token
	});
}

export function logout() {
	secureUserStore.logout();
}

// Initialize the store on module load
if (typeof window !== 'undefined') {
	secureUserStore.init();
}
