# =========================================================
# Stage 1 — Build
# =========================================================
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
    -o /app/notification-worker ./cmd/worker

# =========================================================
# Stage 2 — Minimal runtime image
# =========================================================
FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /app/notification-worker /notification-worker

# Run as non-root UID 65534 (nobody).
USER 65534

ENTRYPOINT ["/notification-worker"]
