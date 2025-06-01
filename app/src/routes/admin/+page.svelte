<script lang="ts">
	import AdminDashboard from '$lib/components/admin/AdminDashboard.svelte';
	import MobileAdminDashboard from '$lib/components/admin/MobileAdminDashboard.svelte';
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import UpcomingShifts from '$lib/components/admin/shifts/UpcomingShifts.svelte';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import SmartphoneIcon from '@lucide/svelte/icons/smartphone';
	import MonitorIcon from '@lucide/svelte/icons/monitor';
	import { createAdminDashboardQuery } from '$lib/queries/admin/dashboard';

	// Create the comprehensive dashboard query
	const dashboardQuery = $derived(createAdminDashboardQuery());

	const isLoading = $derived($dashboardQuery.isLoading);
	const isError = $derived($dashboardQuery.isError);
	const error = $derived($dashboardQuery.error || undefined);
	const dashboardData = $derived($dashboardQuery.data);

	// View mode state
	let mobileView = $state(false);

	// Auto-detect mobile on load
	$effect(() => {
		if (typeof window !== 'undefined') {
			const isMobile = window.innerWidth < 768;
			mobileView = isMobile;
		}
	});
</script>

<svelte:head>
	<title>Admin Dashboard</title>
</svelte:head>

<SidebarPage title="Upcoming Shifts">
	{#snippet listContent()}
		<UpcomingShifts maxItems={8} />
	{/snippet}

	<div class="space-y-6">
		<!-- View Toggle -->
		<div class="flex items-center justify-between">
			<div>
				<h1 class="text-2xl font-bold tracking-tight">Admin Dashboard</h1>
				<p class="text-muted-foreground">
					{mobileView
						? 'Mobile-optimized view for quick insights'
						: 'Comprehensive dashboard with detailed metrics'}
				</p>
			</div>
			<div class="flex items-center gap-2">
				<Button
					variant={!mobileView ? 'default' : 'outline'}
					size="sm"
					onclick={() => (mobileView = false)}
					class="gap-2"
				>
					<MonitorIcon class="h-4 w-4" />
					Desktop
				</Button>
				<Button
					variant={mobileView ? 'default' : 'outline'}
					size="sm"
					onclick={() => (mobileView = true)}
					class="gap-2"
				>
					<SmartphoneIcon class="h-4 w-4" />
					Mobile
					<Badge variant="secondary" class="text-xs">NEW</Badge>
				</Button>
			</div>
		</div>

		<!-- Dashboard Content -->
		{#if mobileView}
			<MobileAdminDashboard {isLoading} {isError} {error} data={dashboardData} />
		{:else}
			<AdminDashboard {isLoading} {isError} {error} data={dashboardData} />
		{/if}
	</div>
</SidebarPage>
