<script lang="ts">
	// === IMPORTS ===
	// UI Components (centralized imports)
	import {
		Button,
		Input,
		Label,
		Textarea,
		Checkbox,
		Card,
		CardContent,
		LoadingState,
		ErrorState
	} from '$lib/components/ui';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';
	import PhoneIcon from '@lucide/svelte/icons/phone';

	// Utilities with new patterns
	import { apiPost, apiPut, apiDelete } from '$lib/utils/api';
	import { classifyError, getErrorMessage } from '$lib/utils/errors';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { useQueryClient } from '@tanstack/svelte-query';

	// Types using our new domain types and API mappings
	import type { EmergencyContact, CreateEmergencyContactData } from '$lib/types/domain';
	import type { BaseComponentProps } from '$lib/types/ui';
	import {
		mapCreateEmergencyContactToAPIRequest,
		mapUpdateEmergencyContactToAPIRequest
	} from '$lib/types/api-mappings';

	// Legacy validation utility (will migrate this later)
	import { validateEmergencyContact } from '$lib/utils/emergencyContacts';

	// === COMPONENT PROPS ===
	interface EmergencyContactFormProps extends BaseComponentProps {
		contact?: EmergencyContact;
		onSuccess?: () => void;
	}

	let {
		contact,
		onSuccess,
		className,
		id,
		'data-testid': testId,
		...props
	}: EmergencyContactFormProps = $props();

	// === STATE MANAGEMENT ===
	const queryClient = useQueryClient();

	// Form values using our domain types
	let formValues = $state<CreateEmergencyContactData>({
		name: contact?.name ?? '',
		number: contact?.number ?? '',
		description: contact?.description ?? '',
		isDefault: contact?.isDefault ?? false,
		displayOrder: contact?.displayOrder ?? 1
	});

	// Form state - simplified approach that works with our domain types
	let formState = $state({
		errors: {} as Partial<Record<keyof CreateEmergencyContactData, string>>,
		touched: {} as Partial<Record<keyof CreateEmergencyContactData, boolean>>,
		dirty: false,
		valid: true,
		submitting: false
	});

	// Derived values to track form state
	const currentFormValues = $derived(formValues);

	// API operation states
	let deleteState = $state<{ loading: boolean; error: Error | null }>({
		loading: false,
		error: null
	});
	let setDefaultState = $state<{ loading: boolean; error: Error | null }>({
		loading: false,
		error: null
	});

	// === DERIVED VALUES ===
	const isEditing = $derived(!!contact);
	const title = $derived(isEditing ? 'Edit Emergency Contact' : 'Create Emergency Contact');
	const submitText = $derived(isEditing ? 'Update Contact' : 'Create Contact');

	// === EFFECTS ===
	// Update form when contact prop changes
	$effect(() => {
		if (contact) {
			const newValues = {
				name: contact.name ?? '',
				number: contact.number ?? '',
				description: contact.description ?? '',
				isDefault: contact.isDefault ?? false,
				displayOrder: contact.displayOrder ?? 1
			};
			formValues = newValues;
			formState.dirty = false;
			formState.errors = {};
			formState.touched = {};
		} else {
			// Reset for new contact creation
			const newValues = {
				name: '',
				number: '',
				description: '',
				isDefault: false,
				displayOrder: 1
			};
			formValues = newValues;
			formState.dirty = false;
			formState.errors = {};
			formState.touched = {};
		}
	});

	// Update form state when values change
	$effect(() => {
		formState.dirty = true;

		// Validate form
		const validation = validateEmergencyContact({
			name: formValues.name,
			number: formValues.number,
			description: formValues.description,
			is_default: formValues.isDefault,
			display_order: formValues.displayOrder
		});

		formState.valid = validation.isValid;
		// Convert validation errors to our FormState format
		formState.errors = validation.isValid
			? {}
			: {
					name: validation.errors.find((e) => e.includes('Name')) || undefined,
					number: validation.errors.find((e) => e.includes('Phone')) || undefined,
					displayOrder: validation.errors.find((e) => e.includes('Display order')) || undefined
				};
	});

	// === FORM HANDLERS ===
	function markFieldTouched(field: keyof CreateEmergencyContactData) {
		formState.touched[field] = true;
	}

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		if (!formState.valid) {
			// Mark all fields as touched to show errors
			(Object.keys(formValues) as Array<keyof CreateEmergencyContactData>).forEach((key) => {
				formState.touched[key] = true;
			});
			toast.error('Please fix the form errors before submitting');
			return;
		}

		formState.submitting = true;

		try {
			if (isEditing && contact?.id) {
				// Update existing contact
				const requestData = mapUpdateEmergencyContactToAPIRequest(formValues);
				await apiPut(`/emergency-contacts/${contact.id}`, requestData);
				toast.success('Contact updated successfully');
			} else {
				// Create new contact
				const requestData = mapCreateEmergencyContactToAPIRequest(formValues);
				await apiPost('/emergency-contacts', requestData);
				toast.success('Contact created successfully');

				// Reset form for new contact creation
				const newValues = {
					name: '',
					number: '',
					description: '',
					isDefault: false,
					displayOrder: 1
				};
				formValues = newValues;
				formState.touched = {};
				formState.dirty = false;
			}

			// Refresh the emergency contacts list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });

			if (onSuccess) {
				onSuccess();
			}
		} catch (error) {
			const appError = classifyError(error);
			toast.error(getErrorMessage(appError));
		} finally {
			formState.submitting = false;
		}
	}

	async function handleDelete() {
		if (!isEditing || !contact?.id) return;

		if (!confirm(`Are you sure you want to delete "${contact.name ?? 'this contact'}"?`)) {
			return;
		}

		deleteState.loading = true;
		deleteState.error = null;

		try {
			await apiDelete(`/emergency-contacts/${contact.id}`);
			toast.success('Contact deleted successfully');

			// Refresh the contacts list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });

			// Navigate back to create form
			goto('/admin/emergency-contacts');
		} catch (error) {
			const appError = classifyError(error);
			deleteState.error = appError;
			toast.error(getErrorMessage(appError));
		} finally {
			deleteState.loading = false;
		}
	}

	async function handleSetDefault() {
		if (!isEditing || !contact?.id) return;

		setDefaultState.loading = true;
		setDefaultState.error = null;

		try {
			await apiPut(`/emergency-contacts/${contact.id}/default`);
			toast.success(`${contact.name ?? 'Contact'} set as default emergency contact`);

			// Refresh the contacts list
			queryClient.invalidateQueries({ queryKey: ['adminEmergencyContacts'] });
		} catch (error) {
			const appError = classifyError(error);
			setDefaultState.error = appError;
			toast.error(getErrorMessage(appError));
		} finally {
			setDefaultState.loading = false;
		}
	}
