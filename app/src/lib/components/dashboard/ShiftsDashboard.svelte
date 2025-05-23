<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton';
	import * as Card from '$lib/components/ui/card';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import DashboardMetrics from './DashboardMetrics.svelte';
	import FillRateChart from './FillRateChart.svelte';
	import ScheduleChart from './ScheduleChart.svelte';

	let { 
		isLoading, 
		isError, 
		error, 
		metrics 
	}: { 
		isLoading: boolean;
		isError: boolean;
		error?: Error;
		metrics?: {
			totalShifts: number;
			filledShifts: number;
			availableShifts: number;
			fillRate: number;
			scheduleData: Array<{ schedule: string; total: number; filled: number; fillRate: number; }>;
			fillRateData: Array<{ label: string; value: number; }>;
		} | null;
	} = $props();
</script>

<div class="p-6">
	<div class="max-w-7xl mx-auto">
		<div class="mb-6">
			<h1 class="text-2xl font-semibold mb-2">Shifts Dashboard</h1>
			<p class="text-muted-foreground">
				Overview of all scheduled shifts and performance metrics
			</p>
		</div>

		{#if isLoading}
			<!-- Loading Dashboard Skeletons -->
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
				{#each Array(4) as _, i (i)}
					<Card.Root class="p-6">
						<Skeleton class="h-4 w-24 mb-2" />
						<Skeleton class="h-8 w-16 mb-1" />
						<Skeleton class="h-3 w-20" />
					</Card.Root>
				{/each}
			</div>
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
				{#each Array(2) as _, i (i)}
					<Card.Root class="p-6">
						<Skeleton class="h-6 w-32 mb-4" />
						<Skeleton class="h-64 w-full" />
					</Card.Root>
				{/each}
			</div>
		{:else if isError}
			<div class="text-center py-12">
				<p class="text-destructive text-lg mb-2">Error Loading Dashboard</p>
				<p class="text-muted-foreground">
					{error?.message || 'Unknown error occurred'}
				</p>
			</div>
		{:else if !metrics}
			<div class="text-center py-12">
				<CalendarIcon class="h-12 w-12 text-muted-foreground mx-auto mb-4" />
				<h2 class="text-xl font-semibold mb-2">No Shifts Found</h2>
				<p class="text-muted-foreground">
					No scheduled shifts found for the next 90 days.
				</p>
			</div>
		{:else}
			<!-- Dashboard Content -->
			<DashboardMetrics {metrics} />

			<!-- Charts Grid -->
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
				<FillRateChart data={metrics.fillRateData} />
				<ScheduleChart data={metrics.scheduleData} />
			</div>
		{/if}
	</div>
</div> 