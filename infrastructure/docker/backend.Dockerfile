# Stage 1: Build the Go application
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git ca-certificates tzdata

# Copy Go module files and download dependencies
# Copy backend go.mod and go.sum to /app/backend/
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /app/backend
RUN go mod download
RUN go mod tidy

# Copy source code
# Copy the rest of the backend code into /app/backend/
COPY backend/ ./

# Lint and Format Check
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
RUN /go/bin/golangci-lint run ./...
RUN test -z $(gofmt -l . | tee /dev/stderr) || (echo "Go files are not formatted. Please run gofmt." && exit 1)

# Build the application with optimizations
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=docker" \
    -o /app/main ./backend/cmd/api

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add tzdata ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/main /app/main

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
CMD ["/app/main"]