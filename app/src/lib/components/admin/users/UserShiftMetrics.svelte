<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import TargetIcon from '@lucide/svelte/icons/target';
	import UsersIcon from '@lucide/svelte/icons/users';
	import TrendingUpIcon from '@lucide/svelte/icons/trending-up';
	import BalanceIcon from '@lucide/svelte/icons/scale';
	import type { UserShiftMetrics } from '$lib/utils/userProcessing';

	let { metrics }: { metrics: UserShiftMetrics } = $props();

	function getBalanceColor(balance: string) {
		switch (balance) {
			case 'balanced': return 'default';
			case 'uneven': return 'secondary';
			case 'concentrated': return 'destructive';
			default: return 'outline';
		}
	}

	function getBalanceLabel(balance: string) {
		switch (balance) {
			case 'balanced': return 'Well Balanced';
			case 'uneven': return 'Slightly Uneven';
			case 'concentrated': return 'Highly Concentrated';
			default: return balance;
		}
	}
</script>

<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-6 mb-8">
	<!-- Total Assigned Shifts -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Total Assigned Shifts</Card.Title>
			<TargetIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.totalShifts}</div>
			<p class="text-xs text-muted-foreground mt-1">
				Shifts currently assigned to volunteers
			</p>
		</Card.Content>
	</Card.Root>

	<!-- Average Shifts per User -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Average per Volunteer</Card.Title>
			<TrendingUpIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.averageShiftsPerUser}</div>
			<p class="text-xs text-muted-foreground mt-1">
				Average shifts per registered user
			</p>
		</Card.Content>
	</Card.Root>

	<!-- Active Volunteers -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Active Volunteers</Card.Title>
			<UsersIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="text-3xl font-bold">{metrics.usersWithShifts}</div>
			<p class="text-xs text-muted-foreground mt-1">
				Users with assigned shifts
			</p>
		</Card.Content>
	</Card.Root>

	<!-- Workload Balance -->
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-3">
			<Card.Title class="text-sm font-medium">Workload Balance</Card.Title>
			<BalanceIcon class="h-5 w-5 text-muted-foreground" />
		</Card.Header>
		<Card.Content>
			<div class="mb-2">
				<Badge variant={getBalanceColor(metrics.workloadBalance)} class="text-sm">
					{getBalanceLabel(metrics.workloadBalance)}
				</Badge>
			</div>
			<p class="text-xs text-muted-foreground mt-1">
				Distribution of shifts across volunteers
			</p>
		</Card.Content>
	</Card.Root>
</div> 