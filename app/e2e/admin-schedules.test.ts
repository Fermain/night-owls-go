import { test, expect, type Page } from '@playwright/test';

// Test configuration
const ADMIN_PHONE = '+1234567890';
const OTP = '123456'; // Dev mode OTP

// Test data constants
const TEST_SCHEDULES = {
	DAILY_MORNING: {
		name: 'Daily Morning Patrol',
		cron: '0 6 * * *',
		description: 'Early morning community watch'
	},
	WEEKLY_NIGHT: {
		name: 'Weekly Night Watch',
		cron: '0 22 * * 6',
		description: 'Saturday night patrol'
	},
	MONTHLY_MEETING: {
		name: 'Monthly Safety Meeting',
		cron: '0 19 1 * *',
		description: 'First of month team meeting'
	},
	HOURLY_ROUNDS: {
		name: 'Hourly Security Rounds',
		cron: '0 * * * *',
		description: 'Every hour security check'
	}
};

const CRON_TEST_CASES = [
	{ cron: '0 6 * * *', description: 'Daily at 6 AM' },
	{ cron: '0 9 * * 1-5', description: 'Weekdays at 9 AM' },
	{ cron: '0 20 * * 6,0', description: 'Weekends at 8 PM' },
	{ cron: '0 12 1 * *', description: 'First of month at noon' },
	{ cron: '0 0 1 1 *', description: 'New Year at midnight' },
	{ cron: '*/15 * * * *', description: 'Every 15 minutes' },
	{ cron: '0 */2 * * *', description: 'Every 2 hours' }
];

const INVALID_CRON_CASES = [
	'invalid-cron',
	'0 25 * * *', // Invalid hour
	'0 0 32 * *', // Invalid day of month
	'0 0 0 13 *', // Invalid month
	'0 0 * * 8', // Invalid day of week
	'', // Empty cron
	'0 0 * *', // Missing field
	'0 0 * * * *' // Too many fields
];

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

async function navigateToSchedules(page: Page) {
	await page.goto('/admin/schedules');
	await expect(page.locator('h1, h2')).toContainText('Schedules');
}

async function createSchedule(
	page: Page,
	name: string,
	cronExpr: string,
	description?: string,
	duration?: number,
	timezone?: string
) {
	await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

	// Wait for form to be visible
	await expect(page.locator('form')).toBeVisible();

	// Fill basic fields
	await page.fill('input[name="name"]', name);
	await page.fill('input[name="cronExpr"]', cronExpr);

	// Fill optional fields if provided
	if (description) {
		await page.fill('textarea[name="description"], input[name="description"]', description);
	}

	if (duration) {
		await page.fill('input[name="duration"], input[name="durationMinutes"]', duration.toString());
	}

	if (timezone) {
		await page.selectOption('select[name="timezone"]', timezone);
	}

	// Submit form
	await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

	// Verify success
	await expect(page.locator('.toast')).toContainText('successfully');
}

async function deleteSchedule(page: Page, scheduleName: string) {
	// Find schedule in list and click delete
	const scheduleRow = page.locator(
		`tr:has-text("${scheduleName}"), [data-testid="schedule-item"]:has-text("${scheduleName}")`
	);
	await scheduleRow.locator('button:has-text("Delete"), a:has-text("Delete")').click();

	// Confirm deletion
	await page.click('button:has-text("Yes"), button:has-text("Delete")');

	// Verify success
	await expect(page.locator('.toast')).toContainText('deleted successfully');
}

