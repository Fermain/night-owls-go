#!/bin/bash

# Manual deployment script for GitHub Container Registry
# Usage: ./deploy-manual.sh [image-tag]

set -e

# Configuration
REPO_NAME="night-owls-go"  # Change this to your actual repo name
GITHUB_USER="fermain"  # Change this to your GitHub username
IMAGE_TAG="${1:-latest}"
SERVER="mm.nightowls.app"
SERVER_USER="deploy"

echo "🚀 Deploying $GITHUB_USER/$REPO_NAME:$IMAGE_TAG to $SERVER"

# Build frontend locally first
echo "🏗️  Building frontend locally..."
cd app
if command -v pnpm &> /dev/null; then
    echo "📦 Installing dependencies with pnpm..."
    pnpm install
    echo "🔨 Building frontend..."
    pnpm run build
else
    echo "📦 Installing dependencies with npm..."
    npm install
    echo "🔨 Building frontend..."
    npm run build
fi

# Verify build directory exists
if [ ! -d "build" ]; then
    echo "❌ Frontend build failed - build directory not found"
    exit 1
fi

echo "✅ Frontend build complete"
cd ..

echo "📦 Copying frontend build and config to server..."
# Create the directory structure on the server
ssh $SERVER_USER@$SERVER "mkdir -p ~/night-owls-go/frontend-build"
# Copy frontend build files
rsync -avz --delete app/build/ $SERVER_USER@$SERVER:~/night-owls-go/frontend-build/
# Copy docker-compose.yml if it doesn't exist or is newer
rsync -avz docker-compose.yml $SERVER_USER@$SERVER:~/night-owls-go/
# Copy Caddyfile
rsync -avz Caddyfile $SERVER_USER@$SERVER:~/night-owls-go/

# Deploy to server
ssh $SERVER_USER@$SERVER << 'EOF'
set -e

cd ~/night-owls-go

echo "📦 Stopping existing containers..."
# Detect Docker Compose version
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

$DOCKER_COMPOSE down 2>/dev/null || true

echo "🧹 Cleaning up old images..."
docker system prune -f

echo "⬇️  Pulling new image from GitHub Container Registry..."
docker pull ghcr.io/fermain/night-owls-go:latest

echo "📋 Verifying config files..."
if [ ! -f docker-compose.yml ]; then
    echo "❌ docker-compose.yml not found"
    exit 1
fi

if [ ! -f .env.production ]; then
    echo "❌ .env.production not found. Please copy it to ~/night-owls-go/"
    echo "   scp .env.production deploy@mm.nightowls.app:~/night-owls-go/"
    exit 1
fi

if [ ! -d frontend-build ]; then
    echo "❌ frontend-build directory not found"
    exit 1
fi

echo "🔍 Frontend files check:"
ls -la frontend-build/ | head -10

echo "🚀 Starting application..."
$DOCKER_COMPOSE up -d

echo "⏳ Waiting for application to start..."
sleep 15

echo "🔍 Health check..."
if curl -f http://localhost:5888/health 2>/dev/null; then
    echo "✅ Application is healthy!"
else
    echo "⚠️  Health check failed, checking container logs..."
    $DOCKER_COMPOSE logs --tail=20
fi

echo "📊 Container status:"
$DOCKER_COMPOSE ps

echo ""
echo "🎉 Deployment complete!"
echo "🌐 Application should be available at: https://mm.nightowls.app"
EOF

echo ""
echo "✅ Manual deployment finished!"
echo "💡 For automated deployments, just push to main branch" 