<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import TrashIcon from '@lucide/svelte/icons/trash';

	let {
		open = $bindable(false),
		title = 'Confirm Deletion',
		description = 'This action cannot be undone.',
		onConfirm,
		isLoading = false
	}: {
		open?: boolean;
		title?: string;
		description?: string;
		onConfirm: () => void;
		isLoading?: boolean;
	} = $props();
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content>
		<AlertDialog.Header>
			<AlertDialog.Title class="flex items-center gap-2">
				<TrashIcon class="h-5 w-5 text-destructive" />
				{title}
			</AlertDialog.Title>
			<AlertDialog.Description>
				{description}
			</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel disabled={isLoading}>Cancel</AlertDialog.Cancel>
			<AlertDialog.Action
				onclick={onConfirm}
				disabled={isLoading}
				class="bg-destructive text-destructive-foreground hover:bg-destructive/90"
			>
				{#if isLoading}
					Deleting...
				{:else}
					Delete
				{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
