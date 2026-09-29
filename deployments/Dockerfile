# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Cache Go modules
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Compile optimized static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=1.0.0" \
    -o /bin/ratelimiter \
    cmd/server/main.go

# Production Minimal Stage
FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /bin/ratelimiter /app/ratelimiter
COPY --from=builder /app/scripts/lua /app/scripts/lua
COPY --from=builder /app/web /app/web

# Expose gRPC (50051) and REST/Metrics (8080)
EXPOSE 8080 50051

ENV HTTP_PORT=8080 \
    GRPC_PORT=50051 \
    REDIS_HOST=redis \
    REDIS_PORT=6379 \
    ENABLE_L1_CACHE=true

HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
    CMD curl -f http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/ratelimiter"]
