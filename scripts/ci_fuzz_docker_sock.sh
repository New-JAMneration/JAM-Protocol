#!/usr/bin/env bash
# Run conformance + fuzzy test_folder against a pre-built fuzz target Docker image.
# Used by .github/workflows/fuzz-validate.yml (job fuzz-sock). Spec: READMERef/VALIDATE_FUZZ.md.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

TARGET_IMAGE="${TARGET_IMAGE:-new-jamneration-target:ci}"
FUZZ_TRACES="${FUZZ_TRACES:-pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces}"
FUZZ_FUZZY_TRACES="${FUZZ_FUZZY_TRACES:-pkg/test_data/jam-test-vectors/traces/fuzzy}"
VALIDATE_FUZZ_OUTPUT="${VALIDATE_FUZZ_OUTPUT:-output.txt}"
VALIDATE_FUZZ_OUTPUT_FUZZY="${VALIDATE_FUZZ_OUTPUT_FUZZY:-output_fuzzy.txt}"
TARGET_CONTAINER="${TARGET_CONTAINER:-fuzz-target-ci}"
HOST_DATA="${JAM_FUZZ_HOST_DIR:-.ci/fuzz_docker_run}"
TARGET_STARTUP_SEC="${TARGET_STARTUP_SEC:-30}"

log() { printf '[ci_fuzz_docker_sock] %s\n' "$*"; }
die() {
	printf '[ci_fuzz_docker_sock] ERROR: %s\n' "$*" >&2
	if docker ps -a --format '{{.Names}}' 2>/dev/null | grep -qx "${TARGET_CONTAINER}"; then
		log "last 80 lines of ${TARGET_CONTAINER} logs:"
		docker logs --tail 80 "${TARGET_CONTAINER}" 2>&1 || true
	fi
	exit 1
}

assert_sock_output() {
	local outfile="$1"
	local trace_path="$2"
	local label="$3"
	[[ -s "$outfile" ]] || die "${label}: empty output ${outfile}"
	if grep -qE 'handshake failed|Handshake failed' "$outfile"; then
		die "${label}: fuzz handshake failed (see ${outfile})"
	fi
	local expected json_count passed_count
	expected="$(find "$trace_path" -type f -name '*.json' 2>/dev/null | wc -l | tr -d ' ')"
	json_count=$(grep -cE 'Found [0-9]+ JSON files' "$outfile" 2>/dev/null || true)
	json_count=${json_count:-0}
	passed_count=$(grep -cE 'PASSED:' "$outfile" 2>/dev/null || true)
	passed_count=${passed_count:-0}
	[[ "$expected" -gt 0 ]] || die "${label}: no *.json under ${trace_path}"
	[[ "$json_count" -gt 0 ]] || die "${label}: missing 'Found N JSON files' in ${outfile}"
	[[ "$passed_count" -eq "$expected" ]] \
		|| die "${label}: expected ${expected} PASSED lines, got ${passed_count} (see ${outfile})"
	if grep -qE 'FAILED!!?:|FAILED:' "$outfile"; then
		grep -E 'FAILED!!?:|FAILED:' "$outfile" | head -20 >&2 || true
		die "${label}: failures in ${outfile}"
	fi
	log "${label} OK: ${passed_count}/${expected} traces, no FAILED in ${outfile}"
}

wait_for_socket() {
	local sock="$1"
	local deadline=$((SECONDS + TARGET_STARTUP_SEC))
	while [[ ! -S "$sock" ]]; do
		[[ "$SECONDS" -lt "$deadline" ]] || die "timeout waiting for socket: $sock"
		sleep 0.2
	done
}

run_test_folder() {
	local trace_path="$1"
	local outfile="$2"
	[[ -d "$trace_path" ]] || die "trace path not found: $trace_path"
	local abs_traces
	abs_traces="$(cd "$trace_path" && pwd)"
	rm -f "$outfile"
	log "test_folder ${trace_path} -> ${outfile}"
	docker run --rm \
		-u "$(id -u):$(id -g)" \
		-v "${HOST_DATA}:/tmp/jam_fuzz" \
		-v "${abs_traces}:/traces:ro" \
		"${TARGET_IMAGE}" \
		test_folder /tmp/jam_fuzz/fuzz.sock /traces >"$outfile" 2>&1 || true
}

docker image inspect "${TARGET_IMAGE}" >/dev/null 2>&1 \
	|| die "image not loaded: ${TARGET_IMAGE} (run build-image job or docker load)"

mkdir -p "${HOST_DATA}"
chmod a+rwx "${HOST_DATA}" 2>/dev/null || true
rm -f "${HOST_DATA}/fuzz.sock"

docker rm -f "${TARGET_CONTAINER}" >/dev/null 2>&1 || true
log "starting target container ${TARGET_CONTAINER} (${TARGET_IMAGE})"
docker run -d --name "${TARGET_CONTAINER}" \
	--init \
	-u "$(id -u):$(id -g)" \
	--cap-add IPC_LOCK \
	-e JAM_FUZZ=1 \
	-e JAM_FUZZ_SPEC=tiny \
	-e JAM_FUZZ_DATA_PATH=/tmp/jam_fuzz \
	-e JAM_FUZZ_SOCK_PATH=/tmp/jam_fuzz/fuzz.sock \
	-e JAM_FUZZ_LOG_LEVEL=ERROR \
	-v "${HOST_DATA}:/tmp/jam_fuzz" \
	"${TARGET_IMAGE}" >/dev/null

cleanup() {
	docker rm -f "${TARGET_CONTAINER}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

wait_for_socket "${HOST_DATA}/fuzz.sock"

run_test_folder "${FUZZ_TRACES}" "${VALIDATE_FUZZ_OUTPUT}"
assert_sock_output "${VALIDATE_FUZZ_OUTPUT}" "${FUZZ_TRACES}" "conformance traces"

run_test_folder "${FUZZ_FUZZY_TRACES}" "${VALIDATE_FUZZ_OUTPUT_FUZZY}"
assert_sock_output "${VALIDATE_FUZZ_OUTPUT_FUZZY}" "${FUZZ_FUZZY_TRACES}" "fuzzy traces"

log "done"
