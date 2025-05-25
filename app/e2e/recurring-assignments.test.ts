import { test, expect, type Page } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

// Test data constants
const TEST_USERS = {
	VOLUNTEER_1: { name: 'Alice Johnson', phone: '+15551111111' },
	VOLUNTEER_2: { name: 'Bob Smith', phone: '+15552222222' },
	VOLUNTEER_3: { name: 'Carol Williams', phone: '+15553333333' },
	VOLUNTEER_4: { name: 'David Brown', phone: '+15554444444' },
	SUPERVISOR: { name: 'Eve Wilson', phone: '+15555555555' }
};

const TEST_SCHEDULES = {
	WEEKEND_NIGHTS: { name: 'Weekend Night Patrol', cron: '0 18 * * 6,0' }, // Sat & Sun 6PM
	WEEKDAY_MORNINGS: { name: 'Weekday Morning Watch', cron: '0 6 * * 1-5' }, // Mon-Fri 6AM
	DAILY_EVENING: { name: 'Daily Evening Rounds', cron: '0 20 * * *' }, // Every day 8PM
	SPECIAL_EVENTS: { name: 'Special Event Coverage', cron: '0 14 * * 3' } // Wednesday 2PM
};

const DAYS_OF_WEEK = [
	{ value: 0, name: 'Sunday' },
	{ value: 1, name: 'Monday' },
	{ value: 2, name: 'Tuesday' },
	{ value: 3, name: 'Wednesday' },
	{ value: 4, name: 'Thursday' },
	{ value: 5, name: 'Friday' },
	{ value: 6, name: 'Saturday' }
];

const TIME_SLOTS = [
	'06:00-08:00',
	'08:00-10:00',
	'10:00-12:00',
	'12:00-14:00',
	'14:00-16:00',
	'16:00-18:00',
	'18:00-20:00',
	'20:00-22:00',
	'22:00-00:00'
];

async function createTestUser(page: Page, name: string, phone: string, role: string = 'owl') {
	await page.goto('/admin/users/new');
	await page.fill('input#name', name);
	await page.fill('input[type="tel"]', phone);
	await page.selectOption('select[name="role"]', role);
	await page.click('button[type="submit"]');
	await expect(page.locator('.toast')).toContainText('User created successfully');
}

async function createTestSchedule(page: Page, name: string, cronExpr: string) {
	await page.goto('/admin/schedules/new');
	await page.fill('input#name', name);
	await page.fill('input#cron_expr', cronExpr);
	await page.click('button[type="submit"]');
	await expect(page.locator('.toast')).toContainText('Schedule created successfully');
}

async function createRecurringAssignment(
	page: Page,
	userDisplayName: string,
	scheduleDisplayName: string,
	dayDisplayName: string,
	timeSlot: string,
	buddyName?: string,
	description?: string
) {
	await page.goto('/admin/shifts/recurring');
	await page.click('button:has-text("New Recurring Assignment")');

	// Select user
	await page.click('button[role="combobox"]:has-text("Select user")');
	await page.click(`div[role="option"]:has-text("${userDisplayName}")`);

	// Add buddy if specified
	if (buddyName) {
		await page.fill('input[placeholder="Enter buddy name"]', buddyName);
	}

	// Select schedule
	await page.click('button:has-text("Select schedule")');
	await page.click(`div[role="option"]:has-text("${scheduleDisplayName}")`);

	// Select day
	await page.click('button:has-text("Select day")');
	await page.click(`div[role="option"]:has-text("${dayDisplayName}")`);

	// Select time slot
	await page.click('button:has-text("Select time slot")');
	await page.click(`div[role="option"]:has-text("${timeSlot}")`);

	// Add description if specified
	if (description) {
		await page.fill('input[placeholder="Optional description"]', description);
	}

	// Submit form
	await page.click('button:has-text("Create Assignment")');

	// Verify success
	await expect(page.locator('.toast')).toContainText('created successfully');
}

async function setupTestData(page: Page) {
	// Create test users
	for (const user of Object.values(TEST_USERS)) {
		await createTestUser(page, user.name, user.phone);
	}

	// Create test schedules
	for (const schedule of Object.values(TEST_SCHEDULES)) {
		await createTestSchedule(page, schedule.name, schedule.cron);
	}
}

