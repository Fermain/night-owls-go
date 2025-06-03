import { writable, derived, get } from 'svelte/store';
import { authenticatedFetch } from '$lib/utils/api';
import { messageStorage } from './messageStorageService';
import { userSession } from '$lib/stores/authStore';

interface NotificationData {
	broadcastId?: number;
	audience?: string;
	shiftId?: number;
	bookingId?: number;
}

interface BroadcastResponse {
	id: number;
	message: string;
	created_at: string;
	audience: string;
}

export interface UserNotification {
	id: number;
	type: 'broadcast' | 'shift_reminder' | 'shift_assignment' | 'system';
	title: string;
	message: string;
	timestamp: string;
	read: boolean;
	data?: NotificationData;
}

export interface NotificationState {
	notifications: UserNotification[];
	unreadCount: number;
	isLoading: boolean;
	lastFetched: string | null;
}

// Create notification store
const createNotificationStore = () => {
	const initialState: NotificationState = {
		notifications: [],
		unreadCount: 0,
		isLoading: false,
		lastFetched: null
	};

	const { subscribe, set, update } = writable(initialState);

	return {
		subscribe,

		// Initialize from IndexedDB
		async init() {
			try {
				// Dexie handles database initialization automatically
				const storedMessages = await messageStorage.getMessages();
				const unreadCount = await messageStorage.getUnreadCount();

				// Convert stored messages to UserNotification format
				const notifications: UserNotification[] = storedMessages.map((stored) => ({
					id: stored.id,
					type: 'broadcast' as const,
					title: stored.title,
					message: stored.message,
					timestamp: stored.timestamp,
					read: stored.read,
					data: {
						broadcastId: stored.id,
						audience: stored.audience
					}
				}));

				update((state) => ({
					...state,
					notifications,
					unreadCount,
					lastFetched: storedMessages.length > 0 ? new Date().toISOString() : null
				}));

				console.log('📦 Notification service initialized with', notifications.length, 'messages');
			} catch (error) {
				console.error('Failed to initialize notifications from storage:', error);
			}
		},

		// Actions
		async fetchNotifications(force = false) {
			// Check if user is authenticated before making API calls
			const session = get(userSession);
			if (!session.isAuthenticated) {
				console.log('📦 Skipping notification fetch - user not authenticated');
				return;
			}

			// Additional safeguard - ensure we have a valid session
			if (!session.id) {
				console.log('📦 Skipping notification fetch - invalid user session');
				return;
			}

			// Don't fetch if already loading and not forced
			const currentState: NotificationState = get({ subscribe });
			if (currentState.isLoading && !force) {
				return;
			}

			// Don't fetch too frequently unless forced
			if (!force && currentState.lastFetched) {
				const timeSinceLastFetch = Date.now() - new Date(currentState.lastFetched).getTime();
				if (timeSinceLastFetch < 5000) {
					// 5 seconds minimum between fetches
					return;
				}
			}

			update((state) => ({ ...state, isLoading: true }));

			try {
				console.log('📦 Fetching notifications for user:', session.id);
				// Fetch broadcasts from API
				const response = await authenticatedFetch('/api/broadcasts');

				if (!response.ok) {
					throw new Error(`HTTP ${response.status}: ${response.statusText}`);
				}

				// Check if response has content
				const contentType = response.headers.get('content-type');
				if (!contentType || !contentType.includes('application/json')) {
					console.warn('📦 Unexpected content type from broadcasts API:', contentType);
					// Return empty array if not JSON
					update((state) => ({
						...state,
						notifications: [],
						unreadCount: 0,
						isLoading: false,
						lastFetched: new Date().toISOString()
					}));
					return;
				}

				const broadcasts: BroadcastResponse[] = await response.json();

				// Debug logging to understand API response
				console.log('📦 Broadcasts API response:', {
					type: typeof broadcasts,
					isArray: Array.isArray(broadcasts),
					isNull: broadcasts === null,
					length: broadcasts?.length,
					data: broadcasts
				});

				// Handle case where API returns null instead of empty array
				const broadcastsArray = broadcasts || [];

				// Validate that it's actually an array
				if (!Array.isArray(broadcastsArray)) {
					console.error('📦 API returned non-array data:', broadcastsArray);
					throw new Error('API returned invalid data format');
				}

				// Transform broadcasts into notifications
				const apiNotifications: UserNotification[] = broadcastsArray.map((broadcast) => ({
					id: broadcast.id,
					type: 'broadcast' as const,
					title: 'New Message',
					message: broadcast.message,
					timestamp: broadcast.created_at,
					read: false, // Will be overridden by stored state
					data: {
						broadcastId: broadcast.id,
						audience: broadcast.audience
					}
				}));

				// Store new messages in IndexedDB (preserves read state)
				await messageStorage.storeMessages(apiNotifications);

				// Get all messages from IndexedDB (with persisted read state)
				const storedMessages = await messageStorage.getMessages();

				// Convert stored messages back to UserNotification format
				const notifications: UserNotification[] = storedMessages.map((stored) => ({
					id: stored.id,
					type: 'broadcast' as const,
					title: stored.title,
					message: stored.message,
					timestamp: stored.timestamp,
					read: stored.read,
					data: {
						broadcastId: stored.id,
						audience: stored.audience
					}
				}));

				// Get unread count from IndexedDB
				const unreadCount = await messageStorage.getUnreadCount();

				update((state) => ({
					...state,
					notifications,
					unreadCount,
					isLoading: false,
					lastFetched: new Date().toISOString()
				}));

				// Clean up old messages (keep last 30 days)
				await messageStorage.clearOldMessages(30);
			} catch (error) {
				console.error('Failed to fetch notifications:', error);
				update((state) => ({ ...state, isLoading: false }));
			}
		},

		// Mark notification as read
		async markAsRead(notificationId: number) {
			try {
				// Update in IndexedDB
				await messageStorage.markAsRead(notificationId);

				// Update in memory state
				const unreadCount = await messageStorage.getUnreadCount();
				update((state) => ({
					...state,
					notifications: state.notifications.map((n) =>
						n.id === notificationId ? { ...n, read: true } : n
					),
					unreadCount
				}));
			} catch (error) {
				console.error('Failed to mark notification as read:', error);
			}
		},

		// Mark all as read
		async markAllAsRead() {
			try {
				// Update in IndexedDB
				await messageStorage.markAllAsRead();

				// Update in memory state
				update((state) => ({
					...state,
					notifications: state.notifications.map((n) => ({ ...n, read: true })),
					unreadCount: 0
				}));
			} catch (error) {
				console.error('Failed to mark all notifications as read:', error);
			}
		},

		// Add new notification (e.g., from push notification)
		addNotification(notification: Omit<UserNotification, 'id'>) {
			update((state) => {
				const newNotification: UserNotification = {
					...notification,
					id: Date.now() // Simple ID generation
				};

				return {
					...state,
					notifications: [newNotification, ...state.notifications],
					unreadCount: notification.read ? state.unreadCount : state.unreadCount + 1
				};
			});
		},

		// Clear all notifications
		clear() {
			set(initialState);
		},

		// Debug utilities
		async getDebugInfo() {
			try {
				const stats = await messageStorage.getStats();
				const currentState = get({ subscribe });

				return {
					database: stats,
					memory: {
						notifications: currentState.notifications.length,
						unreadCount: currentState.unreadCount,
						lastFetched: currentState.lastFetched,
						isLoading: currentState.isLoading
					},
					performance: {
						memoryVsDatabase: {
							memory: currentState.notifications.length,
							database: stats.total,
							synced: currentState.notifications.length === stats.total
						}
					}
				};
			} catch (error) {
				console.error('Failed to get debug info:', error);
				return null;
			}
		}
	};
};

// Export the store
export const notificationStore = createNotificationStore();

// Derived stores for convenience
export const unreadCount = derived(notificationStore, ($store) => $store.unreadCount);
export const hasUnread = derived(unreadCount, ($count) => $count > 0);
