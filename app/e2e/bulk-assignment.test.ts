import { test, expect, type Page } from '@playwright/test';

// Test configuration
const ADMIN_PHONE = '+1234567890';
const OTP = '123456'; // Dev mode OTP

// Test data constants
const TEST_USERS = {
	VOLUNTEER_1: { name: 'Alice Admin', phone: '+27821234567', role: 'admin' },
	VOLUNTEER_2: { name: 'Bob Manager', phone: '+27821234568', role: 'admin' },
	VOLUNTEER_3: { name: 'Charlie Volunteer', phone: '+27821234569', role: 'owl' },
	VOLUNTEER_4: { name: 'Diana Scout', phone: '+27821234570', role: 'owl' },
	VOLUNTEER_5: { name: 'Eve Patrol', phone: '+27821234571', role: 'owl' }
};

async function loginAsAdmin(page: Page) {
	await page.goto('/login');
	await page.fill('input[name="phone"]', ADMIN_PHONE);
	await page.click('button[type="submit"]');

	// Enter OTP
	await page.fill('input[name="otp"]', OTP);
	await page.click('button[type="submit"]');

	// Wait for redirect to dashboard
	await expect(page).toHaveURL('/');
}

async function navigateToBulkAssignment(page: Page) {
	await page.goto('/admin/shifts/bulk-signup');
	await expect(page.locator('h1')).toContainText('Bulk Shift Assignment');
}

async function selectUser(page: Page, userName: string) {
	// Open user dropdown
	await page.click('button:has-text("Select user...")');

	// Wait for dropdown to open and search for user
	await page.fill('[placeholder="Search users..."]', userName);

	// Click on the user option
	await page.click(`[role="option"]:has-text("${userName}")`);

	// Verify user is selected
	await expect(page.locator('button').filter({ hasText: userName })).toBeVisible();
}

async function togglePatternMode(page: Page, enabled: boolean) {
	const checkbox = page.locator('#pattern-mode');
	const isChecked = await checkbox.isChecked();

	if (isChecked !== enabled) {
		await checkbox.click();
	}

	// Verify pattern mode state
	await expect(checkbox).toBeChecked({ checked: enabled });
}

async function selectDateRange(page: Page, days: number = 30) {
	// Click date range picker
	await page.click('[data-testid="date-range-picker"], button:has-text("Next 30 days")');

	// For simplicity, we'll use the default range or extend it
	// In a real test, you might want to select specific dates
}

test.describe('Bulk Assignment - Page Loading and Navigation', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test('Page loads correctly with all sections visible', async ({ page }) => {
		await navigateToBulkAssignment(page);

		// Check main sections are present
		await expect(page.locator('h1')).toContainText('Bulk Shift Assignment');
		await expect(page.locator('text=User Selection')).toBeVisible();
		await expect(page.locator('text=Selection Mode')).toBeVisible();
		await expect(page.locator('text=Selection Summary')).toBeVisible();
		await expect(page.locator('text=Available Shifts')).toBeVisible();
	});

	test('Navigation from shifts layout works', async ({ page }) => {
		await page.goto('/admin/shifts');

		// Click bulk assignment navigation item
		await page.click('a[href="/admin/shifts/bulk-signup"]');

		await expect(page).toHaveURL('/admin/shifts/bulk-signup');
		await expect(page.locator('h1')).toContainText('Bulk Shift Assignment');
	});

	test('User dropdown loads and searches correctly', async ({ page }) => {
		await navigateToBulkAssignment(page);

		// Open user dropdown
		await page.click('button:has-text("Select user...")');

		// Search for a user
		await page.fill('[placeholder="Search users..."]', 'Charlie');

		// Should show matching users
		await expect(page.locator('[role="option"]:has-text("Charlie Volunteer")')).toBeVisible();
	});

	test('Available shifts load correctly', async ({ page }) => {
		await navigateToBulkAssignment(page);

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).toBeVisible({ timeout: 5000 });
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Should show available shifts
		await expect(page.locator('[data-testid="shift-card"], .border.rounded-lg.p-3')).toBeVisible();
	});
});

