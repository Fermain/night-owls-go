<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import {
		Dialog,
		DialogContent,
		DialogHeader,
		DialogTitle,
		DialogTrigger
	} from '$lib/components/ui/dialog';
	import {
		Shield,
		User,
		UserPlus,
		UserX,
		Settings,
		Eye,
		Clock,
		MapPin,
		Monitor,
		MoreHorizontal
	} from 'lucide-svelte';
	import { formatDistanceToNow } from 'date-fns';

	// Use our domain types
	import type { AuditEvent } from '$lib/types/domain';

	export let events: AuditEvent[];

	// Get icon for event type
	function getEventIcon(eventType: string, action?: string) {
		// Extract action from eventType if not provided separately
		const eventAction = action || eventType.split('.')[1] || '';

		if (eventType.startsWith('user.')) {
			switch (eventAction) {
				case 'login':
					return Shield;
				case 'created':
					return UserPlus;
				case 'updated':
					return Settings;
				case 'role_changed':
					return Shield;
				case 'deleted':
					return UserX;
				case 'bulk_deleted':
					return UserX;
				default:
					return User;
			}
		}
		return Eye;
	}

	// Get color scheme for event type
	function getEventColors(eventType: string, action?: string) {
		// Extract action from eventType if not provided separately
		const eventAction = action || eventType.split('.')[1] || '';

		if (eventType.startsWith('user.')) {
			switch (eventAction) {
				case 'login':
					return {
						bg: 'bg-green-50 dark:bg-green-950/30',
						border: 'border-green-200 dark:border-green-800',
						icon: 'text-green-600 dark:text-green-400',
						badge: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
					};
				case 'created':
					return {
						bg: 'bg-blue-50 dark:bg-blue-950/30',
						border: 'border-blue-200 dark:border-blue-800',
						icon: 'text-blue-600 dark:text-blue-400',
						badge: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300'
					};
				case 'updated':
					return {
						bg: 'bg-yellow-50 dark:bg-yellow-950/30',
						border: 'border-yellow-200 dark:border-yellow-800',
						icon: 'text-yellow-600 dark:text-yellow-400',
						badge: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
					};
				case 'role_changed':
					return {
						bg: 'bg-purple-50 dark:bg-purple-950/30',
						border: 'border-purple-200 dark:border-purple-800',
						icon: 'text-purple-600 dark:text-purple-400',
						badge: 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-300'
					};
				case 'deleted':
				case 'bulk_deleted':
					return {
						bg: 'bg-red-50 dark:bg-red-950/30',
						border: 'border-red-200 dark:border-red-800',
						icon: 'text-red-600 dark:text-red-400',
						badge: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
					};
			}
		}
		return {
			bg: 'bg-gray-50 dark:bg-gray-950/30',
			border: 'border-gray-200 dark:border-gray-800',
			icon: 'text-gray-600 dark:text-gray-400',
			badge: 'bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-300'
		};
	}

	// Format event description
	function getEventDescription(event: AuditEvent): string {
		const actor = event.userName || 'System';
		const target = event.targetUserName;

		switch (event.eventType) {
			case 'user.login':
				return `${actor} logged in`;
			case 'user.created':
				return `${actor} created user ${target}`;
			case 'user.updated':
				return `${actor} updated user ${target}`;
			case 'user.role_changed':
				return `${actor} changed role for ${target}`;
			case 'user.deleted':
				return `${actor} deleted user ${target}`;
			case 'user.bulk_deleted':
				return `${actor} bulk deleted users`;
			default: {
				const action = event.eventType.split('.')[1] || 'performed action';
				const entityType = event.eventType.split('.')[0] || 'entity';
				return `${actor} ${action} on ${entityType}`;
			}
		}
	}

	// Format details for display
	function formatDetails(details: Record<string, unknown> | undefined): string {
		if (!details) return '';

		const parts: string[] = [];

		// Handle role changes
		if (details.old_role && details.new_role) {
			parts.push(`Role: ${details.old_role} → ${details.new_role}`);
		}

		// Handle field changes
		if (details.name && typeof details.name === 'object') {
			const nameObj = details.name as { old: string; new: string };
			parts.push(`Name: "${nameObj.old}" → "${nameObj.new}"`);
		}
		if (details.phone && typeof details.phone === 'object') {
			const phoneObj = details.phone as { old: string; new: string };
			parts.push(`Phone: ${phoneObj.old} → ${phoneObj.new}`);
		}

		// Handle creation details
		if (details.target_user_name) {
			parts.push(`Name: ${details.target_user_name}`);
		}
		if (details.target_user_phone) {
			parts.push(`Phone: ${details.target_user_phone}`);
		}
		if (details.target_role) {
			parts.push(`Role: ${details.target_role}`);
		}

		// Handle bulk operations
		if (details.deleted_count) {
			parts.push(`${details.deleted_count} users deleted`);
		}

		return parts.join(' • ');
	}

	// Parse User-Agent for better display
	function parseUserAgent(userAgent: string): string {
		if (!userAgent || userAgent === 'Unknown') return 'Unknown device';

		// Simple parsing - could be enhanced
		if (userAgent.includes('curl')) return 'cURL/API';
		if (userAgent.includes('Chrome')) return 'Chrome Browser';
		if (userAgent.includes('Firefox')) return 'Firefox Browser';
		if (userAgent.includes('Safari')) return 'Safari Browser';
		if (userAgent.includes('Edge')) return 'Edge Browser';

		return 'Browser';
	}

	// Format time
	function formatTime(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString();
	}

	function formatRelativeTime(dateString: string): string {
		const date = new Date(dateString);
		return formatDistanceToNow(date, { addSuffix: true });
	}
