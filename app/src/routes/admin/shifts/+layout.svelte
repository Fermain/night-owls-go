<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
	import ShiftFilters from '$lib/components/admin/shifts/ShiftFilters.svelte';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import CalendarDaysIcon from '@lucide/svelte/icons/calendar-days';
	import { page } from '$app/state';

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
	let shiftStartTimeFromUrl = $derived(page.url.searchParams.get('shiftStartTime'));

	let { children } = $props();
</script>

{#snippet shiftsListContent()}
	<div class="flex flex-col h-full">
		<!-- Top navigation -->
		{#each shiftsNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={page.url.pathname === item.url && !shiftStartTimeFromUrl}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Filters Section -->
		<ShiftFilters bind:showFilled bind:showUnfilled />

		<!-- Shifts List -->
		<div class="flex-grow overflow-y-auto">
			<UpcomingShifts maxItems={15} className="h-full" />
		</div>
	</div>
{/snippet}

<SidebarPage listContent={shiftsListContent}>
	{@render children()}
</SidebarPage>
