# Development Workflow

This project uses automated code quality tools and git hooks to maintain consistent code standards.

## Pre-commit Hooks

### Husky
- **Purpose**: Manages git hooks
- **Configuration**: `.husky/` directory
- **Installation**: Automatically installed via `pnpm install` (prepare script)

### Lint-staged
- **Purpose**: Runs linters only on staged files for faster execution
- **Configuration**: `.lintstagedrc.js`
- **Runs on**: Pre-commit hook

#### What gets linted:
- **JavaScript/TypeScript/Svelte files**: ESLint + Prettier
- **JSON/Markdown/YAML files**: Prettier formatting
- **Svelte files**: Additional svelte-check for type checking

### Commitlint
- **Purpose**: Enforces conventional commit message format
- **Configuration**: `commitlint.config.js`
- **Runs on**: commit-msg hook

## Commit Message Format

Follow [Conventional Commits](https://conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Allowed Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `build`: Build system changes
- `ci`: CI/CD changes
- `chore`: Maintenance tasks
- `revert`: Reverting previous commits

### Examples:
- `feat: add user authentication`
- `fix: resolve login validation error`
- `docs: update API documentation`
- `style: format code with prettier`

## Available Scripts

```bash
# Run lint-staged manually
pnpm run lint:staged

# Test commit message format
echo "feat: your message" | pnpm run commitlint

# Install/reinstall hooks
pnpm run hooks:install

# Remove hooks (if needed)
pnpm run hooks:uninstall
```

## Workflow

1. **Make changes** to your code
2. **Stage files** with `git add`
3. **Commit** with conventional message format
   - Pre-commit hook runs lint-staged automatically
   - commit-msg hook validates your commit message
4. **Push** to remote repository

If any hooks fail, fix the issues and commit again. The hooks ensure code quality and consistent commit history. 