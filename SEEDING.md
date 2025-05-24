# Database Seeding Infrastructure

This document describes the database seeding infrastructure for the Night Owls Go application.

## Overview

The seeding system provides comprehensive development data including users, schedules, recurring assignments, and sample bookings. It uses the actual migration system to ensure perfect consistency with production schema.

## Quick Start

```bash
# Preview what would be seeded
make seed-preview

# Seed the configured database
make seed

# Reset and seed from scratch
make seed-reset

# Set up complete development environment
make dev-setup

# Basic seeding
make seed-reset          # Reset and seed with 10 users
make seed-preview        # Preview seed data (dry run)

# Enhanced seeding options
make seed-minimal        # 3 users only
make seed-large          # 50 users  
make seed-future         # Include 30 days of future bookings
make seed-export         # Export seed data to JSON
make seed-demo           # Full demo: 100 users, future bookings, export

# Custom seeding
./cmd/seed/seed --users 25 --future-bookings --export data.json --verbose
```

## Seeding Command

The seed command is located at `cmd/seed/main.go` and can be used directly:

```bash
# Build the seed command
go build -o cmd/seed/seed cmd/seed/main.go

# Basic usage
./cmd/seed/seed

# Command line options
./cmd/seed/seed --help
./cmd/seed/seed --db path/to/database.db     # Use custom database
./cmd/seed/seed --reset                      # Reset before seeding
./cmd/seed/seed --dry-run                    # Preview without changes
```

## Seeded Data

### Users (10 total)

#### Administrators (2)
- **Alice Admin** - `+27821234567` - Primary admin user
- **Bob Manager** - `+27821234568` - Secondary admin user

#### Owl Volunteers (6) 
- **Charlie Volunteer** - `+27821234569` - Weekend morning shifts
- **Diana Scout** - `+27821234570` - Weekend morning partner
- **Eve Patrol** - `+27821234571` - Evening patrol regular
- **Frank Guard** - `+27821234572` - Weekday lunch security
- **Grace Watch** - `+27821234573` - Summer patrol volunteer
- **Henry Security** - `+27821234574` - Security backup

#### Guests (2)
- **Iris Guest** - `+27821234575` - Guest user
- **Jack Visitor** - `+27821234576` - Visitor user

### Schedules (3 additional + 2 from migrations)

#### Development Schedules
1. **Daily Evening Patrol** - `0 18 * * *` (Every day at 6 PM, 2 hours)
2. **Weekend Morning Watch** - `0 6,10 * * 6,0` (Sat/Sun at 6 AM & 10 AM, 4 hours)
3. **Weekday Lunch Security** - `0 12 * * 1-5` (Mon-Fri at noon, 1 hour)

#### From Migrations
4. **Summer Patrol (Nov-Apr)** - Already seeded via migration
5. **Winter Patrol (May-Oct)** - Already seeded via migration

### Recurring Assignments (7 total)

| User | Schedule | Day | Time Slot | Buddy/Description |
|------|----------|-----|-----------|-------------------|
| Charlie | Weekend Morning Watch | Saturday | 06:00-10:00 | Diana Scout |
| Diana | Weekend Morning Watch | Sunday | 10:00-14:00 | Charlie Volunteer |
| Eve | Daily Evening Patrol | Monday | 18:00-20:00 | Monday evening patrol |
| Eve | Daily Evening Patrol | Wednesday | 18:00-20:00 | Wednesday evening patrol |
| Frank | Weekday Lunch Security | Tuesday | 12:00-13:00 | Tuesday lunch security |
| Frank | Weekday Lunch Security | Thursday | 12:00-13:00 | Thursday lunch security |
| Grace | Summer Patrol (Nov-Apr) | Saturday | 00:00-02:00 | Henry Security |

### Sample Bookings (4 historical)

| User | Schedule | Date/Time | Buddy | Attended |
|------|----------|-----------|-------|----------|
| Charlie | Daily Evening Patrol | 2024-11-25 18:00 | Diana Scout | ✅ |
| Diana | Weekend Morning Watch | 2024-11-24 06:00 | Charlie Volunteer | ✅ |
| Eve | Daily Evening Patrol | 2024-11-26 18:00 | - | ❌ |
| Frank | Weekday Lunch Security | 2024-11-26 12:00 | - | ✅ |

## Development Workflow

