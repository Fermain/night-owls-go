<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { formatDistanceToNow } from 'date-fns';
	import { authenticatedFetch } from '$lib/utils/api';
	import type { AdminShiftSlot } from '$lib/types';
	import ScheduleEditDialog from '$lib/components/admin/dialogs/ScheduleEditDialog.svelte';
	import { formatTimeSlot, formatRelativeTime, normalizeDateRange } from '$lib/utils/dateFormatting';
	import { SchedulesApiService } from '$lib/services/api';

	let searchTerm = $state('');

	// Filters state
	let dateRangeStart = $state<string | null>(null);
	let dateRangeEnd = $state<string | null>(null);
	let showOnlyAvailable = $state(false);
	let showOnlyFilled = $state(false);

	// Settings dialog state
	let showSettingsDialog = $state(false);

	// Define navigation items for the shifts section (only dashboard now)
	const shiftsNavItems = [
		{
			title: 'Calendar Dashboard',
			url: '/admin/shifts',
			icon: LayoutDashboardIcon
		}
	];

	// Get selected shift from URL
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));

	// Data Fetching using API service
	async function fetchShiftSlots(from?: string, to?: string) {
		try {
			return await SchedulesApiService.getAllSlots({ from, to });
		} catch (error) {
			console.error('Error fetching shift slots:', error);
			throw error;
		}
	}

	// Main query for shift slots with filtering
	const shiftsQuery = $derived.by(() => {
		const { from: fromDate, to: toDate } = normalizeDateRange(dateRangeStart, dateRangeEnd, 30);

		return createQuery<AdminShiftSlot[], Error>({
			queryKey: ['adminShiftSlots', fromDate, toDate, showOnlyAvailable, showOnlyFilled],
			queryFn: () => fetchShiftSlots(fromDate, toDate),
			staleTime: 1000 * 60 * 5
		});
	});

	// Filtered shifts for display
	const filteredShifts = $derived.by(() => {
		const shifts = $shiftsQuery.data ?? [];
		return shifts.filter((shift) => {
			if (showOnlyAvailable && shift.is_booked) return false;
			if (showOnlyFilled && !shift.is_booked) return false;
			return true;
		});
	});

	// Upcoming shifts (next 7 days) for sidebar
	const upcomingShifts = $derived.by(() => {
		const now = new Date();
		const nextWeek = new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000);
		const upcoming = filteredShifts
			.filter((shift) => {
				const shiftDate = new Date(shift.start_time);
				return shiftDate >= now && shiftDate <= nextWeek;
			})
			.sort((a, b) => new Date(a.start_time).getTime() - new Date(b.start_time).getTime())
			.slice(0, 20); // Limit to 20 items for sidebar

		return upcoming;
	});

	// Event Handlers
	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;
	}

	function selectShift(shift: AdminShiftSlot) {
		goto(`/admin/shifts?shiftStartTime=${encodeURIComponent(shift.start_time)}`);
	}

	function clearFilters() {
		dateRangeStart = null;
		dateRangeEnd = null;
		showOnlyAvailable = false;
		showOnlyFilled = false;
	}

	function openSettingsDialog() {
		showSettingsDialog = true;
	}

	let { children } = $props();
</script>

