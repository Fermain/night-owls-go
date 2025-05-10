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
	{:else if $schedulesQuery.data}
		<SchedulesDataTable {columns} data={tableData} />
		<!-- Debug output, remove later -->
		<!-- <div class="mt-4 p-2 bg-gray-100 rounded">
            <h3 class="font-semibold">Raw Data:</h3>
            <pre class="text-xs whitespace-pre-wrap">{JSON.stringify($schedulesQuery.data, null, 2)}</pre>
        </div> -->
	{:else}
		<p>No schedules found.</p>
	{/if}

	<hr class="my-6" />

	<div class="mt-4 p-4 border rounded shadow-sm">
		<h2 class="text-xl font-semibold mb-2">API Reachability Test (with Svelte Query)</h2>
		{#if $schedulesQuery.isLoading}
			<p>Status: Loading...</p>
		{:else if $schedulesQuery.isError}
			<p>Status: <span class="text-red-500">Error: {$schedulesQuery.error?.message}</span></p>
		{:else if $schedulesQuery.data}
			<p>
				Status: <span class="text-green-500"
					>Successfully fetched {$schedulesQuery.data.length} schedule(s).</span
				>
			</p>
			{#if $schedulesQuery.data.length > 0}
				<p>First schedule name: {$schedulesQuery.data[0].name}</p>
			{/if}
		{:else}
			<p>Status: No data yet.</p>
		{/if}
		<p class="text-sm text-gray-600 mt-2">
			This page attempts to fetch data from the <code>/schedules</code> API endpoint using
			<code>@tanstack/svelte-query</code>.
		</p>
	</div>
</div>
