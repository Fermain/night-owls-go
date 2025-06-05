/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { SvelteKitPWA } from '@vite-pwa/sveltekit';

export default defineConfig(({ mode: _mode }) => {
	// Disable proxy during e2e tests to let MSW handle requests
	const isE2ETesting = process.env.NODE_ENV === 'test' || process.env.PLAYWRIGHT_TEST === '1';

	return {
		plugins: [
			sveltekit(),
			SvelteKitPWA({
				srcDir: './src',
				mode: 'production',
				scope: '/',
				base: '/',
				selfDestroying: process.env.NODE_ENV === 'development',
				strategies: 'injectManifest',
				filename: 'service-worker.js',
				injectRegister: 'script-defer',
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
				registerType: 'autoUpdate',
				devOptions: {
					enabled: true,
					suppressWarnings: true,
					navigateFallback: '/',
					navigateFallbackAllowlist: [/^\/$/],
					type: 'module'
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
								secure: false
							}
						}
		},

		define: {
			'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development')
		}
	};
});
