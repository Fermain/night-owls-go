/**
 * Error handling utilities for Night Owls application
 * Provides consistent error classification, formatting, and handling
 */

// === ERROR TYPES ===

export enum ErrorType {
	// Network errors
	NETWORK_ERROR = 'NETWORK_ERROR',
	TIMEOUT_ERROR = 'TIMEOUT_ERROR',

	// Authentication errors
	UNAUTHORIZED = 'UNAUTHORIZED',
	FORBIDDEN = 'FORBIDDEN',
	TOKEN_EXPIRED = 'TOKEN_EXPIRED',

	// Validation errors
	VALIDATION_ERROR = 'VALIDATION_ERROR',
	INVALID_INPUT = 'INVALID_INPUT',

	// API errors
	NOT_FOUND = 'NOT_FOUND',
	CONFLICT = 'CONFLICT',
	SERVER_ERROR = 'SERVER_ERROR',

	// Client errors
	OFFLINE_ERROR = 'OFFLINE_ERROR',
	STORAGE_ERROR = 'STORAGE_ERROR',

	// Business logic errors
	BOOKING_CONFLICT = 'BOOKING_CONFLICT',
	SCHEDULE_CONFLICT = 'SCHEDULE_CONFLICT',
	PERMISSION_DENIED = 'PERMISSION_DENIED',

	// Database & Migration errors
	DATABASE_CONNECTION = 'DATABASE_CONNECTION',
	DATABASE_MIGRATION = 'DATABASE_MIGRATION',
	DATABASE_DIRTY = 'DATABASE_DIRTY',
	DATABASE_VERSION = 'DATABASE_VERSION',
	MIGRATION_FAILED = 'MIGRATION_FAILED',
	MIGRATION_ROLLBACK = 'MIGRATION_ROLLBACK',
	SCHEMA_MISMATCH = 'SCHEMA_MISMATCH',

	// Unknown
	UNKNOWN_ERROR = 'UNKNOWN_ERROR'
}

export interface AppError extends Error {
	type: ErrorType;
	code?: string;
	statusCode?: number;
	details?: Record<string, unknown>;
	timestamp: Date;
	retryable: boolean;
	userMessage: string;
}

export interface ValidationError {
	field: string;
	message: string;
	code?: string;
	value?: unknown;
}

// === ERROR FACTORY ===

export function createAppError(
	type: ErrorType,
	message: string,
	options: {
		code?: string;
		statusCode?: number;
		details?: Record<string, unknown>;
		cause?: Error;
		retryable?: boolean;
		userMessage?: string;
	} = {}
): AppError {
	const error = new Error(message) as AppError;

	error.type = type;
	error.code = options.code;
	error.statusCode = options.statusCode;
	error.details = options.details;
	error.timestamp = new Date();
	error.retryable = options.retryable ?? false;
	error.userMessage = options.userMessage ?? getDefaultUserMessage(type);
	error.cause = options.cause;

	return error;
}

// === ERROR CLASSIFICATION ===

export function classifyError(error: Error | unknown): AppError {
	if (isAppError(error)) {
		return error;
	}

	// If it's a Response object (fetch error)
	if (error instanceof Response) {
		return classifyResponseError(error);
	}

	// If it's a standard Error
	if (error instanceof Error) {
		const message = error.message;

		// Database error classification
		if (
			message.includes('database') ||
			message.includes('migration') ||
			message.includes('sqlite')
		) {
			const dbErrorType = mapDatabaseError(message);
			return createAppError(dbErrorType, message, {
				cause: error,
				retryable:
					dbErrorType === ErrorType.DATABASE_CONNECTION ||
					dbErrorType === ErrorType.MIGRATION_FAILED,
				statusCode: 500
			});
		}

		return classifyStandardError(error);
	}

	// If it's a string
	if (typeof error === 'string') {
		return createAppError(ErrorType.UNKNOWN_ERROR, error);
	}

	// Unknown error type
	return createAppError(ErrorType.UNKNOWN_ERROR, 'An unknown error occurred', {
		details: { originalError: error }
	});
}

export function classifyResponseError(response: Response): AppError {
	const statusCode = response.status;

	switch (statusCode) {
		case 400:
			return createAppError(ErrorType.VALIDATION_ERROR, 'Invalid request', {
				statusCode,
				retryable: false,
				userMessage: 'Please check your input and try again.'
			});

		case 401:
			return createAppError(ErrorType.UNAUTHORIZED, 'Unauthorized', {
				statusCode,
				retryable: false,
				userMessage: 'Please log in to continue.'
			});

		case 403:
			return createAppError(ErrorType.FORBIDDEN, 'Forbidden', {
				statusCode,
				retryable: false,
				userMessage: 'You do not have permission to perform this action.'
			});

		case 404:
			return createAppError(ErrorType.NOT_FOUND, 'Resource not found', {
				statusCode,
				retryable: false,
				userMessage: 'The requested resource was not found.'
			});

		case 409:
			return createAppError(ErrorType.CONFLICT, 'Conflict', {
				statusCode,
				retryable: false,
				userMessage: 'This action conflicts with existing data.'
			});

		case 429:
			return createAppError(ErrorType.NETWORK_ERROR, 'Too many requests', {
				statusCode,
				retryable: true,
				userMessage: 'Too many requests. Please try again in a moment.'
			});

		case 500:
		case 502:
		case 503:
		case 504:
			return createAppError(ErrorType.SERVER_ERROR, 'Server error', {
				statusCode,
				retryable: true,
				userMessage: 'A server error occurred. Please try again later.'
			});

		default:
			return createAppError(ErrorType.UNKNOWN_ERROR, `HTTP ${statusCode}`, {
				statusCode,
				retryable: statusCode >= 500,
				userMessage: 'An unexpected error occurred. Please try again.'
			});
	}
}

