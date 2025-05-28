<script lang="ts">
	import { onMount } from 'svelte';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import HeartHandshakeIcon from '@lucide/svelte/icons/heart-handshake';

	interface EmergencyContact {
		id: number;
		name: string;
		number: string;
		description: string;
		is_default: boolean;
		display_order: number;
	}

	let contacts: EmergencyContact[] = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		try {
			const response = await fetch('/api/emergency-contacts');
			if (!response.ok) {
				throw new Error('Failed to load emergency contacts');
			}
			contacts = await response.json();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load emergency contacts';
		} finally {
			loading = false;
		}
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

	function callNumber(number: string, name: string) {
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
		</Card.Title>
		<Card.Description>Quick access to emergency services and security response</Card.Description>
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
			<div class="text-center py-6 text-muted-foreground">
				<AlertTriangleIcon class="h-8 w-8 mx-auto mb-2 text-red-500" />
				<p class="text-sm">{error}</p>
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
								{#if contact.is_default}
									<Badge variant="secondary" class="text-xs">Default</Badge>
								{/if}
							</div>
							<p class="text-xs text-muted-foreground truncate">{contact.description}</p>
							<p class="text-xs font-mono text-muted-foreground">{contact.number}</p>
						</div>
						<div class="flex-shrink-0">
							<Button
								size="sm"
								variant={contact.is_default ? 'default' : 'outline'}
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
		{/if}
	</Card.Content>
</Card.Root>
