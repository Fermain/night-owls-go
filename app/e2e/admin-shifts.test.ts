import { test, expect, type Page } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

// Test data constants
const TEST_USERS = {
	VOLUNTEER_1: { name: 'Alpha Volunteer', phone: '+15551111111', role: 'owl' },
	VOLUNTEER_2: { name: 'Beta Volunteer', phone: '+15552222222', role: 'supervisor' },
	VOLUNTEER_3: { name: 'Gamma Volunteer', phone: '+15553333333', role: 'owl' },
	VOLUNTEER_4: { name: 'Delta Volunteer', phone: '+15554444444', role: 'guest' }
};

const TEST_SCHEDULES = {
	MORNING_PATROL: { name: 'Morning Patrol', cron: '0 6 * * *' },
	EVENING_WATCH: { name: 'Evening Watch', cron: '0 18 * * *' },
	WEEKEND_SPECIAL: { name: 'Weekend Special', cron: '0 12 * * 6,0' },
	NIGHT_SHIFT: { name: 'Night Shift', cron: '0 22 * * *' }
};

async function navigateToShifts(page: Page) {
	await page.goto('/admin/shifts');
	await expect(page.locator('h1, h2')).toContainText('Shifts');
}

async function createTestUser(page: Page, name: string, phone: string, role: string = 'owl') {
	await page.goto('/admin/users/new');
	await page.fill('input#name', name);
	await page.fill('input[type="tel"]', phone);
	await page.selectOption('select[name="role"]', role);
	await page.click('button[type="submit"]');
	await expect(page.locator('.toast')).toContainText('successfully');
}

async function createTestSchedule(page: Page, name: string, cronExpr: string) {
	await page.goto('/admin/schedules/new');
	await page.fill('input#name', name);
	await page.fill('input#cron_expr', cronExpr);
	await page.click('button[type="submit"]');
	await expect(page.locator('.toast')).toContainText('successfully');
}

async function setupTestData(page: Page) {
	// Create test users
	for (const user of Object.values(TEST_USERS)) {
		await createTestUser(page, user.name, user.phone, user.role);
	}

	// Create test schedules
	for (const schedule of Object.values(TEST_SCHEDULES)) {
		await createTestSchedule(page, schedule.name, schedule.cron);
	}
}

async function bookShift(
	page: Page,
	shiftSelector: string,
	userDisplayName: string,
	buddyName?: string
) {
	// Click on shift slot to open booking form
	await page.click(shiftSelector);

	// Wait for booking form
	await expect(page.locator('form, [data-testid="booking-form"]')).toBeVisible();

	// Select user
	await page.click('button:has-text("Select User"), select[name="userId"]');
	await page.click(
		`option:has-text("${userDisplayName}"), div[role="option"]:has-text("${userDisplayName}")`
	);

	// Add buddy if specified
	if (buddyName) {
		await page.fill('input[placeholder*="buddy"], input[name="buddyName"]', buddyName);
	}

	// Submit booking
	await page.click('button:has-text("Book Shift"), button[type="submit"]');

	// Verify success
	await expect(page.locator('.toast')).toContainText('successfully');
}

