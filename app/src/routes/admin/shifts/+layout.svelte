<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Switch } from '$lib/components/ui/switch';
	import { Label } from '$lib/components/ui/label';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/stores';
	import type { AdminShiftSlot } from '$lib/types';
	import { normalizeDateRange } from '$lib/utils/dateFormatting';
	import { SchedulesApiService } from '$lib/services/api';

	// Filters state - active by default with both selected, hardcoded to 2 weeks
	let showFilled = $state(true);
	let showUnfilled = $state(true);

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
		},
		{
			title: 'Settings',
			url: '/admin/shifts/settings',
			icon: SettingsIcon,
			description: 'Manage schedules'
		}
	];

	// Get selected shift from URL
	let shiftStartTimeFromUrl = $derived($page.url.searchParams.get('shiftStartTime'));

	// Data Fetching using API service - hardcoded to 2 weeks
	async function fetchShiftSlots() {
		try {
			const { from: fromDate, to: toDate } = normalizeDateRange(null, null, 14); // 2 weeks
			return await SchedulesApiService.getAllSlots({ from: fromDate, to: toDate });
		} catch (error) {
			console.error('Error fetching shift slots:', error);
			throw error;
		}
	}

	// Main query for shift slots with filtering
	const shiftsQuery = $derived.by(() => {
		return createQuery<AdminShiftSlot[], Error>({
			queryKey: ['adminShiftSlots', showFilled, showUnfilled],
			queryFn: fetchShiftSlots,
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

	function clearFilters() {
		showFilled = true;
		showUnfilled = true;
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
				class:active={$page.url.pathname === item.url && !shiftStartTimeFromUrl}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Inline Filters Section -->
		<div class="p-4 border-b">
			<div class="flex items-center gap-4">
				<div class="flex items-center gap-2">
					<FilterIcon class="h-4 w-4" />
					<Label class="text-sm font-medium">Filters:</Label>
				</div>
				
				<div class="flex items-center gap-4">
					<div class="flex items-center space-x-2">
						<Switch id="filled-filter" bind:checked={showFilled} />
						<Label for="filled-filter" class="text-sm cursor-pointer">Filled</Label>
					</div>
					<div class="flex items-center space-x-2">
						<Switch id="unfilled-filter" bind:checked={showUnfilled} />
						<Label for="unfilled-filter" class="text-sm cursor-pointer">Unfilled</Label>
					</div>
				</div>

				{#if !showFilled || !showUnfilled}
					<Button variant="outline" size="sm" onclick={clearFilters}>
						<FilterIcon class="h-4 w-4 mr-2" />
						Reset
					</Button>
				{/if}
			</div>
		</div>

		<!-- Shifts List (no heading, no rounded edges, no gaps) -->
		<div class="flex-grow overflow-y-auto">
			<UpcomingShifts 
				maxItems={15}
				className="h-full"
				hideHeading={true}
				compactStyle={true}
			/>
		</div>
	</div>
{/snippet}

<SidebarPage listContent={shiftsListContent}>
	{@render children()}
</SidebarPage>

