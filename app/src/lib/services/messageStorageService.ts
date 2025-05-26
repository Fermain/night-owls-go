import Dexie, { type EntityTable } from 'dexie';
import type { UserNotification } from './notificationService';

interface StoredMessage {
	id: number;
	message: string;
	title: string;
	timestamp: string;
	audience: string;
	read: boolean;
	readAt?: string;
	lastSeen: string;
}

class MessageDatabase extends Dexie {
	messages!: EntityTable<StoredMessage, 'id'>;

	constructor() {
		super('NightOwlsMessages');
		
		this.version(1).stores({
			messages: 'id, timestamp, read, audience, lastSeen'
		});
	}
}

class MessageStorageService {
	private db = new MessageDatabase();

	async storeMessages(notifications: UserNotification[]): Promise<void> {
		try {
			// Use Dexie's bulkPut with conflict resolution
			const messagesToStore: StoredMessage[] = [];
			
			for (const notification of notifications) {
				// Check if message already exists to preserve read state
				const existing = await this.db.messages.get(notification.id);
				
				const storedMessage: StoredMessage = {
					id: notification.id,
					message: notification.message,
					title: notification.title,
					timestamp: notification.timestamp,
					audience: notification.data?.audience || 'all',
					read: existing?.read || false, // Preserve existing read state
					readAt: existing?.readAt,
					lastSeen: new Date().toISOString()
				};
				
				messagesToStore.push(storedMessage);
			}
			
			await this.db.messages.bulkPut(messagesToStore);
		} catch (error) {
			console.error('Failed to store messages:', error);
			throw error;
		}
	}

	async getMessages(): Promise<StoredMessage[]> {
		try {
			// Get all messages ordered by timestamp (newest first)
			return await this.db.messages
				.orderBy('timestamp')
				.reverse()
				.toArray();
		} catch (error) {
			console.error('Failed to get messages:', error);
			return [];
		}
	}

	async markAsRead(messageId: number): Promise<void> {
		try {
			await this.db.messages.update(messageId, {
				read: true,
				readAt: new Date().toISOString()
			});
		} catch (error) {
			console.error('Failed to mark message as read:', error);
			throw error;
		}
	}

	async markAllAsRead(): Promise<void> {
		try {
			const now = new Date().toISOString();
			await this.db.messages
				.filter(msg => !msg.read)
				.modify({
					read: true,
					readAt: now
				});
		} catch (error) {
			console.error('Failed to mark all messages as read:', error);
			throw error;
		}
	}

	async getUnreadCount(): Promise<number> {
		try {
			return await this.db.messages
				.filter(msg => !msg.read)
				.count();
		} catch (error) {
			console.error('Failed to get unread count:', error);
			return 0;
		}
	}

	async clearOldMessages(daysToKeep: number = 30): Promise<void> {
		try {
			const cutoffDate = new Date();
			cutoffDate.setDate(cutoffDate.getDate() - daysToKeep);
			const cutoffTimestamp = cutoffDate.toISOString();

			await this.db.messages
				.where('timestamp')
				.below(cutoffTimestamp)
				.delete();
		} catch (error) {
			console.error('Failed to clear old messages:', error);
			// Don't throw - this is a cleanup operation
		}
	}

	async getMessagesByAudience(audience: string): Promise<StoredMessage[]> {
		try {
			return await this.db.messages
				.where('audience')
				.equals(audience)
				.reverse()
				.sortBy('timestamp');
		} catch (error) {
			console.error('Failed to get messages by audience:', error);
			return [];
		}
	}

	async searchMessages(query: string): Promise<StoredMessage[]> {
		try {
			const lowerQuery = query.toLowerCase();
			return await this.db.messages
				.filter(message => 
					message.message.toLowerCase().includes(lowerQuery) ||
					message.title.toLowerCase().includes(lowerQuery)
				)
				.reverse()
				.sortBy('timestamp');
		} catch (error) {
			console.error('Failed to search messages:', error);
			return [];
		}
	}

	async getStats(): Promise<{
		total: number;
		unread: number;
		byAudience: Record<string, number>;
		oldestMessage?: string;
		newestMessage?: string;
	}> {
		try {
			const [total, unread, allMessages] = await Promise.all([
				this.db.messages.count(),
				this.getUnreadCount(),
				this.db.messages.orderBy('timestamp').toArray()
			]);

			// Group by audience
			const byAudience: Record<string, number> = {};
			allMessages.forEach(msg => {
				byAudience[msg.audience] = (byAudience[msg.audience] || 0) + 1;
			});

			return {
				total,
				unread,
				byAudience,
				oldestMessage: allMessages[0]?.timestamp,
				newestMessage: allMessages[allMessages.length - 1]?.timestamp
			};
		} catch (error) {
			console.error('Failed to get message stats:', error);
			return { total: 0, unread: 0, byAudience: {} };
		}
	}

	// Initialize database (Dexie handles this automatically, but we can add custom logic)
	async init(): Promise<void> {
		try {
			await this.db.open();
			console.log('📦 Message database initialized successfully');
		} catch (error) {
			console.error('Failed to initialize message database:', error);
			throw error;
		}
	}

	// Close database connection
	async close(): Promise<void> {
		try {
			await this.db.close();
		} catch (error) {
			console.error('Failed to close message database:', error);
		}
	}
}

export const messageStorage = new MessageStorageService(); 