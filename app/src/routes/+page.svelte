<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { createQuery } from '@tanstack/svelte-query';
	// import SchedulesDataTable from '$lib/components/schedules_table/schedules-data-table.svelte'; // MISSING COMPONENT
	// import { columns } from '$lib/components/schedules_table/columns'; // Columns are for the missing table
	import type { Schedule } from '$lib/types'; // Import Schedule

	// type SchedulesAPIResponse = Schedule[]; // Removed local definition

	const fetchSchedules = async (): Promise<Schedule[]> => {
		// const response = await fetch('/schedules'); // Ensure this is the correct API endpoint. For now, assuming it's public
		const response = await fetch('/api/schedules'); // Corrected to /api/schedules, assuming this needs to go via proxy
		if (!response.ok) {
			throw new Error('Network response was not ok');
		}
		return response.json();
	};

	const query = createQuery<Schedule[], Error>({
		queryKey: ['schedules'],
		queryFn: fetchSchedules
	});

	let data = $derived($query.data ?? []);
	// ...
</script>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-4">Community Watch Schedules</h1>

	{#if $query.isLoading}
		<p>Loading schedules...</p>
	{:else if $query.isError}
		<p class="text-red-500">Error fetching schedules: {$query.error?.message}</p>
	{:else if $query.data && $query.data.length > 0}
		<!-- <SchedulesDataTable {columns} data={data} /> --> <!-- Usage removed due to missing component -->
		<p>Schedule data loaded, but table component is missing.</p> 
		<ul>
			{#each data as schedule (schedule.schedule_id)}
				<li>{schedule.name} - {schedule.cron_expr}</li>
			{/each}
		</ul>
	{:else}
		<p>No schedules available at the moment.</p>
	{/if}
</div>