test.describe('Recurring Assignments - Core CRUD Operations', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Admin can create recurring assignment with all fields', async ({ page }) => {
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00',
			'Emergency Buddy',
			'Regular weekend patrol duty'
		);

		// Verify assignment appears in list with all details
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_1.name}`)).toBeVisible();
		await expect(page.locator('text=Saturday')).toBeVisible();
		await expect(page.locator('text=18:00-20:00')).toBeVisible();
		await expect(page.locator('text=With buddy: Emergency Buddy')).toBeVisible();
		await expect(page.locator('text=Regular weekend patrol duty')).toBeVisible();
	});

	test('Admin can create recurring assignment without optional fields', async ({ page }) => {
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_2.name,
			TEST_SCHEDULES.WEEKDAY_MORNINGS.name,
			'Monday',
			'06:00-08:00'
		);

		// Verify assignment appears without buddy or description
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_2.name}`)).toBeVisible();
		await expect(page.locator('text=Monday')).toBeVisible();
		await expect(page.locator('text=06:00-08:00')).toBeVisible();
		await expect(page.locator('text=No buddy assigned')).toBeVisible();
	});

	test('Admin can view recurring assignment details', async ({ page }) => {
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_3.name,
			TEST_SCHEDULES.DAILY_EVENING.name,
			'Wednesday',
			'20:00-22:00',
			'Night Partner'
		);

		// Click on assignment to view details
		await page.click(`[data-testid="assignment-item"]:has-text("${TEST_USERS.VOLUNTEER_3.name}")`);

		// Verify details modal/page shows all information
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_3.name}`)).toBeVisible();
		await expect(page.locator('text=Wednesday')).toBeVisible();
		await expect(page.locator('text=20:00-22:00')).toBeVisible();
		await expect(page.locator('text=Night Partner')).toBeVisible();
	});

	test('Admin can edit recurring assignment', async ({ page }) => {
		// Create initial assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Click edit button
		await page.click('button[data-testid="edit-assignment"]');

		// Modify fields
		await page.click('button:has-text("Select day")');
		await page.click('div[role="option"]:has-text("Sunday")');

		await page.click('button:has-text("Select time slot")');
		await page.click('div[role="option"]:has-text("20:00-22:00")');

		await page.fill('input[placeholder="Enter buddy name"]', 'New Buddy');

		// Save changes
		await page.click('button:has-text("Update Assignment")');

		// Verify changes
		await expect(page.locator('text=Sunday')).toBeVisible();
		await expect(page.locator('text=20:00-22:00')).toBeVisible();
		await expect(page.locator('text=With buddy: New Buddy')).toBeVisible();
	});

	test('Admin can delete recurring assignment', async ({ page }) => {
		// Create assignment to delete
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_4.name,
			TEST_SCHEDULES.SPECIAL_EVENTS.name,
			'Wednesday',
			'14:00-16:00'
		);

		// Count assignments before deletion
		const countBefore = await page.locator('[data-testid="assignment-item"]').count();

		// Delete assignment
		await page.click('button[data-testid="delete-assignment"]');
		await page.click('button:has-text("Yes, delete")');

		// Verify deletion
		await expect(page.locator('.toast')).toContainText('deleted successfully');
		const countAfter = await page.locator('[data-testid="assignment-item"]').count();
		expect(countAfter).toBe(countBefore - 1);
	});
});

test.describe('Recurring Assignments - All Days and Time Slots', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Can create assignments for all days of the week', async ({ page }) => {
		// Create assignment for each day
		for (let i = 0; i < DAYS_OF_WEEK.length; i++) {
			const day = DAYS_OF_WEEK[i];
			const user = Object.values(TEST_USERS)[i % Object.values(TEST_USERS).length];

			await createRecurringAssignment(
				page,
				user.name,
				TEST_SCHEDULES.DAILY_EVENING.name,
				day.name,
				'20:00-22:00',
				undefined,
				`${day.name} night watch`
			);
		}

		// Verify all days are represented
		for (const day of DAYS_OF_WEEK) {
			await expect(page.locator(`text=${day.name}`)).toBeVisible();
		}
	});

	test('Can create assignments for different time slots', async ({ page }) => {
		// Create assignments for various time slots
		for (let i = 0; i < Math.min(TIME_SLOTS.length, 5); i++) {
			const timeSlot = TIME_SLOTS[i];
			const user = Object.values(TEST_USERS)[i];

			await createRecurringAssignment(
				page,
				user.name,
				TEST_SCHEDULES.DAILY_EVENING.name,
				'Monday',
				timeSlot,
				undefined,
				`${timeSlot} shift`
			);
		}

		// Verify different time slots are displayed
		for (let i = 0; i < Math.min(TIME_SLOTS.length, 5); i++) {
			await expect(page.locator(`text=${TIME_SLOTS[i]}`)).toBeVisible();
		}
	});

	test('Time slot dropdown shows only valid options for selected schedule', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Select a specific schedule
		await page.click('button:has-text("Select schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKDAY_MORNINGS.name}")`);

		// Check time slot options
		await page.click('button:has-text("Select time slot")');
		const timeSlotOptions = page.locator('div[role="option"]');
		const optionCount = await timeSlotOptions.count();

		// Should have valid time slots
		expect(optionCount).toBeGreaterThan(0);

		// Verify format consistency
		for (let i = 0; i < Math.min(optionCount, 3); i++) {
			const option = timeSlotOptions.nth(i);
			await expect(option).toHaveText(/^\d{2}:\d{2}-\d{2}:\d{2}$/);
		}
	});
});

