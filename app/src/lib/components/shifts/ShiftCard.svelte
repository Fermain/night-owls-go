<script lang="ts">
	import { Badge } from '$lib/components/ui/badge';
	import { Button } from '$lib/components/ui/button';
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import ClockIcon from '@lucide/svelte/icons/clock';
	import UsersIcon from '@lucide/svelte/icons/users';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import type { ProcessedShift } from '$lib/types/shifts';

	export let shift: ProcessedShift;
	export let isAuthenticated: boolean = false;
	export let onBookShift: (shift: ProcessedShift) => void = () => {};
	export let onSignIn: () => void = () => {};

	function formatShiftTime(startTime: string, endTime: string) {
		const start = new Date(startTime);
		const end = new Date(endTime);
		return `${start.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })} - ${end.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })}`;
	}

	function formatShiftTitle(startTime: string, endTime: string): string {
		const start = new Date(startTime);
		const end = new Date(endTime);
		
		const previousDay = new Date(start);
		previousDay.setDate(previousDay.getDate() - 1);
		
		const dayName = previousDay.toLocaleDateString('en-US', { weekday: 'long', timeZone: 'UTC' });
		
		const startHour = start.getUTCHours();
		const endHour = end.getUTCHours();
		
		const formatHour = (hour: number) => hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
		const getAmPm = (hour: number) => hour < 12 ? 'AM' : 'PM';
		
		const startHour12 = formatHour(startHour);
		const endHour12 = formatHour(endHour);
		const endAmPm = getAmPm(endHour);
		
		const timeRange = `${startHour12}-${endHour12}${endAmPm}`;
		
		return `${dayName} Night ${timeRange}`;
	}

	function formatShiftDate(startTime: string) {
		const shiftDate = new Date(startTime);
		const today = new Date();
		const tomorrow = new Date(today);
		tomorrow.setDate(today.getDate() + 1);

		const displayDate = new Date(shiftDate);
		displayDate.setDate(displayDate.getDate() - 1);

		const displayDateString = displayDate.toDateString();
		const todayString = today.toDateString();
		const tomorrowString = tomorrow.toDateString();

		if (displayDateString === todayString) {
			return 'Tonight';
		} else if (displayDateString === tomorrowString) {
			return 'Tomorrow Night';
		} else {
			return displayDate.toLocaleDateString('en-GB', {
				weekday: 'short',
				month: 'short',
				day: 'numeric'
			});
		}
	}

	function getShiftDuration(startTime: string, endTime: string) {
		const start = new Date(startTime);
		const end = new Date(endTime);
		const diffHours = (end.getTime() - start.getTime()) / (1000 * 60 * 60);
		return `${diffHours}h`;
	}

	function getStatusText(shift: ProcessedShift) {
		if (shift.is_tonight && !shift.is_booked) return 'Tonight';
		if (shift.is_booked) return 'Urgent';
		return 'Available';
	}
</script>

<div class="p-4 {shift.priority === 'high' ? 'bg-muted/30' : 'bg-background'}">
	<div class="flex items-start justify-between">
		<div class="flex-1 space-y-3">
			<div class="flex items-center gap-3">
				<div class="border border-border p-2 rounded">
					<ShieldIcon class="h-4 w-4 text-foreground" />
				</div>
				<div>
					<h3 class="font-medium text-foreground">
						{formatShiftTitle(shift.start_time, shift.end_time)}
					</h3>
					<div class="flex items-center gap-2 mt-1">
						<Badge variant="outline" class="text-xs h-4 px-1">
							{shift.schedule_name}
						</Badge>
						<Badge variant={shift.priority === 'high' ? 'secondary' : 'outline'} class="text-xs h-4 px-1">
							{getStatusText(shift)}
						</Badge>
						{#if shift.priority === 'high'}
							<Badge variant="secondary" class="text-xs h-4 px-1">Urgent</Badge>
						{/if}
					</div>
				</div>
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-2 text-sm">
				<div class="flex items-center gap-2 text-muted-foreground">
					<CalendarIcon class="h-4 w-4 flex-shrink-0" />
					<span class="text-foreground">{formatShiftDate(shift.start_time)}</span>
				</div>

				<div class="flex items-center gap-2 text-muted-foreground">
					<ClockIcon class="h-4 w-4 flex-shrink-0" />
					<span class="text-foreground">{formatShiftTime(shift.start_time, shift.end_time)}</span>
					<Badge variant="outline" class="text-xs h-4 px-1 ml-1">
						{getShiftDuration(shift.start_time, shift.end_time)}
					</Badge>
				</div>
			</div>
		</div>

		<div class="text-right space-y-2 flex-shrink-0 ml-4">
			<div class="flex items-center gap-1 text-sm text-muted-foreground justify-end">
				<UsersIcon class="h-4 w-4" />
				<span class="text-foreground">{shift.slots_available}/{shift.total_slots}</span>
			</div>

			{#if !isAuthenticated}
				<Button size="sm" variant="outline" onclick={onSignIn} class="w-full min-w-[80px] h-8">
					Sign in
				</Button>
			{:else if !shift.is_booked}
				<Button
					size="sm"
					class="w-full min-w-[80px] h-8"
					onclick={() => onBookShift(shift)}
				>
					Join
				</Button>
			{:else}
				<Button size="sm" variant="outline" disabled class="w-full min-w-[80px] h-8">
					Full
				</Button>
			{/if}
		</div>
	</div>

	{#if shift.slots_available === 1 && shift.total_slots > 1}
		<div class="mt-3 p-2 border border-border rounded bg-muted/30">
			<div class="flex items-center gap-2 text-sm text-foreground">
				⚠️ Only 1 patrol slot remaining - join now!
			</div>
		</div>
	{/if}
</div> 