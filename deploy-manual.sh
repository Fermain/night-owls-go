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

# Deploy to server
ssh $SERVER_USER@$SERVER << EOF
set -e

cd ~/night-owls-app 2>/dev/null || mkdir -p ~/night-owls-app
cd ~/night-owls-app

echo "📦 Stopping existing containers..."
# Detect Docker Compose version
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE="docker compose"
else
    DOCKER_COMPOSE="docker-compose"
fi

\$DOCKER_COMPOSE down 2>/dev/null || true

echo "🧹 Cleaning up old images..."
docker system prune -f

echo "⬇️  Pulling new image from GitHub Container Registry..."
docker pull ghcr.io/$GITHUB_USER/$REPO_NAME:$IMAGE_TAG

echo "🏷️  Tagging image for docker-compose..."
docker tag ghcr.io/$GITHUB_USER/$REPO_NAME:$IMAGE_TAG night-owls-go:latest

echo "📋 Ensuring config files are present..."
if [ ! -f docker-compose.yml ]; then
    echo "❌ docker-compose.yml not found. Please copy it to ~/night-owls-app/"
    echo "   scp docker-compose.yml $SERVER_USER@$SERVER:~/night-owls-app/"
    exit 1
fi

if [ ! -f .env.production ]; then
    echo "❌ .env.production not found. Please copy it to ~/night-owls-app/"
    echo "   scp .env.production $SERVER_USER@$SERVER:~/night-owls-app/"
    exit 1
fi

echo "🚀 Starting application..."
\$DOCKER_COMPOSE up -d

echo "⏳ Waiting for application to start..."
sleep 10

echo "🔍 Health check..."
if curl -f http://localhost:5888/health 2>/dev/null; then
    echo "✅ Application is healthy!"
else
    echo "⚠️  Health check failed, but container might still be starting..."
fi

echo "📊 Container status:"
\$DOCKER_COMPOSE ps

echo ""
echo "🎉 Deployment complete!"
echo "🌐 Application should be available at: https://mm.nightowls.app"
EOF

echo ""
echo "✅ Manual deployment finished!"
echo "💡 For automated deployments, just push to main branch" 