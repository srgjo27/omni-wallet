.PHONY: dev build stop-docker start-docker

# ---------------------------------------------------------------------------
# dev — run wallet-service locally, pointing at Docker-hosted dependencies.
#
# Prerequisites (run once if not started):
#   make start-docker
#
# What this does:
#   1. Loads .env.local (uses localhost:* to reach Docker-exposed ports)
#   2. Starts the service with `go run` so logs are printed directly here
# ---------------------------------------------------------------------------
dev:
	@echo "[wallet-service] Starting locally with .env.local ..."
	@set -a && . ./.env.local && set +a && go run ./cmd/api

# Build a local binary (optional, faster restart than go run)
build:
	@echo "[wallet-service] Building binary ..."
	go build -o bin/wallet-service ./cmd/api

# Run the pre-built binary with local env
run:
	@set -a && . ./.env.local && set +a && ./bin/wallet-service

# ---------------------------------------------------------------------------
# Docker helpers — manage the wallet-service container without touching other
# services. Use these to switch between local and Docker modes.
# ---------------------------------------------------------------------------

# Stop only wallet-service in Docker (free port 8082 for local process)
stop-docker:
	@echo "[docker] Stopping wallet-service container ..."
	docker compose -f ../docker-compose.yml stop wallet-service

# Start wallet-service back in Docker
start-docker:
	@echo "[docker] Starting wallet-service container ..."
	docker compose -f ../docker-compose.yml up -d wallet-service
