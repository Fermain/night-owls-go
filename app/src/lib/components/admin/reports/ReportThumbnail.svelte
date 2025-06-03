<script lang="ts">
	import type { components } from '$lib/types/api';
	import { getSeverityIcon, getSeverityColor } from '$lib/utils/reports';
	import { formatRelativeTime } from '$lib/utils/dateFormatting';
	import TrashIcon from '@lucide/svelte/icons/trash-2';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { ReportsApiService } from '$lib/services/api/reports';
	import { toast } from 'svelte-sonner';

	type AdminReport = components['schemas']['api.AdminReportResponse'];

	let {
		report,
		isSelected = false,
		onSelect
	}: {
		report: AdminReport;
		isSelected?: boolean;
		onSelect: (report: AdminReport) => void;
	} = $props();

	const SeverityIcon = $derived(getSeverityIcon(report.severity ?? 0));
	const queryClient = useQueryClient();

	// Delete mutation
	const deleteMutation = createMutation({
		mutationFn: ReportsApiService.delete,
		onSuccess: () => {
			toast.success('Report deleted successfully');
			// Refresh the reports list
			queryClient.invalidateQueries({ queryKey: ['adminReportsForLayout'] });
			queryClient.invalidateQueries({ queryKey: ['adminReports'] });
		},
		onError: (error: Error) => {
			toast.error(`Failed to delete report: ${error.message}`);
		}
	});

	function handleDelete(event: Event) {
		event.preventDefault();
		event.stopPropagation();
		if (!report.report_id) {
			toast.error('Invalid report ID');
			return;
		}
		if (confirm('Are you sure you want to delete this report? This action cannot be undone.')) {
			$deleteMutation.mutate(report.report_id);
		}
	}
</script>

<div
	class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0 relative group {isSelected
		? 'active'
		: ''}"
>
	<!-- Delete button - appears on hover -->
	<button
		class="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity duration-200 p-1 rounded hover:bg-destructive/20 text-destructive"
		onclick={handleDelete}
		disabled={$deleteMutation.isPending}
		title="Delete report"
		aria-label="Delete report"
	>
		<TrashIcon class="h-3 w-3" />
	</button>

	<a
		href={`/admin/reports?reportId=${report.report_id}`}
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
			<div class="font-medium truncate">Report #{report.report_id}</div>
			<div class="text-xs text-muted-foreground truncate">
				{report.user_name} • {formatRelativeTime(report.created_at ?? '')}
			</div>
		</div>
	</a>
</div>
