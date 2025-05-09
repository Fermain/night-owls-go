# Project Health Report & Diagnosis

This document outlines the current state of the `night-owls-go` application, covering its structure, build processes, dependencies, and potential areas for attention.

## I. Project Overview

The `night-owls-go` project consists of two main components:

1.  **Go Backend:** A RESTful API server built with Go, responsible for business logic, data persistence, and serving the frontend application.
2.  **SvelteKit Frontend:** A web application built with SvelteKit and Vite, located in the `app/` directory.

### 1.1. Backend (Go - `night-owls-go` module)

*   **Purpose:** Provides API endpoints for user authentication (OTP-based), managing community watch schedules, bookings, user reports, and web push notifications.
*   **Location:** Project root, primarily within `cmd/` and `internal/` directories.
*   **Key Technologies:**
    *   Go (v1.23.4 as per `go.mod`)
    *   Chi (v5) for HTTP routing
    *   SQLite for the database
    *   `sqlc` for generating type-safe Go code from SQL queries
    *   `golang-migrate` (v4) for database schema migrations
    *   `slog` for structured logging
    *   `robfig/cron` (v3) for background job scheduling (e.g., outbox processing)
    *   `golang-jwt/jwt` (v5) for JWT authentication
    *   `swaggo/swag` for Swagger/OpenAPI documentation generation
*   **Entry Point:** `cmd/server/main.go`
*   **Configuration:** Via `.env` file or environment variables (see `internal/config/config.go`). Defaults exist if `.env` is not present.
*   **Database Interaction:**
    *   Schema migrations: `internal/db/migrations/`
    *   SQL queries for `sqlc`: `internal/db/queries/`
    *   `sqlc` generated code: `internal/db/sqlc_generated/`
*   **Static Asset Serving:** The Go backend is configured to serve static files, presumably the built frontend application, from a directory specified by the `StaticDir` configuration (defaults to `./frontend/dist`).

### 1.2. Frontend (SvelteKit - `app/`)

*   **Purpose:** Provides the user interface for interacting with the backend services.
*   **Location:** `app/` directory.
*   **Key Technologies:**
    *   SvelteKit
    *   Svelte (v5)
    *   Vite
    *   TailwindCSS for styling
    *   TypeScript
    *   Prettier for code formatting
    *   ESLint for linting
    *   Vitest for unit testing
    *   Playwright for end-to-end testing
*   **Package Manager:** `pnpm` (inferred from `pnpm` block in `app/package.json`)
*   **Build Output:** Typically a `dist/` or `build/` directory within `app/`. The `app/build` directory seen in the project structure is likely the output. The exact output is configured in `app/svelte.config.js` and the build script in `app/package.json` (`vite build`).

## II. Initial Diagnostic Checks & Potential Issues

This section will be populated as diagnostic steps are performed.

### 2.1. Backend Health

*   **Dependencies:**
    *   `go mod tidy` completed successfully. Some dependencies were downloaded/updated.
*   **Build:**
    *   `go build ./cmd/server/...` completed successfully.
*   **Database Migrations:**
    *   The application attempts to run migrations on startup. This will be verified when attempting to run the server.
    *   `sqlc generate` (data access layer code generation) completed successfully.
*   **Configuration:**
    *   A `.env` file is required to run the backend. Based on `internal/config/config.go` and `README.md`, it should contain at least:
        *   `SERVER_PORT` (e.g., `8080`)
        *   `DATABASE_PATH` (e.g., `./night-owls.dev.db`)
        *   `JWT_SECRET` (a strong secret)
        *   `OTP_LOG_PATH` (e.g., `./sms_outbox.log`)
        *   `STATIC_DIR="app/build"` (Crucial for serving the frontend correctly. This overrides the default `./frontend/dist`)
        *   `VAPID_PUBLIC`, `VAPID_PRIVATE`, `VAPID_SUBJECT` (for web push, can be placeholders for initial run)
    *   **`StaticDir` Mismatch:** The Go backend's default `StaticDir` (`./frontend/dist`) would not work. The SvelteKit app builds to `app/build/`. The `.env` file **must** set `STATIC_DIR=app/build`.
*   **Runtime:**
    *   Attempting to run the server is the next step, pending creation of the `.env` file by the user with `STATIC_DIR=app/build`.
