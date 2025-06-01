# Testing Strategy

This project implements a **three-tier testing strategy** for comprehensive coverage while maintaining fast development feedback loops.

## 🧪 Test Categories

### 1. **Unit/Component Tests**

- **Location**: `src/**/*.test.ts`
- **Purpose**: Test individual components and functions in isolation
- **Strategy**: Mock all external dependencies
- **Speed**: Very fast (~1-2 seconds)
- **When to run**: During development, pre-commit hooks

```bash
pnpm test:unit          # Run unit tests
pnpm test:unit:watch    # Watch mode for development
```

### 2. **Integration Tests**

- **Location**: `e2e/integration-real-api.test.ts`
- **Purpose**: Test real API integration and backend functionality
- **Strategy**: Use actual Go backend on localhost:5888
- **Speed**: Medium (~10-30 seconds)
- **When to run**: Pre-push, CI/CD pipeline
- **Requirements**: Go backend must be running

```bash
pnpm test:integration   # Run integration tests
```

### 3. **E2E User Journey Tests**

- **Location**: `e2e/api-integration.test.ts`, `e2e/smoke.test.ts`
- **Purpose**: Test complete user workflows and UI interactions
- **Strategy**: Mock all external APIs for predictable testing
- **Speed**: Slower (~30-60 seconds)
- **When to run**: Pre-release, CI/CD pipeline

```bash
pnpm test:e2e           # Run all e2e tests
pnpm test:e2e:ui        # Run with Playwright UI
```

## 🎯 Testing Matrix

| Test Type       | Speed     | Dependencies     | Purpose         | Mock APIs       |
| --------------- | --------- | ---------------- | --------------- | --------------- |
| **Unit**        | ⚡ Fast   | None             | Component logic | ✅ All mocked   |
| **Integration** | 🔄 Medium | Real backend     | API contracts   | ❌ Real backend |
| **E2E**         | 🐌 Slow   | SvelteKit server | User journeys   | ✅ All mocked   |

## 🔧 Configuration

### Playwright Configuration

- **Base URL**: `http://localhost:4173` (SvelteKit preview)
- **Global Setup**: MSW server initialization
- **Route Interception**: API mocking for e2e tests
- **Screenshots**: On failure only
- **Videos**: Retained on failure

### Test Environment Variables

```bash
PLAYWRIGHT_TEST=1       # Enables test mode in SvelteKit
DEBUG=playwright:*      # Enable Playwright debugging
```

## 🚀 Running Tests

### Development Workflow

```bash
# Quick feedback during development
pnpm test:unit:watch

# Periodic integration check
pnpm test:integration

# Before major commits
pnpm test:e2e
```

### CI/CD Pipeline

```bash
# Full test suite
pnpm test               # Runs unit + integration + e2e
```

## 📋 Test Guidelines

### Unit Tests

- Test one component/function at a time
- Mock all external dependencies
- Focus on business logic and edge cases
- Fast and deterministic

### Integration Tests

- Test real API endpoints
- Verify data flow between frontend and backend
- Test authentication, authorization, and data persistence
- Require actual backend services

### E2E Tests

- Test complete user workflows
- Mock external APIs for consistency
- Focus on user interactions and UI behavior
- Test error handling and edge cases

## 🐛 Debugging Tests

### Failed Tests

```bash
# View last test results
pnpm exec playwright show-report

# Debug specific test
pnpm exec playwright test --debug api-integration.test.ts

# Run single test file
pnpm exec playwright test smoke.test.ts
```

### API Issues

- **Unit tests**: Check component mocks
- **Integration tests**: Verify backend is running on localhost:5888
- **E2E tests**: Check route interception in `setup/api-mocks.ts`

## 📊 Test Coverage

### What Each Type Covers

**Unit Tests**:

- Component rendering
- State management
- Utility functions
- Form validation

**Integration Tests**:

- Authentication flow
- API endpoints
- Database operations
- JWT token handling

**E2E Tests**:

- User registration journey
- Shift booking workflow
- Admin dashboard operations
- Error handling UX

## 🔄 Migration Notes

### Changes from Previous Setup

- ✅ **Fixed**: Route interception now works properly
- ✅ **Separated**: Real API tests vs mocked API tests
- ✅ **Organized**: Clear test categories and purposes
- ✅ **Performance**: Faster feedback loops

### Deprecated Patterns

- ❌ Mixed testing strategies in single files
- ❌ Unclear separation between unit and integration tests
- ❌ Failing route interception

## 🏗️ Future Improvements

1. **Visual Regression Tests**: Add screenshot comparisons
2. **Performance Tests**: API response time monitoring
3. **Accessibility Tests**: Automated a11y testing
4. **Load Tests**: Backend stress testing
