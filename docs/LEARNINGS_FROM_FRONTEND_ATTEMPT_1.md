# Learnings from Frontend Attempt 1 (Svelte + Vite)

This document summarizes the progress and key takeaways from the initial attempt to build the frontend application using plain Svelte 5 with Vite. These notes are intended to inform the next attempt, which will use SvelteKit and Tailwind CSS v3.

## 1. Accomplishments

*   **Project Scaffolding:** Successfully created a Svelte 5 + TypeScript project using Vite in a `frontend/` directory.
*   **Styling Setup:**
    *   Integrated Tailwind CSS v4 (though this presented challenges).
    *   Initialized `shadcn-svelte @next` and configured its path aliases (`$lib`) in `tsconfig.app.json` and `vite.config.ts`.
    *   Basic CSS variable setup in `src/app.css` for `shadcn-svelte` theming.
*   **Linting & Formatting:**
    *   ESLint configured for Svelte 5 and TypeScript using a flat `eslint.config.js`.
    *   Prettier configured with `prettier-plugin-svelte` and integrated with ESLint via `eslint-config-prettier`.
    *   Stylelint configured for CSS and Svelte files, including Tailwind CSS rules.
    *   Added npm scripts for linting and formatting.
*   **Basic Application Structure:**
    *   Created an API client (`src/lib/api/client.ts`) with JWT handling (from localStorage) and methods for API calls. Configured to use Vite environment variables (`import.meta.env.VITE_API_BASE_URL`).
    *   Implemented a Svelte 5 Runes-based authentication store (`src/lib/stores/auth.ts`) for managing token and user state.
    *   Scaffolded placeholder pages for Login (`AuthLogin.svelte`) and OTP Verification (`AuthVerify.svelte`) in `src/pages/`.
    *   Set up a basic client-side router (`Router.svelte`) using `svelte-navaid`.
    *   Initialized `@tanstack/svelte-query` and made its `QueryClient` available via Svelte context in `App.svelte`.
*   **Development Environment:**
    *   Created a root `package.json` with `npm-run-all` to concurrently run the Go backend and the Svelte frontend dev server.
*   **Documentation:**
    *   Created `docs/SVELTE5_FEATURES.md` detailing Runes and new Svelte 5 concepts.

## 2. Key Challenges & Learnings

*   **Tailwind CSS v4 and `@apply`:** Encountered issues with `npm run build` failing due to `Cannot apply unknown utility class: ...` when using `@apply` with CSS variable-based utilities (like `border-border`, `bg-background`) defined by `shadcn-svelte` in the global `src/app.css`. This was specific to Tailwind v4's processing. The workaround was to replace these specific `@apply` rules with direct CSS variable assignments (e.g., `border-color: hsl(var(--border));`).
    *   *Decision for next attempt:* Use **Tailwind CSS v3** to ensure smoother compatibility with the current ecosystem, especially `shadcn-svelte`.
*   **TanStack Component Compatibility with Svelte 5:**
    *   `@tanstack/svelte-table`: The stable version has a peer dependency on Svelte 3/4 and does not officially support Svelte 5 yet. Installation failed due to this.
    *   `@tanstack/svelte-form`: Similar concerns for Svelte 5 compatibility exist.
    *   `@tanstack/svelte-query`: Installed successfully and appears to be compatible with Svelte 5.
    *   *Decision for next attempt:* While Svelte Query will be used, proceed with caution for Table and Form, or be prepared to find Svelte 5 compatible alternatives if official support is still pending.
*   **Svelte 5 Rune Usage & Tooling:**
    *   Initial confusion with importing Runes (e.g., `$state`) - they are keywords and not imported.
    *   Typing dynamic components with `<svelte:component this={...}>` and Svelte 5's `ComponentType` presented some challenges, temporarily resolved by using `any` for the component variable in the router. The `vite-plugin-svelte` also showed a deprecation warning for `<svelte:component>` in runes mode, though it's still the standard way for a truly dynamic component variable.
    *   Event handling directives like `on:submit` are deprecated in favor of `onsubmit` attributes.
    *   `setContext` for Svelte Query needed to be called within a component's lifecycle, not at the top level of `main.ts`.