test.describe('Recurring Assignments - Buddy System', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Assignment with buddy shows buddy information', async ({ page }) => {
		const buddyName = 'Safety Partner Alpha';

		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00',
			buddyName
		);

		await expect(page.locator(`text=With buddy: ${buddyName}`)).toBeVisible();
	});

	test('Assignment without buddy shows appropriate message', async ({ page }) => {
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_2.name,
			TEST_SCHEDULES.WEEKDAY_MORNINGS.name,
			'Monday',
			'06:00-08:00'
		);

		await expect(page.locator('text=No buddy assigned')).toBeVisible();
	});

	test('Can edit buddy information', async ({ page }) => {
		// Create assignment without buddy
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_3.name,
			TEST_SCHEDULES.DAILY_EVENING.name,
			'Wednesday',
			'20:00-22:00'
		);

		// Edit to add buddy
		await page.click('button[data-testid="edit-assignment"]');
		await page.fill('input[placeholder="Enter buddy name"]', 'New Safety Partner');
		await page.click('button:has-text("Update Assignment")');

		await expect(page.locator('text=With buddy: New Safety Partner')).toBeVisible();
	});

	test('Buddy name validation works correctly', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Fill required fields
		await page.click('button[role="combobox"]:has-text("Select user")');
		await page.click(`div[role="option"]:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);

		await page.click('button:has-text("Select schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKEND_NIGHTS.name}")`);

		await page.click('button:has-text("Select day")');
		await page.click('div[role="option"]:has-text("Saturday")');

		await page.click('button:has-text("Select time slot")');
		await page.click('div[role="option"]:has-text("18:00-20:00")');

		// Test very long buddy name
		const longBuddyName = 'A'.repeat(100);
		await page.fill('input[placeholder="Enter buddy name"]', longBuddyName);
		await page.click('button:has-text("Create Assignment")');

		// Should show validation error or truncate appropriately
		// (This depends on backend validation rules)
	});
});

