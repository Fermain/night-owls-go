/**
 * Enhanced API utilities for Night Owls application
 * Provides type-safe, error-handled API requests with authentication
 */

import { userSession, logout } from '../stores/authStore';
import { get } from 'svelte/store';
import {
	classifyError,
	createAppError,
	ErrorType,
	withErrorHandling,
	type AppError
} from './errors';
import type { PaginatedResponse } from '$lib/types/domain';

// === CONFIGURATION ===

const API_TIMEOUT = 30000; // 30 seconds
const API_BASE_URL = '/api';

// === CORE API CLIENT ===

/**
 * Enhanced fetch wrapper with authentication, error handling, and timeouts
 */
export async function authenticatedFetch(
	input: RequestInfo | URL,
	init?: RequestInit & {
		timeout?: number;
		retries?: number;
		skipErrorClassification?: boolean;
	}
): Promise<Response> {
	const {
		timeout = API_TIMEOUT,
		retries = 0,
		skipErrorClassification = false,
		...fetchInit
	} = init || {};

	const session = get(userSession);
	const requestUrl =
		typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;

	const headers = new Headers(fetchInit?.headers);

	// Add authentication for API requests
	if (
		session.isAuthenticated &&
		session.token &&
		(requestUrl.startsWith('/api/') || !requestUrl.startsWith('http'))
	) {
		headers.set('Authorization', `Bearer ${session.token}`);
	}

	// Ensure JSON content type for non-GET requests with body
	if (fetchInit?.body && !headers.has('Content-Type')) {
		headers.set('Content-Type', 'application/json');
	}

	const controller = new AbortController();
	const timeoutId = setTimeout(() => controller.abort(), timeout);

	const modifiedInit: RequestInit = {
		...fetchInit,
		headers,
		signal: controller.signal
	};

	try {
		const response = await fetch(input, modifiedInit);
		clearTimeout(timeoutId);

		if (!response.ok && !skipErrorClassification) {
			// Check for 401 Unauthorized responses and automatically logout
			if (response.status === 401) {
				console.log('Token expired or invalid, logging out user');
				logout(); // This will clear session and redirect to login
			}
			throw classifyError(response);
		}

		return response;
	} catch (error) {
		clearTimeout(timeoutId);

		if (error instanceof DOMException && error.name === 'AbortError') {
			throw createAppError(ErrorType.TIMEOUT_ERROR, 'Request timed out', {
				retryable: true,
				userMessage: 'Request timed out. Please try again.'
			});
		}

		if (skipErrorClassification) {
			throw error;
		}

		throw classifyError(error);
	}
}

// === REQUEST BUILDERS ===

/**
 * Type-safe GET request
 */
export async function apiGet<T = unknown>(
	endpoint: string,
	options: {
		params?: Record<string, string | number | boolean>;
		timeout?: number;
		retries?: number;
	} = {}
): Promise<T> {
	const { params, ...fetchOptions } = options;

	let url = endpoint.startsWith('/') ? endpoint : `${API_BASE_URL}/${endpoint}`;

	if (params) {
		const searchParams = new URLSearchParams();
		Object.entries(params).forEach(([key, value]) => {
			if (value !== undefined && value !== null) {
				searchParams.append(key, String(value));
			}
		});
		url += `?${searchParams.toString()}`;
	}

	const response = await authenticatedFetch(url, {
		method: 'GET',
		...fetchOptions
	});

	return response.json();
}

/**
 * Type-safe POST request
 */
export async function apiPost<TRequest = unknown, TResponse = unknown>(
	endpoint: string,
	data?: TRequest,
	options: {
		timeout?: number;
		retries?: number;
	} = {}
): Promise<TResponse> {
	const url = endpoint.startsWith('/') ? endpoint : `${API_BASE_URL}/${endpoint}`;

	const response = await authenticatedFetch(url, {
		method: 'POST',
		body: data ? JSON.stringify(data) : undefined,
		...options
	});

	return response.json();
}

/**
 * Type-safe PUT request
 */
export async function apiPut<TRequest = unknown, TResponse = unknown>(
	endpoint: string,
	data?: TRequest,
	options: {
		timeout?: number;
		retries?: number;
	} = {}
): Promise<TResponse> {
	const url = endpoint.startsWith('/') ? endpoint : `${API_BASE_URL}/${endpoint}`;

	const response = await authenticatedFetch(url, {
		method: 'PUT',
		body: data ? JSON.stringify(data) : undefined,
		...options
	});

	return response.json();
}

