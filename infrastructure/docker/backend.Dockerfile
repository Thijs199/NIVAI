# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git ca-certificates tzdata

# Copy Go module files and download dependencies
COPY backend/go.mod ./
COPY backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git rev-parse --short HEAD)" \
    -o nivai-api ./cmd/api

# Stage 2: Create minimal runtime image
FROM alpine:3.18

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/nivai-api .

# Create a non-root user for running the application
RUN adduser -D -g '' appuser
USER appuser

# Expose API port
EXPOSE 8080

# Set environment variables
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0
ENV CONFIG_PATH=/app/config.json

# Run the application
ENTRYPOINT ["/app/nivai-api"]