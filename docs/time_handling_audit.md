# Time Handling Audit and Path to Sanity

## 1. Introduction

This document outlines the current state of time data handling in the Night Owls application, identifies areas of confusion or potential bugs, and proposes a path towards a more robust, consistent, and understandable approach. Accurate time handling is critical for a scheduling application, especially when user safety and operational reliability depend on it.

## 2. Current State Analysis

Based on an initial codebase review, here's how time-related data is managed across the backend (Go) and frontend (Svelte/TypeScript).

### 2.1. Backend (Go)

#### 2.1.1. Data Types & Storage:
*   **Schedule `start_date`, `end_date`:**
    *   API Input: Strings in "YYYY-MM-DD" format (`AdminCreateScheduleRequest`, `AdminUpdateScheduleRequest` in `internal/api/admin_schedule_handlers.go`).
    *   Parsing: `time.Parse("2006-01-02", dateStr)` is used, which creates a `time.Time` object at `00:00:00` on that date, in **UTC**.
    *   Database Storage: Stored using `sql.NullTime` in the `schedules` table. `sql.NullTime` will store the `time.Time` object (which is UTC). The underlying database type (e.g., SQLite `DATETIME` or `TEXT`) will typically store this as an ISO 8601 string with a 'Z' suffix (e.g., `2023-10-26T00:00:00Z`).
*   **Schedule `timezone`:**
    *   API Input: String (e.g., "America/New_York", "Europe/London").
    *   Database Storage: Stored using `sql.NullString`.
*   **Shift Slot `start_time`, `end_time` (e.g., in `AvailableShiftSlot`):**
    *   Generation: Calculated in `internal/service/schedule_service.go` using `cronexpr.Next()`. These times are generated *in the schedule's specific timezone* (e.g., if `loc = time.LoadLocation("America/New_York")`, the `time.Time` objects will have this location set).
    *   In-memory: `time.Time` objects with specific location information.
*   **API Query Parameters (e.g., `from`, `to` for available shifts):**
    *   Input: RFC3339 strings (e.g., `2023-10-26T10:00:00-04:00` or `2023-10-27T14:00:00Z`).
    *   Parsing: `time.Parse(time.RFC3339, fromStr)` correctly parses these into `time.Time` objects that represent the specific instant intended by the client.
*   **API Response (JSON):**
    *   Go's `encoding/json` marshals `time.Time` objects into RFC3339 strings. If a `time.Time` object has a specific location (e.g., from `time.LoadLocation`), the offset for that location at that time is included (e.g., `...-04:00`). If it's UTC, it will have a 'Z' (e.g., `...Z`).

#### 2.1.2. Key Logic & Transformations:
*   **Schedule Activation (`internal/service/schedule_service.go`):**
    *   The system uses `time.Now().UTC()` as a baseline for "current time".
    *   Query parameters (`queryFrom`, `queryTo`) are converted to UTC (`.UTC()`).
    *   A schedule's `StartDate.Time` (which is UTC from DB, e.g., `2023-10-26T00:00:00Z`) has its year, month, and day extracted. A new `time.Time` is then constructed at `00:00:00` (for start) or `23:59:59.999...` (for end) *in the schedule's specific timezone* (`loc`).
    *   **Interpretation:** A `start_date` of "2023-10-26" for a schedule in "America/New_York" means the schedule becomes active from `2023-10-26T00:00:00` *New York time*.
*   **Timezone Loading:**
    *   `time.LoadLocation(schedule.Timezone.String)` is used. Defaults to UTC if the timezone string is invalid or empty, with a warning log.
*   **Database Queries (`ListActiveSchedules` in `schedules.sql`):**
    *   Uses `date(?) >= start_date` and `date(?) <= end_date`.
    *   `start_date`/`end_date` in DB are UTC timestamps (e.g., `2023-10-26T00:00:00Z`).
    *   The `?` parameter (current date) is likely also passed as a UTC timestamp.
    *   SQLite's `date()` function, when applied to an ISO timestamp string, extracts the date part. This means the comparison effectively happens on UTC dates.

