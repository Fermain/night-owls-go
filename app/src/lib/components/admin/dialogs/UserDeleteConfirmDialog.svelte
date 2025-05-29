<script lang="ts">
	import DeleteConfirmDialog from './DeleteConfirmDialog.svelte';
	import type { CreateMutationResult } from '@tanstack/svelte-query';

	interface User {
		id: number;
		name: string | null;
		[key: string]: unknown;
	}

	let {
		open = $bindable(false),
		userName = '',
		onConfirm = () => {},
		isLoading = false,
		_user = null,
		_mutation = null
	}: {
		open?: boolean;
		userName?: string;
		onConfirm?: () => void;
		isLoading?: boolean;
		_user?: User | null;
		_mutation?: CreateMutationResult<unknown, Error, number, unknown> | null;
	} = $props();
</script>

<DeleteConfirmDialog
	bind:open
	title="Delete User"
	description="Are you sure you want to delete {userName}? This will permanently remove their account and all associated data. This action cannot be undone."
	{onConfirm}
	{isLoading}
/>
