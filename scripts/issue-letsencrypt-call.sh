#!/usr/bin/env bash
# Let's Encrypt (webroot). Nginx в Docker должен быть запущен.
# По умолчанию: chicloon.ru + www (DNS уже на VPS).
# Добавить call.chicloon.ru: A-запись на сервер, затем:
#   sudo LETSENCRYPT_EXPAND=1 LETSENCRYPT_EXTRA_DOMAINS="call.chicloon.ru" ./scripts/issue-letsencrypt-call.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
WEBROOT="$ROOT/certbot/www"
SSL_DIR="$ROOT/ssl"
mkdir -p "$WEBROOT" "$SSL_DIR"

EXTRA=()
if [[ "${LETSENCRYPT_STAGING:-}" == "1" ]]; then EXTRA+=(--staging); fi

BASE_DOMAINS=(chicloon.ru www.chicloon.ru)
if [[ -n "${LETSENCRYPT_EXTRA_DOMAINS:-}" ]]; then
  read -r -a MORE <<< "${LETSENCRYPT_EXTRA_DOMAINS}"
  BASE_DOMAINS+=("${MORE[@]}")
fi

CERT_NAME="${LETSENCRYPT_CERT_NAME:-chicloon.ru}"
LIVE="/etc/letsencrypt/live/$CERT_NAME"

DOMAIN_ARGS=()
for d in "${BASE_DOMAINS[@]}"; do DOMAIN_ARGS+=(-d "$d"); done

if [[ "${LETSENCRYPT_EXPAND:-}" == "1" ]] && sudo test -d "$LIVE"; then
  EXTRA+=(--expand)
fi

sudo certbot certonly \
  --webroot -w "$WEBROOT" \
  "${DOMAIN_ARGS[@]}" \
  --cert-name "$CERT_NAME" \
  --non-interactive --agree-tos \
  --email "${LETSENCRYPT_EMAIL:-admin@chicloon.ru}" \
  "${EXTRA[@]}"

sudo test -f "$LIVE/fullchain.pem"
sudo install -m 644 "$LIVE/fullchain.pem" "$SSL_DIR/fullchain.pem"
sudo install -m 600 "$LIVE/privkey.pem" "$SSL_DIR/privkey.pem"

cd "$ROOT"
sudo docker compose exec -T nginx nginx -s reload
echo "OK: сертификат $CERT_NAME → $SSL_DIR, nginx reload."
