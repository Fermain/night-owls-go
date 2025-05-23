<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Chart from '$lib/components/ui/chart';
	import { BarChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import BarChartIcon from '@lucide/svelte/icons/bar-chart-3';
	import type { UserShiftDistribution } from '$lib/utils/userProcessing';

	let { distribution }: { distribution: UserShiftDistribution[] } = $props();

	const chartConfig = {
		shiftCount: { label: 'Shift Count', color: 'var(--color-chart-1)' }
	} satisfies Chart.ChartConfig;

	// Filter to only show users with shifts and limit to top 10
	const chartData = $derived(
		distribution
			.filter((user) => user.shiftCount > 0)
			.slice(0, 10)
			.map((user) => ({
				userName:
					user.userName.length > 12 ? user.userName.substring(0, 12) + '...' : user.userName,
				shiftCount: user.shiftCount,
				percentage: user.percentage,
				fullName: user.userName
			}))
	);

	const totalUsersWithShifts = $derived(distribution.filter((u) => u.shiftCount > 0).length);
</script>

<Card.Root class="col-span-2">
	<Card.Header class="pb-4">
		<div class="flex items-center gap-2">
			<BarChartIcon class="h-5 w-5 text-muted-foreground" />
			<Card.Title>Shift Distribution</Card.Title>
		</div>
		<Card.Description>
			Number of shifts per volunteer (showing top {Math.min(10, totalUsersWithShifts)} of {totalUsersWithShifts})
		</Card.Description>
	</Card.Header>
	<Card.Content>
		{#if chartData.length === 0}
			<div class="text-center py-12 text-muted-foreground">No shift assignments to display</div>
		{:else}
			<Chart.Container config={chartConfig} class="h-80">
				<BarChart data={chartData} x="userName" y="shiftCount" xScale={scaleBand().padding(0.2)} />
			</Chart.Container>
		{/if}
	</Card.Content>
	<Card.Footer class="pt-4">
		<div class="flex w-full items-start gap-2 text-sm">
			<div class="grid gap-2">
				<div class="flex items-center gap-2 font-medium leading-none">
					Distribution of volunteer workload
				</div>
				<div class="flex items-center gap-2 leading-none text-muted-foreground">
					Helps identify workload balance across the team
				</div>
			</div>
		</div>
	</Card.Footer>
</Card.Root>
