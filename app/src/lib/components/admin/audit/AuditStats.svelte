<script lang="ts">
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { BarChart3, Activity, Users, Clock, Calendar } from 'lucide-svelte';
	import { formatDistanceToNow } from 'date-fns';

	// Types
	interface AuditStats {
		total_events: number;
		unique_actors: number;
		unique_event_types: number;
		earliest_event: string;
		latest_event: string;
	}

	interface EventTypeStats {
		event_type: string;
		event_count: number;
		latest_event: string;
	}

	export let stats: AuditStats;
	export let typeStats: EventTypeStats[];

	// Calculate percentages for progress bars
	function getEventTypePercentage(count: number): number {
		return stats.total_events > 0 ? (count / stats.total_events) * 100 : 0;
	}

	// Format event type for display
	function formatEventType(eventType: string): string {
		return eventType.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase());
	}

	// Get color for event type
	function getEventTypeColor(eventType: string): string {
		if (eventType.includes('login')) return 'bg-green-500';
		if (eventType.includes('created')) return 'bg-blue-500';
		if (eventType.includes('updated')) return 'bg-yellow-500';
		if (eventType.includes('role_changed')) return 'bg-purple-500';
		if (eventType.includes('deleted')) return 'bg-red-500';
		return 'bg-gray-500';
	}

	// Format relative time
	function formatRelativeTime(dateString: string): string {
		if (!dateString) return 'Unknown';
		const date = new Date(dateString);
		return formatDistanceToNow(date, { addSuffix: true });
	}
</script>

<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
	<!-- Overall Statistics -->
	<Card class="lg:col-span-2">
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<BarChart3 class="h-5 w-5" />
				Overview Statistics
			</CardTitle>
			<CardDescription>High-level metrics for audit trail activity</CardDescription>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<!-- Total Events -->
				<div class="text-center">
					<div
						class="flex items-center justify-center w-12 h-12 mx-auto mb-2 bg-primary/10 rounded-lg"
					>
						<Activity class="h-6 w-6 text-primary" />
					</div>
					<div class="text-2xl font-bold">{stats.total_events.toLocaleString()}</div>
					<div class="text-sm text-muted-foreground">Total Events</div>
				</div>

				<!-- Unique Actors -->
				<div class="text-center">
					<div
						class="flex items-center justify-center w-12 h-12 mx-auto mb-2 bg-blue-500/10 rounded-lg"
					>
						<Users class="h-6 w-6 text-blue-500" />
					</div>
					<div class="text-2xl font-bold">{stats.unique_actors}</div>
					<div class="text-sm text-muted-foreground">Unique Actors</div>
				</div>

				<!-- Event Types -->
				<div class="text-center">
					<div
						class="flex items-center justify-center w-12 h-12 mx-auto mb-2 bg-green-500/10 rounded-lg"
					>
						<Clock class="h-6 w-6 text-green-500" />
					</div>
					<div class="text-2xl font-bold">{stats.unique_event_types}</div>
					<div class="text-sm text-muted-foreground">Event Types</div>
				</div>

				<!-- Time Range -->
				<div class="text-center">
					<div
						class="flex items-center justify-center w-12 h-12 mx-auto mb-2 bg-purple-500/10 rounded-lg"
					>
						<Calendar class="h-6 w-6 text-purple-500" />
					</div>
					<div class="text-2xl font-bold">
						{formatRelativeTime(stats.earliest_event).split(' ')[0]}
					</div>
					<div class="text-sm text-muted-foreground">Time Span</div>
				</div>
			</div>

			<!-- Time Range Details -->
			<div class="mt-6 pt-4 border-t">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
					<div>
						<span class="text-muted-foreground">Earliest event:</span>
						<span class="ml-2 font-medium">{formatRelativeTime(stats.earliest_event)}</span>
					</div>
					<div>
						<span class="text-muted-foreground">Latest event:</span>
						<span class="ml-2 font-medium">{formatRelativeTime(stats.latest_event)}</span>
					</div>
				</div>
			</div>
		</CardContent>
	</Card>

	<!-- Event Type Breakdown -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<Activity class="h-5 w-5" />
				Event Types
			</CardTitle>
			<CardDescription>Breakdown by event type with activity distribution</CardDescription>
		</CardHeader>
		<CardContent>
			<div class="space-y-4">
				{#each typeStats.slice(0, 6) as stat (stat.event_type)}
					{@const percentage = getEventTypePercentage(stat.event_count)}
					{@const color = getEventTypeColor(stat.event_type)}

					<div class="space-y-2">
						<div class="flex items-center justify-between text-sm">
							<div class="flex items-center gap-2">
								<div class="w-3 h-3 rounded-full {color}"></div>
								<span class="font-medium">{formatEventType(stat.event_type)}</span>
							</div>
							<Badge variant="secondary" class="text-xs">
								{stat.event_count}
							</Badge>
						</div>

						<div class="h-2 bg-gray-200 rounded-full overflow-hidden">
							<div class="h-2 bg-gray-500" style="width: {percentage}%"></div>
						</div>

						<div class="text-xs text-muted-foreground">
							{percentage.toFixed(1)}% • Last: {formatRelativeTime(stat.latest_event)}
						</div>
					</div>
				{/each}

				{#if typeStats.length > 6}
					<div class="text-center text-sm text-muted-foreground mt-4">
						And {typeStats.length - 6} more event types...
					</div>
				{/if}
			</div>
		</CardContent>
	</Card>
</div>
