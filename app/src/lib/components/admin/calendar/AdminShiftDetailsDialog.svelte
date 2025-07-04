<script lang="ts">
	import { formatTimeSlot } from '$lib/utils/dateFormatting';
	import { getRelativeTime } from '$lib/utils/shifts';
	import type { AdminShiftSlot } from '$lib/types';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Dialog, DialogContent, DialogHeader, DialogTitle } from '$lib/components/ui/dialog';
	import { Button } from '$lib/components/ui/button';
	import { UserIcon, ClockIcon, CalendarIcon, PhoneIcon } from 'lucide-svelte';

	let {
		shift,
		open = $bindable(),
		onReassignClick
	}: {
		shift: AdminShiftSlot | null;
		open: boolean;
		onReassignClick?: (shift: AdminShiftSlot) => void;
	} = $props();

	const isShiftFilled = $derived(shift?.is_booked && shift?.user_name);

	// Comprehensive date/time formatting
	const shiftDate = $derived.by(() => {
		if (!shift) return null;
		const date = new Date(shift.start_time);
		return {
			// Full date: "Wednesday, January 15, 2025"
			fullDate: date.toLocaleDateString('en-US', {
				weekday: 'long',
				year: 'numeric',
				month: 'long',
				day: 'numeric'
			}),
			// Time range: "18:00 - 06:00"
			timeRange: formatTimeSlot(shift.start_time, shift.end_time),
			// Relative: "in 2 days", "tomorrow", "in 3 hours"
			relative: getRelativeTime(shift.start_time)
		};
	});
</script>

<Dialog bind:open>
	<DialogContent class="max-w-md">
		{#if shift && shiftDate}
			<DialogHeader>
				<DialogTitle class="flex items-center gap-2">
					<CalendarIcon class="h-5 w-5" />
					Shift Details
				</DialogTitle>
			</DialogHeader>

			<div class="space-y-4">
				<!-- Date & Time Information -->
				<Card>
					<CardContent class="pt-6 space-y-4">
						<!-- Full date display -->
						<div class="text-center border-b pb-4">
							<div class="text-lg font-semibold">{shiftDate.fullDate}</div>
							<div class="text-muted-foreground text-sm">{shiftDate.relative}</div>
						</div>

						<!-- Shift details -->
						<div class="space-y-3">
							<div class="flex items-center gap-2 text-sm">
								<ClockIcon class="h-4 w-4 text-muted-foreground" />
								<span class="font-medium">{shiftDate.timeRange}</span>
							</div>
							<div class="flex items-center gap-2 text-sm">
								<CalendarIcon class="h-4 w-4 text-muted-foreground" />
								<span>{shift.schedule_name}</span>
							</div>
							<div class="flex items-center gap-2">
								<Badge
									variant={isShiftFilled ? 'default' : 'secondary'}
									class={isShiftFilled
										? 'bg-green-100 text-green-700 border-green-200'
										: 'bg-orange-100 text-orange-700 border-orange-200'}
								>
									{isShiftFilled ? 'Assigned' : 'Available'}
								</Badge>
								{#if shift.is_recurring_reservation}
									<Badge variant="outline" class="text-xs">Recurring</Badge>
								{/if}
							</div>
						</div>
					</CardContent>
				</Card>

				<!-- Assignment Info (if filled) -->
				{#if isShiftFilled}
					<Card>
						<CardHeader class="pb-3">
							<CardTitle class="text-sm flex items-center gap-2">
								<UserIcon class="h-4 w-4" />
								Current Assignment
							</CardTitle>
						</CardHeader>
						<CardContent class="space-y-3">
							<div>
								<p class="font-medium">{shift.user_name}</p>
								{#if shift.user_phone}
									<div class="flex items-center gap-2 text-sm text-muted-foreground mt-1">
										<PhoneIcon class="h-3 w-3" />
										<span>{shift.user_phone}</span>
									</div>
								{/if}
							</div>

							{#if shift.buddy_name}
								<div class="pt-2 border-t">
									<p class="text-sm text-muted-foreground">Buddy</p>
									<p class="font-medium">{shift.buddy_name}</p>
								</div>
							{/if}

							<!-- Action Buttons -->
							<div class="flex gap-2 pt-2">
								<Button
									variant="outline"
									size="sm"
									onclick={() => {
										open = false;
										onReassignClick?.(shift);
									}}
								>
									Reassign
								</Button>
							</div>
						</CardContent>
					</Card>
				{/if}
			</div>
		{/if}
	</DialogContent>
</Dialog>
