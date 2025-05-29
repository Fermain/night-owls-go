<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Chart from '$lib/components/ui/chart';
	import { BarChart } from 'layerchart';
	import { scaleBand } from 'd3-scale';

	let {
		data
	}: {
		data: Array<{
			schedule: string;
			total: number;
			filled: number;
			fillRate: number;
		}>;
	} = $props();

	const chartConfig = {
		total: { label: 'Total Shifts', color: 'var(--color-chart-1)' }
	} satisfies Chart.ChartConfig;
</script>

<Card.Root class="p-6">
	<Card.Header class="pb-4">
		<Card.Title class="text-lg font-semibold">Shifts by Schedule</Card.Title>
		<Card.Description>Total shifts per schedule type</Card.Description>
	</Card.Header>
	<Card.Content>
		<Chart.Container config={chartConfig} class="h-64">
			<BarChart {data} x="schedule" y="total" xScale={scaleBand().padding(0.25)} />
		</Chart.Container>
	</Card.Content>
</Card.Root>
