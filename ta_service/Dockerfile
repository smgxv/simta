# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install runtime dependencies and tools
RUN apk add --no-cache ca-certificates tzdata netcat-openbsd

WORKDIR /app

# Copy the binary and other necessary files
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
COPY --from=builder /app/cert ./cert

# Create script to check and start the application
RUN echo '#!/bin/sh\n\
echo "Checking SSL certificates..."\n\
ls -la /app/cert\n\
echo "Testing ports..."\n\
nc -zv localhost 8443 || echo "Warning: Port 8443 not responding"\n\
echo "Starting server..."\n\
/app/main' > /app/start.sh && chmod +x /app/start.sh

# Set proper permissions
RUN chmod -R 755 /app/static && \
    chmod 600 /app/cert/server.key && \
    chmod 644 /app/cert/server.crt

# Expose ports
EXPOSE 8080 8443

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 8443 || exit 1

# Start the application
CMD ["/app/start.sh"]