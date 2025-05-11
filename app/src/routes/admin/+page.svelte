<script lang="ts">
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
</script>

<svelte:head>
	<title>Admin Overview</title>
</svelte:head>

<div class="p-4 md:p-8">
	<h1 class="text-2xl font-semibold mb-6">Admin Overview</h1>
	<p class="mb-6 text-muted-foreground">
		High-level statistics and quick actions will be displayed here.
	</p>
	<div class="space-y-4">
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each [1, 2, 3] as i (i)}
				<div class="p-4 border rounded-lg">
					<Skeleton class="h-6 w-3/4 mb-2" />
					<Skeleton class="h-10 w-1/2 mb-4" />
					<Skeleton class="h-4 w-full" />
					<Skeleton class="h-4 w-5/6 mt-1" />
				</div>
			{/each}
		</div>
		<div class="p-4 border rounded-lg">
			<Skeleton class="h-8 w-1/4 mb-4" />
			<Skeleton class="h-48 w-full" />
		</div>
	</div>
	<p class="mt-8 text-sm text-muted-foreground">
		Use the navigation panel on the left to access different admin sections.
	</p>
</div>
