<script lang="ts">
	import UsersIcon from '@lucide/svelte/icons/users';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import type { BroadcastData } from '$lib/schemas/broadcast';
	import {
		formatRelativeTime,
		getAudienceLabel,
		getBroadcastStatusStyle,
		hasDeliveryIssues,
		getDeliveryStatusText
	} from '$lib/utils/broadcasts';

	let {
		broadcast,
		onSelect
	}: {
		broadcast: BroadcastData;
		onSelect?: (broadcast: BroadcastData) => void;
	} = $props();
</script>

<div
	class="p-3 border-b hover:bg-muted/50 transition-all duration-200 cursor-pointer group"
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
	<div class="space-y-2">
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
