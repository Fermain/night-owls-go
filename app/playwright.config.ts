import { defineConfig } from '@playwright/test';

export default defineConfig({
	webServer: {
		command: "echo 'Playwright webServer: Using server started by dev.sh'", // Dummy command to satisfy TS
		url: 'http://localhost:5173', // Playwright will wait for this URL to be available
		reuseExistingServer: true, // Important: use the server started by dev.sh
		// cwd: '.', // Not needed if the command is a simple echo
		timeout: 120 * 1000, // Timeout for the web server to be ready (Playwright still checks the URL)
	},
	use: {
		baseURL: 'http://localhost:5173',
		trace: 'on', // Enable tracing for all tests
	},
	testDir: 'e2e'
});
