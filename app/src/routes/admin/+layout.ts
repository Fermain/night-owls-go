import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { browser } from '$app/environment';

// With SSR disabled, we need client-side authentication
export const load: LayoutLoad = async ({ fetch, url }) => {
	// Skip auth check during build time
	if (!browser) {
		return {
			user: null
		};
	}

	try {
		// Validate token with backend
		const response = await fetch('/api/auth/validate', {
			method: 'GET',
			credentials: 'include'
		});

		if (!response.ok) {
			// Invalid token - redirect to login
			throw redirect(302, '/login?redirect=' + encodeURIComponent(url.pathname));
		}

		const user = await response.json();

		// Check admin role
		if (user.role !== 'admin') {
			// Not an admin - redirect to regular user area
			throw redirect(302, '/');
		}

		// Valid admin - return user data
		return {
			user: {
				id: user.id,
				name: user.name,
				role: user.role,
				phone: user.phone
			}
		};
	} catch (error) {
		// If it's already a redirect, throw it
		if (error && typeof error === 'object' && 'status' in error && 'location' in error) {
			throw error;
		}

		// Network error or invalid response - redirect to login
		throw redirect(302, '/login?redirect=' + encodeURIComponent(url.pathname));
	}
};
