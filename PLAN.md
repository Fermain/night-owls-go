# Community Watch Backend - Project Plan

This document outlines the development plan for the Go backend of the Community Watch Shift Scheduler.
Remember to use Git effectively throughout the project: commit frequently after completing logical units of work, use meaningful commit messages, and consider feature branches for larger tasks.

## Phase 1: Core Setup & Database

- [ ] **Project Initialization**
    - [ ] Initialize Go module: `go mod init night-owls-go` (or your chosen module path)
    - [ ] Create base directory structure:
        ```
        night-owls-go/
        ├── cmd/
        │   └── server/
        │       └── main.go
        ├── internal/
        │   ├── api/
        │   ├── auth/
        │   ├── config/
        │   ├── core/      # (for core domain types if needed)
        │   ├── db/
        │   │   ├── migrations/
        │   │   └── queries/   # (for sqlc .sql files)
        │   ├── logging/
        │   ├── outbox/
        │   └── service/
        ├── pkg/             # (optional, if any truly reusable packages emerge)
        ├── go.mod
        ├── go.sum
        ├── PLAN.md
        └── Guide.md
        ```
- [ ] **Configuration Setup (`internal/config`)**
    - [ ] Define a `Config` struct (e.g., for server port, database path, JWT secret, default shift duration).
    - [ ] Implement loading configuration from environment variables (e.g., using `os.Getenv` or a library like `joho/godotenv` for local dev).
- [ ] **Logging Setup (`internal/logging`)**
    - [ ] Choose and set up a structured logger (e.g., `log/slog` from Go 1.21+, or `Zap`/`Logrus`).
    - [ ] Initialize logger in `main.go` and make it accessible.
- [ ] **Database Schema & Migrations (`internal/db/migrations`)**
    - [ ] Install `golang-migrate/migrate` CLI.
    - [ ] Initialize `golang-migrate` for SQLite.
    - [ ] Create initial migration file (`000001_init_schema.up.sql` and `.down.sql`):
        - [ ] `users` table: `user_id INTEGER PK AUTOINCREMENT`, `phone TEXT UNIQUE NOT NULL`, `name TEXT`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`.
        - [ ] `schedules` table: `schedule_id INTEGER PK AUTOINCREMENT`, `name TEXT NOT NULL`, `cron_expr TEXT NOT NULL`, `start_date DATE`, `end_date DATE`, `duration_minutes INTEGER NOT NULL DEFAULT 120`, `timezone TEXT`.
        - [ ] `bookings` table: `booking_id INTEGER PK AUTOINCREMENT`, `user_id INTEGER REFERENCES users(user_id) NOT NULL`, `schedule_id INTEGER REFERENCES schedules(schedule_id) NOT NULL`, `shift_start DATETIME NOT NULL`, `shift_end DATETIME NOT NULL`, `buddy_user_id INTEGER REFERENCES users(user_id)`, `buddy_name TEXT`, `attended BOOLEAN DEFAULT 0 NOT NULL`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`, `UNIQUE(schedule_id, shift_start)`.
        - [ ] `reports` table: `report_id INTEGER PK AUTOINCREMENT`, `booking_id INTEGER REFERENCES bookings(booking_id) ON DELETE CASCADE NOT NULL`, `severity INTEGER NOT NULL`, `message TEXT`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`.
        - [ ] `outbox` table: `outbox_id INTEGER PK AUTOINCREMENT`, `message_type TEXT NOT NULL`, `recipient TEXT NOT NULL`, `payload TEXT`, `status TEXT DEFAULT 'pending' NOT NULL`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`, `sent_at DATETIME`.
    - [ ] Create a second migration file (`000002_seed_schedules.up.sql` and `.down.sql`):
        - [ ] Insert "Summer Patrol (Nov-Apr)": `cron_expr="0 0,2 * 11-12,1-4 6,0,1"`, `duration_minutes=120`, appropriate `start_date` (e.g., '2024-11-01'), `end_date` (e.g., '2025-04-30').
        - [ ] Insert "Winter Patrol (May-Oct)": `cron_expr="0 1,3 * 5-10 6,0,1"`, `duration_minutes=120`, appropriate `start_date` (e.g., '2025-05-01'), `end_date` (e.g., '2025-10-31').
- [ ] **Database Interaction Layer (`internal/db`)**
    - [ ] Add `sqlc` (e.g. `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`).
    - [ ] Create `sqlc.yaml` configuration file.
    - [ ] Write initial SQL queries (`.sql` files in `internal/db/queries/`) for:
        - Users: Create, GetByPhone, GetByID.
        - Schedules: Create, GetByID, ListActive (based on current date between `start_date` and `end_date`).
        - Bookings: Create, GetByScheduleAndStartTime, GetByID, ListByUserID, UpdateAttendance.
        - Reports: Create, GetByBookingID, ListByUserID.
        - Outbox: Create, GetPending, UpdateStatus.
    - [ ] Generate Go code using `sqlc generate`.
    - [ ] Implement database connection setup in `main.go` (using `github.com/mattn/go-sqlite3`) and ensure migrations are run on startup.

