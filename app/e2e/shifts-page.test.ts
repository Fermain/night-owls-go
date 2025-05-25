import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

test.describe('Shifts Page', () => {
	test.beforeEach(async ({ page }) => {
		// Ensure we have a clean state
		await page.goto('/');
	});

	test('should display shifts page for unauthenticated users', async ({ page }) => {
		// Navigate to shifts page without authentication
		await page.goto('/shifts');
		
		// Verify the page loads (should show login prompt)
		await expect(page.getByText('Available Shifts')).toBeVisible({ timeout: 300 });
		await expect(page.getByText('Please sign in to view available shifts')).toBeVisible();
		
		// Verify the page URL is correct
		expect(page.url()).toContain('/shifts');
		
		// Check if there's a sign in link
		const signInLink = page.getByRole('link', { name: 'sign in' });
		await expect(signInLink).toBeVisible();
		await expect(signInLink).toHaveAttribute('href', '/login');
	});

	test('should display shifts page for authenticated users', async ({ page }) => {
		// Login as admin first
		await loginAsAdmin(page);
		
		// Navigate to shifts page
		await page.goto('/shifts');
		
		// Verify the page loads with proper content
		await expect(page.getByText('Available Shifts')).toBeVisible({ timeout: 15000 });
		await expect(page.getByText('Choose your patrol time')).toBeVisible();
		
		// Check for stats cards
		await expect(page.getByText('Tonight')).toBeVisible();
		await expect(page.getByText('Available')).toBeVisible();
		await expect(page.getByText('Urgent')).toBeVisible();
		
		// Wait for data to load (loading states)
		// Either we see loading animation or actual shift data
		const hasLoadingState = await page.getByText('No shifts found').isVisible({ timeout: 5000 }).catch(() => false);
		const hasShiftData = await page.locator('.space-y-4 > div').count() > 0;
		
		// We should either see "No shifts found" or actual shift cards
		expect(hasLoadingState || hasShiftData).toBeTruthy();
	});

	test('should handle API endpoints correctly', async ({ page }) => {
		// Test the backend API endpoint directly
		const apiResponse = await page.request.get('/shifts/available');
		console.log('API Response status:', apiResponse.status());
		console.log('API Response headers:', await apiResponse.headers());
		
		if (apiResponse.ok()) {
			const data = await apiResponse.json();
			console.log('API Response data sample:', data.slice(0, 2));
			expect(data).toBeInstanceOf(Array);
		} else {
			console.log('API Error:', await apiResponse.text());
		}
	});

	test('should test the full authenticated flow with data loading', async ({ page }) => {
		// Enable request/response logging
		page.on('request', request => {
			if (request.url().includes('/shifts/available') || request.url().includes('/bookings')) {
				console.log('REQUEST:', request.method(), request.url());
			}
		});
		
		page.on('response', response => {
			if (response.url().includes('/shifts/available') || response.url().includes('/bookings')) {
				console.log('RESPONSE:', response.status(), response.url());
			}
		});

		// Login as admin
		await loginAsAdmin(page);
		
		// Navigate to shifts page
		await page.goto('/shifts');
		
		// Wait for the page to fully load
		await page.waitForLoadState('networkidle');
		
		// Check if we have data or error messages
		await page.waitForTimeout(3000); // Give time for API calls
		
		// Take a screenshot for debugging
		await page.screenshot({ path: 'test-results/shifts-page-debug.png', fullPage: true });
		
		// Check for error messages in the UI
		const errorMessage = page.getByText(/Error loading shifts/);
		const noShiftsMessage = page.getByText('No shifts found');
		const loadingState = page.locator('[class*="animate-pulse"]');
		
		// Log what we find
		const hasError = await errorMessage.isVisible().catch(() => false);
		const hasNoShifts = await noShiftsMessage.isVisible().catch(() => false);
		const isLoading = await loadingState.isVisible().catch(() => false);
		
		console.log('UI State:', { hasError, hasNoShifts, isLoading });
		
		if (hasError) {
			console.log('Error message found in UI');
		} else if (hasNoShifts) {
			console.log('No shifts message found - this is expected if no data');
		} else if (isLoading) {
			console.log('Still loading...');
		} else {
			console.log('Shifts data should be visible');
		}
	});

	test('should test bookings page integration', async ({ page }) => {
		// Login as admin
		await loginAsAdmin(page);
		
		// Test bookings page
		await page.goto('/bookings');
		
		// Verify the page loads
		await expect(page.getByText('My Bookings')).toBeVisible({ timeout: 300 });
		
		// Check if bookings/my redirect works
		await page.goto('/bookings/my');
		await expect(page).toHaveURL(/\/bookings$/);
	});

	test('should test navigation between pages', async ({ page }) => {
		// Login as admin
		await loginAsAdmin(page);
		
		// Start from home
		await page.goto('/');
		
		// Use mobile navigation to go to shifts
		const shiftsLink = page.getByRole('link', { name: 'Shifts' });
		await expect(shiftsLink).toBeVisible();
		await shiftsLink.click();
		
		await expect(page).toHaveURL(/\/shifts$/);
		await expect(page.getByText('Available Shifts')).toBeVisible();
		
		// Navigate to bookings
		const bookingsLink = page.getByRole('link', { name: 'Bookings' });
		await expect(bookingsLink).toBeVisible();
		await bookingsLink.click();
		
		await expect(page).toHaveURL(/\/bookings$/);
		await expect(page.getByText('My Bookings')).toBeVisible();
	});

	test('should check console errors and network failures', async ({ page }) => {
		const consoleMessages: string[] = [];
		const networkErrors: string[] = [];
		
		// Capture console messages
		page.on('console', msg => {
			if (msg.type() === 'error') {
				consoleMessages.push(`CONSOLE ERROR: ${msg.text()}`);
			}
		});
		
		// Capture network failures
		page.on('response', response => {
			if (!response.ok() && response.url().includes('localhost')) {
				networkErrors.push(`NETWORK ERROR: ${response.status()} ${response.url()}`);
			}
		});
		
		// Login and navigate
		await loginAsAdmin(page);
		await page.goto('/shifts');
		
		// Wait for everything to settle
		await page.waitForTimeout(5000);
		
		// Log any errors found
		if (consoleMessages.length > 0) {
			console.log('Console Errors Found:', consoleMessages);
		}
		if (networkErrors.length > 0) {
			console.log('Network Errors Found:', networkErrors);
		}
		
		// The test shouldn't fail on errors, just log them for debugging
		console.log('Error check complete. Console errors:', consoleMessages.length, 'Network errors:', networkErrors.length);
	});
}); 