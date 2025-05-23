<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';

	let {
		metrics
	}: {
		metrics: {
			totalShifts: number;
			filledShifts: number;
			availableShifts: number;
			fillRate: number;
		};
	} = $props();
</script>

<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
	<Card.Root class="p-6">
		<div class="flex items-center justify-between">
			<div>
				<p class="text-sm font-medium text-muted-foreground">Total Shifts</p>
				<p class="text-2xl font-bold">{metrics.totalShifts}</p>
			</div>
			<CalendarIcon class="h-8 w-8 text-muted-foreground" />
		</div>
	</Card.Root>

	<Card.Root class="p-6">
		<div class="flex items-center justify-between">
			<div>
				<p class="text-sm font-medium text-muted-foreground">Filled Shifts</p>
				<p class="text-2xl font-bold text-green-600">{metrics.filledShifts}</p>
			</div>
			<CheckIcon class="h-8 w-8 text-green-600" />
		</div>
	</Card.Root>

	<Card.Root class="p-6">
		<div class="flex items-center justify-between">
			<div>
				<p class="text-sm font-medium text-muted-foreground">Available</p>
				<p class="text-2xl font-bold text-orange-600">{metrics.availableShifts}</p>
			</div>
			<ClockIcon class="h-8 w-8 text-orange-600" />
		</div>
	</Card.Root>

	<Card.Root class="p-6">
		<div class="flex items-center justify-between">
			<div>
				<p class="text-sm font-medium text-muted-foreground">Fill Rate</p>
				<p class="text-2xl font-bold">{metrics.fillRate}%</p>
			</div>
			<TrendingUpIcon class="h-8 w-8 text-blue-600" />
		</div>
	</Card.Root>
</div>
