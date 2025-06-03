import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Default to 30 days
const DEFAULT_DAY_RANGE = '30';

// Key for localStorage
const STORAGE_KEY = 'nightowls-shift-day-range';

// Initialize the store with value from localStorage if available
function createShiftDayRangeStore() {
	let initialValue = DEFAULT_DAY_RANGE;

	if (browser) {
		const stored = localStorage.getItem(STORAGE_KEY);
		if (stored) {
			initialValue = stored;
		}
	}

	const { subscribe, set, update } = writable(initialValue);

	return {
		subscribe,
		set: (value: string) => {
			if (browser) {
				localStorage.setItem(STORAGE_KEY, value);
			}
			set(value);
		},
		update,
		reset: () => {
			if (browser) {
				localStorage.removeItem(STORAGE_KEY);
			}
			set(DEFAULT_DAY_RANGE);
		}
	};
}

export const selectedDayRange = createShiftDayRangeStore();

// Day range options
export const dayRangeOptions = [
	{ value: '7', label: 'Next 7 days' },
	{ value: '14', label: 'Next 14 days' },
	{ value: '30', label: 'Next 30 days' },
	{ value: '60', label: 'Next 60 days' }
];

// Helper function to get current day range label
export function getDayRangeLabel(value: string): string {
	return dayRangeOptions.find((opt) => opt.value === value)?.label || 'Next 30 days';
}

// Helper function to get date range for API calls
export function getShiftDateRange(dayRange: string): { from: string; to: string } {
	const days = parseInt(dayRange);
	const from = new Date().toISOString();
	const to = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
	return { from, to };
}