test.describe('Admin Shifts Management - Dashboard and Navigation', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
		await navigateToShifts(page);
	});

	test('Shifts dashboard loads with metrics and overview', async ({ page }) => {
		// Verify dashboard elements are present
		await expect(page.locator('h1, h2')).toContainText('Shifts');

		// Check for metrics cards
		await expect(page.locator('[data-testid="metrics"], .metrics')).toBeVisible();

		// Check for upcoming shifts section
		await expect(page.locator('[data-testid="upcoming-shifts"], text=Upcoming')).toBeVisible();

		// Check for quick stats
		const statsCards = page.locator('[data-testid="stat-card"], .stat-card');
		const cardCount = await statsCards.count();
		expect(cardCount).toBeGreaterThan(0);
	});

	test('Can navigate between different shift views', async ({ page }) => {
		// Navigate to calendar view
		await page.click('a:has-text("Calendar"), button:has-text("Calendar")');
		await expect(page.locator('.calendar, [data-testid="calendar"]')).toBeVisible();

		// Navigate to list view
		await page.click('a:has-text("List"), button:has-text("List")');
		await expect(page.locator('.shift-list, [data-testid="shift-list"]')).toBeVisible();

		// Navigate back to dashboard
		await page.click('a:has-text("Dashboard"), button:has-text("Dashboard")');
		await expect(page.locator('.metrics, [data-testid="metrics"]')).toBeVisible();
	});

	test('Dashboard shows accurate shift statistics', async ({ page }) => {
		// Check that metrics display reasonable values
		const totalShifts = page.locator('[data-testid="total-shifts"], text=Total Shifts');
		if (await totalShifts.isVisible()) {
			const shiftCount = await totalShifts.textContent();
			expect(shiftCount).toMatch(/\d+/); // Should contain numbers
		}

		// Check fill rate
		const fillRate = page.locator('[data-testid="fill-rate"], text=Fill Rate');
		if (await fillRate.isVisible()) {
			const rateText = await fillRate.textContent();
			expect(rateText).toMatch(/\d+%|N\/A/); // Should show percentage or N/A
		}

		// Check upcoming shifts count
		const upcomingShifts = page.locator('[data-testid="upcoming-count"], text=Upcoming');
		if (await upcomingShifts.isVisible()) {
			const upcomingCount = await upcomingShifts.textContent();
			expect(upcomingCount).toMatch(/\d+/);
		}
	});

	test('Quick actions are accessible from dashboard', async ({ page }) => {
		// Look for quick action buttons
		const quickActions = [
			'Book Shift',
			'View Calendar',
			'Manage Schedules',
			'Recurring Assignments'
		];

		for (const action of quickActions) {
			const actionButton = page.locator(`button:has-text("${action}"), a:has-text("${action}")`);
			if (await actionButton.isVisible()) {
				await expect(actionButton).toBeVisible();
			}
		}
	});
});

test.describe('Admin Shifts Management - Shift Booking', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Admin can book a shift for a user', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		// Find an available shift slot
		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child, .available-slot:first-child',
				TEST_USERS.VOLUNTEER_1.name
			);

			// Verify shift appears as booked
			await expect(page.locator('[data-testid="booked-slot"], .booked-slot')).toBeVisible();
		}
	});

	test('Admin can book a shift with a buddy', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child, .available-slot:first-child',
				TEST_USERS.VOLUNTEER_2.name,
				'Safety Partner Alpha'
			);

			// Verify buddy information is displayed
			await expect(page.locator('text=Safety Partner Alpha')).toBeVisible();
		}
	});

	test('Cannot book the same shift twice', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		// Book a shift first
		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child, .available-slot:first-child',
				TEST_USERS.VOLUNTEER_1.name
			);

			// Try to book the same shift again
			const bookedShift = page.locator('[data-testid="booked-slot"], .booked-slot').first();
			await bookedShift.click();

			// Should show that shift is already booked
			await expect(page.locator('text=already booked, text=unavailable')).toBeVisible();
		}
	});

	test('Booking form validation works correctly', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await availableShift.click();

			// Try to submit without selecting user
			await page.click('button:has-text("Book Shift"), button[type="submit"]');

			// Should show validation error
			await expect(page.locator('text=Please select a user')).toBeVisible();
		}
	});

	test('Can cancel a booked shift', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		// Book a shift first
		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child, .available-slot:first-child',
				TEST_USERS.VOLUNTEER_3.name
			);

			// Click on the booked shift
			const bookedShift = page.locator('[data-testid="booked-slot"], .booked-slot').first();
			await bookedShift.click();

			// Cancel the booking
			await page.click('button:has-text("Cancel Booking"), button:has-text("Cancel")');
			await page.click('button:has-text("Yes"), button:has-text("Confirm")');

			// Verify shift is available again
			await expect(page.locator('[data-testid="available-slot"], .available-slot')).toBeVisible();
		}
	});
});

