import { test, expect } from '@playwright/test';

test('🔍 Homepage Diagnostic - See What Is Actually There', async ({ page }) => {
	await page.goto('/');
	
	// Wait for page to fully load
	await page.waitForLoadState('networkidle');
	
	// Check authentication state
	const authState = await page.evaluate(() => {
		const userSession = localStorage.getItem('user-session');
		return userSession ? JSON.parse(userSession) : null;
	});
	
	console.log('🔍 Auth State:', authState);
	
	// Get all buttons on the page
	const buttons = await page.locator('button').allTextContents();
	console.log('🔍 All Buttons:', buttons);
	
	// Get all links on the page  
	const links = await page.locator('a').allTextContents();
	console.log('🔍 All Links:', links);
	
	// Get main heading
	const headings = await page.locator('h1').allTextContents();
	console.log('🔍 Main Headings:', headings);
	
	// Try to find "Become an Owl" as a link instead of button
	const becomeOwlLink = page.getByRole('link', { name: /become an owl/i });
	const isLinkVisible = await becomeOwlLink.isVisible();
	console.log('🔍 "Become an Owl" link visible:', isLinkVisible);
	
	// Try to find "Become an Owl" as button  
	const becomeOwlButton = page.getByRole('button', { name: /become an owl/i });
	const isButtonVisible = await becomeOwlButton.isVisible();
	console.log('🔍 "Become an Owl" button visible:', isButtonVisible);
	
	// Take screenshot for debugging
	await page.screenshot({ path: 'test-results/homepage-diagnostic.png', fullPage: true });
	
	// Basic assertion to make test pass/fail
	await expect(page.locator('body')).toBeVisible();
}); 