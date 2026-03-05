# ============================================================
# Multi-stage Dockerfile for OmniWallet User Service
# Stage 1 (builder): compiles the Go binary
# Stage 2 (runner):  runs the minimal production image
# ============================================================

# --- Stage 1: Build ---
FROM golang:1.22-alpine AS builder

# Install build dependencies (needed for CGO-free builds on Alpine).
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy module files first for layer caching — dependencies are re-downloaded
# only when go.mod / go.sum change, not on every source change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build a statically linked binary with no debug symbols for a minimal image size.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /app/user-service ./cmd/api

# --- Stage 2: Run ---
FROM scratch

# Import CA certificates and timezone data from the builder stage.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the compiled binary.
COPY --from=builder /app/user-service /user-service

# Expose the application port (must match APP_PORT in .env).
EXPOSE 8081

# Run as non-root for security hardening.
USER 65534:65534

ENTRYPOINT ["/user-service"]
