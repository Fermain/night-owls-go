<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import PlusIcon from '@lucide/svelte/icons/plus-circle';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';

	// Utilities with new patterns
	import { apiGet } from '$lib/utils/api';
	import { classifyError } from '$lib/utils/errors';

	// Components
	import EmergencyContactThumbnail from '$lib/components/admin/emergency-contacts/EmergencyContactThumbnail.svelte';

	// Types using our new domain types and API mappings
	import type { EmergencyContact } from '$lib/types/domain';
	import type { components } from '$lib/types/api';
	import { mapAPIEmergencyContactArrayToDomain } from '$lib/types/api-mappings';

	let searchTerm = $state('');

	// Define navigation items for the emergency contacts section
	const emergencyContactsNavItems = [
		{
			title: 'Create Contact',
			url: '/admin/emergency-contacts',
			icon: LayoutDashboardIcon
		}
	];

	// Create a query for emergency contacts using our new API utilities
	const contactsQuery = $derived(
		createQuery<EmergencyContact[], Error>({
			queryKey: ['adminEmergencyContacts'],
			queryFn: async () => {
				try {
					const apiContacts = await apiGet<components['schemas']['api.EmergencyContactResponse'][]>(
						'/api/admin/emergency-contacts'
					);
					return mapAPIEmergencyContactArrayToDomain(apiContacts);
				} catch (error) {
					throw classifyError(error);
				}
			},
			staleTime: 1000 * 60 * 5, // 5 minutes
			gcTime: 1000 * 60 * 10, // 10 minutes
			retry: 2
		})
	);

	// Domain-aware filter function
	function filterEmergencyContacts(
		contacts: EmergencyContact[],
		searchTerm: string
	): EmergencyContact[] {
		if (!searchTerm) return sortContactsByDisplayOrder(contacts);

		const term = searchTerm.toLowerCase();
		return contacts
			.filter(
				(contact) =>
					contact.name?.toLowerCase().includes(term) ||
					contact.number?.includes(term) ||
					contact.description?.toLowerCase().includes(term)
			)
			.sort((a, b) => (a.displayOrder ?? 0) - (b.displayOrder ?? 0));
	}

	// Sort contacts by display order (domain-aware)
	function sortContactsByDisplayOrder(contacts: EmergencyContact[]): EmergencyContact[] {
		return [...contacts].sort((a, b) => (a.displayOrder ?? 0) - (b.displayOrder ?? 0));
	}

	// Filtered contacts for display in sidebar
	const filteredContacts = $derived.by(() => {
		const contacts = $contactsQuery.data ?? [];
		return filterEmergencyContacts(contacts, searchTerm);
	});

	// Handle selecting a contact for editing
	const selectContactForEditing = (contact: EmergencyContact) => {
		goto(`/admin/emergency-contacts?contactId=${contact.id}`);
	};

	// Get current selected contact ID from URL query parameters
	const currentSelectedContactId = $derived(() => {
		const contactId = page.url.searchParams.get('contactId');
		return contactId ? Number(contactId) : undefined;
	});

	let { children } = $props();
</script>

{#snippet contactListContent()}
	<div class="flex flex-col h-full">
		<!-- Top static nav items (Create Contact) -->
		{#each emergencyContactsNavItems as item (item.title)}
			<a
				href={item.url}
				class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight"
				class:active={page.url.pathname === '/admin/emergency-contacts' &&
					!currentSelectedContactId()}
			>
				{#if item.icon}
					<item.icon class="h-4 w-4" />
				{/if}
				<span>{item.title}</span>
			</a>
		{/each}

		<!-- Contact list (potentially scrollable) -->
		<div class="flex-grow overflow-y-auto">
			{#if $contactsQuery.isLoading}
				<div class="p-4 text-sm">Loading contacts...</div>
			{:else if $contactsQuery.isError}
				<div class="p-4 text-sm text-destructive">
					Error loading contacts: {$contactsQuery.error?.message || 'Failed to load contacts'}
				</div>
			{:else if filteredContacts && filteredContacts.length > 0}
				{#each filteredContacts as contact (contact.id)}
					<EmergencyContactThumbnail
						{contact}
						isSelected={currentSelectedContactId() === contact.id}
						onSelect={selectContactForEditing}
					/>
				{/each}
			{:else if $contactsQuery.data}
				<div class="p-4 text-sm text-muted-foreground">
					{searchTerm
						? `No contacts found matching "${searchTerm}".`
						: 'No emergency contacts found.'}
				</div>
			{/if}
		</div>

		<!-- Create Contact button at the bottom -->
		<div class="p-3 border-t mt-auto">
			<Button
				href="/admin/emergency-contacts"
				class="w-full"
				variant={page.url.pathname === '/admin/emergency-contacts' && !currentSelectedContactId()
					? 'default'
					: 'outline'}
			>
				<PlusIcon class="h-4 w-4 mr-2" />
				Create Contact
			</Button>
		</div>
	</div>
{/snippet}

<SidebarPage listContent={contactListContent} title="Emergency Contacts" bind:searchTerm>
	{@render children()}
</SidebarPage>
