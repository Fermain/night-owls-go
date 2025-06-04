<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import StarIcon from '@lucide/svelte/icons/star';

	// Use our domain types
	import type { EmergencyContact } from '$lib/types/domain';

	// Phone number formatting utility
	function formatPhoneNumber(number: string): string {
		// Remove all non-digit characters
		const cleaned = number.replace(/\D/g, '');

		// Basic South African number formatting
		if (cleaned.length === 10 && cleaned.startsWith('0')) {
			// Format: 012 345 6789
			return `${cleaned.slice(0, 3)} ${cleaned.slice(3, 6)} ${cleaned.slice(6)}`;
		}

		if (cleaned.length === 9 && !cleaned.startsWith('0')) {
			// Format: 12 345 6789 (without leading 0)
			return `${cleaned.slice(0, 2)} ${cleaned.slice(2, 5)} ${cleaned.slice(5)}`;
		}

		// Return original if no pattern matches
		return number;
	}

	let {
		contact,
		isSelected = false,
		onSelect
	}: {
		contact: EmergencyContact;
		isSelected?: boolean;
		onSelect: (contact: EmergencyContact) => void;
	} = $props();
</script>

<div
	class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0 {isSelected
		? 'active'
		: ''}"
>
	<a
		href={`/admin/emergency-contacts?contactId=${contact.id}`}
		class="flex items-center gap-2 w-full"
		onclick={(event) => {
			event.preventDefault();
			onSelect(contact);
		}}
		role="button"
		tabindex="0"
		aria-label={`Edit emergency contact: ${contact.name ?? 'Unknown'}`}
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') {
				e.preventDefault();
				onSelect(contact);
			}
		}}
	>
		<PhoneIcon class="h-4 w-4 flex-shrink-0" />
		<div class="flex-grow min-w-0">
			<div class="flex items-center gap-2">
				<span class="truncate font-medium">{contact.name ?? 'Unknown'}</span>
				{#if contact.isDefault}
					<Badge variant="default" class="text-xs">
						<StarIcon class="h-2 w-2 mr-1" />
						Default
					</Badge>
				{/if}
			</div>
			<div class="text-xs text-muted-foreground truncate">
				{formatPhoneNumber(contact.number ?? '')}
			</div>
			{#if contact.description}
				<div class="text-xs text-muted-foreground truncate">{contact.description}</div>
			{/if}
		</div>
	</a>
</div>
