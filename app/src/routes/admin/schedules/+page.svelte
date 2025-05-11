<script lang="ts">
	// import { createQuery } from '@tanstack/svelte-query'; // No longer needed for adminSchedulesQuery here
	import type { Schedule } from '$lib/components/schedules_table/columns';
	// import { columns as publicColumns } from '$lib/components/schedules_table/columns'; // Not used for dashboard
	// import SchedulesDataTable from '$lib/components/schedules_table/schedules-data-table.svelte'; // Not used for dashboard
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import ScheduleFormNew from '$lib/components/admin/schedules/ScheduleFormNew.svelte';
	import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js'; // For dashboard placeholders

	// Get the currently selected schedule from the store
	let currentScheduleToEdit = $derived($selectedScheduleForForm);

	function handleCreateNew() {
		selectedScheduleForForm.set(undefined);
		goto('/admin/schedules/new');
	}
</script>

<svelte:head>
	<title>
		Admin - {currentScheduleToEdit
			? 'Edit Schedule'
			: page.url.pathname.endsWith('/new')
				? 'New Schedule'
				: 'Schedules Dashboard'}
	</title>
</svelte:head>

<div class="container mx-auto p-4">
	{#if currentScheduleToEdit}
		<ScheduleFormNew schedule={currentScheduleToEdit} />
	{:else if page.url.pathname.endsWith('/new')}
		<ScheduleFormNew />
	{:else}
		<!-- Schedules Dashboard View -->
		<div class="p-4 md:p-8">
			<div class="flex justify-between items-center mb-6">
				<h1 class="text-2xl font-semibold">Schedules Dashboard</h1>
				<Button onclick={handleCreateNew}>Create New Schedule</Button>
			</div>
			<p class="mb-6 text-muted-foreground">
				Overview and management tools for schedules will be displayed here.
			</p>
			<div class="space-y-4">
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each [1, 2, 3] as i (i)}
						<div class="p-4 border rounded-lg bg-card">
							<Skeleton class="h-6 w-3/4 mb-2" />
							<Skeleton class="h-10 w-1/2 mb-4" />
							<Skeleton class="h-4 w-full" />
							<Skeleton class="h-4 w-5/6 mt-1" />
						</div>
					{/each}
				</div>
				<div class="p-4 border rounded-lg bg-card">
					<Skeleton class="h-8 w-1/4 mb-4" />
					<Skeleton class="h-48 w-full" />
				</div>
			</div>
		</div>
	{/if}
</div>
