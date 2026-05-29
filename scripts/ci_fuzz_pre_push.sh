#!/usr/bin/env bash
# Local smoke before push — mirrors CI jobs that do not need GitHub Actions.
# Does not run jam-testing (slow; needs npm). Requires submodules already init locally.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

log() { printf '[ci_fuzz_pre_push] %s\n' "$*"; }
die() { printf '[ci_fuzz_pre_push] ERROR: %s\n' "$*" >&2; exit 1; }

command -v docker >/dev/null || die "docker required"
command -v go >/dev/null || die "go required"

for sub in pkg/Rust-VRF pkg/test_data/jam-test-vectors pkg/test_data/jam-conformance; do
  [[ -d "$sub/.git" || -f "$sub/.git" ]] || die "submodule not initialized: $sub (run: git submodule update --init $sub)"
done

[[ -f pkg/Rust-VRF/vrf-func-ffi/target/release/libbandersnatch_vrfs_ffi.so ]] || {
  log "building VRF FFI (same as CI test-vectors job)..."
  (cd pkg/Rust-VRF/vrf-func-ffi && cargo build --release)
}

export TARGET_IMAGE="${TARGET_IMAGE:-new-jamneration-target:ci}"

if ! docker image inspect "$TARGET_IMAGE" >/dev/null 2>&1; then
  log "building Docker image $TARGET_IMAGE (same as CI build-image job)..."
  make fuzz-docker-build JAM_FUZZ_IMAGE="$TARGET_IMAGE"
fi

log "=== validate-fuzz-ci (steps 1,2,3,fuzzy — like test-vectors + test-trace + fuzz-sock) ==="
make validate-fuzz-ci

log "=== ci_fuzz_docker_sock (Docker target — CI fuzz-sock) ==="
./scripts/ci_fuzz_docker_sock.sh

log "pre-push smoke passed (optional: ./scripts/ci_jam_testing_full.sh for jam-testing job)"