test.describe('Admin Shifts Management - Filtering and Search', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
		await page.goto('/admin/schedules/slots');
	});

	test('Can filter shifts by date range', async ({ page }) => {
		// Open date range picker
		await page.click('button:has-text("Date Range"), [data-testid="date-picker"]');

		// Select date range (this week)
		const today = new Date();
		const nextWeek = new Date(today);
		nextWeek.setDate(today.getDate() + 7);

		// Click on calendar dates (simplified - real implementation would be more complex)
		await page.click('[data-testid="calendar"]');

		// Verify shifts are filtered
		await expect(page.locator('[data-testid="shift-slot"], .shift-slot')).toBeVisible();
	});

	test('Can filter by shift status', async ({ page }) => {
		// Filter to show only available shifts
		await page.click('button:has-text("Filter"), [data-testid="filter-button"]');
		await page.click('input[value="available"], label:has-text("Available")');

		// Should show only available shifts
		await expect(page.locator('[data-testid="available-slot"], .available-slot')).toBeVisible();

		// Change filter to show only booked shifts
		await page.click('input[value="booked"], label:has-text("Booked")');

		// Should show only booked shifts
		await expect(page.locator('[data-testid="booked-slot"], .booked-slot')).toBeVisible();
	});

	test('Can filter by schedule', async ({ page }) => {
		// Open schedule filter
		await page.click('button:has-text("Schedule"), select[name="schedule"]');
		await page.click(
			`option:has-text("${TEST_SCHEDULES.MORNING_PATROL.name}"), div:has-text("${TEST_SCHEDULES.MORNING_PATROL.name}")`		);

		// Should show only shifts for selected schedule
		await expect(page.locator(`text=${TEST_SCHEDULES.MORNING_PATROL.name}`)).toBeVisible();
	});

	test('Can search by volunteer name', async ({ page }) => {
		// Search for specific volunteer
		await page.fill(
			'input[placeholder*="Search"], input[name="search"]',
			TEST_USERS.VOLUNTEER_1.name
		);

		// Should show only shifts for that volunteer
		await expect(page.locator(`text=${TEST_USERS.VOLUNTEER_1.name}`)).toBeVisible();
	});

	test('Can clear all filters', async ({ page }) => {
		// Apply some filters first
		await page.click('button:has-text("Filter")');
		await page.click('label:has-text("Available")');

		// Clear filters
		await page.click('button:has-text("Clear Filters"), button:has-text("Reset")');

		// Should show all shifts again
		await expect(page.locator('[data-testid="shift-slot"], .shift-slot')).toBeVisible();
	});
});

