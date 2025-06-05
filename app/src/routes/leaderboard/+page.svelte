<script lang="ts">
	import LeaderboardContainer from '$lib/components/user/leaderboard/LeaderboardContainer.svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		// Check authentication
		const token = localStorage.getItem('auth_token');
		if (!token) {
			goto('/login');
			return;
		}
	});
</script>

<svelte:head>
	<title>Leaderboard - Night Owls</title>
	<meta
		name="description"
		content="View community rankings, achievements, and your points history"
	/>
</svelte:head>

<LeaderboardContainer />