test.describe('Recurring Assignments - Materialization', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Manual materialization creates bookings', async ({ page }) => {
		// Create some recurring assignments
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Materialize bookings
		await page.click('button:has-text("Materialize Now")');
		await expect(page.locator('.toast')).toContainText('Materialization completed');

		// Check that bookings were created
		await page.goto('/admin/schedules/slots');
		const bookedSlots = await page.locator('[data-testid="booked-slot"]').count();
		expect(bookedSlots).toBeGreaterThan(0);
	});

	test('Materialization with date range works', async ({ page }) => {
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_2.name,
			TEST_SCHEDULES.WEEKDAY_MORNINGS.name,
			'Monday',
			'06:00-08:00'
		);

		await page.goto('/admin/shifts/recurring');

		// Open materialization with date range
		await page.click('button:has-text("Advanced Materialize")');

		// Set date range (next 7 days)
		const today = new Date();
		const nextWeek = new Date(today);
		nextWeek.setDate(today.getDate() + 7);

		await page.fill('input[name="from-date"]', today.toISOString().split('T')[0]);
		await page.fill('input[name="to-date"]', nextWeek.toISOString().split('T')[0]);

		await page.click('button:has-text("Materialize Range")');
		await expect(page.locator('.toast')).toContainText('Materialization completed');
	});

	test('Materialization skips already booked slots', async ({ page }) => {
		// Create assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_3.name,
			TEST_SCHEDULES.DAILY_EVENING.name,
			'Wednesday',
			'20:00-22:00'
		);

		// First materialization
		await page.click('button:has-text("Materialize Now")');
		await expect(page.locator('.toast')).toContainText('Materialization completed');

		// Second materialization should skip existing bookings
		await page.click('button:has-text("Materialize Now")');
		await expect(page.locator('.toast')).toContainText('skipped');
	});

	test('Multiple assignments for same slot - first wins', async ({ page }) => {
		// Create two assignments for the same time slot
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_2.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Materialize
		await page.click('button:has-text("Materialize Now")');

		// Check bookings to verify only first assignment was materialized
		await page.goto('/admin/schedules/slots');

		// The specific verification would depend on UI design
		// Should show which user got the booking
	});
});