/**
 * Type-safe PATCH request
 */
export async function apiPatch<TRequest = unknown, TResponse = unknown>(
	endpoint: string,
	data?: TRequest,
	options: {
		timeout?: number;
		retries?: number;
	} = {}
): Promise<TResponse> {
	const url = endpoint.startsWith('/') ? endpoint : `${API_BASE_URL}/${endpoint}`;

	const response = await authenticatedFetch(url, {
		method: 'PATCH',
		body: data ? JSON.stringify(data) : undefined,
		...options
	});

	return response.json();
}

/**
 * Type-safe DELETE request
 */
export async function apiDelete<TResponse = unknown>(
	endpoint: string,
	options: {
		timeout?: number;
		retries?: number;
	} = {}
): Promise<TResponse> {
	const url = endpoint.startsWith('/') ? endpoint : `${API_BASE_URL}/${endpoint}`;

	const response = await authenticatedFetch(url, {
		method: 'DELETE',
		...options
	});

	// Some DELETE endpoints return no content
	if (response.status === 204) {
		return {} as TResponse;
	}

	return response.json();
}

// === SPECIALIZED REQUEST PATTERNS ===

/**
 * Paginated GET request with standardized response format
 */
export async function apiGetPaginated<T = unknown>(
	endpoint: string,
	options: {
		limit?: number;
		offset?: number;
		params?: Record<string, string | number | boolean>;
		timeout?: number;
		retries?: number;
	} = {}
): Promise<PaginatedResponse<T>> {
	const { limit = 20, offset = 0, params = {}, ...fetchOptions } = options;

	const allParams = {
		limit,
		offset,
		...params
	};

	return apiGet<PaginatedResponse<T>>(endpoint, {
		params: allParams,
		...fetchOptions
	});
}

/**
 * File upload request
 */
export async function apiUpload<TResponse = unknown>(
	endpoint: string,
	file: File,
	options: {
		additionalData?: Record<string, string>;
		timeout?: number;
		onProgress?: (progress: number) => void;
	} = {}
): Promise<TResponse> {
	const { additionalData, timeout = 60000 } = options;
	const url = endpoint.startsWith('/') ? endpoint : `${API_BASE_URL}/${endpoint}`;

	const formData = new FormData();
	formData.append('file', file);

	if (additionalData) {
		Object.entries(additionalData).forEach(([key, value]) => {
			formData.append(key, value);
		});
	}

	// Note: Don't set Content-Type header for FormData - let browser set it with boundary
	const response = await authenticatedFetch(url, {
		method: 'POST',
		body: formData,
		timeout
	});

	return response.json();
}

// === ERROR HANDLING WRAPPERS ===

/**
 * API request with automatic retry and error handling
 */
export async function apiWithRetry<T>(
	operation: () => Promise<T>,
	options: {
		retries?: number;
		retryDelay?: number;
		onRetry?: (error: AppError, attempt: number) => void;
		onError?: (error: AppError) => void;
	} = {}
): Promise<T> {
	return withErrorHandling(operation, options);
}

// === UTILITY FUNCTIONS ===

/**
 * Check if user is authenticated
 */
export function isAuthenticated(): boolean {
	const session = get(userSession);
	return session.isAuthenticated && !!session.token;
}

/**
 * Get current user from session
 */
export function getCurrentUser() {
	const session = get(userSession);
	if (!session.isAuthenticated) {
		return null;
	}

	return {
		id: session.id,
		name: session.name,
		phone: session.phone,
		role: session.role
	};
}

/**
 * Build API URL with base path
 */
export function buildApiUrl(endpoint: string): string {
	if (endpoint.startsWith('/')) {
		return endpoint;
	}
	return `${API_BASE_URL}/${endpoint}`;
}

/**
 * Extract error message from API response
 */
export async function extractApiError(response: Response): Promise<string> {
	try {
		const data = await response.json();
		return data.message || data.error || `HTTP ${response.status} Error`;
	} catch {
		return `HTTP ${response.status} Error`;
	}
}

// === LEGACY COMPATIBILITY ===
// Keep the original function name for backward compatibility
export { authenticatedFetch as fetch };
