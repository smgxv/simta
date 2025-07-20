# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod tidy

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy static files
COPY static/ static/

# Create necessary directories and set permissions
RUN mkdir -p cert && \
    chmod 755 /app && \
    chmod 755 /app/static && \
    chmod 755 /app/cert

# Add debugging utilities
RUN apk add --no-cache curl tzdata ca-certificates openssl netcat-openbsd

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD nc -zv localhost 8443 || exit 1

# Expose ports
EXPOSE 8080 8443

# Start script
COPY <<'EOF' /app/start.sh
#!/bin/sh
echo "Checking SSL certificates..."
ls -la /app/cert
echo "Testing ports..."
nc -zv localhost 8443 || echo "Warning: Port 8443 not responding"
echo "Starting server..."
./main
EOF

RUN chmod +x /app/start.sh

# Command to run the application
CMD ["/app/start.sh"]