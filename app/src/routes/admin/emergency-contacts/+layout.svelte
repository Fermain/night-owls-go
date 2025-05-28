<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import LayoutDashboardIcon from '@lucide/svelte/icons/layout-dashboard';
	import PlusIcon from '@lucide/svelte/icons/plus-circle';
	import { createQuery } from '@tanstack/svelte-query';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { authenticatedFetch } from '$lib/utils/api';
	import EmergencyContactThumbnail from '$lib/components/admin/emergency-contacts/EmergencyContactThumbnail.svelte';
	import type { EmergencyContact } from '$lib/utils/emergencyContacts';
	import { filterContacts } from '$lib/utils/emergencyContacts';

	let searchTerm = $state('');

	// Define navigation items for the emergency contacts section
	const emergencyContactsNavItems = [
		{
			title: 'Create Contact',
			url: '/admin/emergency-contacts',
			icon: LayoutDashboardIcon
		}
	];

	// Create a query for emergency contacts
	const contactsQuery = $derived(
		createQuery<EmergencyContact[], Error, EmergencyContact[], [string]>({
			queryKey: ['adminEmergencyContacts'],
			queryFn: async () => {
				const response = await authenticatedFetch('/api/admin/emergency-contacts');
				if (!response.ok) {
					throw new Error('Failed to load emergency contacts');
				}
				return response.json();
			},
			staleTime: 1000 * 60 * 5, // 5 minutes
			gcTime: 1000 * 60 * 10, // 10 minutes
			retry: 2
		})
	);

	// Filtered contacts for display in sidebar
	const filteredContacts = $derived.by(() => {
		const contacts = $contactsQuery.data ?? [];
		return filterContacts(contacts, searchTerm);
	});

	// Handle selecting a contact for editing
	const selectContactForEditing = (contact: EmergencyContact) => {
		goto(`/admin/emergency-contacts/${contact.id}`);
	};

	// Get current selected contact ID from URL
	const currentSelectedContactId = $derived(() => {
		const pathParts = page.url.pathname.split('/');
		const contactId = pathParts[pathParts.length - 1];
		return contactId && !isNaN(Number(contactId)) ? Number(contactId) : undefined;
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
					Error loading contacts: {$contactsQuery.error.message}
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
