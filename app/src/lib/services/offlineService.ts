import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { emergencyContactStorage, type EmergencyContact } from './emergencyContactStorageService';
import { incidentReportQueue, type QueuedIncidentReport } from './incidentReportQueueService';
import { messageStorage } from './messageStorageService';

// Re-export types for convenience
export type { EmergencyContact, QueuedIncidentReport };

interface OfflineState {
	isOnline: boolean;
	lastOnline: string | null;
	emergencyContactsAvailable: boolean;
	queuedReports: number;
	syncInProgress: boolean;
	lastSync: string | null;
}

interface EmergencyContactApiResponse {
	id: number;
	name: string;
	number: string;
	description: string;
	is_default: boolean;
	display_order: number;
}

interface SyncStatus {
	emergencyContacts: boolean;
	reports: boolean;
	messages: boolean;
}

class OfflineService {
	private isInitialized = false;
	private syncInterval: number | null = null;

	// Store for offline state
	public state = writable<OfflineState>({
		isOnline: browser ? navigator.onLine : true,
		lastOnline: null,
		emergencyContactsAvailable: false,
		queuedReports: 0,
		syncInProgress: false,
		lastSync: null
	});

	/**
	 * Initialize the offline service
	 */
	async init(): Promise<void> {
		if (!browser || this.isInitialized) return;

		try {
			// Initialize storage services
			await Promise.all([
				emergencyContactStorage.init(),
				incidentReportQueue.init(),
				messageStorage.init()
			]);

			// Set up network monitoring
			this.setupNetworkMonitoring();

			// Check initial offline capabilities
			await this.checkOfflineCapabilities();

			// Set up periodic sync
			this.setupPeriodicSync();

			this.isInitialized = true;
			console.log('🌐 Offline service initialized successfully');
		} catch (error) {
			console.error('Failed to initialize offline service:', error);
			throw error;
		}
	}

	/**
	 * Set up network status monitoring
	 */
	private setupNetworkMonitoring(): void {
		if (!browser) return;

		const updateOnlineStatus = () => {
			const isOnline = navigator.onLine;
			
			this.state.update(state => ({
				...state,
				isOnline,
				lastOnline: isOnline ? new Date().toISOString() : state.lastOnline
			}));

			if (isOnline) {
				console.log('🌐 Network connection restored');
				this.handleOnlineEvent();
			} else {
				console.log('📵 Network connection lost');
				this.handleOfflineEvent();
			}
		};

		window.addEventListener('online', updateOnlineStatus);
		window.addEventListener('offline', updateOnlineStatus);

		// Initial status
		updateOnlineStatus();
	}

	/**
	 * Handle coming back online
	 */
	private async handleOnlineEvent(): Promise<void> {
		try {
			// Trigger immediate sync
			await this.syncAllData();
		} catch (error) {
			console.error('Failed to sync data after coming online:', error);
		}
	}

	/**
	 * Handle going offline
	 */
	private handleOfflineEvent(): void {
		console.log('📱 App is now in offline mode');
		// Could show offline notification here
	}

	/**
	 * Check offline capabilities
	 */
	async checkOfflineCapabilities(): Promise<void> {
		try {
			const [hasEmergencyContacts, queueStats] = await Promise.all([
				emergencyContactStorage.hasContacts(),
				incidentReportQueue.getQueueStats()
			]);

			this.state.update(state => ({
				...state,
				emergencyContactsAvailable: hasEmergencyContacts,
				queuedReports: queueStats.drafts + queueStats.queued + queueStats.failed
			}));
		} catch (error) {
			console.error('Failed to check offline capabilities:', error);
		}
	}

