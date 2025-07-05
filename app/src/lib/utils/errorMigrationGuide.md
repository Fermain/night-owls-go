# Error Handling Migration Guide

This guide shows how to migrate from direct toast calls to the new centralized error handling system.

## Quick Reference

### Before (Old Pattern)

```typescript
import { toast } from 'svelte-sonner';

try {
	await someAPICall();
	toast.success('Operation successful!');
} catch (error) {
	toast.error(`Failed: ${error.message}`);
}
```

### After (New Pattern)

```typescript
import { NotificationService, handleNetworkError } from '$lib/utils/errorHandling';
import { apiClient } from '$lib/utils/apiClient';

const response = await apiClient.post('/some-endpoint', data);
if (response.success) {
	NotificationService.success('Operation successful!');
} else {
	NotificationService.error(response.error);
}
```

## Migration Patterns

### 1. Simple Success/Error Messages

**Before:**

```typescript
toast.success('User created successfully');
toast.error('Failed to create user');
```

**After:**

```typescript
NotificationService.success('User created successfully');
NotificationService.error(error); // Automatically classifies and formats
```

### 2. Network Error Handling with Retry

**Before:**

```typescript
try {
	await fetch('/api/data');
} catch (error) {
	toast.error('Network error occurred');
}
```

**After:**

```typescript
const retry = () => fetchData();
if (!handleNetworkError(error, retry)) {
	NotificationService.error(error);
}
```

### 3. Authentication Errors

**Before:**

```typescript
if (response.status === 401) {
	toast.error('Please log in again', {
		action: { label: 'Login', onClick: () => goto('/login') }
	});
}
```

**After:**

```typescript
if (!handleAuthError(error)) {
	NotificationService.error(error);
}
```

### 4. Form Validation Errors

**Before:**

```typescript
if (!isValid) {
	toast.error('Please fix form errors');
	return;
}
```

**After:**

```typescript
if (!isValid) {
	NotificationService.error('Please fix form errors');
	return;
}
```

### 5. Component Error Boundaries

**Before:**

```svelte
<!-- No error boundary -->
<SomeComponent />
```

**After:**

```svelte
<ErrorBoundary fallbackMessage="Failed to load component">
	<SomeComponent />
</ErrorBoundary>
```

### 6. API Client Usage

**Before:**

```typescript
const response = await fetch('/api/users', {
	method: 'POST',
	headers: { 'Content-Type': 'application/json' },
	body: JSON.stringify(userData)
});

if (!response.ok) {
	const error = await response.text();
	toast.error(`Failed to create user: ${error}`);
	return;
}

const user = await response.json();
toast.success('User created successfully');
```

**After:**

```typescript
const response = await apiClient.post('/users', userData);
if (response.success) {
	NotificationService.success('User created successfully');
	return response.data;
} else {
	NotificationService.error(response.error);
	return null;
}
```

## Component Error Boundaries

### Usage Examples

**Basic Error Boundary:**

```svelte
<ErrorBoundary>
	<ComplexComponent />
</ErrorBoundary>
```

**With Custom Fallback:**

```svelte
<ErrorBoundary fallbackMessage="Dashboard temporarily unavailable" showDetails={false}>
	<AdminDashboard />
</ErrorBoundary>
```

**With Error Callback:**

```svelte
<ErrorBoundary
	onError={(error) => {
		// Custom error handling
		analytics.track('component_error', { error });
	}}
>
	<CriticalComponent />
</ErrorBoundary>
```

## Automatic Error Classification

The system automatically classifies errors:

- **Rate Limit**: "Too many requests" → Shows retry timer
- **Network**: Connection issues → Shows retry button
- **Auth**: 401/403 errors → Shows login button
- **Validation**: Form errors → Focused error display
- **Server**: 5xx errors → Generic server error message

## Production Monitoring

Errors are automatically logged with context:

```typescript
// Development: Console logging with details
// Production: Send to monitoring service

ErrorLogger.logError({
	type: 'AUTHENTICATION_ERROR',
	message: 'Token expired',
	details: {
		requestId: 'REQ_123',
		url: '/api/users',
		userAgent: '...'
	}
});
```

## Best Practices

1. **Use Error Boundaries** around major components
2. **Use API Client** for all HTTP requests
3. **Handle specific errors** with dedicated handlers first
4. **Fallback to NotificationService** for generic errors
5. **Test error states** in development mode

## Gradual Migration

1. Start with new components using the new system
2. Wrap existing components in ErrorBoundary
3. Replace direct toast calls component by component
4. Update API calls to use the new client
5. Remove old toast imports

This migration can be done gradually without breaking existing functionality.
