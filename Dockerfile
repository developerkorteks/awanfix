# Multi-stage build for RcloneStorage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/rclonestorage cmd/server/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    curl \
    sqlite \
    rclone \
    tzdata \
    bash

# Create app user
RUN addgroup -g 1001 appgroup && \
    adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/rclonestorage .

# Copy configuration files and templates
COPY --from=builder /app/web ./web
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/.env.example ./.env

# Create necessary directories with proper permissions
RUN mkdir -p cache/{files,metadata,temp} data logs && \
    chown -R appuser:appgroup /app && \
    chmod +x ./rclonestorage

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 5601

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:5601/health || exit 1

# Start the application
CMD ["./rclonestorage"]