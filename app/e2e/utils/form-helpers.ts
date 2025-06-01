import { type Page } from '@playwright/test';

/**
 * Properly fill a phone input component that uses svelte-tel-input
 * This ensures the component's validation logic is triggered correctly
 */
export async function fillPhoneInput(page: Page, phoneNumber: string) {
	const phoneField = page.locator('input[type="tel"]');
	
	// Clear the field first
	await phoneField.clear();
	
	// Type character by character to trigger validation
	await phoneField.type(phoneNumber, { delay: 50 });
	
	// Trigger validation events
	await phoneField.dispatchEvent('input');
	await phoneField.dispatchEvent('change');
	await phoneField.blur();
	
	// Wait for validation to complete
	await page.waitForTimeout(500);
}

/**
 * Fill a complete registration form with proper validation
 */
export async function fillRegistrationForm(page: Page, name: string, phone: string) {
	const nameField = page.getByLabel('Full Name');
	
	// Fill name field
	await nameField.fill(name);
	await nameField.blur();
	
	// Fill phone field with proper validation
	await fillPhoneInput(page, phone);
	
	// Wait for all validation to complete
	await page.waitForTimeout(1000);
}

/**
 * Wait for a form submit button to become enabled
 */
export async function waitForSubmitButton(page: Page, buttonText: RegExp, timeout = 5000) {
	const button = page.getByRole('button', { name: buttonText });
	await button.waitFor({ state: 'visible' });
	await button.waitFor({ state: 'attached' });
	
	// Wait for the button to be enabled
	await button.waitFor({ state: 'visible', timeout });
	
	// Check if enabled with polling
	let attempts = 0;
	const maxAttempts = timeout / 200;
	
	while (attempts < maxAttempts) {
		const isEnabled = await button.isEnabled();
		if (isEnabled) {
			return button;
		}
		await page.waitForTimeout(200);
		attempts++;
	}
	
	throw new Error(`Submit button did not become enabled within ${timeout}ms`);
} 