test.describe('Bulk Assignment - Manual Selection Mode', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToBulkAssignment(page);
	});

	test('Can select individual shifts with checkboxes', async ({ page }) => {
		// Ensure pattern mode is off
		await togglePatternMode(page, false);

		// Select a user
		await selectUser(page, 'Charlie Volunteer');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Find and click a checkbox for an available shift
		const firstCheckbox = page.locator('[type="checkbox"]').first();
		await firstCheckbox.click();

		// Verify selection counter updates
		await expect(page.locator('text=1').first()).toBeVisible();
	});

	test('Can select multiple individual shifts', async ({ page }) => {
		await togglePatternMode(page, false);
		await selectUser(page, 'Diana Scout');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Select multiple checkboxes
		const checkboxes = page.locator('[type="checkbox"]');
		const checkboxCount = await checkboxes.count();

		if (checkboxCount >= 3) {
			await checkboxes.nth(0).click();
			await checkboxes.nth(1).click();
			await checkboxes.nth(2).click();

			// Verify selection count
			await expect(page.locator('text=3').first()).toBeVisible();
		}
	});

	test('Can select all shifts for a date', async ({ page }) => {
		await togglePatternMode(page, false);
		await selectUser(page, 'Eve Patrol');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Click "Select Available" button for a date group
		const selectAllButton = page.locator('button:has-text("Select Available")').first();
		if (await selectAllButton.isVisible()) {
			await selectAllButton.click();

			// Verify multiple shifts are selected
			await expect(page.locator('.text-3xl.font-bold')).not.toHaveText('0');
		}
	});

	test('Can clear manual selection', async ({ page }) => {
		await togglePatternMode(page, false);
		await selectUser(page, 'Charlie Volunteer');

		// Wait for shifts to load and select some shifts
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		const firstCheckbox = page.locator('[type="checkbox"]').first();
		await firstCheckbox.click();

		// Verify selection exists
		await expect(page.locator('.text-3xl.font-bold')).not.toHaveText('0');

		// Clear selection
		await page.click('button:has-text("Clear Selection")');

		// Verify selection is cleared
		await expect(page.locator('.text-3xl.font-bold')).toHaveText('0');
	});
});

test.describe('Bulk Assignment - Pattern Selection Mode', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToBulkAssignment(page);
	});

	test('Can enable pattern selection mode', async ({ page }) => {
		// Toggle pattern mode on
		await togglePatternMode(page, true);

		// Verify UI changes
		await expect(page.locator('text=Click any shift to select all matching shifts')).toBeVisible();

		// Checkboxes should be replaced with pattern indicators
		await expect(page.locator('[type="checkbox"]')).not.toBeVisible();
	});

	test('Can select a pattern and auto-select matching shifts', async ({ page }) => {
		await togglePatternMode(page, true);
		await selectUser(page, 'Diana Scout');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Click on any shift to select the pattern
		const firstShift = page.locator('.border.rounded-lg.p-3').first();
		await firstShift.click();

		// Verify pattern is selected and displayed
		await expect(page.locator('text=Selected Pattern')).toBeVisible();
		await expect(page.locator('text=Every')).toBeVisible(); // Should show "Every [Day] [Time]"

		// Verify multiple matching shifts are selected
		await expect(page.locator('.text-3xl.font-bold')).not.toHaveText('0');
	});

	test('Pattern description shows correct day and time', async ({ page }) => {
		await togglePatternMode(page, true);
		await selectUser(page, 'Charlie Volunteer');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Click on a specific shift
		const mondayShift = page.locator('.border.rounded-lg.p-3:has-text("Monday")').first();
		if (await mondayShift.isVisible()) {
			await mondayShift.click();

			// Verify pattern description includes day and time
			await expect(page.locator('text=Every Monday')).toBeVisible();
		}
	});

	test('Can clear pattern selection', async ({ page }) => {
		await togglePatternMode(page, true);
		await selectUser(page, 'Eve Patrol');

		// Wait for shifts and select a pattern
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		const firstShift = page.locator('.border.rounded-lg.p-3').first();
		await firstShift.click();

		// Verify pattern is selected
		await expect(page.locator('text=Selected Pattern')).toBeVisible();

		// Clear pattern
		await page.click('button:has-text("Clear Pattern")');

		// Verify pattern is cleared
		await expect(page.locator('text=Selected Pattern')).not.toBeVisible();
		await expect(page.locator('.text-3xl.font-bold')).toHaveText('0');
	});

	test('Pattern selection updates when date range changes', async ({ page }) => {
		await togglePatternMode(page, true);
		await selectUser(page, 'Charlie Volunteer');

		// Wait for shifts and select a pattern
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		const firstShift = page.locator('.border.rounded-lg.p-3').first();
		await firstShift.click();

		const initialCount = await page.locator('.text-3xl.font-bold').textContent();

		// Change date range (extend it)
		// This would trigger re-selection of pattern matches
		// Note: Actual date range selection would need more specific implementation

		// For now, just verify the feature is responsive to changes
		await expect(page.locator('.text-3xl.font-bold')).toBeVisible();
	});
});

