import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
import InfoIcon from '@lucide/svelte/icons/info';
import { differenceInHours, differenceInDays, parseISO, isValid } from 'date-fns';

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

/**
 * Format a date string as relative time (e.g., "2h ago", "Yesterday")
 */
export function formatRelativeTime(dateString: string): string {
	try {
		const date = parseISO(dateString);
		if (!isValid(date)) {
			return 'Unknown time';
		}

		const now = new Date();
		const diffInHours = differenceInHours(now, date);
		const diffInDays = differenceInDays(now, date);

		if (diffInHours < 1) return 'Just now';
		if (diffInHours < 24) return `${diffInHours}h ago`;
		if (diffInDays === 1) return 'Yesterday';
		if (diffInDays < 7) return `${diffInDays}d ago`;
		
		// For older dates, show the actual date
		return date.toLocaleDateString('en-ZA', {
			month: 'short',
			day: 'numeric',
			year: diffInDays > 365 ? 'numeric' : undefined
		});
	} catch (error) {
		console.error('Error formatting relative time:', error);
		return 'Unknown time';
	}
}
