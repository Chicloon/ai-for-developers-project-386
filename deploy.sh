#!/bin/bash
set -e

LOG_FILE="deploy-$(date '+%Y-%m-%d').log"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

log "→ Pulling images..."
docker compose pull

log "→ Starting containers..."
docker compose up -d

log "→ Cleaning up Docker..."
docker builder prune -f
docker image prune -a -f
docker container prune -f

log "Done!"