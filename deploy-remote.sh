#!/bin/bash
set -e

echo "🚀 Night Owls Control - Remote Deployment"
echo "=========================================="

# Configuration
REMOTE_HOST="mm.nightowls.app"
REMOTE_USER="deploy"
REMOTE_DIR="night-owls-go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we can connect to the server
print_status "Testing SSH connection to $REMOTE_HOST..."
if ! ssh -q $REMOTE_USER@$REMOTE_HOST exit; then
    print_error "Cannot connect to $REMOTE_HOST. Check your SSH configuration."
    exit 1
fi

print_success "SSH connection successful"

# Deploy via SSH
print_status "Deploying to $REMOTE_HOST..."
ssh $REMOTE_USER@$REMOTE_HOST << 'ENDSSH'
set -e

# Colors for remote output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[REMOTE]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[REMOTE]${NC} $1"
}

print_error() {
    echo -e "${RED}[REMOTE]${NC} $1"
}

print_status "Starting remote deployment..."

# Navigate to project directory
cd night-owls-go

# Pull latest changes
print_status "Pulling latest changes from GitHub..."
git pull origin main

# Check if .env.production exists
if [ ! -f ".env.production" ]; then
    print_error ".env.production not found! Please configure it first."
    exit 1
fi

# Run deployment
print_status "Running deployment script..."
./deploy-docker.sh

print_success "Remote deployment completed!"
ENDSSH

if [ $? -eq 0 ]; then
    print_success "Deployment completed successfully!"
    print_status "Application should be available at: https://mm.nightowls.app"
    
    # Test the deployment
    print_status "Testing deployment..."
    if curl -f -s https://mm.nightowls.app/health > /dev/null; then
        print_success "Health check passed! Deployment is working."
    else
        print_warning "Health check failed. Please check the application logs."
    fi
else
    print_error "Deployment failed. Check the output above for details."
    exit 1
fi 