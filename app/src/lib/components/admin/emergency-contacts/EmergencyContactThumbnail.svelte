<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import StarIcon from '@lucide/svelte/icons/star';
	import type { EmergencyContact } from '$lib/utils/emergencyContacts';
	import { formatPhoneNumber } from '$lib/utils/emergencyContacts';

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
		href={`/admin/emergency-contacts/${contact.id}`}
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
				{#if contact.is_default}
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