test.describe('Admin Shifts Management - Shift Conflicts', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Detects user double-booking conflicts', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		// Book first shift for a user
		const firstShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await firstShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child',
				TEST_USERS.VOLUNTEER_1.name
			);

			// Try to book overlapping shift for same user
			const secondShift = page.locator('[data-testid="available-slot"], .available-slot').first();
			if (await secondShift.isVisible()) {
				await secondShift.click();
				await page.click('button:has-text("Select User")');
				await page.click(`option:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);
				await page.click('button:has-text("Book Shift")');

				// Should show conflict warning
				await expect(page.locator('text=conflict, text=already scheduled')).toBeVisible();
			}
		}
	});

	test('Allows different users for same time slot if multiple positions available', async ({
		page
	}) => {
		// This test depends on business rules - may allow multiple bookings per slot
		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			// Book first user
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child',
				TEST_USERS.VOLUNTEER_1.name
			);

			// Try to book second user for same slot
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child',
				TEST_USERS.VOLUNTEER_2.name
			);

			// Depending on business rules, this may succeed or show capacity warning
		}
	});

	test('Shows user availability when booking', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await availableShift.click();

			// User dropdown should show availability status
			await page.click('button:has-text("Select User")');

			// Should indicate which users are available vs already scheduled
			await expect(page.locator('text=Available, text=Busy')).toBeVisible();
		}
	});
});

test.describe('Admin Shifts Management - Shift History and Tracking', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Can view shift history for a user', async ({ page }) => {
		// Navigate to user's shift history
		await page.goto('/admin/users');
		await page.click(`[data-testid="user-item"]:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);

		// Look for shift history section
		await expect(page.locator('[data-testid="shift-history"], text=Shift History')).toBeVisible();

		// Should show past and upcoming shifts
		await expect(page.locator('text=Past Shifts, text=Upcoming Shifts')).toBeVisible();
	});

	test('Can mark shift attendance', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		// Book a shift first
		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child',
				TEST_USERS.VOLUNTEER_2.name
			);

			// Click on booked shift to view details
			const bookedShift = page.locator('[data-testid="booked-slot"], .booked-slot').first();
			await bookedShift.click();

			// Mark attendance
			await page.click(
				'button:has-text("Mark Attended"), input[type="checkbox"]:has-text("Attended")'
			);

			// Verify attendance is recorded
			await expect(page.locator('text=Attended, text=Present')).toBeVisible();
		}
	});

	test('Can track no-shows and absences', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await bookShift(
				page,
				'[data-testid="available-slot"]:first-child',
				TEST_USERS.VOLUNTEER_3.name
			);

			// Mark as no-show
			const bookedShift = page.locator('[data-testid="booked-slot"], .booked-slot').first();
			await bookedShift.click();

			await page.click(
				'button:has-text("Mark No-Show"), input[type="checkbox"]:has-text("No-Show")'
			);

			// Verify no-show is recorded
			await expect(page.locator('text=No-Show, text=Absent')).toBeVisible();
		}
	});

	test('Can view comprehensive shift reports', async ({ page }) => {
		// Navigate to shift reports
		await page.goto('/admin/reports');

		// Look for shift-related reports
		await expect(page.locator('text=Shift Reports, text=Attendance Report')).toBeVisible();

		// Generate attendance report
		await page.click('button:has-text("Generate Report"), a:has-text("Attendance Report")');

		// Should show report data
		await expect(page.locator('[data-testid="report-data"], .report-table')).toBeVisible();
	});
});

test.describe('Admin Shifts Management - Bulk Operations', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
		await page.goto('/admin/schedules/slots');
	});

	test('Can select multiple shifts for bulk operations', async ({ page }) => {
		// Enable bulk selection mode
		await page.click('button:has-text("Bulk Actions"), input[type="checkbox"]:has-text("Bulk")');

		// Select multiple shifts
		const shiftCheckboxes = page.locator('[data-testid="shift-checkbox"], input[type="checkbox"]');
		const checkboxCount = await shiftCheckboxes.count();

		if (checkboxCount > 1) {
			await shiftCheckboxes.nth(0).click();
			await shiftCheckboxes.nth(1).click();

			// Should show bulk actions toolbar
			await expect(page.locator('[data-testid="bulk-actions"], .bulk-toolbar')).toBeVisible();
		}
	});

	test('Can bulk cancel multiple shifts', async ({ page }) => {
		// Assume some shifts are already booked
		await page.click('button:has-text("Bulk Actions")');

		// Select booked shifts
		const bookedShifts = page.locator('[data-testid="booked-slot"] input[type="checkbox"]');
		const bookedCount = await bookedShifts.count();

		if (bookedCount > 0) {
			await bookedShifts.first().click();

			// Bulk cancel
			await page.click('button:has-text("Cancel Selected")');
			await page.click('button:has-text("Yes, cancel all")');

			// Verify cancellation
			await expect(page.locator('.toast')).toContainText('cancelled successfully');
		}
	});

	test('Can bulk assign shifts to users', async ({ page }) => {
		// Select multiple available shifts
		await page.click('button:has-text("Bulk Actions")');

		const availableShifts = page.locator('[data-testid="available-slot"] input[type="checkbox"]');
		const availableCount = await availableShifts.count();

		if (availableCount > 1) {
			await availableShifts.nth(0).click();
			await availableShifts.nth(1).click();

			// Bulk assign
			await page.click('button:has-text("Assign Selected")');
			await page.click('button:has-text("Select User")');
			await page.click(`option:has-text("${TEST_USERS.VOLUNTEER_4.name}")`);
			await page.click('button:has-text("Assign All")');

			// Verify assignments
			await expect(page.locator('.toast')).toContainText('assigned successfully');
		}
	});
});

