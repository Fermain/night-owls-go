#!/bin/bash
set -e

echo "🚀 Night Owls Control - pnpm Setup"
echo "=================================="

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

# Check if pnpm is installed
if ! command -v pnpm &> /dev/null; then
    print_status "pnpm not found. Installing pnpm..."
    
    # Install pnpm using corepack (recommended method)
    if command -v corepack &> /dev/null; then
        corepack enable pnpm
        print_success "pnpm installed via corepack"
    else
        # Fallback to npm installation
        npm install -g pnpm
        print_success "pnpm installed via npm"
    fi
else
    print_success "pnpm is already installed"
fi

# Install frontend dependencies
print_status "Installing frontend dependencies..."
cd app
pnpm install
print_success "Frontend dependencies installed"

# Generate VAPID keys if needed
print_status "Checking for VAPID keys configuration..."
if [ ! -f "../.env" ] || ! grep -q "VAPID_PUBLIC_KEY" ../.env; then
    print_warning "VAPID keys not found in .env file"
    print_status "Generating VAPID keys for web push notifications..."
    
    echo ""
    echo "🔑 Generated VAPID Keys:"
    echo "========================"
    pnpm dlx web-push generate-vapid-keys
    echo ""
    print_warning "Please add these keys to your .env file:"
    echo "VAPID_PUBLIC_KEY=<public_key_from_above>"
    echo "VAPID_PRIVATE_KEY=<private_key_from_above>"
    echo "VAPID_SUBJECT=mailto:admin@mm.nightowls.app"
else
    print_success "VAPID keys already configured"
fi

cd ..

print_success "Setup complete!"
print_status "You can now run:"
echo "  - Frontend dev: cd app && pnpm dev"
echo "  - Build frontend: cd app && pnpm build"
echo "  - Run tests: cd app && pnpm test"
echo "  - Deploy with Docker: ./deploy-docker.sh" 