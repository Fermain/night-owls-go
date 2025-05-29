import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: 'e2e',

	// Global setup to start MSW
	globalSetup: './e2e/setup/global-setup.ts',

	// Use a local test server instead of depending on external services
	use: {
		baseURL: 'http://localhost:4173',
		trace: 'on-first-retry',
		screenshot: 'only-on-failure',
		video: 'retain-on-failure'
	},

	// Configure projects for different test types
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		}
	],

	// Test configuration
	timeout: 30 * 1000,
	expect: {
		timeout: 5 * 1000
	},

	// Run tests in parallel
	workers: 1,

	// Retry configuration
	retries: 0,

	// Reporter configuration
	reporter: [['html'], ['line']],

	// Output directory
	outputDir: 'test-results/',

	webServer: {
		command: 'PLAYWRIGHT_TEST=1 npm run build && PLAYWRIGHT_TEST=1 npm run preview',
		port: 4173,
		env: {
			PLAYWRIGHT_TEST: '1'
		}
	}
});
