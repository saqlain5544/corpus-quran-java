#!/bin/bash
# start/restart the quranreader server.
# Usage: ./run.sh [--rebuild] [--port N]
#   --rebuild  force a full rebuild even if Go thinks nothing changed
#   --port N   port to listen on (default: 8080)
set -euo pipefail

cd "$(dirname "$0")"
ROOT=$(pwd)

# ── args ────────────────────────────────────────────────────────────────────
REBUILD=false
PORT="${PORT:-8080}"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --rebuild) REBUILD=true; shift ;;
    --port)    PORT="$2"; shift 2 ;;
    *)         echo "unknown flag: $1"; exit 1 ;;
  esac
done

# ── stop old server ──────────────────────────────────────────────────────────
if lsof -ti:"$PORT" >/dev/null 2>&1; then
  echo "Port $PORT in use — stopping old server…"
  # SIGTERM first so the server can shut down cleanly.
  kill $(lsof -ti:"$PORT") 2>/dev/null || true
  sleep 1
  # If still alive, SIGKILL.
  if lsof -ti:"$PORT" >/dev/null 2>&1; then
    kill -9 $(lsof -ti:"$PORT") 2>/dev/null || true
    sleep 0.5
  fi
  echo "Port $PORT is now free."
fi

# ── build ────────────────────────────────────────────────────────────────────
echo "Building…"
if $REBUILD; then
  # -a forces rebuild of all packages so static file changes are always embedded.
  go build -a -o ./bin/quranreader ./backend
else
  go build -o ./bin/quranreader ./backend
fi

# Touch the binary so dependent tools (e.g. hot-reload watchers) see a fresh mtime.
touch ./bin/quranreader

# ── start ───────────────────────────────────────────────────────────────────
LOG="/tmp/quranreader-$PORT.log"
echo "Starting on :$PORT  (log: $LOG)"
./bin/quranreader \
  -addr ":$PORT" \
  -db ./data/new/detailed-quran.db \
  -translations ./data/quran/translations \
  &>"$LOG" &

SERVER_PID=$!
echo "$SERVER_PID" > /tmp/quranreader-$PORT.pid
echo "Server PID: $SERVER_PID"

# Wait for it to open the port.
for i in $(seq 1 10); do
  if lsof -ti:"$PORT" >/dev/null 2>&1; then
    echo "Server is up on :$PORT."
    exit 0
  fi
  sleep 0.3
done

echo "ERROR: server failed to start — see $LOG"
cat "$LOG"
exit 1
