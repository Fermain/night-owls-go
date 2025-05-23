<script lang="ts">
	import SidebarPage from '$lib/components/sidebar-page.svelte';
	import { createQuery } from '@tanstack/svelte-query';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import FileTextIcon from '@lucide/svelte/icons/file-text';
	import AlertTriangleIcon from '@lucide/svelte/icons/alert-triangle';
	import ShieldAlertIcon from '@lucide/svelte/icons/shield-alert';
	import InfoIcon from '@lucide/svelte/icons/info';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';

	let searchTerm = $state('');
	let { children } = $props();

	// Fetch shift reports (simulated for now since API exists)
	const reportsQuery = $derived(
		createQuery({
			queryKey: ['shiftReportsForLayout'],
			queryFn: async () => {
				// This would use the real API: GET /api/admin/reports
				// For now, simulate the data structure
				await new Promise((resolve) => setTimeout(resolve, 800));

				const mockReports = [
					{
						report_id: 1,
						booking_id: 123,
						message: 'Visitor seemed intoxicated and was asked to leave. No incidents.',
						severity: 1,
						created_at: new Date(Date.now() - 2 * 60 * 60 * 1000).toISOString(),
						user_name: 'John Doe',
						user_phone: '+27123456789',
						shift_start: new Date(Date.now() - 4 * 60 * 60 * 1000).toISOString(),
						schedule_name: 'Friday Night Security'
					},
					{
						report_id: 2,
						booking_id: 124,
						message: 'All quiet during shift. Routine patrol completed.',
						severity: 0,
						created_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
						user_name: 'Jane Smith',
						user_phone: '+27987654321',
						shift_start: new Date(Date.now() - 26 * 60 * 60 * 1000).toISOString(),
						schedule_name: 'Thursday Night Security'
					},
					{
						report_id: 3,
						booking_id: 125,
						message:
							'Attempted break-in at rear entrance. Police called and responded. Suspect fled.',
						severity: 2,
						created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
						user_name: 'Mike Johnson',
						user_phone: '+27555666777',
						shift_start: new Date(
							Date.now() - 3 * 24 * 60 * 60 * 1000 - 2 * 60 * 60 * 1000
						).toISOString(),
						schedule_name: 'Tuesday Night Security'
					},
					{
						report_id: 4,
						booking_id: 126,
						message: 'Minor equipment malfunction in security booth. Reported to maintenance.',
						severity: 1,
						created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
						user_name: 'Sarah Wilson',
						user_phone: '+27444555666',
						shift_start: new Date(
							Date.now() - 5 * 24 * 60 * 60 * 1000 - 1 * 60 * 60 * 1000
						).toISOString(),
						schedule_name: 'Monday Night Security'
					}
				];

				return mockReports;
			}
		})
	);

	function getSeverityIcon(severity: number) {
		switch (severity) {
			case 0:
				return InfoIcon;
			case 1:
				return AlertTriangleIcon;
			case 2:
				return ShieldAlertIcon;
			default:
				return InfoIcon;
		}
	}

	function getSeverityColor(severity: number) {
		switch (severity) {
			case 0:
				return 'text-blue-600';
			case 1:
				return 'text-orange-600';
			case 2:
				return 'text-red-600';
			default:
				return 'text-gray-600';
		}
	}

	function formatRelativeTime(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));

		if (diffInHours < 1) return 'Just now';
		if (diffInHours < 24) return `${diffInHours}h ago`;
		if (diffInHours < 48) return 'Yesterday';
		return `${Math.floor(diffInHours / 24)}d ago`;
	}

	// Filter reports based on search term
	const filteredReports = $derived.by(() => {
		const reports = $reportsQuery.data ?? [];
		if (!searchTerm) return reports;

		return reports.filter(
			(report) =>
				report.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
				report.user_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
				report.schedule_name.toLowerCase().includes(searchTerm.toLowerCase())
		);
	});
</script>

{#snippet reportsListContent()}
	<div class="flex flex-col h-full">
		<!-- Header -->
		<div class="p-3 border-b bg-muted/50">
			<div class="flex items-center gap-2">
				<FileTextIcon class="h-4 w-4" />
				<span class="text-sm font-medium">Recent Reports</span>
			</div>
			<p class="text-xs text-muted-foreground">Latest incident reports</p>
		</div>

		<!-- Reports List -->
		<div class="flex-grow overflow-y-auto">
			{#if $reportsQuery.isLoading}
				<div class="p-3 space-y-3">
					{#each Array(4) as _, i (i)}
						<div class="border rounded-lg p-3">
							<Skeleton class="h-4 w-3/4 mb-2" />
							<Skeleton class="h-3 w-1/2 mb-1" />
							<Skeleton class="h-3 w-1/3" />
						</div>
					{/each}
				</div>
			{:else if $reportsQuery.isError}
				<div class="p-3 text-center">
					<FileTextIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-xs text-muted-foreground">Failed to load reports</p>
				</div>
			{:else if filteredReports.length === 0}
				<div class="p-3 text-center">
					<FileTextIcon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-xs text-muted-foreground">
						{searchTerm ? `No reports match "${searchTerm}"` : 'No reports found'}
					</p>
				</div>
			{:else}
				<div class="p-2">
					{#each filteredReports as report (report.report_id)}
						<div class="border rounded-lg p-3 mb-2 hover:bg-accent transition-colors">
							<div class="space-y-2">
								{#snippet reportIcon()}
									{@const SeverityIcon = getSeverityIcon(report.severity)}
									<SeverityIcon class="h-4 w-4 {getSeverityColor(report.severity)}" />
								{/snippet}

								<div class="flex items-start gap-2">
									{@render reportIcon()}
									<div class="flex-1 min-w-0">
										<p class="text-sm font-medium line-clamp-2">{report.message}</p>
									</div>
								</div>

								<div class="space-y-1 text-xs text-muted-foreground">
									<div class="flex items-center gap-1">
										<UserIcon class="h-3 w-3" />
										<span class="truncate">{report.user_name}</span>
									</div>
									<div class="flex items-center gap-1">
										<ClockIcon class="h-3 w-3" />
										<span>{formatRelativeTime(report.created_at)}</span>
									</div>
									<div class="text-xs truncate">
										{report.schedule_name}
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	</div>
{/snippet}

<SidebarPage listContent={reportsListContent} title="Reports" bind:searchTerm>
	{@render children()}
</SidebarPage>
