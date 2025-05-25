import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig(() => {
	// Disable proxy during e2e tests to let MSW handle requests
	const isE2ETesting = process.env.NODE_ENV === 'test' || process.env.PLAYWRIGHT_TEST === '1';
	
	return {
		plugins: [sveltekit()],
		server: {
			proxy: isE2ETesting ? undefined : {
				'/api': {
					target: 'http://localhost:8080',
					changeOrigin: true
				},
				'/shifts': {
					target: 'http://localhost:8080',
					changeOrigin: true
				},
				'/bookings': {
					target: 'http://localhost:8080',
					changeOrigin: true
				}
			}
		}
	};
});
