<script lang="ts" generics="TData, TValue">
	import type {
		ColumnDef,
		PaginationState,
		SortingState,
		ColumnFiltersState,
		VisibilityState,
		Updater,
		RowSelectionState
	} from '@tanstack/table-core';
	import {
		getCoreRowModel,
		getPaginationRowModel,
		getSortedRowModel,
		getFilteredRowModel
	} from '@tanstack/table-core';
	import { createSvelteTable, FlexRender } from '$lib/components/ui/data-table';
	import * as Table from '$lib/components/ui/table';
	import { Button } from '$lib/components/ui/button';
	import SchedulesTableControls from './schedules-table-controls.svelte';
	import SchedulesTablePagination from './schedules-table-pagination.svelte';

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
	});
</script>

<div class="w-full space-y-4">
	<!-- Filtering Input (Example for 'name' column) -->
	<!-- Bulk Delete Button and Dialog -->
	<!-- Column Visibility Dropdown -->
	<SchedulesTableControls
		{table}
		{columnFilters}
		setColumnFilters={(updater) => (columnFilters = updater)}
		{rowSelection}
		setRowSelection={(updater) => (rowSelection = updater)}
		{columnVisibility}
		setColumnVisibility={(updater) => (columnVisibility = updater)}
	/>

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
											{@const sorted = header.column.getIsSorted() as 'asc' | 'dsc' | undefined}
											{sorted ? (sorted === 'asc' ? ' ↑' : ' ↓') : ''}
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
	<SchedulesTablePagination {table} />
</div>
