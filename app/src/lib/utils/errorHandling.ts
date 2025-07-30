/**
 * Comprehensive error handling utilities for Night Owls application
 * Centralizes error processing, user notifications, and logging
 */

import { toast } from 'svelte-sonner';

// === ERROR TYPES ===

export interface AppError {
	type: ErrorType;
	message: string;
	code?: string;
	details?: Record<string, unknown>;
	cause?: Error;
}

export type ErrorType =
	| 'NETWORK_ERROR'
	| 'VALIDATION_ERROR'
	| 'AUTHENTICATION_ERROR'
	| 'AUTHORIZATION_ERROR'
	| 'RATE_LIMIT_ERROR'
	| 'SERVER_ERROR'
	| 'NOT_FOUND_ERROR'
	| 'CONFLICT_ERROR'
	| 'UNKNOWN_ERROR';

export interface NotificationOptions {
	duration?: number;
	action?: {
		label: string;
		onClick: () => void;
	};
}

// === ERROR CLASSIFICATION ===

/**
 * Classify errors based on common patterns
 */
export function classifyError(error: unknown): AppError {
	const message = getErrorMessage(error);
	const lowerMessage = message.toLowerCase();

	// Rate limiting
	if (lowerMessage.includes('too many requests') || lowerMessage.includes('rate limit')) {
		return {
			type: 'RATE_LIMIT_ERROR',
			message,
			code: 'RATE_LIMITED'
		};
	}

	// Network errors
	if (
		lowerMessage.includes('network') ||
		lowerMessage.includes('fetch') ||
		lowerMessage.includes('connection') ||
		lowerMessage.includes('timeout')
	) {
		return {
			type: 'NETWORK_ERROR',
			message: 'Network connection error. Please check your internet connection.',
			cause: error instanceof Error ? error : undefined
		};
	}

	// Authentication errors
	if (
		lowerMessage.includes('unauthorized') ||
		lowerMessage.includes('invalid token') ||
		lowerMessage.includes('authentication failed') ||
		lowerMessage.includes('not authenticated')
	) {
		return {
			type: 'AUTHENTICATION_ERROR',
			message: 'Authentication failed. Please log in again.',
			code: 'AUTH_FAILED'
		};
	}

	// Authorization errors
	if (
		lowerMessage.includes('forbidden') ||
		lowerMessage.includes('not authorized') ||
		lowerMessage.includes('access denied') ||
		lowerMessage.includes('permission')
	) {
		return {
			type: 'AUTHORIZATION_ERROR',
			message: 'You do not have permission to perform this action.',
			code: 'ACCESS_DENIED'
		};
	}

	// Validation errors
	if (
		lowerMessage.includes('validation') ||
		lowerMessage.includes('invalid') ||
		lowerMessage.includes('required') ||
		lowerMessage.includes('must be')
	) {
		return {
			type: 'VALIDATION_ERROR',
			message,
			code: 'VALIDATION_FAILED'
		};
	}

	// Not found errors
	if (lowerMessage.includes('not found') || lowerMessage.includes('does not exist')) {
		return {
			type: 'NOT_FOUND_ERROR',
			message,
			code: 'NOT_FOUND'
		};
	}

	// Conflict errors
	if (
		lowerMessage.includes('conflict') ||
		lowerMessage.includes('already exists') ||
		lowerMessage.includes('duplicate')
	) {
		return {
			type: 'CONFLICT_ERROR',
			message,
			code: 'CONFLICT'
		};
	}

	// Server errors
	if (
		lowerMessage.includes('internal server error') ||
		lowerMessage.includes('server error') ||
		lowerMessage.includes('500') ||
		lowerMessage.includes('503')
	) {
		return {
			type: 'SERVER_ERROR',
			message: 'Server error occurred. Please try again later.',
			code: 'SERVER_ERROR'
		};
	}

	// Unknown errors
	return {
		type: 'UNKNOWN_ERROR',
		message: message || 'An unexpected error occurred',
		cause: error instanceof Error ? error : undefined
	};
}

// === NOTIFICATION SERVICE ===

export class NotificationService {
	/**
	 * Show success notification
	 */
	static success(message: string, options?: NotificationOptions): void {
		toast.success(message, {
			duration: options?.duration || 4000,
			action: options?.action
		});
	}

	/**
	 * Show error notification with automatic classification
	 */
	static error(error: unknown, options?: NotificationOptions): void {
		const appError = classifyError(error);

		const toastOptions = {
			duration: this.getErrorDuration(appError.type),
			action: options?.action,
			...options
		};

		toast.error(appError.message, toastOptions);

		// Log error for debugging/monitoring
		ErrorLogger.logError(appError);
	}

	/**
	 * Show warning notification
	 */
	static warning(message: string, options?: NotificationOptions): void {
		toast.warning(message, {
			duration: options?.duration || 5000,
			action: options?.action
		});
	}

