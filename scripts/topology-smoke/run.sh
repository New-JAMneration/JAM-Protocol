#!/usr/bin/env bash
# Run 3 Go validator nodes with topology smoke config and collect results.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SMOKE_DIR="${SMOKE_DIR:-/tmp/jam-topology-smoke}"
BIN="${BIN:-/tmp/jam-node-test}"
BASE_PORT=10101
WAIT_SECS=20

"$ROOT/scripts/topology-smoke/setup.sh"

echo "Building node..."
(cd "$ROOT" && go build -o "$BIN" ./cmd/node)

PID_FILE="$SMOKE_DIR/pids.txt"
LOG_DIR="$SMOKE_DIR/logs"
rm -f "$PID_FILE"
mkdir -p "$LOG_DIR"

cleanup() {
  echo "Stopping smoke nodes..."
  if [[ -f "$PID_FILE" ]]; then
    while read -r pid; do
      kill "$pid" 2>/dev/null || true
    done < "$PID_FILE"
  fi
}
trap cleanup EXIT INT TERM

for i in 0 1 2; do
  NODE_DIR="$SMOKE_DIR/node$i"
  PORT=$((BASE_PORT + i))
  LOG="$LOG_DIR/node$i.log"
  (
    cd "$NODE_DIR"
    export USE_MINI_REDIS=true
    export JAM_TOPOLOGY_SMOKE_CONFIG="$SMOKE_DIR/smoke_config.json"
    exec "$BIN" \
      --role validator \
      --listen-addr "127.0.0.1:$PORT" \
      --chain "$NODE_DIR/dev.chainspec.json" \
      --config "$ROOT/example.json"
  ) >"$LOG" 2>&1 &
  echo $! >> "$PID_FILE"
  echo "Started node$i pid=$! port=$PORT log=$LOG"
done

echo "Waiting ${WAIT_SECS}s for reconcile + dials..."
sleep "$WAIT_SECS"

echo ""
echo "=== Smoke results ==="
PASS=0
FAIL=0

for i in 0 1 2; do
  LOG="$LOG_DIR/node$i.log"
  echo "--- node$i ($LOG) ---"
  if grep -q "topology smoke: seeded 3 validators" "$LOG"; then
    echo "  [OK] smoke config applied"
    PASS=$((PASS + 1))
  else
    echo "  [MISS] smoke config not applied"
    FAIL=$((FAIL + 1))
  fi
  if grep -q "safrole: connectivity applied" "$LOG"; then
    echo "  [OK] ConnectivityApplied published"
    PASS=$((PASS + 1))
  else
    echo "  [MISS] no ConnectivityApplied log"
    FAIL=$((FAIL + 1))
  fi
  if grep -qE "TLS connection established|on peer updated: peer=127.0.0.1:1010" "$LOG"; then
    echo "  [OK] QUIC peer connectivity"
    grep -E "TLS connection established|on peer updated: peer=127.0.0.1:1010" "$LOG" | tail -5 | sed 's/^/    /'
    PASS=$((PASS + 1))
  else
    echo "  [MISS] no peer connectivity logs"
    FAIL=$((FAIL + 1))
  fi
  if grep -qi "failed to connect\|dial failed" "$LOG"; then
    echo "  [WARN] dial errors present:"
    grep -iE "failed to connect|dial failed" "$LOG" | tail -3 | sed 's/^/    /'
  fi
  echo ""
done

echo "Checks passed: $PASS, issues: $FAIL"
if [[ "$FAIL" -gt 0 ]]; then
  echo "SMOKE: PARTIAL (see logs under $LOG_DIR)"
  exit 1
fi
echo "SMOKE: PASS"
exit 0
