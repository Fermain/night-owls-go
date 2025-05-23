# End-to-End (E2E) Tests

This directory contains Playwright end-to-end tests for the Community Watch application.

## Prerequisites

Before running e2e tests, make sure you have:

1. **Development servers running**: Both the Go backend and SvelteKit frontend must be running
   ```bash
   # From the project root
   ./dev.sh
   ```

2. **Dependencies installed**: Make sure all npm dependencies are installed
   ```bash
   # From the app directory  
   pnpm install
   ```

## Running Tests

### Run All E2E Tests
```bash
# From the app directory
pnpm test:e2e
```

### Run Only Authentication Tests
```bash
pnpm test:e2e:auth
```

### Run Tests with UI (Visual Mode)
```bash
pnpm test:e2e:ui
```

### Debug Tests
```bash
pnpm test:e2e:debug
```

## Test Structure

### Authentication Flow Tests (`auth-flow.test.ts`)

This comprehensive test suite covers the complete authentication journey:

#### Test Cases

1. **Full Registration and Login Flow**
   - Visits homepage → clicks "Join Us" → fills registration form → gets OTP → verifies code → logs in
   - Tests the entire new user journey from discovery to authenticated state

2. **Existing User Login Flow**
   - Tests login for users who have already registered
   - Verifies that returning users can get new OTPs and log in successfully

3. **Invalid OTP Handling**
   - Tests error handling when users enter incorrect verification codes
   - Verifies that the UI properly clears invalid input and shows error messages

4. **Authenticated User Redirects**
   - Tests that logged-in users are automatically redirected away from auth pages
   - Verifies proper session handling and route protection

5. **Back Navigation**
   - Tests the "go back" functionality in the auth flow
   - Verifies that form state is preserved when navigating backwards

6. **Outbox Processing Delays**
   - Tests handling of slow OTP generation and processing
   - Verifies robustness against database and background job delays

#### Technical Features

- **Dynamic Test Data**: Each test run uses unique phone numbers and names to avoid conflicts
- **Database Integration**: Tests interact directly with the SQLite database to retrieve OTPs
- **Robust OTP Handling**: Includes retry logic and timeout handling for OTP retrieval
- **Automatic Cleanup**: Test data is cleaned up before and after each test

### Test Utilities (`test-utils.ts`)

#### DatabaseHelper Class
- `getLatestOTP(phone)`: Retrieves the most recent OTP for a phone number
- `getAllOTPs(phone)`: Gets all OTPs for debugging purposes
- `cleanupTestUser(phone)`: Removes test data from database
- `userExists(phone)`: Checks if a user exists
- `getUserByPhone(phone)`: Retrieves user details
- `waitForOutboxProcessing(phone, timeout)`: Waits for OTP to be generated and processed

#### AuthTestHelper Class
- `generateTestPhone()`: Creates unique test phone numbers
- `generateTestName()`: Creates unique test names

#### TEST_CONFIG
- Constants for timeouts, intervals, and default test values

## Database Requirements

The tests expect:
- SQLite database at `./night-owls.test.db` (relative to project root)
- `outbox` table for OTP retrieval
- `users` table for user data
- Background cron job processing the outbox

## Troubleshooting

### Common Issues

1. **"No OTP found"**: 
   - Check that the development servers are running
   - Verify the outbox cron job is processing messages
   - Check database connectivity

2. **Tests timing out**:
   - Increase timeout values in `TEST_CONFIG`
   - Check that both frontend and backend are responding
   - Verify database operations are working

3. **Element not found errors**:
   - Check that the UI components have the expected text/attributes
   - Verify the application is loading correctly
   - Update selectors if the UI has changed

### Debug Tips

1. **Use UI mode** for visual debugging: `pnpm test:e2e:ui`
2. **Check console logs** in the test output for OTP values
3. **Verify database state** using SQLite commands:
   ```bash
   sqlite3 night-owls.test.db "SELECT * FROM outbox WHERE message_type = 'OTP_VERIFICATION';"
   ```

## Configuration

### Playwright Config
- Tests run against `http://localhost:5173` (frontend)
- Backend expected at `http://localhost:8080`
- Tracing enabled for debugging
- Uses existing dev servers (doesn't start its own)

### TypeScript Config
- Node.js types included for database operations
- Playwright types included for test framework

## Contributing

When adding new e2e tests:

1. **Use the test utilities** for database operations and test data generation
2. **Clean up test data** in beforeEach/afterEach hooks
3. **Use unique test data** to avoid conflicts between test runs
4. **Add appropriate timeouts** for async operations
5. **Document new test cases** in this README

## Example Test Run

```bash
# Start the development environment
./dev.sh

# In another terminal, run auth tests
cd app
pnpm test:e2e:auth
```

Expected output:
```
Running 6 tests using 1 worker

✓ Authentication Flow › should complete full registration and login flow (15s)
✓ Authentication Flow › should handle existing user login flow (12s)
✓ Authentication Flow › should handle invalid OTP gracefully (8s)
✓ Authentication Flow › should redirect authenticated users away from auth pages (18s)
✓ Authentication Flow › should handle back navigation in auth flow (6s)
✓ Authentication Flow › should handle outbox processing delays (25s)

  6 passed (1m 84s)
``` 