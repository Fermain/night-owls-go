import { goto } from '$app/navigation';
import { userSession } from '$lib/stores/authStore';
import { get } from 'svelte/store'; // To get store value non-reactively in load
import { toast } from 'svelte-sonner'; // Added

export const load = async () => {
	const session = get(userSession); // Get current session state

	if (!session.isAuthenticated) {
		toast.error('You must be logged in to access this page.'); // Added
		// If not authenticated, redirect to the login page.
		// The `replaceState: true` option prevents the guarded route
		// from being added to the browser's history stack.
		await goto('/login', { replaceState: true });
		// Return an empty object or specific props if needed for the layout when redirecting,
		// though often the redirect itself is enough and the component won't fully render.
		return { user: null, accessDenied: true }; // Signal to layout/page to not render sensitive things
	}

	if (session.role !== 'admin') {
		toast.error('You are not authorized to access this admin area.'); // Added
		await goto('/login', { replaceState: true }); // Or a generic home '/ '
		return { user: null, accessDenied: true }; // Signal to layout/page
	}

	// If authenticated and is an admin, proceed.
	return {
		user: {
			name: session.name,
			role: session.role,
			phone: session.phone,
			id: session.id
		},
		accessDenied: false
	};
};
