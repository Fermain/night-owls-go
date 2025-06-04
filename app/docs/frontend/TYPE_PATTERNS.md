# Type Patterns and Conventions

This document outlines the type system patterns established for the Night Owls frontend application.

## Overview

The type system is organized into three main categories:

1. **Domain Types** (`types/domain.ts`) - Business logic and core entities
2. **UI Types** (`types/ui.ts`) - Component props and UI patterns
3. **API Mappings** (`types/api-mappings.ts`) - Bridge between generated API types and domain types

## File Organization

### Domain Types (`types/domain.ts`)

Contains business entities that represent core application concepts:

```typescript
// Core entities
export interface User { ... }
export interface Schedule { ... }
export interface Booking { ... }

// Enums and constants
export enum ReportSeverity { ... }
export const USER_ROLE_LABELS = { ... }

// Value objects
export interface DateRange { ... }
export interface Location { ... }
```

**Conventions:**

- Use interfaces for entities with behavior
- Use type aliases for unions and primitives
- Include display constants alongside enums
- Document relationships in comments

### UI Types (`types/ui.ts`)

Contains component props, state management, and UI patterns:

```typescript
// Component props extend BaseComponentProps
export interface LoadingProps extends BaseComponentProps {
	isLoading: boolean;
	// ...
}

// Consistent prop patterns
export interface BaseComponentProps {
	className?: string;
	id?: string;
	'data-testid'?: string;
}
```

**Conventions:**

- All component props extend `BaseComponentProps`
- Use consistent naming: `Props` suffix for component props
- Include variant types for component styling
- Document expected children types

### API Mappings (`types/api-mappings.ts`)

Bridges auto-generated API types with domain types:

```typescript
// Mapping functions
export function mapAPIUserToDomain(apiUser: APIUser): User { ... }
export function mapDomainUserToAPI(user: Partial<User>): Partial<APIUser> { ... }

// Batch utilities
export function mapAPIUserArrayToDomain(apiUsers: APIUser[]): User[] { ... }
```

**Conventions:**

- Use `mapAPI[Entity]ToDomain` naming pattern
- Use `mapDomain[Entity]ToAPI` for reverse mapping
- Include batch mapping utilities
- Handle null/undefined gracefully with fallbacks

## Type Safety Patterns

### Null Safety

```typescript
// Always handle null/undefined explicitly
export function safeParseDate(dateString: string | null | undefined): Date | null {
	if (!dateString) return null;
	// ...
}

// Use nullish coalescing for fallbacks
const name = apiUser.name ?? '';
const role = (apiUser.role as UserRole) ?? 'guest';
```

### Type Guards

```typescript
// Create type guards for runtime checking
export function isAppError(error: unknown): error is AppError {
	return error instanceof Error && 'type' in error && 'timestamp' in error;
}
```

### Generic Utilities

```typescript
// Use generics for reusable patterns
export async function apiGet<T = unknown>(endpoint: string): Promise<T> { ... }

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  // ...
}
```

## Component Prop Patterns

### Standard Props

All components should accept these standard props:

```typescript
interface BaseComponentProps {
	className?: string; // For styling
	id?: string; // For accessibility
	'data-testid'?: string; // For testing
}
```

### Event Handler Props

```typescript
// Use consistent naming for event handlers
interface ComponentProps {
	onSubmit?: (data: FormData) => void;
	onCancel?: () => void;
	onChange?: (value: string) => void;
}
```

### Variant Props

```typescript
// Use string literals for variants
type ButtonVariant = 'default' | 'destructive' | 'outline' | 'secondary';
type ButtonSize = 'sm' | 'md' | 'lg';

interface ButtonProps {
	variant?: ButtonVariant;
	size?: ButtonSize;
}
```

## API Integration Patterns

### Request/Response Types

```typescript
// Separate request and response types
interface CreateUserRequest {
	name: string;
	phone: string;
	role?: UserRole;
}

interface CreateUserResponse {
	user: User;
	token: string;
}
```

### Error Handling

```typescript
// Use our error utilities
import { classifyError, getErrorMessage } from '$lib/utils/errors';

try {
	const data = await apiGet<User[]>('/users');
} catch (error) {
	const appError = classifyError(error);
	console.error(getErrorMessage(appError));
}
```

## Form Patterns

### Form State

```typescript
interface FormState<T> {
	values: T;
	errors: Partial<Record<keyof T, string>>;
	touched: Partial<Record<keyof T, boolean>>;
	dirty: boolean;
	valid: boolean;
	submitting: boolean;
}
```

### Validation

```typescript
interface ValidationRule<T> {
	required?: boolean | string;
	minLength?: { value: number; message: string };
	pattern?: { value: RegExp; message: string };
	custom?: (value: T) => string | null;
}
```

## Import Conventions

### Preferred Import Patterns

```typescript
// Use centralized UI exports
import { Button, Card, LoadingState } from '$lib/components/ui';

// Group domain imports
import type { User, Schedule, Booking } from '$lib/types/domain';

// Use specific utility imports
import { formatDateTime, isToday } from '$lib/utils/datetime';
import { classifyError, getErrorMessage } from '$lib/utils/errors';
```

### Avoid These Patterns

```typescript
// ❌ Avoid deep imports
import Button from '$lib/components/ui/button/Button.svelte';

// ❌ Avoid mixing types and values
import { User, UserRole, type Schedule } from '$lib/types/domain';

// ✅ Use separate imports
import type { User, Schedule } from '$lib/types/domain';
import { UserRole } from '$lib/types/domain';
```

## Migration Guidelines

When updating existing code to use these patterns:

1. **Start with types** - Update type definitions first
2. **Update imports** - Use centralized exports where possible
3. **Add error handling** - Replace basic try/catch with our error utilities
4. **Standardize props** - Ensure components extend BaseComponentProps
5. **Test incrementally** - Update and test one component at a time

## Common Pitfalls

1. **Mixing null and undefined** - Be consistent, prefer null for API responses, undefined for optional props
2. **Missing error classification** - Always use our error utilities for API calls
3. **Deep component imports** - Use the centralized UI exports
4. **Inconsistent prop naming** - Follow the established patterns
5. **Forgetting data-testid** - Include for testing infrastructure

## Tools and Validation

- **TypeScript strict mode** - Enabled for maximum type safety
- **ESLint rules** - Enforce import patterns and prop conventions
- **Type-only imports** - Use `import type` for type-only dependencies
- **Generic constraints** - Use extends for better IntelliSense
