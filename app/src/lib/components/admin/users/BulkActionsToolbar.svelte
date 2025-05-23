<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { Trash2Icon, XIcon } from 'lucide-svelte';
	import { createBulkDeleteUsersMutation } from '$lib/queries/admin/users/bulkDeleteUsersMutation';
	import type { UserData } from '$lib/schemas/user';

	// Props
	let { 
		selectedUsers,
		allUsers,
		onExitBulkMode,
		onClearSelection
	}: {
		selectedUsers: UserData[];
		allUsers: UserData[];
		onExitBulkMode: () => void;
		onClearSelection: () => void;
	} = $props();

	// State
	let showDeleteConfirmDialog = $state(false);
	
	// Mutations
	const bulkDeleteMutation = createBulkDeleteUsersMutation(() => {
		showDeleteConfirmDialog = false;
		onExitBulkMode();
	});

	// Computed
	const selectedCount = $derived(selectedUsers.length);
	const totalCount = $derived(allUsers.length);
	const hasSelection = $derived(selectedCount > 0);

	// Actions
	function handleBulkDelete() {
		showDeleteConfirmDialog = true;
	}

	function confirmBulkDelete() {
		const userIds = selectedUsers.map(user => user.id);
		$bulkDeleteMutation.mutate(userIds);
	}

	// Keyboard shortcuts
	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Escape') {
			onExitBulkMode();
		}
	}
</script>

<svelte:window onkeydown={handleKeydown} />

<div class="bg-primary text-primary-foreground border-b border-border">

	<!-- Action buttons - stacked for narrow spaces -->
	<div class="space-y-2">
		{#if hasSelection}
			<div class="flex flex-col gap-1">
				<Button
					variant="destructive"
					size="sm"
					onclick={handleBulkDelete}
					disabled={$bulkDeleteMutation.isPending}
					class="w-full rounded-none"
				>
					<Trash2Icon class="w-3 h-3 mr-1" />
					Delete {selectedCount} user{selectedCount === 1 ? '' : 's'}
				</Button>
			</div>
		{/if}
	</div>
</div>

<!-- Bulk Delete Confirmation Dialog -->
<AlertDialog.Root bind:open={showDeleteConfirmDialog}>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title>Delete {selectedCount} User{selectedCount === 1 ? '' : 's'}?</AlertDialog.Title>
			<AlertDialog.Description>
				This action cannot be undone. You are about to permanently delete the following 
				{selectedCount} user{selectedCount === 1 ? '' : 's'}:
				
				<div class="mt-3 max-h-32 overflow-y-auto bg-muted p-3 rounded">
					<ul class="space-y-1 text-sm">
						{#each selectedUsers as user (user.id)}
							<li class="flex items-center gap-2">
								<span class="font-medium">{user.name || 'Unnamed User'}</span>
								<span class="text-muted-foreground">({user.phone})</span>
								{#if user.role === 'admin'}
									<span class="inline-flex items-center rounded-md bg-destructive text-destructive-foreground px-2 py-1 text-xs font-medium">Admin</span>
								{/if}
							</li>
						{/each}
					</ul>
				</div>
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				onclick={confirmBulkDelete}
				disabled={$bulkDeleteMutation.isPending}
				class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
			>
				{#if $bulkDeleteMutation.isPending}
					Deleting...
				{:else}
					<Trash2Icon class="w-4 h-4 mr-2" />
					Delete {selectedCount} User{selectedCount === 1 ? '' : 's'}
				{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root> 