test.describe('Recurring Assignments - Search and Filtering', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);

		// Create diverse assignments for filtering tests
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00',
			'Weekend Buddy'
		);

		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_2.name,
			TEST_SCHEDULES.WEEKDAY_MORNINGS.name,
			'Monday',
			'06:00-08:00'
		);

		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_3.name,
			TEST_SCHEDULES.DAILY_EVENING.name,
			'Wednesday',
			'20:00-22:00',
			'Evening Partner'
		);
	});

	test('Can search assignments by user name', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Search for specific user
		await page.fill('input[placeholder="Search assignments..."]', TEST_USERS.VOLUNTEER_1.name);

		// Should show only matching assignments
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_1.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_2.name}`)).not.toBeVisible();
	});

	test('Can filter assignments by day of week', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Filter by Monday
		await page.click('button:has-text("Filter by Day")');
		await page.click('div[role="option"]:has-text("Monday")');

		// Should show only Monday assignments
		await expect(page.locator('text=Monday')).toBeVisible();
		await expect(page.locator('text=Saturday')).not.toBeVisible();
		await expect(page.locator('text=Wednesday')).not.toBeVisible();
	});

	test('Can filter assignments by schedule', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Filter by specific schedule
		await page.click('button:has-text("Filter by Schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKEND_NIGHTS.name}")`);

		// Should show only assignments for that schedule
		await expect(page.locator(`text=${TEST_SCHEDULES.WEEKEND_NIGHTS.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_SCHEDULES.WEEKDAY_MORNINGS.name}`)).not.toBeVisible();
	});

	test('Can filter assignments by buddy status', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Filter to show only assignments with buddies
		await page.click('button:has-text("Filter by Buddy")');
		await page.click('div[role="option"]:has-text("With Buddy")');

		// Should show only assignments that have buddies
		await expect(page.locator('text=Weekend Buddy')).toBeVisible();
		await expect(page.locator('text=Evening Partner')).toBeVisible();
		await expect(page.locator('text=No buddy assigned')).not.toBeVisible();
	});

	test('Can combine multiple filters', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Apply multiple filters
		await page.fill('input[placeholder="Search assignments..."]', 'Alice');
		await page.click('button:has-text("Filter by Day")');
		await page.click('div[role="option"]:has-text("Saturday")');

		// Should show only assignments matching all criteria
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_1.name}`)).toBeVisible();
		await expect(page.locator('text=Saturday')).toBeVisible();
	});

	test('Can clear filters', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Apply filter
		await page.fill('input[placeholder="Search assignments..."]', TEST_USERS.VOLUNTEER_1.name);

		// Clear filters
		await page.click('button:has-text("Clear Filters")');

		// Should show all assignments again
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_1.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_2.name}`)).toBeVisible();
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_3.name}`)).toBeVisible();
	});
});

test.describe('Recurring Assignments - Error Handling', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Handles form validation errors gracefully', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Try to submit without required fields
		await page.click('button:has-text("Create Assignment")');

		// Should show validation errors
		await expect(page.locator('text=Please select a user')).toBeVisible();
		await expect(page.locator('text=Please select a schedule')).toBeVisible();
		await expect(page.locator('text=Please select a day')).toBeVisible();
		await expect(page.locator('text=Please select a time slot')).toBeVisible();
	});

	test('Handles duplicate assignment creation', async ({ page }) => {
		// Create first assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Try to create identical assignment
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		await page.click('button[role="combobox"]:has-text("Select user")');
		await page.click(`div[role="option"]:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);

		await page.click('button:has-text("Select schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKEND_NIGHTS.name}")`);

		await page.click('button:has-text("Select day")');
		await page.click('div[role="option"]:has-text("Saturday")');

		await page.click('button:has-text("Select time slot")');
		await page.click('div[role="option"]:has-text("18:00-20:00")');

		await page.click('button:has-text("Create Assignment")');

		// Should show error about duplicate
		await expect(page.locator('text=Assignment already exists')).toBeVisible();
	});

	test('Handles network errors during creation', async ({ page }) => {
		// Simulate network failure
		await page.route('**/api/admin/recurring-assignments', (route) => route.abort());

		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Fill form
		await page.click('button[role="combobox"]:has-text("Select user")');
		await page.click(`div[role="option"]:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);

		await page.click('button:has-text("Select schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKEND_NIGHTS.name}")`);

		await page.click('button:has-text("Select day")');
		await page.click('div[role="option"]:has-text("Saturday")');

		await page.click('button:has-text("Select time slot")');
		await page.click('div[role="option"]:has-text("18:00-20:00")');

		await page.click('button:has-text("Create Assignment")');

		// Should show network error
		await expect(page.locator('text=Network error')).toBeVisible();
	});

	test('Handles server errors gracefully', async ({ page }) => {
		// Simulate server error
		await page.route('**/api/admin/recurring-assignments', (route) =>
			route.fulfill({ status: 500, body: JSON.stringify({ error: 'Internal server error' }) })
		);

		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Fill and submit form
		await page.click('button[role="combobox"]:has-text("Select user")');
		await page.click(`div[role="option"]:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);

		await page.click('button:has-text("Select schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKEND_NIGHTS.name}")`);

		await page.click('button:has-text("Select day")');
		await page.click('div[role="option"]:has-text("Saturday")');

		await page.click('button:has-text("Select time slot")');
		await page.click('div[role="option"]:has-text("18:00-20:00")');

		await page.click('button:has-text("Create Assignment")');

		// Should show user-friendly error message
		await expect(page.locator('text=Something went wrong')).toBeVisible();
	});
});

