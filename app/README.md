# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```bash
# create a new project in the current directory
npx sv create

# create a new project in my-app
npx sv create my-app
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```bash
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```bash
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## ESLint Optimization

Due to the complexity of our TypeScript + Svelte setup, ESLint caching is not available (configuration contains non-serializable functions). However, we've implemented several optimization strategies:

### Available Lint Scripts

- `npm run lint` - Full lint check (includes Prettier + ESLint on all files) - **~5+ minutes**
- `npm run lint:quick` - **Fastest option** - Only shows errors, suppresses warnings - **~50 seconds**
- `npm run lint:errors` - Only fails on errors (warnings allowed) - **~50 seconds**
- `npm run lint:fast` - Lint only src/ directory - **~1.5 minutes**
- `npm run lint:components` - Lint only components directory - **~45 seconds**
- `npm run lint:routes` - Lint only routes directory
- `npm run lint:utils` - Lint only utils and services directories
- `npm run lint:fix` - Auto-fix all fixable issues
- `npm run lint:unused` - Focus on unused imports

### Recommended Development Workflow

1. **During development**: Use `npm run lint:quick` for fast error-only feedback
2. **Before committing**: Use `npm run lint:fix` to auto-fix issues
3. **Final check**: Use `npm run lint` for complete validation

### Current Status

- **TypeScript Errors**: 0 ✅
- **ESLint Errors**: 0 ✅ (All explicit `any` types fixed!)
- **ESLint Warnings**: 22 (down from 35, mostly unused variables)
