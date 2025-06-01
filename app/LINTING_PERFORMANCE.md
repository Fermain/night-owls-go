# ESLint Performance Optimization

This project uses **fast linting by default** for development speed, with full type-checking available for comprehensive validation.

## Performance Results

- **Default linting**: ~4.5 seconds for `src/` directory
- **Full type-aware linting**: ~76 seconds for `src/` directory
- **Performance improvement**: **17x faster** for daily development

## Linting Scripts

### Default (Fast) - Recommended for Development

```bash
pnpm lint              # Fast lint (default) - 4.5 seconds
pnpm lint:fix          # Fast lint with auto-fix
```

### Full Type-Aware (Comprehensive)

```bash
pnpm lint:full         # Complete linting with prettier + type-checking
pnpm lint:ci           # Same as lint:full, optimized for CI/CD
```

### Specialized

```bash
pnpm lint:errors       # Lint with zero warnings tolerance
pnpm lint:components   # Lint only components directory
pnpm lint:routes       # Lint only routes directory
pnpm lint:utils        # Lint only utils and services
```

## Configuration Details

### Default Configuration (Fast)

- **Source**: Uses `eslint.config.fast.js`
- **Performance**: ~4.5 seconds
- **Features**: Essential linting rules, syntax checking, unused imports
- **Trade-off**: No TypeScript type-checking for speed
- **Use case**: Daily development, quick feedback

### Full Configuration (Comprehensive)

- **Source**: Uses `eslint.config.js`
- **Performance**: ~76 seconds
- **Features**: Complete TypeScript type-checking, all rules
- **Use case**: CI/CD, pre-push hooks, comprehensive validation

## Migration from Previous Setup

**Old scripts** → **New scripts**:

- `lint:fast` → `lint` (now default)
- `lint:fast-fix` → `lint:fix` (now default)
- `lint` → `lint:full` (full type-checking)
- `lint:quick` → **deprecated** (use `lint` instead)

## Optimization Techniques Applied

1. **Fast-by-Default Philosophy**

   - Development speed prioritized for daily workflow
   - Full validation available when needed

2. **Disabled TypeScript Project Service** in default config

   - Removes `projectService: true` for development speed
   - Eliminates full project type analysis

3. **Separate Configurations**

   - Fast config for development iteration
   - Full config for comprehensive checking

4. **ESLint Caching** for full type-checking
   - Speeds up subsequent full lint runs
   - Note: Not compatible with Svelte config serialization in fast mode

## Recommended Workflow

### Daily Development

```bash
pnpm lint              # Quick feedback (4.5s)
pnpm lint:fix          # Auto-fix issues
```

### Before Important Commits

```bash
pnpm lint:full         # Comprehensive validation (76s)
```

### CI/CD Pipeline

```bash
pnpm lint:ci           # Full validation with caching
```

## Common Issues

### Missing Type-Aware Rules in Default Lint

This is intentional for performance. Use `pnpm lint:full` when you need comprehensive type checking.

### Cache Serialization Error

If you see "function value cannot be serialized" errors with `lint:full`, it's due to Svelte config incompatibility with ESLint caching. This doesn't affect the default fast linting.
