import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

/**
 * Creates an SSR-safe persisted store that saves to localStorage
 * @param key The localStorage key
 * @param initialValue The initial value if nothing is stored
 * @returns A writable store that persists to localStorage
 */
export function persisted<T>(key: string, initialValue: T): Writable<T> {
	// Determine the initial value based on environment
	let startValue = initialValue;
	
	// Only try to read from localStorage in the browser
	if (browser) {
		const stored = localStorage.getItem(key);
		if (stored) {
			try {
				startValue = JSON.parse(stored);
			} catch (e) {
				console.warn(`Failed to parse stored value for ${key}:`, e);
			}
		}
	}
	
	// Create a regular writable store with the appropriate initial value
	const store = writable<T>(startValue);
	const { subscribe, set, update } = store;
	
	// Custom set that persists to localStorage
	const customSet = (value: T) => {
		set(value);
		if (browser) {
			try {
				localStorage.setItem(key, JSON.stringify(value));
			} catch (e) {
				console.warn(`Failed to save value for ${key}:`, e);
			}
		}
	};
	
	// Custom update that persists to localStorage
	const customUpdate = (updater: (value: T) => T) => {
		update((value) => {
			const newValue = updater(value);
			if (browser) {
				try {
					localStorage.setItem(key, JSON.stringify(newValue));
				} catch (e) {
					console.warn(`Failed to save value for ${key}:`, e);
				}
			}
			return newValue;
		});
	};
	
	return {
		subscribe,
		set: customSet,
		update: customUpdate
	};
} 