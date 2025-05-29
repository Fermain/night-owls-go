<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Card from '$lib/components/ui/card';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';
	import PhoneIcon from '@lucide/svelte/icons/phone';
	import { toast } from 'svelte-sonner';
	import { authenticatedFetch } from '$lib/utils/api';
	import { goto } from '$app/navigation';
	import { useQueryClient } from '@tanstack/svelte-query';
	import {
		validateEmergencyContact,
		type EmergencyContact,
		type CreateEmergencyContactRequest
	} from '$lib/utils/emergencyContacts';

	interface Props {
		contact?: EmergencyContact;
		onSuccess?: () => void;
	}

	let { contact, onSuccess }: Props = $props();

	const queryClient = useQueryClient();

	// Form state - handle optional properties from OpenAPI types
	let formName = $state(contact?.name ?? '');
	let formNumber = $state(contact?.number ?? '');
	let formDescription = $state(contact?.description ?? '');
	let formIsDefault = $state(contact?.is_default ?? false);
	let formDisplayOrder = $state(contact?.display_order ?? 1);
	let formSubmitting = $state(false);

	// Update form fields when contact prop changes
	$effect(() => {
		if (contact) {
			formName = contact.name ?? '';
			formNumber = contact.number ?? '';
			formDescription = contact.description ?? '';
			formIsDefault = contact.is_default ?? false;
			formDisplayOrder = contact.display_order ?? 1;
		} else {
			// Reset form for new contact creation
			formName = '';
			formNumber = '';
			formDescription = '';
			formIsDefault = false;
			formDisplayOrder = 1;
		}
	});

	const isEditing = $derived(!!contact);
	const title = $derived(isEditing ? 'Edit Emergency Contact' : 'Create Emergency Contact');
	const submitText = $derived(isEditing ? 'Update Contact' : 'Create Contact');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		// Validate form data using centralized validation
		const formData: CreateEmergencyContactRequest = {
			name: formName,
			number: formNumber,
			description: formDescription,
			is_default: formIsDefault,
			display_order: formDisplayOrder
		};

		const validation = validateEmergencyContact(formData);
		if (!validation.isValid) {
			toast.error(validation.errors.join(', '));
			return;
		}

		formSubmitting = true;

		try {
			const url =
				isEditing && contact?.id
					? `/api/admin/emergency-contacts/${contact.id}`
					: '/api/admin/emergency-contacts';

			const method = isEditing ? 'PUT' : 'POST';

			const response = await authenticatedFetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					name: formData.name?.trim(),
					number: formData.number?.trim(),
					description: formData.description?.trim(),
					is_default: formData.is_default,
					display_order: formData.display_order
				})
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to save contact');
			}

			toast.success(isEditing ? 'Contact updated successfully' : 'Contact created successfully');

			// Invalidate the query to refresh the sidebar list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });

			if (onSuccess) {
				onSuccess();
			} else if (!isEditing) {
				// Reset form for new contact creation
				formName = '';
				formNumber = '';
				formDescription = '';
				formIsDefault = false;
				formDisplayOrder = 1;
			}
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to save contact';
			toast.error(message);
		} finally {
			formSubmitting = false;
		}
	}

	async function handleDelete() {
		if (!isEditing || !contact?.id) return;

		if (!confirm(`Are you sure you want to delete "${contact.name ?? 'this contact'}"?`)) {
			return;
		}

		try {
			const response = await authenticatedFetch(`/api/admin/emergency-contacts/${contact.id}`, {
				method: 'DELETE'
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to delete contact');
			}

			toast.success('Contact deleted successfully');

			// Invalidate the query to refresh the sidebar list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });

			// Navigate back to the create form
			goto('/admin/emergency-contacts');
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to delete contact';
			toast.error(message);
		}
	}

	async function handleSetDefault() {
		if (!isEditing || !contact?.id) return;

		try {
			const response = await authenticatedFetch(
				`/api/admin/emergency-contacts/${contact.id}/default`,
				{
					method: 'PUT'
				}
			);

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to set default contact');
			}

			toast.success(`${contact.name ?? 'Contact'} set as default emergency contact`);

			// Invalidate the query to refresh the sidebar list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to set default contact';
			toast.error(message);
		}
	}
</script>

<div class="container mx-auto p-6 max-w-6xl">
	<AdminPageHeader
		icon={PhoneIcon}
		heading={title}
		subheading={isEditing
			? 'Update the emergency contact information'
			: 'Add a new emergency contact for the community'}
	/>

	<Card.Root>
		<Card.Content class="p-6">
			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2">
					<Label for="name">Name *</Label>
					<Input id="name" bind:value={formName} placeholder="e.g., RUSA, SAPS, ER24" required />
				</div>

				<div class="space-y-2">
					<Label for="number">Phone Number *</Label>
					<Input id="number" bind:value={formNumber} placeholder="e.g., 086 123 4333" required />
				</div>

				<div class="space-y-2">
					<Label for="description">Description</Label>
					<Textarea
						id="description"
						bind:value={formDescription}
						placeholder="Brief description of the service"
						rows={2}
					/>
				</div>

				<div class="space-y-2">
					<Label for="display-order">Display Order</Label>
					<Input id="display-order" type="number" bind:value={formDisplayOrder} min="1" />
				</div>

				<div class="flex items-center space-x-2">
					<Checkbox id="is-default" bind:checked={formIsDefault} />
					<Label for="is-default" class="text-sm">Set as default emergency contact</Label>
				</div>

				<div class="flex gap-2 pt-4">
					<Button type="submit" disabled={formSubmitting} class="flex-1">
						{formSubmitting ? 'Saving...' : submitText}
					</Button>

					{#if isEditing && contact}
						{#if !contact.is_default}
							<Button type="button" variant="outline" onclick={handleSetDefault}>
								Set as Default
							</Button>
						{/if}

						{#if !contact.is_default}
							<Button type="button" variant="destructive" onclick={handleDelete}>Delete</Button>
						{/if}
					{/if}
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div>
