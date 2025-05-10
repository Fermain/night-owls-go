<script lang="ts">
	import { createQuery } from '@tanstack/svelte-query';
	import type { Schedule } from '$lib/components/schedules_table/columns'; // Import Schedule type
	import { columns } from '$lib/components/schedules_table/columns'; // Import columns definition
	import SchedulesDataTable from '$lib/components/schedules_table/schedules-data-table.svelte'; // Import the table component

	// Define the type for the API response (array of schedules)
	// This should match the structure returned by your /schedules endpoint
	type SchedulesAPIResponse = Schedule[];

	const fetchSchedules = async (): Promise<SchedulesAPIResponse> => {
		const response = await fetch('/schedules'); // Ensure this is the correct API endpoint
		if (!response.ok) {
			throw new Error(`API request failed: ${response.status} ${response.statusText}`);
		}
		return response.json();
	};

	const schedulesQuery = createQuery<SchedulesAPIResponse, Error, SchedulesAPIResponse, string[]>({
		queryKey: ['schedules'],
		queryFn: fetchSchedules
	});

	// Reactive assignment for data to pass to the table
	// Ensure $schedulesQuery.data is not undefined before passing
	let tableData: SchedulesAPIResponse = $derived($schedulesQuery.data ?? []);

	// Svelte 5 runes for easier derived state (optional, could also use $query.data, $query.isLoading etc. in template)
	// const schedules = $derived(query.data);
	// const isLoading = $derived(query.isLoading);
	// const error = $derived(query.error);
</script>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-4">Community Watch Schedules</h1>

	{#if $schedulesQuery.isLoading}
		<p>Loading schedules...</p>
	{:else if $schedulesQuery.isError}
		<p class="text-red-500">Error fetching schedules: {$schedulesQuery.error?.message}</p>
	{:else if $schedulesQuery.data && $schedulesQuery.data.length > 0}
		<SchedulesDataTable {columns} data={tableData} />
	{:else}
		<p>No schedules available at the moment.</p>
	{/if}
</div>
