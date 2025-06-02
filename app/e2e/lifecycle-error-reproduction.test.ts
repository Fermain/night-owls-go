import { test, expect } from '@playwright/test';

test.describe('Lifecycle Error Reproduction', () => {
	test('login page should render without lifecycle errors', async ({ page }) => {
		// Enhanced console error tracking
		const errors: any[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				console.log('🚨 Console Error:', msg.text());
				console.log('🚨 Error Args:', msg.args());
				errors.push({
					text: msg.text(),
					type: msg.type(),
					location: msg.location()
				});
			}
		});

		// Enhanced page error tracking
		page.on('pageerror', (error) => {
			console.log('🚨 Page Error:', error.message);
			console.log('🚨 Stack:', error.stack);
			errors.push({
				message: error.message,
				stack: error.stack,
				type: 'pageerror'
			});
		});

		try {
			console.log('🧪 Loading login page...');
			await page.goto('/login');

			// Wait a bit for any async errors
			await page.waitForTimeout(2000);

			// Check for lifecycle errors
			const hasLifecycleError = errors.some(
				(error) =>
					error.text?.includes('lifecycle_outside_component') ||
					error.message?.includes('lifecycle_outside_component') ||
					error.stack?.includes('lifecycle_outside_component')
			);

			if (hasLifecycleError) {
				console.log('🚨 Found lifecycle errors:', JSON.stringify(errors, null, 2));
			}

			// Check if page rendered at all
			console.log('🔍 Checking if page rendered...');
			await expect(page.locator('body')).toBeVisible();

			// Look for any text content to verify the page actually loaded
			const hasAnyText = await page.locator('body').innerText();
			console.log(`📄 Page content length: ${hasAnyText.length}`);

			// Look for login form elements
			console.log('🔍 Looking for login form elements...');
			const phoneInputs = await page.locator('input[type="tel"]').count();
			const buttons = await page.locator('button').count();
			const forms = await page.locator('form').count();

			console.log(`✅ Found ${phoneInputs} input[type="tel"] elements`);
			console.log(`✅ Found ${buttons} button elements`);
			console.log(`✅ Found ${forms} form elements`);

			expect(hasLifecycleError).toBe(false);
			console.log('✅ Login page loaded successfully without lifecycle errors');
		} catch (error) {
			console.log('❌ Test failed with error:', error);
			console.log('🚨 All captured errors:', JSON.stringify(errors, null, 2));
			throw error;
		}
	});

	test('home page should render without lifecycle errors', async ({ page }) => {
		// Enhanced error tracking
		const errors: any[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				console.log('🚨 Console Error:', msg.text());
				errors.push({ text: msg.text(), type: msg.type(), location: msg.location() });
			}
		});

		page.on('pageerror', (error) => {
			console.log('🚨 Page Error:', error.message);
			console.log('🚨 Stack:', error.stack);
			errors.push({ message: error.message, stack: error.stack, type: 'pageerror' });
		});

		try {
			console.log('🧪 Loading home page...');
			await page.goto('/');
			await page.waitForTimeout(2000);

			const hasLifecycleError = errors.some(
				(error) =>
					error.text?.includes('lifecycle_outside_component') ||
					error.message?.includes('lifecycle_outside_component')
			);

			if (hasLifecycleError) {
				console.log('🚨 Found lifecycle errors:', JSON.stringify(errors, null, 2));
			}

			expect(hasLifecycleError).toBe(false);
			await expect(page.locator('body')).toBeVisible();

			console.log('✅ Home page loaded successfully');
		} catch (error) {
			console.log('❌ Test failed with error:', error);
			console.log('🚨 All captured errors:', JSON.stringify(errors, null, 2));
			throw error;
		}
	});

	test('register page should render without lifecycle errors', async ({ page }) => {
		// Enhanced error tracking
		const errors: any[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				console.log('🚨 Console Error:', msg.text());
				errors.push({ text: msg.text(), type: msg.type(), location: msg.location() });
			}
		});

		page.on('pageerror', (error) => {
			console.log('🚨 Page Error:', error.message);
			console.log('🚨 Stack:', error.stack);
			errors.push({ message: error.message, stack: error.stack, type: 'pageerror' });
		});

		try {
			console.log('🧪 Loading register page...');
			await page.goto('/register');
			await page.waitForTimeout(2000);

			const hasLifecycleError = errors.some(
				(error) =>
					error.text?.includes('lifecycle_outside_component') ||
					error.message?.includes('lifecycle_outside_component')
			);

			if (hasLifecycleError) {
				console.log('🚨 Found lifecycle errors:', JSON.stringify(errors, null, 2));
			}

			expect(hasLifecycleError).toBe(false);
			await expect(page.locator('body')).toBeVisible();

			console.log('✅ Register page loaded successfully');
		} catch (error) {
			console.log('❌ Test failed with error:', error);
			console.log('🚨 All captured errors:', JSON.stringify(errors, null, 2));
			throw error;
		}
	});
});
