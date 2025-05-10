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
	import { Button, buttonVariants } from '$lib/components/ui/button';
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

	// TEMP: Hardcode locale and timezone for DateFormatter to test
	const df = new DateFormatter('en-ZA', {
		dateStyle: 'medium',
		timeZone: 'UTC' // Using UTC for simplicity, can be any valid IANA timezone
	});

	// Internal state for the RangeCalendar, using DateValue objects
	let currentRange = $state<BitsDateRange | undefined>(undefined);
	let calendarPlaceholder = $state<DateValue>(today(getLocalTimeZone()));
	let popoverOpen = $state(false);

	// Effect to initialize or update currentRange when props change
	$effect(() => {
		const newPropStart = parseYyyyMmDdToCalendarDate(initialStartDate);
		const newPropEnd = parseYyyyMmDdToCalendarDate(initialEndDate);

		const internalStart = currentRange?.start;
		const internalEnd = currentRange?.end;

		let needsUpdate = false;

		// Check if start date needs update
		if (newPropStart && internalStart) {
			if (newPropStart.compare(internalStart) !== 0) needsUpdate = true;
		} else if (newPropStart || internalStart) {
			// One is defined, the other is not (XOR condition)
			if (newPropStart !== internalStart) needsUpdate = true;
		}

		// Check if end date needs update, only if start doesn't already necessitate an update
		if (!needsUpdate) {
			if (newPropEnd && internalEnd) {
				if (newPropEnd.compare(internalEnd) !== 0) needsUpdate = true;
			} else if (newPropEnd || internalEnd) {
				// One is defined, the other is not (XOR condition)
				if (newPropEnd !== internalEnd) needsUpdate = true;
			}
		}

		if (needsUpdate) {
			if (newPropStart && newPropEnd && newPropStart.compare(newPropEnd) <= 0) {
				currentRange = { start: newPropStart, end: newPropEnd };
			} else if (newPropStart) {
				currentRange = { start: newPropStart, end: undefined };
			} else {
				currentRange = undefined;
			}
		}

		// Update placeholder if no range is set (or specifically, if no start date is set)
		if (!currentRange?.start) {
			calendarPlaceholder = newPropStart || today(getLocalTimeZone());
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
			// Pass the hardcoded timezone to toDate() as well if it expects one,
			// though DateValue.toDate() usually converts to a JS Date in the system's local TZ by default.
			// For consistency with DateFormatter, specifying it if needed for toDate is safer.
			return `${df.format(currentRange.start.toDate('UTC'))} - ${df.format(currentRange.end.toDate('UTC'))}`;
		} else if (currentRange?.start) {
			return df.format(currentRange.start.toDate('UTC'));
		}
		return placeholderText;
	}
</script>

<Popover.Root bind:open={popoverOpen}>
	<Popover.Trigger
		class={cn(
			buttonVariants({ variant: 'outline' }),
			'w-full justify-start text-left font-normal',
			!currentRange?.start && 'text-muted-foreground'
		)}
	>
		<CalendarIcon class="mr-2 h-4 w-4" />
		<span>{getButtonLabel()}</span>
	</Popover.Trigger>
	<Popover.Content class="w-auto p-0" align="start">
		<RangeCalendar
			bind:value={currentRange}
			bind:placeholder={calendarPlaceholder}
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
