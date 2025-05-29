#!/bin/bash
# Build locally and push to Docker Hub

# Set your Docker Hub username (replace with your actual username)
DOCKER_USERNAME="${DOCKER_USERNAME:-nightowlsapp}"  # Change this to your Docker Hub username!
IMAGE_NAME="night-owls-go"
TAG="latest"

# Clean local Docker first
echo "Cleaning local Docker to free space..."
docker system prune -f

# Ensure we're logged in
echo "Logging in to Docker Hub..."
docker login

echo "Building image locally with optimized Dockerfile..."
docker build -f Dockerfile.prod -t $DOCKER_USERNAME/$IMAGE_NAME:$TAG .

if [ $? -eq 0 ]; then
    echo "Pushing image to Docker Hub..."
    docker push $DOCKER_USERNAME/$IMAGE_NAME:$TAG
    
    echo -e "\n✅ Success! Image pushed to Docker Hub."
    echo -e "\nTo deploy on your server, run:"
    echo "DOCKER_USERNAME=$DOCKER_USERNAME ./deploy-remote.sh --dockerhub"
else
    echo "❌ Build failed!"
    exit 1
fi 