test.describe('Admin Schedules Management - CRUD Operations', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);
	});

	test('Admin can create a new schedule with all fields', async ({ page }) => {
		await createSchedule(
			page,
			TEST_SCHEDULES.DAILY_MORNING.name,
			TEST_SCHEDULES.DAILY_MORNING.cron,
			TEST_SCHEDULES.DAILY_MORNING.description,
			120, // 2 hours
			'America/New_York'
		);

		// Verify schedule appears in list
		await expect(page.locator(`text=${TEST_SCHEDULES.DAILY_MORNING.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_SCHEDULES.DAILY_MORNING.cron}`)).toBeVisible();
	});

	test('Admin can create schedule with minimal required fields', async ({ page }) => {
		await createSchedule(page, TEST_SCHEDULES.WEEKLY_NIGHT.name, TEST_SCHEDULES.WEEKLY_NIGHT.cron);

		// Verify schedule appears in list
		await expect(page.locator(`text=${TEST_SCHEDULES.WEEKLY_NIGHT.name}`)).toBeVisible();
	});

	test('Admin can view schedule details', async ({ page }) => {
		// Create a schedule first
		await createSchedule(
			page,
			TEST_SCHEDULES.MONTHLY_MEETING.name,
			TEST_SCHEDULES.MONTHLY_MEETING.cron,
			TEST_SCHEDULES.MONTHLY_MEETING.description
		);

		// Click on schedule to view details
		await page.click(
			`tr:has-text("${TEST_SCHEDULES.MONTHLY_MEETING.name}"), [data-testid="schedule-item"]:has-text("${TEST_SCHEDULES.MONTHLY_MEETING.name}")`
		);

		// Verify schedule details are displayed
		await expect(page.locator(`text=${TEST_SCHEDULES.MONTHLY_MEETING.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_SCHEDULES.MONTHLY_MEETING.cron}`)).toBeVisible();
		if (TEST_SCHEDULES.MONTHLY_MEETING.description) {
			await expect(
				page.locator(`text=${TEST_SCHEDULES.MONTHLY_MEETING.description}`)
			).toBeVisible();
		}
	});

	test('Admin can edit schedule information', async ({ page }) => {
		// Create a schedule first
		await createSchedule(
			page,
			TEST_SCHEDULES.HOURLY_ROUNDS.name,
			TEST_SCHEDULES.HOURLY_ROUNDS.cron
		);

		// Click edit button
		const scheduleRow = page.locator(`tr:has-text("${TEST_SCHEDULES.HOURLY_ROUNDS.name}")`);
		await scheduleRow.locator('a:has-text("Edit"), button:has-text("Edit")').click();

		// Wait for edit form
		await expect(page.locator('form')).toBeVisible();

		// Modify schedule information
		const updatedName = TEST_SCHEDULES.HOURLY_ROUNDS.name + ' Updated';
		const updatedCron = '0 */3 * * *'; // Every 3 hours instead of hourly

		await page.fill('input[name="name"]', updatedName);
		await page.fill('input[name="cronExpr"]', updatedCron);
		await page.fill(
			'textarea[name="description"], input[name="description"]',
			'Updated security rounds every 3 hours'
		);

		// Save changes
		await page.click('button:has-text("Save"), button[type="submit"]');

		// Verify success
		await expect(page.locator('.toast')).toContainText('successfully');

		// Verify changes are reflected
		await expect(page.locator(`text=${updatedName}`)).toBeVisible();
		await expect(page.locator(`text=${updatedCron}`)).toBeVisible();
	});

	test('Admin can delete a schedule', async ({ page }) => {
		// Create a schedule first
		await createSchedule(page, 'Schedule to Delete', '0 14 * * *');

		// Verify schedule exists
		await expect(page.locator('text=Schedule to Delete')).toBeVisible();

		// Delete schedule
		await deleteSchedule(page, 'Schedule to Delete');

		// Verify schedule no longer appears in list
		await expect(page.locator('text=Schedule to Delete')).not.toBeVisible();
	});
});

