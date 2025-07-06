/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';
import { readFileSync } from 'fs';
import { resolve } from 'path';

// Read version from package.json at build time
const packageJson = JSON.parse(readFileSync(resolve(__dirname, 'package.json'), 'utf-8'));
const appVersion = packageJson.version;

export default defineConfig(({ mode: _mode }) => {
	// Disable proxy during e2e tests to let MSW handle requests
	const _isE2ETesting = process.env.NODE_ENV === 'test' || process.env.PLAYWRIGHT_TEST === '1';

	return {
		plugins: [
			sveltekit(),
			SvelteKitPWA({
				strategies: 'generateSW',
				srcDir: './src',
				scope: '/',
				base: '/',
				mode: 'development',
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
				workbox: {
					globPatterns: ['**/*.{js,css,html,ico,png,svg,webmanifest}'],
					importScripts: ['/sw-push-handlers.js'],
					runtimeCaching: [
						{
							urlPattern: /^https:\/\/api\//,
							handler: 'NetworkFirst',
							options: {
								cacheName: 'api-cache',
								expiration: {
									maxEntries: 100,
									maxAgeSeconds: 60 * 60 * 24 // 24 hours
								}
							}
						}
					],
					additionalManifestEntries: [],
					skipWaiting: true,
					clientsClaim: true
				},
				devOptions: {
					enabled: true,
					suppressWarnings: process.env.SUPPRESS_WARNING === 'true',
					type: 'module',
					navigateFallback: '/',
					webManifestUrl: '/manifest.webmanifest'
				}
			})
		],

		test: {
			environment: 'jsdom',
			include: ['src/**/*.{test,spec}.{js,ts}'],
			exclude: ['e2e/**/*'],
			globals: true,
			setupFiles: []
		},

		server: {
			proxy:
				_mode === 'production'
					? undefined
					: {
							'/api': {
								target: process.env.PUBLIC_API_BASE_URL || 'http://localhost:5888',
								changeOrigin: true,
								secure: false,
								configure: (proxy, _options) => {
									proxy.on('proxyRes', (proxyRes, req, _res) => {
										// Fix trailer handling for empty responses (204, 304, etc.)
										if (proxyRes.statusCode === 204 || proxyRes.statusCode === 304) {
											// Remove problematic headers that cause TRAILERS issues
											delete proxyRes.headers['transfer-encoding'];
											delete proxyRes.headers['trailer'];
											delete proxyRes.headers['connection'];

											// Explicitly set content-length to 0 for empty responses
											proxyRes.headers['content-length'] = '0';

											// Remove any keep-alive headers that might cause issues
											if (proxyRes.headers['connection']) {
												delete proxyRes.headers['connection'];
											}
										}

										// General cleanup for DELETE method responses
										if (
											req.method === 'DELETE' &&
											proxyRes.statusCode &&
											proxyRes.statusCode < 300
										) {
											delete proxyRes.headers['transfer-encoding'];
											delete proxyRes.headers['trailer'];
											if (!proxyRes.headers['content-length']) {
												proxyRes.headers['content-length'] = '0';
											}
										}
									});

									// Handle proxy errors more gracefully
									proxy.on('error', (err, _req, res) => {
										console.error('Proxy error:', err);
										if (!res.headersSent) {
											res.writeHead(502, { 'Content-Type': 'application/json' });
											res.end(JSON.stringify({ error: 'Proxy error occurred' }));
										}
									});
								}
							}
						}
		},

		define: {
			'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
			// Inject app version from package.json at build time
			__APP_VERSION__: JSON.stringify(appVersion),
			__APP_NAME__: JSON.stringify(packageJson.name),
			__APP_DESCRIPTION__: JSON.stringify(
				packageJson.description || 'Night Owls Community Watch Platform'
			)
		}
	};
});
