<script lang="ts">
	import { createUpcomingShiftsQuery } from '$lib/queries/admin/shifts';
	import { Badge } from '$lib/components/ui/badge';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { goto } from '$app/navigation';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UserIcon from '@lucide/svelte/icons/user';
	import UserCheckIcon from '@lucide/svelte/icons/user-check';
	import ExternalLinkIcon from '@lucide/svelte/icons/external-link';
	
	// Props for customization
	let {
		maxItems = 10,
		className = '',
		showTitle = true,
		title = 'Upcoming Shifts',
		hideHeading = false,
		compactStyle = false
	}: {
		maxItems?: number;
		className?: string;
		showTitle?: boolean;
		title?: string;
		hideHeading?: boolean;
		compactStyle?: boolean;
	} = $props();

	// Query for upcoming shifts
	const upcomingShiftsQuery = $derived(createUpcomingShiftsQuery());
	
	const isLoading = $derived($upcomingShiftsQuery.isLoading);
	const isError = $derived($upcomingShiftsQuery.isError);
	const shifts = $derived($upcomingShiftsQuery.data?.slice(0, maxItems) || []);

	// Helper function to get relative time
	function getRelativeTime(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = date.getTime() - now.getTime();
		const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));
		
		if (diffDays === 0) return 'Today';
		if (diffDays === 1) return 'Tomorrow';
		if (diffDays <= 7) return `In ${diffDays} days`;
		return date.toLocaleDateString('en-ZA', { month: 'short', day: 'numeric' });
	}

	// Navigation function
	function navigateToShiftDetail(shift: any) {
		const shiftStartTime = encodeURIComponent(shift.start_time);
		goto(`/admin/shifts?shiftStartTime=${shiftStartTime}`);
	}
</script>

<div class="space-y-3 {className}">
	{#if showTitle && !hideHeading}
		<div class="flex items-center gap-2">
			<CalendarIcon class="h-4 w-4 text-muted-foreground" />
			<h3 class="font-medium text-sm">{title}</h3>
		</div>
	{/if}

	{#if isLoading}
		<!-- Loading skeletons -->
		<div class={compactStyle ? "space-y-0" : "space-y-2"}>
			{#each Array(3) as _}
				<div class={compactStyle ? "space-y-2 p-3 border-b" : "space-y-2 p-3 border rounded-lg"}>
					<Skeleton class="h-4 w-3/4" />
					<Skeleton class="h-3 w-1/2" />
					<Skeleton class="h-3 w-2/3" />
				</div>
			{/each}
		</div>
	{:else if isError}
		<!-- Error state -->
		<div class={compactStyle ? "text-sm text-muted-foreground p-3 border-b border-destructive/20 bg-destructive/5" : "text-sm text-muted-foreground p-3 border rounded-lg border-destructive/20 bg-destructive/5"}>
			Failed to load upcoming shifts
		</div>
	{:else if shifts.length === 0}
		<!-- Empty state -->
		<div class={compactStyle ? "text-sm text-muted-foreground p-3 border-b text-center" : "text-sm text-muted-foreground p-3 border rounded-lg text-center"}>
			No upcoming shifts in the next 2 weeks
		</div>
	{:else}
		<!-- Shifts list -->
		<div class={compactStyle ? "space-y-0" : "space-y-2"}>
			{#each shifts as shift (shift.schedule_id + shift.start_time)}
				<div 
					class={compactStyle 
						? "p-3 border-b hover:bg-muted/50 transition-all duration-200 cursor-pointer group"
						: "p-3 border rounded-lg hover:bg-muted/50 hover:border-muted-foreground/20 transition-all duration-200 cursor-pointer group"}
					onclick={() => navigateToShiftDetail(shift)}
					onkeydown={(e) => {
						if (e.key === 'Enter' || e.key === ' ') {
							e.preventDefault();
							navigateToShiftDetail(shift);
						}
					}}
					role="button"
					tabindex="0"
					aria-label={`View details for ${shift.schedule_name} shift on ${getRelativeTime(shift.start_time)}`}
				>
					<!-- Schedule name and time -->
					<div class="flex items-start justify-between gap-2 mb-2">
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2">
								<h4 class="font-medium text-sm truncate">{shift.schedule_name}</h4>
								<ExternalLinkIcon class="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity flex-shrink-0" />
							</div>
							<div class="flex items-center gap-1 text-xs text-muted-foreground mt-1">
								<ClockIcon class="h-3 w-3" />
								<span>{new Date(shift.start_time).toLocaleTimeString('en-ZA', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })} - {new Date(shift.end_time).toLocaleTimeString('en-ZA', { hour: '2-digit', minute: '2-digit', hour12: false, timeZone: 'UTC' })}</span>
							</div>
						</div>
						<Badge variant={shift.is_booked ? 'default' : 'secondary'} class="text-xs">
							{getRelativeTime(shift.start_time)}
						</Badge>
					</div>

					<!-- Booking status -->
					<div class="flex items-center justify-between gap-2">
						{#if shift.is_booked}
							<div class="flex items-center gap-1 text-xs text-green-600 min-w-0 flex-1">
								<UserCheckIcon class="h-3 w-3 flex-shrink-0" />
								<div class="min-w-0 flex-1">
									{#if shift.buddy_name}
										<!-- Two-person assignment - stack vertically -->
										<div class="space-y-0.5">
											<div class="truncate font-medium">{shift.user_name || 'Primary'}</div>
											<div class="truncate text-muted-foreground">+ {shift.buddy_name}</div>
										</div>
									{:else}
										<!-- Single assignment -->
										<span class="truncate">{shift.user_name || 'Booked'}</span>
									{/if}
								</div>
							</div>
						{:else}
							<div class="flex items-center gap-1 text-xs text-orange-600">
								<UserIcon class="h-3 w-3" />
								<span>Available</span>
							</div>
						{/if}
						
						<div class="flex-shrink-0">
							{#if shift.is_recurring_reservation}
								<Badge variant="outline" class="text-xs">Recurring</Badge>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>

		{#if shifts.length === maxItems && $upcomingShiftsQuery.data && $upcomingShiftsQuery.data.length > maxItems}
			<div class="text-xs text-muted-foreground text-center pt-2">
				Showing {maxItems} of {$upcomingShiftsQuery.data.length} upcoming shifts
			</div>
		{/if}
	{/if}
</div> 