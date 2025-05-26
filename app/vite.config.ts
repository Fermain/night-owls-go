import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
// import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig(() => {
	// Disable proxy during e2e tests to let MSW handle requests
	const isE2ETesting = process.env.NODE_ENV === 'test' || process.env.PLAYWRIGHT_TEST === '1';

	return {
		plugins: [
			sveltekit()
			// TODO: Re-enable PWA plugin when we integrate our service worker properly
			// SvelteKitPWA({...})
		],
		server: {
							proxy: isE2ETesting
				? undefined
				: {
						'/api': {
							target: 'http://localhost:5888',
							changeOrigin: true
						},
						'/shifts': {
							target: 'http://localhost:5888',
							changeOrigin: true
						},
						'/bookings': {
							target: 'http://localhost:5888',
							changeOrigin: true
						}
					}
		}
	};
});
