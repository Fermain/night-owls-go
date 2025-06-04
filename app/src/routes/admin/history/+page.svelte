<script lang="ts">
	import { onMount } from 'svelte';
	import { authenticatedFetch } from '$lib/utils/api';
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

	// Types
	interface AuditEventDetails {
		old_role?: string;
		new_role?: string;
		name?: { old: string; new: string };
		phone?: { old: string; new: string };
		target_user_name?: string;
		target_user_phone?: string;
		target_role?: string;
		deleted_count?: number;
		[key: string]: unknown; // For any additional properties
	}

	interface AuditEvent {
		event_id: number;
		event_type: string;
		actor_user_id?: number;
		actor_name: string;
		actor_phone: string;
		target_user_id?: number;
		target_name: string;
		target_phone: string;
		entity_type: string;
		entity_id?: number;
		action: string;
		details?: AuditEventDetails;
		ip_address: string;
		user_agent: string;
		created_at: string;
	}

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

	// State
	let events: AuditEvent[] = [];
	let stats: AuditStats | null = null;
	let typeStats: EventTypeStats[] = [];
	let loading = true;
	let error = '';

	// Filters
	let currentFilters = {
		event_type: '',
		actor_user_id: '',
		target_user_id: '',
		limit: 50,
		offset: 0
	};

	// Load audit data
	async function loadAuditData() {
		loading = true;
		error = '';

		try {
			// Build query parameters
			const params = new URLSearchParams();
			if (currentFilters.event_type) params.append('event_type', currentFilters.event_type);
			if (currentFilters.actor_user_id)
				params.append('actor_user_id', currentFilters.actor_user_id);
			if (currentFilters.target_user_id)
				params.append('target_user_id', currentFilters.target_user_id);
			params.append('limit', currentFilters.limit.toString());
			params.append('offset', currentFilters.offset.toString());

			// Fetch events using authenticatedFetch
			const eventsResponse = await authenticatedFetch(`/api/admin/audit-events?${params}`);

			if (!eventsResponse.ok) {
				throw new Error(`Failed to load audit events: ${eventsResponse.status}`);
			}

			events = await eventsResponse.json();

			// Load stats only on initial load
			if (
				currentFilters.offset === 0 &&
				!currentFilters.event_type &&
				!currentFilters.actor_user_id &&
				!currentFilters.target_user_id
			) {
				await loadStats();
			}
		} catch (err) {
			console.error('Error loading audit data:', err);
			error = err instanceof Error ? err.message : 'Failed to load audit data';
		} finally {
			loading = false;
		}
	}

	async function loadStats() {
		try {
			// Load overall stats using authenticatedFetch
			const statsResponse = await authenticatedFetch('/api/admin/audit-events/stats');

			if (statsResponse.ok) {
				stats = await statsResponse.json();
			}

			// Load type stats using authenticatedFetch
			const typeStatsResponse = await authenticatedFetch('/api/admin/audit-events/type-stats');

			if (typeStatsResponse.ok) {
				typeStats = await typeStatsResponse.json();
			}
		} catch (err) {
			console.error('Error loading stats:', err);
		}
	}

	// Handle filter changes
	function handleFiltersChange(newFilters: typeof currentFilters) {
		currentFilters = { ...newFilters, offset: 0 }; // Reset pagination
		loadAuditData();
	}

	// Handle pagination
	function loadMore() {
		currentFilters.offset += currentFilters.limit;
		loadMoreEvents();
	}

	async function loadMoreEvents() {
		try {
			const params = new URLSearchParams();
			if (currentFilters.event_type) params.append('event_type', currentFilters.event_type);
			if (currentFilters.actor_user_id)
				params.append('actor_user_id', currentFilters.actor_user_id);
			if (currentFilters.target_user_id)
				params.append('target_user_id', currentFilters.target_user_id);
			params.append('limit', currentFilters.limit.toString());
			params.append('offset', currentFilters.offset.toString());

			const response = await authenticatedFetch(`/api/admin/audit-events?${params}`);

			if (response.ok) {
				const moreEvents = await response.json();
				events = [...events, ...moreEvents];
			}
		} catch (err) {
			console.error('Error loading more events:', err);
		}
	}

	// Load data on mount
	onMount(() => {
		loadAuditData();
	});
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
		{#if stats}
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
	{#if stats && typeStats.length > 0}
		<AuditStats {stats} {typeStats} />
	{/if}

	<!-- Filters -->
	<AuditFilters
		filters={currentFilters}
		{typeStats}
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
			{#if loading && events.length === 0}
				<div class="flex items-center justify-center py-12">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
				</div>
			{:else if error}
				<div class="flex items-center justify-center py-12">
					<div class="text-center">
						<p class="text-destructive font-medium mb-2">Error loading audit events</p>
						<p class="text-sm text-muted-foreground mb-4">{error}</p>
						<Button onclick={loadAuditData} variant="outline" size="sm">Try Again</Button>
					</div>
				</div>
			{:else if events.length === 0}
				<div class="flex items-center justify-center py-12">
					<div class="text-center">
						<Shield class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
						<p class="text-lg font-medium mb-2">No audit events found</p>
						<p class="text-sm text-muted-foreground">No events match the current filters.</p>
					</div>
				</div>
			{:else}
				<AuditTimeline {events} />

				<!-- Load More Button -->
				{#if events.length >= currentFilters.limit}
					<div class="flex justify-center mt-6">
						<Button onclick={loadMore} variant="outline" disabled={loading}>
							{#if loading}
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
