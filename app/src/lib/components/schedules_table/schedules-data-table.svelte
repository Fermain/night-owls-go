<script lang="ts" generics="TData, TValue">
	import type {
		ColumnDef,
		PaginationState,
		SortingState,
		ColumnFiltersState,
		VisibilityState,
		Updater,
		FilterFn,
		RowSelectionState
	} from '@tanstack/table-core';
	import {
		getCoreRowModel,
		getPaginationRowModel,
		getSortedRowModel,
		getFilteredRowModel,
		getFacetedRowModel,
		getFacetedUniqueValues,
		getFacetedMinMaxValues
	} from '@tanstack/table-core';
	import { createSvelteTable, FlexRender } from '$lib/components/ui/data-table';
	import * as Table from '$lib/components/ui/table';
	import { Button, buttonVariants } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import ChevronDown from 'lucide-svelte/icons/chevron-down';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';
	import type { Schedule } from './columns'; // Assuming Schedule type is exported from columns.ts

	// Define the props for this component
	type DataTableProps = {
		columns: ColumnDef<TData, TValue>[];
		data: TData[];
	};

	let { data, columns }: DataTableProps = $props();

	// Table state using Svelte 5 runes
	let pagination = $state<PaginationState>({ pageIndex: 0, pageSize: 10 });
	let sorting = $state<SortingState>([]);
	let columnFilters = $state<ColumnFiltersState>([]);
	let columnVisibility = $state<VisibilityState>({});
	// For row selection (optional, can be added later if needed)
	let rowSelection = $state<RowSelectionState>({});

	const table = createSvelteTable<TData>({
		// Getter for data ensures reactivity if the prop changes
		get data() {
			return data;
		},
		columns,
		state: {
			// Pass reactive state variables
			get pagination() {
				return pagination;
			},
			get sorting() {
				return sorting;
			},
			get columnFilters() {
				return columnFilters;
			},
			get columnVisibility() {
				return columnVisibility;
			},
			get rowSelection() {
				return rowSelection;
			} // If using row selection
		},
		enableRowSelection: true, // enable row selection
		// Enable features by providing their row model getters
		getCoreRowModel: getCoreRowModel(),
		getPaginationRowModel: getPaginationRowModel(),
		getSortedRowModel: getSortedRowModel(),
		getFilteredRowModel: getFilteredRowModel(),
		// Event handlers to update state
		onPaginationChange: (updater: Updater<PaginationState>) => {
			pagination = typeof updater === 'function' ? updater(pagination) : updater;
		},
		onSortingChange: (updater: Updater<SortingState>) => {
			sorting = typeof updater === 'function' ? updater(sorting) : updater;
		},
		onColumnFiltersChange: (updater: Updater<ColumnFiltersState>) => {
			columnFilters = typeof updater === 'function' ? updater(columnFilters) : updater;
		},
		onColumnVisibilityChange: (updater: Updater<VisibilityState>) => {
			columnVisibility = typeof updater === 'function' ? updater(columnVisibility) : updater;
		},
		onRowSelectionChange: (updater: Updater<RowSelectionState>) => {
			// If using row selection
			rowSelection = typeof updater === 'function' ? updater(rowSelection) : updater;
		}
		// Optional: default column settings, e.g., min size
		// defaultColumn: {
		//   minSize: 20,
		//   maxSize: 500,
		// },
		// debugTable: dev, // Enable debug logging in dev mode
	});

	// Helper to get filter value for a column (e.g., for an input binding)
	const getFilterValue = (columnId: string): string => {
		const filter = columnFilters.find((f: { id: string; value: unknown }) => f.id === columnId);
		return (filter?.value as string) ?? '';
	};

	// Helper to set filter value for a column
	const setFilterValue = (columnId: string, value: any) => {
		// Remove existing filter for this column
		const newFilters = columnFilters.filter((f: { id: string }) => f.id !== columnId);
		// Add new filter if value is not empty
		if (value !== null && value !== undefined && value !== '') {
			newFilters.push({ id: columnId, value });
		}
		columnFilters = newFilters;
	};

	const queryClient = useQueryClient();

	const bulkDeleteMutation = createMutation<unknown, Error, number[], unknown>({
		mutationFn: async (scheduleIds: number[]) => {
			const response = await fetch('/api/admin/schedules', {
				method: 'DELETE',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ schedule_ids: scheduleIds })
			});
			if (!response.ok) {
				const errorData = await response
					.json()
					.catch(() => ({ message: 'Failed to delete schedules. Unknown error.' }));
				throw new Error(errorData.message || 'Failed to delete schedules.');
			}
			return response.json();
		},
		onSuccess: () => {
			toast.success('Schedules deleted successfully!');
			queryClient.invalidateQueries({ queryKey: ['adminSchedules'] }); // Matches the queryKey in admin/schedules/+page.svelte
			rowSelection = {}; // Clear selection
		},
		onError: (error) => {
			toast.error(`Error deleting schedules: ${error.message}`);
		}
	});
</script>

