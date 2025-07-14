# Night Owls Scripts

This directory contains utility scripts for managing the Night Owls application.

## Database Backup Script

`backup-db.sh` - Creates backups of the production SQLite database from the Docker volume.

### Quick Start

```bash
# Basic backup
./scripts/backup-db.sh

# Backup with verification
./scripts/backup-db.sh --verify

# List recent backups
./scripts/backup-db.sh --list
```

### Features

- ✅ Backs up SQLite database from Docker volume
- ✅ Timestamps all backups automatically
- ✅ Verifies backup integrity (if sqlite3 is available)
- ✅ No downtime required
- ✅ Configurable backup location
- ✅ Lists recent backups
- ✅ Provides restore commands

### Requirements

- Docker installed and running
- Night Owls container and volume exist
- Write access to backup directory
- Optional: `sqlite3` command for integrity verification

## Migration Commands

With the updated Docker build, the migration tool is now available inside the container:

### Run Migration (Recommended approach)

```bash
# 1. Create backup first
./scripts/backup-db.sh --verify

# 2. Run dry-run to see what will happen
docker exec -it night-owls-go ./migrate-points --dry-run

# 3. Run actual migration
docker exec -it night-owls-go ./migrate-points

# 4. Force migration without confirmation (use with caution)
docker exec -it night-owls-go ./migrate-points --force
```

### Migration Features

- ✅ Awards historical points for completed shifts
- ✅ Dry-run mode to preview changes
- ✅ Interactive confirmation by default
- ✅ Force mode for automated deployments
- ✅ Comprehensive logging
- ✅ Database transaction safety

## Deployment Workflow

For production deployments with migration:

```bash
# 1. Backup database
./scripts/backup-db.sh --verify

# 2. Deploy new container with migration tool
docker-compose pull
docker-compose up -d

# 3. Run migration
docker exec -it night-owls-go ./migrate-points --dry-run
docker exec -it night-owls-go ./migrate-points

# 4. Verify application is working
curl -f http://localhost:5888/health
```

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