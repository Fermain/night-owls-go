# Build backend only (frontend now built locally for deployment efficiency)
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

# Build arguments for version information
ARG GIT_SHA=""
ARG BUILD_TIME=""

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -linkmode external -extldflags '-static' -X main.GitSHA=${GIT_SHA} -X main.BuildTime=${BUILD_TIME}" \
    -o night-owls-server ./cmd/server

# Build migration tool
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-w -s -linkmode external -extldflags '-static'" \
    -o migrate-points ./cmd/migrate-points

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
COPY internal/db/migrations/ ./internal/db/migrations/

# Copy binaries
COPY --from=backend-builder /app/night-owls-server .
COPY --from=backend-builder /app/migrate-points .

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