export function classifyStandardError(error: Error): AppError {
	const message = error.message.toLowerCase();

	// Network errors
	if (message.includes('fetch') || message.includes('network') || message.includes('connection')) {
		return createAppError(ErrorType.NETWORK_ERROR, error.message, {
			cause: error,
			retryable: true,
			userMessage: 'Network error. Please check your connection and try again.'
		});
	}

	// Timeout errors
	if (message.includes('timeout') || message.includes('aborted')) {
		return createAppError(ErrorType.TIMEOUT_ERROR, error.message, {
			cause: error,
			retryable: true,
			userMessage: 'Request timed out. Please try again.'
		});
	}

	// Validation errors
	if (message.includes('validation') || message.includes('invalid')) {
		return createAppError(ErrorType.VALIDATION_ERROR, error.message, {
			cause: error,
			retryable: false,
			userMessage: 'Please check your input and try again.'
		});
	}

	// Default to unknown error
	return createAppError(ErrorType.UNKNOWN_ERROR, error.message, {
		cause: error,
		retryable: false
	});
}

// === ERROR UTILITIES ===

export function isAppError(error: unknown): error is AppError {
	return error instanceof Error && 'type' in error && 'timestamp' in error;
}

export function isRetryableError(error: unknown): boolean {
	if (isAppError(error)) {
		return error.retryable;
	}

	// Default retry logic for non-AppErrors
	if (error instanceof Error) {
		const message = error.message.toLowerCase();
		return (
			message.includes('network') ||
			message.includes('timeout') ||
			message.includes('server') ||
			message.includes('503') ||
			message.includes('502')
		);
	}

	return false;
}

export function getErrorMessage(error: unknown): string {
	if (isAppError(error)) {
		return error.userMessage || error.message;
	}

	if (error instanceof Error) {
		return error.message;
	}

	if (typeof error === 'string') {
		return error;
	}

	return 'An unknown error occurred';
}

export function getErrorDetails(error: unknown): Record<string, unknown> {
	if (isAppError(error)) {
		return {
			type: error.type,
			code: error.code,
			statusCode: error.statusCode,
			timestamp: error.timestamp,
			retryable: error.retryable,
			details: error.details
		};
	}

	if (error instanceof Error) {
		return {
			name: error.name,
			message: error.message,
			stack: error.stack
		};
	}

	return { error };
}

// === DEFAULT USER MESSAGES ===

function getDefaultUserMessage(type: ErrorType): string {
	switch (type) {
		case ErrorType.NETWORK_ERROR:
			return 'Network error. Please check your connection and try again.';

		case ErrorType.TIMEOUT_ERROR:
			return 'Request timed out. Please try again.';

		case ErrorType.UNAUTHORIZED:
			return 'Please log in to continue.';

		case ErrorType.FORBIDDEN:
			return 'You do not have permission to perform this action.';

		case ErrorType.TOKEN_EXPIRED:
			return 'Your session has expired. Please log in again.';

		case ErrorType.VALIDATION_ERROR:
		case ErrorType.INVALID_INPUT:
			return 'Please check your input and try again.';

		case ErrorType.NOT_FOUND:
			return 'The requested resource was not found.';

		case ErrorType.CONFLICT:
		case ErrorType.BOOKING_CONFLICT:
		case ErrorType.SCHEDULE_CONFLICT:
			return 'This action conflicts with existing data.';

		case ErrorType.SERVER_ERROR:
			return 'A server error occurred. Please try again later.';

		case ErrorType.OFFLINE_ERROR:
			return 'You are currently offline. Please check your connection.';

		case ErrorType.STORAGE_ERROR:
			return 'Unable to access local storage. Please try again.';

		case ErrorType.PERMISSION_DENIED:
			return 'You do not have permission to perform this action.';

		// Database error messages
		case ErrorType.DATABASE_CONNECTION:
			return 'Unable to connect to database. Please try again later.';

		case ErrorType.DATABASE_DIRTY:
			return 'Database needs manual recovery. Please contact support.';

		case ErrorType.DATABASE_MIGRATION:
		case ErrorType.MIGRATION_FAILED:
			return 'Database update failed. Please try restarting the application.';

		case ErrorType.DATABASE_VERSION:
			return 'Database is up to date.';

		case ErrorType.MIGRATION_ROLLBACK:
			return 'Database rollback required. Please contact support.';

		case ErrorType.SCHEMA_MISMATCH:
			return 'Database schema mismatch. Please contact support.';

		default:
			return 'An unexpected error occurred. Please try again.';
	}
}

