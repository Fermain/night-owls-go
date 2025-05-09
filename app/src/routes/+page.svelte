<script lang="ts">
    import { createQuery } from '@tanstack/svelte-query';

    // Define a function to fetch schedules
    const fetchSchedules = async () => {
        const response = await fetch('/schedules'); // Assumes Go backend serves this path
        if (!response.ok) {
            throw new Error(`API request failed: ${response.status} ${response.statusText}`);
        }
        return response.json();
    };

    // Use createQuery to fetch and manage the schedules data
    const query = createQuery({
        queryKey: ['schedules'], // Unique key for this query
        queryFn: fetchSchedules,  // The function that performs the fetching
        // Options can be added here, e.g., staleTime, gcTime (previously cacheTime)
    });

    // Svelte 5 runes for easier derived state (optional, could also use $query.data, $query.isLoading etc. in template)
    // const schedules = $derived(query.data);
    // const isLoading = $derived(query.isLoading);
    // const error = $derived(query.error);

</script>

<h1>API Reachability Test (with Svelte Query)</h1>

{#if $query.isLoading}
    <p>Loading schedules...</p>
{:else if $query.isError}
    <p style="color: red;">Error: {$query.error?.message || 'Unknown error fetching schedules'}</p>
    <p>Check browser console and ensure backend is running and accessible.</p>
{:else if $query.data}
    <p>Status: Successfully fetched {$query.data.length} schedule(s).</p>
    {#if $query.data.length > 0}
        <pre>First schedule name: {$query.data[0].name}</pre>
    {:else}
        <pre>No schedules returned.</pre>
    {/if}
    <!-- To display raw data:
    <pre>{JSON.stringify($query.data, null, 2)}</pre>
    -->
{:else}
    <p>No data available.</p> <!-- This case should ideally be covered by isLoading or isError -->
{/if}

<p>
    This page attempts to fetch data from the <code>/schedules</code> API endpoint using @tanstack/svelte-query.
</p>
