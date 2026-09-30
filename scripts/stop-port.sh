#!/usr/bin/env bash
# Verilen portu dinleyen süreçleri durdurur.
# Kullanım: scripts/stop-port.sh [port]   (varsayılan: 8080)
set -euo pipefail

port="${1:-8080}"

if ! [[ "$port" =~ ^[0-9]+$ ]] || [ "$port" -lt 1 ] || [ "$port" -gt 65535 ]; then
  echo "geçersiz port: $port" >&2
  exit 2
fi

if ! command -v lsof >/dev/null 2>&1; then
  echo "lsof bulunamadı" >&2
  exit 1
fi

pids="$(lsof -nP -t -iTCP:"$port" -sTCP:LISTEN || true)"
if [ -z "$pids" ]; then
  echo "port $port boş"
  exit 0
fi

echo "port $port: durduruluyor -> $(echo "$pids" | tr '\n' ' ')"
# shellcheck disable=SC2086
kill $pids 2>/dev/null || true

# Süreçlere düzgün kapanma için kısa süre tanı, sonra zorla
for _ in 1 2 3 4 5 6 7 8 9 10; do
  sleep 0.3
  pids="$(lsof -nP -t -iTCP:"$port" -sTCP:LISTEN || true)"
  [ -z "$pids" ] && { echo "port $port boşaltıldı"; exit 0; }
done

echo "süreç kapanmadı, SIGKILL gönderiliyor -> $(echo "$pids" | tr '\n' ' ')"
# shellcheck disable=SC2086
kill -9 $pids 2>/dev/null || true
echo "port $port boşaltıldı"