*   **Build Tooling & Configuration:**
    *   Initial issues with `tailwindcss init -p` not finding the executable, resolved by manual config file creation or reinstalling specific packages.
    *   Ensuring `tsconfig.node.json` was present and correctly configured for `npm run check`.
    *   Correct ESLint flat config structure for Svelte plugins took a few iterations.
    *   JSON files (like `.stylelintrc.json`) do not support comments.

## 3. Plan for Next Attempt (Project "app" with SvelteKit)

1.  **Initialize with SvelteKit:** Use `npm create svelte@latest app` and choose the SvelteKit skeleton or demo app template with TypeScript.
2.  **Tailwind CSS v3:** Install and configure Tailwind CSS v3, PostCSS, and Autoprefixer.
    *   `npm install -D tailwindcss@^3 postcss autoprefixer svelte-preprocess`
    *   Configure `tailwind.config.js`, `postcss.config.js`.
    *   Update `svelte.config.js` to use `vitePreprocess` and `svelte-preprocess` for Tailwind.
    *   Create `src/app.pcss` (or similar) for Tailwind directives and import it in `+layout.svelte` or `app.html`.
3.  **Shadcn-svelte:** Initialize `shadcn-svelte` (which should work well with Tailwind v3).
    *   Configure path aliases (`$lib`) in `tsconfig.json` (SvelteKit often sets this up) and `vite.config.ts`.
4.  **TanStack Components:**
    *   Install `@tanstack/svelte-query` (v5 confirmed working with Svelte 5).
    *   Attempt to install latest stable `@tanstack/svelte-table` and `@tanstack/svelte-form`. If Svelte 5 peer dependency issues persist with their stable versions even with Tailwind v3, we will use the `--legacy-peer-deps` flag for installation and test thoroughly, or seek alternatives/defer.
    *   Configure `QueryClient` in a root layout (`+layout.svelte` or `+layout.ts`) using SvelteKit's context module if appropriate, or standard Svelte context.
5.  **Environment Variables:** Use SvelteKit's `$env/static/public` for client-side environment variables (e.g., `PUBLIC_API_BASE_URL`).
6.  **Linting & Formatting:** Set up ESLint (with `eslint-plugin-svelte`), Prettier (with `prettier-plugin-svelte`), and Stylelint similarly to the first attempt, adapting paths for SvelteKit structure.
7.  **API Client & Auth Store:** Re-implement `src/lib/api/client.ts` and `src/lib/stores/auth.ts` using SvelteKit conventions where applicable (e.g., for env vars).
8.  **Routing & Pages:** Utilize SvelteKit's file-system based routing to create login and OTP verification pages in `src/routes`.
9.  **Static Build:** Configure `adapter-static` in `svelte.config.js` with a fallback for SPA behavior, and set `ssr = false`, `prerender = true` in the root `+layout.ts` to ensure a client-side only build, as per the provided guide.

## 4. Configurations to Remember

*   **`vite.config.ts` for path aliases:**
    ```typescript
    import path from 'path';
    // ...
    resolve: {
      alias: {
        '$lib': path.resolve(__dirname, './src/lib')
      }
    }
    ```
*   **`tsconfig.json` for path aliases (SvelteKit handles this slightly differently but concept is similar):**
    ```json
    "baseUrl": ".",
    "paths": {
      "$lib": ["src/lib"],
      "$lib/*": ["src/lib/*"]
    }
    ```
*   **`postcss.config.js` (for Tailwind v3):**
    ```javascript
    module.exports = {
      plugins: {
        tailwindcss: {},
        autoprefixer: {},
      },
    };
    ```
*   **ESLint (`eslint.config.js` structure for flat config was generally good towards the end):** Refer to the working version from the previous attempt.
*   **`.stylelintrc.json` (working version):**
    ```json
    {
      "extends": [
        "stylelint-config-standard",
        "stylelint-config-tailwindcss"
      ],
      "rules": {
        "at-rule-no-unknown": [true, { "ignoreAtRules": ["tailwind", "apply", "layer", "config", "variants", "responsive", "screen"] } ],
        "function-no-unknown": [true, { "ignoreFunctions": ["theme"] } ],
        "at-rule-no-deprecated": null
      },
      "overrides": [ { "files": ["**/*.svelte"], "customSyntax": "postcss-html" } ]
    }
    ```

This document should serve as a good checkpoint. 