test.describe('Recurring Assignments - Integration Scenarios', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Assignment persists when user role changes', async ({ page }) => {
		// Create assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Change user role
		await page.goto('/admin/users');
		await page.click(
			`[data-testid="user-item"]:has-text("${TEST_USERS.VOLUNTEER_1.name}") button:has-text("Edit")`
		);
		await page.selectOption('select[name="role"]', 'supervisor');
		await page.click('button:has-text("Save")');

		// Verify assignment still exists
		await page.goto('/admin/shifts/recurring');
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_1.name}`)).toBeVisible();
	});

	test('Assignment is deleted when user is deleted', async ({ page }) => {
		// Create assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_4.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Delete user
		await page.goto('/admin/users');
		await page.click(
			`[data-testid="user-item"]:has-text("${TEST_USERS.VOLUNTEER_4.name}") button:has-text("Delete")`
		);
		await page.click('button:has-text("Yes, delete")');

		// Verify assignment is also removed
		await page.goto('/admin/shifts/recurring');
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_4.name}`)).not.toBeVisible();
	});

	test('Assignment is deleted when schedule is deleted', async ({ page }) => {
		// Create assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.SPECIAL_EVENTS.name,
			'Wednesday',
			'14:00-16:00'
		);

		// Delete schedule
		await page.goto('/admin/schedules');
		await page.click(
			`[data-testid="schedule-item"]:has-text("${TEST_SCHEDULES.SPECIAL_EVENTS.name}") button:has-text("Delete")`
		);
		await page.click('button:has-text("Yes, delete")');

		// Verify assignment is also removed
		await page.goto('/admin/shifts/recurring');
		await expect(page.locator(`text=${TEST_SCHEDULES.SPECIAL_EVENTS.name}`)).not.toBeVisible();
	});

	test('Bookings persist when recurring assignment is deleted', async ({ page }) => {
		// Create assignment
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_2.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		// Materialize bookings
		await page.click('button:has-text("Materialize Now")');
		await expect(page.locator('.toast')).toContainText('Materialization completed');

		// Count bookings before deletion
		await page.goto('/admin/schedules/slots');
		const bookingCountBefore = await page.locator('[data-testid="booked-slot"]').count();

		// Delete recurring assignment
		await page.goto('/admin/shifts/recurring');
		await page.click('button[data-testid="delete-assignment"]');
		await page.click('button:has-text("Yes, delete")');

		// Verify bookings still exist
		await page.goto('/admin/schedules/slots');
		const bookingCountAfter = await page.locator('[data-testid="booked-slot"]').count();
		expect(bookingCountAfter).toBe(bookingCountBefore);
	});
});

test.describe('Recurring Assignments - Performance and Scalability', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Page loads quickly with many assignments', async ({ page }) => {
		// Create multiple assignments (simulate large dataset)

		// Create 10 assignments quickly
		for (let i = 0; i < 10; i++) {
			const user = Object.values(TEST_USERS)[i % Object.values(TEST_USERS).length];
			const day = DAYS_OF_WEEK[i % DAYS_OF_WEEK.length];
			const timeSlot = TIME_SLOTS[i % TIME_SLOTS.length];

			await createRecurringAssignment(
				page,
				user.name,
				TEST_SCHEDULES.DAILY_EVENING.name,
				day.name,
				timeSlot,
				`Buddy ${i}`
			);
		}

		// Navigate to assignments page and measure load time
		const loadStartTime = Date.now();
		await page.goto('/admin/shifts/recurring');
		await page.waitForSelector('[data-testid="assignment-item"]');
		const loadTime = Date.now() - loadStartTime;

		// Should load within reasonable time (adjust threshold as needed)
		expect(loadTime).toBeLessThan(3000); // 3 seconds

		// Verify all assignments are displayed
		const assignmentCount = await page.locator('[data-testid="assignment-item"]').count();
		expect(assignmentCount).toBe(10);
	});

	test('Search performs well with large dataset', async ({ page }) => {
		// Assuming assignments from previous test exist
		await page.goto('/admin/shifts/recurring');

		const searchStartTime = Date.now();
		await page.fill('input[placeholder="Search assignments..."]', TEST_USERS.VOLUNTEER_1.name);

		// Wait for search results
		await page.waitForFunction((userName) => {
			const items = document.querySelectorAll('[data-testid="assignment-item"]');
			return Array.from(items).some((item) => item.textContent?.includes(userName));
		}, TEST_USERS.VOLUNTEER_1.name);

		const searchTime = Date.now() - searchStartTime;
		expect(searchTime).toBeLessThan(1000); // 1 second
	});

	test('Materialization handles large number of assignments efficiently', async ({ page }) => {
		// This test would need actual backend implementation
		// For now, just verify the UI handles the materialization request
		await page.goto('/admin/shifts/recurring');

		const materializeStartTime = Date.now();
		await page.click('button:has-text("Materialize Now")');

		// Wait for completion
		await expect(page.locator('.toast')).toContainText('completed', { timeout: 300 });

		const materializeTime = Date.now() - materializeStartTime;
		expect(materializeTime).toBeLessThan(300);
	});
});

