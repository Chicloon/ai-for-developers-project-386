#!/usr/bin/env sh
# Самоподписанный сертификат для локального docker compose (nginx слушает 443).
set -e
cd "$(dirname "$0")/.."
mkdir -p ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout ssl/privkey.pem \
  -out ssl/fullchain.pem \
  -subj "/CN=call.chicloon.ru" \
  -addext "subjectAltName=DNS:localhost,DNS:call.chicloon.ru,IP:127.0.0.1"
chmod 600 ssl/privkey.pem
echo "OK: ssl/fullchain.pem и ssl/privkey.pem"
