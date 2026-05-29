#!/usr/bin/env bash
# Full minifuzz + picofuzz via FluffyLabs/jam-testing (npm/tsx), against a local target image.
# See jam-testing .github/workflows/reusable-picofuzz.yml
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

TARGET_IMAGE="${TARGET_IMAGE:-new-jamneration-target:ci}"
TARGET_NAME="${TARGET_NAME:-new-jamneration}"
TARGET_CMD="${TARGET_CMD:-{TARGET_SOCK}}"
TARGET_MEMORY="${TARGET_MEMORY:-2g}"
JAM_TESTING_DIR="${JAM_TESTING_DIR:-${REPO_ROOT}/.ci/jam-testing}"
JAM_TESTING_REF="${JAM_TESTING_REF:-main}"
JAM_TESTING_REPO="${JAM_TESTING_REPO:-https://github.com/FluffyLabs/jam-testing.git}"

# Cursor prompt 4 suites + official jam-testing extras (forks / no_forks)
MINIFUZZ_SUITES="${MINIFUZZ_SUITES:-forks no_forks fallback safrole storage storage_light}"
PICOFUZZ_SUITES="${PICOFUZZ_SUITES:-fallback safrole storage storage_light}"

log() { printf '[ci_jam_testing] %s\n' "$*"; }
die() { printf '[ci_jam_testing] ERROR: %s\n' "$*" >&2; exit 1; }

command -v docker >/dev/null || die "docker required"
command -v npm >/dev/null || die "npm required (Node 20+)"

docker image inspect "${TARGET_IMAGE}" >/dev/null 2>&1 \
	|| die "image not loaded: ${TARGET_IMAGE}"

if [[ ! -d "${JAM_TESTING_DIR}/.git" ]]; then
	mkdir -p "$(dirname "${JAM_TESTING_DIR}")"
	log "cloning ${JAM_TESTING_REPO} (ref=${JAM_TESTING_REF}) -> ${JAM_TESTING_DIR}"
	git clone --depth 1 --branch "${JAM_TESTING_REF}" "${JAM_TESTING_REPO}" "${JAM_TESTING_DIR}"
fi

(
	cd "${JAM_TESTING_DIR}"
	git fetch origin "${JAM_TESTING_REF}" --depth 1 2>/dev/null || true
	git checkout "${JAM_TESTING_REF}" 2>/dev/null || git checkout "origin/${JAM_TESTING_REF}"
	log "initializing jam-testing submodules (HTTPS; skip dashboard)"
	# picofuzz-data nests jam-test-vectors via git@github.com — breaks non-interactive CI/WSL.
	# forks/no_forks need picofuzz-conformance-data; picofuzz STF needs picofuzz-stf-data only (no nested jam-test-vectors).
	git -c url.https://github.com/.insteadOf=git@github.com: \
		submodule update --init --recursive picofuzz-conformance-data
	git submodule update --init picofuzz-stf-data
)

log "npm ci + build client images in jam-testing"
(
	cd "${JAM_TESTING_DIR}"
	npm ci
	npm run build-docker -w @fluffylabs/picofuzz
	npm run build -w @fluffylabs/minifuzz
)

export TARGET_NAME TARGET_IMAGE TARGET_CMD TARGET_MEMORY
export TARGET_ENV="${TARGET_ENV:-}"

run_tsx_test() {
	local kind="$1"
	local suite="$2"
	local timeout_min="${3:-10}"
	log "=== ${kind} ${suite} (timeout ${timeout_min}m) ==="
	(
		cd "${JAM_TESTING_DIR}"
		TIMEOUT_MINUTES="${timeout_min}" npm exec tsx -- --test "tests/${kind}/${suite}.test.ts"
	)
}

for suite in ${MINIFUZZ_SUITES}; do
	timeout=10
	[[ "$suite" == "storage" || "$suite" == "storage_light" ]] && timeout=15
	run_tsx_test minifuzz "$suite" "$timeout"
done

for suite in ${PICOFUZZ_SUITES}; do
	timeout=10
	[[ "$suite" == "storage" || "$suite" == "storage_light" ]] && timeout=15
	mkdir -p "${JAM_TESTING_DIR}/picofuzz-result"
	chmod 777 "${JAM_TESTING_DIR}/picofuzz-result" 2>/dev/null || true
	run_tsx_test picofuzz "$suite" "$timeout"
done

log "all jam-testing suites passed"
