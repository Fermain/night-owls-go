import { userSession } from '../stores/authStore';
import { get } from 'svelte/store';

/**
 * Custom fetch wrapper to automatically inject the auth token
 * for API requests (paths starting with '/api/').
 */
export async function authenticatedFetch(
	input: RequestInfo | URL,
	init?: RequestInit
): Promise<Response> {
	const session = get(userSession);
	const requestUrl =
		typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;

	const headers = new Headers(init?.headers);

	// Check if the request is for our API (starts with '/api/' or is a relative path)
	// and if the user is authenticated with a token.
	if (
		session.isAuthenticated &&
		session.token &&
		(requestUrl.startsWith('/api/') || !requestUrl.startsWith('http'))
	) {
		headers.set('Authorization', `Bearer ${session.token}`);
	}

	const modifiedInit = {
		...init,
		headers
	};

	return fetch(input, modifiedInit);
}

// You can also export specific API call functions here if you prefer, e.g.:
// export async function getSomeData(id: string) {
//     const response = await authenticatedFetch(`/api/some-data/${id}`);
//     if (!response.ok) throw new Error('Failed to fetch some data');
//     return response.json();
// }
