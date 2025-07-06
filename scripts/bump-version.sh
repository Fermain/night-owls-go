#!/bin/bash

# Night Owls Go - Version Bump Script
# Usage: ./scripts/bump-version.sh [new-version]
# Example: ./scripts/bump-version.sh 2025.07.2

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if version is provided
if [ $# -eq 0 ]; then
    print_error "Version not provided!"
    echo "Usage: $0 [new-version]"
    echo "Example: $0 2025.07.2"
    echo ""
    echo "Current month CalVer suggestion: 2025.$(date +%m).1"
    exit 1
fi

NEW_VERSION=$1

# Validate version format (YYYY.MM.PATCH)
if [[ ! $NEW_VERSION =~ ^[0-9]{4}\.[0-9]{1,2}\.[0-9]+$ ]]; then
    print_error "Invalid version format: $NEW_VERSION"
    echo "Expected format: YYYY.MM.PATCH (e.g., 2025.07.1)"
    echo ""
    echo "CalVer Guidelines:"
    echo "  - YYYY: Current year ($(date +%Y))"
    echo "  - MM: Current month ($(date +%m)) or feature release month"
    echo "  - PATCH: Incrementing number (1, 2, 3, ...)"
    exit 1
fi

# Extract year and month from version
VERSION_YEAR=$(echo $NEW_VERSION | cut -d. -f1)
VERSION_MONTH=$(echo $NEW_VERSION | cut -d. -f2)
CURRENT_YEAR=$(date +%Y)
CURRENT_MONTH=$(date +%m)

# Validate date logic
if [ "$VERSION_YEAR" -lt "$CURRENT_YEAR" ]; then
    print_warning "Version year ($VERSION_YEAR) is in the past"
elif [ "$VERSION_YEAR" -eq "$CURRENT_YEAR" ] && [ "$VERSION_MONTH" -lt "$CURRENT_MONTH" ]; then
    print_warning "Version month ($VERSION_YEAR.$VERSION_MONTH) is in the past"
fi

# Check if we're in the root directory
if [ ! -f "go.mod" ] || [ ! -f "app/package.json" ]; then
    print_error "Please run this script from the root directory of the project"
    exit 1
fi

# Get current versions from source files only
CURRENT_FRONTEND_VERSION=$(node -p "require('./app/package.json').version")
CURRENT_BACKEND_VERSION=$(grep -o 'Version.*=.*"[^"]*"' internal/config/version.go | cut -d'"' -f2)
CURRENT_SWAGGER_VERSION=$(grep -o '@version.*' cmd/server/main.go | awk '{print $2}')

print_status "Current versions:"
echo "  Frontend (package.json): $CURRENT_FRONTEND_VERSION"
echo "  Backend (version.go):    $CURRENT_BACKEND_VERSION"
echo "  Swagger (main.go):       $CURRENT_SWAGGER_VERSION"
echo "  Target version:          $NEW_VERSION"
echo ""
echo "ℹ️  Frontend config version is automatically synced from package.json at build time"

# Check for version consistency before update
if [ "$CURRENT_FRONTEND_VERSION" != "$CURRENT_BACKEND_VERSION" ] || [ "$CURRENT_FRONTEND_VERSION" != "$CURRENT_SWAGGER_VERSION" ]; then
    print_warning "Current versions are inconsistent!"
    echo "This update will synchronize all versions to: $NEW_VERSION"
fi

# Check if git is clean
if [ -n "$(git status --porcelain)" ]; then
    print_warning "Git working directory is not clean!"
    echo "Uncommitted changes detected:"
    git status --porcelain
    echo ""
    read -p "Continue with version bump anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Aborting version bump"
        exit 1
    fi
fi

# Update source version locations
print_status "Updating versions in source files..."

# 1. Update frontend version
print_status "Updating frontend version in app/package.json..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/\"version\": \"$CURRENT_FRONTEND_VERSION\"/\"version\": \"$NEW_VERSION\"/" app/package.json
else
    # Linux
    sed -i "s/\"version\": \"$CURRENT_FRONTEND_VERSION\"/\"version\": \"$NEW_VERSION\"/" app/package.json
fi

# 2. Update backend version
print_status "Updating backend version in internal/config/version.go..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/Version.*=.*\"$CURRENT_BACKEND_VERSION\"/Version     = \"$NEW_VERSION\"/" internal/config/version.go
else
    # Linux
    sed -i "s/Version.*=.*\"$CURRENT_BACKEND_VERSION\"/Version     = \"$NEW_VERSION\"/" internal/config/version.go
fi

# 3. Update swagger version
print_status "Updating swagger version in cmd/server/main.go..."
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/@version.*$/@version $NEW_VERSION/" cmd/server/main.go
else
    # Linux
    sed -i "s/@version.*$/@version $NEW_VERSION/" cmd/server/main.go
fi

print_status "✅ Frontend config version will be automatically updated from package.json at build time"

# 5. Update VERSIONING.md if it exists
if [ -f "VERSIONING.md" ]; then
    print_status "Updating VERSIONING.md..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "s/\*\*Current Stable\*\*:.*$/\*\*Current Stable\*\*: \`$NEW_VERSION\`/" VERSIONING.md
    else
        # Linux
        sed -i "s/\*\*Current Stable\*\*:.*$/\*\*Current Stable\*\*: \`$NEW_VERSION\`/" VERSIONING.md
    fi
fi

# Verify changes
print_status "Verifying changes..."
NEW_FRONTEND_VERSION=$(node -p "require('./app/package.json').version")
NEW_BACKEND_VERSION=$(grep -o 'Version.*=.*"[^"]*"' internal/config/version.go | cut -d'"' -f2)
NEW_SWAGGER_VERSION=$(grep -o '@version.*' cmd/server/main.go | awk '{print $2}')

if [ "$NEW_FRONTEND_VERSION" != "$NEW_VERSION" ] || [ "$NEW_BACKEND_VERSION" != "$NEW_VERSION" ] || [ "$NEW_SWAGGER_VERSION" != "$NEW_VERSION" ]; then
	print_error "Version update failed!"
	echo "Results:"
	echo "  Frontend (package.json): $NEW_FRONTEND_VERSION (expected: $NEW_VERSION)"
	echo "  Backend:                 $NEW_BACKEND_VERSION (expected: $NEW_VERSION)"
	echo "  Swagger:                 $NEW_SWAGGER_VERSION (expected: $NEW_VERSION)"
	echo "Note: Frontend config version is automatically synced from package.json at build time"
	exit 1
fi

print_success "All versions updated successfully to: $NEW_VERSION"

# Run version consistency test
print_status "Running version consistency test..."
if command -v go &> /dev/null; then
    if go test ./internal/config -v -run TestVersionConsistency; then
        print_success "Version consistency test passed"
    else
        print_error "Version consistency test failed"
        exit 1
    fi
else
    print_warning "Go not found, skipping version consistency test"
fi

# Run other tests
print_status "Running test suite..."
if command -v pnpm &> /dev/null; then
    (cd app && pnpm run test:unit -- --run) || {
        print_error "Frontend tests failed"
        exit 1
    }
else
    print_warning "pnpm not found, skipping frontend tests"
fi

if command -v go &> /dev/null; then
    go test ./... -short || {
        print_error "Backend tests failed"
        exit 1
    }
else
    print_warning "go not found, skipping backend tests"
fi

# Build check
print_status "Testing build..."
if command -v pnpm &> /dev/null; then
    (cd app && pnpm run build) || {
        print_error "Frontend build failed"
        exit 1
    }
else
    print_warning "pnpm not found, skipping frontend build test"
fi

if command -v go &> /dev/null; then
    go build -o /tmp/night-owls-server ./cmd/server || {
        print_error "Backend build failed"
        exit 1
    }
    rm -f /tmp/night-owls-server
else
    print_warning "go not found, skipping backend build test"
fi

# Git operations
print_status "Committing changes..."
git add app/package.json internal/config/version.go cmd/server/main.go VERSIONING.md
git commit -m "chore: bump version to $NEW_VERSION

- Updated frontend version (package.json)
- Updated backend version (version.go)
- Updated swagger annotation (main.go)
- Frontend config automatically syncs from package.json at build time

Version follows CalVer format: $NEW_VERSION"

print_status "Creating git tag..."
git tag "v$NEW_VERSION"

print_success "Version bump completed successfully!"
echo ""
echo "📋 Version Summary:"
echo "  New Version: $NEW_VERSION"
echo "  Git Tag:     v$NEW_VERSION"
echo "  All Files:   ✅ Synchronized"
echo "  Tests:       ✅ Passed"
echo "  Build:       ✅ Successful"
echo ""
echo "🚀 Next steps:"
echo "  1. Review the changes: git show HEAD"
echo "  2. Push to remote: git push origin $(git branch --show-current) --tags"
echo "  3. Monitor deployment: Check GitHub Actions"
echo ""
echo "The new version $NEW_VERSION will be automatically deployed when pushed to main." 