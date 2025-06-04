import { browser } from '$app/environment';
import { isOnline, getQueuedForms, removeQueuedForm } from './offline';
import { apiPost, apiPut, apiDelete } from './api';

let syncInProgress = false;

// Process queued forms when back online
export async function processOfflineQueue(): Promise<void> {
	if (!browser || syncInProgress || !navigator.onLine) {
		return;
	}

	const queue = getQueuedForms();
	if (queue.length === 0) {
		return;
	}

	syncInProgress = true;
	console.log(`Processing ${queue.length} queued form submissions...`);

	for (const item of queue) {
		try {
			// Use our existing API utilities
			switch (item.method.toUpperCase()) {
				case 'POST':
					await apiPost(item.endpoint, item.data);
					break;
				case 'PUT':
					await apiPut(item.endpoint, item.data);
					break;
				case 'DELETE':
					await apiDelete(item.endpoint);
					break;
				default:
					console.warn(`Unsupported method: ${item.method}`);
					continue;
			}

			// Remove successfully processed item
			removeQueuedForm(item.id);
			console.log(`Synced queued submission: ${item.endpoint}`);
		} catch (error) {
			console.error(`Failed to sync queued submission ${item.id}:`, error);
			// Keep item in queue for retry
		}
	}

	syncInProgress = false;
}

// Initialize background sync when online
if (browser) {
	// Process queue when coming back online
	window.addEventListener('online', () => {
		setTimeout(processOfflineQueue, 1000); // Small delay to ensure connection is stable
	});

	// Process queue when page loads if online
	if (navigator.onLine) {
		setTimeout(processOfflineQueue, 2000);
	}

	// Subscribe to our offline store changes
	isOnline.subscribe((online) => {
		if (online) {
			setTimeout(processOfflineQueue, 1000);
		}
	});
}