### 2.2. Frontend (Svelte/TypeScript)

#### 2.2.1. Data Types & Handling:
*   **Receiving Times from API:**
    *   Shift `start_time`, `end_time` arrive as ISO 8601 strings (RFC3339 format from Go, e.g., `2023-10-26T10:00:00-04:00` or `...Z`).
    *   Schedule `start_date`, `end_date` also arrive as ISO 8601 strings representing `YYYY-MM-DDT00:00:00Z`.
*   **JavaScript `Date` Object:**
    *   `new Date(isoString)` is used to parse these strings. JS `Date` objects internally represent a specific instant in time (milliseconds since UTC epoch). The timezone offset in the ISO string is used for correct parsing.
*   **Displaying Times:**
    *   `date.toLocaleString(undefined, options)`: Displays the date/time in the **user's local timezone**. This is generally correct for user-facing display of shift times.
    *   `date.toLocaleDateString()`: Used for displaying schedule start/end dates.
*   **Sending Dates to API (Schedule Forms):**
    *   Schedule `start_date`, `end_date` are sent as "YYYY-MM-DD" strings.
    *   Utilities like `parseIsoToYyyyMmDd` exist to convert from potentially fuller ISO strings to this format for form population.

#### 2.2.2. Key Components:
*   **Date Pickers (`DateRangePicker`, etc.):**
    *   Often work with "local date" concepts (e.g., `CalendarDate` from `@internationalized/date`). The interaction between a "local date" selected by the user and the "YYYY-MM-DD" string sent to the backend needs to be clear, especially considering the schedule's target timezone.

## 3. Identified Issues & Areas of Concern

### 3.1. Ambiguity of "Date-Only" Fields (Schedule `start_date`, `end_date`)
*   **The Core Problem:** A "date" like "2023-10-26" is inherently ambiguous without a timezone. Does it start at `00:00 UTC`, `00:00` in the user's local timezone, or `00:00` in the schedule's *target* timezone?
*   **Current Handling:**
    1.  Frontend sends "YYYY-MM-DD".
    2.  Backend parses this as `YYYY-MM-DDT00:00:00Z` (UTC).
    3.  Backend stores this UTC `time.Time`.
    4.  `schedule_service` later interprets this stored UTC date by taking its Y/M/D components and creating a new `time.Time` at `00:00:00` *in the schedule's specific timezone*.
*   **Confusion Points:**
    *   The transformation from a simple "YYYY-MM-DD" input to a UTC timestamp, then to a timezone-specific start/end of day, is complex and not immediately obvious.
    *   It relies on the `timezone` field of the schedule being correctly set *at the time of input/processing*.
    *   If a user intends "October 26th in New York" and the system goes through UTC, it works out, but it's indirect.

### 3.2. `ListActiveSchedules` SQL Query
*   The current SQL query compares dates at the UTC level (`date(param_utc) >= date(db_stored_utc_date)`).
*   **Potential Issue:** A schedule might be considered "active" or "inactive" based on UTC day boundaries, which might differ from its actual activity period in its specific timezone.
    *   Example: A schedule in "America/Los_Angeles" (UTC-7) intended to start "2023-10-26" (local) should become active at `2023-10-26T00:00:00` PDT, which is `2023-10-26T07:00:00Z`.
    *   If `now_utc` is `2023-10-26T01:00:00Z`, `date(now_utc)` is "2023-10-26". `date(db_stored_utc_date)` (which is `2023-10-26T00:00:00Z`) is also "2023-10-26". The query would match.
    *   However, at `2023-10-26T01:00:00Z`, it is still "2023-10-25" in Los Angeles. The schedule shouldn't be active yet from the perspective of its own timezone.
*   This discrepancy means that filtering active schedules at the database level using only UTC dates might not align with the business logic's interpretation of schedule activity based on its specific timezone.

