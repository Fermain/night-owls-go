import { type FullConfig } from '@playwright/test';
import { setupServer } from 'msw/node';
import { handlers } from './mocks';

const server = setupServer(...handlers);

async function globalSetup(config: FullConfig) {
  // Start MSW server
  server.listen({
    onUnhandledRequest: 'warn'
  });

  console.log('✅ MSW server started for e2e tests');

  // Setup cleanup on process exit
  const handleExit = () => server.close();
  globalThis.process?.on?.('SIGTERM', handleExit);
  globalThis.process?.on?.('SIGINT', handleExit);
}

export default globalSetup; 