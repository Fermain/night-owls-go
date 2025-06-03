<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';

	let { firstMonthName }: { firstMonthName: string } = $props();
</script>

<div class="flex items-center gap-2">
	<CalendarIcon class="h-4 w-4" />
	<h3 class="text-base font-semibold">Shift Calendar</h3>
	<span class="text-sm text-muted-foreground">- {firstMonthName}</span>
</div>