*   **Tests:**
    *   `go test ./...` **FAILED** due to build errors in test files. This is a critical issue blocking test execution.
        *   **Reason 1: `outbox.NewDispatcherService` signature change:** Test files in `internal/outbox/dispatcher_test.go` and `internal/api/auth_handlers_integration_test.go` are calling `outbox.NewDispatcherService` with an outdated signature (missing the `pushSender *service.PushSender` argument).
            *   **File affected:** `internal/outbox/dispatcher.go` shows the new signature: `func NewDispatcherService(querier db.Querier, smsSender MessageSender, pushSender *service.PushSender, logger *slog.Logger, cfg *config.Config) *DispatcherService`
        *   **Reason 2: `db.Querier` interface change:** Mock queriers in `internal/service/booking_service_test.go` and `internal/service/report_service_test.go` do not implement the new `DeleteSubscription` method from the `db.Querier` interface.
            *   **File affected:** `internal/db/sqlc_generated/querier.go` shows the `Querier` interface now includes `DeleteSubscription(ctx context.Context, arg DeleteSubscriptionParams) error`.
    *   The `golang-migrate` issue mentioned in the README for integration tests (manual SQL execution as workaround) was not directly hit due to these preceding build failures in tests.
*   **API Documentation:**
    *   `swag init -g cmd/server/main.go -o ./docs/swagger` (Swagger/OpenAPI doc generation) completed successfully. A common benign warning about package name in root dir was noted.

### 2.2. Frontend Health (`app/` directory)

*   **Dependencies:**
    *   `pnpm install` completed successfully. Updated/added several packages. A warning about ignored build scripts for `svelte-preprocess` was noted.
*   **Build:**
    *   `pnpm build` (which runs `vite build`) completed successfully.
    *   **Build Output Directory:** Confirmed to be `app/build/`.
    *   A SvelteKit convention warning was noted during the build: `src/routes/+layout.svelte \`export const ssr\` will be ignored — move it to +layout(.server).js/ts instead.` This should be addressed for best practices but did not fail the build.
*   **Linting & Formatting:**
*   **Tests:**

### 2.3. Overall Integration

*   **Frontend asset serving by backend:**
    *   **Confirmed `StaticDir` Mismatch:** The Go backend's default `StaticDir` configuration is `./frontend/dist` (from `internal/config/config.go`). The SvelteKit frontend builds its static assets to `app/build/`. For the Go backend to correctly serve the frontend, the `StaticDir` configuration (via `.env` or direct modification) must be updated to point to `app/build/` (relative to the Go executable, so likely just `app/build` if the Go server runs from the project root).

### 2.4. API Specification (Swagger/OpenAPI)

*   **Current Version:** The project uses `swaggo/swag` to generate API documentation. The `swag init` command successfully generates `docs/swagger/swagger.json`, which is an **OpenAPI 2.0 (Swagger 2.0)** specification, as confirmed by the `"swagger": "2.0"` field in the generated file.
*   **Serving:** The Go backend serves this specification using `github.com/swaggo/http-swagger`, which is compatible with OpenAPI 2.0.
*   **OpenAPI 3.0 Consideration:** If there is an external requirement for an OpenAPI 3.0 specification, the current tooling (`swaggo/swag`) has limited support for it and might require different usage or replacement with a tool that has more robust OpenAPI 3.0 capabilities. The existing Go doc comments are parsed for OpenAPI 2.0.

## III. Recommended Actions & Next Steps

1.  **Create a `.env.example` file:** Based on the current `.env` provided and the application's configuration structure (`internal/config/config.go`), it's highly recommended to create a `.env.example` file in the project root to guide future environment setup. Its content should be:
    ```dotenv
    # Server Configuration
    SERVER_PORT=8080

    # Database Configuration
    # For development, a local SQLite file is typical:
    DATABASE_PATH=./night-owls.dev.db
    # For tests, you might use a different file:
    # DATABASE_PATH=./night-owls.test.db

    # JWT Configuration
    # IMPORTANT: Replace with a strong, randomly generated secret key in production!
    JWT_SECRET=your_strong_random_jwt_secret_here_please_change
    # JWT token expiration in hours (e.g., 24 for 1 day, 168 for 7 days)
    JWT_EXPIRATION_HOURS=24

    # Default duration for shifts if not specified by the schedule itself (in hours)
    # DEFAULT_SHIFT_DURATION_HOURS=2

    # OTP/SMS Mocking & Configuration
    # For development/testing, OTPs can be logged to a file.
    OTP_LOG_PATH=./sms_outbox.log
    # OTP validity period in minutes
    OTP_VALIDITY_MINUTES=5

    # Logging Configuration (Optional)
    # Log level: debug, info, warn, error
    LOG_LEVEL=info
    # Log format: json, text
    LOG_FORMAT=json

    # Outbox Configuration (Optional - for background message processing)
    OUTBOX_BATCH_SIZE=10
    OUTBOX_MAX_RETRIES=3

    # PWA / WebPush Configuration
    # Generate VAPID keys using a command like: npx web-push generate-vapid-keys --json
    # The VAPID_SUBJECT should be a URL or mailto URI for your application/contact.
    VAPID_PUBLIC=
    VAPID_PRIVATE=
    VAPID_SUBJECT=mailto:admin@example.com

    # Static File Directory for Frontend
    # CRITICAL: This path MUST point to the build output of your SvelteKit frontend.
    # For this project, the SvelteKit app (in app/) builds to "app/build/".
    STATIC_DIR=app/build
    ```
