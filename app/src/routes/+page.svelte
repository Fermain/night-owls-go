<script lang="ts">
    import { onMount } from 'svelte';

    let statusMessage: string = 'Loading...';
    let schedulesData: string | null = null;

    onMount(async () => {
        try {
            const response = await fetch('/schedules'); // Assumes Go backend serves this path
            if (!response.ok) {
                throw new Error(`API request failed: ${response.status} ${response.statusText}`);
            }
            const data = await response.json();
            statusMessage = `Successfully fetched ${data.length} schedule(s).`;
            // To display raw data, you could use:
            // schedulesData = JSON.stringify(data, null, 2);
            // For now, just showing the count and a snippet if available:
            if (data.length > 0) {
                schedulesData = `First schedule name: ${data[0].name}`;
            } else {
                schedulesData = "No schedules returned.";
            }
        } catch (error: any) {
            console.error('Failed to fetch schedules:', error);
            statusMessage = `Error: ${error.message}`;
            schedulesData = 'Check browser console and ensure backend is running and accessible.';
        }
    });
</script>

<h1>API Reachability Test</h1>

<p><strong>Status:</strong> {statusMessage}</p>

{#if schedulesData}
    <pre>{schedulesData}</pre>
{/if}

<p>
    This page attempts to fetch data from the <code>/schedules</code> API endpoint served by the Go backend.
</p>
