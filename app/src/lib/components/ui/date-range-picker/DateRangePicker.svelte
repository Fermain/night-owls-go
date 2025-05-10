<script lang="ts">
	import CalendarIcon from '@lucide/svelte/icons/calendar';
	import type { DateRange as BitsDateRange } from 'bits-ui'; // Underlying type from bits-ui for RangeCalendar
	import {
		DateFormatter,
		type DateValue,
		getLocalTimeZone,
		today,
		CalendarDate
	} from '@internationalized/date';
	import { cn } from '$lib/utils';
	import { Button } from '$lib/components/ui/button';
	import { RangeCalendar } from '$lib/components/ui/range-calendar';
	import * as Popover from '$lib/components/ui/popover';
	import { parseYyyyMmDdToCalendarDate, formatCalendarDateToYyyyMmDd } from '$lib/utils/date';
	import { createEventDispatcher } from 'svelte';

	type Props = {
		initialStartDate?: string | null; // YYYY-MM-DD
		initialEndDate?: string | null; // YYYY-MM-DD
		placeholderText?: string;
	};

	const {
		initialStartDate = undefined,
		initialEndDate = undefined,
		placeholderText = 'Select a date range'
	}: Props = $props();

	const dispatch = createEventDispatcher<{
		change: { start: string | null; end: string | null };
	}>();

	const df = new DateFormatter(getLocalTimeZone(), {
		dateStyle: 'medium'
	});

	// Internal state for the RangeCalendar, using DateValue objects
	let currentRange = $state<BitsDateRange | undefined>(undefined);
	let calendarPlaceholder = $state<DateValue>(today(getLocalTimeZone()));
	let popoverOpen = $state(false);

	// Effect to initialize or update currentRange when props change
	$effect(() => {
		const startVal = parseYyyyMmDdToCalendarDate(initialStartDate);
		const endVal = parseYyyyMmDdToCalendarDate(initialEndDate);
		const currentStartString = formatCalendarDateToYyyyMmDd(currentRange?.start);
		const currentEndString = formatCalendarDateToYyyyMmDd(currentRange?.end);

		// Only update from props if the string versions are different
		// to avoid resetting user interaction unnecessarily or causing loops.
		if (initialStartDate !== currentStartString || initialEndDate !== currentEndString) {
			if (startVal && endVal && startVal.compare(endVal) <= 0) {
				currentRange = { start: startVal, end: endVal };
			} else if (startVal) {
				currentRange = { start: startVal, end: undefined };
			} else {
				currentRange = undefined;
			}
		}
		// Update placeholder if no range is set
		if (!currentRange?.start) {
			calendarPlaceholder = startVal || today(getLocalTimeZone());
		}
	});

	// Effect to dispatch change when currentRange is updated by the calendar
	// This needs to be distinct from the prop-driven update to avoid loops
	// and to ensure we only dispatch when the user interacts with the calendar.
	$effect(() => {
		const newStartString = formatCalendarDateToYyyyMmDd(currentRange?.start);
		const newEndString = formatCalendarDateToYyyyMmDd(currentRange?.end);

		// Check if the actual selected values (now strings) differ from initial props
		// to decide if a meaningful change occurred that needs dispatching.
		// This helps avoid dispatching events when the component initializes or props are set externally
		// without actual user interaction leading to a new distinct range.
		if (newStartString !== initialStartDate || newEndString !== initialEndDate) {
			dispatch('change', {
				start: newStartString,
				end: newEndString
			});
		}
	});

	function getButtonLabel(): string {
		if (currentRange?.start && currentRange?.end) {
			return `${df.format(currentRange.start.toDate(getLocalTimeZone()))} - ${df.format(currentRange.end.toDate(getLocalTimeZone()))}`;
		} else if (currentRange?.start) {
			// If only a start date is picked in the range (end is undefined)
			return df.format(currentRange.start.toDate(getLocalTimeZone()));
		}
		return placeholderText;
	}
</script>

<Popover.Root bind:open={popoverOpen}>
	<Popover.Trigger asChild={true}>
		<Button
			variant="outline"
			class={cn(
				'w-[300px] justify-start text-left font-normal',
				!currentRange?.start && 'text-muted-foreground' // Show muted if no start date
			)}
		>
			<CalendarIcon class="mr-2 h-4 w-4" />
			<span>{getButtonLabel()}</span>
		</Button>
	</Popover.Trigger>
	<Popover.Content class="w-auto p-0" align="start">
		<RangeCalendar
			bind:value={currentRange}
			bind:placeholder={calendarPlaceholder}
			initialFocus={true}
			numberOfMonths={2}
			minValue={new CalendarDate(1900, 1, 1)}
			onValueChange={(v) => {
				if (v?.start && v.end && v.start.compare(v.end) <= 0) {
					popoverOpen = false;
				}
			}}
		/>
	</Popover.Content>
</Popover.Root>
