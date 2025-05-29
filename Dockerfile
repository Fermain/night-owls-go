# Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend
COPY app/package*.json ./
RUN npm ci --production=false

COPY app/ ./
RUN npm run build

# Build backend
FROM golang:1.24-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o night-owls-server ./cmd/server

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