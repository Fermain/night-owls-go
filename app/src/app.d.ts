// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}

	// Vite build-time constants injected from package.json
	const __APP_VERSION__: string;
	const __APP_NAME__: string;
	const __APP_DESCRIPTION__: string;
}

export {};
