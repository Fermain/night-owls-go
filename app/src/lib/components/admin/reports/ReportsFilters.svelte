<script lang="ts">
	import * as Select from '$lib/components/ui/select';
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import DateRangePicker from '$lib/components/ui/date-range-picker/DateRangePicker.svelte';
	import SearchIcon from '@lucide/svelte/icons/search';
	import FilterIcon from '@lucide/svelte/icons/filter';
	import XIcon from '@lucide/svelte/icons/x';
	import { getSeverityIcon, getSeverityColor, getSeverityLabel } from '$lib/utils/reports';

	let {
		searchQuery = $bindable(''),
		severityFilter = $bindable('all'),
		scheduleFilter = $bindable('all'),
		dateRangeStart = $bindable(null),
		dateRangeEnd = $bindable(null),
		sortBy = $bindable('newest'),
		resultCount = 0
	}: {
		searchQuery: string;
		severityFilter: string;
		scheduleFilter: string;
		dateRangeStart: string | null;
		dateRangeEnd: string | null;
		sortBy: string;
		resultCount?: number;
	} = $props();

	// Filter options
	const severityOptions = [
		{ value: 'all', label: 'All Severities' },
		{ value: '0', label: 'Normal' },
		{ value: '1', label: 'Suspicion' },
		{ value: '2', label: 'Incident' }
	];

	const scheduleOptions = [
		{ value: 'all', label: 'All Schedules' },
		{ value: 'Old schedule', label: 'Old schedule' },
		{ value: 'New schedule', label: 'New schedule' }
	];

	const sortOptions = [
		{ value: 'newest', label: 'Newest First' },
		{ value: 'oldest', label: 'Oldest First' },
		{ value: 'severity', label: 'Severity (High to Low)' },
		{ value: 'schedule', label: 'Schedule Name' }
	];

	// Check if any filters are active
	const hasActiveFilters = $derived.by(() => {
		return (
			searchQuery.trim() !== '' ||
			severityFilter !== 'all' ||
			scheduleFilter !== 'all' ||
			dateRangeStart !== null ||
			dateRangeEnd !== null ||
			sortBy !== 'newest'
		);
	});

	function clearFilters() {
		searchQuery = '';
		severityFilter = 'all';
		scheduleFilter = 'all';
		dateRangeStart = null;
		dateRangeEnd = null;
		sortBy = 'newest';
	}

	function handleDateRangeChange(range: { start: string | null; end: string | null }) {
		dateRangeStart = range.start;
		dateRangeEnd = range.end;
	}
</script>

<div class="bg-card border rounded-lg p-4 space-y-4">
	<!-- Header with results count -->
	<div class="flex items-center justify-between">
		<div class="flex items-center gap-2">
			<FilterIcon class="h-4 w-4 text-muted-foreground" />
			<span class="text-sm font-medium">Filters</span>
			{#if hasActiveFilters}
				<Badge variant="secondary" class="text-xs">
					{resultCount} results
				</Badge>
			{/if}
		</div>
		{#if hasActiveFilters}
			<Button variant="ghost" size="sm" onclick={clearFilters} class="h-8 px-2">
				<XIcon class="h-3 w-3 mr-1" />
				Clear
			</Button>
		{/if}
	</div>

	<!-- Search bar -->
	<div class="relative">
		<SearchIcon
			class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground"
		/>
		<Input
			bind:value={searchQuery}
			placeholder="Search reports by message, user, schedule, or severity..."
			class="pl-10"
		/>
	</div>

	<!-- Filter controls in a compact grid -->
	<div class="grid grid-cols-2 md:grid-cols-4 gap-3">
		<!-- Severity Filter -->
		<div class="space-y-1">
			<Label class="text-xs text-muted-foreground">Severity</Label>
			<Select.Root type="single" bind:value={severityFilter}>
				<Select.Trigger class="h-9">
					{#if severityFilter === 'all'}
						All Severities
					{:else}
						{@const SeverityIcon = getSeverityIcon(parseInt(severityFilter))}
						<div class="flex items-center gap-2">
							<SeverityIcon class="h-3 w-3 {getSeverityColor(parseInt(severityFilter))}" />
							{getSeverityLabel(parseInt(severityFilter))}
						</div>
					{/if}
				</Select.Trigger>
				<Select.Content>
					{#each severityOptions as option (option.value)}
						<Select.Item value={option.value} label={option.label}>
							{#if option.value === 'all'}
								{option.label}
							{:else}
								{@const SeverityIcon = getSeverityIcon(parseInt(option.value))}
								<div class="flex items-center gap-2">
									<SeverityIcon class="h-3 w-3 {getSeverityColor(parseInt(option.value))}" />
									{option.label}
								</div>
							{/if}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- Schedule Filter -->
		<div class="space-y-1">
			<Label class="text-xs text-muted-foreground">Schedule</Label>
			<Select.Root type="single" bind:value={scheduleFilter}>
				<Select.Trigger class="h-9">
					{scheduleOptions.find((opt) => opt.value === scheduleFilter)?.label ?? 'Select schedule'}
				</Select.Trigger>
				<Select.Content>
					{#each scheduleOptions as option (option.value)}
						<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- Date Range Filter -->
		<div class="space-y-1">
			<Label class="text-xs text-muted-foreground">Date Range</Label>
			<DateRangePicker
				initialStartDate={dateRangeStart}
				initialEndDate={dateRangeEnd}
				change={handleDateRangeChange}
				placeholderText="Select range"
			/>
		</div>

		<!-- Sort Filter -->
		<div class="space-y-1">
			<Label class="text-xs text-muted-foreground">Sort By</Label>
			<Select.Root type="single" bind:value={sortBy}>
				<Select.Trigger class="h-9">
					{sortOptions.find((opt) => opt.value === sortBy)?.label ?? 'Select sort'}
				</Select.Trigger>
				<Select.Content>
					{#each sortOptions as option (option.value)}
						<Select.Item value={option.value} label={option.label}>{option.label}</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>
	</div>
</div>