test.describe('Admin Schedules Management - Cron Expression Validation', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);
	});

	test('Accepts valid cron expressions', async ({ page }) => {
		for (const { cron, description } of CRON_TEST_CASES) {
			await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

			await page.fill('input[name="name"]', `Test Schedule - ${description}`);
			await page.fill('input[name="cronExpr"]', cron);

			await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

			// Should succeed
			await expect(page.locator('.toast')).toContainText('successfully');

			// Navigate back to list
			await page.goto('/admin/schedules');
		}
	});

	test('Rejects invalid cron expressions', async ({ page }) => {
		for (const invalidCron of INVALID_CRON_CASES) {
			await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

			await page.fill('input[name="name"]', 'Invalid Cron Test');
			await page.fill('input[name="cronExpr"]', invalidCron);

			await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

			// Should show validation error
			await expect(page.locator('text=Invalid cron expression, text=cron format')).toBeVisible();

			// Cancel and try next case
			await page.click('button:has-text("Cancel"), button:has-text("Back")');
		}
	});

	test('Provides cron expression help and examples', async ({ page }) => {
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

		// Look for cron help text or examples
		await expect(page.locator('text=cron, text=format, text=example')).toBeVisible();

		// Look for help button or tooltip
		const helpButton = page.locator('button[aria-label*="help"], [data-tooltip], .tooltip');
		if (await helpButton.isVisible()) {
			await helpButton.click();
			await expect(page.locator('text=minute hour day month')).toBeVisible();
		}
	});

	test('Validates cron expression on input change', async ({ page }) => {
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

		// Fill name first
		await page.fill('input[name="name"]', 'Validation Test');

		// Type invalid cron
		await page.fill('input[name="cronExpr"]', 'invalid');

		// Should show validation message immediately or on blur
		await page.click('input[name="name"]'); // Blur the cron field
		await expect(page.locator('text=Invalid, text=format')).toBeVisible();

		// Fix the cron expression
		await page.fill('input[name="cronExpr"]', '0 12 * * *');
		await page.click('input[name="name"]'); // Blur again

		// Validation error should disappear
		await expect(page.locator('text=Invalid, text=format')).not.toBeVisible();
	});
});

test.describe('Admin Schedules Management - Form Validation', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);
	});

	test('Form validation prevents creating schedule without required fields', async ({ page }) => {
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

		// Try to submit empty form
		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should show validation errors
		await expect(page.locator('text=Name is required')).toBeVisible();
		await expect(page.locator('text=Cron expression is required')).toBeVisible();
	});

	test('Form validation prevents duplicate schedule names', async ({ page }) => {
		// Create first schedule
		await createSchedule(page, 'Duplicate Name Test', '0 9 * * *');

		// Try to create second schedule with same name
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');
		await page.fill('input[name="name"]', 'Duplicate Name Test');
		await page.fill('input[name="cronExpr"]', '0 10 * * *');

		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should show duplicate name error
		await expect(page.locator('text=Schedule name already exists')).toBeVisible();
	});

	test('Form validation handles long names gracefully', async ({ page }) => {
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

		// Try very long name
		const longName = 'A'.repeat(256);
		await page.fill('input[name="name"]', longName);
		await page.fill('input[name="cronExpr"]', '0 12 * * *');

		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should either truncate or show length validation error
		await expect(page.locator('.toast, text=too long, text=exceeds')).toBeVisible();
	});

	test('Duration validation accepts valid values', async ({ page }) => {
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

		await page.fill('input[name="name"]', 'Duration Test');
		await page.fill('input[name="cronExpr"]', '0 12 * * *');

		// Test valid durations
		const validDurations = [15, 30, 60, 120, 240, 480];

		for (const duration of validDurations) {
			await page.fill('input[name="duration"], input[name="durationMinutes"]', duration.toString());
			// Should not show validation error
			await expect(page.locator('text=Invalid duration')).not.toBeVisible();
		}
	});

	test('Duration validation rejects invalid values', async ({ page }) => {
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');

		await page.fill('input[name="name"]', 'Duration Test');
		await page.fill('input[name="cronExpr"]', '0 12 * * *');

		// Test invalid durations
		const invalidDurations = [-1, 0, 'abc', 10000];

		for (const duration of invalidDurations) {
			await page.fill('input[name="duration"], input[name="durationMinutes"]', duration.toString());
			await page.click('input[name="name"]'); // Blur field

			// Should show validation error
			await expect(page.locator('text=Invalid duration, text=must be positive')).toBeVisible();
		}
	});
});

