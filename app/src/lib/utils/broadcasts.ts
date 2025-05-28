import {
	formatDistanceToNow,
	parseISO,
	isValid,
	differenceInHours,
	differenceInDays
} from 'date-fns';

/**
 * Audience options for broadcast targeting
 */
export const AUDIENCE_OPTIONS = [
	{ value: 'all', label: 'All Users' },
	{ value: 'admins', label: 'Admins Only' },
	{ value: 'owls', label: 'Owls Only' },
	{ value: 'active', label: 'Active Users (last 30 days)' }
] as const;

/**
 * Broadcast status types
 */
export const BROADCAST_STATUS = {
	SENT: 'sent',
	PENDING: 'pending',
	SENDING: 'sending',
	FAILED: 'failed'
} as const;

/**
 * Get relative time description for a broadcast using date-fns
 * Uses a more concise format similar to the original implementation
 */
export function formatRelativeTime(dateString: string): string {
	try {
		const date = parseISO(dateString);

		if (!isValid(date)) {
			return 'Invalid Date';
		}

		const now = new Date();
		const hoursAgo = differenceInHours(now, date);

		if (hoursAgo < 1) return 'Just now';
		if (hoursAgo < 24) return `${hoursAgo}h ago`;

		const daysAgo = differenceInDays(now, date);
		if (daysAgo === 1) return 'Yesterday';
		if (daysAgo < 7) return `${daysAgo}d ago`;

		// For older broadcasts, use the full distance format
		return formatDistanceToNow(date, { addSuffix: true });
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Get detailed relative time description (alternative format)
 */
export function formatDetailedRelativeTime(dateString: string): string {
	try {
		const date = parseISO(dateString);

		if (!isValid(date)) {
			return 'Invalid Date';
		}

		return formatDistanceToNow(date, { addSuffix: true });
	} catch {
		return 'Invalid Date';
	}
}

/**
 * Get audience label from audience value
 */
export function getAudienceLabel(audience: string): string {
	return AUDIENCE_OPTIONS.find((opt) => opt.value === audience)?.label ?? 'Unknown';
}

/**
 * Get status styling for broadcast status badges
 */
export function getBroadcastStatusStyle(status: string): string {
	switch (status) {
		case BROADCAST_STATUS.SENT:
			return 'bg-green-100 text-green-800';
		case BROADCAST_STATUS.PENDING:
			return 'bg-yellow-100 text-yellow-800';
		case BROADCAST_STATUS.SENDING:
			return 'bg-blue-100 text-blue-800';
		case BROADCAST_STATUS.FAILED:
		default:
			return 'bg-red-100 text-red-800';
	}
}

/**
 * Check if broadcast has delivery issues
 */
export function hasDeliveryIssues(broadcast: {
	status: string;
	sent_count: number;
	recipient_count: number;
}): boolean {
	return (
		broadcast.status === BROADCAST_STATUS.SENT && broadcast.sent_count !== broadcast.recipient_count
	);
}

/**
 * Get delivery status text
 */
export function getDeliveryStatusText(broadcast: {
	sent_count: number;
	recipient_count: number;
}): string {
	return `${broadcast.sent_count}/${broadcast.recipient_count} delivered`;
}
