import type { ColumnDef } from '@tanstack/svelte-table';
import { Checkbox } from '$lib/components/ui/checkbox';
import DataTableColumnHeader from '$lib/components/ui/data-table/data-table-column-header.svelte';
import DataTableRowActions from './schedule-actions.svelte';
import type { Schedule } from '$lib/types';
import CronView from '$lib/components/cron/cron-view.svelte';

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
		header: ({ table }) => {
			return Checkbox({
				checked: table.getIsAllPageRowsSelected()
					? true
					: table.getIsSomePageRowsSelected()
					? 'indeterminate'
					: false,
				onCheckedChange: (value) => table.toggleAllPageRowsSelected(!!value),
				'aria-label': 'Select all',
				class: 'translate-y-[2px]'
			});
		},
		cell: ({ row }) => {
			return Checkbox({
				checked: row.getIsSelected(),
				onCheckedChange: (value) => row.toggleSelected(!!value),
				'aria-label': 'Select row',
				class: 'translate-y-[2px]'
			});
		},
		enableSorting: false,
		enableHiding: false,
		size: 50
	},
	{
		accessorKey: 'schedule_id',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'ID' }),
		cell: (info) => info.getValue(),
		size: 50
	},
	{
		accessorKey: 'name',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'Name' }),
		cell: (info) => info.getValue() as string
	},
	{
		accessorKey: 'cron_expr',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'CRON' }),
		cell: (info) => info.getValue() as string
	},
	{
		accessorKey: 'is_active',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'Active' }),
		cell: (info) => (info.getValue() ? 'Yes' : 'No')
	},
	{
		accessorKey: 'start_date',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'Start' }),
		cell: (info) => formatDateToLocaleString(info.getValue() as string | null)
	},
	{
		accessorKey: 'end_date',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'End' }),
		cell: (info) => formatDateToLocaleString(info.getValue() as string | null)
	},
	{
		accessorKey: 'timezone',
		header: ({ column }) => DataTableColumnHeader({ column, title: 'Timezone' }),
		cell: (info) => (info.getValue() as string | null) || 'N/A'
	},
	{
		id: 'actions',
		header: () => 'Actions',
		cell: ({ row }) => DataTableRowActions({ scheduleId: row.original.schedule_id })
	}
];
