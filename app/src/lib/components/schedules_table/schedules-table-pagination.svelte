<script lang="ts" generics="TData">
	import type { Table, PaginationState, Updater } from '@tanstack/table-core';
	import { Button } from '$lib/components/ui/button';

	// Props passed from the parent component
	let { table }: { table: Table<TData> } = $props();
</script>

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
