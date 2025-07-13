<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';

	// Utilities with new patterns
	import { apiGet } from '$lib/utils/api';
	import { classifyError } from '$lib/utils/errors';

	// Components
	import AuditTimeline from '$lib/components/admin/audit/AuditTimeline.svelte';
	import AuditFilters from '$lib/components/admin/audit/AuditFilters.svelte';
	import AuditStats from '$lib/components/admin/audit/AuditStats.svelte';
	import { Button } from '$lib/components/ui/button';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { Shield, Clock, Activity, Users } from 'lucide-svelte';

	// Types using our new domain types and API mappings
	import type { AuditEvent } from '$lib/types/domain';
	import { mapAPIAuditEventToDomain } from '$lib/types/api-mappings';

	// Legacy interface for stats (until backend provides typed responses)
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

	// Filters
	let currentFilters = $state({
		event_type: '',
		actor_user_id: '',
		target_user_id: '',
		limit: 50,
		offset: 0
	});

	// Accumulated events for load-more functionality
	let allEvents = $state<AuditEvent[]>([]);
	let hasLoadedInitialData = $state(false);

	// Query for audit events using our new API utilities
	const auditEventsQuery = $derived(
		createQuery<AuditEvent[], Error>({
			queryKey: ['auditEvents', currentFilters],
			queryFn: async () => {
				try {
					const params: Record<string, string | number> = {
						limit: currentFilters.limit,
						offset: currentFilters.offset
					};

					if (currentFilters.event_type) params.event_type = currentFilters.event_type;
					if (currentFilters.actor_user_id) params.actor_user_id = currentFilters.actor_user_id;
					if (currentFilters.target_user_id) params.target_user_id = currentFilters.target_user_id;

					const apiEvents = await apiGet<Record<string, unknown>[]>('/api/admin/audit-events', {
						params
					});
					return apiEvents.map(mapAPIAuditEventToDomain);
				} catch (error) {
					throw classifyError(error);
				}
			}
		})
	);

	// Query for stats using our new API utilities
	const statsQuery = $derived(
		createQuery<AuditStats, Error>({
			queryKey: ['auditStats'],
			queryFn: async () => {
				try {
					return await apiGet<AuditStats>('/api/admin/audit-events/stats');
				} catch (error) {
					throw classifyError(error);
				}
			},
			// Only fetch stats once and cache them
			staleTime: 1000 * 60 * 10 // 10 minutes
		})
	);

	// Query for type stats using our new API utilities
	const typeStatsQuery = $derived(
		createQuery<EventTypeStats[], Error>({
			queryKey: ['auditTypeStats'],
			queryFn: async () => {
				try {
					return await apiGet<EventTypeStats[]>('/api/admin/audit-events/type-stats');
				} catch (error) {
					throw classifyError(error);
				}
			},
			// Only fetch type stats once and cache them
			staleTime: 1000 * 60 * 10 // 10 minutes
		})
	);

	// Handle filter changes
	function handleFiltersChange(newFilters: typeof currentFilters) {
		currentFilters = { ...newFilters, offset: 0 }; // Reset pagination
		allEvents = []; // Clear accumulated events
		hasLoadedInitialData = false;
	}

	// Handle pagination - load more events
	function loadMore() {
		currentFilters = {
			...currentFilters,
			offset: currentFilters.offset + currentFilters.limit
		};
	}

	// Effect to handle event accumulation when query data changes
	$effect(() => {
		const events = $auditEventsQuery.data;
		if (events && events.length > 0) {
			if (currentFilters.offset === 0) {
				// New search - replace events
				allEvents = events;
				hasLoadedInitialData = true;
			} else {
				// Load more - append events if not already present
				const newEvents = events.filter(
					(event) => !allEvents.some((existing) => existing.id === event.id)
				);
				allEvents = [...allEvents, ...newEvents];
			}
		} else if (events && events.length === 0 && currentFilters.offset === 0) {
			// Empty result for new search
			allEvents = [];
			hasLoadedInitialData = true;
		}
	});

	// Derived state for UI
	const isInitialLoading = $derived($auditEventsQuery.isLoading && !hasLoadedInitialData);
	const isLoadingMore = $derived($auditEventsQuery.isLoading && hasLoadedInitialData);
	const hasMoreData = $derived(($auditEventsQuery.data?.length ?? 0) >= currentFilters.limit);
	const displayEvents = $derived(allEvents);
