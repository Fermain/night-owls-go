import { persisted } from 'svelte-persisted-store';
import { derived, type Readable } from 'svelte/store';

export type ThemeMode = 'light' | 'dark' | 'system';

interface ThemeState {
	mode: ThemeMode;
}

const initialThemeState: ThemeState = {
	mode: 'system'
};

// Persisted store for theme preference
export const themeState = persisted<ThemeState>('theme-state', initialThemeState);

// Derived store that calculates the actual theme to apply
export const actualTheme: Readable<'light' | 'dark'> = derived(themeState, (state, set) => {
	if (typeof window === 'undefined') {
		set('light');
		return;
	}

	if (state.mode === 'system') {
		// Listen to system preference changes
		const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');

		const updateTheme = () => {
			set(mediaQuery.matches ? 'dark' : 'light');
		};

		// Set initial value
		updateTheme();

		// Listen for changes
		mediaQuery.addEventListener('change', updateTheme);

		// Cleanup
		return () => {
			mediaQuery.removeEventListener('change', updateTheme);
		};
	} else {
		set(state.mode as 'light' | 'dark');
	}
});

// Actions for theme management
export const themeActions = {
	setMode(mode: ThemeMode) {
		themeState.update((state) => ({
			...state,
			mode
		}));
	},

	// Apply theme to document
	applyTheme(theme: 'light' | 'dark') {
		if (typeof document === 'undefined') return;

		if (theme === 'dark') {
			document.documentElement.classList.add('dark');
		} else {
			document.documentElement.classList.remove('dark');
		}
	}
};
