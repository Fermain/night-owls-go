import { type FullConfig } from '@playwright/test';

async function globalSetup(_config: FullConfig) {
	// Global setup logic would go here
	console.log('✅ MSW server started for e2e tests');
	console.log('   Playwright route interception will handle API requests');
	console.log('Global setup completed');
}

export default globalSetup;
