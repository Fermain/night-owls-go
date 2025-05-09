# Building a Go Backend for a Community Watch Shift Scheduler

## Core Functionality Overview

The shift scheduling system needs to handle recurring volunteer shifts, user bookings, attendance tracking, and incident reporting. Below is a summary of the core requirements and how to approach them:

* **Recurring Shifts with Seasonal Schedules:** Define volunteer shift schedules using Cron expressions to capture recurring times (e.g. daily/weekly shifts). Each schedule can have a *seasonal validity window* – for example, a schedule might only be active during certain months or date ranges (e.g. a “winter schedule” active from May through August). This can be implemented by pairing a Cron expression with a start and end date range for validity. The Cron expression defines the pattern of days/times, and the date range limits when it applies.

* **Ephemeral Shifts (No Pre-Generated Records):** Do **not** pre-store every shift instance in the database. Instead, calculate upcoming shift times on the fly from the schedule definitions (Cron/recurrence rules). A shift **only becomes a booking record once a user volunteers for it.** In other words, shift occurrences are ephemeral until booked. This keeps the database clean – you only persist a shift when someone actually signs up. To implement this, you’ll fetch/generate the next few upcoming shift times from each schedule (within the valid season) and present them to users. When a user books one, you then create a booking entry in the database.

* **User Identification and Auth via Phone (Passwordless):** Volunteers are identified by their phone number. Instead of passwords, use a password-less authentication flow (e.g. one-time codes via SMS or WhatsApp). For example, a user provides their phone number, you generate a one-time code and send it via SMS/WhatsApp, and the user enters the code to verify and log in. This can be implemented by integrating with an SMS API (or WhatsApp Business API), but for development you can **mock** this out. The system should still have endpoints to request an auth code and to verify the code. Once verified, the user is considered authenticated (you can issue a JWT or session token). The phone number serves as the unique user ID (you’ll likely have a Users table keyed by phone).

* **Booking Shifts (with Optional Buddy):** Volunteers can book an available shift slot. When booking, a user may specify a *“buddy”* who will join them on that shift. The buddy can either be another registered user (by reference, e.g. selecting from contacts or entering their phone number) or just free-text (e.g. entering a name if the buddy is not registered). The booking record should capture the buddy information: perhaps store a `buddy_user_id` if the buddy is a registered user, or a `buddy_name` (free text) if not. Only one booking should be allowed per shift slot – once a slot is taken, attempts to book it by someone else should be rejected with a **409 Conflict** error (since it conflicts with the current state of that shift being already booked). In other words, **two people cannot book the same slot** – one will succeed, the second should get a graceful error.

* **Manual Attendance Recording:** After a shift has passed, the volunteer who booked it should manually record whether they (and their buddy, if any) attended. This can be as simple as a checkbox or button in the UI that updates an “attended” flag on the booking. The backend should provide an endpoint to mark a booking as attended (or absent). This data is useful for reporting and accountability but does not need complex logic – just store whether the person showed up. Only the volunteer who booked (or an admin) should be allowed to mark attendance for a booking.

* **Incident Reports (Optional, Linked to Bookings):** The system allows volunteers to file a brief report after their shift if something noteworthy happened. Each report is linked to a specific booking (shift) and has a **severity level** (e.g. 0 = low/no issue, 1 = moderate, 2 = serious) and a text description. Reports are optional – most shifts won’t have a report. If a volunteer submits a report, it should reference the booking ID. In the database, you might have a `Reports` table with fields like `booking_id`, `severity` (0–2), `message`. This could be extended in the future (e.g. for an admin to review reports), but for now it’s a simple attachment to a shift’s record.

**Production considerations:** All core features should be implemented with **robust error handling and input validation**. For example, ensure the Cron expressions for schedules are valid, user phone numbers are unique and properly formatted, a buddy reference is handled correctly (if a buddy user ID is provided, it should exist in the Users table, otherwise reject or treat it as a name). Use appropriate HTTP status codes (e.g. 400 for bad input, 401 for unauthorized, 403 for forbidden actions, 409 for booking conflicts, etc.) and return clear error messages. This makes the API consumer’s job easier and the system behavior more predictable.

## Database Schema Design (SQLite)

Using **SQLite** as the database is a good choice for a lightweight, file-based storage – suitable for development and modest production use. SQLite doesn’t require running a separate DB server, and it supports the SQL features we need. We will design a relational schema for the core entities: Users, Schedules, Bookings, Reports, and an Outbox for messages (for notifications, explained later). Each table will have a primary key and appropriate foreign keys or unique constraints to enforce business rules. Below is a possible schema layout:

* **Users:** Stores volunteer user accounts.

  * `user_id` – primary key (could be an auto-increment integer or the phone number itself if you prefer).
  * `phone` – text, unique (identifies the user, used for login).
  * `name` – text, optional name of the user.
  * `created_at` – timestamp of registration.
  * (If implementing auth codes: you might have a separate table for login OTP codes or a field in Users for a pending OTP, but that can also be handled in memory or cache since SMS is out-of-scope.)

* **Schedules:** Defines recurring shift schedules.

  * `schedule_id` – primary key.
  * `name` – text label (e.g. "Weekday Evening Patrol", "Saturday Night Shift", etc.).
  * `cron_expr` – text, the Cron expression for the recurrence (e.g. `"0 18 * * 5"` for every Friday 6:00pm, assuming a 5-field Cron format for minute/hour/day/etc).
  * `start_date` – date or datetime, when this schedule becomes active (seasonal start).
  * `end_date` – date or datetime, when this schedule stops being active (seasonal end). These fields allow “seasonal” validity; you can ignore occurrences outside this range.
  * `timezone` – text (optional, e.g. `"Africa/Johannesburg"`) if needed to interpret cron times seasonally (this matters if DST shifts times; you might assume a fixed zone).
  * (Alternatively to `start_date`/`end_date`, one could encode month restrictions in the Cron expression or use recurrence rules – see **Schedule math** below – but having explicit fields is simplest.)

