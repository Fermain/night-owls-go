<script lang="ts">
	import * as Card from '$lib/components/ui/card';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import MapPinIcon from '@lucide/svelte/icons/map-pin';

	let {
		isLoading = false,
		_latitude,
		_longitude,
		_accuracy,
		_severity,
		_className = ''
	}: {
		isLoading?: boolean;
		_latitude?: number;
		_longitude?: number;
		_accuracy?: number;
		_severity?: number;
		_className?: string;
	} = $props();
</script>

<Card.Root class="p-6">
	<div class="flex items-center gap-2 mb-4">
		<MapPinIcon class="h-5 w-5 text-muted-foreground" />
		<h2 class="text-lg font-semibold">Report Map</h2>
	</div>
	{#if isLoading}
		<Skeleton class="h-64 w-full" />
	{:else}
		<div class="text-center py-16 text-muted-foreground">
			<MapPinIcon class="h-12 w-12 mx-auto mb-4 opacity-50" />
			<p class="text-sm">Interactive map will be displayed here</p>
		</div>
	{/if}
</Card.Root>
