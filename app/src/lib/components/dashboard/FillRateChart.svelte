<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Chart from '$lib/components/ui/chart';
	import { PieChart } from 'layerchart';

	let { data }: { data: Array<{ label: string; value: number }> } = $props();

	const chartConfig = {
		filled: { label: 'Filled', color: 'var(--color-chart-1)' },
		available: { label: 'Available', color: 'var(--color-chart-2)' }
	} satisfies Chart.ChartConfig;
</script>

<Card.Root class="p-6">
	<Card.Header class="pb-4">
		<Card.Title class="text-lg font-semibold">Fill Rate Breakdown</Card.Title>
		<Card.Description>Current shift assignment status</Card.Description>
	</Card.Header>
	<Card.Content>
		<Chart.Container config={chartConfig} class="h-64">
			<PieChart 
				data={data.map((item, i) => ({
					...item, 
					fill: i === 0 ? 'var(--color-chart-1)' : 'var(--color-chart-2)'
				}))}
				innerRadius={60}
			/>
		</Chart.Container>
	</Card.Content>
</Card.Root> 