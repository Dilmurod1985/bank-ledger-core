# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
ENV GOPROXY=direct
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Create data directory for SQLite with proper permissions
RUN mkdir -p /app/data && \
    chmod 755 /app/data

# Create a startup script to handle SQLite migration issues
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'set -e' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Remove existing database if it exists to prevent migration conflicts' >> /app/entrypoint.sh && \
    echo 'if [ -f "/app/data/bank_ledger.db" ]; then' >> /app/entrypoint.sh && \
    echo '    echo "Removing existing database to prevent migration conflicts..."' >> /app/entrypoint.sh && \
    echo '    rm -f /app/data/bank_ledger.db' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Set database path environment variable' >> /app/entrypoint.sh && \
    echo 'export DB_PATH="/app/data/bank_ledger.db"' >> /app/entrypoint.sh && \
    echo '' >> /app/entrypoint.sh && \
    echo '# Start the application' >> /app/entrypoint.sh && \
    echo 'exec ./main' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Use the entrypoint script
ENTRYPOINT ["/app/entrypoint.sh"]