	/**
	 * Show info notification
	 */
	static info(message: string, options?: NotificationOptions): void {
		toast.info(message, {
			duration: options?.duration || 4000,
			action: options?.action
		});
	}

	/**
	 * Get appropriate duration based on error type
	 */
	private static getErrorDuration(errorType: ErrorType): number {
		switch (errorType) {
			case 'RATE_LIMIT_ERROR':
				return 8000; // Longer for rate limits
			case 'NETWORK_ERROR':
				return 6000; // Longer for network issues
			case 'AUTHENTICATION_ERROR':
			case 'AUTHORIZATION_ERROR':
				return 7000; // Longer for auth issues
			default:
				return 5000;
		}
	}
}

// === ERROR LOGGER ===

export class ErrorLogger {
	/**
	 * Log errors for monitoring and debugging
	 */
	static logError(error: AppError): void {
		// Console logging for development
		if (import.meta.env.DEV) {
			console.group(`🚨 ${error.type}: ${error.message}`);
			console.error('Error details:', error);
			if (error.cause) {
				console.error('Caused by:', error.cause);
			}
			console.groupEnd();
		}

		// In production, this would send to monitoring service
		// Example: Sentry, LogRocket, DataDog, etc.
		if (import.meta.env.PROD) {
			this.sendToMonitoringService(error);
		}
	}

	/**
	 * Send error to monitoring service (placeholder)
	 */
	private static sendToMonitoringService(_error: AppError): void {
		// Example implementation for production monitoring
		// Replace with your preferred monitoring service

		try {
			// Example: Send to Sentry
			// Sentry.captureException(error);

			// Example: Send to custom logging endpoint
			// fetch('/api/errors', {
			//   method: 'POST',
			//   headers: { 'Content-Type': 'application/json' },
			//   body: JSON.stringify({
			//     type: error.type,
			//     message: error.message,
			//     code: error.code,
			//     url: window.location.href,
			//     timestamp: new Date().toISOString(),
			//     userAgent: navigator.userAgent
			//   })
			// });

			console.warn('Error logging to monitoring service not implemented');
		} catch (logError) {
			console.error('Failed to log error to monitoring service:', logError);
		}
	}
}

// === SPECIFIC ERROR HANDLERS ===

/**
 * Handle authentication errors with appropriate actions
 */
export function handleAuthError(error: unknown): boolean {
	const appError = classifyError(error);

	if (appError.type === 'AUTHENTICATION_ERROR') {
		NotificationService.error(error, {
			action: {
				label: 'Login',
				onClick: () => (window.location.href = '/login')
			}
		});
		return true;
	}

	return false;
}

/**
 * Handle rate limit errors with consistent messaging
 */
export function handleRateLimitError(errorMessage: string): boolean {
	if (errorMessage.includes('Too many requests') || errorMessage.includes('rate limit')) {
		NotificationService.warning(
			'Too many attempts. Please wait a few minutes before trying again.',
			{
				duration: 8000
			}
		);
		return true;
	}
	return false;
}

/**
 * Handle network errors with retry options
 */
export function handleNetworkError(error: unknown, retryFn?: () => void): boolean {
	const appError = classifyError(error);

	if (appError.type === 'NETWORK_ERROR') {
		NotificationService.error(error, {
			action: retryFn
				? {
						label: 'Retry',
						onClick: retryFn
					}
				: undefined
		});
		return true;
	}

	return false;
}

/**
 * Handle validation errors
 */
export function handleValidationError(error: unknown): boolean {
	const appError = classifyError(error);

	if (appError.type === 'VALIDATION_ERROR') {
		NotificationService.error(error);
		return true;
	}

	return false;
}

// === UTILITY FUNCTIONS ===

/**
 * Extract user-friendly error message from any error type
 */
export function getErrorMessage(error: unknown): string {
	if (error instanceof Error) {
		return error.message;
	}
	if (typeof error === 'string') {
		return error;
	}
	if (error && typeof error === 'object' && 'message' in error) {
		return String(error.message);
	}
	return 'An unexpected error occurred';
}

/**
 * Parse API error responses consistently
 */
export function parseAPIError(response: Response, fallbackMessage: string): AppError {
	// Try to extract error from response
	// This would be customized based on your API error format
	return {
		type: response.status >= 500 ? 'SERVER_ERROR' : 'VALIDATION_ERROR',
		message: fallbackMessage,
		code: response.status.toString()
	};
}

// === LEGACY COMPATIBILITY ===
// Keep these for backward compatibility during migration

/** @deprecated Use NotificationService.error instead */
export const showErrorToast = (error: unknown) => NotificationService.error(error);

/** @deprecated Use NotificationService.success instead */
export const showSuccessToast = (message: string) => NotificationService.success(message);

/** @deprecated Use NotificationService.warning instead */
export const showWarningToast = (message: string) => NotificationService.warning(message);
