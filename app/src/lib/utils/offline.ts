import { writable } from 'svelte/store';
import { browser } from '$app/environment';

// Simple offline store using native browser API
export const isOnline = writable(browser ? navigator.onLine : true);

// Initialize offline detection in browser
if (browser) {
	window.addEventListener('online', () => isOnline.set(true));
	window.addEventListener('offline', () => isOnline.set(false));
}

// Simple offline form queue using localStorage
interface QueuedFormData {
	id: string;
	endpoint: string;
	method: string;
	data: Record<string, unknown>;
	timestamp: number;
}

const QUEUE_KEY = 'night-owls-offline-queue';

// Get queued forms from localStorage
export function getQueuedForms(): QueuedFormData[] {
	if (!browser) return [];
	try {
		const stored = localStorage.getItem(QUEUE_KEY);
		return stored ? JSON.parse(stored) : [];
	} catch {
		return [];
	}
}

// Queue a form submission for when back online
export function queueFormSubmission(
	endpoint: string,
	method: string,
	data: Record<string, unknown>
): string {
	if (!browser) return '';

	const id = `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
	const queued: QueuedFormData = {
		id,
		endpoint,
		method,
		data,
		timestamp: Date.now()
	};

	const queue = getQueuedForms();
	queue.push(queued);
	localStorage.setItem(QUEUE_KEY, JSON.stringify(queue));

	return id;
}

// Remove a queued form (after successful submission)
export function removeQueuedForm(id: string): void {
	if (!browser) return;

	const queue = getQueuedForms().filter((item) => item.id !== id);
	localStorage.setItem(QUEUE_KEY, JSON.stringify(queue));
}

// Clear all queued forms
export function clearQueue(): void {
	if (!browser) return;
	localStorage.removeItem(QUEUE_KEY);
}