## Phase 2: User Authentication & Management

- [ ] **Core Authentication Logic (`internal/auth`)**
    - [ ] Implement JWT generation and validation functions.
    - [ ] Define mock OTP generation (e.g., fixed OTP for dev, or random 6-digit string).
    - [ ] Define OTP storage/cache (e.g., in-memory map for simplicity during dev, keyed by phone, with expiry).
- [ ] **User Service (`internal/service/user_service.go`)**
    - [ ] `RegisterOrLoginUser(phone, name)`: handles user creation if not exists, generates OTP, stores it, and queues OTP message to outbox.
    - [ ] `VerifyOTP(phone, otp)`: validates OTP, generates JWT on success.
- [ ] **API Endpoints & Handlers (`internal/api/auth_handlers.go`)**
    - [ ] `POST /auth/register` (Handler: `RegisterHandler`):
        - Input: `{ "phone": "+27821234567", "name": "Alice" }`.
        - Validates phone number.
        - Calls user service to handle registration/login & OTP.
        - Response: `200 OK {"message": "OTP sent to sms_outbox.log"}`.
    - [ ] `POST /auth/verify` (Handler: `VerifyHandler`):
        - Input: `{ "phone": "...", "code": "123456" }`.
        - Calls user service to verify OTP.
        - Response: `200 OK {"token": "<JWT>"}` on success, `401 Unauthorized` on failure.
- [ ] **Authentication Middleware (`internal/api/middleware.go`)**
    - [ ] Create `AuthMiddleware` to protect routes: parses JWT from `Authorization: Bearer <token>` header, validates it, and adds user ID to request context.

## Phase 3: Schedules & Shift Availability

- [ ] **Scheduling Service (`internal/service/schedule_service.go`)**
    - [ ] Integrate `github.com/gorhill/cronexpr`.
    - [ ] `GetUpcomingAvailableSlots(fromTime, toTime, limit)`:
        - Fetches active schedules from DB (based on current date between `schedule.start_date` and `schedule.end_date`).
        - For each active schedule:
            - Uses `cronexpr` to generate occurrences within the schedule's `start_date`/`end_date` and the query's `fromTime`/`toTime`.
            - Calculates `shift_end` using `shift_start + schedule.duration_minutes`.
            - Checks against `bookings` table to filter out already booked slots (`UNIQUE(schedule_id, shift_start)`).
        - Compiles a list of available `ShiftSlot` objects (custom struct with schedule info, start/end times).
        - Sorts and limits results.
- [ ] **API Endpoints & Handlers (`internal/api/schedule_handlers.go`)**
    - [ ] `GET /schedules` (Handler: `ListSchedulesHandler`):
        - Retrieves and returns all defined schedules (or active ones).
        - Response: `200 OK` with array of schedule details.
    - [ ] `GET /shifts/available` (Handler: `ListAvailableShiftsHandler`):
        - Accepts optional query params: `from` (date), `to` (date), `limit` (int).
        - Calls `schedule_service.GetUpcomingAvailableSlots`.
        - Response: `200 OK` with array of available shift slots.

## Phase 4: Bookings Management

- [ ] **Booking Service (`internal/service/booking_service.go`)**
    - [ ] `CreateBooking(userID, scheduleID, startTime, buddyPhone, buddyName)`:
        - Validates that `scheduleID` is valid and active.
        - Validates that `startTime` aligns with the schedule's cron expression and is within its active `start_date`/`end_date`.
        - Calculates `shift_end` using schedule's `duration_minutes`.
        - Handles buddy logic:
            - If `buddyPhone` provided, try to find user. If found, `buddy_user_id` = found user's ID, `buddy_name` = found user's name.
            - Else, `buddy_user_id` = NULL, `buddy_name` = provided `buddyName`.
        - Attempts to insert booking into DB (uses `sqlc` method which respects `UNIQUE(schedule_id, shift_start)`).
        - If DB insert fails due to unique constraint, return specific error for 409 Conflict.
        - On successful booking, queues confirmation message to outbox for `userID`.
    - [ ] `MarkAttendance(bookingID, userID, attendedStatus)`:
        - Validates `bookingID` exists and belongs to `userID` (or user is admin - admin role not yet defined, assume owner only for now).
        - Updates `attended` status in `bookings` table.
