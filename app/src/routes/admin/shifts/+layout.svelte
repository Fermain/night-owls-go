<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { Label } from '$lib/components/ui/label';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { Badge } from '$lib/components/ui/badge';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { AdminShiftSlot } from '$lib/types';
	import ScheduleEditDialog from '$lib/components/admin/dialogs/ScheduleEditDialog.svelte';
	import {
		formatTimeSlot,
		formatRelativeTime,
		normalizeDateRange
	} from '$lib/utils/dateFormatting';

	// Format shift title using watch tradition (previous day + "Night" + time)
	function formatShiftTitle(startTime: string, endTime: string): string {
		const start = new Date(startTime);
		const end = new Date(endTime);
		
		// Get the previous day for the "Night" reference
		const previousDay = new Date(start);
		previousDay.setDate(previousDay.getDate() - 1);
		
		const dayName = previousDay.toLocaleDateString('en-US', { weekday: 'long', timeZone: 'UTC' });
		
		// Format time as condensed AM/PM (e.g., "1-3AM")
		const startHour = start.getUTCHours();
		const endHour = end.getUTCHours();
		
		// Convert to 12-hour format
		const formatHour = (hour: number) => hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
		const getAmPm = (hour: number) => hour < 12 ? 'AM' : 'PM';
		
		const startHour12 = formatHour(startHour);
		const endHour12 = formatHour(endHour);
		const endAmPm = getAmPm(endHour);
		
		const timeRange = `${startHour12}-${endHour12}${endAmPm}`;
		
		return `${dayName} Night ${timeRange}`;
	}
	import { SchedulesApiService } from '$lib/services/api';

	// Filters state - active by default with both selected
	let dateRangeStart = $state<string | null>(null);
	let dateRangeEnd = $state<string | null>(null);
	let showFilled = $state(true);
	let showUnfilled = $state(true);

	// Settings dialog state
	let showSettingsDialog = $state(false);

	// Define navigation items for the shifts section
	const shiftsNavItems = [
		{
			title: 'Dashboard',
			url: '/admin/shifts',
			icon: LayoutDashboardIcon,
			description: 'Calendar view'
		},
		{
			title: 'Bulk Assignment',
			url: '/admin/shifts/bulk-signup',
			icon: CalendarDaysIcon,
			description: 'Individual & pattern selection'
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
			queryKey: ['adminShiftSlots', fromDate, toDate, showFilled, showUnfilled],
			queryFn: () => fetchShiftSlots(fromDate, toDate),
			staleTime: 1000 * 60 * 5
		});
	});

	// Filtered shifts for display
	const filteredShifts = $derived.by(() => {
		const shifts = $shiftsQuery.data ?? [];
		return shifts.filter((shift) => {
			if (showFilled && shift.is_booked) return true;
			if (showUnfilled && !shift.is_booked) return true;
			return false;
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
		showFilled = true;
		showUnfilled = true;
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
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-3 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={item.url === '/admin/shifts' &&
					$page.url.pathname === '/admin/shifts' &&
					!shiftStartTimeFromUrl}
			>
				<item.icon class="h-4 w-4 shrink-0" />
				<div class="flex flex-col">
					<span class="font-medium">{item.title}</span>
					<span class="text-xs text-muted-foreground">{item.description}</span>
				</div>
			</a>
		{/each}

		<!-- Settings Button -->
		<button
			onclick={openSettingsDialog}
			class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-3 whitespace-nowrap border-b p-4 text-sm leading-tight text-left"
		>
			<SettingsIcon class="h-4 w-4 shrink-0" />
			<div class="flex flex-col">
				<span class="font-medium">Settings</span>
				<span class="text-xs text-muted-foreground">Manage schedules</span>
			</div>
		</button>

		<!-- Filters Section -->
		<div class="p-3 border-b space-y-3">
			<div class="flex items-center gap-2 mb-2">
				<FilterIcon class="h-4 w-4" />
				<Label class="text-sm font-medium">Filters</Label>
			</div>

			<div class="space-y-3">
				<div class="flex items-center space-x-2">
					<Switch id="filled-filter" bind:checked={showFilled} />
					<Label for="filled-filter" class="text-sm cursor-pointer">Filled</Label>
				</div>
				<div class="flex items-center space-x-2">
					<Switch id="unfilled-filter" bind:checked={showUnfilled} />
					<Label for="unfilled-filter" class="text-sm cursor-pointer">Unfilled</Label>
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

			{#if dateRangeStart || dateRangeEnd || !showFilled || !showUnfilled}
				<Button variant="outline" size="sm" onclick={clearFilters} class="w-full">
					<FilterIcon class="h-4 w-4 mr-2" />
					Reset Filters
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
							class="group w-full p-3 mb-2 text-left rounded-lg border transition-all duration-200 hover:shadow-sm
								{shiftStartTimeFromUrl === shift.start_time
								? 'border-primary/50 bg-primary/10 shadow-primary/5'
								: 'border-border/50 bg-card/50 hover:bg-accent/50 hover:border-border'}"
							onclick={() => selectShift(shift)}
						>
							<div class="space-y-2">
								<div class="flex items-center justify-between">
									<h3 class="font-medium text-sm text-card-foreground truncate group-hover:text-accent-foreground transition-colors">{formatShiftTitle(shift.start_time, shift.end_time)}</h3>
									{#if shift.is_booked}
										<span class="status-safe text-xs">
											Filled
										</span>
									{:else}
										<span class="status-warning text-xs">
											Available
										</span>
									{/if}
								</div>
								<div class="flex items-center gap-2">
									<Badge variant="secondary" class="text-xs">{shift.schedule_name}</Badge>
								</div>
								<p class="text-xs text-muted-foreground">
									{formatRelativeTime(shift.start_time)}
								</p>
								{#if shift.is_booked && shift.user_name}
									<p class="text-xs font-medium" style="color: hsl(var(--safety-green))">
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

<SidebarPage listContent={shiftsListContent}>
	{@render children()}
</SidebarPage>

<!-- Schedule Settings Dialog -->
<ScheduleEditDialog
	bind:open={showSettingsDialog}
	mode="create"
	onCancel={() => (showSettingsDialog = false)}
	onSuccess={() => (showSettingsDialog = false)}
/>
