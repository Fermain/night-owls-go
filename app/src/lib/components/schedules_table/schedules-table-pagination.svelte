<script lang="ts" generics="TData">
	import type { Table } from '@tanstack/table-core';
	import { Button } from '$lib/components/ui/button';
	import * as Pagination from '$lib/components/ui/pagination';
	import ChevronLeft from 'lucide-svelte/icons/chevron-left';
	import ChevronRight from 'lucide-svelte/icons/chevron-right';

	// Props passed from the parent component
	let { table }: { table: Table<TData> } = $props();
</script>

<!-- Pagination Controls -->
{#if table.getPageCount() > 1}
	<Pagination.Root
		count={table.getPageCount()}
		perPage={table.getState().pagination.pageSize}
		page={table.getState().pagination.pageIndex + 1}
	>
		{#snippet children({ pages, currentPage })}
			<Pagination.Content>
				<Pagination.Item>
					<Pagination.PrevButton
						onclick={() => table.previousPage()}
						disabled={!table.getCanPreviousPage()}
					>
						<ChevronLeft class="h-4 w-4" />
						<span class="sr-only">Previous page</span>
					</Pagination.PrevButton>
				</Pagination.Item>

				{#each pages as page (page.key)}
					{#if page.type === 'ellipsis'}
						<Pagination.Item>
							<Pagination.Ellipsis />
						</Pagination.Item>
					{:else}
						<Pagination.Item>
							<Pagination.Link
								{page}
								isActive={currentPage === page.value}
								onclick={() => table.setPageIndex(page.value - 1)}
							>
								{page.value}
							</Pagination.Link>
						</Pagination.Item>
					{/if}
				{/each}

				<Pagination.Item>
					<Pagination.NextButton
						onclick={() => table.nextPage()}
						disabled={!table.getCanNextPage()}
					>
						<span class="sr-only">Next page</span>
						<ChevronRight class="h-4 w-4" />
					</Pagination.NextButton>
				</Pagination.Item>
			</Pagination.Content>
		{/snippet}
	</Pagination.Root>
{/if}
