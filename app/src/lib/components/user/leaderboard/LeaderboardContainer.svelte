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

<div class="container mx-auto px-4 py-4">
	<!-- Header - Compact -->
	<div class="mb-4">
		<h1 class="text-xl font-bold mb-1">Leaderboard</h1>
		<p class="text-sm text-muted-foreground">Community rankings and activity</p>
	</div>

	<!-- Tab Navigation - Mobile Optimized -->
	<div class="grid grid-cols-4 rounded-lg bg-muted p-1 mb-4">
		<Button
			variant={activeTab === 'leaderboard' ? 'default' : 'ghost'}
			size="sm"
			class="h-8 text-xs"
			onclick={() => setActiveTab('leaderboard')}
		>
			<Trophy class="h-3 w-3 mr-1" />
			Ranks
		</Button>
		<Button
			variant={activeTab === 'achievements' ? 'default' : 'ghost'}
			size="sm"
			class="h-8 text-xs"
			onclick={() => setActiveTab('achievements')}
		>
			<Award class="h-3 w-3 mr-1" />
			Badges
		</Button>
		<Button
			variant={activeTab === 'history' ? 'default' : 'ghost'}
			size="sm"
			class="h-8 text-xs"
			onclick={() => setActiveTab('history')}
		>
			<History class="h-3 w-3 mr-1" />
			History
		</Button>
		<Button
			variant={activeTab === 'activity' ? 'default' : 'ghost'}
			size="sm"
			class="h-8 text-xs"
			onclick={() => setActiveTab('activity')}
		>
			<Activity class="h-3 w-3 mr-1" />
			Feed
		</Button>
	</div>

	<!-- Tab Content -->
	<div>
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
