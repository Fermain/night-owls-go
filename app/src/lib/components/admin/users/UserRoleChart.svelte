<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import * as Chart from '$lib/components/ui/chart';
	import { PieChart } from 'layerchart';
	import type { UserMetrics } from '$lib/utils/userProcessing';

	let { metrics }: { metrics: UserMetrics } = $props();

	const chartConfig = {
		admin: { label: 'Admin', color: 'var(--color-chart-1)' },
		owl: { label: 'Owl', color: 'var(--color-chart-2)' },
		guest: { label: 'Guest', color: 'var(--color-chart-3)' }
	} satisfies Chart.ChartConfig;

	// Transform data for the pie chart
	const chartData = $derived(
		metrics.roleDistribution
			.map((item) => ({
				role: item.role.toLowerCase(),
				count: item.count,
				percentage: item.percentage,
				label: item.role,
				fill: `var(--color-chart-${item.role === 'Admin' ? '1' : item.role === 'Owl' ? '2' : '3'})`
			}))
			.filter((item) => item.count > 0)
	);
</script>

<Card.Root class="flex flex-col">
	<Card.Header class="items-center pb-4">
		<Card.Title>Role Distribution</Card.Title>
		<Card.Description>Breakdown of user roles</Card.Description>
	</Card.Header>
	<Card.Content class="flex-1 pb-4">
		<Chart.Container config={chartConfig} class="mx-auto aspect-square max-h-72">
			<PieChart data={chartData} value="count" />
		</Chart.Container>
	</Card.Content>
	<Card.Footer class="flex-col gap-3 text-sm pt-4">
		<div class="flex items-center gap-2 font-medium leading-none">
			User roles across the platform
		</div>
		<div class="leading-none text-muted-foreground">
			{#each metrics.roleDistribution as role}
				{#if role.count > 0}
					<span class="inline-flex items-center gap-2 mr-6 mb-2">
						<div
							class="w-3 h-3 rounded-full {role.role === 'Admin'
								? 'bg-chart-1'
								: role.role === 'Owl'
									? 'bg-chart-2'
									: 'bg-chart-3'}"
						></div>
						{role.role}: {role.count} ({role.percentage}%)
					</span>
				{/if}
			{/each}
		</div>
	</Card.Footer>
</Card.Root>
