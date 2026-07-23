# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy source
COPY src/ .

# Build
RUN go build -o gateway ./cmd/gateway

# Runtime stage
FROM alpine:3.18

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/gateway .

# Copy config
COPY src/config.yaml .

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./gateway"]
