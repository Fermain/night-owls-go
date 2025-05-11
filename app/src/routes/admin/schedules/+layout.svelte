<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { page } from '$app/state'; // Use $app/stores for $page in layout
	import { goto } from '$app/navigation';
	import { createQuery, type QueryKey, type CreateQueryResult } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import { Button } from '$lib/components/ui/button';
	import type { Snippet } from 'svelte';

	type Schedule = {
		id: number;
		name: string;
		start_date: string;
		end_date: string;
		timezone: string;
	};

	let searchTerm = $state('');
	let { children } = $props();

	const fetchSchedules = async (): Promise<Schedule[]> => {
		const response = await fetch('/api/admin/schedules');
		if (!response.ok) {
			toast.error('Failed to fetch schedules');
			throw new Error('Failed to fetch schedules');
		}
		return response.json();
	};

	const schedulesQuery: CreateQueryResult<Schedule[], Error> = createQuery<Schedule[], Error, Schedule[], QueryKey>({
		queryKey: ['adminSchedulesForLayout'],
		queryFn: fetchSchedules
	});

	const schedulesForTemplate = $derived.by(() => {
		const data = $schedulesQuery.data;
		if (!data) return [];
		if (!searchTerm) return data;
		return data.filter(
			(schedule) => schedule.name.toLowerCase().includes(searchTerm.toLowerCase())
		);
	});

	const title = 'Schedules';

	const isNewSchedulePage = $derived(page.url.pathname === '/admin/schedules/new');
	const editScheduleId = $derived(page.url.pathname.match(/\/admin\/schedules\/(\d+)\/edit/)?.[1]);

</script>

{#snippet scheduleListContent()}
	<Sidebar.Menu class="p-2">
		<Sidebar.MenuItem class="mb-2">
			<Button
				variant={isNewSchedulePage ? 'default' : 'outline'}
				class="w-full justify-start"
				onclick={() => goto('/admin/schedules/new')}
			>
				+ Create New Schedule
			</Button>
		</Sidebar.MenuItem>

		{#if $schedulesQuery.isLoading}
			<p class="p-2 text-sm text-muted-foreground">Loading schedules...</p>
		{:else if $schedulesQuery.isError}
			<p class="p-2 text-sm text-destructive">Error loading schedules.</p>
		{:else if schedulesForTemplate.length > 0}
			{#each schedulesForTemplate as schedule (schedule.id)}
				<Sidebar.MenuItem>
					<Sidebar.MenuButton
						onclick={() => goto(`/admin/schedules/${schedule.id}/edit`)}
						isActive={editScheduleId === String(schedule.id)}
					>
						{schedule.name}
					</Sidebar.MenuButton>
				</Sidebar.MenuItem>
			{/each}
		{:else}
			<p class="p-2 text-sm text-muted-foreground">No schedules found.</p>
		{/if}
	</Sidebar.Menu>
{/snippet}

<SidebarPage listContent={scheduleListContent} {title} bind:searchTerm>
	{@render children()}
</SidebarPage>
