import { writable } from 'svelte/store';

// Assuming UserData is exported from UserForm.svelte or a central types file
// If UserForm.svelte is the source, we might need to ensure its script context="module" part is accessible
// For now, let's define a local UserData or import if readily available.
// Ideally, UserData would be in a shared types definition.

export interface UserData {
	id: number; // Primary key, non-optional
	phone: string;
	name: string | null;
	created_at: string; // Assuming this is available and non-optional for existing users
	role: string; // Added role
}

export const selectedUserForForm = writable<UserData | undefined>(undefined);
