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

# Create necessary directories
RUN mkdir -p cert && \
    chmod -R 755 /app

# Add debugging utilities
RUN apk add --no-cache curl tzdata ca-certificates

# Expose both HTTP and HTTPS ports
EXPOSE 8083 8446

# Command to run the application
CMD ["./main"]
