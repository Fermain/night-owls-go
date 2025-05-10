import { expect, test } from '@playwright/test';

test.describe('Admin Schedule Editing', () => {
	const scheduleListName = 'Test Schedule E2E';
	const scheduleCronExpr = '0 0 1 1 *'; // Yearly on Jan 1st at midnight
	const scheduleDuration = '30';
	const editedSuffix = ' - Edited by Playwright';

	test.beforeEach(async ({ page }) => {
		// Navigate to the admin schedules list
		await page.goto('/admin/schedules');

		// Check if the schedule we want to test with exists. If not, create it.
		const scheduleRowLocator = page.locator(`tr:has-text("${scheduleListName}")`);
		const count = await scheduleRowLocator.count();

		if (count === 0) {
			// Create the schedule if it doesn't exist
			await page.getByRole('link', { name: 'Create New Schedule' }).click();
			await expect(page).toHaveURL(/.*\/admin\/schedules\/new/);

			await page.locator('input#name').fill(scheduleListName);
			await page.locator('input#cron_expr').fill(scheduleCronExpr);
			await page.locator('input#duration_minutes').fill(scheduleDuration);
			
			await page.getByRole('button', { name: 'Create Schedule' }).click();
			await expect(page).toHaveURL(/.*\/admin\/schedules/); // Should redirect to list
			await expect(page.locator(`tr:has-text("${scheduleListName}")`)).toBeVisible();
		}
	});

	test('should allow editing an existing schedule name and not duplicate it', async ({ page }) => {
		await page.goto('/admin/schedules');
		
		// Get the total number of schedule rows before editing
		const initialScheduleRows = page.locator('table tbody tr');
		const initialRowCount = await initialScheduleRows.count();
		expect(initialRowCount).toBeGreaterThan(0); // Ensure there are schedules to edit

		// Find the row with the specific schedule name
		const scheduleRow = page.locator(`tr:has-text("${scheduleListName}")`).first();
		await expect(scheduleRow).toBeVisible();

		// Click the "Edit" button in that row
		// The Edit button is within the ScheduleActions component, rendered in the last cell
		await scheduleRow.locator('td:last-child a:has-text("Edit")').click();

		// Verify navigation to the edit page
		// URL should contain /edit and the schedule ID (which we don't know explicitly here, so regex)
		await expect(page).toHaveURL(/.*\/admin\/schedules\/\d+\/edit/);
		
		// The name input should be pre-filled with the current name
		const nameInput = page.locator('input#name');
		await expect(nameInput).toHaveValue(scheduleListName);

		// Modify the name
		const editedScheduleName = scheduleListName + editedSuffix;
		await nameInput.fill(editedScheduleName);

		// Save the changes
		// Using a more robust selector for the submit button
		await page.locator('form button[type="submit"]').click();

		// Verify redirection back to the schedules list
		await expect(page).toHaveURL(/.*\/admin\/schedules/);

		// Verify the schedule with the new name is visible
		await expect(page.locator(`tr:has-text("${editedScheduleName}")`)).toBeVisible();

		// Verify the schedule with the old name is NOT visible (unless it was a partial edit of a different schedule)
		await expect(page.locator(`tr:has-text("${scheduleListName}")`)).not.toBeVisible();
		
		// Verify the total number of schedules has not increased
		const finalScheduleRows = page.locator('table tbody tr');
		const finalRowCount = await finalScheduleRows.count();
		expect(finalRowCount).toEqual(initialRowCount);
	});

	// Clean up: delete the schedule after tests if needed (optional)
	// test.afterAll(async ({ page }) => {
		// This would require a delete button and functionality
	// });
}); 