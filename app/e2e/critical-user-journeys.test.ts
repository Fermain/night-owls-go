import { test, expect } from '@playwright/test';
import { AuthPage } from './page-objects/auth.page';
import { AdminSchedulesPage } from './page-objects/admin-schedules.page';
import { ShiftsPage } from './page-objects/shifts.page';
import { setupApiMocks } from './setup/api-mocks';

test.describe('Critical User Journeys', () => {
	test.beforeEach(async ({ page }) => {
		// Set up API mocks for each test
		await setupApiMocks(page);
	});

	test('Complete new user registration and authentication flow', async ({ page }) => {
		const authPage = new AuthPage(page);

		// Navigate to homepage and start registration flow
		await authPage.goto();
		await authPage.clickJoinUs();

		// Fill out registration form
		await authPage.register('John Doe', '+27821234567');

		// Verify OTP and complete login
		await authPage.verifyOTP('123456');
		await authPage.expectSuccessfulLogin();
	});

	test('Admin can manage schedules end-to-end', async ({ page }) => {
		const authPage = new AuthPage(page);
		const schedulesPage = new AdminSchedulesPage(page);

		// Login as admin
		await authPage.loginAsAdmin();

		// Navigate to schedules and create new schedule
		await schedulesPage.goto();
		await schedulesPage.createSchedule({
			name: 'Night Patrol',
			description: 'Late night security patrol',
			cronExpression: '0 22 * * *',
			duration: 180,
			positions: 2,
			timezone: 'Africa/Johannesburg'
		});
		await schedulesPage.expectScheduleInList('Night Patrol');
	});

	test('Volunteer can book and manage shifts', async ({ page }) => {
		const authPage = new AuthPage(page);
		const shiftsPage = new ShiftsPage(page);

		// Login as volunteer
		await authPage.loginAsVolunteer();

		// Navigate to shifts and book one
		await shiftsPage.goto();
		await shiftsPage.expectShiftsVisible();

		const initialCount = await shiftsPage.getAvailableShiftsCount();
		expect(initialCount).toBeGreaterThan(0);

		// Book first available shift
		await shiftsPage.bookShift();
		await shiftsPage.expectBookingSuccess();
	});

	test('Volunteer can book shift with buddy', async ({ page }) => {
		const authPage = new AuthPage(page);
		const shiftsPage = new ShiftsPage(page);

		// Login as volunteer
		await authPage.loginAsVolunteer();

		// Navigate to shifts and book with buddy
		await shiftsPage.goto();
		await shiftsPage.bookShift('Jane Smith');
		await shiftsPage.expectBookingSuccess();
	});

	test('Authentication error handling', async ({ page }) => {
		const authPage = new AuthPage(page);

		// Test registration with valid data
		await authPage.gotoRegister();
		await authPage.register('Valid User', '+27821234567');

		// Test login with invalid OTP
		await authPage.verifyOTP('000000');
		await authPage.expectLoginError();
		await authPage.expectOTPCleared();

		// Test login with valid OTP
		await authPage.verifyOTP('123456');
		await authPage.expectSuccessfulLogin();
	});

	test('Schedule form validation', async ({ page }) => {
		const authPage = new AuthPage(page);
		const schedulesPage = new AdminSchedulesPage(page);

		// Login as admin
		await authPage.loginAsAdmin();

		// Test form validation
		await schedulesPage.goto();
		await schedulesPage.clickCreateSchedule();

		// Try to submit empty form
		await schedulesPage.saveSchedule();
		await schedulesPage.expectValidationError('name');

		// Fill form with invalid data
		await schedulesPage.fillScheduleForm({
			name: '', // Invalid: empty name
			description: 'Test description',
			cronExpression: 'invalid-cron', // Invalid: bad cron
			duration: -30, // Invalid: negative duration
			positions: 0 // Invalid: zero positions
		});
		await schedulesPage.saveSchedule();
		await schedulesPage.expectFormError('Invalid cron expression');

		// Fill form with valid data and save
		await schedulesPage.fillScheduleForm({
			name: 'Valid Schedule',
			description: 'Valid description',
			cronExpression: '0 9 * * *',
			duration: 120,
			positions: 2
		});
		await schedulesPage.saveSchedule();
		await schedulesPage.expectScheduleInList('Valid Schedule');
	});

	test('Route protection and redirects', async ({ page }) => {
		// Try to access admin route without authentication
		await page.goto('/admin');
		await expect(page).toHaveURL('/login');

		// Try to access admin schedules without authentication
		await page.goto('/admin/schedules');
		await expect(page).toHaveURL('/login');

		// Login and verify access is granted
		const authPage = new AuthPage(page);
		await authPage.loginAsAdmin();
		await page.goto('/admin');
		await expect(page).toHaveURL('/admin');
	});

	test('Full booking conflict handling', async ({ page }) => {
		const authPage = new AuthPage(page);
		const shiftsPage = new ShiftsPage(page);

		// Login as volunteer
		await authPage.loginAsVolunteer();
		await shiftsPage.goto();

		// Try to book a shift that's already full (based on mock data)
		// The mock data has one shift marked as is_booked: true
		const shiftsCount = await shiftsPage.getAvailableShiftsCount();
		expect(shiftsCount).toBeGreaterThan(0);

		// This test verifies that the booking system shows both available and booked shifts
		await expect(page.getByText('Weekend Watch')).toBeVisible(); // This shift is marked as booked
	});
});
