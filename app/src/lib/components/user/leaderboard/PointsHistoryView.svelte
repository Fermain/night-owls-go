<script lang="ts">
	import { onMount } from 'svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { authenticatedFetch } from '$lib/utils/api';

	interface PointsHistoryEntry {
		history_id: number;
		booking_id: number | null;
		points_awarded: number;
		reason: string;
		multiplier: number;
		awarded_at: string;
	}

	interface ReasonInfo {
		label: string;
		color: string;
		description: string;
	}

	let pointsHistory: PointsHistoryEntry[] = [];
	let loading = true;
	let error = '';
	let limit = 20;

	async function fetchPointsHistory() {
		try {
			const response = await authenticatedFetch(`/api/user/points/history?limit=${limit}`);
			pointsHistory = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function getReasonInfo(reason: string): ReasonInfo {
		const reasonMap: Record<string, ReasonInfo> = {
			shift_checkin: {
				label: 'Check-in',
				color: 'text-blue-600',
				description: 'Shift check-in'
			},
			shift_completion: {
				label: 'Completion',
				color: 'text-green-600',
				description: 'Shift completed'
			},
			early_checkin: {
				label: 'Early Check-in',
				color: 'text-purple-600',
				description: 'Early arrival bonus'
			},
			weekend_bonus: {
				label: 'Weekend Bonus',
				color: 'text-orange-600',
				description: 'Weekend shift'
			},
			late_night_bonus: {
				label: 'Night Bonus',
				color: 'text-indigo-600',
				description: 'Late night shift'
			},
			frequency_bonus: {
				label: 'Frequency Bonus',
				color: 'text-pink-600',
				description: 'Multiple shifts'
			},
			level2_report: {
				label: 'Incident Report',
				color: 'text-red-600',
				description: 'Serious incident reported'
			},
			report_filed: {
				label: 'Report',
				color: 'text-green-500',
				description: 'Report filed'
			}
		};
		return (
			reasonMap[reason] || {
				label: 'Points',
				color: 'text-gray-600',
				description: 'Points awarded'
			}
		);
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString();
	}

	function loadMore() {
		limit += 20;
		fetchPointsHistory();
	}

	function getTotalPoints(): number {
		return pointsHistory.reduce((total, entry) => total + entry.points_awarded, 0);
	}

	onMount(() => {
		fetchPointsHistory();
	});
</script>

<div class="space-y-4">
	<!-- Summary -->
	<div class="grid grid-cols-2 gap-3">
		<div class="bg-muted/50 rounded-lg p-3 text-center">
			<div class="text-lg font-bold">{getTotalPoints()}</div>
			<div class="text-xs text-muted-foreground">Total Points</div>
		</div>
		<div class="bg-muted/50 rounded-lg p-3 text-center">
			<div class="text-lg font-bold">{pointsHistory.length}</div>
			<div class="text-xs text-muted-foreground">Activities</div>
		</div>
	</div>

	<!-- Points History -->
	{#if loading}
		<div class="flex justify-center py-8">
			<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-primary"></div>
		</div>
	{:else if error}
		<div class="text-center py-4 text-sm text-destructive">
			{error}
		</div>
	{:else if pointsHistory.length === 0}
		<div class="text-center py-8 text-sm text-muted-foreground">No points history yet</div>
	{:else}
		<div class="space-y-2">
			{#each pointsHistory as entry (entry.history_id)}
				{@const reasonInfo = getReasonInfo(entry.reason)}
				<div class="flex items-center gap-3 p-2 rounded-lg bg-background border">
					<!-- Color indicator -->
					<div class="w-3 h-3 rounded-full {reasonInfo.color.replace('text-', 'bg-')}"></div>

					<!-- Content -->
					<div class="flex-1 min-w-0">
						<div class="font-medium text-sm">{reasonInfo.label}</div>
						<div class="text-xs text-muted-foreground">{formatDate(entry.awarded_at)}</div>
						{#if entry.booking_id}
							<div class="text-xs text-muted-foreground">Shift #{entry.booking_id}</div>
						{/if}
					</div>

					<!-- Points -->
					<div class="text-right">
						<div class="font-bold text-sm {entry.points_awarded >= 0 ? 'text-green-600' : 'text-red-600'}">
							{entry.points_awarded >= 0 ? '+' : ''}{entry.points_awarded}
						</div>
						{#if entry.multiplier !== 1}
							<Badge variant="outline" class="text-xs py-0 px-1">
								{entry.multiplier}x
							</Badge>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		<!-- Load More -->
		<div class="text-center pt-2">
			<Button variant="outline" size="sm" onclick={loadMore} class="w-full">Load More</Button>
		</div>
	{/if}
</div>
