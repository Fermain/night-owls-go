<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import type { UserData } from '$lib/schemas/user';

	type Role = 'admin' | 'owl' | 'guest';

	let {
		open = $bindable(false),
		user,
		currentRole = $bindable('guest' as Role),
		onConfirm // Callback when confirm is clicked
	} = $props<{
		open?: boolean;
		user: UserData | undefined | null;
		currentRole?: Role;
		onConfirm: (newRole: Role) => void;
	}>();

	const roleDisplayValues: Record<Role, string> = {
		admin: 'Admin',
		owl: 'Owl',
		guest: 'Guest'
	};

	// Internal state for the selection within the dialog
	let selectedRoleInDialog: Role = $state(currentRole);

	$effect(() => {
		// Sync internal dialog state if the prop changes from outside
		selectedRoleInDialog = currentRole;
	});

	function handleConfirm() {
		onConfirm(selectedRoleInDialog);
		open = false;
	}

	function handleCancel() {
		// Reset internal state to the original role before closing
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