{#snippet shiftsListContent()}
	<div class="flex flex-col h-full">
		<!-- Top navigation -->
		{#each shiftsNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={item.url === '/admin/shifts' && $page.url.pathname === '/admin/shifts' && !shiftStartTimeFromUrl}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}
		
		<!-- Settings Button -->
		<button
			onclick={openSettingsDialog}
			class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight text-left"
		>
			<SettingsIcon class="h-4 w-4" />
			<span>Schedule Settings</span>
		</button>

		<!-- Filters Section -->
		<div class="p-3 border-b space-y-3">
			<div class="flex items-center gap-2 mb-2">
				<FilterIcon class="h-4 w-4" />
				<Label class="text-sm font-medium">Filters</Label>
			</div>

			<div class="space-y-3">
				<div class="flex items-center space-x-2">
					<Switch id="available-only" bind:checked={showOnlyAvailable} />
					<Label for="available-only" class="text-sm cursor-pointer">Available only</Label>
				</div>
				<div class="flex items-center space-x-2">
					<Switch id="filled-only" bind:checked={showOnlyFilled} />
					<Label for="filled-only" class="text-sm cursor-pointer">Filled only</Label>
				</div>
			</div>

			<div class="space-y-2">
				<Label class="text-sm font-medium">Date Range</Label>
				<DateRangePicker
					initialStartDate={dateRangeStart}
					initialEndDate={dateRangeEnd}
					change={handleDateRangeChange}
					placeholderText="Select range"
				/>
			</div>

			{#if dateRangeStart || dateRangeEnd || showOnlyAvailable || showOnlyFilled}
				<Button variant="outline" size="sm" onclick={clearFilters} class="w-full">
					<FilterIcon class="h-4 w-4 mr-2" />
					Clear Filters
				</Button>
			{/if}
		</div>

		<!-- Upcoming Shifts Header -->
		<div class="p-3 border-b bg-muted/50">
			<div class="flex items-center gap-2">
				<ClockIcon class="h-4 w-4" />
				<span class="text-sm font-medium">Upcoming Shifts</span>
			</div>
			<p class="text-xs text-muted-foreground">Next 7 days</p>
		</div>

		<!-- Upcoming Shifts List -->
		<div class="flex-grow overflow-y-auto">
			{#if $shiftsQuery.isLoading}
				<div class="p-3 space-y-3">
					{#each Array(5) as _, index (index)}
						<Skeleton class="h-16 w-full" />
					{/each}
				</div>
			{:else if $shiftsQuery.isError}
				<div class="p-3 text-sm text-destructive">
					Error loading shifts: {$shiftsQuery.error.message}
				</div>
			{:else if upcomingShifts.length === 0}
				<div class="p-3 text-sm text-muted-foreground">
					No upcoming shifts found matching your filters.
				</div>
			{:else}
				<div class="p-2">
					{#each upcomingShifts as shift (shift.schedule_id + '-' + shift.start_time)}
						<button
							class="w-full p-3 mb-2 text-left rounded-lg border transition-colors hover:bg-accent
								{shiftStartTimeFromUrl === shift.start_time 
									? 'border-primary bg-primary/10' 
									: shift.is_booked 
										? 'border-green-200 bg-green-50 hover:bg-green-100' 
										: 'border-orange-200 bg-orange-50 hover:bg-orange-100'}"
							onclick={() => selectShift(shift)}
						>
							<div class="space-y-1">
								<div class="flex items-center justify-between">
									<h3 class="font-medium text-sm truncate">{shift.schedule_name}</h3>
									{#if shift.is_booked}
										<span class="text-xs font-medium text-green-700 bg-green-100 px-2 py-1 rounded-full">
											Filled
										</span>
									{:else}
										<span class="text-xs font-medium text-orange-700 bg-orange-100 px-2 py-1 rounded-full">
											Available
										</span>
									{/if}
								</div>
								<p class="text-xs text-muted-foreground">
									{formatTimeSlot(shift.start_time, shift.end_time)}
								</p>
								<p class="text-xs text-muted-foreground">
									{formatRelativeTime(shift.start_time)}
								</p>
								{#if shift.is_booked && shift.user_name}
									<p class="text-xs text-green-700 font-medium">
										Assigned to: {shift.user_name}
									</p>
								{/if}
							</div>
						</button>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={shiftsListContent} bind:searchTerm>
	{@render children()}
</SidebarPage>

<!-- Schedule Settings Dialog -->
<ScheduleEditDialog 
	bind:open={showSettingsDialog}
	mode="create"
	onCancel={() => showSettingsDialog = false}
	onSuccess={() => showSettingsDialog = false}
/>
