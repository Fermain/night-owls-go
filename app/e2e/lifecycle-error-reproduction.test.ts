import { test, expect } from '@playwright/test';

test.describe('Lifecycle Error Reproduction', () => {
	test('login page should render without lifecycle errors', async ({ page }) => {
		// Track JavaScript errors
		const jsErrors: string[] = [];
		page.on('pageerror', (error) => {
			jsErrors.push(error.message);
			console.log('JavaScript Error:', error.message);
		});

		// Track console errors  
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				console.log('Console Error:', msg.text());
				jsErrors.push(msg.text());
			}
		});

		// Load the login page
		console.log('🧪 Loading login page...');
		await page.goto('/login');

		// Wait a moment for JavaScript to execute
		await page.waitForTimeout(2000);

		// Check if page rendered at all
		console.log('🔍 Checking if page rendered...');
		await expect(page.locator('body')).toBeVisible();

		// Look for any text content to verify the page actually loaded
		const hasAnyText = await page.locator('body').innerText();
		console.log('📄 Page content length:', hasAnyText.length);
		
		// Check specifically for lifecycle errors
		const hasLifecycleError = jsErrors.some(error => 
			error.includes('lifecycle_outside_component') || 
			error.includes('BWToSyqr.js')
		);

		if (hasLifecycleError) {
			console.log('❌ LIFECYCLE ERROR FOUND:', jsErrors);
			throw new Error(`Lifecycle error detected: ${jsErrors.join(', ')}`);
		}

		// Verify the page actually has content (not just empty body)
		expect(hasAnyText.length).toBeGreaterThan(0);
		
		// Look for expected login page elements
		console.log('🔍 Looking for login form elements...');
		
		// We should see some kind of login interface
		const possibleSelectors = [
			'input[type="tel"]',
			'input[type="email"]', 
			'input[type="text"]',
			'button',
			'form',
			'[role="button"]'
		];

		let foundElements = 0;
		for (const selector of possibleSelectors) {
			const elements = await page.locator(selector).count();
			if (elements > 0) {
				foundElements++;
				console.log(`✅ Found ${elements} ${selector} elements`);
			}
		}

		// If no interactive elements found, the page likely crashed
		if (foundElements === 0) {
			throw new Error('No interactive elements found - page may have crashed due to lifecycle error');
		}

		console.log('✅ Login page loaded successfully without lifecycle errors');
	});

	test('home page should render without lifecycle errors', async ({ page }) => {
		// Track JavaScript errors
		const jsErrors: string[] = [];
		page.on('pageerror', (error) => {
			jsErrors.push(error.message);
		});

		await page.goto('/');
		await page.waitForTimeout(2000);

		// Check for lifecycle errors
		const hasLifecycleError = jsErrors.some(error => 
			error.includes('lifecycle_outside_component')
		);

		expect(hasLifecycleError).toBe(false);
		await expect(page.locator('body')).toBeVisible();
		
		console.log('✅ Home page loaded successfully');
	});

	test('register page should render without lifecycle errors', async ({ page }) => {
		// Track JavaScript errors
		const jsErrors: string[] = [];
		page.on('pageerror', (error) => {
			jsErrors.push(error.message);
		});

		await page.goto('/register');
		await page.waitForTimeout(2000);

		// Check for lifecycle errors
		const hasLifecycleError = jsErrors.some(error => 
			error.includes('lifecycle_outside_component')
		);

		expect(hasLifecycleError).toBe(false);
		await expect(page.locator('body')).toBeVisible();
		
		console.log('✅ Register page loaded successfully');
	});
}); 