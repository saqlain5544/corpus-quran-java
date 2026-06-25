#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"
PORT="${PORT:-8080}"
if lsof -ti:"$PORT" >/dev/null 2>&1; then
  echo "Port $PORT in use — killing…"
  kill -9 $(lsof -ti:"$PORT") 2>/dev/null || true
  sleep 0.5
fi
echo "Building…"
go build -o ./bin/quranreader ./backend
echo "Starting on :$PORT"
exec ./bin/quranreader -addr ":$PORT" -db ./data/new/detailed-quran.db -translations ./data/quran/translations
