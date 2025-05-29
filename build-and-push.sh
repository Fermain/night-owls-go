#!/bin/bash
# Build locally and push to Docker Hub

# Set your Docker Hub username
DOCKER_USERNAME="your-dockerhub-username"
IMAGE_NAME="night-owls-go"
TAG="latest"

echo "Building image locally..."
docker build -f Dockerfile.prod -t $DOCKER_USERNAME/$IMAGE_NAME:$TAG .

echo "Logging in to Docker Hub..."
docker login

echo "Pushing image..."
docker push $DOCKER_USERNAME/$IMAGE_NAME:$TAG

echo "Image pushed successfully!"
echo "On the server, run: docker pull $DOCKER_USERNAME/$IMAGE_NAME:$TAG" 