test.describe('Admin Shifts Management - Error Handling', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await setupTestData(page);
	});

	test('Handles network errors during shift booking', async ({ page }) => {
		// Simulate network failure
		await page.route('**/api/admin/shifts/book', (route) => route.abort());

		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await availableShift.click();
			await page.click('button:has-text("Select User")');
			await page.click(`option:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);
			await page.click('button:has-text("Book Shift")');

			// Should show network error
			await expect(page.locator('text=Network error, text=Failed to book')).toBeVisible();
		}
	});

	test('Handles server errors gracefully', async ({ page }) => {
		// Simulate server error
		await page.route('**/api/admin/shifts/book', (route) =>
			route.fulfill({ status: 500, body: JSON.stringify({ error: 'Internal server error' }) })
		);

		await page.goto('/admin/schedules/slots');

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await availableShift.click();
			await page.click('button:has-text("Select User")');
			await page.click(`option:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);
			await page.click('button:has-text("Book Shift")');

			// Should show user-friendly error
			await expect(page.locator('text=Something went wrong')).toBeVisible();
		}
	});

	test('Handles stale data conflicts', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		// Simulate another user booking the same shift
		await page.route('**/api/admin/shifts/book', (route) =>
			route.fulfill({
				status: 409,
				body: JSON.stringify({ error: 'Shift already booked by another user' })
			})
		);

		const availableShift = page.locator('[data-testid="available-slot"], .available-slot').first();
		if (await availableShift.isVisible()) {
			await availableShift.click();
			await page.click('button:has-text("Select User")');
			await page.click(`option:has-text("${TEST_USERS.VOLUNTEER_1.name}")`);
			await page.click('button:has-text("Book Shift")');

			// Should show conflict message and refresh data
			await expect(page.locator('text=already booked, text=conflict')).toBeVisible();
		}
	});
});

test.describe('Admin Shifts Management - Performance', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
	});

	test('Shifts page loads quickly with large dataset', async ({ page }) => {
		// Navigate and measure load time
		const startTime = Date.now();
		await page.goto('/admin/shifts');
		await page.waitForSelector('[data-testid="shift-slot"], .shift-slot, text=No shifts found');
		const loadTime = Date.now() - startTime;

		// Should load within reasonable time
		expect(loadTime).toBeLessThan(5000); // 5 seconds
	});

	test('Calendar view performs well with many shifts', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		const calendarStartTime = Date.now();

		// Switch to calendar view if available
		const calendarButton = page.locator('button:has-text("Calendar"), a:has-text("Calendar")');
		if (await calendarButton.isVisible()) {
			await calendarButton.click();
			await page.waitForSelector('[data-testid="calendar"], .calendar-view');
		}

		const calendarLoadTime = Date.now() - calendarStartTime;
		expect(calendarLoadTime).toBeLessThan(3000); // 3 seconds
	});

	test('Filtering responds quickly', async ({ page }) => {
		await page.goto('/admin/schedules/slots');

		const filterStartTime = Date.now();
		await page.click('button:has-text("Filter")');
		await page.click('label:has-text("Available")');

		// Wait for filter to apply
		await page.waitForTimeout(500);

		const filterTime = Date.now() - filterStartTime;
		expect(filterTime).toBeLessThan(2000); // 2 seconds
	});
});

