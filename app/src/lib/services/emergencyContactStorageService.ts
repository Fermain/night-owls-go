import Dexie, { type EntityTable } from 'dexie';

export interface EmergencyContact {
	id: number;
	name: string;
	number: string;
	description: string;
	isDefault: boolean;
	displayOrder: number;
	lastUpdated: string;
}

class EmergencyContactDatabase extends Dexie {
	emergencyContacts!: EntityTable<EmergencyContact, 'id'>;

	constructor() {
		super('NightOwlsEmergencyContacts');

		this.version(1).stores({
			emergencyContacts: 'id, displayOrder, isDefault, lastUpdated'
		});
	}
}

class EmergencyContactStorageService {
	private db = new EmergencyContactDatabase();

	/**
	 * Store emergency contacts in local storage
	 */
	async storeContacts(contacts: EmergencyContact[]): Promise<void> {
		try {
			const contactsToStore = contacts.map((contact) => ({
				...contact,
				lastUpdated: new Date().toISOString()
			}));

			await this.db.emergencyContacts.clear();
			await this.db.emergencyContacts.bulkAdd(contactsToStore);

			console.log('📞 Emergency contacts cached for offline access:', contacts.length);
		} catch (error) {
			console.error('Failed to store emergency contacts:', error);
			throw error;
		}
	}

	/**
	 * Get all emergency contacts from local storage
	 */
	async getContacts(): Promise<EmergencyContact[]> {
		try {
			const contacts = await this.db.emergencyContacts.orderBy('displayOrder').toArray();
			return contacts;
		} catch (error) {
			console.error('Failed to get emergency contacts from storage:', error);
			return [];
		}
	}

	/**
	 * Get the default emergency contact
	 */
	async getDefaultContact(): Promise<EmergencyContact | null> {
		try {
			const contact = await this.db.emergencyContacts
				.filter((contact) => contact.isDefault === true)
				.first();
			return contact || null;
		} catch (error) {
			console.error('Failed to get default emergency contact:', error);
			return null;
		}
	}

	/**
	 * Check if emergency contacts are available offline
	 */
	async hasContacts(): Promise<boolean> {
		try {
			const count = await this.db.emergencyContacts.count();
			return count > 0;
		} catch (error) {
			console.error('Failed to check emergency contacts availability:', error);
			return false;
		}
	}

	/**
	 * Get cache freshness info
	 */
	async getCacheInfo(): Promise<{ lastUpdated: string | null; count: number }> {
		try {
			const contacts = await this.db.emergencyContacts
				.orderBy('lastUpdated')
				.reverse()
				.limit(1)
				.toArray();
			const count = await this.db.emergencyContacts.count();

			return {
				lastUpdated: contacts[0]?.lastUpdated || null,
				count
			};
		} catch (error) {
			console.error('Failed to get cache info:', error);
			return { lastUpdated: null, count: 0 };
		}
	}

	/**
	 * Clear all stored emergency contacts
	 */
	async clearContacts(): Promise<void> {
		try {
			await this.db.emergencyContacts.clear();
			console.log('🗑️ Emergency contacts cache cleared');
		} catch (error) {
			console.error('Failed to clear emergency contacts:', error);
			throw error;
		}
	}

	/**
	 * Initialize database
	 */
	async init(): Promise<void> {
		try {
			await this.db.open();
			console.log('📞 Emergency contacts database initialized');
		} catch (error) {
			console.error('Failed to initialize emergency contacts database:', error);
			throw error;
		}
	}
}

export const emergencyContactStorage = new EmergencyContactStorageService();
