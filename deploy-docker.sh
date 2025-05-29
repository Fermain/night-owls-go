#!/bin/bash
set -e

echo "🐳 Night Owls Control - Low Memory Docker Deployment"
echo "=========================================="

# Configuration
DOCKER_IMAGE="night-owls-go"
DOCKER_TAG="latest"
COMPOSE_FILE="docker-compose.yml"

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

# Check if .env.production exists
if [ ! -f ".env.production" ]; then
    print_warning ".env.production not found"
    print_status "Creating from template..."
    cp .env.production.example .env.production
    print_warning "Please edit .env.production with your actual values before continuing"
    print_status "Required variables to update:"
    echo "  - JWT_SECRET (generate with: openssl rand -base64 32)"
    echo "  - VAPID_PUBLIC_KEY and VAPID_PRIVATE_KEY (generate with: pnpm dlx web-push generate-vapid-keys)"
    echo "  - Twilio credentials for SMS"
    exit 1
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    print_error "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if docker compose is available (try both V2 and legacy)
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
    print_status "Using Docker Compose V2 (plugin)"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE="docker-compose"
    print_status "Using Docker Compose V1 (legacy)"
else
    print_error "Docker Compose not found. Please install docker-compose or docker compose plugin."
    exit 1
fi

print_status "Stopping existing containers..."
$DOCKER_COMPOSE -f $COMPOSE_FILE --env-file .env.production down

print_status "Cleaning up Docker to free memory..."
docker system prune -f

print_status "Building Docker image with low-memory optimizations..."
docker build --memory="900m" --memory-swap="4g" -f Dockerfile.lowmem -t $DOCKER_IMAGE:$DOCKER_TAG .

print_status "Starting services..."
$DOCKER_COMPOSE -f $COMPOSE_FILE --env-file .env.production up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
timeout=120
elapsed=0
while [ $elapsed -lt $timeout ]; do
    if $DOCKER_COMPOSE -f $COMPOSE_FILE ps | grep -q "healthy"; then
        break
    fi
    sleep 5
    elapsed=$((elapsed + 5))
    print_status "Waiting... ($elapsed/${timeout}s)"
done

# Check final status
print_status "Checking service status..."
$DOCKER_COMPOSE -f $COMPOSE_FILE ps

# Test health endpoint
print_status "Testing health endpoint..."
if curl -f -s http://localhost/health > /dev/null; then
    print_success "Health check passed!"
else
    print_warning "Health check failed. Check logs with: $DOCKER_COMPOSE logs"
fi

# Show logs
print_status "Recent logs:"
$DOCKER_COMPOSE -f $COMPOSE_FILE logs --tail=20

print_success "Deployment complete!"
print_status "Application should be available at: https://mm.nightowls.app"
print_status "Local access (if testing): http://localhost"
print_status ""
print_status "Useful commands:"
echo "  - View logs: $DOCKER_COMPOSE logs -f"
echo "  - Stop services: $DOCKER_COMPOSE down"
echo "  - Restart: $DOCKER_COMPOSE restart"
echo "  - Update: ./deploy-docker.sh" 