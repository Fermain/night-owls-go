#!/bin/bash
# Night Owls Control - Deployment Script
# Usage: ./deploy.sh [production|staging]

set -e

# Configuration
ENVIRONMENT=${1:-production}
REMOTE_USER="deploy"
REMOTE_HOST="your-server-ip"  # TODO: Replace with your server IP
REMOTE_DIR="/home/deploy/night-owls-go"
BRANCH="main"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting deployment to $ENVIRONMENT...${NC}"

# Check if remote host is configured
if [ "$REMOTE_HOST" = "your-server-ip" ]; then
    echo -e "${RED}Error: Please configure REMOTE_HOST in deploy.sh${NC}"
    exit 1
fi

# Ensure we're on the correct branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "$BRANCH" ]; then
    echo -e "${YELLOW}Warning: You're on branch '$CURRENT_BRANCH', not '$BRANCH'${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check for uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo -e "${YELLOW}Warning: You have uncommitted changes${NC}"
    git status --short
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Run tests
echo -e "${GREEN}📋 Running tests...${NC}"
go test ./... || {
    echo -e "${RED}Tests failed! Aborting deployment.${NC}"
    exit 1
}

# Build frontend
echo -e "${GREEN}📦 Building frontend...${NC}"
cd app
npm ci
npm run check || {
    echo -e "${RED}Frontend type check failed!${NC}"
    exit 1
}
npm run build || {
    echo -e "${RED}Frontend build failed!${NC}"
    exit 1
}
cd ..

# Build backend for Linux
echo -e "${GREEN}🔨 Building backend for Linux...${NC}"
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o night-owls-server ./cmd/server || {
    echo -e "${RED}Backend build failed!${NC}"
    exit 1
}

# Create deployment archive
echo -e "${GREEN}📤 Creating deployment archive...${NC}"
tar -czf deploy.tar.gz \
    night-owls-server \
    app/build \
    internal/db/migrations \
    env.example \
    DEPLOYMENT.md \
    PRODUCTION_CHECKLIST.md

# Upload to server
echo -e "${GREEN}📤 Uploading to server...${NC}"
scp deploy.tar.gz $REMOTE_USER@$REMOTE_HOST:/tmp/

# Deploy on server
echo -e "${GREEN}🔄 Deploying on server...${NC}"
ssh $REMOTE_USER@$REMOTE_HOST << 'ENDSSH'
set -e

# Backup current deployment
if [ -d "/home/deploy/night-owls-go" ]; then
    echo "Backing up current deployment..."
    sudo systemctl stop night-owls-backend night-owls-frontend || true
    tar -czf /home/deploy/backups/night-owls-backup-$(date +%Y%m%d-%H%M%S).tar.gz \
        -C /home/deploy night-owls-go
fi

# Backup database
if [ -f "/home/deploy/night-owls-data/production.db" ]; then
    echo "Backing up database..."
    cp /home/deploy/night-owls-data/production.db \
       /home/deploy/night-owls-data/backups/production-$(date +%Y%m%d-%H%M%S).db
fi

# Extract new deployment
echo "Extracting new deployment..."
cd /home/deploy
rm -rf night-owls-go-new
mkdir night-owls-go-new
cd night-owls-go-new
tar -xzf /tmp/deploy.tar.gz

# Copy configuration if it doesn't exist
if [ ! -f "/home/deploy/night-owls-go/.env.production" ]; then
    echo "Warning: .env.production not found. Please configure it!"
fi

# Move configuration
if [ -f "/home/deploy/night-owls-go/.env.production" ]; then
    cp /home/deploy/night-owls-go/.env.production .
fi

# Atomic swap
echo "Swapping deployments..."
if [ -d "/home/deploy/night-owls-go" ]; then
    mv /home/deploy/night-owls-go /home/deploy/night-owls-go-old
fi
mv /home/deploy/night-owls-go-new /home/deploy/night-owls-go

# Make backend executable
chmod +x /home/deploy/night-owls-go/night-owls-server

# Run migrations
echo "Running database migrations..."
cd /home/deploy/night-owls-go
./night-owls-server -migrate-only

# Restart services
echo "Restarting services..."
sudo systemctl start night-owls-backend
sudo systemctl start night-owls-frontend

# Wait for services to start
sleep 5

# Health check
echo "Running health check..."
if curl -f http://localhost:5888/health > /dev/null 2>&1; then
    echo "✅ Health check passed!"
    # Clean up old deployment
    rm -rf /home/deploy/night-owls-go-old
    rm /tmp/deploy.tar.gz
else
    echo "❌ Health check failed! Rolling back..."
    sudo systemctl stop night-owls-backend night-owls-frontend
    rm -rf /home/deploy/night-owls-go
    mv /home/deploy/night-owls-go-old /home/deploy/night-owls-go
    sudo systemctl start night-owls-backend night-owls-frontend
    exit 1
fi

echo "🎉 Deployment successful!"
ENDSSH

# Clean up local files
rm deploy.tar.gz
rm night-owls-server

echo -e "${GREEN}✅ Deployment to $ENVIRONMENT complete!${NC}"

# Run smoke tests
echo -e "${GREEN}🧪 Running smoke tests...${NC}"
sleep 5
if [ "$ENVIRONMENT" = "production" ]; then
    HEALTH_URL="https://$REMOTE_HOST/health"
else
    HEALTH_URL="http://$REMOTE_HOST/health"
fi

if curl -f "$HEALTH_URL" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Remote health check passed!${NC}"
else
    echo -e "${YELLOW}⚠️  Remote health check failed. Please check the deployment.${NC}"
fi 