* **Bookings:** Stores actual booked shift instances (only created when a user takes a slot).

  * `booking_id` – primary key.
  * `user_id` – foreign key to Users (the volunteer who booked the shift).
  * `schedule_id` – foreign key to Schedules (which schedule template this booking came from).
  * `shift_start` – datetime of the shift **start time** (the exact date and time the shift occurs).
  * `shift_end` – datetime of shift end (optional; could be derived or stored if shifts have a fixed duration).
  * `buddy_user_id` – foreign key to Users, optional (if the buddy is a registered user).
  * `buddy_name` – text, optional (if the buddy was entered as free text and not a registered user).
  * `attended` – boolean or tiny int, default false (whether attendance was confirmed).
  * `created_at` – timestamp when the booking was made.
  * **Unique Constraint:** `(schedule_id, shift_start)` should be unique to prevent two bookings for the same schedule at the same time. This is critical for enforcing that only one person can book a given slot. With SQLite, you can declare `UNIQUE(schedule_id, shift_start)` so any attempt to insert a duplicate will fail (you catch that error to return 409 Conflict).

* **Reports:** Stores incident reports linked to a booking.

  * `report_id` – primary key.
  * `booking_id` – foreign key to Bookings (on delete cascade perhaps, so if a booking is removed, its report is too).
  * `severity` – integer, allowed 0,1,2 (could also enforce via CHECK constraint).
  * `message` – text, the report content.
  * `created_at` – timestamp of report submission.
  * (You might also store who reported it, but since presumably the volunteer who did the shift reports, it could be derived via the booking’s user. If you want, add `user_id` for the reporter for clarity.)

* **Outbox:** (for message dispatch, explained in **Tech Stack** below) A table to queue outgoing messages (SMS/WhatsApp notifications).

  * `outbox_id` – primary key.
  * `message_type` – text (e.g. `"SMS"` or `"WhatsApp"` or maybe `"SHIFT_REMINDER"` or categories).
  * `recipient` – text (e.g. phone number or user id reference).
  * `payload` – text or JSON (the message content or data needed to construct it).
  * `created_at` – timestamp queued.
  * `sent_at` – timestamp when actually sent (null if not sent yet).
  * `status` – text or int (e.g. `"pending"`, `"sent"`, `"failed"`).
  * You would insert into this table within the same transaction as other changes (for example, when a booking is created, queue a “booking confirmation” message to the user). A background task will later read new entries and send out the messages.

**Migrations:** Since we’re using SQLite, we can manage the schema either by writing SQL DDL statements ourselves or using a tool. For a student project, you can keep it simple:

* Create a SQL script for the initial schema (CREATE TABLE statements for all tables with constraints).
* Use a migration tool like **golang-migrate** or just execute the SQL statements on startup if the tables don’t exist. If using an ORM like Ent (which has migration capabilities) or a generator like sqlc (which doesn’t handle migrations), you’ll handle migrations separately.

For example, with plain SQL, your migration SQL might look like:

```sql
CREATE TABLE users (
  user_id    INTEGER PRIMARY KEY AUTOINCREMENT,
  phone      TEXT NOT NULL UNIQUE,
  name       TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE schedules (
  schedule_id INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL,
  cron_expr   TEXT NOT NULL,
  start_date  DATE,
  end_date    DATE,
  timezone    TEXT
);

CREATE TABLE bookings (
  booking_id    INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id       INTEGER NOT NULL REFERENCES users(user_id),
  schedule_id   INTEGER NOT NULL REFERENCES schedules(schedule_id),
  shift_start   DATETIME NOT NULL,
  shift_end     DATETIME,
  buddy_user_id INTEGER REFERENCES users(user_id),
  buddy_name    TEXT,
  attended      BOOLEAN NOT NULL DEFAULT 0,
  created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(schedule_id, shift_start)
);

CREATE TABLE reports (
  report_id   INTEGER PRIMARY KEY AUTOINCREMENT,
  booking_id  INTEGER NOT NULL REFERENCES bookings(booking_id) ON DELETE CASCADE,
  severity    INTEGER NOT NULL,
  message     TEXT,
  created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE outbox (
  outbox_id    INTEGER PRIMARY KEY AUTOINCREMENT,
  message_type TEXT NOT NULL,
  recipient    TEXT NOT NULL,
  payload      TEXT,
  status       TEXT NOT NULL DEFAULT 'pending',
  created_at   DATETIME DEFAULT CURRENT_TIMESTAMP,
  sent_at      DATETIME
);
```

The above is just an illustration of schema; the exact types (e.g. using `DATETIME` vs storing UNIX timestamps) can vary. SQLite is flexible with types (it uses dynamic typing), but sticking to conventions (TEXT for ISO8601 timestamps or INTEGER for Unix epoch) and being consistent will help. With this schema in place, you have the foundation to implement the required features.

## Technology Stack and Libraries

We will use **Go** for the backend, taking advantage of its strong concurrency, standard library, and ecosystem of libraries for database access and scheduling. Below are the key components of the tech stack and how they fit into our project:

* **Go + net/http (REST API):** The service will be a Go program using the standard `net/http` package (or a minimal framework/router like Gin, Echo, or Chi for convenience) to expose a JSON REST API. Go’s standard library is sufficient for a production-ready HTTP server. We will ensure to use context with requests (Go’s `*http.Request` carries context) for cancellation and timeouts where appropriate. JSON encoding/decoding will be done with `encoding/json` or possibly a library like `jsoniter` if needed for performance (probably not necessary here).

