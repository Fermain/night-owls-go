<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import {
		History,
		TrendingUp,
		Calendar,
		Plus,
		Clock,
		Star,
		MapPin,
		Moon,
		Sun
	} from 'lucide-svelte';
	import { authenticatedFetch } from '$lib/utils/api';

	interface PointsHistoryEntry {
		history_id: number;
		booking_id: number | null;
		points_awarded: number;
		reason: string;
		multiplier: number;
		awarded_at: string;
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

	function getReasonInfo(reason: string) {
		const reasonMap = {
			shift_checkin: {
				icon: Clock,
				label: 'Shift Check-in',
				description: 'Checked in to shift on time',
				color: 'text-blue-600'
			},
			shift_completion: {
				icon: Star,
				label: 'Shift Completion',
				description: 'Completed full shift with report',
				color: 'text-green-600'
			},
			report_filed: {
				icon: MapPin,
				label: 'Report Filed',
				description: 'Filed incident report',
				color: 'text-orange-600'
			},
			early_checkin: {
				icon: Sun,
				label: 'Early Check-in',
				description: 'Checked in 15+ minutes early',
				color: 'text-yellow-600'
			},
			level2_report: {
				icon: TrendingUp,
				label: 'Level 2 Report',
				description: 'Reported serious incident',
				color: 'text-red-600'
			},
			weekend_bonus: {
				icon: Calendar,
				label: 'Weekend Bonus',
				description: 'Weekend shift completed',
				color: 'text-purple-600'
			},
			late_night_bonus: {
				icon: Moon,
				label: 'Late Night Bonus',
				description: 'Night shift (10 PM - 5 AM)',
				color: 'text-indigo-600'
			},
			frequency_bonus: {
				icon: TrendingUp,
				label: 'Frequency Bonus',
				description: 'Multiple shifts this month',
				color: 'text-pink-600'
			}
		};

		return (
			reasonMap[reason as keyof typeof reasonMap] || {
				icon: Plus,
				label: reason.replace('_', ' ').replace(/\b\w/g, (l) => l.toUpperCase()),
				description: 'Points awarded',
				color: 'text-gray-600'
			}
		);
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
		const diffDays = Math.floor(diffHours / 24);

		if (diffHours < 1) {
			return 'Just now';
		} else if (diffHours < 24) {
			return `${diffHours}h ago`;
		} else if (diffDays < 7) {
			return `${diffDays}d ago`;
		} else {
			return date.toLocaleDateString();
		}
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

<div class="space-y-6">
	<!-- Summary Card -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<History class="h-5 w-5" />
				Points History Summary
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-2 gap-4 text-center">
				<div>
					<div class="text-3xl font-bold text-primary">{getTotalPoints()}</div>
					<div class="text-sm text-muted-foreground">Total Points Shown</div>
				</div>
				<div>
					<div class="text-3xl font-bold text-primary">{pointsHistory.length}</div>
					<div class="text-sm text-muted-foreground">Recent Activities</div>
				</div>
			</div>
		</CardContent>
	</Card>

	<!-- Points History List -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<TrendingUp class="h-5 w-5" />
				Recent Point Awards
			</CardTitle>
		</CardHeader>
		<CardContent>
			{#if loading}
				<div class="flex items-center justify-center py-8">
					<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
				</div>
			{:else if error}
				<div class="text-center py-8 text-destructive">
					{error}
				</div>
			{:else if pointsHistory.length === 0}
				<div class="text-center py-8 text-muted-foreground">
					<History class="h-12 w-12 mx-auto mb-4 opacity-50" />
					<p>No points history yet.</p>
					<p class="text-sm">Check in to your first shift to start earning points!</p>
				</div>
			{:else}
				<div class="space-y-3">
					{#each pointsHistory as entry (entry.history_id)}
						{@const reasonInfo = getReasonInfo(entry.reason)}
						<div
							class="flex items-center justify-between p-3 rounded-lg bg-muted/50 hover:bg-muted transition-colors"
						>
							<div class="flex items-center gap-3">
								<div class="flex items-center justify-center w-10 h-10 rounded-full bg-background">
									<svelte:component this={reasonInfo.icon} class="h-5 w-5 {reasonInfo.color}" />
								</div>

								<div>
									<div class="font-medium">{reasonInfo.label}</div>
									<div class="text-sm text-muted-foreground">{reasonInfo.description}</div>
									{#if entry.booking_id}
										<div class="text-xs text-muted-foreground">Shift #{entry.booking_id}</div>
									{/if}
								</div>
							</div>

							<div class="text-right">
								<div class="font-bold text-lg text-green-600">
									+{entry.points_awarded}
								</div>
								<div class="text-xs text-muted-foreground">
									{formatDate(entry.awarded_at)}
								</div>
								{#if entry.multiplier !== 1}
									<Badge variant="secondary" class="text-xs mt-1">
										{entry.multiplier}x multiplier
									</Badge>
								{/if}
							</div>
						</div>
					{/each}
				</div>

				{#if pointsHistory.length >= limit}
					<div class="text-center mt-6">
						<Button variant="outline" onclick={loadMore} disabled={loading}>
							Load More History
						</Button>
					</div>
				{/if}
			{/if}
		</CardContent>
	</Card>

	<!-- Points Guide -->
	<Card>
		<CardHeader>
			<CardTitle class="flex items-center gap-2">
				<Star class="h-5 w-5" />
				How You Earn Points
			</CardTitle>
		</CardHeader>
		<CardContent>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
				<div class="space-y-3">
					<div class="flex items-center gap-3">
						<Clock class="h-4 w-4 text-blue-600" />
						<div>
							<div class="font-medium">Check-in (10 pts)</div>
							<div class="text-muted-foreground">For checking in to your shift</div>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<Star class="h-4 w-4 text-green-600" />
						<div>
							<div class="font-medium">Completion (15 pts)</div>
							<div class="text-muted-foreground">For completing shift with report</div>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<MapPin class="h-4 w-4 text-orange-600" />
						<div>
							<div class="font-medium">Report Filed (5 pts)</div>
							<div class="text-muted-foreground">For filing any incident report</div>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<TrendingUp class="h-4 w-4 text-red-600" />
						<div>
							<div class="font-medium">Level 2 Report (+10 pts)</div>
							<div class="text-muted-foreground">Extra for serious incidents</div>
						</div>
					</div>
				</div>
				<div class="space-y-3">
					<div class="flex items-center gap-3">
						<Sun class="h-4 w-4 text-yellow-600" />
						<div>
							<div class="font-medium">Early Check-in (+3 pts)</div>
							<div class="text-muted-foreground">15+ minutes before shift</div>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<Calendar class="h-4 w-4 text-purple-600" />
						<div>
							<div class="font-medium">Weekend Bonus (+5 pts)</div>
							<div class="text-muted-foreground">Saturday/Sunday shifts</div>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<Moon class="h-4 w-4 text-indigo-600" />
						<div>
							<div class="font-medium">Late Night (+3 pts)</div>
							<div class="text-muted-foreground">Shifts between 10 PM - 5 AM</div>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<TrendingUp class="h-4 w-4 text-pink-600" />
						<div>
							<div class="font-medium">Frequency Bonus (+10 pts)</div>
							<div class="text-muted-foreground">Multiple shifts per month</div>
						</div>
					</div>
				</div>
			</div>
		</CardContent>
	</Card>
</div>
