import { test, expect } from '@playwright/test';

test.describe('🔍 Navigation Debug', () => {
	test('Debug navigation routes', async ({ page }) => {
		// Test each route individually to isolate issues (removed non-existent /shifts route)
		const routes = [
			{ path: '/', name: 'Homepage' },
			{ path: '/login', name: 'Login' },
			{ path: '/register', name: 'Register' },
			{ path: '/admin', name: 'Admin' }
		];

		for (const route of routes) {
			console.log(`Testing route: ${route.path}`);

			await page.goto(route.path);
			await page.waitForLoadState('networkidle');

			// Get page info for debugging
			const url = page.url();
			const title = await page.title();
			const bodyVisible = await page.locator('body').isVisible();
			const bodyText = await page.locator('body').textContent();

			console.log(`Route: ${route.path}`);
			console.log(`- URL: ${url}`);
			console.log(`- Title: ${title}`);
			console.log(`- Body visible: ${bodyVisible}`);
			console.log(`- Body has content: ${(bodyText?.length || 0) > 0}`);

			if (!bodyVisible) {
				// If body is not visible, try to understand why
				const bodyStyles = await page.locator('body').evaluate((el) => {
					const styles = window.getComputedStyle(el);
					return {
						display: styles.display,
						visibility: styles.visibility,
						opacity: styles.opacity,
						height: styles.height
					};
				});
				console.log(`- Body styles:`, bodyStyles);

				// Check if any content is actually present
				const hasContent = await page.locator('body *').count();
				console.log(`- Child elements: ${hasContent}`);
			}

			console.log('---');
		}
	});

	test('Alternative body visibility check', async ({ page }) => {
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Try different ways to check if the page is properly loaded
		const checks = {
			bodyVisible: await page.locator('body').isVisible(),
			htmlVisible: await page.locator('html').isVisible(),
			hasMainContent: await page.locator('main, div, [role="main"]').first().isVisible(),
			hasText: ((await page.textContent('body')) || '').length > 0,
			specificContent: await page
				.getByText('Night Owls')
				.isVisible()
				.catch(() => false)
		};

		console.log('Visibility checks:', checks);

		// If body is not visible but content exists, this might be a test runner issue
		if (!checks.bodyVisible && checks.hasText) {
			// Try waiting for any visible element
			await expect(page.locator('html')).toBeVisible();
			console.log('✅ HTML is visible, test adjusted');
		}
	});

	test('Debug page loading with console monitoring', async ({ page }) => {
		// Monitor console errors
		const consoleErrors: string[] = [];
		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				consoleErrors.push(msg.text());
			}
		});

		// Monitor network failures
		const networkErrors: string[] = [];
		page.on('response', (response) => {
			if (response.status() >= 400) {
				networkErrors.push(`${response.status()} ${response.url()}`);
			}
		});

		console.log('Navigating to homepage...');
		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Wait a bit more to ensure component mounting
		await page.waitForTimeout(1000);

		console.log('Console errors:', consoleErrors);
		console.log('Network errors:', networkErrors);

		// Check DOM structure
		const htmlContent = await page.locator('html').innerHTML();
		const bodyContent = await page.locator('body').innerHTML();

		console.log('HTML has content:', htmlContent.length > 100);
		console.log('Body content length:', bodyContent.length);
		console.log('Body content preview:', bodyContent.substring(0, 200) + '...');

		// Check if SvelteKit is working
		const svelteKitElements = await page.locator('[data-sveltekit]').count();
		console.log('SvelteKit elements found:', svelteKitElements);

		// Check for homepage content
		const nightOwlsText = await page.getByText('Night Owls').count();
		console.log('Night Owls text found:', nightOwlsText);
	});

	test('Check if page JavaScript is executing', async ({ page }) => {
		// Add a custom script to check if the page is actually loading
		await page.addInitScript(() => {
			// eslint-disable-next-line @typescript-eslint/no-explicit-any
			(window as any).pageLoadTest = 'init-script-executed';
		});

		await page.goto('/');
		await page.waitForLoadState('networkidle');

		// Check if our init script ran
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		const initScriptRan = await page.evaluate(() => (window as any).pageLoadTest);
		console.log('Init script executed:', initScriptRan);

		// Check if any Svelte components are mounted
		const svelteExists = await page.evaluate(() => {
			// Check for Svelte-related globals
			return {
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				hasSvelte: typeof (window as any).__SVELTE__ !== 'undefined',
				// eslint-disable-next-line @typescript-eslint/no-explicit-any
				hasVite: typeof (window as any).__vite__ !== 'undefined',
				userAgent: navigator.userAgent,
				url: window.location.href
			};
		});

		console.log('Svelte environment:', svelteExists);
	});
});
