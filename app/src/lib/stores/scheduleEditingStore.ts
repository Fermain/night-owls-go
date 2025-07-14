// First, try to import the specific Schedule type we need for the form/detail.
// This might be the same as the one from columns, or a more detailed one if available.
// For now, let's assume it's the same as the one used in the schedules table.
import type { Schedule } from '$lib/types';

// Define the type for the store's value.
// It can hold a Schedule object or be undefined if no schedule is selected for editing/creation.
export type SelectedScheduleState = Schedule | undefined;

// Create a writable store for the selected schedule.
// We are NOT persisting this one to localStorage like selectedUserForForm,
// as schedule selection is primarily driven by URL and cleared on navigation typically.
// If persistence is desired later, `persisted` could be used.

// Using a simple Svelte writable store for now.
// If you find you need `svelte/store` 's `writable` directly:
import { writable } from 'svelte/store';

// Initialize the store with an undefined value, meaning no schedule is selected initially.
export const selectedScheduleForForm = writable<Schedule | undefined>(undefined);
