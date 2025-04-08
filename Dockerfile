# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build both servers
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/grpc-server ./cmd/grpc-server

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy built binaries from builder
COPY --from=builder /app/server /app/server
COPY --from=builder /app/grpc-server /app/grpc-server
COPY --from=builder /app/scripts/entrypoint.sh /app/entrypoint.sh

# Make entrypoint script executable
RUN chmod +x /app/entrypoint.sh

# Use non-root user
USER appuser

# Expose ports for both REST API and gRPC
EXPOSE 8080 50051

# Set the entrypoint script
ENTRYPOINT ["/app/entrypoint.sh"] 