import { type FullConfig } from '@playwright/test';

async function globalSetup(config: FullConfig) {
  console.log('✅ MSW server started for e2e tests');
  console.log('   Playwright route interception will handle API requests');
}

export default globalSetup; 