// === VALIDATION ERROR UTILITIES ===

export function createValidationError(
	field: string,
	message: string,
	options: {
		code?: string;
		value?: unknown;
	} = {}
): ValidationError {
	return {
		field,
		message,
		code: options.code,
		value: options.value
	};
}

export function formatValidationErrors(errors: ValidationError[]): string {
	if (errors.length === 0) return '';

	if (errors.length === 1) {
		return errors[0].message;
	}

	return errors.map((error) => `${error.field}: ${error.message}`).join(', ');
}

// === ASYNC ERROR HANDLING ===

export async function withErrorHandling<T>(
	operation: () => Promise<T>,
	options: {
		retries?: number;
		retryDelay?: number;
		onRetry?: (error: AppError, attempt: number) => void;
		onError?: (error: AppError) => void;
	} = {}
): Promise<T> {
	const { retries = 0, retryDelay = 1000, onRetry, onError } = options;

	let lastError: AppError;

	for (let attempt = 0; attempt <= retries; attempt++) {
		try {
			return await operation();
		} catch (error) {
			lastError = classifyError(error);

			if (attempt < retries && isRetryableError(lastError)) {
				onRetry?.(lastError, attempt + 1);
				await delay(retryDelay * (attempt + 1)); // Exponential backoff
				continue;
			}

			onError?.(lastError);
			throw lastError;
		}
	}

	throw lastError!;
}

function delay(ms: number): Promise<void> {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

// === BUSINESS DOMAIN ERRORS ===

export function createBookingConflictError(
	message = 'This shift slot is already booked'
): AppError {
	return createAppError(ErrorType.BOOKING_CONFLICT, message, {
		retryable: false,
		userMessage: 'This shift is no longer available. Please choose another time slot.'
	});
}

export function createScheduleConflictError(message = 'Schedule conflict detected'): AppError {
	return createAppError(ErrorType.SCHEDULE_CONFLICT, message, {
		retryable: false,
		userMessage: 'This schedule conflicts with existing schedules.'
	});
}

export function createPermissionDeniedError(message = 'Permission denied'): AppError {
	return createAppError(ErrorType.PERMISSION_DENIED, message, {
		retryable: false,
		userMessage: 'You do not have permission to perform this action.'
	});
}

export function createTokenExpiredError(): AppError {
	return createAppError(ErrorType.TOKEN_EXPIRED, 'Authentication token has expired', {
		retryable: false,
		userMessage: 'Your session has expired. Please log in again.'
	});
}

export function createOfflineError(): AppError {
	return createAppError(ErrorType.OFFLINE_ERROR, 'Application is offline', {
		retryable: true,
		userMessage: 'You are currently offline. Please check your connection.'
	});
}

// === DATABASE DOMAIN ERRORS ===

export function createDatabaseDirtyError(message = 'Database is in dirty state'): AppError {
	return createAppError(ErrorType.DATABASE_DIRTY, message, {
		retryable: false,
		userMessage: 'Database needs manual recovery. Please contact support.',
		details: {
			recoveryInstructions: [
				'Check database migration state',
				'Force migration version if safe',
				'Restore from backup if necessary'
			]
		}
	});
}

export function createMigrationFailedError(message = 'Database migration failed'): AppError {
	return createAppError(ErrorType.MIGRATION_FAILED, message, {
		retryable: true,
		userMessage: 'Database update failed. Please try restarting the application.'
	});
}

export function createDatabaseConnectionError(message = 'Database connection failed'): AppError {
	return createAppError(ErrorType.DATABASE_CONNECTION, message, {
		retryable: true,
		userMessage: 'Unable to connect to database. Please try again later.'
	});
}

/**
 * Maps common database error messages to our error types
 */
function mapDatabaseError(message: string): ErrorType {
	const lowerMessage = message.toLowerCase();

	if (lowerMessage.includes('dirty database version')) {
		return ErrorType.DATABASE_DIRTY;
	}
	if (lowerMessage.includes('no change') || lowerMessage.includes('no migration')) {
		return ErrorType.DATABASE_VERSION;
	}
	if (lowerMessage.includes('migration') && lowerMessage.includes('fail')) {
		return ErrorType.MIGRATION_FAILED;
	}
	if (lowerMessage.includes('database') && lowerMessage.includes('connect')) {
		return ErrorType.DATABASE_CONNECTION;
	}
	if (lowerMessage.includes('schema')) {
		return ErrorType.SCHEMA_MISMATCH;
	}
	if (lowerMessage.includes('migration')) {
		return ErrorType.DATABASE_MIGRATION;
	}

	return ErrorType.UNKNOWN_ERROR;
}
