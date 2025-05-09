import eslintPluginSvelte from 'eslint-plugin-svelte';
import typescriptParser from '@typescript-eslint/parser';
import eslintConfigPrettier from 'eslint-config-prettier';
import globals from 'globals'; // For browser/node globals

export default [
  // Global ignores
  {
    ignores: [
      'build/',
      '.svelte-kit/',
      'dist/', // Common output directory
      'node_modules/',
      '*.config.js', // Ignoring our own config files for now
      '*.config.ts',
    ],
  },
  {
    files: ['**/*.{js,mjs,cjs,ts}'],
    languageOptions: {
      parser: typescriptParser,
      globals: {
        ...globals.browser,
        ...globals.node,
        ...globals.es2021,
      },
    },
    rules: {
      // Add any specific JS/TS rules here
      // Example:
      // 'no-unused-vars': ['warn', { 'argsIgnorePattern': '^_' }],
    },
  },
  // Svelte configurations from plugin (this is expected to be an array of config objects)
  ...eslintPluginSvelte.configs['flat/recommended'],

  // Add any Svelte-specific overrides or additional settings in a new object
  // This ensures it applies after the defaults from 'flat/recommended'
  {
    files: ['**/*.svelte'], // Re-target Svelte files for overrides
    languageOptions: {
      parserOptions: {
        // The recommended config should set the Svelte parser,
        // but we ensure the TS parser is known for script blocks.
        parser: typescriptParser,
      },
      globals: {
        ...globals.browser, // Ensure browser globals for Svelte files
        ...globals.es2021,
      },
    },
    rules: {
      // Your custom Svelte rules or overrides here
      // These will override rules from 'flat/recommended' if they conflict
      // e.g., 'svelte/no-at-html-tags': 'off',
    },
  },
  eslintConfigPrettier, // Add Prettier config last
];