</script>

<div {id} data-testid={testId} class="container mx-auto p-6 max-w-6xl {className}" {...props}>
	<AdminPageHeader
		icon={PhoneIcon}
		heading={title}
		subheading={isEditing
			? 'Update the emergency contact information'
			: 'Add a new emergency contact for the community'}
	/>

	<Card>
		<CardContent class="p-6">
			{#if formState.submitting}
				<LoadingState isLoading={true} loadingText="Saving contact..." />
			{:else}
				<form onsubmit={handleSubmit} class="space-y-4">
					<!-- Name Field -->
					<div class="space-y-2">
						<Label for="name">Name *</Label>
						<Input
							id="name"
							bind:value={formValues.name}
							onblur={() => markFieldTouched('name')}
							placeholder="e.g., RUSA, SAPS, ER24"
							required
							class={formState.touched.name && formState.errors.name ? 'border-destructive' : ''}
						/>
						{#if formState.touched.name && formState.errors.name}
							<p class="text-sm text-destructive">{formState.errors.name}</p>
						{/if}
					</div>

					<!-- Number Field -->
					<div class="space-y-2">
						<Label for="number">Phone Number *</Label>
						<Input
							id="number"
							bind:value={formValues.number}
							onblur={() => markFieldTouched('number')}
							placeholder="e.g., 086 123 4333"
							required
							class={formState.touched.number && formState.errors.number
								? 'border-destructive'
								: ''}
						/>
						{#if formState.touched.number && formState.errors.number}
							<p class="text-sm text-destructive">{formState.errors.number}</p>
						{/if}
					</div>

					<!-- Description Field -->
					<div class="space-y-2">
						<Label for="description">Description</Label>
						<Textarea
							id="description"
							bind:value={formValues.description}
							onblur={() => markFieldTouched('description')}
							placeholder="Brief description of the service"
							rows={2}
						/>
					</div>

					<!-- Display Order Field -->
					<div class="space-y-2">
						<Label for="display-order">Display Order</Label>
						<Input
							id="display-order"
							type="number"
							bind:value={formValues.displayOrder}
							onblur={() => markFieldTouched('displayOrder')}
							min="1"
							class={formState.touched.displayOrder && formState.errors.displayOrder
								? 'border-destructive'
								: ''}
						/>
						{#if formState.touched.displayOrder && formState.errors.displayOrder}
							<p class="text-sm text-destructive">{formState.errors.displayOrder}</p>
						{/if}
					</div>

					<!-- Default Checkbox -->
					<div class="flex items-center space-x-2">
						<Checkbox id="is-default" bind:checked={formValues.isDefault} />
						<Label for="is-default" class="text-sm">Set as default emergency contact</Label>
					</div>

					<!-- Action Buttons -->
					<div class="flex gap-2 pt-4">
						<Button
							type="submit"
							disabled={formState.submitting || !formState.valid}
							class="flex-1"
						>
							{submitText}
						</Button>

						{#if isEditing && contact}
							{#if !contact.isDefault}
								<Button
									type="button"
									variant="outline"
									onclick={handleSetDefault}
									disabled={setDefaultState.loading}
								>
									{setDefaultState.loading ? 'Setting...' : 'Set as Default'}
								</Button>
							{/if}

							{#if !contact.isDefault}
								<Button
									type="button"
									variant="destructive"
									onclick={handleDelete}
									disabled={deleteState.loading}
								>
									{deleteState.loading ? 'Deleting...' : 'Delete'}
								</Button>
							{/if}
						{/if}
					</div>
				</form>

				<!-- Error States for Operations -->
				{#if setDefaultState.error}
					<div class="mt-4">
						<ErrorState
							error={setDefaultState.error}
							title="Failed to set default"
							showRetry={true}
							onRetry={handleSetDefault}
						/>
					</div>
				{/if}

				{#if deleteState.error}
					<div class="mt-4">
						<ErrorState
							error={deleteState.error}
							title="Failed to delete contact"
							showRetry={true}
							onRetry={handleDelete}
						/>
					</div>
				{/if}
			{/if}
		</CardContent>
	</Card>
</div>
