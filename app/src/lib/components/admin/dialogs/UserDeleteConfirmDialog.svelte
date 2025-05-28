<script lang="ts">
	import DeleteConfirmDialog from './DeleteConfirmDialog.svelte';

	let {
		open = $bindable(false),
		userName = '',
		onConfirm = () => {},
		isLoading = false,
		user = null,
		mutation = null
	}: {
		open?: boolean;
		userName?: string;
		onConfirm?: () => void;
		isLoading?: boolean;
		user?: any;
		mutation?: any;
	} = $props();
</script>

<DeleteConfirmDialog
	bind:open
	title="Delete User"
	description="Are you sure you want to delete {userName}? This will permanently remove their account and all associated data. This action cannot be undone."
	{onConfirm}
	{isLoading}
/>