</script>

<div class="space-y-4">
	{#each events as event, index (event.id)}
		{@const colors = getEventColors(event.eventType)}
		{@const IconComponent = getEventIcon(event.eventType)}
		{@const description = getEventDescription(event)}
		{@const detailsText = formatDetails(event.details)}

		<div
			class="relative flex gap-4 p-4 rounded-lg border transition-all hover:shadow-sm {colors.bg} {colors.border}"
		>
			<!-- Timeline Line -->
			{#if index < events.length - 1}
				<div class="absolute left-6 top-12 w-px h-full bg-border"></div>
			{/if}

			<!-- Event Icon -->
			<div
				class="flex-shrink-0 w-8 h-8 rounded-full bg-background border-2 flex items-center justify-center z-10 {colors.border}"
			>
				<svelte:component this={IconComponent} class="h-4 w-4 {colors.icon}" />
			</div>

			<!-- Event Content -->
			<div class="flex-1 min-w-0 space-y-2">
				<!-- Event Header -->
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<h3 class="font-medium text-sm leading-none">
							{description}
						</h3>
						<Badge variant="secondary" class="text-xs {colors.badge}">
							{event.eventType}
						</Badge>
					</div>

					<div class="flex items-center gap-2 text-xs text-muted-foreground">
						<Clock class="h-3 w-3" />
						<span title={formatTime(event.createdAt)}>
							{formatRelativeTime(event.createdAt)}
						</span>
					</div>
				</div>

				<!-- Event Details -->
				{#if detailsText}
					<p class="text-sm text-muted-foreground">
						{detailsText}
					</p>
				{/if}

				<!-- Metadata -->
				<div class="flex items-center gap-4 text-xs text-muted-foreground">
					{#if event.ipAddress}
						<div class="flex items-center gap-1">
							<MapPin class="h-3 w-3" />
							<span>{event.ipAddress}</span>
						</div>
					{/if}
					{#if event.userAgent}
						<div class="flex items-center gap-1">
							<Monitor class="h-3 w-3" />
							<span>{parseUserAgent(event.userAgent)}</span>
						</div>
					{/if}
				</div>

				<!-- View Details Button -->
				{#if (event.details && Object.keys(event.details).length > 0) || (event.userAgent && event.userAgent !== 'Unknown' && event.userAgent !== 'cURL/API')}
					<Dialog>
						<DialogTrigger>
							<button
								class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-6 px-2 text-xs"
							>
								<MoreHorizontal class="h-3 w-3 mr-1" />
								Details
							</button>
						</DialogTrigger>
						<DialogContent class="max-w-2xl">
							<DialogHeader>
								<DialogTitle class="flex items-center gap-2">
									<svelte:component this={IconComponent} class="h-5 w-5 {colors.icon}" />
									Audit Event Details
								</DialogTitle>
							</DialogHeader>

							<div class="space-y-4">
								<!-- Basic Info -->
								<div class="grid grid-cols-2 gap-4">
									<div>
										<h4 class="text-sm font-medium mb-2">Event Information</h4>
										<dl class="space-y-1 text-sm">
											<div class="flex justify-between">
												<dt class="text-muted-foreground">Event ID:</dt>
												<dd class="font-mono">{event.id}</dd>
											</div>
											<div class="flex justify-between">
												<dt class="text-muted-foreground">Type:</dt>
												<dd>
													<Badge variant="secondary" class={colors.badge}>{event.eventType}</Badge>
												</dd>
											</div>
											<div class="flex justify-between">
												<dt class="text-muted-foreground">Timestamp:</dt>
												<dd class="font-mono">{formatTime(event.createdAt)}</dd>
											</div>
										</dl>
									</div>

									<div>
										<h4 class="text-sm font-medium mb-2">User Information</h4>
										<dl class="space-y-1 text-sm">
											{#if event.userName}
												<div class="flex justify-between">
													<dt class="text-muted-foreground">Actor:</dt>
													<dd>{event.userName}</dd>
												</div>
											{/if}
											{#if event.targetUserName}
												<div class="flex justify-between">
													<dt class="text-muted-foreground">Target:</dt>
													<dd>{event.targetUserName}</dd>
												</div>
											{/if}
											{#if event.ipAddress}
												<div class="flex justify-between">
													<dt class="text-muted-foreground">IP Address:</dt>
													<dd class="font-mono">{event.ipAddress}</dd>
												</div>
											{/if}
										</dl>
									</div>
								</div>

								<!-- Details JSON -->
								{#if event.details && Object.keys(event.details).length > 0}
									<div>
										<h4 class="text-sm font-medium mb-2">Event Details</h4>
										<pre class="text-xs bg-muted p-3 rounded-md overflow-auto">{JSON.stringify(
												event.details,
												null,
												2
											)}</pre>
									</div>
								{/if}

								<!-- User Agent -->
								{#if event.userAgent && event.userAgent !== 'Unknown'}
									<div>
										<h4 class="text-sm font-medium mb-2">User Agent</h4>
										<p class="text-xs bg-muted p-3 rounded-md font-mono break-all">
											{event.userAgent}
										</p>
									</div>
								{/if}
							</div>
						</DialogContent>
					</Dialog>
				{/if}
			</div>
		</div>
	{/each}
</div>
