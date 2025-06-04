<script lang="ts">
	// Types using our new domain types
	import type { Report } from '$lib/types/domain';

	// Utilities
	import { getSeverityIcon, getSeverityColor } from '$lib/utils/reports';
	import { formatRelativeTime } from '$lib/utils/dateFormatting';

	// Components and icons
	import ArchiveIcon from '@lucide/svelte/icons/archive';

	// API and state management
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { ReportsApiService } from '$lib/services/api/reports';
	import { toast } from 'svelte-sonner';

	let {
		report,
		isSelected = false,
		onSelect
	}: {
		report: Report;
		isSelected?: boolean;
		onSelect: (report: Report) => void;
	} = $props();

	const SeverityIcon = $derived(getSeverityIcon(report.severity ?? 0));
	const queryClient = useQueryClient();

	// Archive mutation (changed from delete)
	const archiveMutation = createMutation({
		mutationFn: ReportsApiService.archive,
		onSuccess: () => {
			toast.success('Report archived successfully');
			// Refresh the reports list
			queryClient.invalidateQueries({ queryKey: ['adminReportsForLayout'] });
			queryClient.invalidateQueries({ queryKey: ['adminReports'] });
		},
		onError: (error: Error) => {
			toast.error(`Failed to archive report: ${error.message}`);
		}
	});

	function handleArchive(event: Event) {
		event.preventDefault();
		event.stopPropagation();
		if (!report.id) {
			toast.error('Invalid report ID');
			return;
		}
		if (
			confirm('Are you sure you want to archive this report? You can unarchive it later if needed.')
		) {
			$archiveMutation.mutate(report.id);
		}
	}
</script>

<div
	class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0 relative group {isSelected
		? 'active'
		: ''}"
>
	<!-- Archive button - appears on hover -->
	<button
		class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 p-1 rounded hover:bg-orange-500/20 text-orange-600"
		onclick={handleArchive}
		disabled={$archiveMutation.isPending}
		title="Archive report"
		aria-label="Archive report"
	>
		<ArchiveIcon class="h-3 w-3" />
	</button>

	<a
		href={`/admin/reports?reportId=${report.id}`}
		class="flex items-center gap-2 w-full pr-6"
		onclick={(event) => {
			event.preventDefault();
			onSelect(report);
		}}
	>
		<div class="p-1 rounded {getSeverityColor(report.severity ?? 0)} bg-opacity-10">
			<SeverityIcon class="h-3 w-3 {getSeverityColor(report.severity ?? 0)}" />
		</div>
		<div class="flex-1 min-w-0">
			<div class="font-medium truncate">Report #{report.id}</div>
			<div class="text-xs text-muted-foreground truncate">
				{report.userName || 'Unknown'} • {formatRelativeTime(report.createdAt ?? '')}
			</div>
		</div>
	</a>
</div>
