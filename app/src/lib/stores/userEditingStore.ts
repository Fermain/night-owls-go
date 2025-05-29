import { writable } from 'svelte/store';
import type { UserData } from '$lib/schemas/user'; // Import UserData from schemas

// UserData interface was previously here, now imported.
// export interface UserData { ... }

export const selectedUserForForm = writable<UserData | undefined>(undefined);
