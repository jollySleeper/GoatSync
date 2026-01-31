# Build stage
FROM golang:1.25-alpine AS builder

# Build arguments for version info
ARG VERSION=dev
ARG COMMIT=unknown

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with version info
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION} -X main.Commit=${COMMIT}" \
    -o /goatsync \
    ./cmd/server

# Final stage - using latest Alpine 3.21 (supported until Nov 2026)
FROM alpine:3.21

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' goatsync

# Copy binary from builder
COPY --from=builder /goatsync /goatsync

# Create data directories
RUN mkdir -p /data/chunks && chown -R goatsync:goatsync /data

# Switch to non-root user
USER goatsync

# Default port (can be overridden via PORT env var)
EXPOSE 3735

# Health check script that uses the PORT environment variable
# Note: The health check runs inside the container, so we use a shell
# to expand the PORT variable at runtime
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-3735}/health || exit 1

# Set environment defaults
ENV PORT=3735 \
    CHUNK_STORAGE_PATH=/data/chunks

# Run
ENTRYPOINT ["/goatsync"]

