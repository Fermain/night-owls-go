<script lang="ts">
	import { onMount } from 'svelte';
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import HeartHandshakeIcon from '@lucide/svelte/icons/heart-handshake';
	import WifiOffIcon from '@lucide/svelte/icons/wifi-off';
	import CloudOffIcon from '@lucide/svelte/icons/cloud-off';
	import { offlineService, type EmergencyContact } from '$lib/services/offlineService';

	let { open = $bindable(false) }: { open?: boolean } = $props();

	let contacts = $state<EmergencyContact[]>([]);
	let loading = $state(true);
	let error = $state('');
	let isOnline = $state(true);
	let usingCachedData = $state(false);
	let showEmergencyConfirm = $state(false);

	onMount(() => {
		let unsubscribe: (() => void) | undefined;

		const initializeService = async () => {
			try {
				// Initialize offline service
				await offlineService.init();

				// Subscribe to online/offline status
				unsubscribe = offlineService.state.subscribe((state) => {
					isOnline = state.isOnline;
				});

				// Try to get emergency contacts (offline-first approach)
				contacts = await offlineService.getEmergencyContacts();

				if (contacts.length > 0) {
					// We have cached data
					usingCachedData = !isOnline;
					loading = false;

					// If online, try to refresh in the background
					if (isOnline) {
						try {
							await offlineService.cacheEmergencyContacts();
							// Refresh contacts after cache update
							contacts = await offlineService.getEmergencyContacts();
							usingCachedData = false;
						} catch (refreshError) {
							console.warn('Failed to refresh emergency contacts:', refreshError);
							// Keep using cached data
						}
					}
				} else {
					// No cached data available
					if (isOnline) {
						// Try to fetch from API
						try {
							await offlineService.cacheEmergencyContacts();
							contacts = await offlineService.getEmergencyContacts();
						} catch (_fetchError) {
							error = 'Failed to load emergency contacts';
						}
					} else {
						error = 'Emergency contacts not available offline. Connect to internet to download.';
					}
				}
			} catch (err) {
				error = err instanceof Error ? err.message : 'Failed to load emergency contacts';
			} finally {
				loading = false;
			}
		};

		// Start the async initialization
		initializeService();

		// Return cleanup function
		return () => {
			if (unsubscribe) {
				unsubscribe();
			}
		};
	});

	function getContactIcon(name: string) {
		const nameLower = name.toLowerCase();
		if (nameLower.includes('rusa') || nameLower.includes('security')) {
			return ShieldIcon;
		}
		if (nameLower.includes('police') || nameLower.includes('saps')) {
			return AlertTriangleIcon;
		}
		if (
			nameLower.includes('medical') ||
			nameLower.includes('er24') ||
			nameLower.includes('ambulance')
		) {
			return HeartHandshakeIcon;
		}
		return PhoneIcon;
	}

	function getContactColor(name: string) {
		const nameLower = name.toLowerCase();
		if (nameLower.includes('rusa') || nameLower.includes('security')) {
			return 'text-blue-600 dark:text-blue-400';
		}
		if (nameLower.includes('police') || nameLower.includes('saps')) {
			return 'text-red-600 dark:text-red-400';
		}
		if (
			nameLower.includes('medical') ||
			nameLower.includes('er24') ||
			nameLower.includes('ambulance')
		) {
			return 'text-green-600 dark:text-green-400';
		}
		return 'text-gray-600 dark:text-gray-400';
	}

	function callNumber(number: string, _name: string) {
		// For mobile devices, use tel: protocol
		if (typeof window !== 'undefined') {
			window.location.href = `tel:${number}`;
		}
		// Close dialog after initiating call
		open = false;
	}

	function confirmEmergencyCall() {
		window.location.href = 'tel:999';
		showEmergencyConfirm = false;
		open = false;
	}

	function cancelEmergencyCall() {
		showEmergencyConfirm = false;
	}

	function _callEmergency(number: string) {
		// Implementation for emergency calling
		window.location.href = `tel:${number}`;
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<AlertTriangleIcon class="h-5 w-5 text-red-500" />
				Emergency Contacts
				{#if !isOnline && usingCachedData}
					<Badge variant="outline" class="text-xs bg-orange-50 border-orange-200 text-orange-700">
						<WifiOffIcon class="h-3 w-3 mr-1" />
						Offline
					</Badge>
				{:else if !isOnline}
					<Badge variant="outline" class="text-xs bg-red-50 border-red-200 text-red-700">
						<CloudOffIcon class="h-3 w-3 mr-1" />
						Unavailable
					</Badge>
				{/if}
			</Dialog.Title>
		</Dialog.Header>

		<div class="border-t pt-4">
			{#if loading}
				<div class="space-y-3">
					{#each Array(3) as _, i (i)}
						<div class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg animate-pulse">
							<div class="w-10 h-10 bg-muted rounded-full"></div>
							<div class="flex-grow space-y-2">
								<div class="h-4 bg-muted rounded w-24"></div>
								<div class="h-3 bg-muted rounded w-32"></div>
							</div>
							<div class="w-16 h-8 bg-muted rounded"></div>
						</div>
					{/each}
				</div>
			{:else if error}
				<div class="text-center py-6">
					<AlertTriangleIcon class="h-8 w-8 mx-auto mb-2 text-red-500" />
					<p class="text-sm text-muted-foreground mb-3">{error}</p>
					{#if !isOnline}
						<div class="p-3 bg-blue-50 rounded-lg border border-blue-200 text-left">
							<p class="text-xs text-blue-700">
								<strong>Emergency calling still works:</strong> You can dial emergency numbers directly
								using your phone's dialer.
							</p>
						</div>
					{/if}
				</div>
			{:else if contacts.length === 0}
				<div class="text-center py-6 text-muted-foreground">
					<PhoneIcon class="h-8 w-8 mx-auto mb-2" />
					<p class="text-sm">No emergency contacts configured</p>
				</div>
			{:else}
				<div class="space-y-3 overflow-y-auto">
					{#each contacts as contact (contact.id)}
						{@const IconComponent = getContactIcon(contact.name)}
						<div
							class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg hover:bg-muted/70 transition-colors"
						>
							<div class="flex-shrink-0">
								<div
									class="w-10 h-10 bg-background rounded-full flex items-center justify-center border"
								>
									<IconComponent class="h-5 w-5 {getContactColor(contact.name)}" />
								</div>
							</div>
							<div class="flex-grow min-w-0">
								<div class="flex items-center gap-2 mb-1">
									<p class="font-medium text-sm truncate">{contact.name}</p>
									{#if contact.isDefault}
										<Badge variant="secondary" class="text-xs">Default</Badge>
									{/if}
								</div>
								<p class="text-xs text-muted-foreground truncate">{contact.description}</p>
								<p class="text-xs font-mono text-muted-foreground">{contact.number}</p>
							</div>
							<div class="flex-shrink-0">
								<Button
									size="sm"
									variant={contact.isDefault ? 'default' : 'outline'}
									class="h-8 px-3"
									onclick={() => callNumber(contact.number, contact.name)}
								>
									<PhoneIcon class="h-3 w-3 mr-1" />
									Call
								</Button>
							</div>
						</div>
					{/each}
				</div>

				{#if usingCachedData}
					<div class="mt-4 p-3 bg-orange-50 rounded-lg border border-orange-200">
						<p class="text-xs text-orange-700">
							<strong>Note:</strong> These contacts are cached from your last connection. Phone calls
							will work normally.
						</p>
					</div>
				{/if}
			{/if}
		</div>
	</Dialog.Content>
</Dialog.Root>

<!-- Emergency Call Confirmation Dialog -->
<Dialog.Root bind:open={showEmergencyConfirm}>
	<Dialog.Content class="sm:max-w-md">
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<AlertTriangleIcon class="h-5 w-5 text-red-500" />
				Emergency Call
			</Dialog.Title>
			<Dialog.Description>
				This will call emergency services immediately. Are you sure you want to proceed?
			</Dialog.Description>
		</Dialog.Header>

		<div
			class="p-3 bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 rounded-lg"
		>
			<p class="text-sm text-red-700 dark:text-red-300">
				⚠️ This action will initiate an emergency call to 999. Only proceed if this is a genuine
				emergency.
			</p>
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={cancelEmergencyCall}>Cancel</Button>
			<Button variant="destructive" onclick={confirmEmergencyCall}>Call Emergency Services</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
