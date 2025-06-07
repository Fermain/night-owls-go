# Scripts

This directory contains development and CI scripts for the Night Owls project.

## API Endpoint Validation

### `validate-api-endpoints.sh`

A CI script that validates frontend API calls against backend endpoints to prevent 404 errors in production.

**Features:**
- Extracts all unique API endpoints from frontend TypeScript/JavaScript/Svelte files
- Parses backend route registrations from `cmd/server/main.go`
- Cross-references with OpenAPI specification if available
- Reports missing endpoints with actionable recommendations

**Usage:**
```bash
# Run directly
./scripts/validate-api-endpoints.sh

# Run via Makefile (recommended)
make validate-api

# Run as part of comprehensive checks
make check
```

**Sample Output:**
```
🔍 Validating API endpoint consistency between frontend and backend...
Extracting API endpoints from frontend...
Found 38 unique API endpoints in frontend code:
  /api/admin/users
  /api/admin/users/{id}
  /api/bookings
  ...

Extracting registered routes from backend...
Found 57 registered routes in backend:
  /api/admin/users
  /api/admin/users/{id}
  /api/bookings
  ...

✅ VALIDATION PASSED: All frontend API endpoints are implemented in backend
🎉 API endpoint validation completed successfully!
```

**Integration:**
- Exit code 0: All endpoints validated successfully
- Exit code 1: Missing endpoints found (CI will fail)
- Integrated into `make check` for comprehensive validation

**What it catches:**
- Frontend calling non-existent backend endpoints
- Typos in API endpoint paths
- Missing route registrations
- Inconsistent parameter patterns (`{id}` vs `{userId}`)

### Adding to CI/CD

Add to your CI pipeline:

```yaml
# GitHub Actions example
- name: Validate API endpoints
  run: make validate-api

# Or as part of comprehensive checks
- name: Run all checks
  run: make check
```

This prevents the common issue where frontend code calls an API endpoint that doesn't exist in the backend, which would result in 404 errors in production. 