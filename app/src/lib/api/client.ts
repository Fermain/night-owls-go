import { PUBLIC_API_BASE_URL } from '$env/static/public';

const BASE_URL = PUBLIC_API_BASE_URL || 'http://localhost:8080'; // Fallback if not set

// authToken is now loaded directly from localStorage and assumed to be in a browser environment.
let authToken: string | null = localStorage.getItem('jwt_token');

export function setAuthToken(token: string | null) {
  authToken = token;
  if (token) {
    localStorage.setItem('jwt_token', token);
  } else {
    localStorage.removeItem('jwt_token');
  }
  // TODO: Notify authStore of token change
}

export function getAuthToken(): string | null {
  // Re-fetch from localStorage if in-memory authToken is null to ensure consistency,
  // though direct localStorage access on every get might also be an option.
  if (authToken === null) { // Check specifically for null if it could have been cleared
      authToken = localStorage.getItem('jwt_token');
  }
  return authToken;
}

interface RequestOptions extends Omit<RequestInit, 'body' | 'headers'> {
  body?: unknown;
  isPublic?: boolean;
  headers?: HeadersInit;
}

// Custom error type for API client errors to allow distinguishing them
export class ApiClientError extends Error {
  status?: number;
  data?: unknown;
  isApiClientError: boolean = true;
  isNetworkError?: boolean;

  constructor(message: string, status?: number, data?: unknown, isNetworkError?: boolean) {
    super(message);
    this.name = 'ApiClientError';
    this.status = status;
    this.data = data;
    this.isNetworkError = isNetworkError;
    Object.setPrototypeOf(this, ApiClientError.prototype); // Maintain prototype chain
  }
}

async function apiClient<T>(endpoint: string, options: RequestOptions = {}): Promise<T> {
  const { body, isPublic, headers: customHeadersInit, ...customConfig } = options;
  const headers = new Headers(customHeadersInit);

  if (body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  if (!isPublic) {
    const token = getAuthToken(); // getAuthToken now directly uses localStorage if needed
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    } else {
      console.warn('Auth token not found for protected API call:', endpoint);
    }
  }

  const config: RequestInit = {
    method: body ? 'POST' : 'GET',
    ...customConfig,
    headers,
  };

  if (body) {
    config.body = JSON.stringify(body);
  }

  try {
    const response = await fetch(`${BASE_URL}${endpoint}`, config);

    if (!response.ok) {
      let errorData: { message?: string } = { message: response.statusText || 'API request failed' };
      try {
        errorData = await response.json();
      } catch /* istanbul ignore next */ { 
        // Ignore error from parsing non-JSON response, fallback to statusText already set
      }
      throw new ApiClientError(errorData.message || 'API request failed', response.status, errorData);
    }

    if (response.status === 204) {
      return Promise.resolve(null as T);
    }
    return response.json();

  } catch (error) {
    // If it's not already our custom ApiClientError, it might be a network error
    if (!(error instanceof ApiClientError)) {
      console.error('Network or unexpected fetch error:', error);
      throw new ApiClientError((error as Error).message || 'Network error or API unreachable', undefined, undefined, true);
    }
    // Re-throw ApiClientError instances directly
    console.error('API Client Error (re-thrown):', error);
    throw error;
  }
}

export const client = {
  get: <T>(endpoint: string, options?: Omit<RequestOptions, 'body'>) => 
    apiClient<T>(endpoint, { ...options, method: 'GET' }),
  post: <T>(endpoint: string, bodyData: unknown, options?: Omit<RequestOptions, 'body'>) => 
    apiClient<T>(endpoint, { ...options, method: 'POST', body: bodyData }),
  patch: <T>(endpoint: string, bodyData: unknown, options?: Omit<RequestOptions, 'body'>) => 
    apiClient<T>(endpoint, { ...options, method: 'PATCH', body: bodyData }),
  put: <T>(endpoint: string, bodyData: unknown, options?: Omit<RequestOptions, 'body'>) => 
    apiClient<T>(endpoint, { ...options, method: 'PUT', body: bodyData }),
  delete: <T>(endpoint: string, options?: Omit<RequestOptions, 'body'>) => 
    apiClient<T>(endpoint, { ...options, method: 'DELETE' }),
};

// TODO: Implement JWT refresh logic. 