<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import HeartHandshakeIcon from '@lucide/svelte/icons/heart-handshake';
	import WifiOffIcon from '@lucide/svelte/icons/wifi-off';
	import CloudOffIcon from '@lucide/svelte/icons/cloud-off';
	import { offlineService, type EmergencyContact } from '$lib/services/offlineService';

	let contacts: EmergencyContact[] = [];
	let loading = true;
	let error = '';
	let isOnline = true;
	let usingCachedData = false;

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
						} catch (fetchError) {
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
	}
</script>

<Card.Root class="w-full">
	<Card.Header class="pb-3">
		<Card.Title class="flex items-center gap-2">
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
		</Card.Title>
		<Card.Description>
			Quick access to emergency services and security response
			{#if usingCachedData}
				<span class="text-orange-600"> - Using cached data</span>
			{/if}
		</Card.Description>
	</Card.Header>
	<Card.Content class="pt-0">
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
			<div class="space-y-3">
				{#each contacts as contact (contact.id)}
					{@const IconComponent = getContactIcon(contact.name)}
					<div
						class="flex items-center gap-3 p-3 bg-muted/50 rounded-lg hover:bg-muted/70 transition-colors"
					>
						<div class="flex-shrink-0">
							<div
								class="w-10 h-10 bg-background rounded-full flex items-center justify-center border"
							>
								<svelte:component
									this={IconComponent}
									class="h-5 w-5 {getContactColor(contact.name)}"
								/>
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
						<strong>Note:</strong> These contacts are cached from your last connection. Phone calls will
						work normally.
					</p>
				</div>
			{/if}
		{/if}
	</Card.Content>
</Card.Root>