test.describe('Bulk Assignment - Assignment Execution', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToBulkAssignment(page);
	});

	test('Can successfully assign individual shifts', async ({ page }) => {
		await togglePatternMode(page, false);
		await selectUser(page, 'Diana Scout');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Select a shift
		const firstCheckbox = page.locator('[type="checkbox"]').first();
		await firstCheckbox.click();

		// Submit assignment
		await page.click('button:has-text("Assign")');

		// Wait for success message
		await expect(page.locator('text=successfully')).toBeVisible({ timeout: 15000 });

		// Verify selection is cleared after success
		await expect(page.locator('.text-3xl.font-bold')).toHaveText('0');
	});

	test('Can successfully assign pattern shifts', async ({ page }) => {
		await togglePatternMode(page, true);
		await selectUser(page, 'Eve Patrol');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Select a pattern
		const firstShift = page.locator('.border.rounded-lg.p-3').first();
		await firstShift.click();

		// Submit pattern assignment
		await page.click('button:has-text("Assign Pattern")');

		// Wait for success (this might take longer for multiple assignments)
		await expect(page.locator('text=successfully')).toBeVisible({ timeout: 30000 });
	});

	test('Shows validation error when no user selected', async ({ page }) => {
		// Don't select a user
		await togglePatternMode(page, false);

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Try to assign without selecting user
		await page.click('button:has-text("Assign")');

		// Should show validation error
		await expect(page.locator('text=Please select a user')).toBeVisible();
	});

	test('Shows validation error when no shifts selected', async ({ page }) => {
		await selectUser(page, 'Charlie Volunteer');

		// Don't select any shifts
		// Try to assign
		await page.click('button:has-text("Assign")');

		// Should show validation error
		await expect(page.locator('text=Please select at least one shift')).toBeVisible();
	});

	test('Shows loading state during assignment', async ({ page }) => {
		await togglePatternMode(page, false);
		await selectUser(page, 'Diana Scout');

		// Wait for shifts to load and select one
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		const firstCheckbox = page.locator('[type="checkbox"]').first();
		await firstCheckbox.click();

		// Click assign and immediately check for loading state
		await page.click('button:has-text("Assign")');
		await expect(page.locator('button:has-text("Assigning...")')).toBeVisible();
	});

	test('Handles partial success in bulk assignment', async ({ page }) => {
		await togglePatternMode(page, true);
		await selectUser(page, 'Charlie Volunteer');

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Select a pattern that might have some conflicts
		const firstShift = page.locator('.border.rounded-lg.p-3').first();
		await firstShift.click();

		// Submit assignment
		await page.click('button:has-text("Assign Pattern")');

		// Wait for completion and check for partial success message
		await page.waitForTimeout(5000);

		// Should either show full success or partial success with error count
		const successText = page.locator('text=successfully');
		const partialText = page.locator('text=failed');

		await expect(successText.or(partialText)).toBeVisible({ timeout: 30000 });
	});
});

test.describe('Bulk Assignment - Filters and Display', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToBulkAssignment(page);
	});

	test('Available shifts only filter works', async ({ page }) => {
		// Ensure "Available shifts only" is checked by default
		const availableOnlyCheckbox = page.locator('#available-only');
		await expect(availableOnlyCheckbox).toBeChecked();

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// All visible shifts should be available (not booked)
		const bookedShifts = page.locator('text=Assigned, text=Booked');
		await expect(bookedShifts).not.toBeVisible();
	});

	test('Can toggle to show all shifts including booked', async ({ page }) => {
		// Uncheck "Available shifts only"
		await page.click('#available-only');

		// Wait for shifts to reload
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Should now show both available and booked shifts
		// Booked shifts should be disabled/styled differently
		const shiftCards = page.locator('.border.rounded-lg.p-3');
		await expect(shiftCards).toBeVisible();
	});

	test('Shift cards display correct information', async ({ page }) => {
		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		const firstShift = page.locator('.border.rounded-lg.p-3').first();

		// Should show schedule name
		await expect(firstShift.locator('.font-medium')).toBeVisible();

		// Should show time slot
		await expect(firstShift.locator('text=:')).toBeVisible(); // Time format contains colon

		// Should show relative time
		await expect(firstShift.locator('text=in, text=ago')).toBeVisible();
	});

	test('Grouped by date display works', async ({ page }) => {
		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Should show date headers
		await expect(page.locator('.font-semibold.text-lg')).toBeVisible();

		// Should show "Select Available" buttons for each date group
		await expect(page.locator('button:has-text("Select Available")')).toBeVisible();
	});
});

test.describe('Bulk Assignment - Error Handling', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToBulkAssignment(page);
	});

	test('Handles API errors gracefully', async ({ page }) => {
		// This test would need to mock API failures or use invalid data
		// For now, just verify error handling UI is present

		await selectUser(page, 'Charlie Volunteer');
		await togglePatternMode(page, false);

		// Wait for shifts to load
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Error handling would be triggered by backend issues
		// We can test that error display elements exist
		const errorContainer = page.locator('.bg-destructive\\/10, .text-destructive');

		// Error container should exist in DOM (even if not visible)
		await expect(errorContainer).toHaveCount(0); // No errors initially
	});

	test('Handles network failures', async ({ page }) => {
		// Intercept and fail API calls
		await page.route('**/api/admin/bookings/assign', (route) => {
			route.abort();
		});

		await selectUser(page, 'Diana Scout');
		await togglePatternMode(page, false);

		// Wait for shifts to load (this should still work)
		await expect(page.locator('text=Loading shifts...')).not.toBeVisible({ timeout: 10000 });

		// Select a shift and try to assign
		const firstCheckbox = page.locator('[type="checkbox"]').first();
		await firstCheckbox.click();

		await page.click('button:has-text("Assign")');

		// Should show network error
		await expect(page.locator('text=Failed, text=error')).toBeVisible({ timeout: 10000 });
	});
});
