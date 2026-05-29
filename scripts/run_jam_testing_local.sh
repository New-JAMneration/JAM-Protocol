#!/usr/bin/env bash
# Optional local jam-testing (minifuzz gate). See READMERef/VALIDATE_FUZZ.md.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
JAM_TESTING_DIR="${JAM_TESTING_DIR:-/tmp/jam-testing}"
TARGET_IMAGE="${TARGET_IMAGE:-new-jamneration-target:latest}"
JAM_FUZZ_SPEC="${JAM_FUZZ_SPEC:-tiny}"

log() { printf '[jam-testing-local] %s\n' "$*"; }
die() { printf '[jam-testing-local] ERROR: %s\n' "$*" >&2; exit 1; }

command -v docker >/dev/null || die "docker required"
command -v npm >/dev/null || die "npm required (Node 20+)"

if [[ ! -d "$JAM_TESTING_DIR/.git" ]]; then
	log "cloning FluffyLabs/jam-testing -> $JAM_TESTING_DIR"
	git clone --depth 1 https://github.com/FluffyLabs/jam-testing.git "$JAM_TESTING_DIR"
fi

if ! docker image inspect "$TARGET_IMAGE" >/dev/null 2>&1; then
	log "building target image (make fuzz-docker-build)..."
	(cd "$REPO_ROOT" && make fuzz-docker-build JAM_FUZZ_IMAGE="$TARGET_IMAGE")
fi

log "npm install in jam-testing..."
(cd "$JAM_TESTING_DIR" && npm install)

export JAM_FUZZ_SPEC
export TARGET_NAME="${TARGET_NAME:-new-jamneration}"
export TARGET_IMAGE
export TARGET_CMD="${TARGET_CMD:-{TARGET_SOCK}}"
export TARGET_MEMORY="${TARGET_MEMORY:-2g}"
export TARGET_ENV="${TARGET_ENV:-USE_MINI_REDIS=true}"

log "running minifuzz fallback suite (smoke; set JAM_TESTING_SUITE=all for full minifuzz)..."
SUITE="${JAM_TESTING_SUITE:-fallback}"
(cd "$JAM_TESTING_DIR" && npx tsx --test "tests/minifuzz/${SUITE}.test.ts")

log "jam-testing local smoke passed (suite=${SUITE})"
