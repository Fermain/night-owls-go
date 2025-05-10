<script lang="ts">
	import ScheduleForm from '$lib/components/admin/schedules/ScheduleForm.svelte';
	import type { PageData } from './$types';

	export let data: PageData;
</script>

{#if data.error}
	<div class="container mx-auto p-4">
		<h1 class="text-2xl font-bold text-destructive mb-4">Error Loading Schedule</h1>
		<p class="text-destructive">{data.error}</p>
		<p class="mt-4">
			<a href="/admin/schedules" class="text-blue-600 hover:underline">Return to schedules list</a>
		</p>
	</div>
{:else if data.schedule}
	<ScheduleForm schedule={data.schedule} />
{:else}
	<!-- This case should ideally not be reached if error handling in load is correct -->
	<div class="container mx-auto p-4">
		<h1 class="text-2xl font-bold text-destructive mb-4">Schedule Not Found</h1>
		<p>The requested schedule (ID: {data.scheduleId}) could not be found or loaded.</p>
		<p class="mt-4">
			<a href="/admin/schedules" class="text-blue-600 hover:underline">Return to schedules list</a>
		</p>
	</div>
{/if}
