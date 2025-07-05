<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';

	let { firstMonthName }: { firstMonthName: string } = $props();
</script>

<div class="flex items-center gap-2">
	<CalendarIcon class="h-5 w-5 text-primary" />
	<h2 class="text-xl font-semibold">Admin Shift Calendar</h2>
	<span class="text-sm text-muted-foreground">- {firstMonthName}</span>
</div>
