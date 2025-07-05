/**
 * Standardized API client for Night Owls application
 * Provides consistent error handling, request/response processing, and logging
 */

import { type AppError, classifyError, ErrorLogger } from './errorHandling';

// === API TYPES ===

export interface APIResponse<T = unknown> {
	data?: T;
	error?: AppError;
	success: boolean;
	message?: string;
}

export interface APIRequestConfig {
	headers?: Record<string, string>;
	timeout?: number;
	retries?: number;
	retryDelay?: number;
}

// === API CLIENT ===

export class APIClient {
	private baseURL: string;
	private defaultHeaders: Record<string, string>;
	private defaultTimeout: number;

	constructor(baseURL = '/api', defaultHeaders: Record<string, string> = {}) {
		this.baseURL = baseURL;
		this.defaultHeaders = {
			'Content-Type': 'application/json',
			...defaultHeaders
		};
		this.defaultTimeout = 10000; // 10 seconds
	}

	/**
	 * GET request with error handling
	 */
	async get<T>(endpoint: string, config?: APIRequestConfig): Promise<APIResponse<T>> {
		return this.request<T>('GET', endpoint, undefined, config);
	}

	/**
	 * POST request with error handling
	 */
	async post<T>(
		endpoint: string,
		data?: unknown,
		config?: APIRequestConfig
	): Promise<APIResponse<T>> {
		return this.request<T>('POST', endpoint, data, config);
	}

	/**
	 * PUT request with error handling
	 */
	async put<T>(
		endpoint: string,
		data?: unknown,
		config?: APIRequestConfig
	): Promise<APIResponse<T>> {
		return this.request<T>('PUT', endpoint, data, config);
	}

	/**
	 * DELETE request with error handling
	 */
	async delete<T>(endpoint: string, config?: APIRequestConfig): Promise<APIResponse<T>> {
		return this.request<T>('DELETE', endpoint, undefined, config);
	}

	/**
	 * Core request method with comprehensive error handling
	 */
	private async request<T>(
		method: string,
		endpoint: string,
		data?: unknown,
		config?: APIRequestConfig
	): Promise<APIResponse<T>> {
		const url = `${this.baseURL}${endpoint}`;
		const requestId = this.generateRequestId();
		const startTime = Date.now();

		try {
			// Prepare request options
			const requestOptions: RequestInit = {
				method,
				headers: {
					...this.defaultHeaders,
					...config?.headers
				},
				signal: AbortSignal.timeout(config?.timeout || this.defaultTimeout)
			};

			// Add body for non-GET requests
			if (data && method !== 'GET') {
				requestOptions.body = JSON.stringify(data);
			}

			// Log request in development
			if (import.meta.env.DEV) {
				console.log(`🔄 API ${method} ${url}`, { requestId, data });
			}

			// Make the request with retries
			const response = await this.requestWithRetries(
				url,
				requestOptions,
				config?.retries || 0,
				config?.retryDelay || 1000
			);
			const duration = Date.now() - startTime;

			// Parse response
			const result = await this.parseResponse<T>(response, requestId, duration);

			// Log successful response in development
			if (import.meta.env.DEV && result.success) {
				console.log(`✅ API ${method} ${url} (${duration}ms)`, { requestId, result });
			}

			return result;
		} catch (error) {
			const duration = Date.now() - startTime;
			const appError = classifyError(error);

			// Enhanced error with request context
			const enhancedError: AppError = {
				...appError,
				details: {
					...appError.details,
					requestId,
					method,
					url,
					duration,
					data
				}
			};

			// Log error
			ErrorLogger.logError(enhancedError);

			// Log error in development
			if (import.meta.env.DEV) {
				console.error(`❌ API ${method} ${url} (${duration}ms)`, {
					requestId,
					error: enhancedError
				});
			}

			return {
				success: false,
				error: enhancedError
			};
		}
	}

	/**
	 * Request with automatic retries for transient failures
	 */
	private async requestWithRetries(
		url: string,
		options: RequestInit,
		retries: number,
		retryDelay: number
	): Promise<Response> {
		let lastError: Error | null = null;

		for (let attempt = 0; attempt <= retries; attempt++) {
			try {
				const response = await fetch(url, options);

				// Don't retry on client errors (4xx) or successful responses
				if (response.ok || (response.status >= 400 && response.status < 500)) {
					return response;
				}

				// Only retry on server errors (5xx) or network issues
				if (attempt < retries) {
					await this.sleep(retryDelay * (attempt + 1)); // Exponential backoff
					continue;
				}

				return response;
			} catch (error) {
				lastError = error instanceof Error ? error : new Error(String(error));

				if (attempt < retries) {
					await this.sleep(retryDelay * (attempt + 1));
					continue;
				}

				throw lastError;
			}
		}

		throw lastError || new Error('Request failed after all retries');
	}

