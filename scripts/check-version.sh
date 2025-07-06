#!/bin/bash

# Night Owls Go - Version Check Script
# Usage: ./scripts/check-version.sh

set -e

# Colors for output
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

print_status "Checking Night Owls Go versions..."
echo ""

# Check if we're in the root directory
if [ ! -f "go.mod" ] || [ ! -f "app/package.json" ]; then
    print_warning "Please run this script from the root directory of the project"
    exit 1
fi

# Get versions from source files
FRONTEND_VERSION=$(node -p "require('./app/package.json').version")
BACKEND_VERSION=$(grep -o 'Version.*=.*"[^"]*"' internal/config/version.go | cut -d'"' -f2)
FRONTEND_CONFIG_VERSION=$(grep -o "APP_VERSION = '[^']*'" app/src/lib/config/version.ts | cut -d"'" -f2 2>/dev/null || echo "N/A")

echo "📦 Source Code Versions:"
echo "  Frontend (package.json): $FRONTEND_VERSION"
echo "  Frontend (config):       $FRONTEND_CONFIG_VERSION"
echo "  Backend (version.go):    $BACKEND_VERSION"

# Check for version consistency
if [ "$FRONTEND_VERSION" = "$BACKEND_VERSION" ] && [ "$FRONTEND_VERSION" = "$FRONTEND_CONFIG_VERSION" ]; then
    print_success "Versions are consistent across all files"
else
    print_warning "Version mismatch detected!"
    echo "  Frontend (package.json): $FRONTEND_VERSION"
    echo "  Frontend (config):       $FRONTEND_CONFIG_VERSION"
    echo "  Backend (version.go):    $BACKEND_VERSION"
fi

echo ""

# Check git tags
print_status "Git Information:"
CURRENT_BRANCH=$(git branch --show-current)
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "No tags found")
CURRENT_SHA=$(git rev-parse --short HEAD)

echo "  Current branch: $CURRENT_BRANCH"
echo "  Latest tag:     $LATEST_TAG"
echo "  Current SHA:    $CURRENT_SHA"

# Check if there are uncommitted changes
if [ -n "$(git status --porcelain)" ]; then
    print_warning "There are uncommitted changes"
else
    print_success "Git working directory is clean"
fi

echo ""

# Check if running server is available
print_status "Runtime Version Check:"
if command -v curl &> /dev/null; then
    # Try to check local development server
    if curl -s http://localhost:5888/api/health &>/dev/null; then
        echo "🔍 Local development server (http://localhost:5888):"
        if command -v jq &> /dev/null; then
            curl -s http://localhost:5888/api/health | jq -r '.build.version // "Version not available"' 2>/dev/null || echo "  Could not parse version"
        else
            echo "  jq not available, skipping version parsing"
        fi
    else
        echo "  Local development server not running"
    fi
    
    # Try to check production server
    if curl -s https://mm.nightowls.app/api/health &>/dev/null; then
        echo "🌐 Production server (https://mm.nightowls.app):"
        if command -v jq &> /dev/null; then
            curl -s https://mm.nightowls.app/api/health | jq -r '.build.version // "Version not available"' 2>/dev/null || echo "  Could not parse version"
        else
            echo "  jq not available, skipping version parsing"
        fi
    else
        echo "  Production server not reachable"
    fi
else
    print_warning "curl not available, skipping runtime version check"
fi

echo ""

# Check Docker image versions if available
print_status "Docker Information:"
if command -v docker &> /dev/null; then
    # Check if there are local images
    LOCAL_IMAGES=$(docker images ghcr.io/*/night-owls-go --format "table {{.Tag}}" --no-trunc 2>/dev/null | tail -n +2 | head -5)
    if [ -n "$LOCAL_IMAGES" ]; then
        echo "🐳 Local Docker images:"
        echo "$LOCAL_IMAGES" | sed 's/^/  /'
    else
        echo "  No local Docker images found"
    fi
else
    print_warning "Docker not available, skipping Docker version check"
fi

echo ""

# Summary
echo "📋 Version Summary:"
echo "  Source Version:    $FRONTEND_VERSION"
echo "  Git Latest Tag:    $LATEST_TAG"
echo "  Current Branch:    $CURRENT_BRANCH"
echo "  Current SHA:       $CURRENT_SHA"

if [ "$FRONTEND_VERSION" = "$BACKEND_VERSION" ]; then
    if [ "v$FRONTEND_VERSION" = "$LATEST_TAG" ]; then
        print_success "All versions are consistent and tagged"
    else
        print_warning "Version not tagged yet - run: git tag v$FRONTEND_VERSION"
    fi
else
    print_warning "Version inconsistency detected - run: ./scripts/bump-version.sh $FRONTEND_VERSION"
fi 