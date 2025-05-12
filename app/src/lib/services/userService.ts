import { derived } from 'svelte/store';
import { userSession } from '../stores/authStore';
import type { UserSessionData } from '../types';

// Derived store for just the authentication status
export const isAuthenticated = derived(userSession, ($userSession) => $userSession.isAuthenticated);

// Derived store for the current user object (or null if not authenticated)
export const currentUser = derived<typeof userSession, UserSessionData | null>(
	userSession,
	($userSession) => ($userSession.isAuthenticated ? $userSession : null)
);

// Derived store for the current user's role (or null)
export const currentUserRole = derived(currentUser, ($currentUser) => $currentUser?.role ?? null);

// Helper function to check if the current user is an admin
export const isAdmin = derived(currentUserRole, ($currentUserRole) => $currentUserRole === 'admin');

// Helper function to get user initials (e.g., for avatars)
export function getUserInitials(name: string | null | undefined): string {
	if (!name) return '?';
	const parts = name.split(' ');
	if (parts.length === 1) {
		return parts[0].substring(0, 2).toUpperCase();
	}
	return (parts[0][0] + (parts[parts.length - 1][0] || '')).toUpperCase();
}

// Derived store for the current user's initials
export const currentUserInitials = derived(currentUser, ($currentUser) =>
	getUserInitials($currentUser?.name)
);

// Function to get the current session data directly (not reactive)
export function getCurrentSessionData(): UserSessionData {
	let sessionData!: UserSessionData;
	const unsubscribe = userSession.subscribe((value) => {
		sessionData = value;
	});
	unsubscribe(); // Immediately unsubscribe to get a one-time value
	return sessionData;
}
