<script lang="ts">
	import ScheduleFormNew from '$lib/components/admin/schedules/ScheduleForm.svelte';
	import { selectedScheduleForForm } from '$lib/stores/scheduleEditingStore';

	// Ensure the store is clear when navigating to the new page directly
	selectedScheduleForForm.set(undefined);
</script>

<ScheduleFormNew />
