<script lang="ts">
	import MessageSquareIcon from '@lucide/svelte/icons/message-square';
	import AdminPageHeader from '$lib/components/admin/AdminPageHeader.svelte';
	import BroadcastForm from '$lib/components/admin/broadcasts/BroadcastForm.svelte';
</script>

<svelte:head>
	<title>Admin - Broadcasts</title>
</svelte:head>

<div class="p-6">
	<div class="max-w-6xl mx-auto">
		<AdminPageHeader icon={MessageSquareIcon} heading="Send Push Alert" />

		<BroadcastForm />
	</div>
</div>
