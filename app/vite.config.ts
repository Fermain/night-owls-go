/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig(({ mode }) => {
	// Disable proxy during e2e tests to let MSW handle requests
	const isE2ETesting = process.env.NODE_ENV === 'test' || process.env.PLAYWRIGHT_TEST === '1';
	
	// Check if we're running tests (Vitest sets this environment variable)
	const isTesting = process.env.VITEST === 'true' || mode === 'test';

	return {
		plugins: [
			sveltekit(),
			SvelteKitPWA({
				srcDir: './src',
				mode: 'production',
				scope: '/',
				base: '/',
				selfDestroying: process.env.NODE_ENV === 'development',
				manifest: {
					name: 'Mount Moreland Night Owls',
					short_name: 'Night Owls',
					description: 'Community watch scheduling and incident reporting for Mount Moreland',
					theme_color: '#1f2937',
					background_color: '#ffffff',
					display: 'standalone',
					scope: '/',
					start_url: '/',
					icons: [
						{
							src: '/icons/icon-192x192.png',
							sizes: '192x192',
							type: 'image/png',
							purpose: 'any maskable'
						},
						{
							src: '/icons/icon-512x512.png',
							sizes: '512x512',
							type: 'image/png',
							purpose: 'any maskable'
						}
					]
				},
				injectManifest: {
					globPatterns: ['client/**/*.{js,css,ico,png,svg,webp,woff,woff2}']
				},
				workbox: {
					globPatterns: ['client/**/*.{js,css,ico,png,svg,webp,woff,woff2}'],
					cleanupOutdatedCaches: true,
					clientsClaim: true,
					skipWaiting: true,
					runtimeCaching: [
						{
							urlPattern: /^https:\/\/fonts\.googleapis\.com\/.*/i,
							handler: 'CacheFirst',
							options: {
								cacheName: 'google-fonts-cache',
								expiration: {
									maxEntries: 10,
									maxAgeSeconds: 60 * 60 * 24 * 365 // <== 365 days
								}
							}
						},
						{
							urlPattern: /^https:\/\/fonts\.gstatic\.com\/.*/i,
							handler: 'CacheFirst',
							options: {
								cacheName: 'gstatic-fonts-cache',
								expiration: {
									maxEntries: 10,
									maxAgeSeconds: 60 * 60 * 24 * 365 // <== 365 days
								}
							}
						},
						{
							urlPattern: /\/api\/.*/i,
							handler: 'NetworkFirst',
							options: {
								cacheName: 'api-cache',
								expiration: {
									maxEntries: 100,
									maxAgeSeconds: 60 * 5 // 5 minutes
								},
								networkTimeoutSeconds: 10
							}
						}
					]
				},
				registerType: 'autoUpdate',
				devOptions: {
					enabled: process.env.NODE_ENV === 'development',
					suppressWarnings: true,
					navigateFallback: '/',
					navigateFallbackAllowlist: [/^\/$/],
					type: 'module'
				}
			})
		],

		// 👇 Fix for Svelte 5 + Vitest: use browser build during tests
		resolve: {
			conditions: isTesting ? ['browser'] : [],
		},

		test: {
			environment: 'jsdom',
			include: ['src/**/*.{test,spec}.{js,ts}'],
			exclude: ['e2e/**/*'],
			globals: true,
			setupFiles: [],
		},

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
						},
						'/reports': {
							target: 'http://localhost:5888',
							changeOrigin: true
						},
						'/broadcasts': {
							target: 'http://localhost:5888',
							changeOrigin: true
						},
						'/push': {
							target: 'http://localhost:5888',
							changeOrigin: true
						},
						'/schedules': {
							target: 'http://localhost:5888',
							changeOrigin: true
						}
					}
		}
	};
});
