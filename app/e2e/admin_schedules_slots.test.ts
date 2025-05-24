import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

test.describe('Admin Schedules Slots Page', () => {
	test('should allow selecting and re-selecting a date range in DateRangePicker', async ({
		page
	}) => {
		await loginAsAdmin(page);
		await page.goto('/admin/schedules/slots');

		// Wait for page to load
		await page.waitForLoadState('networkidle');

		// Debug: Print all buttons on the page
		const allButtons = await page.locator('button').all();
		console.log(`Found ${allButtons.length} buttons on the page`);
		
		// Look for date/calendar related buttons
		for (let i = 0; i < allButtons.length; i++) {
			const text = await allButtons[i].textContent();
			const hasCalendarIcon = await allButtons[i].locator('svg').count() > 0;
			if (text?.includes('Select') || text?.includes('range') || text?.includes('Date') || hasCalendarIcon) {
				console.log(`Button ${i + 1}: "${text}" (hasCalendarIcon: ${hasCalendarIcon})`);
			}
		}

		// Look specifically for DateRangePicker in the sidebar filters section
		const filtersSection = page.locator('.space-y-3, [data-testid="filters"]').first();
		await expect(filtersSection).toBeVisible({ timeout: 5000 });
		
		// Look for the date picker within the filters section
		const datePickerButton = filtersSection.locator('button:has(svg)').first();
		
		console.log('Attempting to click date picker button in filters section');

		// --- First selection ---
		await expect(datePickerButton).toBeVisible({ timeout: 10000 });
		await datePickerButton.click();

		// Wait for the calendar popover to be visible
		const calendarRoot = page.locator('[data-range-calendar-root], .range-calendar, [role="dialog"]').first();
		await expect(calendarRoot).toBeVisible({ timeout: 3000 });

		// Select the 1st day of the current month that is not an 'outside' day
		const firstDayOfMonth = page
			.locator('[data-range-calendar-day]:not([data-outside-month]), .calendar-day:not(.outside), [role="gridcell"]')
			.filter({ hasText: '1' })
			.first();
		const tenthDayOfMonth = page
			.locator('[data-range-calendar-day]:not([data-outside-month]), .calendar-day:not(.outside), [role="gridcell"]')
			.filter({ hasText: '10' })
			.first();

		await expect(firstDayOfMonth).toBeVisible({ timeout: 3000 }); // Ensure day is visible before click
		await firstDayOfMonth.click();
		await expect(tenthDayOfMonth).toBeVisible({ timeout: 3000 }); // Ensure day is visible before click
		await tenthDayOfMonth.click(); // Popover should close after this

		// Verify the button text updated.
		const firstSelectedRangeText = await datePickerButton.textContent();
		console.log('First selected range text:', firstSelectedRangeText);
		expect(firstSelectedRangeText).not.toBeNull();
		expect(firstSelectedRangeText?.trim()).not.toBe('');

		// --- Second selection (testing re-selection) ---
		await datePickerButton.click(); // Re-open the picker
		await expect(calendarRoot).toBeVisible({ timeout: 3000 }); // Wait for calendar again

		// Pick 15th and 20th
		const fifteenthDayOfMonth = page
			.locator('[data-range-calendar-day]:not([data-outside-month]), .calendar-day:not(.outside), [role="gridcell"]')
			.filter({ hasText: '15' })
			.first();
		const twentiethDayOfMonth = page
			.locator('[data-range-calendar-day]:not([data-outside-month]), .calendar-day:not(.outside), [role="gridcell"]')
			.filter({ hasText: '20' })
			.first();

		await expect(fifteenthDayOfMonth).toBeVisible({ timeout: 3000 });
		await fifteenthDayOfMonth.click();
		await expect(twentiethDayOfMonth).toBeVisible({ timeout: 3000 });
		await twentiethDayOfMonth.click(); // Popover should close

		// Verify button text updated to the new range
		const secondSelectedRangeText = await datePickerButton.textContent();
		console.log('Second selected range text:', secondSelectedRangeText);
		expect(secondSelectedRangeText).not.toBeNull();
		expect(secondSelectedRangeText?.trim()).not.toBe('');
		expect(secondSelectedRangeText?.trim()).not.toBe(firstSelectedRangeText?.trim());
	});
});
