import { test, expect } from '@playwright/test';
import { AuthPage } from './page-objects/auth.page';
import { AdminSchedulesPage } from './page-objects/admin-schedules.page';
import { ShiftsPage } from './page-objects/shifts.page';
import { testOTPs, generateUniqueTestData } from './fixtures/test-data';

test.describe('Critical User Journeys', () => {
  
  test('Complete new user registration and authentication flow', async ({ page }) => {
    const authPage = new AuthPage(page);
    const uniqueData = generateUniqueTestData();

    // Step 1: Homepage to registration
    await authPage.goto();
    await expect(page.getByRole('heading', { name: 'Protecting Our Community Together' })).toBeVisible();
    await authPage.clickJoinUs();

    // Step 2: Complete registration
    await authPage.register(uniqueData.user.name, uniqueData.user.phone);

    // Step 3: Verify and login with OTP
    await authPage.verifyOTP(testOTPs.valid);
    await authPage.expectSuccessfulLogin();
  });

  test('Admin can manage schedules end-to-end', async ({ page }) => {
    const authPage = new AuthPage(page);
    const schedulesPage = new AdminSchedulesPage(page);
    const uniqueData = generateUniqueTestData();

    // Login as admin
    await authPage.loginAsAdmin();

    // Navigate to schedules
    await schedulesPage.goto();

    // Create a new schedule
    await schedulesPage.createSchedule(uniqueData.schedule);
    await schedulesPage.expectScheduleInList(uniqueData.schedule.name);

    // Edit the schedule
    await schedulesPage.editSchedule(uniqueData.schedule.name, {
      description: 'Updated description',
      duration: 180
    });

    // Search for the schedule
    await schedulesPage.searchSchedules(uniqueData.schedule.name);
    await schedulesPage.expectScheduleInList(uniqueData.schedule.name);

    // Clean up - delete the schedule
    await schedulesPage.deleteSchedule(uniqueData.schedule.name);
    await schedulesPage.expectScheduleNotInList(uniqueData.schedule.name);
  });

  test('Volunteer can book and manage shifts', async ({ page }) => {
    const authPage = new AuthPage(page);
    const shiftsPage = new ShiftsPage(page);

    // Login as volunteer
    await authPage.loginAsVolunteer();

    // Navigate to shifts
    await shiftsPage.goto();
    await shiftsPage.expectShiftsVisible();

    // Verify available shifts are shown
    const shiftCount = await shiftsPage.getAvailableShiftsCount();
    expect(shiftCount).toBeGreaterThan(0);

    // Book a shift without buddy
    await shiftsPage.bookShift();
    await shiftsPage.expectBookingSuccess();

    // Verify booking appears in my bookings
    await shiftsPage.expectBookingInMyBookings('Evening Patrol');
  });

  test('Volunteer can book shift with buddy', async ({ page }) => {
    const authPage = new AuthPage(page);
    const shiftsPage = new ShiftsPage(page);

    // Login as volunteer
    await authPage.loginAsVolunteer();
    await shiftsPage.goto();

    // Book a shift with buddy
    await shiftsPage.bookShift(undefined, 'Jane Partner');
    await shiftsPage.expectBookingSuccess();
    await shiftsPage.expectBuddyDisplayed('Jane Partner');
  });

  test('Authentication error handling', async ({ page }) => {
    const authPage = new AuthPage(page);
    const uniqueData = generateUniqueTestData();

    // Register new user
    await authPage.gotoRegister();
    await authPage.register(uniqueData.user.name, uniqueData.user.phone);

    // Try invalid OTP
    await authPage.verifyOTP(testOTPs.invalid);
    await authPage.expectLoginError();
    await authPage.expectOTPCleared();

    // Try with valid OTP
    await authPage.verifyOTP(testOTPs.valid);
    await authPage.expectSuccessfulLogin();
  });

  test('Schedule form validation', async ({ page }) => {
    const authPage = new AuthPage(page);
    const schedulesPage = new AdminSchedulesPage(page);

    await authPage.loginAsAdmin();
    await schedulesPage.goto();
    await schedulesPage.clickCreateSchedule();

    // Try to save empty form
    await schedulesPage.saveSchedule();
    await schedulesPage.expectValidationError('name');

    // Fill with invalid data
    await schedulesPage.fillScheduleForm({
      name: 'Test Schedule',
      description: 'Test description',
      cronExpression: 'invalid-cron',
      duration: -1,
      positions: 0
    });

    await schedulesPage.saveSchedule();
    await schedulesPage.expectFormError('Invalid cron expression');
  });

  test('Authenticated users are redirected from auth pages', async ({ page }) => {
    const authPage = new AuthPage(page);

    // Login as admin
    await authPage.loginAsAdmin();

    // Try to visit auth pages - should redirect to admin
    await page.goto('/register');
    await expect(page).toHaveURL('/admin');

    await page.goto('/login');
    await expect(page).toHaveURL('/admin');

    // Logout and verify redirection works
    await authPage.logout();
    await page.goto('/admin');
    await expect(page).toHaveURL('/login');
  });

  test('Full booking conflict handling', async ({ page }) => {
    const authPage = new AuthPage(page);
    const shiftsPage = new ShiftsPage(page);

    await authPage.loginAsVolunteer();
    await shiftsPage.goto();

    // Try to book a shift that's already full (based on mock data)
    const fullShiftSelector = '[data-testid="shift-2"]'; // This shift is full in our mock data
    await page.locator(fullShiftSelector).click();
    await shiftsPage.bookShiftButton.click();
    await shiftsPage.confirmBookingButton.click();

    await shiftsPage.expectBookingError('Shift is full');
  });
}); 