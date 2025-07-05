<script lang="ts">
	import LeaderboardContainer from '$lib/components/user/leaderboard/LeaderboardContainer.svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { generateOpenGraphTags } from '$lib/utils/opengraph';

	// OpenGraph tags for leaderboard
	const ogTags = generateOpenGraphTags({
		title: 'Community Leaderboard - Mount Moreland Night Owls',
		description:
			"View community rankings, achievements, and patrol contributions. See who's leading the charge in keeping our neighborhood safe.",
		type: 'website'
	});

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
	<title>{ogTags.title}</title>
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.description}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogImageAlt}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogType}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.ogSiteName}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterCard}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterTitle}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterDescription}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImage}
	<!-- eslint-disable-next-line svelte/no-at-html-tags -->
	{@html ogTags.twitterImageAlt}
</svelte:head>

<LeaderboardContainer />