### 3.3. Timezone String Validity
*   The system relies on `time.LoadLocation()` successfully parsing timezone strings stored in the database (e.g., "America/New_York").
*   If an invalid string is stored, it defaults to UTC. While there's a log, this could lead to subtle incorrect behavior for schedules if the intended timezone isn't applied. Input validation for timezone strings is important.

### 3.4. Frontend Date Input for Schedules
*   When a user picks a `start_date` for a schedule using a date picker, the date they pick ("October 26th") is generally understood by them in the context of *that schedule's location*.
*   The current process of sending "YYYY-MM-DD" and then having the backend re-interpret it via UTC and then into the schedule's timezone is functional but indirect.

### 3.5. Consistency in API Time Formatting
*   While Go's default JSON marshalling of `time.Time` to RFC3339 is good, ensuring all time-related API parameters (`from`, `to` query params) also strictly expect RFC3339 (as `ListAvailableShiftsHandler` does) is crucial. The admin schedule creation/update handlers currently expect "YYYY-MM-DD" which is a different style.

## 4. Proposed Path to Sanity & Standardization

The goal is to make time handling explicit, reduce ambiguity, and ensure consistency across the stack.

### 4.1. Guiding Principles
1.  **UTC for Storage and API (mostly):** Store all absolute timestamps in the database as UTC equivalent (e.g., SQLite `DATETIME` storing `YYYY-MM-DDTHH:MM:SSZ`). API endpoints dealing with specific instants should primarily accept and return RFC3339 strings.
2.  **Explicit Timezones:** Always be explicit about timezones. Schedules *must* have a valid IANA timezone identifier.
3.  **Local Time for User Interface:** Display times to users in their local timezone. For inputs that define "a day" (like schedule start/end dates), be clear about which timezone defines that day.
4.  **Server-Side Authority:** The backend should be the authority for all time calculations, timezone conversions, and slot generation.

### 4.2. Recommendations

#### 4.2.1. Schedule Start/End Dates ("Floating Dates" Done Right)
*   **Concept:** The `start_date` and `end_date` of a schedule represent a "floating" date (like a birthday) that only becomes a concrete range of instants when combined with the schedule's timezone. "October 26th" is the start date, regardless of whether it's in NY or London.
*   **Backend API (`AdminCreateScheduleRequest`, `AdminUpdateScheduleRequest`):**
    *   Continue accepting `start_date` and `end_date` as "YYYY-MM-DD" strings.
    *   **Crucially, also require the `timezone` string in the same request.**
*   **Backend Storage (`schedules` table):**
    *   Store `start_date_str` and `end_date_str` as plain "YYYY-MM-DD" strings (e.g., `TEXT` type in DB).
    *   Store `timezone` as a validated IANA string (e.g., `TEXT` type).
    *   **Remove** the `sql.NullTime` columns for `start_date` and `end_date` from `db.Schedule` if they are solely to represent these YYYY-MM-DD values. If they are used for other timestamp purposes, they need renaming and re-evaluation. *Initial review suggests they are for YYYY-MM-DD.*
*   **Service Logic (`schedule_service.go`):**
    *   When needing the actual start/end *instants* of a schedule:
        1.  Parse the "YYYY-MM-DD" string from `start_date_str`.
        2.  Load the `time.Location` from the schedule's `timezone` string.
        3.  Construct the `time.Time` for the start instant: `time.Date(year, month, day, 0, 0, 0, 0, loc)`.
        4.  Construct the `time.Time` for the end instant: `time.Date(year, month, day, 23, 59, 59, 999999999, loc)`.
    *   This makes the interpretation explicit: "the schedule runs from the beginning of `start_date_str` to the end of `end_date_str`, *within its own timezone*."

#### 4.2.2. Shift Slot Times (`start_time`, `end_time`)
*   **No change to generation:** Continue generating these in the schedule's specific timezone using `cronexpr` and `time.Time` with location.
*   **API Output:** These will be marshalled to RFC3339 by Go, including the correct offset (e.g., `...-04:00`). This is correct.
*   **Frontend:** Continue parsing with `new Date()` and displaying with `toLocaleString()`. This is correct.