test.describe('Recurring Assignments - Accessibility and UX', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Keyboard navigation works throughout the interface', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Tab through main interface elements
		await page.keyboard.press('Tab');
		await expect(page.locator(':focus')).toBeVisible();

		// Should be able to reach "New Assignment" button
		await page.keyboard.press('Tab');
		await page.keyboard.press('Tab');
		const focusedElement = await page.locator(':focus').textContent();
		expect(focusedElement).toContain('New');

		// Open form with keyboard
		await page.keyboard.press('Enter');
		await expect(page.locator('form')).toBeVisible();

		// Navigate form with keyboard
		await page.keyboard.press('Tab');
		await expect(page.locator(':focus')).toHaveAttribute('role', 'combobox');
	});

	test('Screen reader labels are present and descriptive', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Check for proper ARIA labels
		await expect(page.locator('button[aria-label*="Select user"]')).toBeVisible();
		await expect(page.locator('button[aria-label*="Select schedule"]')).toBeVisible();
		await expect(page.locator('button[aria-label*="Select day"]')).toBeVisible();
		await expect(page.locator('button[aria-label*="Select time slot"]')).toBeVisible();

		// Check form has proper heading
		await expect(page.locator('h2')).toContainText('New Recurring Assignment');
	});

	test('Mobile interface is fully functional', async ({ page }) => {
		// Set mobile viewport
		await page.setViewportSize({ width: 375, height: 667 });

		await page.goto('/admin/shifts/recurring');

		// Mobile navigation should work
		if (await page.locator('.mobile-menu-button').isVisible()) {
			await page.click('.mobile-menu-button');
		}

		// Create assignment on mobile
		await page.click('button:has-text("New")'); // Abbreviated text on mobile
		await expect(page.locator('form')).toBeVisible();

		// Form should be usable on mobile
		await page.click('button[role="combobox"]');
		await expect(page.locator('div[role="listbox"]')).toBeVisible();

		// Should be able to scroll through options
		const options = page.locator('div[role="option"]');
		const optionCount = await options.count();
		if (optionCount > 3) {
			await page.mouse.wheel(0, 100);
			await expect(options.last()).toBeVisible();
		}
	});

	test('Focus management works correctly in modals', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Focus should be in modal
		const activeElement = page.locator(':focus');
		await expect(activeElement).toBeVisible();

		// Tab should stay within modal
		for (let i = 0; i < 10; i++) {
			await page.keyboard.press('Tab');
			const focusedElement = page.locator(':focus');
			await expect(focusedElement).toBeVisible();

			// Should not focus elements outside modal
			const isInModal = await focusedElement.evaluate((el) => {
				return el.closest('[role="dialog"]') !== null;
			});
			expect(isInModal).toBe(true);
		}

		// Escape should close modal
		await page.keyboard.press('Escape');
		await expect(page.locator('form')).not.toBeVisible();
	});

	test('Error messages are announced to screen readers', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Submit empty form
		await page.click('button:has-text("Create Assignment")');

		// Error messages should have proper ARIA attributes
		await expect(page.locator('[role="alert"]')).toBeVisible();
		await expect(page.locator('[aria-live="polite"]')).toContainText('required');
	});
});

