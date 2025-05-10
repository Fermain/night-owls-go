<script context="module" lang="ts">
	export type UserData = {
		id: number;
		phone: string;
		name: string | null;
		created_at: string;
	};
</script>

<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { toast } from 'svelte-sonner';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import { z } from 'zod';

	// Prop for existing user data (undefined if creating a new one)
	export let user: UserData | undefined = undefined;

	// Define schema with Zod
	const userSchema = z.object({
		phone: z
			.string()
			.min(1, 'Phone number is required')
			.regex(
				/^\+?[0-9]{10,15}$/,
				'Please enter a valid phone number (10-15 digits, can start with +)'
			),
		name: z.string().nullable()
	});

	type FormValues = z.infer<typeof userSchema>;

	// Local Svelte state for form data
	let formData: FormValues = {
		phone: '',
		name: null
	};

	// State for validation errors
	let errors: Partial<Record<keyof FormValues, string>> = {};

	onMount(() => {
		if (user) {
			formData = {
				phone: user.phone,
				name: user.name
			};
		}
	});

	const queryClient = useQueryClient();

	type MutationVariables = {
		payload: FormValues;
		userId?: number;
	};

	const mutation = createMutation<
		Response,
		Error,
		MutationVariables
	>({
		mutationFn: async (vars) => {
			const { payload, userId: currentUserIdToUse } = vars;
			const currentIsEditMode = currentUserIdToUse !== undefined;

			const url = currentIsEditMode ? `/api/admin/users/${currentUserIdToUse}` : '/api/admin/users';
			const method = currentIsEditMode ? 'PUT' : 'POST';

			console.log('Sending request:', { method, url, payload });

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

	function validateForm(): boolean {
		const result = userSchema.safeParse(formData);
		if (!result.success) {
			const newErrors: Partial<Record<keyof FormValues, string>> = {};
			for (const issue of result.error.issues) {
				if (issue.path.length > 0) {
					newErrors[issue.path[0] as keyof FormValues] = issue.message;
				}
			}
			errors = newErrors;
			return false;
		}
		errors = {};
		return true;
	}

	function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		const currentUserIdFromProp = user?.id;

		const payloadForSubmit: FormValues = {
			phone: formData.phone.trim(),
			name: formData.name?.trim() === '' ? null : formData.name
		};

		const mutationVars: MutationVariables = {
			payload: payloadForSubmit,
			userId: currentUserIdFromProp
		};

		$mutation.mutate(mutationVars);
	}
</script>

<svelte:head>
	<title>{user?.id !== undefined ? 'Edit' : 'Create New'} User</title>
</svelte:head>

<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold mb-6">
		{user?.id !== undefined ? 'Edit' : 'Create New'} User
	</h1>

	<form on:submit|preventDefault={handleSubmit} class="space-y-6 max-w-lg">
		<div>
			<Label for="phone">Phone Number</Label>
			<Input
				id="phone"
				type="text"
				bind:value={formData.phone}
				required
				class={errors.phone ? 'border-red-500' : ''}
			/>
			{#if errors.phone}
				<p class="text-sm text-destructive mt-1">{errors.phone}</p>
			{:else}
				<p class="text-sm text-muted-foreground mt-1">Required. Format: +1234567890</p>
			{/if}
		</div>

		<div>
			<Label for="name">Name (Optional)</Label>
			<Input
				id="name"
				type="text"
				bind:value={formData.name}
				class={errors.name ? 'border-red-500' : ''}
			/>
			{#if errors.name}
				<p class="text-sm text-destructive mt-1">{errors.name}</p>
			{/if}
			<p class="text-sm text-muted-foreground mt-1">User's full name</p>
		</div>

		{#if user?.id !== undefined}
			<div>
				<Label>Created At</Label>
				<p class="text-sm text-muted-foreground mt-1">
					{new Date(user.created_at).toLocaleString()}
				</p>
			</div>
		{/if}

		<div class="flex gap-4">
			<Button type="submit" disabled={$mutation.isPending}>
				{#if $mutation.isPending}
					Saving...
				{:else}
					{user?.id !== undefined ? 'Update' : 'Create'} User
				{/if}
			</Button>
			<Button type="button" variant="outline" onclick={() => goto('/admin/users')}>Cancel</Button>
		</div>
	</form>
</div>
