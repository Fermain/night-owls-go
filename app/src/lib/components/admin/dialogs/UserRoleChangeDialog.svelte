<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import UserIcon from '@lucide/svelte/icons/user';
	import type { CreateMutationResult } from '@tanstack/svelte-query';
	import type { UserRole } from '$lib/types';

	interface User {
		id: number;
		name: string | null;
		role: string;
		[key: string]: unknown;
	}

	let {
		open = $bindable(false),
		userName = '',
		currentRole = $bindable(''),
		newRole = '',
		onConfirm = () => {},
		isLoading = false,
		_user = null,
		_mutation = null
	}: {
		open?: boolean;
		userName?: string;
		currentRole?: string;
		newRole?: string;
		onConfirm?: (role: UserRole) => void;
		isLoading?: boolean;
		_user?: User | null;
		_mutation?: CreateMutationResult<unknown, Error, unknown, unknown> | null;
	} = $props();

	const roleOptions = [
		{ value: 'guest', label: 'Guest' },
		{ value: 'owl', label: 'Owl' },
		{ value: 'admin', label: 'Admin' }
	];

	const getRoleDisplayName = (role: string) => {
		const option = roleOptions.find((opt) => opt.value === role);
		return option ? option.label : role;
	};
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title class="flex items-center gap-2">
				<UserIcon class="h-5 w-5" />
				Change User Role
			</AlertDialog.Title>
			<AlertDialog.Description>
				Change the role for {userName} from {getRoleDisplayName(currentRole)} to a new role.
			</AlertDialog.Description>
		</AlertDialog.Header>
		<div class="py-4">
			<div class="space-y-2">
				<Label>New Role</Label>
				<Select.Root type="single" bind:value={newRole}>
					<Select.Trigger>
						{roleOptions.find((opt) => opt.value === newRole)?.label || 'Select new role'}
					</Select.Trigger>
					<Select.Content>
						{#each roleOptions as option (option.value)}
							<Select.Item value={option.value} label={option.label}>
								{option.label}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</div>
		</div>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={isLoading}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				onclick={() => onConfirm(newRole as UserRole)}
				disabled={isLoading || !newRole || newRole === currentRole}
			>
				{#if isLoading}
					Updating...
				{:else}
					Update Role
				{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