- [ ] **API Endpoints & Handlers (`internal/api/booking_handlers.go`)**
    - [ ] `POST /bookings` (Handler: `CreateBookingHandler`, protected by AuthMiddleware):
        - Input: `{ "schedule_id": 1, "start_time": "YYYY-MM-DDTHH:MM:SSZ", "buddy_phone": "...", "buddy_name": "..." }`.
        - Extracts `userID` from context.
        - Calls `booking_service.CreateBooking`.
        - Response: `201 Created` with booking details, or `409 Conflict` if slot taken, or `400 Bad Request` for invalid input.
    - [ ] `PATCH /bookings/{id}/attendance` (Handler: `MarkAttendanceHandler`, protected by AuthMiddleware):
        - Input: `{ "attended": true }`. Booking ID from path.
        - Extracts `userID` from context.
        - Calls `booking_service.MarkAttendance`.
        - Response: `200 OK` with updated booking or `204 No Content`. `403 Forbidden` if not owner, `404 Not Found`.
    - [ ] `DELETE /bookings/{id}` (Handler: `CancelBookingHandler`, protected, optional as per guide - consider for future).

## Phase 5: Incident Reporting

- [ ] **Report Service (`internal/service/report_service.go`)**
    - [ ] `CreateReport(userID, bookingID, severity, message)`:
        - Validates `bookingID` exists and user is authorized to report (e.g., booked user).
        - Validates `severity` (0-2).
        - Inserts report into `reports` table.
- [ ] **API Endpoints & Handlers (`internal/api/report_handlers.go`)**
    - [ ] `POST /bookings/{id}/report` (Handler: `CreateReportHandler`, protected by AuthMiddleware):
        - Booking ID from path. Input: `{ "severity": 0, "message": "..." }`.
        - Extracts `userID` from context.
        - Calls `report_service.CreateReport`.
        - Response: `201 Created` with report details. `400 Bad Request` for invalid severity. `403/404` for booking issues.
    - [ ] `GET /reports` (Handler: `ListReportsHandler`, protected, optional - consider for future).

## Phase 6: Outbox & Notifications

- [ ] **Message Sender Interface & Mock (`internal/outbox`)**
    - [ ] Define `MessageSender` interface (`Send(recipient, messageType, payload) error`).
    - [ ] Implement `LogFileMessageSender` that writes messages to `sms_outbox.log` (e.g., "To: [recipient], Type: [messageType], Body: [payload]").
- [ ] **Outbox Dispatcher Service (`internal/outbox/dispatcher.go`)**
    - [ ] `ProcessPendingOutboxMessages()`:
        - Queries `outbox` table for `status='pending'` messages.
        - For each message, calls `MessageSender.Send()`.
        - Updates message status in `outbox` to `sent` or `failed` (with `sent_at` timestamp).
        - Implement basic retry logic for failed sends (e.g., increment a retry counter, mark as failed after X retries).
- [ ] **Background Job for Dispatcher (`cmd/server/main.go`)**
    - [ ] Integrate `github.com/robfig/cron/v3`.
    - [ ] Schedule `outbox_service.ProcessPendingOutboxMessages` to run periodically (e.g., every 1 minute).
    - [ ] Ensure cron scheduler is started in `main.go` and stopped gracefully on shutdown.

## Phase 7: Finalization & Testing

- [ ] **Main Application Setup (`cmd/server/main.go`)**
    - [ ] Initialize config, logger, DB connection (run migrations).
    - [ ] Initialize `sqlc.New(db)` querier.
    - [ ] Initialize all services, injecting dependencies (querier, config, logger, message sender).
    - [ ] Set up `chi` router:
        - Mount middleware (logger, recovery, CORS if needed, AuthMiddleware for protected routes).
        - Register all API handlers.
    - [ ] Start HTTP server.
    - [ ] Implement graceful shutdown for HTTP server and cron jobs.
- [ ] **Testing**
    - [ ] **Unit Tests (`_test.go` files):**
        - Test critical service logic: schedule generation, booking validation (slot taken, invalid time), OTP verification, buddy resolution.
        - Mock database interactions where appropriate or use helper functions to set up in-memory DB state.
    - [ ] **Integration Tests (`_test.go` files, possibly in `internal/api`):**
        - Use `httptest` to test API endpoints.
        - Set up an in-memory SQLite database for each test or test suite, applying migrations.
        - Test scenarios from `Guide.md`:
            - User registration & login flow.
            - Viewing available shifts (empty, with data, after booking).
            - Booking a slot (success, conflict 409, invalid data 400).
            - Marking attendance (success, unauthorized).
            - Submitting reports.
            - Authentication enforcement (401/403).
    - [ ] Run tests with `go test ./...` and aim for good coverage.
- [ ] **Documentation**
    - [ ] Create/Update `README.md`:
        - Project overview.
        - Setup instructions (Go version, migrate CLI, env vars).
        - How to run (dev server, tests).
        - Basic API endpoint documentation (or link to Postman collection/Swagger spec if generated).
- [ ] **Final Review & Refinement**
    - [ ] Code review for clarity, consistency, error handling.
    - [ ] Check for hardcoded values that should be in config.
    - [ ] Ensure all requirements from `Guide.md` are met.

This plan provides a structured approach. We can adjust and add detail as we go.
Let's start building! 