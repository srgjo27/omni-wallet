# =========================================================
# Stage 1 — Build
# =========================================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o /app/api-gateway ./cmd/api

# =========================================================
# Stage 2 — Minimal runtime image
# =========================================================
FROM scratch

# CA certificates for outbound HTTPS calls to upstream services.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/api-gateway /api-gateway

# Run as non-root UID 65534 (nobody).
USER 65534

EXPOSE 8080

ENTRYPOINT ["/api-gateway"]
