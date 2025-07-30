# Build stage
FROM golang:1.24.5-alpine3.22 AS builder

# Install build dependencies and update packages
RUN apk update && apk upgrade && apk add --no-cache git make gcc musl-dev librdkafka-dev pkgconf

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o go-base-ms cmd/server/main.go

# Runtime stage
FROM alpine:3.22

# Update and install runtime dependencies
RUN apk update && apk upgrade && apk add --no-cache ca-certificates tzdata librdkafka

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create api directory
RUN mkdir -p api

# Copy binary from builder
COPY --from=builder /build/go-base-ms .

# Copy API specifications
COPY --from=builder /build/api/openapi.yaml ./api/
COPY --from=builder /build/api/openapi.json ./api/

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

# Run the application
ENTRYPOINT ["./go-base-ms"]