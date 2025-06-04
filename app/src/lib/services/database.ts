/**
 * Database Service
 * Handles database health checks, migration management, and recovery operations
 * Uses our domain types and error classification system
 */

import type { DatabaseHealth, MigrationState, MigrationResult } from '$lib/types/domain';
import {
	ErrorType,
	createDatabaseDirtyError,
	createMigrationFailedError,
	classifyError
} from '$lib/utils/errors';
import { apiGet, apiPost } from '$lib/utils/api';

export class DatabaseService {
	private baseUrl: string;

	constructor(baseUrl = '/api') {
		this.baseUrl = baseUrl;
	}

	/**
	 * Check overall database health
	 */
	async checkDatabaseHealth(): Promise<DatabaseHealth> {
		try {
			const response = await apiGet<{
				status: string;
				database: string;
				uptime?: string;
				version?: string;
				error?: string;
			}>(`${this.baseUrl}/health`);

			const migrationState = await this.getMigrationState().catch(() => null);

			return {
				status: response.database === 'up' ? 'healthy' : 'unhealthy',
				connectionStatus: response.database === 'up' ? 'connected' : 'error',
				migrationStatus: migrationState?.dirty
					? 'dirty'
					: migrationState
						? 'up-to-date'
						: 'pending',
				lastChecked: new Date().toISOString(),
				version: migrationState?.version,
				error: response.error
			};
		} catch (error) {
			const appError = classifyError(error);
			return {
				status: 'unhealthy',
				connectionStatus: 'error',
				migrationStatus: 'failed',
				lastChecked: new Date().toISOString(),
				error: appError.message
			};
		}
	}

	/**
	 * Get current migration state
	 */
	async getMigrationState(): Promise<MigrationState | null> {
		try {
			const response = await apiGet<{
				version: number;
				dirty: boolean;
				description?: string;
				applied_at?: string;
			}>(`${this.baseUrl}/admin/database/migration-state`);

			return {
				version: response.version,
				dirty: response.dirty,
				description: response.description,
				appliedAt: response.applied_at
			};
		} catch (error) {
			// If endpoint doesn't exist or fails, return null
			console.warn('Migration state endpoint not available:', error);
			return null;
		}
	}

	/**
	 * Handle dirty database state with recovery options
	 */
	async handleDirtyDatabase(version?: number): Promise<MigrationResult> {
		try {
			// First, attempt to get current state
			const currentState = await this.getMigrationState();

			if (!currentState?.dirty) {
				return {
					success: true,
					currentVersion: currentState?.version,
					migrationsApplied: 0
				};
			}

			// Attempt to force clean the migration state
			const response = await apiPost<
				{ version?: number; force: boolean },
				{ success: boolean; version: number; message: string }
			>(`${this.baseUrl}/admin/database/force-migration-version`, {
				version: version || currentState.version,
				force: true
			});

			if (response.success) {
				return {
					success: true,
					previousVersion: currentState.version,
					currentVersion: response.version,
					migrationsApplied: 0 // No migrations applied, just cleaned state
				};
			} else {
				throw createDatabaseDirtyError(`Failed to force migration version: ${response.message}`);
			}
		} catch (error) {
			const appError = classifyError(error);

			return {
				success: false,
				migrationsApplied: 0,
				error: appError.message,
				requiresManualIntervention: true,
				recoveryInstructions: [
					'Check database migration logs for specific error details',
					'Verify database backup is available before proceeding',
					'Consider running: `make migrate-force VERSION=N` where N is the target version',
					'If migrations are corrupted, restore from backup and reapply',
					'Contact system administrator if issue persists'
				]
			};
		}
	}

	/**
	 * Run database migrations with proper error handling
	 */
	async runMigrations(): Promise<MigrationResult> {
		try {
			const response = await apiPost<
				Record<string, never>,
				{
					success: boolean;
					migrations_applied: number;
					current_version: number;
					previous_version?: number;
					error?: string;
				}
			>(`${this.baseUrl}/admin/database/migrate`, {});

			if (response.success) {
				return {
					success: true,
					previousVersion: response.previous_version,
					currentVersion: response.current_version,
					migrationsApplied: response.migrations_applied
				};
			} else {
				throw createMigrationFailedError(response.error || 'Migration failed');
			}
		} catch (error) {
			const appError = classifyError(error);

			// Check if it's a dirty database error
			if (appError.type === ErrorType.DATABASE_DIRTY) {
				return {
					success: false,
					migrationsApplied: 0,
					error: appError.message,
					requiresManualIntervention: true,
					recoveryInstructions: [
						'Database is in dirty state - migration was interrupted',
						'Run database recovery: handleDirtyDatabase()',
						'Check logs for the specific migration that failed',
						'Verify database integrity before proceeding'
					]
				};
			}

			return {
				success: false,
				migrationsApplied: 0,
				error: appError.message,
				requiresManualIntervention: appError.type !== ErrorType.DATABASE_CONNECTION
			};
		}
	}

	/**
	 * Test database connection
	 */
	async testConnection(): Promise<boolean> {
		try {
			const health = await this.checkDatabaseHealth();
			return health.connectionStatus === 'connected';
		} catch (error) {
			return false;
		}
	}

	/**
	 * Get database recovery instructions based on current state
	 */
	async getRecoveryInstructions(): Promise<string[]> {
		try {
			const health = await this.checkDatabaseHealth();
			const instructions: string[] = [];

			if (health.connectionStatus === 'error') {
				instructions.push(
					'Database connection failed',
					'Check if database server is running',
					'Verify database connection string',
					'Check database file permissions (for SQLite)'
				);
			}

			if (health.migrationStatus === 'dirty') {
				instructions.push(
					'Database migration state is dirty',
					'A previous migration was interrupted',
					'Use handleDirtyDatabase() to attempt recovery',
					'Consider restoring from backup if recovery fails'
				);
			}

			if (health.migrationStatus === 'pending') {
				instructions.push(
					'Database migrations are pending',
					'Run runMigrations() to apply pending migrations',
					'Ensure application is stopped during migration',
					'Backup database before applying migrations'
				);
			}

			return instructions;
		} catch (error) {
			return [
				'Unable to determine database state',
				'Check database connectivity',
				'Verify application configuration',
				'Contact system administrator'
			];
		}
	}
}

// Export singleton instance
export const databaseService = new DatabaseService();