#### 4.2.3. API Query Parameters for Time Ranges (e.g., `from`, `to`)
*   **Standardize on RFC3339:** All API endpoints accepting a time range (`from`, `to` parameters) MUST expect full RFC3339 strings. This ensures the client can specify an exact instant with its timezone.
*   The backend should parse these using `time.Parse(time.RFC3339, ...)`. The resulting `time.Time` object will accurately represent that instant.
*   Backend logic can then convert these to UTC (`.UTC()`) if needed for internal comparisons or database queries against UTC timestamps.

#### 4.2.4. Database Queries for Active Schedules / Slots
*   **Option A (Preferred for accuracy): Filter in Go.**
    *   Fetch a broader set of schedules/slots from the database (e.g., based on a wider UTC range or other criteria).
    *   Perform precise active/within-range checks in Go service code, using the full timezone logic described in 4.2.1 and 4.2.2. This ensures business logic for timezones is correctly applied.
*   **Option B (If DB filtering is essential and complex):**
    *   This is much harder. It would require storing pre-calculated UTC start/end *instants* for each schedule's active period, or passing timezone-aware start/end instants into the query that correctly define the query window in UTC.
    *   For instance, to find schedules active between `queryFrom` (RFC3339) and `queryTo` (RFC3339):
        1. Convert `queryFrom` and `queryTo` to UTC.
        2. For each schedule, calculate its `effectiveStartUTC` and `effectiveEndUTC` based on its "YYYY-MM-DD" start/end and its specific timezone.
        3. The query would then be: `(effectiveStartUTC < queryToUTC AND effectiveEndUTC > queryFromUTC)`.
        This is complex to do purely in SQL without potentially Stored Procedures or more advanced DB features. It's simpler in Go.

#### 4.2.5. Frontend Schedule Date Inputs
*   The `DateRangePicker` should be configured to primarily deal with "YYYY-MM-DD" strings.
*   The user experience should clearly indicate that these dates apply to the schedule's timezone (which should also be an input field on the same form).
*   No complex client-side timezone conversions are needed for *these specific inputs* if the backend adopts 4.2.1.

#### 4.2.6. Timezone Database & Validation
*   Ensure a robust way to validate IANA timezone strings on input (e.g., by attempting `time.LoadLocation` and rejecting if it errors).
*   Consider if the list of selectable timezones in the UI should be curated or if any valid IANA string is allowed.

### 4.3. Documentation and Comments
*   Update inline code comments and create clear documentation (like this document) explaining the chosen time handling strategy.
*   Specifically, document the meaning of "YYYY-MM-DD" fields vs. RFC3339 timestamp fields.

## 5. Next Steps
1.  **Review and Agreement:** Discuss these findings and recommendations with the team.
2.  **Prioritize Changes:** Identify which changes are most critical and can be implemented first.
    *   **High Priority:** Clarifying Schedule Start/End Dates (4.2.1), Standardizing API Query Params (4.2.3), and moving schedule active filtering to Go (4.2.4 Option A).
3.  **Implementation Plan:**
    *   Backend:
        *   Modify `AdminCreateScheduleRequest`, `AdminUpdateScheduleRequest` to ensure `timezone` is always present with "YYYY-MM-DD" dates.
        *   Change DB schema for `schedules` to store `start_date_str`, `end_date_str` as TEXT. Update `sqlc` queries and Go structs.
        *   Refactor `schedule_service.go` to use these string dates and the schedule's timezone to calculate start/end instants.
        *   Refactor `ListActiveSchedules` (and similar) logic to perform filtering in Go.
    *   Frontend:
        *   Ensure schedule forms submit "YYYY-MM-DD" and the associated timezone.
        *   Verify date pickers are straightforward for "YYYY-MM-DD" input.
4.  **Testing:** Thoroughly test all changes, focusing on edge cases involving different timezones, DST transitions, and date boundaries.

This structured approach should significantly improve the clarity, correctness, and maintainability of time handling in the application. 