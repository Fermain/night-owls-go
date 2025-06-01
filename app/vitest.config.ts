/// <reference types="vitest" />
import { defineConfig } from 'vite';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
	plugins: [sveltekit()],

	// 👇 Critical fix for Svelte 5: force browser build in tests
	resolve: {
		conditions: ['browser']
	},

	test: {
		environment: 'jsdom',
		include: ['src/**/*.{test,spec}.{js,ts}'],
		exclude: ['e2e/**/*'],
		globals: true,
		setupFiles: ['src/tests/setup.ts']
	}
});
