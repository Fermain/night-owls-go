<script lang="ts" generics="TData">
	import type {
		Table,
		ColumnFiltersState,
		VisibilityState,
		RowSelectionState
	} from '@tanstack/table-core';
	import { Button, buttonVariants } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import ChevronDown from 'lucide-svelte/icons/chevron-down';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { createMutation, useQueryClient } from '@tanstack/svelte-query';
	import { toast } from 'svelte-sonner';

	// Props passed from the parent component
	let {
		table,
		columnFilters,
		setColumnFilters,
		setRowSelection,
	}: {
		table: Table<TData>;
		columnFilters: ColumnFiltersState;
		setColumnFilters: (updater: ColumnFiltersState) => void;
		rowSelection: RowSelectionState;
		setRowSelection: (updater: RowSelectionState) => void;
		columnVisibility: VisibilityState;
		setColumnVisibility: (updater: VisibilityState) => void;
	} = $props();

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
		setColumnFilters(newFilters);
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
			setRowSelection({}); // Clear selection
		},
		onError: (error) => {
			toast.error(`Error deleting schedules: ${error.message}`);
		}
	});

	let isBulkDeleteDialogOpen = $state(false);
	let pendingBulkDeleteIds = $state<number[]>([]);

	// Need to react to changes in rowSelection to update pendingBulkDeleteIds
	$effect(() => {
		const selectedOriginalRows = table.getSelectedRowModel().rows.map((row) => row.original);
		// Assuming TData has a schedule_id property for bulk delete functionality
		pendingBulkDeleteIds = selectedOriginalRows.map((schedule: any) => schedule.schedule_id);
	});
</script>

<div class="flex items-center py-4">
	<!-- Filtering Input (Example for 'name' column) -->
	<Input
		placeholder="Filter by name..."
		value={getFilterValue('name')}
		oninput={(event) => setFilterValue('name', event.currentTarget.value)}
		class="max-w-sm"
	/>

	{#if table.getSelectedRowModel().rows.length > 0}
		<AlertDialog.Root bind:open={isBulkDeleteDialogOpen}>
			<AlertDialog.Trigger>
				<Button
					variant="destructive"
					class="ml-4"
					disabled={$bulkDeleteMutation.isPending || pendingBulkDeleteIds.length === 0}
					onclick={() => {
						isBulkDeleteDialogOpen = true;
					}}
				>
					{#if $bulkDeleteMutation.isPending}
						Deleting...
					{:else}
						Delete Selected ({table.getSelectedRowModel().rows.length})
					{/if}
				</Button>
			</AlertDialog.Trigger>
			<AlertDialog.Content>
				<AlertDialog.Header>
					<AlertDialog.Title>Are you absolutely sure?</AlertDialog.Title>
					<AlertDialog.Description>
						This action cannot be undone. This will permanently delete {pendingBulkDeleteIds.length}
						schedule(s).
					</AlertDialog.Description>
				</AlertDialog.Header>
				<AlertDialog.Footer>
					<AlertDialog.Cancel
						onmousedown={() => {
							isBulkDeleteDialogOpen = false;
						}}
						onkeydown={() => {
							isBulkDeleteDialogOpen = false;
						}}>Cancel</AlertDialog.Cancel
					>
					<AlertDialog.Action
						onclick={() => {
							if (pendingBulkDeleteIds.length > 0) {
								$bulkDeleteMutation.mutate(pendingBulkDeleteIds);
							}
						}}
						disabled={$bulkDeleteMutation.isPending}
					>
						Yes, delete selected
					</AlertDialog.Action>
				</AlertDialog.Footer>
			</AlertDialog.Content>
		</AlertDialog.Root>
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
