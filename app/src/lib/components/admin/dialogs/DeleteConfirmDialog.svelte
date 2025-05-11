<script lang="ts">
	import * as AlertDialog from '$lib/components/ui/alert-dialog/index.js';
	import type { CreateMutationResult } from '@tanstack/svelte-query';

	let { open, name, id, mutation } = $props<{
		open: boolean;
		name: string;
		id: number;
		mutation: CreateMutationResult<any, Error, number, unknown>;
	}>();
</script>

<AlertDialog.Root bind:open>
	<AlertDialog.Content>
		<AlertDialog.Header
			><AlertDialog.Title>Are you sure?</AlertDialog.Title>
			<AlertDialog.Description>This will permanently delete "{name}".</AlertDialog.Description>
		</AlertDialog.Header>
		<AlertDialog.Footer>
			<AlertDialog.Cancel onclick={() => (open = false)} disabled={$mutation.isPending}
				>Cancel</AlertDialog.Cancel
			>
			<AlertDialog.Action
				onclick={() => $mutation.mutate(id)}
				disabled={$mutation.isPending}
				class={$mutation.isPending ? 'bg-destructive/50' : 'bg-destructive'}
			>
				{#if $mutation.isPending}Deleting...{:else}Yes, delete{/if}
			</AlertDialog.Action>
		</AlertDialog.Footer>
	</AlertDialog.Content>
</AlertDialog.Root>