</script>

<svelte:head>
	<title>Admin History - Night Owls Control</title>
</svelte:head>

<div class="container mx-auto py-6 space-y-6">
	<!-- Page Header -->
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-3">
			<div class="bg-primary/10 p-2 rounded-lg">
				<Shield class="h-6 w-6 text-primary" />
			</div>
			<div>
				<h1 class="text-3xl font-bold tracking-tight">Admin History</h1>
				<p class="text-muted-foreground">Security audit trail and activity monitoring</p>
			</div>
		</div>

		<!-- Quick Stats -->
		{#if $statsQuery.data}
			{@const stats = $statsQuery.data}
			<div class="flex items-center gap-4">
				<div class="flex items-center gap-2 text-sm">
					<Activity class="h-4 w-4 text-primary" />
					<span class="font-medium">{stats.total_events}</span>
					<span class="text-muted-foreground">events</span>
				</div>
				<div class="flex items-center gap-2 text-sm">
					<Users class="h-4 w-4 text-blue-500" />
					<span class="font-medium">{stats.unique_actors}</span>
					<span class="text-muted-foreground">actors</span>
				</div>
				<div class="flex items-center gap-2 text-sm">
					<Clock class="h-4 w-4 text-green-500" />
					<span class="font-medium">{stats.unique_event_types}</span>
					<span class="text-muted-foreground">types</span>
				</div>
			</div>
		{/if}
	</div>

	<!-- Statistics Overview -->
	{#if $statsQuery.data && $typeStatsQuery.data && $typeStatsQuery.data.length > 0}
		<AuditStats stats={$statsQuery.data} typeStats={$typeStatsQuery.data} />
	{/if}

	<!-- Filters -->
	<AuditFilters
		filters={currentFilters}
		typeStats={$typeStatsQuery.data ?? []}
		on:change={(e) => handleFiltersChange(e.detail)}
	/>

	<!-- Timeline -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<Clock class="h-5 w-5" />
				Activity Timeline
			</CardTitle>
			<CardDescription>Chronological view of all administrative and user actions</CardDescription>
		</CardHeader>
		<CardContent>
			{#if isInitialLoading}
				<div class="flex items-center justify-center py-12">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
				</div>
			{:else if $auditEventsQuery.isError}
				<div class="flex items-center justify-center py-12">
					<div class="text-center">
						<p class="text-destructive font-medium mb-2">Error loading audit events</p>
						<p class="text-sm text-muted-foreground mb-4">
							{$auditEventsQuery.error?.message || 'Failed to load audit data'}
						</p>
						<Button onclick={() => $auditEventsQuery.refetch()} variant="outline" size="sm"
							>Try Again</Button
						>
					</div>
				</div>
			{:else if displayEvents.length === 0}
				<div class="flex items-center justify-center py-12">
					<div class="text-center">
						<Shield class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
						<p class="text-lg font-medium mb-2">No audit events found</p>
						<p class="text-sm text-muted-foreground">No events match the current filters.</p>
					</div>
				</div>
			{:else}
				<AuditTimeline events={displayEvents} />

				<!-- Load More Button -->
				{#if hasMoreData}
					<div class="flex justify-center mt-6">
						<Button onclick={loadMore} variant="outline" disabled={isLoadingMore}>
							{#if isLoadingMore}
								<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-current mr-2"></div>
							{/if}
							Load More Events
						</Button>
					</div>
				{/if}
			{/if}
		</CardContent>
	</Card>
</div>
