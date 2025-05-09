# Night Owls Go - Community Watch Shift Scheduler Backend

This project is a Go backend service for managing volunteer shifts for a community watch program.
It provides a RESTful API for user authentication (passwordless via OTP), viewing available shifts, booking shifts (with an optional buddy), marking attendance, and submitting incident reports.

Key features include:
- Recurring shifts defined by Cron expressions with seasonal validity.
- Ephemeral shifts: shift instances are only created in the database upon booking.
- Transactional outbox pattern for reliable (mocked) SMS/notification dispatch.
- API documentation with Swagger/OpenAPI
- Converter-based approach for transforming DB models to API responses

## Tech Stack

- **Go** (version 1.21+ recommended)
- **SQLite** for the database
- **Chi** for HTTP routing
- **sqlc** for type-safe SQL query generation
- **golang-migrate** for database migrations
- **slog** for structured logging
- **robfig/cron** for background job scheduling (outbox dispatcher)
- **golang-jwt/jwt** for JWT handling
- **swaggo/swag** for Swagger/OpenAPI documentation

## Setup Instructions

### Prerequisites

1.  **Go:** Ensure you have Go installed (version 1.21 or later is recommended).
    See [golang.org/dl/](https://golang.org/dl/).
2.  **golang-migrate CLI:** Install the CLI tool for running database migrations.
    ```bash
    go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
    ```
    Ensure `$GOPATH/bin` or `$HOME/go/bin` is in your system's `PATH`.
3.  **sqlc CLI:** Install the CLI tool for generating Go code from SQL.
    ```bash
    go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
    ```
4.  **swag CLI:** Install the CLI tool for generating Swagger documentation.
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```

### Project Setup

1.  **Clone the repository (if applicable) or ensure you are in the project root.**
2.  **Install Go dependencies:**
    ```bash
    go mod tidy
    ```
3.  **Configuration via `.env` file:**
    Create a `.env` file in the project root by copying `.env.example` (if one exists) or by creating it manually.
    This file is used for local development. In production, environment variables should be set directly.

    Example `.env` content:
    ```env
    # Server Configuration
    SERVER_PORT=8080

    # Database Configuration
    DATABASE_PATH=./community_watch.db

    # JWT Configuration
    JWT_SECRET=a_very_secure_secret_for_development_only_change_this
    # DEFAULT_SHIFT_DURATION_HOURS=2 # Optional, defaults to 2 hours

    # OTP/SMS Mocking Configuration
    OTP_LOG_PATH=./sms_outbox.log
    ```
    *Note: `JWT_SECRET` should be a strong, unique secret in a production environment.*

### Database Migrations

Database migrations are managed using `golang-migrate` and SQL files located in `internal/db/migrations/`.

-   **To apply all pending up migrations:**
    The application attempts to run migrations automatically on startup using the `DATABASE_PATH` from the config.
    Alternatively, you can run them manually using the `migrate` CLI (ensure `DATABASE_PATH` in your `.env` points to the correct file location relative to where you run the `migrate` command, or use an absolute path):
    ```bash
    # Ensure DATABASE_PATH in .env is correct if it's a relative path
    # Or provide the database URL directly:
    migrate -database "sqlite3://$(grep DATABASE_PATH .env | cut -d '=' -f2)" -path internal/db/migrations up
    ```

-   **To roll back the last migration:**
    ```bash
    migrate -database "sqlite3://$(grep DATABASE_PATH .env | cut -d '=' -f2)" -path internal/db/migrations down 1
    ```

-   **To create a new migration:**
    ```bash
    migrate create -ext sql -dir internal/db/migrations -seq <migration_name>
    ```

### SQLC Code Generation

If you modify any SQL queries in `internal/db/queries/` or change the schema (and create new migrations), you need to regenerate the Go code for `sqlc`:

```bash
sqlc generate
```

### Swagger Documentation Generation

To generate or update the Swagger documentation after making API changes:

```bash
swag init -g cmd/server/main.go -o ./docs/swagger
```

## How to Run

### Development Server

To start the backend server for development:

```bash
go run ./cmd/server/main.go
```

The server will start (by default on port 8080, as per config) and apply database migrations if needed.
Output, including mock OTPs and other notifications, will be logged to the console (structured JSON) and potentially to `sms_outbox.log` (as per `OTP_LOG_PATH` config).

Swagger API documentation will be available at: http://localhost:8080/swagger/index.html

### Running Tests

To run all tests:

```bash
go test ./...
```

## API Endpoints

The API is fully documented using Swagger/OpenAPI. When the server is running, you can access the interactive documentation at: http://localhost:8080/swagger/index.html

### Authentication

-   `POST /auth/register`
    -   Registers a new user or initiates login by sending an OTP.
    -   Request: `{ "phone": "+27...", "name": "Optional Name" }`
    -   Response: `200 OK { "message": "OTP sent..." }`
-   `POST /auth/verify`
    -   Verifies the OTP and returns a JWT.
    -   Request: `{ "phone": "+27...", "code": "123456" }`
    -   Response: `200 OK { "token": "<jwt_token>" }`

### Schedules & Shifts

-   `GET /schedules`
    -   Lists defined shift schedules.
-   `GET /shifts/available`
    -   Lists upcoming available shift slots.
    -   Query Params: `from` (RFC3339), `to` (RFC3339), `limit` (int)
    -   Response: `200 OK` with array of `AvailableShiftSlot` objects.

### Bookings (Protected - Requires JWT)

-   `POST /bookings`
    -   Books a shift slot.
    -   Request: `{ "schedule_id": 1, "start_time": "YYYY-MM-DDTHH:MM:SSZ", "buddy_phone": "...", "buddy_name": "..." }`
    -   Response: `201 Created` with booking details, or `409 Conflict` / `400 Bad Request`.
-   `PATCH /bookings/{id}/attendance`
    -   Marks attendance for a booking.
    -   Request: `{ "attended": true }`
    -   Response: `200 OK` with updated booking.

### Reports (Protected - Requires JWT)

-   `POST /bookings/{id}/report`
    -   Submits an incident report for a booking.
    -   Request: `{ "severity": 0-2, "message": "..." }`
    -   Response: `201 Created` with report details.

## Project Architecture

The project follows a clean architecture approach with the following layers:

1. **API Layer (`internal/api/`)**: HTTP handlers, middleware, and request/response models
2. **Service Layer (`internal/service/`)**: Business logic and orchestration
3. **Data Layer (`internal/db/`)**: Database queries and models

The project uses a converter-based approach for transforming database models to API responses, ensuring clean separation between internal data structures and external API contracts.

## Future Enhancements

For a comprehensive list of potential future enhancements to the project, see the [ENHANCEMENTS.md](docs/ENHANCEMENTS.md) document. This includes:

- Security improvements (rate limiting, JWT enhancements)
- Error handling and observability enhancements
- Performance optimizations
- Architecture improvements
- Testing improvements

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Troubleshooting / Known Issues

### `golang-migrate` "unknown driver" in Tests

During the development of integration tests (specifically in `internal/api/*_test.go`), we encountered persistent "unknown driver 'sqlite3'" (and later "unknown driver 'sqlite'") errors when trying to use `golang-migrate/migrate/v4` programmatically via `migrate.New()` to set up an in-memory SQLite database, despite having the correct blank imports for the database and source drivers (e.g., `_ "github.com/golang-migrate/migrate/v4/database/sqlite"`).

This issue occurred even after trying CGo flags, different relative paths for migration files, and ensuring `go mod tidy` was run. The root cause seems to be related to how the `golang-migrate` library registers its drivers or how those registrations are picked up within the specific `go test` binary compilation and execution context for these test packages.

**Workaround for Integration Tests:**
To ensure reliable database schema setup for integration tests, the `newTestApp` helper function in `internal/api/auth_handlers_integration_test.go` (and potentially other future integration test files) manually reads and executes the `.up.sql` migration files directly on the `*sql.DB` connection. This bypasses the programmatic use of `golang-migrate` for test database setup, resolving the driver registration issue in that context.

The main application (`cmd/server/main.go`) uses `golang-migrate` programmatically without issue, as does the `migrate` CLI tool. 