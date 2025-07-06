#!/bin/bash

# Night Owls Production Deployment Script
# Usage: ./scripts/deploy.sh [version|latest]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REGISTRY="ghcr.io"
IMAGE_NAME="fermain/night-owls-go"
COMPOSE_FILE="docker-compose.yml"

print_usage() {
    echo "Usage: $0 [version|latest]"
    echo ""
    echo "Examples:"
    echo "  $0 2025.07.1    # Deploy specific version"
    echo "  $0 latest       # Deploy latest tag (not recommended)"
    echo "  $0              # Deploy version from package.json"
    echo ""
    echo "Options:"
    echo "  --demo          # Deploy to demo environment"
    echo "  --rollback      # Show available versions for rollback"
    echo "  --status        # Show current running version"
}

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

get_package_version() {
    if [ -f "app/package.json" ]; then
        node -p "require('./app/package.json').version"
    else
        print_error "app/package.json not found"
        exit 1
    fi
}

get_current_version() {
    local container_name=${1:-"night-owls-go"}
    docker inspect $container_name --format='{{index .Config.Image}}' 2>/dev/null | cut -d':' -f2 || echo "unknown"
}

show_status() {
    print_status "Current deployment status:"
    echo ""
    
    local prod_version=$(get_current_version "night-owls-go")
    local demo_version=$(get_current_version "night-owls-demo")
    
    echo "🏭 Production: $prod_version"
    echo "🎭 Demo:       $demo_version"
    echo ""
    
    if docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" | grep -E "(night-owls|caddy)"; then
        print_success "Services are running"
    else
        print_warning "Some services may not be running"
    fi
}

show_available_versions() {
    print_status "Fetching available versions from registry..."
    
    # This requires GitHub CLI or curl to query the registry
    if command -v gh &> /dev/null; then
        gh api repos/fermain/night-owls-go/packages/container/night-owls-go/versions --paginate | jq -r '.[].metadata.container.tags[]' | grep -E '^[0-9]{4}\.[0-9]{2}\.[0-9]+$' | sort -V | tail -10
    else
        print_warning "GitHub CLI not available. Install 'gh' to see available versions."
        print_status "Recent versions should follow pattern: YYYY.MM.PATCH (e.g., 2025.07.1)"
    fi
}

deploy_version() {
    local version=$1
    local is_demo=${2:-false}
    
    if [ "$is_demo" = true ]; then
        local compose_file="docker-compose.demo.yml"
        local env_name="demo"
    else
        local compose_file="docker-compose.yml"
        local env_name="production"
    fi
    
    print_status "Deploying version $version to $env_name..."
    
    # Check if image exists
    local full_image="$REGISTRY/$IMAGE_NAME:$version"
    print_status "Checking if image exists: $full_image"
    
    if ! docker manifest inspect $full_image &>/dev/null; then
        print_error "Image $full_image not found in registry"
        print_status "Available versions:"
        show_available_versions
        exit 1
    fi
    
    # Pull the specific image
    print_status "Pulling image: $full_image"
    docker pull $full_image
    
    # Stop existing services
    print_status "Stopping existing services..."
    docker compose -f $compose_file down || true
    
    # Start with specific image tag
    print_status "Starting services with version $version..."
    IMAGE_TAG=$version docker compose -f $compose_file up -d --force-recreate
    
    # Wait for health check
    print_status "Waiting for services to be healthy..."
    sleep 15
    
    # Verify deployment
    if [ "$is_demo" = true ]; then
        local health_url="http://localhost:5889/health"
    else
        local health_url="http://localhost:5888/health"
    fi
    
    if curl -f $health_url &>/dev/null; then
        print_success "Deployment successful! Version $version is running on $env_name"
    else
        print_error "Deployment failed - health check failed"
        print_status "Container logs:"
        docker compose -f $compose_file logs --tail=20
        exit 1
    fi
}

# Parse arguments
DEMO=false
ROLLBACK=false
STATUS=false
VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --demo)
            DEMO=true
            shift
            ;;
        --rollback)
            ROLLBACK=true
            shift
            ;;
        --status)
            STATUS=true
            shift
            ;;
        --help|-h)
            print_usage
            exit 0
            ;;
        *)
            if [ -z "$VERSION" ]; then
                VERSION=$1
            else
                print_error "Unknown option: $1"
                print_usage
                exit 1
            fi
            shift
            ;;
    esac
done

# Handle different modes
if [ "$STATUS" = true ]; then
    show_status
    exit 0
fi

if [ "$ROLLBACK" = true ]; then
    show_available_versions
    exit 0
fi

# Determine version to deploy
if [ -z "$VERSION" ]; then
    VERSION=$(get_package_version)
    print_status "No version specified, using package.json version: $VERSION"
fi

# Deploy
deploy_version "$VERSION" "$DEMO"