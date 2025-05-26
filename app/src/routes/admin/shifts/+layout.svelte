<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
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



	// Event Handlers
	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;
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

		<!-- Upcoming Shifts using reusable component -->
		<div class="flex-grow overflow-y-auto p-3">
			<UpcomingShifts 
				maxItems={15}
				className="h-full"
			/>
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
