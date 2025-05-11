import { goto } from '$app/navigation';
import { userSession } from '$lib/stores/authStore';
import { get } from 'svelte/store'; // To get store value non-reactively in load

export const load = async () => {
	const session = get(userSession); // Get current session state

	if (!session.isAuthenticated) {
		// If not authenticated, redirect to the login page.
		// The `replaceState: true` option prevents the guarded route
		// from being added to the browser's history stack.
		await goto('/login', { replaceState: true });
		// Return an empty object or specific props if needed for the layout when redirecting,
		// though often the redirect itself is enough and the component won't fully render.
		return { redirected: true };
	}

	// If authenticated, proceed to load the page.
	// You can pass user data to child pages through the return value if needed.
	return {
		user: {
			name: session.name,
			role: session.role,
			phone: session.phone,
			id: session.id
		}
	};
};
