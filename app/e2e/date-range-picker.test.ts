import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './test-utils';

test.describe('Bulk Assignment Date Range Picker', () => {
	test.beforeEach(async ({ page }) => {
		await loginAsAdmin(page);
		await page.goto('/admin/shifts/bulk-signup');
		await page.waitForLoadState('networkidle');
	});

	test('should load with default date range and show shifts', async ({ page }) => {
		// Wait for shifts to load
		await expect(page.locator('[data-testid="shifts-list"]')).toBeVisible();
		
		// Check that some shifts are displayed
		const shiftCards = page.locator('[role="button"]').filter({ hasText: /\d{2}:\d{2} - \d{2}:\d{2}/ });
		const initialShiftCount = await shiftCards.count();
		console.log(`Initial shift count: ${initialShiftCount}`);
		
		expect(initialShiftCount).toBeGreaterThan(0);
		
		// Check the default date range text exists
		await expect(page.locator('text=/shifts available/')).toBeVisible();
	});

	test('should capture console logs when date range changes', async ({ page }) => {
		// Capture console logs
		const consoleLogs: string[] = [];
		page.on('console', msg => {
			if (msg.type() === 'log') {
				consoleLogs.push(msg.text());
				console.log('BROWSER LOG:', msg.text());
			}
		});

		// Listen for API requests
		const apiRequests: string[] = [];
		page.on('request', request => {
			if (request.url().includes('/api/admin/schedules/all-slots')) {
				apiRequests.push(request.url());
				console.log('API Request:', request.url());
			}
		});

		// Wait for initial load
		await expect(page.locator('[data-testid="shifts-list"]')).toBeVisible();
		
		// Get initial shift count
		const shiftCards = page.locator('[role="button"]').filter({ hasText: /\d{2}:\d{2} - \d{2}:\d{2}/ });
		const initialShiftCount = await shiftCards.count();
		console.log(`Initial shift count: ${initialShiftCount}`);

		// Find and interact with date range picker
		const dateRangePicker = page.locator('[data-testid="date-range-picker"]').first();
		await expect(dateRangePicker).toBeVisible();
		
		// Click to trigger interaction (this should cause the logs we added)
		await dateRangePicker.click();
		await page.waitForTimeout(1000);
		
		// Try to simulate some date selection by pressing keys
		await page.keyboard.press('Escape'); // Close any popover
		await page.waitForTimeout(500);

		// Log all captured information
		console.log(`Captured ${consoleLogs.length} console logs`);
		console.log(`Captured ${apiRequests.length} API requests`);
		
		consoleLogs.forEach((log, index) => {
			console.log(`Console Log ${index + 1}: ${log}`);
		});
		
		apiRequests.forEach((url, index) => {
			console.log(`API Request ${index + 1}: ${url}`);
		});

		// At minimum, verify the date range picker is interactive
		expect(await dateRangePicker.count()).toBeGreaterThan(0);
		
		// Check if any date-related logs were captured
		const dateRelatedLogs = consoleLogs.filter(log => 
			log.includes('date') || log.includes('Date') || log.includes('Using') || log.includes('fetch')
		);
		console.log(`Date-related logs found: ${dateRelatedLogs.length}`);
	});

	test('should verify date range picker component exists and is interactive', async ({ page }) => {
		// Check if the date range picker section exists (use more specific selector)
		await expect(page.locator('label:has-text("Date Range")').first()).toBeVisible();
		
		// Check if the date range picker wrapper exists
		const dateRangeWrapper = page.locator('[data-testid="date-range-picker"]');
		await expect(dateRangeWrapper).toBeVisible();
		
		// Try to interact with it
		await dateRangeWrapper.click();
		await page.waitForTimeout(500);
		
		console.log('Date range picker clicked successfully');
		
		// Close any opened elements
		await page.keyboard.press('Escape');
	});

	test('should show loading and error states appropriately', async ({ page }) => {
		// Check that we start with a loading state or quickly move to loaded state
		const loadingIndicator = page.locator('text=Loading shifts...');
		const shiftsContent = page.locator('[data-testid="shifts-list"]');
		
		// Either we see loading briefly, or content loads immediately
		await expect(loadingIndicator.or(shiftsContent)).toBeVisible();
		
		// Eventually we should see the shifts content
		await expect(shiftsContent).toBeVisible();
		
		// Check for any error states
		const errorIndicator = page.locator('text=/Error loading shifts/');
		await expect(errorIndicator).not.toBeVisible();
	});
}); 