* **Database Access – sqlc (or Ent ORM):** For interacting with SQLite, we have options:

  * **sqlc:** A code generation tool that **compiles SQL queries into type-safe Go code**. With sqlc, you write SQL queries (e.g. in a `.sql` file), and it generates Go functions and types for those queries. This gives you the safety of compile-time checks (if your SQL or schema changes, mismatches will cause compile errors) without the complexity of a full ORM. For example, you might write a query `-- name: CreateBooking :one INSERT INTO bookings ... RETURNING booking_id, user_id, ...` and sqlc will generate a `CreateBooking(ctx, params) (Booking, error)` method for you.

  * **Ent (optional alternative):** Ent is an entity framework (an ORM) for Go that uses a **code-first approach** (you define schema in Go structs and it generates query methods). Ent can simplify the definition of relationships and provide an idiomatic API, and it also has migration tools. It’s a powerful option if you prefer working in Go code instead of writing SQL. The trade-off is that Ent has a learning curve and can be overkill if your project is not very complex. Given that the students have looked at these tools, you could choose either. If you want fine control and to write raw SQL, use sqlc. If you prefer an ORM style and auto-generated schema, use Ent. Both will give you type-safe database interactions.

  Regardless of choice, *use context and transactions with the database*. For example, when creating a booking, you will likely open a transaction, perform the insert, and commit. If any step fails, rollback. Both sqlc and Ent support transactions (sqlc just uses `database/sql` under the hood, so you can use `Tx` with it; Ent has a `Tx` type as well).

* **Scheduling & Recurrence Libraries:** The combination of **Cron expressions** and possibly more complex recurrence rules requires some specialized libraries:

  * **cronexpr (github.com/gorhill/cronexpr):** This is a Go library for parsing Cron expressions and getting the next occurrence(s) of the schedule. For example, given a cron string `"0 18 * * FRI"` (meaning Fridays at 18:00) and a start time, `cronexpr` can tell you the next time it occurs. You can use it to iterate multiple future occurrences (by feeding it the last result to get the next). This is perfect for generating upcoming shift times based on our stored `cron_expr`.

    * *Usage Example:*

      ```go
      expr := cronexpr.MustParse(cronString)
      nextTime := expr.Next(time.Now())
      ```

      This yields the next occurrence after now. You could loop to get all occurrences in the next X days.

  * **rrule (github.com/teambition/rrule-go):** This library implements iCalendar recurrence rules (RFC 5545). It’s more powerful than Cron expressions – for example, you can specify rules like “Every Monday and Wednesday until 2024-12-31” or “Every year in June and July on the first Friday” using the RRULE syntax. Under the hood, you can either construct rules with code or parse an RRULE string. We might use rrule-go to handle *seasonal* logic if Cron alone isn’t sufficient. For instance, a “schedule with seasonal validity” could be represented by an RRULE that includes a `BYMONTH` filter or an `UNTIL` date. However, if we already have `start_date` and `end_date` in our Schedule, we might simply post-filter the cron occurrences (e.g., generate upcoming times with cronexpr and skip any outside the valid range). So, **rrule-go is optional** – it’s good to know if you need advanced patterns. The library is efficient and can generate sets of occurrences or next occurrence given a rule. We may include it for extensibility (perhaps future upgrade: allow admin to specify rules in RRULE format).

  * **robfig/cron (github.com/robfig/cron v3):** This is a scheduling library that runs functions on a schedule (essentially a Cron job runner inside your Go app). We will use this to schedule background jobs within our service. For example, we might schedule a nightly task to send reminders to next day’s volunteers, or a periodic task (every few minutes) to check the Outbox table and send pending messages. The usage is straightforward: you create a Cron instance, then add functions with a cron schedule string. For instance:

    ```go
    c := cron.New()
    c.AddFunc("@daily", func(){ sendTomorrowReminders() })  // runs every day at midnight by default
    c.AddFunc("@every 5m", func(){ dispatchOutboxMessages() })  // runs every 5 minutes
    c.Start()
    ```

    The library handles running each job in its own goroutine. It supports standard cron format (seconds optional) and some shortcuts like "@daily". We should be careful to structure these jobs so they don’t conflict and to handle errors (e.g., log if a send fails, etc.). Also, on shutdown, call `c.Stop()` to cleanly stop any scheduled jobs.

* **Message Outbox Pattern for Notifications:** We anticipate sending notifications (such as SMS or WhatsApp messages) for things like login codes, shift reminders, booking confirmations, etc. Directly sending these inside our request handlers can be problematic for reliability – if the external SMS service fails or if we crash after updating the DB but before sending the SMS, we’d have an inconsistency. To solve this, we implement a **transactional outbox pattern**.

  &#x20;*Using an outbox table to reliably send messages after committing data.*

  In this pattern, when an event occurs that requires a notification (e.g. a booking is created), we **insert a record into the Outbox table within the same database transaction** as the booking. If the transaction commits, we have a guaranteed record of the message to send. A separate background process (could be a goroutine launched at startup, or a cron job running every couple minutes) will look at unsent Outbox entries and attempt to send them. After sending (or on failure), it updates the status (and maybe records `sent_at` or an error for retry). This decouples message delivery from the main logic and ensures *at-least-once* delivery of notifications. For our project:

  * We might configure an interface `MessageSender` (with methods like `SendSMS(phone, text)`) and have an implementation that actually calls an API (Twilio, etc). In development, this can be a stub that just logs or marks the outbox entry as sent.
  * The Outbox dispatcher will periodically query for `status='pending'` messages in the outbox and send them. We can use `robfig/cron` to schedule this job, or simply a loop with `time.Sleep()` in a goroutine. Using cron has the advantage of easily adjusting frequency via schedule string.
  * Using SQLite, reading and updating the outbox in a transaction with a `FOR UPDATE` (SQLite doesn’t have `FOR UPDATE` but we can achieve similar effect by serializing dispatches) or simply marking each as in-progress can help avoid two dispatch routines picking the same message. In a single-instance deployment this isn’t a big concern, but it’s worth noting.

  *Example:* When a new booking is created, in the same transaction we insert a row into outbox: `{message_type: "SMS", recipient: "<user phone>", payload: "Thank you for volunteering for <date>...", status: "pending"}`. A cron job every 5 minutes reads pending messages, sends via SMS API, and updates status to "sent". This way, even if the app crashes after booking, the message is in the DB and will be sent when the dispatcher runs next. (For now, since actual SMS integration is out of scope, you might simulate sending by printing to console or writing to a log file).

