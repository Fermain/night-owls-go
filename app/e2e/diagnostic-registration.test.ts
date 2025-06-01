import { test, expect } from '@playwright/test';

test('🔍 Registration Page Diagnostic - Form Fields Analysis', async ({ page }) => {
	await page.goto('/register');
	
	// Wait for page to fully load
	await page.waitForLoadState('networkidle');
	
	console.log('🔍 Registration page loaded');
	
	// Get all form labels
	const labels = await page.locator('label').allTextContents();
	console.log('🔍 All Labels:', labels);
	
	// Get all input placeholders
	const placeholders = await page.locator('input').evaluateAll(inputs => 
		inputs.map(input => (input as HTMLInputElement).placeholder || '')
	);
	console.log('🔍 All Placeholders:', placeholders);
	
	// Get all button texts
	const buttons = await page.locator('button').allTextContents();
	console.log('🔍 All Buttons:', buttons);
	
	// Check specifically for phone-related elements
	const phoneElements = await page.locator('*:has-text("phone")').allTextContents();
	console.log('🔍 Phone-related elements:', phoneElements);
	
	// Check specifically for name-related elements  
	const nameElements = await page.locator('*:has-text("name")').allTextContents();
	console.log('🔍 Name-related elements:', nameElements);
	
	// Try different phone field selectors
	const phoneInput1 = page.getByLabel('Phone Number');
	const phoneInput2 = page.getByLabel(/phone/i);
	const phoneInput3 = page.getByPlaceholder(/phone/i);
	const phoneInput4 = page.locator('input[type="tel"]');
	
	console.log('🔍 Phone Number (exact):', await phoneInput1.isVisible());
	console.log('🔍 Phone (regex):', await phoneInput2.isVisible());
	console.log('🔍 Phone (placeholder):', await phoneInput3.isVisible());
	console.log('🔍 Phone (tel type):', await phoneInput4.isVisible());
	
	// Try different name field selectors
	const nameInput1 = page.getByLabel('Full Name');
	const nameInput2 = page.getByLabel(/name/i);
	const nameInput3 = page.getByPlaceholder(/name/i);
	
	console.log('🔍 Full Name (exact):', await nameInput1.isVisible());
	console.log('🔍 Name (regex):', await nameInput2.isVisible());
	console.log('🔍 Name (placeholder):', await nameInput3.isVisible());
	
	// Take screenshot for debugging
	await page.screenshot({ path: 'test-results/registration-diagnostic.png', fullPage: true });
	
	// Basic assertion to make test pass/fail
	await expect(page.locator('body')).toBeVisible();
}); 