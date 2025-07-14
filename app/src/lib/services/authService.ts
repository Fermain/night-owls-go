import { userSession } from '$lib/stores/authStore';

// API types based on backend schema
export interface RegisterRequest {
	phone: string;
	name?: string;
}

export interface RegisterResponse {
	message: string;
	dev_otp?: string; // Development mode only
}

export interface VerifyRequest {
	phone: string;
	code: string;
}

export interface VerifyResponse {
	token: string;
}

export interface ApiError {
	error: string;
}

interface JWTPayload {
	user_id: number;
	phone: string;
	name: string;
	role: string;
	exp: number;
	iat: number;
	iss: string;
}

// Utility function to decode JWT token
function decodeJWT(token: string): JWTPayload | null {
	try {
		const base64Url = token.split('.')[1];
		const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
		const jsonPayload = decodeURIComponent(
			atob(base64)
				.split('')
				.map(function (c) {
					return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
				})
				.join('')
		);
		return JSON.parse(jsonPayload) as JWTPayload;
	} catch (error) {
		console.error('Failed to decode JWT:', error);
		return null;
	}
}

class AuthService {
	private baseUrl = '/api';

	async register(data: RegisterRequest): Promise<RegisterResponse> {
		const response = await fetch(`${this.baseUrl}/auth/register`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		});

		const result = await response.json();

		if (!response.ok) {
			throw new Error(result.error || 'Registration failed');
		}

		return result;
	}

	async verify(data: VerifyRequest): Promise<VerifyResponse> {
		const response = await fetch(`${this.baseUrl}/auth/verify`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		});

		const result = await response.json();

		if (!response.ok) {
			throw new Error(result.error || 'Verification failed');
		}

		return result;
	}

	async login(phoneNumber: string, _name: string, code: string): Promise<void> {
		const verifyResponse = await this.verify({ phone: phoneNumber, code });

		// Update the user session with the real token
		const decodedToken = decodeJWT(verifyResponse.token);
		if (decodedToken) {
			userSession.set({
				isAuthenticated: true,
				id: decodedToken.user_id.toString(),
				name: decodedToken.name || 'User',
				phone: phoneNumber,
				role: decodedToken.role as 'admin' | 'owl' | 'guest',
				token: verifyResponse.token
			});
		} else {
			throw new Error('Failed to decode JWT token');
		}
	}

	logout(): void {
		userSession.set({
			isAuthenticated: false,
			id: null,
			name: null,
			phone: null,
			role: null,
			token: null
		});
	}
}

export const authService = new AuthService();