<div class="w-full space-y-4">
	<!-- Filtering Input (Example for 'name' column) -->
	<div class="flex items-center py-4">
		<Input
			placeholder="Filter by name..."
			value={getFilterValue('name')}
			oninput={(event) => setFilterValue('name', event.currentTarget.value)}
			class="max-w-sm"
		/>

		{#if table.getSelectedRowModel().rows.length > 0}
			<Button
				variant="destructive"
				class="ml-4"
				disabled={$bulkDeleteMutation.isPending}
				onclick={() => {
					const selectedOriginalRows = table
						.getSelectedRowModel()
						.rows.map((row) => row.original as Schedule);
					const selectedIds = selectedOriginalRows.map((schedule) => schedule.schedule_id);
					if (selectedIds.length > 0) {
						$bulkDeleteMutation.mutate(selectedIds);
					}
				}}
			>
				{#if $bulkDeleteMutation.isPending}
					Deleting...
				{:else}
					Delete Selected ({table.getSelectedRowModel().rows.length})
				{/if}
			</Button>
		{/if}

		<!-- Column Visibility Dropdown -->
		<DropdownMenu.Root>
			<DropdownMenu.Trigger class={buttonVariants({ variant: 'outline', class: 'ml-auto' })}>
				Columns <ChevronDown class="ml-2 h-4 w-4" />
			</DropdownMenu.Trigger>
			<DropdownMenu.Content align="end">
				{#each table.getAllColumns().filter((col) => col.getCanHide()) as column (column.id)}
					<DropdownMenu.CheckboxItem
						class="capitalize"
						checked={column.getIsVisible()}
						onCheckedChange={(value) => {
							column.toggleVisibility(!!value);
						}}
					>
						{column.id}
					</DropdownMenu.CheckboxItem>
				{/each}
			</DropdownMenu.Content>
		</DropdownMenu.Root>
	</div>

	<!-- Table -->
	<div class="rounded-md border">
		<Table.Root>
			<Table.Header>
				{#each table.getHeaderGroups() as headerGroup (headerGroup.id)}
					<Table.Row>
						{#each headerGroup.headers as header (header.id)}
							<Table.Head class="[&:has([role=checkbox])]:pl-3">
								{#if !header.isPlaceholder}
									{@const sortHandler = header.column.getCanSort()
										? header.column.getToggleSortingHandler()
										: undefined}
									{#if sortHandler}
										<Button variant="ghost" onclick={sortHandler}>
											<FlexRender
												content={header.column.columnDef.header}
												context={header.getContext()}
											/>
											{{ asc: ' ↑', desc: ' ↓' }[header.column.getIsSorted() as string] ?? ''}
										</Button>
									{:else}
										<FlexRender
											content={header.column.columnDef.header}
											context={header.getContext()}
										/>
									{/if}
								{/if}
							</Table.Head>
						{/each}
					</Table.Row>
				{/each}
			</Table.Header>
			<Table.Body>
				{#if table.getRowModel().rows.length}
					{#each table.getRowModel().rows as row (row.id)}
						<Table.Row data-state={row.getIsSelected() && 'selected'}>
							{#each row.getVisibleCells() as cell (cell.id)}
								<Table.Cell class="[&:has([role=checkbox])]:pl-3">
									<FlexRender content={cell.column.columnDef.cell} context={cell.getContext()} />
								</Table.Cell>
							{/each}
						</Table.Row>
					{/each}
				{:else}
					<Table.Row>
						<Table.Cell colspan={columns.length} class="h-24 text-center">No results.</Table.Cell>
					</Table.Row>
				{/if}
			</Table.Body>
		</Table.Root>
	</div>

	<!-- Pagination Controls -->
	<div class="flex items-center justify-between space-x-2 py-4">
		<div class="text-muted-foreground flex-1 text-sm">
			<!-- Row selection count (if enabled) -->
			{table.getFilteredSelectedRowModel().rows.length} of {' '}
			{table.getFilteredRowModel().rows.length} row(s) selected. ({table.getFilteredRowModel().rows
				.length} total rows found)
		</div>
		<div class="flex items-center space-x-2">
			<Button
				variant="outline"
				size="sm"
				onclick={() => table.setPageIndex(0)}
				disabled={!table.getCanPreviousPage()}
			>
				First
			</Button>
			<Button
				variant="outline"
				size="sm"
				onclick={() => table.previousPage()}
				disabled={!table.getCanPreviousPage()}
			>
				Previous
			</Button>
			<span class="text-sm">
				Page {table.getState().pagination.pageIndex + 1} of {table.getPageCount() > 0
					? table.getPageCount()
					: 1}
			</span>
			<Button
				variant="outline"
				size="sm"
				onclick={() => table.nextPage()}
				disabled={!table.getCanNextPage()}
			>
				Next
			</Button>
			<Button
				variant="outline"
				size="sm"
				onclick={() => table.setPageIndex(table.getPageCount() - 1)}
				disabled={!table.getCanNextPage()}
			>
				Last
			</Button>
		</div>
		<div class="flex items-center space-x-2">
			<span class="text-sm">Rows per page:</span>
			<select
				class="border-input bg-background ring-offset-background placeholder:text-muted-foreground focus-visible:ring-ring flex h-10 w-auto rounded-md border px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
				value={table.getState().pagination.pageSize}
				onchange={(e) => {
					table.setPageSize(Number(e.currentTarget.value));
				}}
			>
				{#each [10, 20, 30, 40, 50] as pageSize (pageSize)}
					<option value={pageSize}>
						{pageSize}
					</option>
				{/each}
			</select>
		</div>
	</div>
</div>
