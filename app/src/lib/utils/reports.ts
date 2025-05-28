import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
import InfoIcon from '@lucide/svelte/icons/info';

/**
 * Report severity levels
 */
export const REPORT_SEVERITY = {
	NORMAL: 0,
	SUSPICION: 1,
	INCIDENT: 2
} as const;

export type ReportSeverity = (typeof REPORT_SEVERITY)[keyof typeof REPORT_SEVERITY];

/**
 * Get the appropriate icon component for a report severity level
 */
export function getSeverityIcon(severity: number) {
	switch (severity) {
		case REPORT_SEVERITY.NORMAL:
			return InfoIcon;
		case REPORT_SEVERITY.SUSPICION:
			return AlertTriangleIcon;
		case REPORT_SEVERITY.INCIDENT:
			return ShieldAlertIcon;
		default:
			return InfoIcon;
	}
}

/**
 * Get the appropriate color class for a report severity level
 */
export function getSeverityColor(severity: number): string {
	switch (severity) {
		case REPORT_SEVERITY.NORMAL:
			return 'text-blue-600';
		case REPORT_SEVERITY.SUSPICION:
			return 'text-orange-600';
		case REPORT_SEVERITY.INCIDENT:
			return 'text-red-600';
		default:
			return 'text-gray-600';
	}
}

/**
 * Get the human-readable label for a report severity level
 */
export function getSeverityLabel(severity: number): string {
	switch (severity) {
		case REPORT_SEVERITY.NORMAL:
			return 'Normal';
		case REPORT_SEVERITY.SUSPICION:
			return 'Suspicion';
		case REPORT_SEVERITY.INCIDENT:
			return 'Incident';
		default:
			return 'Unknown';
	}
}

/**
 * Get the background color class for a report severity level
 */
export function getSeverityBgColor(severity: number): string {
	switch (severity) {
		case REPORT_SEVERITY.NORMAL:
			return 'bg-blue-100';
		case REPORT_SEVERITY.SUSPICION:
			return 'bg-orange-100';
		case REPORT_SEVERITY.INCIDENT:
			return 'bg-red-100';
		default:
			return 'bg-gray-100';
	}
}
