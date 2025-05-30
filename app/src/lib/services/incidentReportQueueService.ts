import Dexie, { type EntityTable } from 'dexie';

export interface QueuedIncidentReport {
	id: string; // UUID for offline reports
	severity: number;
	message: string;
	latitude?: number;
	longitude?: number;
	accuracy?: number;
	locationTimestamp?: string;
	bookingId?: number; // For on-shift reports
	isOffShift: boolean;
	status: 'draft' | 'queued' | 'syncing' | 'synced' | 'failed';
	createdAt: string;
	lastAttempt?: string;
	syncAttempts: number;
	error?: string;
}

class IncidentReportQueueDatabase extends Dexie {
	reports!: EntityTable<QueuedIncidentReport, 'id'>;

	constructor() {
		super('NightOwlsIncidentReports');

		this.version(1).stores({
			reports: 'id, status, createdAt, isOffShift, severity'
		});
	}
}

class IncidentReportQueueService {
	private db = new IncidentReportQueueDatabase();

	/**
	 * Create a new draft report (for offline creation)
	 */
	async createDraft(report: Omit<QueuedIncidentReport, 'id' | 'status' | 'createdAt' | 'syncAttempts'>): Promise<string> {
		try {
			const reportId = crypto.randomUUID();
			const draftReport: QueuedIncidentReport = {
				...report,
				id: reportId,
				status: 'draft',
				createdAt: new Date().toISOString(),
				syncAttempts: 0
			};

			await this.db.reports.add(draftReport);
			console.log('📝 Incident report draft created:', reportId);
			return reportId;
		} catch (error) {
			console.error('Failed to create incident report draft:', error);
			throw error;
		}
	}

	/**
	 * Queue a report for syncing when online
	 */
	async queueForSync(reportId: string): Promise<void> {
		try {
			await this.db.reports.update(reportId, {
				status: 'queued',
				lastAttempt: new Date().toISOString()
			});
			console.log('📤 Incident report queued for sync:', reportId);
		} catch (error) {
			console.error('Failed to queue incident report:', error);
			throw error;
		}
	}

	/**
	 * Get all queued reports ready for syncing
	 */
	async getQueuedReports(): Promise<QueuedIncidentReport[]> {
		try {
			return await this.db.reports
				.where('status')
				.equals('queued')
				.sortBy('createdAt');
		} catch (error) {
			console.error('Failed to get queued reports:', error);
			return [];
		}
	}

	/**
	 * Get all draft reports
	 */
	async getDraftReports(): Promise<QueuedIncidentReport[]> {
		try {
			return await this.db.reports
				.where('status')
				.equals('draft')
				.sortBy('createdAt');
		} catch (error) {
			console.error('Failed to get draft reports:', error);
			return [];
		}
	}

	/**
	 * Get all pending reports (drafts + queued)
	 */
	async getPendingReports(): Promise<QueuedIncidentReport[]> {
		try {
			return await this.db.reports
				.where('status')
				.anyOf(['draft', 'queued', 'failed'])
				.sortBy('createdAt');
		} catch (error) {
			console.error('Failed to get pending reports:', error);
			return [];
		}
	}

	/**
	 * Mark report as syncing
	 */
	async markSyncing(reportId: string): Promise<void> {
		try {
			await this.db.reports.update(reportId, {
				status: 'syncing',
				lastAttempt: new Date().toISOString()
			});
		} catch (error) {
			console.error('Failed to mark report as syncing:', error);
			throw error;
		}
	}

	/**
	 * Mark report as successfully synced
	 */
	async markSynced(reportId: string): Promise<void> {
		try {
			await this.db.reports.update(reportId, {
				status: 'synced',
				lastAttempt: new Date().toISOString(),
				error: undefined
			});
			console.log('✅ Incident report synced successfully:', reportId);
		} catch (error) {
			console.error('Failed to mark report as synced:', error);
			throw error;
		}
	}

	/**
	 * Mark report as failed with error
	 */
	async markFailed(reportId: string, error: string): Promise<void> {
		try {
			const report = await this.db.reports.get(reportId);
			const syncAttempts = (report?.syncAttempts || 0) + 1;

			await this.db.reports.update(reportId, {
				status: 'failed',
				lastAttempt: new Date().toISOString(),
				syncAttempts,
				error
			});
			console.warn('❌ Incident report sync failed:', reportId, error);
		} catch (err) {
			console.error('Failed to mark report as failed:', err);
			throw err;
		}
	}

	/**
	 * Delete a report (usually after successful sync)
	 */
	async deleteReport(reportId: string): Promise<void> {
		try {
			await this.db.reports.delete(reportId);
			console.log('🗑️ Incident report deleted:', reportId);
		} catch (error) {
			console.error('Failed to delete incident report:', error);
			throw error;
		}
	}

	/**
	 * Update a draft report
	 */
	async updateDraft(reportId: string, updates: Partial<QueuedIncidentReport>): Promise<void> {
		try {
			await this.db.reports.update(reportId, updates);
			console.log('📝 Incident report draft updated:', reportId);
		} catch (error) {
			console.error('Failed to update incident report draft:', error);
			throw error;
		}
	}

	/**
	 * Get queue statistics
	 */
	async getQueueStats(): Promise<{
		total: number;
		drafts: number;
		queued: number;
		failed: number;
		synced: number;
	}> {
		try {
			const [total, drafts, queued, failed, synced] = await Promise.all([
				this.db.reports.count(),
				this.db.reports.where('status').equals('draft').count(),
				this.db.reports.where('status').equals('queued').count(),
				this.db.reports.where('status').equals('failed').count(),
				this.db.reports.where('status').equals('synced').count()
			]);

			return { total, drafts, queued, failed, synced };
		} catch (error) {
			console.error('Failed to get queue stats:', error);
			return { total: 0, drafts: 0, queued: 0, failed: 0, synced: 0 };
		}
	}

	/**
	 * Clean up old synced reports (keep for 30 days)
	 */
	async cleanupOldReports(daysToKeep: number = 30): Promise<void> {
		try {
			const cutoffDate = new Date();
			cutoffDate.setDate(cutoffDate.getDate() - daysToKeep);
			const cutoffTimestamp = cutoffDate.toISOString();

			const deletedCount = await this.db.reports
				.where('status')
				.equals('synced')
				.and(report => report.createdAt < cutoffTimestamp)
				.delete();

			if (deletedCount > 0) {
				console.log(`🧹 Cleaned up ${deletedCount} old synced reports`);
			}
		} catch (error) {
			console.error('Failed to cleanup old reports:', error);
		}
	}

	/**
	 * Initialize database
	 */
	async init(): Promise<void> {
		try {
			await this.db.open();
			console.log('📋 Incident report queue database initialized');
			
			// Cleanup old reports on init
			await this.cleanupOldReports();
		} catch (error) {
			console.error('Failed to initialize incident report queue database:', error);
			throw error;
		}
	}
}

export const incidentReportQueue = new IncidentReportQueueService(); 