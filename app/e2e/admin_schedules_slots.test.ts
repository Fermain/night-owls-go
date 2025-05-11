import { test, expect } from '@playwright/test';

test.describe('Admin Schedules Slots Page', () => {
	test('should allow selecting and re-selecting a date range in DateRangePicker', async ({
		page
	}) => {
		await page.goto('/admin/schedules/slots');

		// The DateRangePicker trigger button
		const datePickerButton = page.locator(
			'button[data-popover-trigger][aria-haspopup="dialog"]:has(svg.lucide-calendar)'
		);

		// --- First selection ---
		await expect(datePickerButton).toBeVisible({ timeout: 5000 });
		await datePickerButton.click();

		// Wait for the calendar popover to be visible using its root attribute
		const calendarRoot = page.locator('div[data-range-calendar-root]');
		await expect(calendarRoot).toBeVisible({ timeout: 3000 });

		// Select the 1st day of the current month that is not an 'outside' day
		const firstDayOfMonth = page
			.locator('div[data-range-calendar-day]:not([data-outside-month])')
			.filter({ hasText: '1' })
			.first();
		const tenthDayOfMonth = page
			.locator('div[data-range-calendar-day]:not([data-outside-month])')
			.filter({ hasText: '10' })
			.first();

		await expect(firstDayOfMonth).toBeVisible({ timeout: 3000 }); // Ensure day is visible before click
		await firstDayOfMonth.click();
		await expect(tenthDayOfMonth).toBeVisible({ timeout: 3000 }); // Ensure day is visible before click
		await tenthDayOfMonth.click(); // Popover should close after this

		// Verify the button text updated.
		await expect(datePickerButton).not.toHaveText('Select date range for slots', { timeout: 3000 });
		const firstSelectedRangeText = await datePickerButton.textContent();
		console.log('First selected range text:', firstSelectedRangeText);
		expect(firstSelectedRangeText).not.toBeNull();
		expect(firstSelectedRangeText?.trim()).not.toBe('');

		// --- Second selection (testing re-selection) ---
		await datePickerButton.click(); // Re-open the picker
		await expect(calendarRoot).toBeVisible({ timeout: 3000 }); // Wait for calendar again

		// Pick 15th and 20th
		const fifteenthDayOfMonth = page
			.locator('div[data-range-calendar-day]:not([data-outside-month])')
			.filter({ hasText: '15' })
			.first();
		const twentiethDayOfMonth = page
			.locator('div[data-range-calendar-day]:not([data-outside-month])')
			.filter({ hasText: '20' })
			.first();

		await expect(fifteenthDayOfMonth).toBeVisible({ timeout: 3000 });
		await fifteenthDayOfMonth.click();
		await expect(twentiethDayOfMonth).toBeVisible({ timeout: 3000 });
		await twentiethDayOfMonth.click(); // Popover should close

		// Verify button text updated to the new range
		await expect(datePickerButton).not.toHaveText(firstSelectedRangeText!, { timeout: 3000 });
		const secondSelectedRangeText = await datePickerButton.textContent();
		console.log('Second selected range text:', secondSelectedRangeText);
		await expect(datePickerButton).not.toHaveText('Select date range for slots', { timeout: 3000 });
		expect(secondSelectedRangeText).not.toBeNull();
		expect(secondSelectedRangeText?.trim()).not.toBe('');
	});
});
