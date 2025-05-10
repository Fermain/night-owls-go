<script lang="ts">
  import { createQuery } from '@tanstack/svelte-query';
  import type { Schedule } from '$lib/components/schedules_table/columns'; // Reusing Schedule type
  import { columns as publicColumns } from '$lib/components/schedules_table/columns'; // Using public columns for now
  import SchedulesDataTable from '$lib/components/schedules_table/schedules-data-table.svelte';
  import { Button } from '$lib/components/ui/button'; // For "Create New" button

  // Define the type for the API response (array of schedules)
  type AdminSchedulesAPIResponse = Schedule[];

  const fetchAdminSchedules = async (): Promise<AdminSchedulesAPIResponse> => {
    const response = await fetch('/api/admin/schedules'); 
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`API request failed: ${response.status} ${response.statusText} - ${errorText}`);
    }
    return response.json();
  };

  const adminSchedulesQuery = createQuery<AdminSchedulesAPIResponse, Error, AdminSchedulesAPIResponse, string[]>({
    queryKey: ['adminSchedules'],
    queryFn: fetchAdminSchedules,
  });

  // For now, we use the public columns. Later we might define admin-specific columns (e.g., with actions).
  // TODO: Define adminColumns by adding an 'actions' column to publicColumns
  let tableColumns = publicColumns; 

  let tableData: AdminSchedulesAPIResponse = $derived($adminSchedulesQuery.data ?? []);

</script>

<svelte:head>
  <title>Admin - Manage Schedules</title>
</svelte:head>

<div class="container mx-auto p-4">
  <div class="flex justify-between items-center mb-6">
    <h1 class="text-2xl font-semibold">Manage Schedules</h1>
    <a href="/admin/schedules/new">
      <Button variant="default">
        <!-- TODO: Add icon e.g. <Plus class="mr-2 h-4 w-4" /> -->
        Create New Schedule
      </Button>
    </a>
  </div>

  {#if $adminSchedulesQuery.isLoading}
    <p>Loading schedules...</p>
  {:else if $adminSchedulesQuery.isError}
    <p class="text-red-500">Error fetching schedules: {$adminSchedulesQuery.error?.message}</p>
    {#if $adminSchedulesQuery.error?.message?.includes("Failed to decode request body")}
        <p class="text-sm text-gray-600 mt-1">This might indicate an issue with the request sent by the client or how the server expects the data.</p>
    {/if}
    {#if $adminSchedulesQuery.error?.message?.includes("Failed to create schedule") || $adminSchedulesQuery.error?.message?.includes("Failed to list schedules") }
        <p class="text-sm text-gray-600 mt-1">This often points to a server-side or database issue. Check server logs.</p>
    {/if}
  {:else if $adminSchedulesQuery.data}
    {#if tableData.length === 0}
      <p>No schedules found. <a href="/admin/schedules/new" class="text-blue-600 hover:underline">Create the first one!</a></p>
    {:else}
      <SchedulesDataTable columns={tableColumns} data={tableData} />
    {/if}
  {:else}
    <p>No schedule data available.</p>
  {/if}
</div> 