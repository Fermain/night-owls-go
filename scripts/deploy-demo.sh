#!/bin/bash

# Night Owls Demo Instance Management Script
# Usage: ./scripts/deploy-demo.sh [start|stop|reset|logs|status]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DEMO_COMPOSE="$PROJECT_ROOT/docker-compose.demo.yml"
MULTI_CADDY="$PROJECT_ROOT/Caddyfile.multi"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️ $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

check_dependencies() {
    if ! command -v docker &> /dev/null; then
        error "Docker is not installed or not in PATH"
    fi
    
    if ! docker compose version &> /dev/null; then
        error "Docker Compose is not available"
    fi
    
    if [ ! -f "$DEMO_COMPOSE" ]; then
        error "Demo compose file not found: $DEMO_COMPOSE"
    fi
}

start_demo() {
    log "Starting Night Owls Demo Instance..."
    
    # Check if production is running and warn
    if docker compose ps | grep -q "night-owls"; then
        warning "Production instance appears to be running"
        warning "Demo will run on different ports (5889 for backend)"
    fi
    
    # Start demo containers
    log "Starting demo containers..."
    docker compose -f "$DEMO_COMPOSE" up -d
    
    # Wait for health check
    log "Waiting for demo instance to be healthy..."
    sleep 10
    
    for i in {1..12}; do
        if docker compose -f "$DEMO_COMPOSE" ps | grep -q "healthy"; then
            success "Demo instance is healthy!"
            break
        fi
        log "Waiting for demo to be ready... ($i/12)"
        sleep 5
    done
    
    # Seed data
    log "Seeding demo data (50 users, future bookings)..."
    docker compose -f "$DEMO_COMPOSE" run --rm night-owls-demo-seed \
        ./seed --reset --users 50 --future-bookings --verbose || warning "Seeding completed with warnings"
    
    success "Demo instance started successfully!"
    echo
    echo "📍 Demo backend: http://localhost:5889"
    echo "🏥 Health check: http://localhost:5889/health"
    echo "📚 API docs: http://localhost:5889/swagger/"
    echo "👤 Admin login: +27821234567 (any 6-digit OTP)"
    echo "🎮 Dev mode enabled - any OTP code will work"
}

stop_demo() {
    log "Stopping Night Owls Demo Instance..."
    docker compose -f "$DEMO_COMPOSE" down
    success "Demo instance stopped"
}

reset_demo() {
    log "Resetting Night Owls Demo Instance..."
    
    # Stop containers
    docker compose -f "$DEMO_COMPOSE" down
    
    # Remove demo data volume
    docker volume rm night_owls_demo_data 2>/dev/null || warning "Demo data volume didn't exist"
    
    # Start fresh
    start_demo
    
    success "Demo instance reset complete!"
}

show_logs() {
    log "Showing demo instance logs..."
    docker compose -f "$DEMO_COMPOSE" logs -f
}

show_status() {
    log "Demo instance status:"
    echo
    
    if docker compose -f "$DEMO_COMPOSE" ps | grep -q "night-owls-demo"; then
        docker compose -f "$DEMO_COMPOSE" ps
        echo
        
        # Test backend health
        if curl -s http://localhost:5889/health >/dev/null 2>&1; then
            success "Backend is healthy"
        else
            error "Backend is not responding"
        fi
        
        # Show some stats
        CONTAINER_ID=$(docker compose -f "$DEMO_COMPOSE" ps -q night-owls-demo)
        if [ -n "$CONTAINER_ID" ]; then
            echo
            log "Container stats:"
            docker stats --no-stream "$CONTAINER_ID"
        fi
    else
        warning "Demo instance is not running"
        echo "Run './scripts/deploy-demo.sh start' to start it"
    fi
}

# Main script logic
case "${1:-}" in
    start)
        check_dependencies
        start_demo
        ;;
    stop)
        check_dependencies
        stop_demo
        ;;
    reset)
        check_dependencies
        reset_demo
        ;;
    logs)
        check_dependencies
        show_logs
        ;;
    status)
        check_dependencies
        show_status
        ;;
    *)
        echo "Night Owls Demo Instance Management"
        echo
        echo "Usage: $0 [command]"
        echo
        echo "Commands:"
        echo "  start   Start the demo instance with fresh seed data"
        echo "  stop    Stop the demo instance"
        echo "  reset   Stop, clear data, and restart with fresh seed"
        echo "  logs    Show demo instance logs (follow mode)"
        echo "  status  Show current status and health"
        echo
        echo "Demo Features:"
        echo "  • 50 users with realistic roles (2 admins, ~40 owls, ~8 guests)"
        echo "  • Historical bookings with attendance patterns"
        echo "  • Future bookings for next 30 days"
        echo "  • Incident reports with realistic severity distribution"
        echo "  • Dev mode enabled (any OTP works)"
        echo "  • Extended JWT expiration (1 week)"
        echo
        echo "Access URLs:"
        echo "  Backend:    http://localhost:5889"
        echo "  Health:     http://localhost:5889/health"
        echo "  API Docs:   http://localhost:5889/swagger/"
        echo
        echo "Demo Login:"
        echo "  Admin:      +27821234567 (Alice Admin)"
        echo "  Manager:    +27821234568 (Bob Manager)"
        echo "  OTP:        Any 6 digits (dev mode)"
        ;;
esac 