<script lang="ts">
	import { Trophy, Award, History, Activity } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';
	import LeaderboardView from './LeaderboardView.svelte';
	import AchievementsView from './AchievementsView.svelte';
	import PointsHistoryView from './PointsHistoryView.svelte';
	import ActivityFeedView from './ActivityFeedView.svelte';

	type TabType = 'leaderboard' | 'achievements' | 'history' | 'activity';

	let activeTab: TabType = 'leaderboard';

	function setActiveTab(tab: TabType) {
		activeTab = tab;
	}
</script>

<div class="container mx-auto px-4 py-6 max-w-4xl">
	<!-- Header -->
	<div class="mb-8 text-center">
		<h1 class="text-3xl font-bold text-foreground mb-2">Community Leaderboard</h1>
		<p class="text-muted-foreground">Track your progress and compete with fellow Night Owls</p>
	</div>

	<!-- Tab Navigation -->
	<div class="flex space-x-1 rounded-lg bg-muted p-1 mb-6">
		<Button
			variant={activeTab === 'leaderboard' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1 transition-all"
			onclick={() => setActiveTab('leaderboard')}
		>
			<Trophy class="h-4 w-4 mr-2" />
			Rankings
		</Button>
		<Button
			variant={activeTab === 'achievements' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1 transition-all"
			onclick={() => setActiveTab('achievements')}
		>
			<Award class="h-4 w-4 mr-2" />
			Achievements
		</Button>
		<Button
			variant={activeTab === 'history' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1 transition-all"
			onclick={() => setActiveTab('history')}
		>
			<History class="h-4 w-4 mr-2" />
			History
		</Button>
		<Button
			variant={activeTab === 'activity' ? 'default' : 'ghost'}
			size="sm"
			class="flex-1 transition-all"
			onclick={() => setActiveTab('activity')}
		>
			<Activity class="h-4 w-4 mr-2" />
			Activity
		</Button>
	</div>

	<!-- Tab Content -->
	<div class="transition-all duration-200">
		{#if activeTab === 'leaderboard'}
			<LeaderboardView />
		{:else if activeTab === 'achievements'}
			<AchievementsView />
		{:else if activeTab === 'history'}
			<PointsHistoryView />
		{:else if activeTab === 'activity'}
			<ActivityFeedView />
		{/if}
	</div>
</div>
