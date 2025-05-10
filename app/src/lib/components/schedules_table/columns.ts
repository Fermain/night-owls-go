import type { ColumnDef } from '@tanstack/table-core';
import { createRawSnippet } from 'svelte';
import { renderSnippet, renderComponent } from '$lib/components/ui/data-table';
import CronView from '$lib/components/cron/cron-view.svelte';
import ScheduleActions from './schedule-actions.svelte';
import Checkbox from '$lib/components/ui/checkbox/checkbox.svelte';

// Define types for how Go's sql.NullString and sql.NullTime are serialized to JSON
export type SQLNullString = {
	String: string;
	Valid: boolean;
};

export type SQLNullTime = {
	Time: string; // Typically an ISO 8601 string from Go's JSON marshaller for time.Time
	Valid: boolean;
};

// This type is used to define the shape of our data.
// We are fetching this from the Go backend.
// Why aren't we using a derived or centralised type?
export type Schedule = {
	schedule_id: number;
	name: string;
	cron_expr: string;
	start_date?: string | null;
	end_date?: string | null;
	timezone?: string | null;
};

// Helper to format date strings from SQLNullTime or return 'N/A'
// This seems over-engineered and I am sure it can be made more concise.
// Also not the responsibility of this component to format dates.
const formatDate = (dateString?: string | null): string => {
	if (!dateString) return 'N/A';
	try {
		// The date from Go might be a full timestamp, extract date part for display.
		return new Date(dateString).toLocaleDateString(undefined, {
			year: 'numeric',
			month: '2-digit',
			day: '2-digit'
		});
	} catch {
		return 'Invalid Date';
	}
};

export const columns: ColumnDef<Schedule>[] = [
	{
		id: 'select',
		header: ({ table }) => {
			const isAllPageRowsSelected = table.getIsAllPageRowsSelected();
			const isSomePageRowsSelected = table.getIsSomePageRowsSelected();

			return renderComponent(Checkbox, {
				checked: isAllPageRowsSelected,
				indeterminate: isSomePageRowsSelected && !isAllPageRowsSelected,
				onCheckedChange: (checked: boolean) => {
					table.toggleAllPageRowsSelected(checked);
				},
				'aria-label': 'Select all rows on current page'
			});
		},
		cell: ({ row }) => {
			return renderComponent(Checkbox, {
				checked: row.getIsSelected(),
				onCheckedChange: (checked: boolean) => {
					row.toggleSelected(checked);
				},
				'aria-label': 'Select row',
				disabled: !row.getCanSelect()
			});
		},
		enableSorting: false,
		enableHiding: false,
		size: 40
	},
	{
		accessorKey: 'name',
		header: 'Name',
		cell: ({ row }) => {
			const name = row.getValue('name') as string;
			// Using createRawSnippet for simple text display; can be enhanced
			const nameSnippet = createRawSnippet<[string]>((getName) => {
				const val = getName();
				return {
					render: () => `<div>${val}</div>`
				};
			});
			return renderSnippet(nameSnippet, name);
		}
	},
	{
		accessorKey: 'cron_expr',
		header: 'Cron Expression',
		cell: ({ row }) => {
			const cronExpr = row.getValue('cron_expr') as string;
			// Display raw CRON expression string
			const cronSnippet = createRawSnippet<[string]>((getExpr) => {
				const val = getExpr();
				return {
					render: () => `<pre>${val}</pre>`
				};
			});
			return renderSnippet(cronSnippet, cronExpr);
		}
	},
	{
		id: 'cron_visualization',
		header: 'Schedule Visualization',
		cell: ({ row }) => {
			const cronExpr = row.original.cron_expr;
			return renderComponent(CronView, { cronExpr: cronExpr });
		},
		enableSorting: false
	},
	{
		accessorKey: 'start_date',
		header: 'Start Date',
		cell: ({ row }) => {
			const startDate = formatDate(row.getValue('start_date') as string | null | undefined);
			const snippet = createRawSnippet<[string]>((getDate) => {
				const val = getDate();
				return { render: () => `<div>${val}</div>` };
			});
			return renderSnippet(snippet, startDate);
		}
	},
	{
		accessorKey: 'end_date',
		header: 'End Date',
		cell: ({ row }) => {
			const endDate = formatDate(row.getValue('end_date') as string | null | undefined);
			const snippet = createRawSnippet<[string]>((getDate) => {
				const val = getDate();
				return { render: () => `<div>${val}</div>` };
			});
			return renderSnippet(snippet, endDate);
		}
	},
	{
		id: 'actions',
		header: 'Actions',
		cell: ({ row }) => {
			const schedule = row.original;
			return renderComponent(ScheduleActions, { scheduleId: schedule.schedule_id });
		},
		enableSorting: false,
		enableHiding: false
	}
];
