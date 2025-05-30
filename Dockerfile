# Build frontend
FROM node:20-alpine AS frontend-builder

# Install pnpm globally for better caching
RUN corepack enable pnpm

WORKDIR /app/frontend

# Copy package files first for better layer caching
COPY app/package.json app/pnpm-lock.yaml* ./

# Install dependencies with cache mount
RUN --mount=type=cache,target=/root/.local/share/pnpm \
    pnpm install --frozen-lockfile

# Copy source and build
COPY app/ ./
RUN pnpm run build

# Build backend
FROM golang:1.24-alpine AS backend-builder

# Install minimal build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Optimize Go build environment
ENV GOPROXY=https://proxy.golang.org,direct
ENV CGO_ENABLED=1

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download modules with cache mount
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy source and build with optimizations
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o night-owls-server ./cmd/server

# Production image - use minimal distroless
FROM alpine:latest

# Install only essential runtime dependencies
RUN apk --no-cache add ca-certificates sqlite tzdata wget

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Create app directory
WORKDIR /app
RUN chown appuser:appgroup /app

# Copy migrations
COPY internal/db/migrations/ ./migrations/

# Copy binaries
COPY --from=backend-builder /app/night-owls-server .

# Copy frontend build (we're still including it even though Caddy serves it)
COPY --from=frontend-builder /app/frontend/build ./static

# Set permissions
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 5888

# Set environment
ENV TZ=UTC \
    SERVER_PORT=5888 \
    DATABASE_PATH=./data/production.db

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:5888/health || exit 1

CMD ["./night-owls-server"] 