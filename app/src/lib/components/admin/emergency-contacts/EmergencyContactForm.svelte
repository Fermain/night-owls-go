<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Checkbox } from '$lib/components/ui/checkbox';
	import * as Card from '$lib/components/ui/card';
	import { toast } from 'svelte-sonner';
	import { authenticatedFetch } from '$lib/utils/api';
	import { goto } from '$app/navigation';
	import { useQueryClient } from '@tanstack/svelte-query';

	interface EmergencyContact {
		id: number;
		name: string;
		number: string;
		description: string;
		is_default: boolean;
		display_order: number;
	}

	interface Props {
		contact?: EmergencyContact;
		onSuccess?: () => void;
	}

	let { contact, onSuccess }: Props = $props();

	const queryClient = useQueryClient();

	// Form state
	let formName = $state(contact?.name || '');
	let formNumber = $state(contact?.number || '');
	let formDescription = $state(contact?.description || '');
	let formIsDefault = $state(contact?.is_default || false);
	let formDisplayOrder = $state(contact?.display_order || 1);
	let formSubmitting = $state(false);

	// Update form fields when contact prop changes
	$effect(() => {
		if (contact) {
			formName = contact.name;
			formNumber = contact.number;
			formDescription = contact.description;
			formIsDefault = contact.is_default;
			formDisplayOrder = contact.display_order;
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
		
		if (!formName.trim() || !formNumber.trim()) {
			toast.error('Name and number are required');
			return;
		}

		formSubmitting = true;

		try {
			const url = isEditing && contact
				? `/api/admin/emergency-contacts/${contact.id}`
				: '/api/admin/emergency-contacts';
			
			const method = isEditing ? 'PUT' : 'POST';

			const response = await authenticatedFetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					name: formName.trim(),
					number: formNumber.trim(),
					description: formDescription.trim(),
					is_default: formIsDefault,
					display_order: formDisplayOrder
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
		if (!isEditing || !contact) return;
		
		if (!confirm(`Are you sure you want to delete "${contact.name}"?`)) {
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
		if (!isEditing || !contact) return;

		try {
			const response = await authenticatedFetch(`/api/admin/emergency-contacts/${contact.id}/default`, {
				method: 'PUT'
			});

			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to set default contact');
			}

			toast.success(`${contact.name} set as default emergency contact`);
			
			// Invalidate the query to refresh the sidebar list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });
		} catch (err) {
			const message = err instanceof Error ? err.message : 'Failed to set default contact';
			toast.error(message);
		}
	}
</script>

<div class="container mx-auto p-6 max-w-2xl">
	<Card.Root>
		<Card.Header>
			<Card.Title>{title}</Card.Title>
			<Card.Description>
				{isEditing 
					? 'Update the emergency contact information' 
					: 'Add a new emergency contact for the community'}
			</Card.Description>
		</Card.Header>
		
		<Card.Content>
			<form onsubmit={handleSubmit} class="space-y-4">
				<div class="space-y-2">
					<Label for="name">Name *</Label>
					<Input
						id="name"
						bind:value={formName}
						placeholder="e.g., RUSA, SAPS, ER24"
						required
					/>
				</div>

				<div class="space-y-2">
					<Label for="number">Phone Number *</Label>
					<Input
						id="number"
						bind:value={formNumber}
						placeholder="e.g., 086 123 4333"
						required
					/>
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
					<Input
						id="display-order"
						type="number"
						bind:value={formDisplayOrder}
						min="1"
					/>
				</div>

				<div class="flex items-center space-x-2">
					<Checkbox
						id="is-default"
						bind:checked={formIsDefault}
					/>
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
							<Button type="button" variant="destructive" onclick={handleDelete}>
								Delete
							</Button>
						{/if}
					{/if}
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div> 