### Initial Setup
```bash
# Complete development setup
make dev-setup

# Start the server
make run
```

### Daily Development
```bash
# Reset to clean state
make seed-reset

# Preview changes before seeding
make seed-preview

# Add new data without reset
make seed
```

### Testing Different Scenarios
```bash
# Create test database for experiments
make seed-test

# Custom database seeding
make seed-custom DB=./my-test.db
```

## Technical Details

### Architecture
- Uses real migration system for schema consistency
- Runs migrations before seeding data
- Gracefully handles existing data conflicts
- Provides detailed logging for debugging

### Schema Validation
The seeding system automatically:
1. Runs all database migrations
2. Validates foreign key relationships
3. Handles unique constraint conflicts
4. Reports seeding progress and errors

### Data Relationships
- All users are linked to their respective roles
- Recurring assignments reference valid users and schedules
- Sample bookings use actual schedule timing
- Buddy relationships are properly established

## Extending the Seed Data

To add new seed data, modify `cmd/seed/main.go`:

### Adding Users
```go
Users: []UserSeed{
    {Name: "New User", Phone: "+27821234577", Role: "owl"},
    // ... existing users
},
```

### Adding Schedules
```go
Schedules: []ScheduleSeed{
    {
        Name:            "My New Schedule",
        CronExpr:        "0 14 * * 2,4", // Tue/Thu at 2 PM
        StartDate:       "2024-01-01",
        EndDate:         "2024-12-31", 
        DurationMinutes: 90,
        Timezone:        "Africa/Johannesburg",
    },
    // ... existing schedules
},
```

### Adding Recurring Assignments
```go
RecurringAssignments: []RecurringAssignmentSeed{
    {
        UserPhone:    "+27821234577", // Must match existing user
        ScheduleName: "My New Schedule", // Must match existing schedule
        DayOfWeek:    2, // Tuesday (0=Sunday, 6=Saturday)
        TimeSlot:     "14:00-15:30",
        BuddyName:    "Partner Name",
        Description:  "Tuesday afternoon shift",
    },
    // ... existing assignments
},
```

### Adding Bookings
```go
Bookings: []BookingSeed{
    {
        UserPhone:    "+27821234577",
        ScheduleName: "My New Schedule",
        ShiftStart:   "2024-12-03T14:00:00Z", // Must be valid RFC3339
        BuddyName:    "Partner Name",
        Attended:     true,
    },
    // ... existing bookings
},
```

## Troubleshooting

### Common Issues

#### Migration Failures
```bash
# Check migration status
go run cmd/server/main.go # Will show migration errors

# Reset database completely
rm ./night-owls.test.db
make seed-reset
```

#### Constraint Violations
- Check that all phone numbers are unique
- Ensure schedule names match exactly
- Verify date formats are RFC3339 compliant
- Confirm user/schedule references exist

#### Permission Issues
```bash
# Ensure database file is writable
chmod 644 ./night-owls.test.db

# Check directory permissions
ls -la ./
```

### Debugging
```bash
# Use dry-run to debug data issues
make seed-preview

# Enable verbose logging (if available)
NIGHT_OWLS_LOG_LEVEL=debug make seed

# Check seed command directly
./cmd/seed/seed --dry-run --db ./debug.db
```

## Integration with Testing

The seeding infrastructure can be used in tests:

### Test Helper Pattern (Recommended)
```go
// Create helper for consistent test setup
func setupTestDBWithSeed(t *testing.T) *sql.DB {
    // Use migrations + seed data instead of manual schema
    // This ensures tests use identical schema to production
}
```

### Benefits Over Manual Schema
- ✅ Always matches production schema
- ✅ Includes all constraints and indexes
- ✅ Tests migration compatibility
- ✅ Reduces maintenance burden
- ❌ Slightly slower than in-memory manual schema

## Production Considerations

⚠️ **Never run seeding against production databases!**

The seeding system is designed for development only:
- Uses test phone numbers (+2782123456x)
- Creates predictable passwords/tokens
- Includes debug-friendly data patterns
- May conflict with real user data

For production data management, use proper database administration tools and scripts.

## Advanced Features

### 1. Custom User Count
Generate any number of users with intelligent role distribution:

```bash
# Generate exactly 5 users (core team)
./cmd/seed/seed --users 5 --reset

# Generate 100 users for load testing  
./cmd/seed/seed --users 100 --reset
```

