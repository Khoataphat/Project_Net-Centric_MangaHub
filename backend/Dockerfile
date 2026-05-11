# Stage 1: Build
FROM golang:1.25-alpine AS builder

# Install build dependencies (needed for SQLite if CGO is enabled, but glebarez/go-sqlite is CGO-free)
# However, it's good practice to have build-base for some Go packages.
RUN apk add --no-cache build-base

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# Use CGO_ENABLED=0 to ensure portability if using glebarez/go-sqlite
RUN CGO_ENABLED=0 GOOS=linux go build -o mangahub-server ./cmd/server/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o mangahub-seeder ./cmd/seed/main.go

# Stage 2: Runtime
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite

# Copy binaries from builder
COPY --from=builder /app/mangahub-server .
COPY --from=builder /app/mangahub-seeder .

# Copy entrypoint script
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

# Create data directory for SQLite persistent storage
RUN mkdir -p /app/data

# Expose ports as defined in PRD
# 8080: HTTP
# 9090: TCP
# 9999: UDP Notifier
# 8888: UDP Bridge
# 50051: gRPC
EXPOSE 8080 9090 9999/udp 8888/udp 50051

# Default Environment Variables
ENV HOST=0.0.0.0
ENV HTTP_PORT=8080
ENV TCP_PORT=9090
ENV UDP_PORT=9999
ENV BRIDGE_PORT=8888
ENV GRPC_PORT=50051
ENV DB_PATH=/app/data/mangahub.db

# Command to run
CMD ["./entrypoint.sh"]
