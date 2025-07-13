#!/bin/bash

# Night Owls Database Backup Script
# Backs up the production SQLite database from the Docker volume

set -euo pipefail

# Configuration
CONTAINER_NAME="${CONTAINER_NAME:-night-owls-go}"
VOLUME_NAME="${VOLUME_NAME:-night_owls_data}"
DB_PATH_IN_VOLUME="${DB_PATH_IN_VOLUME:-data/production.db}"
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="night_owls_backup_${TIMESTAMP}.db"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Night Owls Database Backup Script

This script creates a backup of the production SQLite database from the Docker volume.
The backup is copied to a local directory with a timestamp.

OPTIONS:
    -h, --help              Show this help message
    -c, --container NAME    Docker container name (default: night-owls-go)
    -v, --volume NAME       Docker volume name (default: night_owls_data)
    -d, --backup-dir DIR    Local backup directory (default: ./backups)
    -f, --filename NAME     Custom backup filename (default: night_owls_backup_TIMESTAMP.db)
    --verify                Verify the backup after creation
    --list                  List recent backups

EXAMPLES:
    # Basic backup
    $0

    # Backup with custom directory
    $0 --backup-dir /path/to/backups

    # Backup with verification
    $0 --verify

    # List recent backups
    $0 --list

ENVIRONMENT VARIABLES:
    CONTAINER_NAME         Override default container name
    VOLUME_NAME           Override default volume name  
    BACKUP_DIR            Override default backup directory

The script will:
1. Check if Docker and the container exist
2. Create the backup directory if needed
3. Copy the database from the Docker volume
4. Optionally verify the backup integrity
5. Show backup information and cleanup instructions

EOF
}

# Function to check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if container exists
    if ! docker ps -a --format "table {{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
        log_error "Container '${CONTAINER_NAME}' not found"
        log_info "Available containers:"
        docker ps -a --format "table {{.Names}}\t{{.Status}}"
        exit 1
    fi
    
    # Check if volume exists
    if ! docker volume ls --format "{{.Name}}" | grep -q "^${VOLUME_NAME}$"; then
        log_error "Volume '${VOLUME_NAME}' not found"
        log_info "Available volumes:"
        docker volume ls
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Function to verify backup
verify_backup() {
    local backup_path="$1"
    
    log_info "Verifying backup integrity..."
    
    # Check if file exists and is not empty
    if [[ ! -f "$backup_path" ]] || [[ ! -s "$backup_path" ]]; then
        log_error "Backup file is missing or empty"
        return 1
    fi
    
    # Use SQLite to check database integrity
    if command -v sqlite3 &> /dev/null; then
        if sqlite3 "$backup_path" "PRAGMA integrity_check;" | grep -q "ok"; then
            log_success "Backup integrity verified"
            
            # Show some basic stats
            local table_count=$(sqlite3 "$backup_path" "SELECT COUNT(*) FROM sqlite_master WHERE type='table';")
            local db_size=$(du -h "$backup_path" | cut -f1)
            log_info "Database contains $table_count tables, backup size: $db_size"
        else
            log_error "Backup integrity check failed"
            return 1
        fi
    else
        log_warn "sqlite3 not available, skipping integrity check"
        log_info "Install sqlite3 to enable backup verification"
    fi
}

# Function to list recent backups
list_backups() {
    log_info "Recent backups in $BACKUP_DIR:"
    
    if [[ -d "$BACKUP_DIR" ]]; then
        find "$BACKUP_DIR" -name "night_owls_backup_*.db" -type f -exec ls -lh {} \; | \
        sort -k 6,8 -r | head -10
    else
        log_warn "Backup directory $BACKUP_DIR does not exist"
    fi
}

# Function to create backup
create_backup() {
    log_info "Starting database backup..."
    log_info "Container: $CONTAINER_NAME"
    log_info "Volume: $VOLUME_NAME" 
    log_info "Database path: $DB_PATH_IN_VOLUME"
    log_info "Backup directory: $BACKUP_DIR"
    log_info "Backup file: $BACKUP_FILE"
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    
    # Full path to backup file
    local backup_path="$BACKUP_DIR/$BACKUP_FILE"
    
    # Copy database from Docker volume
    log_info "Copying database from Docker volume..."
    
    # Use a temporary container to access the volume and copy the database
    docker run --rm \
        -v "${VOLUME_NAME}:/source:ro" \
        -v "$(realpath "$BACKUP_DIR"):/backup" \
        alpine:latest \
        cp "/source/${DB_PATH_IN_VOLUME}" "/backup/${BACKUP_FILE}"
    
    # Check if backup was created successfully
    if [[ -f "$backup_path" ]]; then
        log_success "Backup created successfully: $backup_path"
        
        # Show backup size
        local backup_size=$(du -h "$backup_path" | cut -f1)
        log_info "Backup size: $backup_size"
        
        return 0
    else
        log_error "Failed to create backup"
        return 1
    fi
}

# Main function
main() {
    local verify_backup_flag=false
    local list_backups_flag=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -c|--container)
                CONTAINER_NAME="$2"
                shift 2
                ;;
            -v|--volume)
                VOLUME_NAME="$2"
                shift 2
                ;;
            -d|--backup-dir)
                BACKUP_DIR="$2"
                shift 2
                ;;
            -f|--filename)
                BACKUP_FILE="$2"
                shift 2
                ;;
            --verify)
                verify_backup_flag=true
                shift
                ;;
            --list)
                list_backups_flag=true
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    # Handle list backups flag
    if [[ "$list_backups_flag" == true ]]; then
        list_backups
        exit 0
    fi
    
    # Main backup process
    log_info "🗄️  Night Owls Database Backup Script"
    log_info "======================================"
    
    check_prerequisites
    
    if create_backup; then
        local backup_path="$BACKUP_DIR/$BACKUP_FILE"
        
        # Verify backup if requested
        if [[ "$verify_backup_flag" == true ]]; then
            verify_backup "$backup_path"
        fi
        
        log_success "✅ Backup completed successfully!"
        echo
        log_info "📁 Backup location: $backup_path"
        log_info "📊 To restore this backup, you can use:"
        echo "   docker cp \"$backup_path\" ${CONTAINER_NAME}:/app/data/production.db"
        echo
        log_info "🧹 To clean up old backups, run:"
        echo "   find \"$BACKUP_DIR\" -name 'night_owls_backup_*.db' -mtime +30 -delete"
        
    else
        log_error "❌ Backup failed!"
        exit 1
    fi
}

# Run main function with all arguments
main "$@"