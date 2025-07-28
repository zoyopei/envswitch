# Build stage
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git (needed for go mod download)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o envswitch .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN adduser -D -s /bin/sh envswitch

# Set working directory
WORKDIR /home/envswitch

# Copy binary from builder stage
COPY --from=builder /app/envswitch .

# Copy web templates and static files
COPY --from=builder /app/web ./web

# Create necessary directories
RUN mkdir -p data backups && \
    chown -R envswitch:envswitch .

# Switch to non-root user
USER envswitch

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Set default command
CMD ["./envswitch", "server", "--port", "8080"] 