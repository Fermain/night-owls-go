import { test, expect } from '@playwright/test';

test('🔍 Form Validation Diagnostic - Button Enable Conditions', async ({ page }) => {
	await page.goto('/register');
	await page.waitForLoadState('networkidle');

	const nameField = page.getByLabel('Full Name');
	const phoneField = page.locator('input[type="tel"]');
	const createAccountButton = page.getByRole('button', { name: /create account/i });

	// Initial state
	console.log('🔍 Initial button enabled:', await createAccountButton.isEnabled());

	// Try filling just name
	await nameField.fill('Test User');
	console.log('🔍 After name fill, button enabled:', await createAccountButton.isEnabled());

	// Try different phone formats to see which enables the button
	const phoneFormats = [
		'+27821234567',
		'0821234567',
		'82 123 4567',
		'071 123 4567',
		'+27 82 123 4567',
		'27821234567'
	];

	for (const phoneFormat of phoneFormats) {
		await phoneField.clear();
		await phoneField.fill(phoneFormat);

		// Wait for validation to complete by checking button state
		try {
			await expect(createAccountButton).toBeEnabled({ timeout: 1000 });
			console.log(`✅ SUCCESS: Button enabled with format: "${phoneFormat}"`);
			break;
		} catch {
			console.log(`🔍 Phone format "${phoneFormat}" - Button still disabled`);
		}
	}

	// Check if there are any visible validation errors
	const errorElements = await page
		.locator('[role="alert"], .error, .text-red-500, .text-destructive')
		.allTextContents();
	console.log('🔍 Validation errors:', errorElements);

	// Check form validity using browser API
	const formValidity = await page.evaluate(() => {
		const nameInput = document.querySelector('input[type="text"]') as HTMLInputElement;
		const phoneInput = document.querySelector('input[type="tel"]') as HTMLInputElement;

		return {
			nameValid: nameInput?.validity.valid,
			nameValue: nameInput?.value,
			phoneValid: phoneInput?.validity.valid,
			phoneValue: phoneInput?.value,
			phoneValidationMessage: phoneInput?.validationMessage
		};
	});

	console.log('🔍 Form validity:', formValidity);

	// Take screenshot for debugging
	await page.screenshot({ path: 'test-results/form-validation-diagnostic.png', fullPage: true });

	await expect(page.locator('body')).toBeVisible();
});
