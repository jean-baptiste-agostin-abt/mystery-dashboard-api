# Mystery Factory API - Production Dockerfile
# Multi-stage build for optimized production image

# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build arguments for version information
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -a -installsuffix cgo \
    -o mysteryfactory-api \
    ./cmd/server

# Production stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/mysteryfactory-api .

# Copy configuration files
COPY --from=builder /app/prompts ./prompts

# Create necessary directories
RUN mkdir -p /app/logs /app/tmp && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV PORT=8080

# Run the application
CMD ["./mysteryfactory-api"]

# Labels for metadata
LABEL maintainer="Mystery Factory Team"
LABEL version="${VERSION}"
LABEL description="Mystery Factory API - AI-powered video content management platform"
LABEL org.opencontainers.image.source="https://github.com/jibe0123/mysteryfactory"
LABEL org.opencontainers.image.documentation="https://github.com/jibe0123/mysteryfactory/blob/main/README.md"
LABEL org.opencontainers.image.licenses="MIT"