**User Generation Logic:**
- **First 10**: Core team with specific names and roles
- **Additional users**: Auto-generated with pattern-based names
- **Role distribution**: Every 4th additional user is a guest, rest are owls
- **Phone numbers**: Sequential starting from +27821234577

### 2. Future Bookings
Generate realistic future bookings for testing and development:

```bash
./cmd/seed/seed --future-bookings --reset
```

**Future Booking Logic:**
- **Next 30 days**: Generates bookings for upcoming month
- **Weekday evenings**: Eve Patrol on Mon-Fri at 6 PM
- **Weekend mornings**: Charlie (Saturday), Diana (Sunday) at 6 AM
- **Realistic data**: All future bookings default to `attended: false`

### 3. Data Export
Export seeded data to structured JSON for analysis or documentation:

```bash
./cmd/seed/seed --export ./my-seed-data.json --reset
```

**Export Structure:**
```json
{
  "exported_at": "2025-05-24T13:47:19.03008Z",
  "version": "1.0", 
  "database": "Night Owls Go",
  "data": {
    "Users": [...],
    "Schedules": [...],
    "RecurringAssignments": [...],
    "Bookings": [...]
  }
}
```

### 4. Verbose Logging
Enhanced debugging and development visibility:

```bash
./cmd/seed/seed --verbose --users 50
```

**Features:**
- Detailed progress logging
- Debug-level information
- Human-readable timestamps
- Enhanced context information

### 5. Smart Data Filtering
Automatically filters data based on available users:

- **Recurring assignments**: Only created for existing users
- **Bookings**: Only generated for users that exist  
- **Relationships**: Maintains data integrity across all entities
- **Scaling**: Works seamlessly from 1 to 1000+ users

## Development Workflows

### Testing Different Scales
```bash
# Minimal for unit testing
make seed-minimal

# Standard for development
make seed-reset  

# Large scale for performance testing
make seed-large
```

### Demo Environment Setup
```bash
# Complete demo environment
make seed-demo

# Results in:
# - 100 users (2 admins, 73 owls, 25 guests)
# - 3 additional schedules + 2 from migrations
# - 7 recurring assignments (filtered by available users)
# - 4 historical + 30 future bookings
# - Complete data export in demo-data.json
```

### Data Analysis Workflow
```bash
# 1. Seed with export
make seed-export

# 2. Analyze the JSON
cat seed-export.json | jq '.data.Users | length'  # Count users
cat seed-export.json | jq '.data.Users | group_by(.Role)'  # Group by role

# 3. Verify database state
sqlite3 night-owls.test.db "SELECT role, COUNT(*) FROM users GROUP BY role;"
```

## Command Line Options

| Flag | Description | Example |
|------|-------------|---------|
| `--users N` | Generate N users | `--users 25` |
| `--future-bookings` | Generate 30 days of future bookings | `--future-bookings` |
| `--export FILE` | Export seed data to JSON file | `--export data.json` |
| `--verbose` | Enable detailed logging | `--verbose` |
| `--reset` | Reset database before seeding | `--reset` |
| `--dry-run` | Preview without making changes | `--dry-run` |
| `--db PATH` | Use specific database file | `--db test.db` |

## User Generation Patterns

### Core Users (1-10)
- **2 Admins**: Alice Admin, Bob Manager
- **6 Owls**: Charlie, Diana, Eve, Frank, Grace, Henry  
- **2 Guests**: Iris, Jack

### Additional Users (11+)
- **Pattern**: Every 4th user is a guest, rest are owls
- **Names**: Leo, Zoe, Max, Ivy, Sam, Ruby, Alex, Nova, Finn, Luna...
- **Phones**: Sequential +27821234577, +27821234578...
- **Roles**: Smart distribution for realistic testing

## Integration with Development

### Docker Development
```bash
# Seed container database
docker exec night-owls-container ./cmd/seed/seed --users 50 --reset

# Export for backup
docker exec night-owls-container ./cmd/seed/seed --export /backup/seed.json
```

### CI/CD Pipeline
```bash
# Test with minimal data  
make seed-minimal && make test

# Performance test with large dataset
make seed-large && make test-performance

# Integration test with future bookings
make seed-future && make test-integration
```

### Production Setup (Staging)
```bash
# Staging environment with realistic scale
./cmd/seed/seed --users 500 --future-bookings --export staging-backup.json --reset
``` 