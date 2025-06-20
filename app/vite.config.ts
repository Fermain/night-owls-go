/// <reference types="vitest" />
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig(({ mode: _mode }) => {
	// Disable proxy during e2e tests to let MSW handle requests
	const isE2ETesting = process.env.NODE_ENV === 'test' || process.env.PLAYWRIGHT_TEST === '1';

	return {
		plugins: [sveltekit()],

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
