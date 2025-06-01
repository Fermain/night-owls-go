import { test, expect } from '@playwright/test';

test.describe('Smoke Tests', () => {
	test('app loads successfully', async ({ page }) => {
		await page.goto('/');

		// Just verify the page loads - we'll add more specific checks later
		await expect(page.locator('body')).toBeVisible();

		// Check for basic content (adjust based on actual homepage content)
		await expect(page.locator('html')).toBeVisible();
	});

	test('registration form UI works', async ({ page }) => {
		// Test the registration UI instead of API calls
		await page.goto('/register');
		
		// Verify form elements are present and interactive
		const nameField = page.getByLabel('Full Name');
		const phoneField = page.locator('input[type="tel"]');
		const createButton = page.getByRole('button', { name: /create account/i });
		
		await expect(nameField).toBeVisible();
		await expect(phoneField).toBeVisible();
		await expect(createButton).toBeVisible();
		
		// Verify we can interact with the form
		await nameField.fill('Test User');
		await expect(nameField).toHaveValue('Test User');
		
		console.log('✅ Registration form UI is functional');
	});

	test('navigation works correctly', async ({ page }) => {
		// Test navigation between key pages
		await page.goto('/');
		
		// Test navigation to registration
		const becomeOwlLink = page.getByRole('link', { name: /become an owl/i });
		await expect(becomeOwlLink).toBeVisible();
		await becomeOwlLink.click();
		await expect(page).toHaveURL('/register');
		
		// Test navigation to login
		await page.goto('/login');
		await expect(page.locator('body')).toBeVisible();
		
		console.log('✅ Navigation between pages works correctly');
	});
});