* **Miscellaneous:** Other notable libraries or components:

  * **HTTP Router:** While you can use Go’s `http.ServeMux`, a third-party router like **Chi** (lightweight, idiomatic), **Gin** or **Echo** (more full-featured) can ease route definitions and middleware. For teaching purposes, using a simple router is fine. The router will map endpoints (like `/api/bookings`) to handler functions.
  * **JSON Library:** Standard library `encoding/json` is fine. If you need to validate JSON inputs, you might write manual checks or use a library like `go-playground/validator` for struct validation tags.
  * **Logging:** Use Go’s `log` or a structured logger (Zap, Logrus) to record important events, errors, and debug info. This is part of production-readiness (e.g. log when a booking is created, or if a cron job fails to send a message).
  * **Configuration:** Manage configuration (like database file path, server port, SMS API keys, etc.) via environment variables or a config file. In production you might use a package like `spf13/viper` for config, but for a small service, simple `os.Getenv` or flags may suffice.

By leveraging these tools and libraries, we ensure our service is **robust, clear, and maintainable**. The use of code-generation (sqlc or Ent) will reduce boilerplate and potential errors, the scheduling libraries will correctly handle recurring times, and the outbox pattern will future-proof our notification system.

## Project Structure and Organization

A well-structured project makes the codebase easier to navigate and maintain. We will use a common Go project layout suitable for a medium-sized service. A possible structure could look like:

```
community-watch-backend/
├── cmd/
│   └── server/           
│       └── main.go         # Application entry point
├── internal/              # Application core packages (not exported to other projects)
│   ├── api/               # HTTP handler logic, request/response models
│   │   ├── bookings.go    # Handlers for booking endpoints
│   │   ├── users.go       # Handlers for auth/user endpoints
│   │   ├── reports.go     # Handlers for report endpoints
│   │   └── ...            # etc.
│   ├── service/           # Business logic layer (optional, could be combined with handlers or db)
│   │   ├── booking.go     # Functions for booking logic (e.g. BookSlot)
│   │   ├── schedule.go    # Schedule calculation logic (cron/rrule utilities)
│   │   └── notification.go# Functions to enqueue notifications to outbox
│   ├── db/                # Database related code (generated or manual queries)
│   │   ├── sqlc.yaml      # sqlc config (if using sqlc)
│   │   ├── queries/       # .sql files for sqlc (if using sqlc)
│   │   ├── models.go      # If using sqlc, generated models; if using Ent, ent generated code in /ent
│   │   └── ...            
│   ├── outbox/            # Outbox dispatcher implementation
│   │   └── dispatcher.go  # background job that sends messages
│   └── migrations/        # SQL migration files or migration logic
├── pkg/ (optional)        # If some packages can be open-sourced or used externally
│   └── ...                # (Not needed if everything is internal for now)
├── go.mod, go.sum         # Go module files
└── README.md              # Documentation for the project
```

**Explanation:**

* **cmd/server/main.go:** The `main.go` will tie everything together. It will parse config (e.g. get the SQLite file path, server port, etc.), open the database (using `sql.Open("sqlite3", ...)`), run migrations if needed (ensuring the latest schema), initialize the services (e.g. if using sqlc, create a `db.Queries` instance; if using Ent, initialize the client), set up the router with all HTTP routes, start any background jobs (cron scheduler for outbox, etc.), and then launch the HTTP server. Essentially, `main.go` orchestrates the startup.

* **internal/api:** This contains HTTP handler functions and route setup. You might have sub-packages per feature (e.g. `internal/api/bookings` package) or one package with multiple files divided by feature. Handlers will parse HTTP requests, call the appropriate business logic, and formulate HTTP responses. Keep handlers thin if possible – any complex logic (like “is this slot available” or “send an SMS”) should be delegated to the service layer or other packages. This improves testability (you can test logic without running an HTTP request) and clarity.

  For example, `bookings.go` might define handlers for:

  * `CreateBookingHandler` (for POST /bookings)
  * `ListAvailableSlotsHandler` (for GET /shifts or /availability)
  * `CancelBookingHandler` (if implemented, for DELETE)
  * `MarkAttendanceHandler` (for POST /bookings/{id}/attendance)

  These handlers would use functions from `internal/service/booking.go` (or directly use `internal/db` queries if logic is simple) to perform the actions.

