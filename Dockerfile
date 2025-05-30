# Build frontend
FROM node:20-alpine AS frontend-builder

# Add build argument for cache busting
ARG BUILD_DATE

# Install pnpm
RUN corepack enable pnpm

WORKDIR /app/frontend
COPY app/package.json app/pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

COPY app/ ./
# Force rebuild with timestamp
RUN echo "Building frontend at ${BUILD_DATE}" && pnpm run build

# Build backend
FROM golang:1.24-alpine AS backend-builder

# Install build dependencies (added git and ca-certificates)
RUN apk add --no-cache gcc musl-dev sqlite-dev git ca-certificates

# Set Go proxy and sumdb settings
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org
ENV GO111MODULE=on
ENV GONOSUMDB=github.com/twilio/*
ENV GOPRIVATE=
ENV CGO_ENABLED=1

WORKDIR /app
COPY go.mod go.sum ./

# Download modules with cache mount for faster rebuilds (requires BuildKit)
# Use DOCKER_BUILDKIT=1 docker build . for best performance
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN echo "Building Go application..." && \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o night-owls-server ./cmd/server

# Production image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create necessary directories
RUN mkdir -p /app/static /app/data /app/migrations
RUN chown -R appuser:appgroup /app

WORKDIR /app

# Copy migrations
COPY internal/db/migrations/ ./migrations/

# Copy backend binary
COPY --from=backend-builder /app/night-owls-server .

# Copy frontend build
COPY --from=frontend-builder /app/frontend/build ./static

# Debug: Show what was copied
RUN ls -la /app/static/ && echo "---" && find /app/static -type f -name "*.js" | head -20

# Set correct permissions for static files
RUN chown -R appuser:appgroup /app/static

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 5888

# Set environment
ENV TZ=UTC
ENV SERVER_PORT=5888
ENV STATIC_DIR=./static
ENV DATABASE_PATH=./data/production.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5888/health || exit 1

CMD ["./night-owls-server"] 