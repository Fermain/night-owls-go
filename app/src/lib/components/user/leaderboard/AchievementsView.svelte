<script lang="ts">
	import { onMount } from 'svelte';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Progress } from '$lib/components/ui/progress';
	import { Award, Lock, Trophy, Star } from 'lucide-svelte';

	interface Achievement {
		achievement_id: number;
		name: string;
		description: string;
		icon: string;
		shifts_threshold: number | null;
		earned_at: string | null;
	}

	interface AvailableAchievement {
		achievement_id: number;
		name: string;
		description: string;
		icon: string;
		shifts_threshold: number | null;
	}

	interface UserStats {
		total_points: number;
		shift_count: number;
		rank: number;
	}

	let earnedAchievements: Achievement[] = [];
	let availableAchievements: AvailableAchievement[] = [];
	let userStats: UserStats | null = null;
	let loading = true;
	let error = '';

	async function fetchAchievements() {
		try {
			const token = localStorage.getItem('auth_token');
			if (!token) {
				error = 'Authentication required';
				return;
			}

			const headers = {
				Authorization: `Bearer ${token}`,
				'Content-Type': 'application/json'
			};

			const [earnedRes, availableRes, statsRes] = await Promise.all([
				fetch('/api/user/achievements', { headers }),
				fetch('/api/user/achievements/available', { headers }),
				fetch('/api/user/stats', { headers })
			]);

			if (!earnedRes.ok || !availableRes.ok || !statsRes.ok) {
				throw new Error('Failed to fetch achievements');
			}

			earnedAchievements = await earnedRes.json();
			availableAchievements = await availableRes.json();
			userStats = await statsRes.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	}

	function getProgressPercentage(threshold: number, current: number): number {
		return Math.min((current / threshold) * 100, 100);
	}

	function formatDate(dateStr: string): string {
		return new Date(dateStr).toLocaleDateString();
	}

	function getAchievementBadgeVariant(earned: boolean): 'default' | 'secondary' | 'outline' {
		return earned ? 'default' : 'outline';
	}

	onMount(() => {
		fetchAchievements();
	});
</script>

<div class="space-y-6">
	<!-- Progress Overview -->
	{#if userStats}
		<Card>
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Trophy class="h-5 w-5" />
					Achievement Progress
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="grid grid-cols-2 gap-4">
					<div class="text-center">
						<div class="text-3xl font-bold text-primary mb-1">{earnedAchievements.length}</div>
						<div class="text-sm text-muted-foreground">Achievements Earned</div>
					</div>
					<div class="text-center">
						<div class="text-3xl font-bold text-primary mb-1">{userStats.shift_count}</div>
						<div class="text-sm text-muted-foreground">Total Shifts Completed</div>
					</div>
				</div>
			</CardContent>
		</Card>
	{/if}

	<!-- Earned Achievements -->
	{#if loading}
		<div class="flex items-center justify-center py-8">
			<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
		</div>
	{:else if error}
		<div class="text-center py-8 text-destructive">
			{error}
		</div>
	{:else}
		<Card>
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Award class="h-5 w-5" />
					Earned Achievements
				</CardTitle>
			</CardHeader>
			<CardContent>
				{#if earnedAchievements.length === 0}
					<div class="text-center py-8 text-muted-foreground">
						<Award class="h-12 w-12 mx-auto mb-4 opacity-50" />
						<p>No achievements earned yet.</p>
						<p class="text-sm">Complete your first shift to get started!</p>
					</div>
				{:else}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						{#each earnedAchievements as achievement (achievement.achievement_id)}
							<div
								class="flex items-center gap-4 p-4 rounded-lg bg-gradient-to-r from-primary/10 to-primary/5 border border-primary/20"
							>
								<div class="text-4xl">{achievement.icon}</div>
								<div class="flex-1">
									<div class="font-semibold text-primary">{achievement.name}</div>
									<div class="text-sm text-muted-foreground mb-2">{achievement.description}</div>
									<Badge variant="default" class="text-xs">
										Earned {achievement.earned_at ? formatDate(achievement.earned_at) : 'Unknown'}
									</Badge>
								</div>
								<div class="text-primary">
									<Star class="h-6 w-6 fill-current" />
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</CardContent>
		</Card>

		<!-- Available Achievements -->
		{#if availableAchievements.length > 0}
			<Card>
				<CardHeader>
					<CardTitle class="flex items-center gap-2">
						<Lock class="h-5 w-5" />
						Available Achievements
					</CardTitle>
				</CardHeader>
				<CardContent>
					<div class="space-y-4">
						{#each availableAchievements as achievement (achievement.achievement_id)}
							{@const current = userStats?.shift_count || 0}
							{@const threshold = achievement.shifts_threshold || 0}
							{@const progress = getProgressPercentage(threshold, current)}

							<div class="flex items-center gap-4 p-4 rounded-lg bg-muted/50 border">
								<div class="text-4xl opacity-50">{achievement.icon}</div>
								<div class="flex-1">
									<div class="font-semibold">{achievement.name}</div>
									<div class="text-sm text-muted-foreground mb-3">{achievement.description}</div>

									{#if achievement.shifts_threshold}
										<div class="space-y-2">
											<div class="flex justify-between text-sm">
												<span>Progress: {current} / {threshold} shifts</span>
												<span>{progress.toFixed(0)}%</span>
											</div>
											<Progress value={progress} class="h-2" />
										</div>
									{/if}
								</div>
								<div class="text-muted-foreground">
									<Lock class="h-6 w-6" />
								</div>
							</div>
						{/each}
					</div>
				</CardContent>
			</Card>
		{/if}

		<!-- Achievement Guide -->
		<Card>
			<CardHeader>
				<CardTitle class="flex items-center gap-2">
					<Star class="h-5 w-5" />
					How to Earn Achievements
				</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-3 text-sm">
					<div class="flex items-start gap-3">
						<div class="text-2xl mt-1">🦜</div>
						<div>
							<div class="font-medium">Owlet</div>
							<div class="text-muted-foreground">Complete your first shift to become an Owlet!</div>
						</div>
					</div>
					<div class="flex items-start gap-3">
						<div class="text-2xl mt-1">🦉</div>
						<div>
							<div class="font-medium">Solid Owl</div>
							<div class="text-muted-foreground">
								Complete 20 shifts to prove your dedication to the community.
							</div>
						</div>
					</div>
					<div class="flex items-start gap-3">
						<div class="text-2xl mt-1">🦅</div>
						<div>
							<div class="font-medium">Wise Owl</div>
							<div class="text-muted-foreground">
								Complete 50 shifts and become a community guardian veteran.
							</div>
						</div>
					</div>
					<div class="flex items-start gap-3">
						<div class="text-2xl mt-1">🔥</div>
						<div>
							<div class="font-medium">Super Owl</div>
							<div class="text-muted-foreground">
								Complete 100 shifts - the ultimate community protector!
							</div>
						</div>
					</div>
				</div>
			</CardContent>
		</Card>
	{/if}
</div>
