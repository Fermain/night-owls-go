<script lang="ts">
	import UsersIcon from '@lucide/svelte/icons/users';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import type { BroadcastData } from '$lib/schemas/broadcast';
	import {
		formatRelativeTime,
		getAudienceLabel,
		getBroadcastStatusStyle,
		hasDeliveryIssues,
		getDeliveryStatusText
	} from '$lib/utils/broadcasts';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { deleteBroadcast } from '$lib/queries/admin/broadcasts';
	import { toast } from 'svelte-sonner';

	let {
		broadcast,
		onSelect
	}: {
		broadcast: BroadcastData;
		onSelect?: (broadcast: BroadcastData) => void;
	} = $props();

	const queryClient = useQueryClient();

	// Delete mutation
	const deleteMutation = createMutation({
		mutationFn: deleteBroadcast,
		onSuccess: () => {
			toast.success('Broadcast deleted successfully');
			// Refresh the broadcasts list
			queryClient.invalidateQueries({ queryKey: ['broadcasts'] });
		},
		onError: (error: Error) => {
			toast.error(`Failed to delete broadcast: ${error.message}`);
		}
	});

	function handleDelete(event: Event) {
		event.stopPropagation();
		if (confirm('Are you sure you want to delete this broadcast? This action cannot be undone.')) {
			$deleteMutation.mutate(broadcast.broadcast_id);
		}
	}
</script>

<div
	class="p-3 border-b hover:bg-muted/50 transition-all duration-200 cursor-pointer group relative"
	onclick={() => onSelect?.(broadcast)}
	onkeydown={(e) => {
		if (e.key === 'Enter' || e.key === ' ') {
			e.preventDefault();
			onSelect?.(broadcast);
		}
	}}
	role="button"
	tabindex="0"
	aria-label={`View broadcast: ${broadcast.message.substring(0, 50)}...`}
>
	<!-- Delete button - appears on hover -->
	<button
		class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 p-1 rounded hover:bg-destructive/20 text-destructive"
		onclick={handleDelete}
		disabled={$deleteMutation.isPending}
		title="Delete broadcast"
		aria-label="Delete broadcast"
	>
		<TrashIcon class="h-3 w-3" />
	</button>

	<div class="space-y-2 pr-6">
		<p class="text-sm font-medium line-clamp-2">{broadcast.message}</p>
		<div class="flex items-center justify-between text-xs text-muted-foreground">
			<div class="flex items-center gap-2">
				<span class="flex items-center gap-1">
					<UsersIcon class="h-3 w-3" />
					{broadcast.recipient_count}
				</span>
				<span class="flex items-center gap-1">
					<ClockIcon class="h-3 w-3" />
					{formatRelativeTime(broadcast.created_at)}
				</span>
			</div>
			{#if broadcast.push_enabled}
				<span class="text-blue-600 font-medium">SMS</span>
			{/if}
		</div>
		<div class="text-xs text-muted-foreground">
			{getAudienceLabel(broadcast.audience)}
		</div>
		<div class="text-xs">
			<span
				class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium {getBroadcastStatusStyle(
					broadcast.status
				)}"
			>
				{broadcast.status}
			</span>
			{#if hasDeliveryIssues(broadcast)}
				<span class="ml-2 text-orange-600 text-xs">
					{getDeliveryStatusText(broadcast)}
				</span>
			{/if}
		</div>
	</div>
</div>