	/**
	 * Parse response with standardized error handling
	 */
	private async parseResponse<T>(
		response: Response,
		requestId: string,
		duration: number
	): Promise<APIResponse<T>> {
		try {
			// Handle successful responses
			if (response.ok) {
				const contentType = response.headers.get('content-type');

				if (contentType?.includes('application/json')) {
					const data = await response.json();
					return {
						success: true,
						data: data as T
					};
				} else {
					// Handle non-JSON responses (like text)
					const text = await response.text();
					return {
						success: true,
						data: text as unknown as T
					};
				}
			}

			// Handle error responses
			let errorMessage = `HTTP ${response.status}: ${response.statusText}`;
			let errorDetails: Record<string, unknown> = {};

			try {
				const errorBody = await response.text();
				if (errorBody) {
					try {
						const parsedError = JSON.parse(errorBody);
						errorMessage = parsedError.message || parsedError.error || errorMessage;
						errorDetails = parsedError;
					} catch {
						// If not JSON, use the text as the error message
						if (errorBody.length < 200) {
							// Avoid huge error messages
							errorMessage = errorBody;
						}
					}
				}
			} catch {
				// Failed to read error body, use status text
			}

			const appError: AppError = {
				type: this.getErrorTypeFromStatus(response.status),
				message: errorMessage,
				code: response.status.toString(),
				details: {
					...errorDetails,
					requestId,
					duration,
					status: response.status,
					statusText: response.statusText
				}
			};

			return {
				success: false,
				error: appError
			};
		} catch (parseError) {
			const appError: AppError = {
				type: 'UNKNOWN_ERROR',
				message: 'Failed to parse response',
				cause: parseError instanceof Error ? parseError : undefined,
				details: {
					requestId,
					duration,
					status: response.status
				}
			};

			return {
				success: false,
				error: appError
			};
		}
	}

	/**
	 * Map HTTP status codes to error types
	 */
	private getErrorTypeFromStatus(status: number): AppError['type'] {
		if (status === 401) return 'AUTHENTICATION_ERROR';
		if (status === 403) return 'AUTHORIZATION_ERROR';
		if (status === 404) return 'NOT_FOUND_ERROR';
		if (status === 409) return 'CONFLICT_ERROR';
		if (status === 422) return 'VALIDATION_ERROR';
		if (status === 429) return 'RATE_LIMIT_ERROR';
		if (status >= 500) return 'SERVER_ERROR';
		if (status >= 400) return 'VALIDATION_ERROR';
		return 'UNKNOWN_ERROR';
	}

	/**
	 * Generate unique request ID for tracking
	 */
	private generateRequestId(): string {
		return `REQ_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
	}

	/**
	 * Sleep utility for retries
	 */
	private sleep(ms: number): Promise<void> {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}
}

// === DEFAULT CLIENT INSTANCE ===

export const apiClient = new APIClient();

// === TYPED API HELPERS ===

/**
 * Create a typed API client for specific endpoints
 */
export function createTypedAPI<T extends Record<string, unknown>>(baseEndpoint: string) {
	return {
		list: () => apiClient.get<T[]>(baseEndpoint),
		get: (id: string | number) => apiClient.get<T>(`${baseEndpoint}/${id}`),
		create: (data: Partial<T>) => apiClient.post<T>(baseEndpoint, data),
		update: (id: string | number, data: Partial<T>) =>
			apiClient.put<T>(`${baseEndpoint}/${id}`, data),
		delete: (id: string | number) => apiClient.delete(`${baseEndpoint}/${id}`)
	};
}

// === AUTHENTICATION HELPERS ===

/**
 * Set authentication token for all requests
 */
export function setAuthToken(token: string): void {
	apiClient['defaultHeaders']['Authorization'] = `Bearer ${token}`;
}

/**
 * Clear authentication token
 */
export function clearAuthToken(): void {
	delete apiClient['defaultHeaders']['Authorization'];
}

// === RESPONSE HELPERS ===

/**
 * Extract data from API response or throw error
 */
export function unwrapResponse<T>(response: APIResponse<T>): T {
	if (response.success && response.data !== undefined) {
		return response.data;
	}

	throw response.error || new Error('API response was not successful');
}

/**
 * Check if API response was successful
 */
export function isSuccessResponse<T>(
	response: APIResponse<T>
): response is APIResponse<T> & { success: true; data: T } {
	return response.success && response.data !== undefined;
}
