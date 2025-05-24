import type { ColumnDef, HeaderContext, CellContext } from '@tanstack/table-core';
// import DataTableColumnHeader from '$lib/components/ui/data-table/data-table-column-header.svelte'; // MISSING
// import DataTableRowActions from './schedule-actions.svelte'; // MISSING
import type { Schedule } from '$lib/types';

const formatDateToLocaleString = (dateString?: string | null): string => {
	if (!dateString) return 'N/A';
	try {
		return new Date(dateString).toLocaleDateString();
	} catch {
		return 'Invalid Date';
	}
};

export const columns: ColumnDef<Schedule>[] = [
	{
		id: 'select',
		header: ({ table }: HeaderContext<Schedule, unknown>) => {
			// Incorrect Checkbox usage - this will not render a Svelte component correctly.
			// Needs a Svelte-specific cell renderer or to be used in a .svelte file.
			// Returning a placeholder string to satisfy type checks for now.
			return 'SelectAll'; // Placeholder
		},
		cell: ({ row }: CellContext<Schedule, unknown>) => {
			// Incorrect Checkbox usage - placeholder string.
			return 'SelectRow'; // Placeholder
		},
		enableSorting: false,
		enableHiding: false,
		size: 50
	},
	{
		accessorKey: 'schedule_id',
		header: 'ID',
		cell: (info) => info.getValue() // TanStack table can infer info type here
	},
	{
		accessorKey: 'name',
		header: 'Name',
		cell: (info) => info.getValue()
	},
	{
		accessorKey: 'cron_expr',
		header: 'CRON',
		cell: (info) => info.getValue()
	},
	{
		accessorKey: 'is_active',
		header: 'Active',
		cell: (info) => (info.getValue() ? 'Yes' : 'No')
	},
	{
		accessorKey: 'start_date',
		header: 'Start',
		cell: (info) => formatDateToLocaleString(info.getValue() as string | null)
	},
	{
		accessorKey: 'end_date',
		header: 'End',
		cell: (info) => formatDateToLocaleString(info.getValue() as string | null)
	},
	{
		accessorKey: 'timezone',
		header: 'Timezone',
		cell: (info) => (info.getValue() as string | null) || 'N/A'
	},
	{
		id: 'actions',
		header: 'Actions',
		cell: ({ row }) => `Actions for ${row.original.schedule_id}`
	}
];
