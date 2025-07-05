import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async ({ cookies, fetch }) => {
	// Check for JWT token in cookies (secure, HTTP-only)
	const token = cookies.get('night-owls-session');

	if (!token) {
		// No token - redirect to login
		throw redirect(302, '/login');
	}

	try {
		// Validate token with backend
		const response = await fetch('/api/auth/validate', {
			method: 'GET',
			headers: {
				Cookie: cookies.toString()
			}
		});

		if (!response.ok) {
			// Invalid token - clear cookie and redirect
			cookies.delete('night-owls-session', { path: '/' });
			throw redirect(302, '/login');
		}

		const user = await response.json();

		// Check admin role on server-side
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
		// Network error or invalid response - redirect to login
		cookies.delete('night-owls-session', { path: '/' });
		throw redirect(302, '/login');
	}
};
