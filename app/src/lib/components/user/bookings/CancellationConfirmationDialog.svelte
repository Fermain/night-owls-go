<script lang="ts">
	import * as Dialog from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';

	let {
		open = $bindable(false),
		shiftDetails = '',
		isLoading = false,
		onConfirm,
		onCancel
	}: {
		open: boolean;
		shiftDetails: string;
		isLoading?: boolean;
		onConfirm: () => void;
		onCancel: () => void;
	} = $props();

	function handleCancel() {
		onCancel();
		open = false;
	}

	function handleConfirm() {
		onConfirm();
	}
</script>

<Dialog.Root bind:open>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title class="flex items-center gap-2">
				<AlertTriangleIcon class="h-5 w-5" />
				Cancel Shift
			</Dialog.Title>
			<Dialog.Description>
				Are you sure you want to cancel your commitment to this shift?
			</Dialog.Description>
		</Dialog.Header>

		<div
			class="p-2 bg-amber-50 dark:bg-amber-950/30 border border-amber-200 dark:border-amber-800 rounded-lg"
		>
			<p class="text-sm text-amber-700 dark:text-amber-300">
				{shiftDetails}
			</p>
		</div>

		<Dialog.Footer>
			<Button variant="outline" onclick={handleCancel} disabled={isLoading}>Keep Shift</Button>
			<Button variant="destructive" onclick={handleConfirm} disabled={isLoading}>
				{isLoading ? 'Cancelling...' : 'Cancel Shift'}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
