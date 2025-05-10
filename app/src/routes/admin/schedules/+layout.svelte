<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import ListIcon from 'lucide-svelte/icons/list';
	import PlusCircleIcon from 'lucide-svelte/icons/plus-circle';
	import CalendarDaysIcon from 'lucide-svelte/icons/calendar-days'; // Icon for individual schedules
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { createQuery } from '@tanstack/svelte-query';
	import type { Schedule as ScheduleData } from '$lib/components/schedules_table/columns'; // Using existing type

	let searchTerm = $state(''); // Moved before $props() and for sidebar's search input
	let { children } = $props();

	// TODO: Potentially create a store for selectedSchedule, similar to selectedUserForForm
	// For now, selection will be driven by URL query param `scheduleId` read by the page component.

	const staticNavItems = [
		{
			title: 'All Schedules', // This could be a "Dashboard" or overview link
			url: '/admin/schedules',
			icon: ListIcon
		}
	];

	// Fetch schedules for the sidebar list
	const fetchSchedulesForSidebar = async (): Promise<ScheduleData[]> => {
		const response = await fetch('/api/admin/schedules'); // No search, gets all
		if (!response.ok) {
			throw new Error('Failed to fetch schedules for sidebar');
		}
		return response.json() as Promise<ScheduleData[]>;
	};

	// Using $derived for the query, simplifying by removing explicit generic types for createQuery
	const schedulesListQuery = $derived(createQuery({
		queryKey: ['adminSchedulesForSidebar'], // Static key, as search term doesn't apply here
		queryFn: fetchSchedulesForSidebar
	}));

	const selectSchedule = (schedule: ScheduleData) => {
		// Navigate to show this schedule's details in the main content area
		// The +page.svelte component for /admin/schedules will need to read this
		goto(`/admin/schedules?scheduleId=${schedule.schedule_id}`);
	};
</script>

{#snippet scheduleListContent()}
	<div class="flex flex-col h-full">
		<!-- Top static nav items -->
		{#each staticNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Dynamic Schedule list (scrollable) -->
		<div class="flex-grow overflow-y-auto border-b">
			{#if $schedulesListQuery.isLoading}
				<div class="p-4 text-sm">Loading schedules...</div>
			{:else if $schedulesListQuery.isError}
				<div class="p-4 text-sm text-destructive">
					Error: {$schedulesListQuery.error.message}
				</div>
			{:else if $schedulesListQuery.data && $schedulesListQuery.data.length > 0}
				{#each $schedulesListQuery.data as schedule (schedule.schedule_id)}
					<a
						href={`/admin/schedules?scheduleId=${schedule.schedule_id}`}
						class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0"
						class:active={page.url.searchParams.get('scheduleId') === schedule.schedule_id.toString()}
						onclick={(event) => {
							event.preventDefault();
							selectSchedule(schedule);
						}}
					>
						<CalendarDaysIcon class="h-4 w-4" />
						<span>{schedule.name || 'Unnamed Schedule'}</span>
					</a>
				{/each}
			{:else}
				<div class="p-4 text-sm text-muted-foreground">No schedules found.</div>
			{/if}
		</div>

		<!-- Create Schedule button at the bottom -->
		<div class="p-3 border-t mt-auto">
			<Button
				href="/admin/schedules/new"
				class="w-full"
				variant={page.url.pathname === '/admin/schedules/new' ? 'default' : 'outline'}
			>
				<PlusCircleIcon class="mr-2 h-4 w-4" />
				Create Schedule
			</Button>
		</div>
	</div>
{/snippet}

<SidebarPage listContent={scheduleListContent} title="Schedules" bind:searchTerm>
	{@render children()}
</SidebarPage> 