	/**
	 * Cache emergency contacts for offline use
	 */
	async cacheEmergencyContacts(): Promise<boolean> {
		try {
			const response = await fetch('/api/emergency-contacts');
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}

			const contactsData: EmergencyContactApiResponse[] = await response.json();
			const contacts: EmergencyContact[] = contactsData.map((contact) => ({
				id: contact.id,
				name: contact.name,
				number: contact.number,
				description: contact.description || '',
				isDefault: contact.is_default,
				displayOrder: contact.display_order,
				lastUpdated: new Date().toISOString()
			}));

			await emergencyContactStorage.storeContacts(contacts);
			
			this.state.update(state => ({
				...state,
				emergencyContactsAvailable: true
			}));

			return true;
		} catch (error) {
			console.error('Failed to cache emergency contacts:', error);
			return false;
		}
	}

	/**
	 * Create an incident report (works offline)
	 */
	async createIncidentReport(report: {
		severity: number;
		message: string;
		latitude?: number;
		longitude?: number;
		accuracy?: number;
		locationTimestamp?: string;
		bookingId?: number;
		isOffShift: boolean;
	}): Promise<string> {
		try {
			// Create as draft initially
			const reportId = await incidentReportQueue.createDraft(report);

			// Try to sync immediately if online
			if (navigator.onLine) {
				await this.syncReport(reportId);
			} else {
				// Queue for later sync
				await incidentReportQueue.queueForSync(reportId);
			}

			// Update queue count
			await this.checkOfflineCapabilities();

			return reportId;
		} catch (error) {
			console.error('Failed to create incident report:', error);
			throw error;
		}
	}

	/**
	 * Sync a single report
	 */
	private async syncReport(reportId: string): Promise<boolean> {
		try {
			const reports = await incidentReportQueue.getQueuedReports();
			const report = reports.find(r => r.id === reportId);
			
			if (!report) {
				console.warn('Report not found for sync:', reportId);
				return false;
			}

			await incidentReportQueue.markSyncing(reportId);

			const payload = {
				severity: report.severity,
				message: report.message,
				...(report.latitude && { latitude: report.latitude }),
				...(report.longitude && { longitude: report.longitude }),
				...(report.accuracy && { accuracy: report.accuracy }),
				...(report.locationTimestamp && { location_timestamp: report.locationTimestamp })
			};

			let response;
			if (report.isOffShift) {
				response = await fetch('/api/reports/off-shift', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
			} else {
				response = await fetch(`/api/bookings/${report.bookingId}/report`, {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify(payload)
				});
			}

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}

			await incidentReportQueue.markSynced(reportId);
			// Optionally delete synced reports after a delay
			setTimeout(() => incidentReportQueue.deleteReport(reportId), 5000);

			return true;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			await incidentReportQueue.markFailed(reportId, errorMessage);
			return false;
		}
	}

	/**
	 * Sync all pending data
	 */
	async syncAllData(): Promise<SyncStatus> {
		const status: SyncStatus = {
			emergencyContacts: false,
			reports: false,
			messages: false
		};

		this.state.update(state => ({ ...state, syncInProgress: true }));

		try {
			// Sync emergency contacts
			status.emergencyContacts = await this.cacheEmergencyContacts();

			// Sync queued reports
			const queuedReports = await incidentReportQueue.getQueuedReports();
			let syncedCount = 0;
			
			for (const report of queuedReports) {
				const synced = await this.syncReport(report.id);
				if (synced) syncedCount++;
			}
			
			status.reports = syncedCount === queuedReports.length;

			// Messages are handled by notificationService separately
			status.messages = true;

			this.state.update(state => ({
				...state,
				syncInProgress: false,
				lastSync: new Date().toISOString()
			}));

			// Update capabilities after sync
			await this.checkOfflineCapabilities();

		} catch (error) {
			console.error('Failed to sync all data:', error);
			this.state.update(state => ({ ...state, syncInProgress: false }));
		}

		return status;
	}

	/**
	 * Get emergency contacts (tries cache first)
	 */
	async getEmergencyContacts(): Promise<EmergencyContact[]> {
		try {
			// Try cache first
			const cachedContacts = await emergencyContactStorage.getContacts();
			
			if (cachedContacts.length > 0) {
				return cachedContacts;
			}

			// If no cache and online, fetch and cache
			if (navigator.onLine) {
				await this.cacheEmergencyContacts();
				return await emergencyContactStorage.getContacts();
			}

			// No cache and offline
			return [];
		} catch (error) {
			console.error('Failed to get emergency contacts:', error);
			return [];
		}
	}

	/**
	 * Set up periodic sync when online
	 */
	private setupPeriodicSync(): void {
		// Sync every 5 minutes when online
		this.syncInterval = window.setInterval(async () => {
			if (navigator.onLine) {
				await this.syncAllData();
			}
		}, 5 * 60 * 1000);
	}

	/**
	 * Get pending reports for UI display
	 */
	async getPendingReports(): Promise<QueuedIncidentReport[]> {
		return await incidentReportQueue.getPendingReports();
	}

	/**
	 * Cleanup and destroy service
	 */
	destroy(): void {
		if (this.syncInterval) {
			clearInterval(this.syncInterval);
			this.syncInterval = null;
		}
		this.isInitialized = false;
	}
}

export const offlineService = new OfflineService(); 