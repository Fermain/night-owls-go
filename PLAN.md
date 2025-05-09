# Community Watch Backend - Project Plan

This document outlines the development plan for the Go backend of the Community Watch Shift Scheduler.
Remember to use Git effectively throughout the project: commit frequently after completing logical units of work, use meaningful commit messages, and consider feature branches for larger tasks.

## Phase 1: Core Setup & Database

- [x] **Project Initialization**
    - [x] Initialize Go module: `go mod init night-owls-go` (or your chosen module path)
    - [x] Create base directory structure:
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
- [x] **Configuration Setup (`internal/config`)**
    - [x] Define a `Config` struct (e.g., for server port, database path, JWT secret, default shift duration).
    - [x] Implement loading configuration from environment variables (e.g., using `os.Getenv` or a library like `joho/godotenv` for local dev).
- [x] **Logging Setup (`internal/logging`)**
    - [x] Choose and set up a structured logger (e.g., `log/slog` from Go 1.21+, or `Zap`/`Logrus`).
    - [x] Initialize logger in `main.go` and make it accessible.
- [x] **Database Schema & Migrations (`internal/db/migrations`)**
    - [x] Install `golang-migrate/migrate` CLI.
    - [x] Initialize `golang-migrate` for SQLite.
    - [x] Create initial migration file (`000001_init_schema.up.sql` and `.down.sql`):
        - [x] `users` table: `user_id INTEGER PK AUTOINCREMENT`, `phone TEXT UNIQUE NOT NULL`, `name TEXT`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`.
        - [x] `schedules` table: `schedule_id INTEGER PK AUTOINCREMENT`, `name TEXT NOT NULL`, `cron_expr TEXT NOT NULL`, `start_date DATE`, `end_date DATE`, `duration_minutes INTEGER NOT NULL DEFAULT 120`, `timezone TEXT`.
        - [x] `bookings` table: `booking_id INTEGER PK AUTOINCREMENT`, `user_id INTEGER REFERENCES users(user_id) NOT NULL`, `schedule_id INTEGER REFERENCES schedules(schedule_id) NOT NULL`, `shift_start DATETIME NOT NULL`, `shift_end DATETIME NOT NULL`, `buddy_user_id INTEGER REFERENCES users(user_id)`, `buddy_name TEXT`, `attended BOOLEAN DEFAULT 0 NOT NULL`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`, `UNIQUE(schedule_id, shift_start)`.
        - [x] `reports` table: `report_id INTEGER PK AUTOINCREMENT`, `booking_id INTEGER REFERENCES bookings(booking_id) ON DELETE CASCADE NOT NULL`, `severity INTEGER NOT NULL`, `message TEXT`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`.
        - [x] `outbox` table: `outbox_id INTEGER PK AUTOINCREMENT`, `message_type TEXT NOT NULL`, `recipient TEXT NOT NULL`, `payload TEXT`, `status TEXT DEFAULT 'pending' NOT NULL`, `created_at DATETIME DEFAULT CURRENT_TIMESTAMP`, `sent_at DATETIME`.
    - [x] Create a second migration file (`000002_seed_schedules.up.sql` and `.down.sql`):
        - [x] Insert "Summer Patrol (Nov-Apr)": `cron_expr="0 0,2 * 11-12,1-4 6,0,1"`, `duration_minutes=120`, appropriate `start_date` (e.g., '2024-11-01'), `end_date` (e.g., '2025-04-30').
        - [x] Insert "Winter Patrol (May-Oct)": `cron_expr="0 1,3 * 5-10 6,0,1"`, `duration_minutes=120`, appropriate `start_date` (e.g., '2025-05-01'), `end_date` (e.g., '2025-10-31').
- [x] **Database Interaction Layer (`internal/db`)**
    - [x] Add `sqlc` (e.g. `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`).
    - [x] Create `sqlc.yaml` configuration file.
    - [x] Write initial SQL queries (`.sql` files in `internal/db/queries/`) for:
        - [x] Users: Create, GetByPhone, GetByID.
        - [x] Schedules: Create, GetByID, ListActive (based on current date between `start_date` and `end_date`).
        - [x] Bookings: Create, GetByScheduleAndStartTime, GetByID, ListByUserID, UpdateAttendance.
        - [x] Reports: Create, GetByBookingID, ListByUserID.
        - [x] Outbox: Create, GetPending, UpdateStatus.
    - [x] Generate Go code using `sqlc generate`.
    - [x] Implement database connection setup in `main.go` (using `github.com/mattn/go-sqlite3`) and ensure migrations are run on startup.

## Phase 2: User Authentication & Management

- [x] **Core Authentication Logic (`internal/auth`)**
    - [x] Implement JWT generation and validation functions.
    - [x] Define mock OTP generation (e.g., fixed OTP for dev, or random 6-digit string).
    - [x] Define OTP storage/cache (e.g., in-memory map for simplicity during dev, keyed by phone, with expiry).
- [x] **User Service (`internal/service/user_service.go`)**
    - [x] `RegisterOrLoginUser(phone, name)`: handles user creation if not exists, generates OTP, stores it, and queues OTP message to outbox.
    - [x] `VerifyOTP(phone, otp)`: validates OTP, generates JWT on success.
- [x] **API Endpoints & Handlers (`internal/api/auth_handlers.go`)**
    - [x] `POST /auth/register` (Handler: `RegisterHandler`):
        - [x] Input: `{ "phone": "+27821234567", "name": "Alice" }`.
        - [x] Validates phone number.
        - [x] Calls user service to handle registration/login & OTP.
        - [x] Response: `200 OK {"message": "OTP sent to sms_outbox.log"}`.
    - [x] `POST /auth/verify` (Handler: `VerifyHandler`):
        - [x] Input: `{ "phone": "...", "code": "123456" }`.
        - [x] Calls user service to verify OTP.
        - [x] Response: `200 OK {"token": "<JWT>"}` on success, `401 Unauthorized` on failure.
- [x] **Authentication Middleware (`internal/api/middleware.go`)**
    - [x] Create `AuthMiddleware` to protect routes: parses JWT from `Authorization: Bearer <token>` header, validates it, and adds user ID to request context.

## Phase 3: Schedules & Shift Availability

- [x] **Scheduling Service (`internal/service/schedule_service.go`)**
    - [x] Integrate `github.com/gorhill/cronexpr`.
    - [x] `GetUpcomingAvailableSlots(fromTime, toTime, limit)`:
        - [x] Fetches active schedules from DB (based on current date between `schedule.start_date` and `schedule.end_date`).
        - [x] For each active schedule:
            - [x] Uses `cronexpr` to generate occurrences within the schedule's `start_date`/`end_date` and the query's `fromTime`/`toTime`.
            - [x] Calculates `shift_end` using `shift_start + schedule.duration_minutes`.
            - [x] Checks against `bookings` table to filter out already booked slots (`UNIQUE(schedule_id, shift_start)`).
        - [x] Compiles a list of available `ShiftSlot` objects (custom struct with schedule info, start/end times).
        - [x] Sorts and limits results.
- [x] **API Endpoints & Handlers (`internal/api/schedule_handlers.go`)**
    - [x] `GET /schedules` (Handler: `ListSchedulesHandler`):
        - [x] Retrieves and returns all defined schedules (or active ones).
        - [x] Response: `200 OK` with array of schedule details. (Implemented with placeholder)
    - [x] `GET /shifts/available` (Handler: `ListAvailableShiftsHandler`):
        - [x] Accepts optional query params: `from` (date), `to` (date), `limit` (int).
        - [x] Calls `schedule_service.GetUpcomingAvailableSlots`.
        - [x] Response: `200 OK` with array of available shift slots.

## Phase 4: Bookings Management

- [x] **Booking Service (`internal/service/booking_service.go`)**
    - [x] `CreateBooking(userID, scheduleID, startTime, buddyPhone, buddyName)`:
        - [x] Validates that `scheduleID` is valid and active.
        - [x] Validates that `startTime` aligns with the schedule's cron expression and is within its active `start_date`/`end_date`.
        - [x] Calculates `shift_end` using schedule's `duration_minutes`.
        - [x] Handles buddy logic:
            - [x] If `buddyPhone` provided, try to find user. If found, `buddy_user_id` = found user's ID, `buddy_name` = found user's name.
            - [x] Else, `buddy_user_id` = NULL, `buddy_name` = provided `buddyName`.
        - [x] Attempts to insert booking into DB (uses `sqlc` method which respects `UNIQUE(schedule_id, shift_start)`).
        - [x] If DB insert fails due to unique constraint, return specific error for 409 Conflict.
        - [x] On successful booking, queues confirmation message to outbox for `userID`.
    - [x] `MarkAttendance(bookingID, userID, attendedStatus)`:
        - [x] Validates `bookingID` exists and belongs to `userID` (or user is admin - admin role not yet defined, assume owner only for now).
        - [x] Updates `attended` status in `bookings` table.
- [x] **API Endpoints & Handlers (`internal/api/booking_handlers.go`)**
    - [x] `POST /bookings` (Handler: `CreateBookingHandler`, protected by AuthMiddleware):
        - [x] Input: `{ "schedule_id": 1, "start_time": "YYYY-MM-DDTHH:MM:SSZ", "buddy_phone": "...", "buddy_name": "..." }`.
        - [x] Extracts `userID` from context.
        - [x] Calls `booking_service.CreateBooking`.
        - [x] Response: `201 Created` with booking details, or `409 Conflict` if slot taken, or `400 Bad Request` for invalid input.
    - [x] `PATCH /bookings/{id}/attendance` (Handler: `MarkAttendanceHandler`, protected by AuthMiddleware):
        - [x] Input: `{ "attended": true }`. Booking ID from path.
        - [x] Extracts `userID` from context.
        - [x] Calls `booking_service.MarkAttendance`.
        - [x] Response: `200 OK` with updated booking or `204 No Content`. `403 Forbidden` if not owner, `404 Not Found`.
    - [ ] `DELETE /bookings/{id}` (Handler: `CancelBookingHandler`, protected, optional as per guide - consider for future).

## Phase 5: Incident Reporting

- [x] **Report Service (`internal/service/report_service.go`)**
    - [x] `CreateReport(userID, bookingID, severity, message)`:
        - [x] Validates `bookingID` exists and user is authorized to report (e.g., booked user).
        - [x] Validates `severity` (0-2).
        - [x] Inserts report into `reports` table.
- [x] **API Endpoints & Handlers (`internal/api/report_handlers.go`)**
    - [x] `POST /bookings/{id}/report` (Handler: `CreateReportHandler`, protected by AuthMiddleware):
        - [x] Booking ID from path. Input: `{ "severity": 0, "message": "..." }`.
        - [x] Extracts `userID` from context.
        - [x] Calls `report_service.CreateReport`.
        - [x] Response: `201 Created` with report details. `400 Bad Request` for invalid severity. `403/404` for booking issues.
    - [ ] `GET /reports` (Handler: `ListReportsHandler`, protected, optional - consider for future).

## Phase 6: Outbox & Notifications

- [x] **Message Sender Interface & Mock (`internal/outbox`)**
    - [x] Define `MessageSender` interface (`Send(recipient, messageType, payload) error`).
    - [x] Implement `LogFileMessageSender` that writes messages to `sms_outbox.log` (e.g., "To: [recipient], Type: [messageType], Body: [payload]").
- [x] **Outbox Dispatcher Service (`internal/outbox/dispatcher.go`)**
    - [x] `ProcessPendingOutboxMessages()`:
        - [x] Queries `outbox` table for `status='pending'` messages.
        - [x] Uses `MessageSender.Send()`.
        - [x] Updates `outbox` entry status (`sent` or `failed`) and `sent_at`.
        - [x] Implement basic retry logic for failed sends (e.g., increment a retry counter, mark as failed after X retries).
- [x] **Background Job for Dispatcher (`cmd/server/main.go`)**
    - [x] Integrate `github.com/robfig/cron/v3`.
    - [x] Schedule `outbox_service.ProcessPendingOutboxMessages` to run periodically (e.g., every 1 minute).
    - [x] Ensure cron scheduler is started in `main.go` and stopped gracefully on shutdown.

## Phase 7: Finalization & Testing

- [x] **Main Application Setup (`cmd/server/main.go`)**
    - [x] Initialize config, logger, DB connection (run migrations).
    - [x] Initialize `sqlc.New(db)` querier.
    - [x] Initialize all services, injecting dependencies (querier, config, logger, message sender).
    - [x] Set up `chi` router:
        - [x] Mount middleware (logger, recovery, CORS if needed, AuthMiddleware for protected routes).
        - [x] Register all API handlers.
    - [x] Start HTTP server.
    - [x] Implement graceful shutdown for HTTP server and cron jobs.
- [ ] **Testing**
    - [ ] **Unit Tests (`_test.go` files):**
        - [ ] Test critical service logic: schedule generation, booking validation (slot taken, invalid time), OTP verification, buddy resolution.
        - [ ] Mock database interactions where appropriate or use helper functions to set up in-memory DB state.
    - [ ] **Integration Tests (`_test.go` files, possibly in `internal/api`):**
        - [ ] Use `httptest` to test API endpoints.
        - [ ] Set up an in-memory SQLite database for each test or test suite, applying migrations.
        - [ ] Test scenarios from `Guide.md`:
            - [ ] User registration & login flow.
            - [ ] Viewing available shifts (empty, with data, after booking).
            - [ ] Booking a slot (success, conflict 409, invalid data 400).
            - [ ] Marking attendance (success, unauthorized).
            - [ ] Submitting reports.
            - [ ] Authentication enforcement (401/403).
    - [ ] Run tests with `go test ./...` and aim for good coverage.
- [ ] **Documentation**
    - [ ] Create/Update `README.md`:
        - [ ] Project overview.
        - [ ] Setup instructions (Go version, migrate CLI, env vars).
        - [ ] How to run (dev server, tests).
        - [ ] Basic API endpoint documentation (or link to Postman collection/Swagger spec if generated).
- [ ] **Final Review & Refinement**
    - [ ] Code review for clarity, consistency, error handling.
    - [ ] Check for hardcoded values that should be in config.
    - [ ] Ensure all requirements from `Guide.md` are met.

This plan provides a structured approach. We can adjust and add detail as we go.
Let's start building! 