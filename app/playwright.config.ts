import { defineConfig } from '@playwright/test';

export default defineConfig({
	use: {
		baseURL: 'http://localhost:8080' // Assuming Go server runs on port 8080
	},
	testDir: 'e2e'
});
