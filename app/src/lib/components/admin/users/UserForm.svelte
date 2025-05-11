<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { TelInput } from 'svelte-tel-input';
	import type { E164Number } from 'svelte-tel-input/types';
	import { createSaveUserMutation } from '$lib/queries/admin/users/saveUserMutation';
	import { createDeleteUserMutation } from '$lib/queries/admin/users/deleteUserMutation';
	import { userSchema, type UserFormValues, type UserData } from '$lib/schemas/user';
	import UserDeleteConfirmDialog from '$lib/components/admin/dialogs/UserDeleteConfirmDialog.svelte';
	import UserRoleChangeDialog from '$lib/components/admin/dialogs/UserRoleChangeDialog.svelte';

	// Use $props() for Svelte 5 runes mode
	let { user }: { user?: UserData } = $props();

	// State for svelte-tel-input validity
	let phoneInputValid = $state(true);

	// Local Svelte state for form data, initialized with user prop data if available
	let formData = $state<UserFormValues>({
		phone: (user?.phone as E164Number) || '',
		name: user?.name || null,
		role: user?.role || 'guest'
	});

	const roleDisplayValues = {
		admin: 'Admin',
		owl: 'Owl',
		guest: 'Guest'
	};

	let showRoleChangeDialog = $state(false);

	$effect(() => {
		// When the user prop changes (e.g., selecting a different user to edit),
		// reset roleInDialog to the new user's current role.
		formData.role = user?.role || 'guest';
	});

	function openRoleDialog() {
		showRoleChangeDialog = true;
	}

	function handleRoleConfirm(newRole: 'admin' | 'owl' | 'guest') {
		formData.role = newRole;
	}

	// State for Zod validation errors
	let zodErrors = $state<Partial<Record<keyof UserFormValues, string>>>({});

	// State for controlling delete confirmation dialog
	let showDeleteConfirm = $state(false);

	const mutation = createSaveUserMutation();

	const deleteUserMutation = createDeleteUserMutation(() => {
		showDeleteConfirm = false;
	});

	function validateForm(): boolean {
		const result = userSchema.safeParse(formData);
		if (!result.success) {
			const newErrors: Partial<Record<keyof UserFormValues, string>> = {};
			for (const issue of result.error.issues) {
				if (issue.path.length > 0) {
					newErrors[issue.path[0] as keyof UserFormValues] = issue.message;
				}
			}
			zodErrors = newErrors;
		} else {
			zodErrors = {};
		}

		if (!phoneInputValid && !zodErrors.phone) {
			zodErrors.phone = 'Invalid phone number format.';
		}

		return result.success && phoneInputValid;
	}

	function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		const currentUserIdFromProp = user?.id;

		if (formData.phone === '' || !phoneInputValid) {
			toast.error('Phone number is invalid or empty.');
			return;
		}

		const payloadForSubmit = {
			phone: formData.phone as E164Number,
			name: formData.name?.trim() === '' ? null : formData.name,
			role: formData.role
		};

		$mutation.mutate({
			payload: payloadForSubmit,
			userId: currentUserIdFromProp
		});
	}

	function handleDeleteClick() {
		if (user?.id) {
			showDeleteConfirm = true;
		}
	}
</script>

<svelte:head>
	<title>{user?.id !== undefined ? 'Edit' : 'Create New'} User</title>
</svelte:head>

<div class="container mr-auto p-4">
	<h1 class="text-2xl font-bold mb-6">
		{user?.id !== undefined ? 'Edit' : 'Create New'} User
	</h1>

	<form
		onsubmit={(event) => {
			event.preventDefault();
			handleSubmit();
		}}
		class="space-y-4"
	>
		<div>
			<Label for="phone" class="block mb-2">Phone Number</Label>
			<TelInput
				disabled={Boolean(user?.id)}
				readonly={Boolean(user?.id)}
				bind:value={formData.phone}
				bind:valid={phoneInputValid}
				country="ZA"
				options={{
					strictCountry: true,
					autoPlaceholder: true,
					format: 'international'
				}}
				class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 invalid:border-red-500"
				required
			/>
			{#if zodErrors.phone}
				<p class="text-sm text-destructive mt-1">{zodErrors.phone}</p>
			{:else if !phoneInputValid && formData.phone !== ''}
				<p class="text-sm text-destructive mt-1">Invalid phone number.</p>
			{/if}
		</div>

		<div>
			<Label for="name" class="block mb-2">Name</Label>
			<Input
				id="name"
				type="text"
				bind:value={formData.name}
				class={zodErrors.name ? 'border-red-500' : ''}
			/>
			{#if zodErrors.name}
				<p class="text-sm text-destructive mt-1">{zodErrors.name}</p>
			{/if}
		</div>

		<div>
			<Label class="block mb-2">Role</Label>
			<div class="flex items-center gap-4">
				<Input disabled readonly value={roleDisplayValues[formData.role]} class="flex-grow" />
				<Button type="button" variant="outline" onclick={openRoleDialog}>Change Role</Button>
			</div>
			{#if zodErrors.role}
				<p class="text-sm text-destructive mt-1">{zodErrors.role}</p>
			{/if}
		</div>

		{#if user?.id !== undefined}
			<div class="text-sm text-muted-foreground">
				<Label>Created</Label>
				<time>
					{new Date(user.created_at).toLocaleString()}
				</time>
			</div>
		{/if}

		<div class="flex gap-4">
			<Button type="submit" disabled={$mutation.isPending} class="flex-1">
				{#if $mutation.isPending}
					<Loader2Icon class="w-4 h-4 mr-2" />
					Saving...
				{:else}
					<UserPlusIcon class="w-4 h-4" />
					{user?.id !== undefined ? 'Update' : 'Create'} User
				{/if}
			</Button>
			<Button type="button" variant="outline" onclick={() => goto('/admin/users')}>Cancel</Button>
			{#if user?.id !== undefined}
				<Button
					type="button"
					variant="destructive"
					onclick={handleDeleteClick}
					disabled={$deleteUserMutation.isPending}
				>
					{#if $deleteUserMutation.isPending}Deleting...{:else}Delete User{/if}
				</Button>
			{/if}
		</div>
	</form>
</div>

{#if showDeleteConfirm}
	<UserDeleteConfirmDialog bind:open={showDeleteConfirm} {user} mutation={deleteUserMutation} />
{/if}

{#if showRoleChangeDialog}
	<UserRoleChangeDialog
		bind:open={showRoleChangeDialog}
		{user}
		bind:currentRole={formData.role}
		onConfirm={handleRoleConfirm}
	/>
{/if}
