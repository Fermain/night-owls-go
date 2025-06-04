<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { Button } from '$lib/components/ui/button';
	import { Label } from '$lib/components/ui/label';
	import * as Select from '$lib/components/ui/select';
	import {
		Card,
		CardContent,
		CardDescription,
		CardHeader,
		CardTitle
	} from '$lib/components/ui/card';
	import { UserMultiSelect } from '$lib/components/ui/user-multiselect';
	import { Filter, X } from 'lucide-svelte';

	// Types
	interface Filters {
		event_type: string;
		actor_user_id: string;
		target_user_id: string;
		limit: number;
		offset: number;
	}

	interface EventTypeStats {
		event_type: string;
		event_count: number;
		latest_event: string;
	}

	// Props
	let { filters, typeStats }: { filters: Filters; typeStats: EventTypeStats[] } = $props();

	const dispatch = createEventDispatcher<{ change: Filters }>();

	let localFilters = { ...filters };

	// Multiselect values
	let selectedActorUserIds = $state<(number | string)[]>([]);
	let selectedTargetUserIds = $state<(number | string)[]>([]);

	// Apply filters
	function applyFilters() {
		// Convert selected user arrays to comma-separated strings for the API
		localFilters.actor_user_id = selectedActorUserIds.join(',');
		localFilters.target_user_id = selectedTargetUserIds.join(',');

		dispatch('change', { ...localFilters, offset: 0 });
	}

	// Clear filters
	function clearFilters() {
		localFilters = {
			event_type: '',
			actor_user_id: '',
			target_user_id: '',
			limit: 50,
			offset: 0
		};
		// Clear multiselect values
		selectedActorUserIds = [];
		selectedTargetUserIds = [];
		dispatch('change', localFilters);
	}

	// Format event type for display
	function formatEventType(eventType: string): string {
		return eventType.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase());
	}

	// Check if any filters are active
	let hasActiveFilters = $derived(
		localFilters.event_type || selectedActorUserIds.length > 0 || selectedTargetUserIds.length > 0
	);

	// Convert limit number to string for select
	let limitValue = $state(localFilters.limit.toString());

	// Update localFilters.limit when limitValue changes
	$effect(() => {
		localFilters.limit = parseInt(limitValue) || 50;
	});

	// Handle event type selection
	let eventTypeValue = $state(localFilters.event_type);

	// Update localFilters.event_type when eventTypeValue changes
	$effect(() => {
		localFilters.event_type = eventTypeValue;
	});
</script>

<Card>
	<CardHeader>
		<CardTitle class="flex items-center gap-2">
			<Filter class="h-5 w-5" />
			Filters
		</CardTitle>
		<CardDescription>Filter audit events by type and users</CardDescription>
	</CardHeader>
	<CardContent>
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
			<!-- Event Type Filter -->
			<div class="space-y-2">
				<Label for="event-type">Event Type</Label>
				<Select.Root type="single" bind:value={eventTypeValue}>
					<Select.Trigger>
						<span>{eventTypeValue ? formatEventType(eventTypeValue) : 'All event types'}</span>
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="">All event types</Select.Item>
						{#each typeStats as stat (stat.event_type)}
							<Select.Item value={stat.event_type}>
								{formatEventType(stat.event_type)} ({stat.event_count})
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</div>

			<!-- Actor Users Filter -->
			<div class="space-y-2">
				<Label>Actor Users</Label>
				<UserMultiSelect
					bind:selectedUserIds={selectedActorUserIds}
					placeholder="Select actors..."
					variant="compact"
					maxDisplayItems={2}
				/>
			</div>

			<!-- Target Users Filter -->
			<div class="space-y-2">
				<Label>Target Users</Label>
				<UserMultiSelect
					bind:selectedUserIds={selectedTargetUserIds}
					placeholder="Select targets..."
					variant="compact"
					maxDisplayItems={2}
				/>
			</div>
		</div>

		<!-- Filter Actions -->
		<div class="flex items-center justify-between mt-4 pt-4 border-t">
			<div class="flex items-center gap-2">
				<Button onclick={applyFilters} size="sm">Apply Filters</Button>
				{#if hasActiveFilters}
					<Button onclick={clearFilters} variant="outline" size="sm">
						<X class="h-4 w-4 mr-1" />
						Clear
					</Button>
				{/if}
			</div>

			<!-- Results per page -->
			<div class="flex items-center gap-2 text-sm">
				<Label for="limit">Per page:</Label>
				<Select.Root type="single" bind:value={limitValue}>
					<Select.Trigger class="w-20">
						<span>{limitValue}</span>
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="25">25</Select.Item>
						<Select.Item value="50">50</Select.Item>
						<Select.Item value="100">100</Select.Item>
					</Select.Content>
				</Select.Root>
			</div>
		</div>
	</CardContent>
</Card>