2.  **Correct `STATIC_DIR` in your active `.env` file:** Ensure your local `.env` file has `STATIC_DIR=app/build` instead of `./frontend/dist` to allow the Go backend to serve the SvelteKit frontend correctly.
3.  **Fix Go Test Failures:** Address the build errors in the Go tests:
    *   Update all calls to `outbox.NewDispatcherService` in test files (`internal/outbox/dispatcher_test.go`, `internal/api/auth_handlers_integration_test.go`) to include the `pushSender *service.PushSender` argument (e.g., passing `nil` or a mock if appropriate for the test context).
    *   Update mock querier implementations (e.g., `MockBookingQuerier` in `internal/service/booking_service_test.go`, `MockReportQuerier` in `internal/service/report_service_test.go`) to include the `DeleteSubscription(ctx context.Context, arg db.DeleteSubscriptionParams) error` method to satisfy the `db.Querier` interface. The mock can return `nil` or a pre-defined error as needed for test cases.
4.  **Address SvelteKit Build Warning:** In `app/src/routes/+layout.svelte`, move the `export const ssr` to a `+layout.server.js` or `+layout.server.ts` file as per SvelteKit conventions. This is for best practice and future compatibility.
5.  **Run Full Test Suites:** After fixing Go test build errors, run:
    *   Go tests: `go test ./...` (from project root)
    *   Frontend tests: `cd app && pnpm test`
6.  **Attempt to Run the Application:** With a corrected `.env` file, try running the backend: `go run ./cmd/server/main.go`. Verify that:
    *   Database migrations run successfully on startup.
    *   The server starts and serves the API (check Swagger UI at `http://localhost:[SERVER_PORT]/swagger/index.html`).
    *   The frontend is served correctly from `http://localhost:[SERVER_PORT]/`.
7.  **Review Frontend `svelte-preprocess` warning:** Investigate the `pnpm install` warning: `Ignored build scripts: svelte-preprocess`. While not currently breaking the build, understand if any action (like `pnpm approve-builds`) is needed for `svelte-preprocess` if it's essential for specific Svelte features being used.

## IV. Frontend Toolchain Upgrade & Verification

### 4.1. TanStack Svelte Query & Devtools

*   **Objective:** Integrate TanStack Svelte Query for server state management and its devtools for easier debugging.
*   **Actions Taken:**
    *   Verified Svelte 5 is listed in `app/package.json`.
    *   Checked `app/components.json`, confirming `shadcn-svelte@next` was already initialized.
    *   Installed `@tanstack/svelte-query` and `@tanstack/svelte-query-devtools` using `pnpm add` in the `app/` directory.
    *   Modified `app/src/routes/+layout.svelte` to:
        *   Import `QueryClient` and `QueryClientProvider` from `@tanstack/svelte-query`.
        *   Instantiate `QueryClient`.
        *   Wrap the main `<slot />` with `<QueryClientProvider>`.
        *   Import `SvelteQueryDevtools` from `@tanstack/svelte-query-devtools` and `dev` from `$app/environment`.
        *   Conditionally render `<SvelteQueryDevtools initialIsOpen={false} />` inside the provider if `dev` is true.
    *   Modified `app/src/routes/+page.svelte` to:
        *   Import `createQuery` from `@tanstack/svelte-query`.
        *   Refactor the existing `fetchSchedules` logic to use `createQuery` to fetch data from the `/schedules` endpoint.
        *   Update the template to display loading, error, and data states from the query object.
    *   Created `docs/frontend/README.md` and populated it with details about SvelteKit, Svelte 5, TanStack Svelte Query (including setup and basic usage examples for `createQuery` and devtools), `shadcn-svelte`, and styling with TailwindCSS.
*   **Verification:**
    *   Ran `pnpm build` in `app/` directory: Build completed successfully.
    *   Ran the Go backend server (`go run ./cmd/server/main.go`).
    *   Ran the SvelteKit development server (`cd app && pnpm dev`).
    *   **Result:** Navigating to the application in the browser confirmed:
        *   The page loaded correctly.
        *   Schedule data was successfully fetched and displayed via TanStack Svelte Query (e.g., "Successfully fetched 2 schedule(s). First schedule name: Summer Patrol (Nov-Apr)").
        *   The TanStack Query Devtools were available and functioning (though user indicated this was not a primary concern for them at the moment, its integration is complete).
*   **Status:** TanStack Svelte Query and Devtools integration is **complete and verified.**

---
*Report generated by AI Assistant on $(date -Iseconds)* 