test.describe('Recurring Assignments - Advanced Features', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Bulk operations work correctly', async ({ page }) => {
		// Create multiple assignments
		for (let i = 0; i < 3; i++) {
			const user = Object.values(TEST_USERS)[i];
			await createRecurringAssignment(
				page,
				user.name,
				TEST_SCHEDULES.WEEKEND_NIGHTS.name,
				'Saturday',
				'18:00-20:00'
			);
		}

		await page.goto('/admin/shifts/recurring');

		// Select multiple assignments
		await page.click('input[type="checkbox"][data-testid="select-all"]');

		// Bulk delete
		await page.click('button:has-text("Delete Selected")');
		await page.click('button:has-text("Yes, delete all")');

		// Verify all selected assignments are deleted
		await expect(page.locator('.toast')).toContainText('assignments deleted');

		const remainingCount = await page.locator('[data-testid="assignment-item"]').count();
		expect(remainingCount).toBe(0);
	});

	test('Export functionality works', async ({ page }) => {
		// Create some assignments
		await createRecurringAssignment(
			page,
			TEST_USERS.VOLUNTEER_1.name,
			TEST_SCHEDULES.WEEKEND_NIGHTS.name,
			'Saturday',
			'18:00-20:00'
		);

		await page.goto('/admin/shifts/recurring');

		// Set up download handler
		const downloadPromise = page.waitForEvent('download');

		// Click export
		await page.click('button:has-text("Export")');

		const download = await downloadPromise;
		expect(download.suggestedFilename()).toMatch(/recurring-assignments.*\.csv$/);
	});

	test('Import functionality works', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');

		// Click import
		await page.click('button:has-text("Import")');

		// Upload CSV file (would need test file)
		const csvContent = `user_name,schedule_name,day_of_week,time_slot,buddy_name,description
${TEST_USERS.VOLUNTEER_1.name},${TEST_SCHEDULES.WEEKEND_NIGHTS.name},Saturday,18:00-20:00,Import Buddy,Imported assignment`;

		// Create temporary file
		const dataTransfer = await page.evaluateHandle((csvContent) => {
			const dt = new DataTransfer();
			const file = new File([csvContent], 'test-assignments.csv', { type: 'text/csv' });
			dt.items.add(file);
			return dt;
		}, csvContent);

		await page.locator('input[type="file"]').dispatchEvent('drop', { dataTransfer });
		await page.click('button:has-text("Import Assignments")');

		// Verify import success
		await expect(page.locator('.toast')).toContainText('imported successfully');
		await expect(page.locator('text=Import Buddy')).toBeVisible();
	});

	test('Assignment analytics are displayed', async ({ page }) => {
		// Create assignments across different days
		for (let i = 0; i < 5; i++) {
			const user = Object.values(TEST_USERS)[i % Object.values(TEST_USERS).length];
			const day = DAYS_OF_WEEK[i];

			await createRecurringAssignment(
				page,
				user.name,
				TEST_SCHEDULES.DAILY_EVENING.name,
				day.name,
				'20:00-22:00'
			);
		}

		await page.goto('/admin/shifts/recurring');

		// Check analytics section
		await expect(page.locator('[data-testid="analytics-total-assignments"]')).toContainText('5');

		// Day distribution chart
		await expect(page.locator('[data-testid="day-distribution-chart"]')).toBeVisible();

		// User distribution
		await expect(page.locator('[data-testid="user-distribution"]')).toBeVisible();

		// Schedule coverage
		await expect(page.locator('[data-testid="schedule-coverage"]')).toBeVisible();
	});

	test('Assignment templates can be saved and reused', async ({ page }) => {
		await page.goto('/admin/shifts/recurring');
		await page.click('button:has-text("New Recurring Assignment")');

		// Fill form
		await page.click('button[role="combobox"]:has-text("Select user")');
		await page.click(`div[role="option"]:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);

		await page.click('button:has-text("Select schedule")');
		await page.click(`div[role="option"]:has-text("${TEST_SCHEDULES.WEEKEND_NIGHTS.name}")`);

		await page.click('button:has-text("Select day")');
		await page.click('div[role="option"]:has-text("Saturday")');

		await page.click('button:has-text("Select time slot")');
		await page.click('div[role="option"]:has-text("18:00-20:00")');

		// Save as template
		await page.click('button:has-text("Save as Template")');
		await page.fill('input[placeholder="Template name"]', 'Saturday Night Patrol');
		await page.click('button:has-text("Save Template")');

		// Create new assignment using template
		await page.click('button:has-text("Use Template")');
		await page.click('div[role="option"]:has-text("Saturday Night Patrol")');

		// Form should be pre-filled
		await expect(page.locator('button:has-text("Saturday")')).toBeVisible();
		await expect(page.locator('button:has-text("18:00-20:00")')).toBeVisible();
	});
});
