<script lang="ts">
	import PushNotificationDebug from '$lib/components/debug/PushNotificationDebug.svelte';
	import SidebarPage from '$lib/components/sidebar-page.svelte';
</script>

<svelte:head>
	<title>Push Notification Testing</title>
</svelte:head>

<SidebarPage title="Push Alerts Test">
	<div class="space-y-6 p-4">
		<div>
			<h1 class="text-2xl font-bold tracking-tight">Push Notification Testing</h1>
			<p class="text-muted-foreground">
				Debug and test push notification functionality with the live backend
			</p>
		</div>

		<PushNotificationDebug />
	</div>
</SidebarPage>