* **internal/service:** (Optional but recommended) This layer implements the core **business logic** independent of HTTP or DB specifics. For example, a `BookSlot` function might:

  1. Validate the requested slot (check that the time aligns with one of the defined schedule’s occurrences and within the valid range).
  2. Use the `internal/db` package to attempt to insert the booking (within a transaction).
  3. On success, enqueue an outbox message (also in the transaction).
  4. Commit and return the booked record.

  Having this in a service package means you can unit-test it by mocking the DB or using a test DB. It also means if you later add a gRPC interface or CLI, you can reuse the same logic. In simple projects, some merge this with handlers, but separation is a good practice for larger projects.

  The service layer can also house the **schedule computation** logic: e.g., a function `GetAvailableSlots(scheduleIDs, from, to)` that uses the cronexpr/rrule libraries to compute upcoming times for each schedule, then checks against existing bookings (via db queries) to filter out taken slots. This function can be tested independently by seeding some booking data.

* **internal/db:** This contains the persistence logic. If using **sqlc**, you’ll have generated code here (e.g. a struct `Queries` with methods for each SQL query). You might also add custom helper functions or interfaces to wrap the generated code if needed (for example, to abstract which DB driver or for easier mocking). If using **Ent**, you will have an `ent/` directory (likely at the root, not internal) with generated ORM code, and you might not need a separate db package. But you could still have a wrapper package for interfacing with Ent if desired.

  Regardless, the idea is to encapsulate direct DB access in one place. The rest of the app should not compose raw SQL or manage statement preparation – they should call methods here. For example, you might have `db.GetBookingsByUser(userID)` or `db.InsertReport(report)` methods (sqlc will auto-generate these from queries; Ent generates methods from your schema).

* **internal/outbox:** Here you implement the logic for processing the outbox. For instance, `dispatcher.go` might have a function that runs in a loop or is invoked by a cron job. It will query the outbox for unsent messages, attempt to send each (perhaps calling an interface `MessageBroker.Send(message)` – where your implementation of MessageBroker uses an API or logs it), and then mark them sent or failed. You can also include retry logic or a dead-letter concept if a message fails repeatedly. Since the actual sending is external, by abstracting it (interface or even a simple function pointer), you make it easy to test (you can inject a fake sender that just records calls).

* **internal/migrations:** This could be SQL files if using a tool, or if using Ent, you might not use raw SQL files at all (Ent’s code can auto-run schema creation). But having a directory for migrations is good for visibility. In a real project, you might use a tool like `golang-migrate` or `dbmate` to apply migrations. For this project, you might manually ensure the tables exist on startup by executing the CREATE TABLE statements if the file is empty or by versioning the schema.

