<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import type { Shift } from '$lib/types/shifts';

	export let shifts: Shift[] = [];
	export let onToggleFilters: () => void = () => {};

	function isTonight(startTime: string): boolean {
		const shiftDate = new Date(startTime);
		const today = new Date();
		return shiftDate.toDateString() === today.toDateString();
	}

	$: totalShifts = shifts.length;
	$: tonightShifts = shifts.filter((s) => isTonight(s.start_time)).length;
	$: availableShifts = shifts.filter((s) => !s.is_booked).length;
</script>

<header class="bg-background border-b border-border sticky top-0 z-40">
	<div class="px-4 py-3">
		<div class="flex items-center justify-between">
			<div class="flex items-center gap-3">
				<div class="border border-border p-2 rounded">
					<ShieldIcon class="h-5 w-5 text-foreground" />
				</div>
				<div>
					<h1 class="text-lg font-semibold text-foreground">Patrol Shifts</h1>
					<p class="text-sm text-muted-foreground">
						{totalShifts} total, {tonightShifts} tonight, {availableShifts} available
					</p>
				</div>
			</div>
			<Button
				variant="outline"
				size="sm"
				class="h-8 px-2"
				onclick={onToggleFilters}
			>
				<FilterIcon class="h-4 w-4" />
			</Button>
		</div>
	</div>
</header> 