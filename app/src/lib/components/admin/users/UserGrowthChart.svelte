<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Chart from '$lib/components/ui/chart';
	import { AreaChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import type { UserGrowthData } from '$lib/utils/userProcessing';

	let { data }: { data: UserGrowthData[] } = $props();

	const chartConfig = {
		total: { label: 'Total Users', color: 'var(--color-chart-1)' },
		new: { label: 'New Users', color: 'var(--color-chart-2)' }
	} satisfies Chart.ChartConfig;

	// Calculate growth percentage from first to last data point
	const growthPercentage = $derived(() => {
		if (data.length < 2) return 0;
		const first = data[0].total;
		const last = data[data.length - 1].total;
		return first > 0 ? Math.round(((last - first) / first) * 100) : 0;
	});
</script>

<Card.Root>
	<Card.Header class="pb-4">
		<Card.Title>User Growth</Card.Title>
		<Card.Description>User registration trends over the last 6 months</Card.Description>
	</Card.Header>
	<Card.Content>
		<Chart.Container config={chartConfig} class="h-72">
			<AreaChart {data} x="period" y="total" xScale={scaleBand().padding(0.1)} />
		</Chart.Container>
	</Card.Content>
	<Card.Footer class="pt-4">
		<div class="flex w-full items-start gap-2 text-sm">
			<div class="grid gap-3">
				<div class="flex items-center gap-2 font-medium leading-none">
					Trending up by {Math.abs(growthPercentage())}% this period
					<TrendingUpIcon class="h-4 w-4" />
				</div>
				<div class="flex items-center gap-2 leading-none text-muted-foreground">
					Showing user growth over 6 months
				</div>
			</div>
		</div>
	</Card.Footer>
</Card.Root>
