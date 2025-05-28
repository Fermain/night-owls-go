<script lang="ts">
	import type { components } from '$lib/types/api';
	import { getSeverityIcon, getSeverityColor } from '$lib/utils/reports';
	import { formatRelativeTime } from '$lib/utils/dateFormatting';

	type Report = components['schemas']['api.ReportResponse'] & {
		user_name: string;
		schedule_name: string;
	};

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
</script>

<div
	class="hover:bg-sidebar-accent hover:text-sidebar-accent-foreground flex items-center gap-2 whitespace-nowrap border-b p-4 text-sm leading-tight last:border-b-0 {isSelected
		? 'active'
		: ''}"
>
	<a
		href={`/admin/reports?reportId=${report.report_id}`}
		class="flex items-center gap-2 w-full"
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
