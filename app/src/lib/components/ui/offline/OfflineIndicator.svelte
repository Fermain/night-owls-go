<script lang="ts">
	import { onMount } from 'svelte';
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import { Separator } from '$lib/components/ui/separator';
	import * as Popover from '$lib/components/ui/popover';
	import WifiIcon from '@lucide/svelte/icons/wifi';
	import WifiOffIcon from '@lucide/svelte/icons/wifi-off';
	import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
	import CloudOffIcon from '@lucide/svelte/icons/cloud-off';
	import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
	import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
	import LoaderIcon from '@lucide/svelte/icons/loader-2';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import { offlineService } from '$lib/services/offlineService';

	let offlineState = $state<{
		isOnline: boolean;
		lastOnline: string | null;
		emergencyContactsAvailable: boolean;
		queuedReports: number;
		syncInProgress: boolean;
		lastSync: string | null;
	}>({
		isOnline: true,
		lastOnline: null,
		emergencyContactsAvailable: false,
		queuedReports: 0,
		syncInProgress: false,
		lastSync: null
	});

	let isVisible = $state(false);

	onMount(() => {
		let unsubscribe: (() => void) | undefined;

		// Initialize offline service
		const initializeService = async () => {
			try {
				await offlineService.init();

				// Subscribe to offline state changes
				unsubscribe = offlineService.state.subscribe((state) => {
					offlineState = state;

					// Auto-show indicator when going offline or when there are queued items
					if (!state.isOnline || state.queuedReports > 0) {
						isVisible = true;
					}
				});
			} catch (error) {
				console.error('Failed to initialize offline service:', error);
			}
		};

		initializeService();

		return () => {
			if (unsubscribe) {
				unsubscribe();
			}
		};
	});

	function getStatusIcon() {
		if (offlineState.syncInProgress) {
			return LoaderIcon;
		}
		if (!offlineState.isOnline) {
			return WifiOffIcon;
		}
		if (offlineState.queuedReports > 0) {
			return RefreshCwIcon;
		}
		return WifiIcon;
	}

	function getStatusColor() {
		if (!offlineState.isOnline) {
			return 'destructive';
		}
		if (offlineState.queuedReports > 0) {
			return 'warning';
		}
		return 'default';
	}

	function getStatusText() {
		if (offlineState.syncInProgress) {
			return 'Syncing...';
		}
		if (!offlineState.isOnline) {
			return 'Offline';
		}
		if (offlineState.queuedReports > 0) {
			return `${offlineState.queuedReports} pending`;
		}
		return 'Online';
	}

	function formatRelativeTime(timestamp: string | null): string {
		if (!timestamp) return 'Never';

		const now = new Date();
		const time = new Date(timestamp);
		const diffMs = now.getTime() - time.getTime();
		const diffMins = Math.floor(diffMs / (1000 * 60));

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;

		const diffHours = Math.floor(diffMins / 60);
		if (diffHours < 24) return `${diffHours}h ago`;

		const diffDays = Math.floor(diffHours / 24);
		return `${diffDays}d ago`;
	}

	async function handleSync() {
		try {
			await offlineService.syncAllData();
		} catch (error) {
			console.error('Manual sync failed:', error);
		}
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
					{#if offlineState.syncInProgress}
						<LoaderIcon class="h-4 w-4 mr-2 animate-spin" />
					{:else}
						<svelte:component this={getStatusIcon()} class="h-4 w-4 mr-2" />
					{/if}
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
								{#if offlineState.isOnline}
									<WifiIcon class="h-4 w-4 text-green-600" />
									<span class="text-sm text-green-600">Connected</span>
								{:else}
									<WifiOffIcon class="h-4 w-4 text-red-600" />
									<span class="text-sm text-red-600">Offline</span>
								{/if}
							</div>
							{#if offlineState.lastOnline}
								<span class="text-xs text-muted-foreground">
									{offlineState.isOnline
										? 'Connected'
										: `Last: ${formatRelativeTime(offlineState.lastOnline)}`}
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
								{#if offlineState.emergencyContactsAvailable}
									<CheckCircleIcon class="h-4 w-4 text-green-600" />
								{:else}
									<AlertCircleIcon class="h-4 w-4 text-orange-600" />
								{/if}
							</div>

							<!-- Incident Reports -->
							<div class="flex items-center justify-between">
								<div class="flex items-center gap-2">
									<FileTextIcon class="h-3 w-3" />
									<span class="text-sm">Incident Reports</span>
								</div>
								<div class="flex items-center gap-1">
									<CheckCircleIcon class="h-4 w-4 text-green-600" />
									{#if offlineState.queuedReports > 0}
										<Badge variant="outline" class="text-xs">
											{offlineState.queuedReports} queued
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

					<Separator />

					<!-- Sync Status -->
					<div class="space-y-2">
						<div class="flex items-center justify-between">
							<h4 class="font-medium text-sm">Sync Status</h4>
							{#if offlineState.syncInProgress}
								<LoaderIcon class="h-4 w-4 animate-spin text-blue-600" />
							{:else if offlineState.isOnline}
								<Button
									variant="ghost"
									size="sm"
									onclick={handleSync}
									disabled={offlineState.syncInProgress}
								>
									<RefreshCwIcon class="h-3 w-3 mr-1" />
									Sync Now
								</Button>
							{:else}
								<CloudOffIcon class="h-4 w-4 text-gray-400" />
							{/if}
						</div>

						{#if offlineState.lastSync}
							<p class="text-xs text-muted-foreground">
								Last sync: {formatRelativeTime(offlineState.lastSync)}
							</p>
						{/if}
					</div>

					<!-- Offline Guidance -->
					{#if !offlineState.isOnline}
						<div class="p-3 bg-blue-50 rounded-lg border border-blue-200">
							<h5 class="text-sm font-medium text-blue-900 mb-1">Offline Mode</h5>
							<p class="text-xs text-blue-700">
								{#if offlineState.emergencyContactsAvailable}
									Emergency contacts and reporting are available offline. Reports will sync when
									connection is restored.
								{:else}
									Limited offline functionality. Emergency calling still works via phone.
								{/if}
							</p>
						</div>
					{/if}

					<!-- Queue Information -->
					{#if offlineState.queuedReports > 0}
						<div class="p-3 bg-orange-50 rounded-lg border border-orange-200">
							<h5 class="text-sm font-medium text-orange-900 mb-1">Pending Sync</h5>
							<p class="text-xs text-orange-700">
								{offlineState.queuedReports} incident report{offlineState.queuedReports > 1
									? 's'
									: ''} waiting to sync.
								{offlineState.isOnline
									? 'Syncing automatically...'
									: 'Will sync when connection returns.'}
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
