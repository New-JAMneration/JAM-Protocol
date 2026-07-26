#!/usr/bin/env bash
# Local test-only harness: generate temp keystores and topology smoke config.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SMOKE_DIR="${SMOKE_DIR:-/tmp/jam-topology-smoke}"
CHAIN="${ROOT}/cmd/node/test_data/dev.chainspec.json"
BASE_PORT=10101

# Dev-chain validator Ed25519 seeds (JIP-5 trivial_seed 0..2) from jip5_key_derivation_test.go
declare -a PUBS=(
  "4418fb8c85bb3985394a8c2756d3643457ce614546202a2f50b093d762499ace"
  "ad93247bd01307550ec7acd757ce6fb805fcf73db364063265b30a949e90d933"
  "cab2b9ff25c2410fbe9b8a717abb298c716a03983c98ceb4def2087500b8e341"
)
declare -a SEEDS=(
  "996542becdf1e78278dc795679c825faca2e9ed2bf101bf3c4a236d3ed79cf59"
  "b81e308145d97464d2bc92d35d227a9e62241a16451af6da5053e309be4f91d7"
  "0093c8c10a88ebbc99b35b72897a26d259313ee9bad97436a437d2e43aaafa0f"
)

rm -rf "$SMOKE_DIR"
mkdir -p "$SMOKE_DIR"

python3 - "$SMOKE_DIR" "$BASE_PORT" "${PUBS[@]}" <<'PY'
import json, sys
smoke_dir, base_port, *pubs = sys.argv[1:]
validators = []
for i, pub in enumerate(pubs):
    validators.append({
        "index": i,
        "ed25519": pub,
        "host": "127.0.0.1",
        "port": int(base_port) + i,
    })
with open(f"{smoke_dir}/smoke_config.json", "w") as f:
    json.dump({"validators": validators}, f, indent=2)
print(f"Wrote {smoke_dir}/smoke_config.json")
PY

for i in 0 1 2; do
  NODE_DIR="$SMOKE_DIR/node$i"
  mkdir -p "$NODE_DIR/keystore/ed25519"
  cat > "$NODE_DIR/keystore/ed25519/keys.json" <<EOF
{
  "${PUBS[$i]}": {"private_key": "${SEEDS[$i]}"}
}
EOF
  cp "$CHAIN" "$NODE_DIR/dev.chainspec.json"
  echo "Prepared $NODE_DIR (listen 127.0.0.1:$((BASE_PORT + i)))"
done

echo "Smoke dir: $SMOKE_DIR"
