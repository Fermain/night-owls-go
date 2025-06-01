<script lang="ts">
	import { page } from '$app/state';
	import { createQuery } from '@tanstack/svelte-query';
	import { authenticatedFetch } from '$lib/utils/api';
	import EmergencyContactForm from '$lib/components/admin/emergency-contacts/EmergencyContactForm.svelte';
	import * as Card from '$lib/components/ui/card';
	import type { EmergencyContact } from '$lib/utils/emergencyContacts';

	// Get contact ID from query parameters
	const contactId = $derived(() => {
		const id = page.url.searchParams.get('contactId');
		return id ? parseInt(id, 10) : undefined;
	});

	// Query for all contacts and find the specific one
	const contactsQuery = $derived(
		createQuery<EmergencyContact[], Error>({
			queryKey: ['adminEmergencyContacts'],
			queryFn: async () => {
				const response = await authenticatedFetch('/api/admin/emergency-contacts');
				if (!response.ok) {
					throw new Error('Failed to load emergency contacts');
				}
				return response.json();
			}
		})
	);

	const contact = $derived(() => {
		const id = contactId();
		if (!id) return undefined;

		const contacts = $contactsQuery.data;
		return contacts?.find((c) => c.id === id);
	});

	const isEditing = $derived(!!contactId());
</script>

<svelte:head>
	<title>{isEditing ? 'Edit Emergency Contact' : 'Create Emergency Contact'} - Admin</title>
</svelte:head>

{#if isEditing}
	{#if $contactsQuery.isLoading}
		<div class="container mx-auto p-6 max-w-2xl">
			<Card.Root>
				<Card.Content class="p-6">
					<div class="flex items-center justify-center">
						<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	{:else if $contactsQuery.isError}
		<div class="container mx-auto p-6 max-w-2xl">
			<Card.Root>
				<Card.Content class="p-6 text-center">
					<p class="text-destructive">Error loading contacts: {$contactsQuery.error.message}</p>
				</Card.Content>
			</Card.Root>
		</div>
	{:else if contact()}
		<EmergencyContactForm contact={contact()} />
	{:else}
		<div class="container mx-auto p-6 max-w-2xl">
			<Card.Root>
				<Card.Content class="p-6 text-center">
					<p class="text-muted-foreground">Contact not found</p>
				</Card.Content>
			</Card.Root>
		</div>
	{/if}
{:else}
	<!-- Create mode -->
	<EmergencyContactForm />
{/if}
