# ============================================
# Stage 1: Build
# ============================================
FROM golang:1.22-alpine AS builder

# Install git (required for fetching Go modules)
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /bin/api ./cmd/api

# ============================================
# Stage 2: Runtime (minimal image)
# ============================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary and assets from builder
COPY --from=builder /bin/api .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/.env.example .env.example

# Set ownership
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["./api"]
