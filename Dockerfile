# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /goatsync \
    ./cmd/server

# Final stage
FROM alpine:3.19

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

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set environment defaults
ENV PORT=8080 \
    CHUNK_STORAGE_PATH=/data/chunks

# Run
ENTRYPOINT ["/goatsync"]

