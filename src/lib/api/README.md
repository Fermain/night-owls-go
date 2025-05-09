# API Type Generation

This directory contains TypeScript type definitions generated from the Go backend's OpenAPI specification.

## How it works

1. The Go backend uses [swaggo/swag](https://github.com/swaggo/swag) to generate Swagger 2.0 documentation from Go code annotations
2. We then convert the Swagger 2.0 spec to OpenAPI 3.0 using [api-spec-converter](https://github.com/LucyBot-Inc/api-spec-converter)
3. Finally, we generate TypeScript types from the OpenAPI 3.0 spec using [openapi-typescript](https://github.com/drwpow/openapi-typescript)

## Workflow

The entire process is automated with npm scripts:

```bash
# Generate everything from scratch
pnpm run generate:api-types

# Or run each step individually:
pnpm run swag                 # Generate Swagger 2.0 docs from Go code
pnpm run convert-to-openapi3  # Convert Swagger 2.0 to OpenAPI 3.0
pnpm run generate-types       # Generate TypeScript types from OpenAPI 3.0
```

## Files

- `openapi3.json` - The converted OpenAPI 3.0 specification
- `schema.d.ts` - The generated TypeScript type definitions

## Using the types

Import types from this module in your TypeScript code:

```typescript
import type { components, paths } from '$lib/api/schema';

// Use component schemas
type User = components['schemas']['User'];

// Use request/response types
type RegisterRequest = paths['/auth/register']['post']['requestBody']['content']['application/json'];
type RegisterResponse = paths['/auth/register']['post']['responses']['200']['content']['application/json'];
```

## Future improvements

In the future, we may want to consider:

1. Switching to a tool that directly generates OpenAPI 3.0 from Go code, such as:
   - [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen)
   - [ogen](https://github.com/ogen-go/ogen)

2. Adding validation to ensure the generated types match the backend implementation 