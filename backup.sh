#!/bin/bash
set -e

echo "📦 Night Owls Control - Backup Script"
echo "====================================="

# Configuration
BACKUP_DIR="./backups"
CONTAINER_NAME="night-owls-app"
DB_PATH="/app/data/production.db"
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
RETENTION_DAYS=7

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Detect docker compose command
if command -v "docker compose" &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
else
    print_warning "Docker Compose not found. Using docker commands directly."
    DOCKER_COMPOSE=""
fi

# Create backup directory if it doesn't exist
mkdir -p $BACKUP_DIR

# Check if container is running
if ! docker ps | grep -q $CONTAINER_NAME; then
    print_warning "Container $CONTAINER_NAME is not running. Cannot create backup."
    exit 1
fi

print_status "Creating database backup..."

# Create backup inside container
docker exec $CONTAINER_NAME sqlite3 $DB_PATH ".backup /tmp/backup-$TIMESTAMP.db"

# Copy backup from container
docker cp $CONTAINER_NAME:/tmp/backup-$TIMESTAMP.db $BACKUP_DIR/

# Remove temporary backup from container
docker exec $CONTAINER_NAME rm /tmp/backup-$TIMESTAMP.db

# Compress backup
print_status "Compressing backup..."
gzip $BACKUP_DIR/backup-$TIMESTAMP.db

BACKUP_FILE="$BACKUP_DIR/backup-$TIMESTAMP.db.gz"
BACKUP_SIZE=$(du -h $BACKUP_FILE | cut -f1)

print_success "Backup created: $BACKUP_FILE ($BACKUP_SIZE)"

# Clean up old backups
print_status "Cleaning up old backups (keeping last $RETENTION_DAYS days)..."
find $BACKUP_DIR -name "backup-*.db.gz" -mtime +$RETENTION_DAYS -delete

REMAINING_BACKUPS=$(ls -1 $BACKUP_DIR/backup-*.db.gz 2>/dev/null | wc -l)
print_status "Remaining backups: $REMAINING_BACKUPS"

# Export logs
print_status "Exporting container logs..."
docker logs $CONTAINER_NAME --since 24h > $BACKUP_DIR/logs-$TIMESTAMP.log 2>&1
gzip $BACKUP_DIR/logs-$TIMESTAMP.log

# Clean up old log exports
find $BACKUP_DIR -name "logs-*.log.gz" -mtime +$RETENTION_DAYS -delete

print_success "Backup process completed!"

# Optional: Upload to cloud storage
# Uncomment and configure for your cloud provider
# 
# print_status "Uploading to cloud storage..."
# aws s3 cp $BACKUP_FILE s3://your-backup-bucket/night-owls/
# print_success "Uploaded to S3" 