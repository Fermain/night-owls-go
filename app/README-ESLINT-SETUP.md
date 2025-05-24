# ESLint Unused Imports Setup

This project is configured with `eslint-plugin-unused-imports` to automatically detect and remove unused imports and variables.

## What's Configured

### ESLint Plugin
- **Plugin**: `eslint-plugin-unused-imports@^4.1.4`
- **Rules**:
  - `unused-imports/no-unused-imports: error` - Detects and auto-removes unused imports
  - `unused-imports/no-unused-vars: warn` - Warns about unused variables (respects `_` prefix for intentionally unused)

### VS Code Integration
- Auto-fix ESLint issues on save
- Organize imports automatically
- Svelte language support

## How to Use

### Command Line
```bash
# Fix all ESLint issues including unused imports
npm run lint:fix

# Focus specifically on unused imports
npm run lint:unused

# Check for issues without fixing
npm run lint
```

### VS Code (Automatic)
With the provided `.vscode/settings.json`:
1. **On Save**: Automatically removes unused imports and fixes other ESLint issues
2. **Manual**: Use `Ctrl/Cmd + Shift + P` → "ESLint: Fix all auto-fixable Problems"

### Recommended Extensions
Install these VS Code extensions (listed in `.vscode/extensions.json`):
- **Svelte for VS Code** - Svelte language support
- **ESLint** - ESLint integration
- **Prettier** - Code formatting
- **Tailwind CSS IntelliSense** - Tailwind support

## Variable Naming Convention

To mark variables as intentionally unused, prefix them with `_`:

```typescript
// ❌ Will warn about unused variable
function example(data, index) {
  return data.name;
}

// ✅ Won't warn - underscore prefix indicates intentional
function example(data, _index) {
  return data.name;
}

// ✅ Alternative - fully underscore named
function example(data, _) {
  return data.name;
}
```

## What Gets Auto-Fixed

### ✅ Automatically Fixed
- Unused imports are completely removed
- Unused variables with `_` prefix are ignored

### ⚠️ Warns but Doesn't Auto-Fix
- Unused variables (to prevent accidental removal of important code)
- Unused function parameters (to maintain function signatures)
- Unused caught errors in try/catch blocks

## Example

**Before auto-fix:**
```typescript
import { format, parseISO, addDays } from 'date-fns'; // addDays is unused
import { createQuery } from '@tanstack/svelte-query';  // unused import

function formatDate(dateString: string) {
  const date = parseISO(dateString);   // used
  const formatted = format(date, 'yyyy-MM-dd'); // used
  const unused = 'test';               // unused variable
  return formatted;
}
```

**After auto-fix:**
```typescript
import { format, parseISO } from 'date-fns'; // addDays removed
// createQuery import completely removed

function formatDate(dateString: string) {
  const date = parseISO(dateString);   // used
  const formatted = format(date, 'yyyy-MM-dd'); // used
  const unused = 'test';               // ⚠️ still warns, not auto-removed
  return formatted;
}
```

## Integration with CI/CD

The linting rules will fail builds if there are unused imports (error level), helping maintain clean code across the team.

## Troubleshooting

### "ESLint fixes not applying on save"
1. Ensure ESLint extension is installed
2. Check that `.vscode/settings.json` is in your workspace
3. Restart VS Code

### "Still seeing unused import warnings"
- Run `npm run lint:fix` to fix all auto-fixable issues
- Some complex cases might need manual review

### "Plugin not found errors"
- Ensure dependencies are installed: `npm install`
- Restart your editor after installing new plugins 