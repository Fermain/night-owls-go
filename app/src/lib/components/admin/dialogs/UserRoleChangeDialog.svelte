<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import type { UserRole } from '$lib/types';
	import type { UserData } from '$lib/schemas/user';

	let {
		open = $bindable(false),
		user,
		currentRole = $bindable('guest' as UserRole),
		onConfirm
	} = $props<{
		open?: boolean;
		user: UserData | undefined | null;
		currentRole?: UserRole;
		onConfirm: (newRole: UserRole) => void;
	}>();

	const roleDisplayValues: Record<UserRole, string> = {
		admin: 'Admin',
		owl: 'Owl',
		guest: 'Guest'
	};

	let selectedRoleInDialog: UserRole = $state(currentRole);

	$effect(() => {
		selectedRoleInDialog = currentRole;
	});

	function handleConfirm() {
		onConfirm(selectedRoleInDialog);
		open = false;
	}

	function handleCancel() {
		selectedRoleInDialog = currentRole;
		open = false;
	}
</script>

<AlertDialog.Root
	bind:open
	onOpenChange={(isOpen) => {
		if (!isOpen) handleCancel();
	}}
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
			<Select.Root type="single" bind:value={selectedRoleInDialog}>
				<Select.Trigger class="w-full" id="dialog-role">
					{roleDisplayValues[selectedRoleInDialog]}
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
			<AlertDialog.Cancel onclick={handleCancel}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action onclick={handleConfirm}>Confirm Change</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
