<script lang="ts">
	import { buttonVariants } from '$lib/components/ui/button';
	import * as AlertDialog from '$lib/components/ui/alert-dialog'; // Import AlertDialog components
	import type { HTMLAnchorAttributes } from 'svelte/elements';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';

	export let scheduleId: number;

	let href: HTMLAnchorAttributes['href'] = `/admin/schedules/${scheduleId}/edit`;
	let isConfirmDialogOpen = false;

	const queryClient = useQueryClient();

	const deleteMutation = createMutation<Response, Error, number>({
		mutationFn: async (idToDelete) => {
			const response = await fetch(`/api/admin/schedules/${idToDelete}`, {
				method: 'DELETE'
			});
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: 'Failed to delete schedule' }));
				throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
			}
			return response;
		},
		onSuccess: () => {
			toast.success('Schedule deleted successfully!');
			queryClient.invalidateQueries({ queryKey: ['adminSchedules'] });
			isConfirmDialogOpen = false; // Close dialog on success
		},
		onError: (error) => {
			toast.error(`Error deleting schedule: ${error.message}`);
			isConfirmDialogOpen = false; // Close dialog on error too, or handle differently
		}
	});

	function confirmDelete() {
		$deleteMutation.mutate(scheduleId);
	}
</script>

<div class="flex space-x-2">
	<a {href} class={buttonVariants({ variant: 'outline', size: 'sm' })} role="button"> Edit </a>

	<AlertDialog.Root bind:open={isConfirmDialogOpen}>
		<AlertDialog.Trigger
			class={buttonVariants({ variant: 'destructive', size: 'sm' })}
			disabled={$deleteMutation.isPending}
		>
			{#if $deleteMutation.isPending}
				Deleting...
			{:else}
				Delete
			{/if}
		</AlertDialog.Trigger>
		<AlertDialog.Content>
			<AlertDialog.Header>
				<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
				<AlertDialog.Description>
					This action cannot be undone. This will permanently delete schedule ID {scheduleId}.
				</AlertDialog.Description>
			</AlertDialog.Header>
			<AlertDialog.Footer>
				<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
				<AlertDialog.Action onclick={confirmDelete} disabled={$deleteMutation.isPending}>
					Yes, delete schedule
				</AlertDialog.Action>
			</AlertDialog.Footer>
		</AlertDialog.Content>
	</AlertDialog.Root>
</div>
