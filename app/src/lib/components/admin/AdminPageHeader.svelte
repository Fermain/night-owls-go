<script lang="ts">
	let {
		icon,
		heading,
		subheading,
		className = ''
	}: {
		icon?: any;
		heading: string;
		subheading?: string;
		className?: string;
	} = $props();
</script>

<div class="mb-6 {className}">
	<h1 class="text-2xl font-bold flex items-center gap-2 mb-3">
		{#if icon}
			<svelte:component this={icon} class="h-6 w-6" />
		{/if}
		{heading}
	</h1>
	{#if subheading}
		<p class="text-muted-foreground text-lg">
			{subheading}
		</p>
	{/if}
</div> 