import { test, expect } from '@playwright/test';

test('🔍 Deep Form Validation - Step by Step Analysis', async ({ page }) => {
	await page.goto('/register');
	await page.waitForLoadState('networkidle');
	
	const nameField = page.getByLabel('Full Name');
	const phoneField = page.locator('input[type="tel"]');
	const createAccountButton = page.getByRole('button', { name: /create account/i });
	
	console.log('🔍 === INITIAL STATE ===');
	console.log('Button enabled:', await createAccountButton.isEnabled());
	console.log('Button disabled attribute:', await createAccountButton.getAttribute('disabled'));
	
	// Step 1: Fill name only
	console.log('\n🔍 === STEP 1: NAME ONLY ===');
	await nameField.fill('Test User');
	await page.waitForTimeout(1000);
	console.log('Button enabled:', await createAccountButton.isEnabled());
	
	// Step 2: Add phone in working format
	console.log('\n🔍 === STEP 2: ADD PHONE ===');
	await phoneField.fill('0821234567');
	await page.waitForTimeout(2000); // Wait longer for validation
	console.log('Button enabled:', await createAccountButton.isEnabled());
	console.log('Button disabled attribute:', await createAccountButton.getAttribute('disabled'));
	
	// Check if there are any hidden validation requirements
	console.log('\n🔍 === FORM STATE ANALYSIS ===');
	const formState = await page.evaluate(() => {
		const nameInput = document.querySelector('input[type="text"]') as HTMLInputElement;
		const phoneInput = document.querySelector('input[type="tel"]') as HTMLInputElement;
		const form = document.querySelector('form');
		const button = document.querySelector('button[type="submit"]') as HTMLButtonElement;
		
		return {
			nameValue: nameInput?.value,
			nameValid: nameInput?.validity.valid,
			phoneValue: phoneInput?.value,
			phoneValid: phoneInput?.validity.valid,
			formValid: form?.checkValidity(),
			buttonDisabled: button?.disabled,
			buttonType: button?.type,
			formNoValidate: form?.noValidate,
			requiredFields: Array.from(document.querySelectorAll('input[required]')).map(input => ({
				name: (input as HTMLInputElement).name,
				type: (input as HTMLInputElement).type,
				value: (input as HTMLInputElement).value,
				valid: (input as HTMLInputElement).validity.valid
			}))
		};
	});
	
	console.log('Form state:', JSON.stringify(formState, null, 2));
	
	// Check for any JavaScript validation or state management
	console.log('\n🔍 === JAVASCRIPT STATE ===');
	const jsState = await page.evaluate(() => {
		// Check for any global state or validation variables
		const globals = Object.keys(window).filter(key => 
			key.includes('valid') || key.includes('form') || key.includes('button') || key.includes('enable')
		);
		
		return {
			windowGlobals: globals,
			// Check for any Svelte stores or state
			svelteStores: '__svelte' in window ? 'present' : 'absent'
		};
	});
	
	console.log('JavaScript state:', JSON.stringify(jsState, null, 2));
	
	// Try triggering all possible events
	console.log('\n🔍 === EVENT TRIGGERING ===');
	await nameField.blur();
	await phoneField.blur();
	await page.waitForTimeout(1000);
	console.log('After blur events - Button enabled:', await createAccountButton.isEnabled());
	
	// Try clicking somewhere else to trigger change events
	await page.click('body');
	await page.waitForTimeout(1000);
	console.log('After body click - Button enabled:', await createAccountButton.isEnabled());
	
	// Take screenshot for final debugging
	await page.screenshot({ path: 'test-results/deep-form-diagnostic.png', fullPage: true });
	
	await expect(page.locator('body')).toBeVisible();
}); 