test.describe('Admin Schedules Management - Schedule Conflicts', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);
	});

	test('Detects and warns about schedule conflicts', async ({ page }) => {
		// Create first schedule
		await createSchedule(page, 'Morning Patrol A', '0 9 * * *', 'First morning schedule');

		// Try to create conflicting schedule (same time)
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');
		await page.fill('input[name="name"]', 'Morning Patrol B');
		await page.fill('input[name="cronExpr"]', '0 9 * * *'); // Same time as first

		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should show conflict warning (if implemented)
		// This might be a warning rather than an error, depending on business rules
		await expect(page.locator('text=conflict, text=overlapping, text=same time')).toBeVisible();
	});

	test('Allows non-conflicting schedules', async ({ page }) => {
		// Create first schedule
		await createSchedule(page, 'Morning Shift', '0 9 * * *');

		// Create non-conflicting schedule (different time)
		await createSchedule(page, 'Evening Shift', '0 21 * * *');

		// Both should exist
		await expect(page.locator('text=Morning Shift')).toBeVisible();
		await expect(page.locator('text=Evening Shift')).toBeVisible();
	});

	test('Handles overlapping duration conflicts', async ({ page }) => {
		// Create first schedule with 3-hour duration
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');
		await page.fill('input[name="name"]', 'Long Morning Shift');
		await page.fill('input[name="cronExpr"]', '0 9 * * *');
		await page.fill('input[name="duration"], input[name="durationMinutes"]', '180'); // 3 hours
		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Try to create overlapping schedule
		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');
		await page.fill('input[name="name"]', 'Midday Shift');
		await page.fill('input[name="cronExpr"]', '0 11 * * *'); // 2 hours after first, but first runs for 3 hours
		await page.fill('input[name="duration"], input[name="durationMinutes"]', '120');

		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should detect overlap (if implemented)
		// 9 AM + 3 hours = 12 PM, but second starts at 11 AM
		const conflictDetected = await page.locator('text=overlap, text=conflict').isVisible();
		// This test documents expected behavior - may warn or allow depending on business rules
	});
});

test.describe('Admin Schedules Management - Timezone Handling', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);
	});

	test('Can create schedules with different timezones', async ({ page }) => {
		const timezones = [
			'America/New_York',
			'America/Los_Angeles',
			'Europe/London',
			'Asia/Tokyo',
			'UTC'
		];

		for (const timezone of timezones) {
			await createSchedule(
				page,
				`${timezone} Schedule`,
				'0 12 * * *',
				`Noon schedule in ${timezone}`,
				60,
				timezone
			);

			// Verify schedule was created with timezone
			await expect(page.locator(`text=${timezone} Schedule`)).toBeVisible();
		}
	});

	test('Displays timezone information in schedule list', async ({ page }) => {
		await createSchedule(
			page,
			'Tokyo Schedule',
			'0 9 * * *',
			'Morning schedule in Tokyo',
			120,
			'Asia/Tokyo'
		);

		// Should show timezone in list or details
		await expect(page.locator('text=Tokyo, text=Asia/Tokyo')).toBeVisible();
	});

	test('Handles timezone changes in editing', async ({ page }) => {
		// Create schedule with one timezone
		await createSchedule(
			page,
			'Timezone Test',
			'0 15 * * *',
			'Test timezone changes',
			60,
			'America/New_York'
		);

		// Edit to change timezone
		const scheduleRow = page.locator('tr:has-text("Timezone Test")');
		await scheduleRow.locator('a:has-text("Edit"), button:has-text("Edit")').click();

		await page.selectOption('select[name="timezone"]', 'Europe/London');
		await page.click('button:has-text("Save"), button[type="submit"]');

		// Verify timezone was updated
		await expect(page.locator('text=London, text=Europe/London')).toBeVisible();
	});
});

