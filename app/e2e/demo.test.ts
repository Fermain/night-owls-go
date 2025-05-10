import { expect, test } from '@playwright/test';

test('API reachability test page successfully fetches schedules', async ({ page }) => {
	// Navigate to the root page where the API reachability test is
	await page.goto('/');

	// Check if the main heading for the test is present
	await expect(page.locator('h1', { hasText: 'API Reachability Test' })).toBeVisible();

	// Wait for the status message to indicate success
	// We'll look for the text "Successfully fetched" which appears on success.
	// Increased timeout because the API call might take a moment, especially on first run or CI.
	const statusLocator = page.locator('p', { hasText: /Status:/ });
	await expect(statusLocator).toContainText(/Successfully fetched \d+ schedule\(s\)\./, {
		timeout: 10000
	});

	// Optionally, check if the <pre> tag with schedule data is present and contains some expected text
	// This confirms not just that the fetch occurred but that data was rendered.
	const dataLocator = page.locator('pre');
	await expect(dataLocator).toBeVisible();
	// Example check: if schedules are seeded, one might be "Morning Patrol"
	// For a generic check, we can see if it says "First schedule name:" or "No schedules returned."
	await expect(dataLocator).toContainText(/(First schedule name:|No schedules returned.)/);
});
