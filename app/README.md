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

Once you've created a project and installed dependencies with `pnpm install` (or `npm install` or `yarn`), start a development server:

```bash
pnpm run dev

# or start the server and open the app in a new browser tab
pnpm run dev -- --open
```

## Building

To create a production version of your app:

```bash
pnpm run build
```

You can preview the production build with `pnpm run preview`.

> To deploy your app, you may need to install an [adapter](https://kit.svelte.dev/docs/adapters) for your target environment.

## Linting Strategy

This project uses a comprehensive linting setup. Choose the right command for your needs:

### Quick Commands (Recommended for Development)
- `pnpm run lint` - Full lint check (includes Prettier + ESLint on all files) - **~5+ minutes**
- `pnpm run lint:quick` - **Fastest option** - Only shows errors, suppresses warnings - **~50 seconds**
- `pnpm run lint:errors` - Only fails on errors (warnings allowed) - **~50 seconds**
- `pnpm run lint:fast` - Lint only src/ directory - **~1.5 minutes**
- `pnpm run lint:components` - Lint only components directory - **~45 seconds**
- `pnpm run lint:routes` - Lint only routes directory
- `pnpm run lint:utils` - Lint only utils and services directories
- `pnpm run lint:fix` - Auto-fix all fixable issues
- `pnpm run lint:unused` - Focus on unused imports

### Recommended Workflow

1. **During development**: Use `pnpm run lint:quick` for fast error-only feedback
2. **Before committing**: Use `pnpm run lint:fix` to auto-fix issues
3. **Final check**: Use `pnpm run lint` for complete validation

### Current Status

- **TypeScript Errors**: 0 ✅
- **ESLint Errors**: 0 ✅ (All explicit `any` types fixed!)
- **ESLint Warnings**: 22 (down from 35, mostly unused variables)
