import { writable, derived } from 'svelte/store';
import { authenticatedFetch } from '$lib/utils/api';

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
		
		// Actions
		async fetchNotifications(force = false) {
			update(state => ({ ...state, isLoading: true }));
			
			try {
				// Fetch broadcasts (main source of notifications for now)
				const response = await authenticatedFetch('/api/broadcasts');
				
				if (!response.ok) {
					throw new Error(`HTTP ${response.status}: ${response.statusText}`);
				}
				
				const broadcasts: BroadcastResponse[] = await response.json();
				
				// Transform broadcasts into notifications
				const notifications: UserNotification[] = broadcasts.map((broadcast) => ({
					id: broadcast.id,
					type: 'broadcast' as const,
					title: 'New Message',
					message: broadcast.message,
					timestamp: broadcast.created_at,
					read: false, // TODO: Track read status per user
					data: {
						broadcastId: broadcast.id,
						audience: broadcast.audience
					}
				}));

				// Sort by timestamp (newest first)
				notifications.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

				update(state => ({
					...state,
					notifications,
					unreadCount: notifications.filter(n => !n.read).length,
					isLoading: false,
					lastFetched: new Date().toISOString()
				}));

			} catch (error) {
				console.error('Failed to fetch notifications:', error);
				update(state => ({ ...state, isLoading: false }));
			}
		},

		// Mark notification as read
		markAsRead(notificationId: number) {
			update(state => ({
				...state,
				notifications: state.notifications.map(n => 
					n.id === notificationId ? { ...n, read: true } : n
				),
				unreadCount: Math.max(0, state.unreadCount - 1)
			}));
		},

		// Mark all as read
		markAllAsRead() {
			update(state => ({
				...state,
				notifications: state.notifications.map(n => ({ ...n, read: true })),
				unreadCount: 0
			}));
		},

		// Add new notification (e.g., from push notification)
		addNotification(notification: Omit<UserNotification, 'id'>) {
			update(state => {
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
		}
	};
};

// Export the store
export const notificationStore = createNotificationStore();

// Derived stores for convenience
export const unreadCount = derived(notificationStore, $store => $store.unreadCount);
export const hasUnread = derived(unreadCount, $count => $count > 0); 