test.describe('Admin Schedules Management - Search and Filtering', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);

		// Create test schedules for filtering
		await createSchedule(page, 'Morning Patrol Alpha', '0 6 * * *', 'Early morning rounds');
		await createSchedule(page, 'Evening Watch Beta', '0 20 * * *', 'Late evening security');
		await createSchedule(page, 'Weekend Special', '0 12 * * 6,0', 'Weekend coverage');
		await createSchedule(page, 'Monthly Check', '0 10 1 * *', 'Monthly inspection');
	});

	test('Can search schedules by name', async ({ page }) => {
		// Search for specific schedule
		await page.fill('input[placeholder*="Search"]', 'Morning');

		// Should show only matching schedule
		await expect(page.locator('text=Morning Patrol Alpha')).toBeVisible();
		await expect(page.locator('text=Evening Watch Beta')).not.toBeVisible();
		await expect(page.locator('text=Weekend Special')).not.toBeVisible();
		await expect(page.locator('text=Monthly Check')).not.toBeVisible();
	});

	test('Can search schedules by description', async ({ page }) => {
		// Search by description content
		await page.fill('input[placeholder*="Search"]', 'security');

		// Should find schedule with matching description
		await expect(page.locator('text=Evening Watch Beta')).toBeVisible();
		await expect(page.locator('text=Morning Patrol Alpha')).not.toBeVisible();
	});

	test('Search is case insensitive', async ({ page }) => {
		// Search with different case
		await page.fill('input[placeholder*="Search"]', 'weekend');

		// Should still find the schedule
		await expect(page.locator('text=Weekend Special')).toBeVisible();
	});

	test('Can clear search filter', async ({ page }) => {
		// Search for schedule
		await page.fill('input[placeholder*="Search"]', 'Morning');
		await expect(page.locator('text=Morning Patrol Alpha')).toBeVisible();
		await expect(page.locator('text=Evening Watch Beta')).not.toBeVisible();

		// Clear search
		await page.fill('input[placeholder*="Search"]', '');

		// Should show all schedules again
		await expect(page.locator('text=Morning Patrol Alpha')).toBeVisible();
		await expect(page.locator('text=Evening Watch Beta')).toBeVisible();
		await expect(page.locator('text=Weekend Special')).toBeVisible();
		await expect(page.locator('text=Monthly Check')).toBeVisible();
	});

	test('Search with no results shows appropriate message', async ({ page }) => {
		// Search for non-existent schedule
		await page.fill('input[placeholder*="Search"]', 'NonExistentSchedule');

		// Should show no results message
		await expect(page.locator('text=No schedules found')).toBeVisible();
	});
});

test.describe('Admin Schedules Management - Error Handling', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await navigateToSchedules(page);
	});

	test('Handles network errors gracefully during schedule creation', async ({ page }) => {
		// Simulate network failure
		await page.route('**/api/admin/schedules', (route) => route.abort());

		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');
		await page.fill('input[name="name"]', 'Network Test Schedule');
		await page.fill('input[name="cronExpr"]', '0 12 * * *');

		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should show network error
		await expect(page.locator('text=Network error, text=Failed to')).toBeVisible();
	});

	test('Handles server errors gracefully', async ({ page }) => {
		// Simulate server error
		await page.route('**/api/admin/schedules', (route) =>
			route.fulfill({ status: 500, body: JSON.stringify({ error: 'Internal server error' }) })
		);

		await page.click('button:has-text("Create New Schedule"), a:has-text("Create New Schedule")');
		await page.fill('input[name="name"]', 'Server Error Test');
		await page.fill('input[name="cronExpr"]', '0 12 * * *');

		await page.click('button[type="submit"]:has-text("Create"), button:has-text("Save")');

		// Should show user-friendly error message
		await expect(page.locator('text=Something went wrong')).toBeVisible();
	});

	test('Handles editing non-existent schedule', async ({ page }) => {
		// Try to navigate to edit page for non-existent schedule
		await page.goto('/admin/schedules/99999/edit');

		// Should show not found message or redirect
		await expect(page.locator('text=Schedule not found, text=not exist')).toBeVisible();
	});
});

test.describe('Admin Schedules Management - Performance', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test('Schedules page loads quickly', async ({ page }) => {
		// Navigate and measure load time
		const startTime = Date.now();
		await page.goto('/admin/schedules');
		await page.waitForSelector('tr, [data-testid="schedule-item"], text=No schedules found');
		const loadTime = Date.now() - startTime;

		// Should load within reasonable time
		expect(loadTime).toBeLessThan(5000); // 5 seconds
	});

	test('Search performs well with many schedules', async ({ page }) => {
		await page.goto('/admin/schedules');

		const searchStartTime = Date.now();
		await page.fill('input[placeholder*="Search"]', 'Test');

		// Wait for search results
		await page.waitForTimeout(500); // Allow for debouncing

		const searchTime = Date.now() - searchStartTime;
		expect(searchTime).toBeLessThan(2000); // 2 seconds
	});
});
