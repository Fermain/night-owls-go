<script lang="ts">
	import { onMount } from 'svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import * as Popover from '$lib/components/ui/popover';
	import WifiIcon from '@lucide/svelte/icons/wifi';
	import WifiOffIcon from '@lucide/svelte/icons/wifi-off';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import { isOnline, getQueuedForms } from '$lib/utils/offline';

	let queuedCount = $state(0);
	let isVisible = $state(false);
	let lastOnline = $state<Date | null>(null);

	onMount(() => {
		// Update queued forms count
		const updateQueuedCount = () => {
			queuedCount = getQueuedForms().length;
		};

		// Track last online time
		const unsubscribe = isOnline.subscribe((online) => {
			if (online) {
				lastOnline = new Date();
			}

			// Show indicator when offline or have queued items
			updateQueuedCount();
			if (!online || queuedCount > 0) {
				isVisible = true;
			}
		});

		// Initial update
		updateQueuedCount();

		// Refresh queued count periodically
		const interval = setInterval(updateQueuedCount, 5000);

		return () => {
			unsubscribe();
			clearInterval(interval);
		};
	});

	function getStatusIcon() {
		if (!$isOnline) {
			return WifiOffIcon;
		}
		if (queuedCount > 0) {
			return RefreshCwIcon;
		}
		return WifiIcon;
	}

	function getStatusColor() {
		if (!$isOnline) {
			return 'destructive';
		}
		if (queuedCount > 0) {
			return 'warning';
		}
		return 'default';
	}

	function getStatusText() {
		if (!$isOnline) {
			return 'Offline';
		}
		if (queuedCount > 0) {
			return `${queuedCount} pending`;
		}
		return 'Online';
	}

	function formatRelativeTime(date: Date | null): string {
		if (!date) return 'Never';

		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / (1000 * 60));

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;

		const diffHours = Math.floor(diffMins / 60);
		if (diffHours < 24) return `${diffHours}h ago`;

		const diffDays = Math.floor(diffHours / 24);
		return `${diffDays}d ago`;
	}
</script>

{#if isVisible}
	<div class="fixed bottom-4 right-4 z-50">
		<Popover.Root>
			<Popover.Trigger>
				<Button
					variant="outline"
					size="sm"
					class="shadow-lg border-2 {getStatusColor() === 'destructive'
						? 'border-red-200 bg-red-50 hover:bg-red-100'
						: getStatusColor() === 'warning'
							? 'border-orange-200 bg-orange-50 hover:bg-orange-100'
							: 'border-green-200 bg-green-50 hover:bg-green-100'}"
				>
					{@const StatusIcon = getStatusIcon()}
					<StatusIcon class="h-4 w-4 mr-2" />
					{getStatusText()}
				</Button>
			</Popover.Trigger>

			<Popover.Content class="w-80" align="end">
				<div class="space-y-4">
					<!-- Network Status -->
					<div class="space-y-2">
						<h4 class="font-medium text-sm">Network Status</h4>
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-2">
								{#if $isOnline}
									<WifiIcon class="h-4 w-4 text-green-600" />
									<span class="text-sm text-green-600">Connected</span>
								{:else}
									<WifiOffIcon class="h-4 w-4 text-red-600" />
									<span class="text-sm text-red-600">Offline</span>
								{/if}
							</div>
							{#if lastOnline}
								<span class="text-xs text-muted-foreground">
									{$isOnline ? 'Connected' : `Last: ${formatRelativeTime(lastOnline)}`}
								</span>
							{/if}
						</div>
					</div>

					<Separator />

					<!-- Offline Capabilities -->
					<div class="space-y-2">
						<h4 class="font-medium text-sm">Offline Features</h4>
						<div class="space-y-2">
							<!-- Emergency Contacts -->
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<PhoneIcon class="h-3 w-3" />
									<span class="text-sm">Emergency Contacts</span>
								</div>
								<CheckCircleIcon class="h-4 w-4 text-green-600" />
							</div>

							<!-- Incident Reports -->
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<FileTextIcon class="h-3 w-3" />
									<span class="text-sm">Incident Reports</span>
								</div>
								<div class="flex items-center gap-1">
									<CheckCircleIcon class="h-4 w-4 text-green-600" />
									{#if queuedCount > 0}
										<Badge variant="outline" class="text-xs">
											{queuedCount} queued
										</Badge>
									{/if}
								</div>
							</div>

							<!-- Messages -->
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<MessageSquareIcon class="h-3 w-3" />
									<span class="text-sm">Messages</span>
								</div>
								<CheckCircleIcon class="h-4 w-4 text-green-600" />
							</div>
						</div>
					</div>

					<!-- Offline Guidance -->
					{#if !$isOnline}
						<div class="p-3 bg-blue-50 rounded-lg border border-blue-200">
							<h5 class="text-sm font-medium text-blue-900 mb-1">Offline Mode</h5>
							<p class="text-xs text-blue-700">
								Form submissions will be queued and synced when connection is restored. Emergency
								calling still works via phone.
							</p>
						</div>
					{/if}

					<!-- Queue Information -->
					{#if queuedCount > 0}
						<div class="p-3 bg-orange-50 rounded-lg border border-orange-200">
							<h5 class="text-sm font-medium text-orange-900 mb-1">Pending Sync</h5>
							<p class="text-xs text-orange-700">
								{queuedCount} form submission{queuedCount > 1 ? 's' : ''} waiting to sync.
								{$isOnline ? 'Will sync automatically...' : 'Will sync when connection returns.'}
							</p>
						</div>
					{/if}

					<!-- Close Button -->
					<div class="flex justify-end">
						<Button variant="ghost" size="sm" onclick={() => (isVisible = false)}>Close</Button>
					</div>
				</div>
			</Popover.Content>
		</Popover.Root>
	</div>
{/if}