* **pkg/**: Often in Go projects, `internal` is for private application code, while `pkg` can contain library code intended to be used by other projects. In our case, we likely don’t have such code to expose, so we might not need a `pkg` directory. Everything can live in `internal` to clearly indicate it’s not a public module API.

* **ent/**: If using Ent, you will have an `ent` folder (not in internal by default) with generated code (clients, models) and an `ent/schema` subfolder with your schema definitions (User, Schedule, Booking, etc.). You would then not use `internal/db/sqlc`, but instead call Ent’s client from service or handlers. Ent’s codegen and usage would replace some of the manual work above (for example, Ent will handle creating the SQLite tables if you call `client.Schema.Create(ctx)` for automatic migration, and you’d use Ent’s fluent API to query and mutate data).

**Dependency management:** Use Go modules (Go 1.16+); your `go.mod` will list dependencies like `github.com/mattn/go-sqlite3` for the SQLite driver, `github.com/gorhill/cronexpr`, `github.com/teambition/rrule-go`, `github.com/robfig/cron/v3`, and `github.com/sqlc-dev/sqlc` (actually, for sqlc, you don’t import it at runtime; it’s a build tool generating code). If using Ent, you’d have `entgo.io/ent`. Ensure to pin versions (use tagged versions or commits for stability).

**Code clarity:** Organize code by feature and responsibility. Keep functions short and focused. Use descriptive names. It’s often helpful to define types for your domain objects (even if you have DB models) – for example, a `ShiftSlot` struct that has a `Start time.Time` and maybe a `ScheduleID` could represent an upcoming shift, separate from the DB Booking or Schedule. This can clarify your logic when dealing with “possible shifts” vs “booked shifts.”

**Production readiness considerations:** Even as a student project, structure it as if it were a production service:

* Include a `config` (could be a struct populated from env vars) for things like the listen port, paths, API keys – avoid hard-coding these.
* Implement graceful shutdown of the server (on `os.Interrupt` signal, stop accepting new requests, wait for in-flight to finish, stop cron jobs).
* Use context timeouts for external calls (like the SMS API).
* Validate all inputs coming from outside (never trust client provided data fully).
* Consider concurrency issues: e.g., what if two people try to book the same slot at nearly the same time? We rely on the DB unique constraint to serialize that – one will fail to insert. We should catch that error (SQLite will throw an error for unique constraint violation) and return a 409 Conflict to the user. This is a safe approach in a single-instance deployment. In a multi-instance scenario, you’d need the same unique constraint (still works) and possibly more coordination, but beyond our scope here.

By following this structure, each part of the code has a clear home, and the separation of concerns (HTTP layer vs business logic vs DB vs background jobs) will make the system easier to test and reason about.

## API Design (RESTful JSON)

We will design a RESTful API with JSON request and response bodies. The API will cover user registration/authentication, viewing available shift slots, booking a slot, marking attendance, and submitting reports. Below is a breakdown of key endpoints, their purpose, and data schemas. All responses will be in JSON, and the API should use standard HTTP status codes to indicate success or error conditions.

### User Registration & Authentication

* **POST `/auth/register`** – Register a new user (or initiate login) with a phone number.
  **Request:** JSON body with the phone number (and optionally a name). For example:

  ```json
  { "phone": "+27821234567", "name": "Alice" }
  ```

  **Response:** On success, return 200 OK with a message like `{"message": "OTP sent"}`. Internally, this would generate an OTP code and enqueue an SMS via the outbox. If the phone is already registered (and perhaps already verified), you might either treat it as a login attempt (send OTP anyway) or return 409 if you want to prevent duplicate accounts – but since this is passwordless, it often doesn’t distinguish new vs returning user at OTP stage. For simplicity, assume idempotent: always send an OTP.

  * If phone is malformed, return 400. If SMS service failure (outbox insert fails), return 500 (though with outbox, that shouldn’t fail unless DB down).

* **POST `/auth/verify`** – Verify the one-time code sent to the phone.
  **Request:** JSON `{"phone": "...", "code": "123456"}`.
  **Response:** If the code matches what was sent, return 200 with an authentication token or session info. For example: `{"token": "<JWT or session id>"}` or set a secure HTTP-only cookie. If the code is wrong or expired, 401 Unauthorized.

  * After verifying, if the user was new, you create their account in Users table (if not already created at register time). If using a JWT, include user ID/phone in its claims for subsequent requests.

*(Note: The exact implementation of auth is flexible. The main idea is an OTP flow. In development, you might skip actual SMS and accept a fixed code for testing, given the focus is not on integrating an SMS gateway.)*

### Viewing Available Shifts

* **GET `/shifts/available`** – Retrieve upcoming available shift slots.
  **Query Params:** You could allow filtering by date range or number of occurrences. For example, `/shifts/available?from=2025-05-01&to=2025-05-31` to get May 2025, or `/shifts/available?limit=10` for next 10 slots. If no params, default to next couple of weeks.
  **Response:** JSON array of available slots. Each slot can include the schedule info and the date/time:

  ```json
  [
    {
      "schedule_id": 1,
      "schedule_name": "Weekday Evening Patrol",
      "start_time": "2025-05-10T18:00:00+02:00",
      "end_time":   "2025-05-10T21:00:00+02:00",
      "available": true
    },
    {
      "schedule_id": 2,
      "schedule_name": "Saturday Night Shift",
      "start_time": "2025-05-12T22:00:00+02:00",
      "end_time":   "2025-05-13T02:00:00+02:00",
      "available": true
    }
  ]
  ```

  Only slots that are not yet booked should be listed as available. Under the hood, the server will compute upcoming shifts from all active schedules (using the cron/rrule logic) and subtract those that have a booking. If performance becomes an issue (generating on the fly for a large range), one might generate and cache some slots, but given moderate scale and an ephemeral model, on-the-fly is fine.

  * If no slots in the range, return an empty array.
  * If the client is not authenticated (we may require login to view or not, depending on use case), then 401. But viewing availability could even be public read-access – up to requirements.

* **GET `/schedules`** – (Optional) List the defined schedules (recurrence rules). This could help clients know what kind of shifts exist. For example:
  **Response:**

  ```json
  [
    { "schedule_id": 1, "name": "Weekday Evening Patrol", "cron_expr": "0 18 * * 1-5", "start_date": "2025-05-01", "end_date": "2025-08-31" },
    { "schedule_id": 2, "name": "Saturday Night Shift", "cron_expr": "0 22 * * 6", "start_date": null, "end_date": null }
  ]
  ```

  This endpoint might be read-only and possibly admin-only if you allow adding/editing schedules via API (not required for now – schedules could be preconfigured).

### Booking and Managing Shifts

* **POST `/bookings`** – Book a specific shift slot (volunteer signs up for a shift).
  **Request:** JSON with the desired slot and optional buddy info. For example:

  ```json
  {
    "schedule_id": 1,
    "start_time": "2025-05-10T18:00:00+02:00",
    "buddy_phone": "+27827654321",
    "buddy_name": "Bob"
  }
  ```

  Here we include both a `buddy_phone` and `buddy_name`. The logic could be: if `buddy_phone` matches a registered user, link that user (buddy\_user\_id); otherwise use the `buddy_name` provided. Alternatively, the client might decide to either send a phone or a name. You could make two fields optional: one of them can be provided. Or a simpler way: just have `buddy_name` as free text always, and if the buddy also registers, that’s fine. It’s up to how much linking you want – for now, store what is provided.
  **Response:** On success (slot was free and booking created), return 201 Created with the booking details. For example:

  ```json
  {
    "booking_id": 42,
    "user_id": 7,
    "schedule_id": 1,
    "start_time": "2025-05-10T18:00:00+02:00",
    "buddy_name": "Bob",
    "buddy_user_id": 12,
    "attended": false,
    "created_at": "2025-05-01T10:15:00+02:00"
  }
  ```

  (Here we assumed buddy\_phone was a registered user with id 12, hence both buddy\_name and buddy\_user\_id are set. If it was not a known user, buddy\_user\_id might be null and we just store the name.)

  * **Conflict case:** If the slot is already booked (another booking with same schedule & time exists), the server should return **409 Conflict** to indicate the slot is no longer available. The client can then show a message like “Sorry, that shift was just taken.” This uses the unique constraint at the DB level; catch the error and map to 409. This is analogous to two people trying to buy the last ticket at once – one succeeds, the other gets a conflict.
  * **Validation cases:** If the requested time doesn’t align with the schedule’s cron expression or is outside the valid window, return 400 Bad Request (client is likely using an outdated list or tampered data). If buddy\_user\_id is provided (in an alternative design) and that user doesn’t exist, return 400 or 404. If user is not authenticated, 401.
  * On success, in addition to the HTTP response, our system would have enqueued a confirmation SMS (outbox entry) to notify the volunteer (and possibly the buddy if we have their contact). That detail is transparent to the client.

* **DELETE `/bookings/{id}`** – Cancel a booking. (This was not explicitly required, but in a real system, volunteers might need to cancel if they can’t make it.)
  **Request:** No body needed. The `{id}` in URL is the booking\_id to cancel. Only the user who made the booking (or an admin) should be allowed.
  **Response:** On success, 204 No Content (the booking is removed or marked canceled). If the shift time is very near or in the past, you might disallow cancellation (409 or 400). If not found or not allowed, 404 or 403 respectively.
  *Implementation:* you could either delete the row or add a `canceled` flag in Bookings. Deletion with cascade would also delete any report attached. For simplicity, deletion is fine. Ensure to also possibly enqueue an outbox message to notify others (maybe notify an admin or update an availability board).

* **PATCH `/bookings/{id}`** (or **POST `/bookings/{id}/attendance`**) – Mark attendance for a shift.
  **Request:** Could be a PATCH with `{"attended": true}` or simply a POST with no body to toggle attendance. We’ll assume a PATCH:

  ```json
  { "attended": true }
  ```

  **Response:** 200 OK with updated booking info (or 204 No Content).
  Only the volunteer who booked should be able to mark their attendance (the server should verify the authenticated user matches the booking’s user\_id, else 403 Forbidden). Alternatively, an admin could also mark it in an admin interface.
  If a booking is not yet occurred, you might allow marking only after or during the shift (this could be enforced on client side mostly). If the booking is not found, 404.
  The update is straightforward: set the `attended` flag to true. If you need to record the exact time they marked attendance, you could add an `attended_at` timestamp field.

### Incident Reporting

* **POST `/bookings/{id}/report`** – Submit a report for a shift.
  **Request:** JSON with `severity` and `message` fields, e.g.

  ```json
  { 
    "severity": 2, 
    "message": "Found a broken streetlight on 5th Ave." 
  }
  ```

  **Response:** 201 Created with the created report (or 200 OK). For example:

  ```json
  {
    "report_id": 101,
    "booking_id": 42,
    "severity": 2,
    "message": "Found a broken streetlight on 5th Ave.",
    "created_at": "2025-05-11T22:10:00+02:00"
  }
  ```

  This links the report to booking 42. You should ensure the `{id}` in URL indeed belongs to the authenticated user (if regular volunteers can only report on their own shifts). Alternatively, perhaps an admin could also post a report on a volunteer’s behalf, but likely it’s the volunteer themselves.
  Validate that severity is 0, 1, or 2 (else 400). The message text could have a reasonable length limit. Once stored, this could trigger notifications if severity is high (e.g. send an alert to community leaders if severity 2 – not required now but a consideration).

* **GET `/reports`** – (Optional) if one wants to retrieve reports. Possibly for an admin view to see all recent reports or for a user to see their own past reports. This could allow filtering by user or severity or date. Since it’s optional, we won’t detail it fully, but it would just read from the Reports table.

**General API Notes:**

* All endpoints that modify data (booking, cancelling, marking attendance, reports) should require authentication (the user’s token from the OTP login). Likely use a Bearer JWT in `Authorization` header or a session cookie. The server must verify this and find the user ID for authorization checks.
* Use consistent JSON formatting and field naming (e.g. snake\_case vs camelCase – Go’s `json` tags default to exporting fields in same case as struct field name, so you might define struct fields in CamelCase and tag them like `` `json:"schedule_id"` `` or use a library to convert automatically).
* Error responses: For ease of client handling, you might return a JSON error body with a message, e.g. `{"error": "Slot already booked"}` along with the 409 status. This is nicer than empty body. Similarly 400 errors can explain what was wrong.
* Hypermedia/HATEOAS is not necessary; a simple REST approach is fine.
* Versioning: If this were a public API, you might prefix with `/api/v1/`. For now we can assume it’s v1.
* Testing the API is important: you can write integration tests using `httptest.NewServer` and making HTTP calls to these endpoints to verify they behave as expected (see Testing section below for scenarios).

By designing the endpoints as above, we cover the required functionality. The API is intuitive: clients will first authenticate via phone, then retrieve available slots, then book one, optionally cancel if needed, mark attendance after done, and file a report if something happened. The use of proper HTTP verbs and codes makes it RESTful and clear.

## Testing Strategy and Examples

Testing is crucial to ensure our scheduler backend works correctly and handles edge cases. We will write tests at multiple levels: unit tests for individual functions (like the scheduling logic, or the booking creation), and integration tests for the HTTP API endpoints (possibly using a test database). Here we outline a testing approach and some example test cases in plain English:

**General Testing Approach:**

* Use Go’s `testing` package. Organize tests in files named `_test.go` alongside the code.
* For database-dependent code, set up a *test database*. Since we use SQLite, we can create a temporary SQLite file or use the special `:memory:` database for fast in-memory tests. Apply migrations to it before tests (possibly automatically).
* When testing higher-level functions (service layer), you can use the real SQLite (fast enough) or mock the database by implementing the same interface as your db package (for example, create a fake `db.Queries` that just returns preset data). However, using a real SQLite in-memory gives more realistic coverage.
* For API endpoint tests, spin up the router with handlers (you can use `httptest.NewRecorder` and `http.NewRequest` to simulate HTTP calls). Populate the test DB with necessary data (e.g. create a user, a schedule, etc.) and then simulate requests like “POST /bookings” and check the response code and body.
* Automate tests to run on every build (e.g. via `go test ./...`). Aim for good coverage of critical logic (booking conflicts, schedule calculations, etc.).

**Test Case Ideas (in plain English with expected outcomes):**

1. **Schedule Recurrence Calculation:** Given a schedule with cron expression `"0 18 * * FRI"` (every Friday 6pm) and validity 2025-05-01 to 2025-05-31, the system should compute that the only occurrences are 2025-05-02, 2025-05-09, 2025-05-16, 2025-05-23, 2025-05-30 at 18:00. Verify that the function that generates occurrences returns those dates and none outside May. (This tests cronexpr integration and filtering by date range.)

2. **View Available Slots – No Bookings:** Seed a schedule and no bookings. Call the GET `/shifts/available` endpoint (or directly the underlying function) for the next two weeks. It should list all expected slots as available. For example, if today is May 1 and there is a Fri 6pm schedule, it should show May 2 18:00 as available. Ensure the list is not empty when it should have an occurrence.

3. **View Available Slots – With Bookings:** Now insert a booking for one of the upcoming slots (say user booked May 9 18:00). Call GET `/shifts/available` again. The output should **not** include the May 9 slot (because it’s taken). It might include May 2, May 16, etc., but May 9 is filtered out. This validates that the system checks the Bookings table to exclude booked times.

4. **Book a Slot Successfully:** Simulate an authenticated user (you might bypass the auth by calling the service function directly or set a dummy auth in context). Have them POST `/bookings` for a slot that is currently free. The response should be 201 Created, and the response JSON should contain the booking details including the correct `schedule_id` and `start_time`. Also verify in the database that a new booking row exists with the right data. Additionally, verify that an outbox entry was created for the confirmation SMS (the outbox table should have one new record for this booking’s user). This ensures the booking flow and outbox integration work.

5. **Booking a Taken Slot (Conflict):** Arrange for a slot to already be booked (insert a booking directly or via prior step). Then have another user (or same user, doesn’t matter) attempt to book the *same* schedule and start\_time. The expected result is a **409 Conflict** response. The response body might say something like `"error": "slot already booked"`. In the service logic, if using transactions and catching DB errors, ensure that a unique constraint violation on (schedule\_id, shift\_start) is correctly translated to a 409. *Assertion:* the HTTP status is 409. This matches the idea that only one can succeed in booking a given slot.

6. **Invalid Booking Request (Bad Request):** Test scenarios where the client provides an invalid time or data:

   * The `start_time` does not align with the schedule’s cron pattern (e.g. schedule is Fridays 6pm and they send a Wednesday date). The server should detect that (maybe by checking `cronexpr.Next(before)` logic or just by generating expected times and seeing a mismatch) and reject with 400 Bad Request, since the request is logically incorrect.
   * The `schedule_id` provided doesn’t exist – should return 404 Not Found or 400.
   * (If implemented) The buddy’s phone is provided but is malformed – return 400.
   * `severity` in a report is out of range (e.g. 5) – the POST /report should return 400 with a message "severity must be 0,1,2".

7. **Authentication Enforcement:** Attempt to access a protected endpoint (like POST /bookings or GET /shifts/available if we make that protected) without a valid token. The response should be 401 Unauthorized. If using JWT, you can create a dummy token and ensure the middleware rejects an invalid one, etc. This ensures security is in place.

8. **Attendance Marking:** Simulate a booking that is completed (e.g. start\_time in the past). Have the user call the attendance endpoint to mark it. The request might be a PATCH with attended true. After the call, fetch the booking from the DB (or via GET if such endpoint) and verify `attended` is now true. Also, if you try marking attendance for a booking as a different user (not the owner), it should return 403 Forbidden and not change the data.

9. **Report Submission and Retrieval:** After a shift, POST a report for it. Then (if you have GET /reports or GET /bookings/{id} that includes report info) check that the report is stored correctly. If you design it so that creating a report when one already exists either updates or creates a new one, test that logic (maybe multiple reports per booking are allowed, then each call creates a new row). Also test that an invalid severity is rejected (as above).

10. **Outbox Dispatch:** This can be tricky to test end-to-end without an actual SMS service. But you can abstract the sender. For testing, implement a fake sender that simply appends the message to a slice or prints. Run the dispatcher logic (which picks an outbox entry and “sends” it). Mark the entry as sent. Then verify that the outbox entry’s status is updated and that the fake sender was called with the expected payload (e.g. the text containing the shift details or OTP). This test ensures your background job works and doesn’t resend things incorrectly. You can also test that if the sender fails (simulate an error), the status might remain “pending” or become “failed” and the message will be retried on next run.

**Testing Tools & Practices:**

* Use **table-driven tests** for functions like schedule parsing: supply an input cron and expected next times in a table, loop through and assert each. This is a common Go practice to cover multiple scenarios cleanly.
* For HTTP tests, you can use `httptest.NewRecorder` to simulate ResponseWriter and check the recorder’s Code and Body.
* Use `reflect.DeepEqual` or better, cmp libraries, to compare expected vs got structures in complex cases.
* Test negative paths (errors) as well as positive.
* Keep tests independent: reset the database or use a fresh in-memory DB for each test (or use transactions and roll back at end of test).
* Continuous integration: Ensure tests run on every build; catch things like broken SQL migrations (e.g., run a migration + connection in a TestMain).

By writing thorough tests, we gain confidence that:

* **“Booking a taken slot returns 409”** (as an example assertion) is verified in code,
* The core logic for scheduling and booking works for all edge cases (no double bookings, correct filtering by date, etc.),
* Refactoring or adding features later won’t unintentionally break existing functionality (tests will catch regressions).

Finally, beyond automated tests, consider some manual testing or using tools like Postman to simulate real client interactions once the server is running. This can help validate that everything works in an integrated manner.

---

**Conclusion:** By following this guide, final-year students can build a robust Go backend for the community watch shift scheduler. The solution uses industry-grade practices (Cron scheduling, transactional outbox for messages, REST API design, layered architecture) while remaining manageable in scope. The emphasis on code clarity, separation of concerns, and testing will help ensure the project is not only functional but also maintainable and ready for production use.
