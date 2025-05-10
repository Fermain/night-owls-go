<script module lang="ts">
	export type UserData = {
		id: number;
		phone: string;
		name: string | null;
		created_at: string;
		role: 'admin' | 'owl' | 'guest';
	};
</script>

<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import Loader2Icon from '@lucide/svelte/icons/loader-2';
	import UserPlusIcon from '@lucide/svelte/icons/user-plus';
	import { toast } from 'svelte-sonner';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { z } from 'zod';
	import { TelInput } from 'svelte-tel-input';
	import type { E164Number } from 'svelte-tel-input/types';
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as Select from '$lib/components/ui/select';

	// Use $props() for Svelte 5 runes mode
	let { user }: { user?: UserData } = $props();

	// Define schema with Zod
	const userSchema = z.object({
		phone: z.string().min(1, 'Phone number is required'),
		name: z.string().nullable(),
		role: z.enum(['admin', 'owl', 'guest'], { message: 'Role must be admin, owl, or guest' })
	});

	type FormValues = {
		phone: E164Number | '';
		name: string | null;
		role: 'admin' | 'owl' | 'guest';
	};

	// State for svelte-tel-input validity
	let phoneInputValid = $state(true);

	// Local Svelte state for form data, initialized with user prop data if available
	let formData = $state<FormValues>({
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
	let roleInDialog = $state(formData.role);

	$effect(() => {
		// When the user prop changes (e.g., selecting a different user to edit),
		// reset roleInDialog to the new user's current role.
		roleInDialog = formData.role;
	});

	function openRoleDialog() {
		roleInDialog = formData.role; // Initialize dialog with current form role
		showRoleChangeDialog = true;
	}

	function confirmRoleChange() {
		formData.role = roleInDialog;
		showRoleChangeDialog = false;
	}

	// State for Zod validation errors
	let zodErrors = $state<Partial<Record<keyof FormValues, string>>>({});

	// State for controlling delete confirmation dialog
	let showDeleteConfirm = $state(false);

	const queryClient = useQueryClient();

	type MutationVariables = {
		payload: { phone: E164Number; name: string | null; role: 'admin' | 'owl' | 'guest' };
		userId?: number;
	};

	const mutation = createMutation<Response, Error, MutationVariables>({
		mutationFn: async (vars) => {
			const { payload, userId: currentUserIdToUse } = vars;
			const currentIsEditMode = currentUserIdToUse !== undefined;

			const url = currentIsEditMode ? `/api/admin/users/${currentUserIdToUse}` : '/api/admin/users';
			const method = currentIsEditMode ? 'PUT' : 'POST';

			const response = await fetch(url, {
				method: method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(payload)
			});

			if (!response.ok) {
				const errorData = await response.json().catch(() => ({
					message: `Failed to ${currentIsEditMode ? 'update' : 'create'} user`
				}));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response;
		},
		onSuccess: async (_data, vars) => {
			const { userId: mutatedUserId } = vars;
			const currentIsEditMode = mutatedUserId !== undefined;
			toast.success(`User ${currentIsEditMode ? 'updated' : 'created'} successfully!`);
			await queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
			if (currentIsEditMode && mutatedUserId) {
				await queryClient.invalidateQueries({ queryKey: ['adminUser', mutatedUserId] });
			}
			goto('/admin/users');
		},
		onError: (error) => {
			toast.error(`Error: ${error.message}`);
		}
	});

	const deleteUserMutation = createMutation<
		Response, // Assuming server returns a success response (e.g. { message: "..." } or just 200/204)
		Error,
		number // Variable type is userId
	>({
		mutationFn: async (userIdToDelete) => {
			const response = await fetch(`/api/admin/users/${userIdToDelete}`, {
				method: 'DELETE'
			});
			if (!response.ok) {
				const errorData = await response.json().catch(() => ({ message: 'Failed to delete user' }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			// For DELETE, response might be empty (204) or have a message (200)
			// We don't strictly need to parse JSON if it might be empty
			return response;
		},
		onSuccess: async () => {
			toast.success('User deleted successfully!');
			await queryClient.invalidateQueries({ queryKey: ['adminUsers'] });
			goto('/admin/users'); // Navigate away from the potentially deleted user's form
			showDeleteConfirm = false; // Close dialog on success
		},
		onError: (error) => {
			toast.error(`Error deleting user: ${error.message}`);
			showDeleteConfirm = false; // Close dialog on error too
		}
	});

	function validateForm(): boolean {
		const result = userSchema.safeParse(formData);
		if (!result.success) {
			const newErrors: Partial<Record<keyof FormValues, string>> = {};
			for (const issue of result.error.issues) {
				if (issue.path.length > 0) {
					newErrors[issue.path[0] as keyof FormValues] = issue.message;
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

		const mutationVars: MutationVariables = {
			payload: payloadForSubmit,
			userId: currentUserIdFromProp
		};

		$mutation.mutate(mutationVars);
	}

	function handleDeleteClick() {
		if (user?.id) {
			showDeleteConfirm = true;
		}
	}

	function confirmDelete() {
		if (user?.id) {
			$deleteUserMutation.mutate(user.id);
		}
	}
</script>

<svelte:head>
	<title>{user?.id !== undefined ? 'Edit' : 'Create New'} User</title>
</svelte:head>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-6">
		{user?.id !== undefined ? 'Edit' : 'Create New'} User
	</h1>

	<form
		onsubmit={(event) => {
			event.preventDefault();
			handleSubmit();
		}}
		class="space-y-6 max-w-lg"
	>
		<div>
			<Label for="phone" class="block mb-2">Phone Number</Label>
			<TelInput
				disabled={Boolean(user?.id)}
				readonly={Boolean(user?.id)}
				bind:value={formData.phone}
				bind:valid={phoneInputValid}
				country={'ZA'}
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
	<AlertDialog.Root open={showDeleteConfirm} onOpenChange={(open) => (showDeleteConfirm = open)}>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Are you sure you want to delete this user?</AlertDialog.Title>
				<AlertDialog.Description>
					This action cannot be undone. This will permanently delete the user
					{user?.name ? ` "${user.name}"` : ''}
					{user?.phone ? `(${user.phone})` : ''}.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel disabled={$deleteUserMutation.isPending}>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action
					onclick={confirmDelete}
					disabled={$deleteUserMutation.isPending}
					class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
				>
					{#if $deleteUserMutation.isPending}Deleting...{:else}Yes, delete user{/if}
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}

{#if showRoleChangeDialog}
	<AlertDialog.Root
		open={showRoleChangeDialog}
		onOpenChange={(open) => (showRoleChangeDialog = open)}
	>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Change User Role</AlertDialog.Title>
				<AlertDialog.Description>
					Select the new role for {user?.name || 'this user'}.
				</AlertDialog.Description>
			</AlertDialog.Header>

			<div class="py-4">
				<Label for="dialog-role" class="block mb-2">New Role</Label>
				<Select.Root type="single" bind:value={formData.role}>
					<Select.Trigger class="w-full" id="dialog-role">
						Select a role
					</Select.Trigger>
					<Select.Content>
						<Select.Group>
							<Select.GroupHeading>User Role</Select.GroupHeading>
							<Select.Item value="guest" label="Guest">Guest</Select.Item>
							<Select.Item value="owl" label="Owl">Owl</Select.Item>
							<Select.Item value="admin" label="Admin">Admin</Select.Item>
						</Select.Group>
					</Select.Content>
				</Select.Root>
			</div>

			<AlertDialog.Footer>
				<AlertDialog.Cancel
					onclick={() => {
						roleInDialog = formData.role;
					}}>Cancel</AlertDialog.Cancel
				>
				<AlertDialog.Action onclick={confirmRoleChange}>Confirm Change</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
{/if}
