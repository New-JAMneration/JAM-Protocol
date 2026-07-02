#!/usr/bin/env bash
# Fuzz validation pipeline — see READMERef/VALIDATE_FUZZ.md for rationale and pass criteria.
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# --- configurable defaults (override via env) ---
JAM_FUZZ_HOST_DIR="${JAM_FUZZ_HOST_DIR:-.jam_fuzz_docker_run}"
FUZZ_SOCK="${FUZZ_SOCK:-$JAM_FUZZ_HOST_DIR/fuzz.sock}"
FUZZ_TRACES="${FUZZ_TRACES:-pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces}"
FUZZ_FUZZY_TRACES="${FUZZ_FUZZY_TRACES:-pkg/test_data/jam-test-vectors/traces/fuzzy}"
VALIDATE_FUZZ_OUTPUT="${VALIDATE_FUZZ_OUTPUT:-output.txt}"
VALIDATE_FUZZ_OUTPUT_FUZZY="${VALIDATE_FUZZ_OUTPUT_FUZZY:-output_fuzzy.txt}"
TEST_SIZE="${TEST_SIZE:-tiny}"
TEST_FORMAT="${TEST_FORMAT:-binary}"
# Steps: 1,2,3 (conformance sock), fuzzy (jam-test-vectors fuzzy sock), 4 (jam-testing hint)
VALIDATE_FUZZ_STEPS="${VALIDATE_FUZZ_STEPS:-1,2,3,fuzzy}"
FUZZ_SMOKE_TRACE_DIR="${FUZZ_SMOKE_TRACE_DIR:-}"
TARGET_STARTUP_SEC="${TARGET_STARTUP_SEC:-30}"
FUZZ_TARGET_BIN="${FUZZ_TARGET_BIN:-${JAM_FUZZ_HOST_DIR}/fuzz-target-bin}"
VALIDATE_FUZZ_TARGET_LOG="${VALIDATE_FUZZ_TARGET_LOG:-${JAM_FUZZ_HOST_DIR}/fuzz_target.log}"
# statistics tiny: the v0.7.x vectors are skipped as v0.8.0-incompatible
# (testdata.v080IncompatibleModes, #1021), so the mode yields 0 vectors.
# Restore the real expectations when official v0.8.0 vectors land (#1012).
STATISTICS_EXPECT_PASSED="${STATISTICS_EXPECT_PASSED:-0}"
STATISTICS_EXPECT_FAILED="${STATISTICS_EXPECT_FAILED:-0}"

log() { printf '[validate_fuzz] %s\n' "$*"; }
die() { printf '[validate_fuzz] ERROR: %s\n' "$*" >&2; exit 1; }

step_enabled() {
	local n="$1"
	[[ ",${VALIDATE_FUZZ_STEPS}," == *",${n},"* ]]
}

parse_test_totals() {
	local logfile="$1"
	grep -E 'Total: [0-9]+, Passed: [0-9]+, Failed: [0-9]+' "$logfile" | tail -1 | sed -E 's/.*Total: ([0-9]+), Passed: ([0-9]+), Failed: ([0-9]+).*/\1 \2 \3/'
}

run_jam_test_vectors_mode() {
	local mode="$1"
	local log
	log="$(mktemp)"
	log "step 1: mode=${mode}"
	if ! go run ./cmd/node test \
		--mode "$mode" --size "$TEST_SIZE" --type jam-test-vectors --format "$TEST_FORMAT" \
		>"$log" 2>&1; then
		cat "$log" >&2
		rm -f "$log"
		return 1
	fi
	local totals shell_total shell_passed shell_failed
	if ! totals="$(parse_test_totals "$log")"; then
		cat "$log" >&2
		rm -f "$log"
		die "mode ${mode}: could not parse Total/Passed/Failed from output"
	fi
	read -r shell_total shell_passed shell_failed <<<"$totals"
	rm -f "$log"
	log "  ${mode}: total=${shell_total} passed=${shell_passed} failed=${shell_failed}"
	if [[ "$mode" == "statistics" ]]; then
		[[ "$shell_passed" -eq "$STATISTICS_EXPECT_PASSED" && "$shell_failed" -eq "$STATISTICS_EXPECT_FAILED" ]] \
			|| die "statistics: expected Passed=${STATISTICS_EXPECT_PASSED} Failed=${STATISTICS_EXPECT_FAILED} (1/3 pass), got ${shell_passed}/${shell_total}"
	else
		[[ "$shell_failed" -eq 0 ]] || die "mode ${mode}: expected 0 failures, got ${shell_failed}"
	fi
}

step1_jam_test_vectors() {
	log "=== Step 1: jam-test-vectors (statistics 1/3 pass, others 100%) ==="
	local modes=(
		safrole assurances preimages disputes history
		accumulate authorizations statistics reports
	)
	local m
	for m in "${modes[@]}"; do
		run_jam_test_vectors_mode "$m"
	done
}

run_trace_mode() {
	local mode="$1"
	local log
	log="$(mktemp)"
	log "step 2: trace mode=${mode}"
	if ! go run ./cmd/node test --type trace --mode "$mode" >"$log" 2>&1; then
		cat "$log" >&2
		rm -f "$log"
		return 1
	fi
	local totals shell_total shell_passed shell_failed
	if ! totals="$(parse_test_totals "$log")"; then
		cat "$log" >&2
		rm -f "$log"
		die "trace ${mode}: could not parse Total/Passed/Failed"
	fi
	read -r shell_total shell_passed shell_failed <<<"$totals"
	rm -f "$log"
	log "  trace ${mode}: total=${shell_total} passed=${shell_passed} failed=${shell_failed}"
	[[ "$shell_failed" -eq 0 ]] || die "trace ${mode}: expected 0 failures"
}

step2_jam_test_vectors_trace() {
	log "=== Step 2: jam-test-vectors trace ==="
	local modes=(
		fallback safrole preimages_light preimages
		storage_light storage fuzzy_light
	)
	local m
	for m in "${modes[@]}"; do
		run_trace_mode "$m"
	done
}

wait_for_socket() {
	local sock="$1"
	local deadline=$((SECONDS + TARGET_STARTUP_SEC))
	while [[ ! -S "$sock" ]]; do
		[[ "$SECONDS" -lt "$deadline" ]] || die "timeout waiting for socket: $sock"
		sleep 0.2
	done
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
		if [[ -f "${VALIDATE_FUZZ_TARGET_LOG:-}" ]]; then
			log "last 30 lines of target log (${VALIDATE_FUZZ_TARGET_LOG}):"
			tail -30 "$VALIDATE_FUZZ_TARGET_LOG" >&2 || true
		fi
		die "${label}: failures in ${outfile}"
	fi
	log "${label} OK: ${passed_count}/${expected} traces, no FAILED in ${outfile}"
}

run_test_folder_to_file() {
	local trace_path="$1"
	local outfile="$2"
	[[ -d "$trace_path" ]] || die "trace path not found: $trace_path"
	rm -f "$outfile"
	log "test_folder ${trace_path} -> ${outfile}"
	CLIENT_QUIET="${CLIENT_QUIET:-1}" go run ./cmd/fuzz/ test_folder \
		"$FUZZ_SOCK" "$trace_path" >"$outfile" 2>&1 || true
}

step_sock_traces() {
	if ! step_enabled 3 && ! step_enabled fuzzy; then
		return 0
	fi

	mkdir -p "$JAM_FUZZ_HOST_DIR"

	FUZZ_TARGET_PID=""
	cleanup_target() {
		if [[ -n "${FUZZ_TARGET_PID:-}" ]]; then
			kill -- -"${FUZZ_TARGET_PID}" 2>/dev/null || kill "${FUZZ_TARGET_PID}" 2>/dev/null || true
			wait "${FUZZ_TARGET_PID}" 2>/dev/null || true
		fi
		rm -f "$FUZZ_SOCK"
	}
	trap cleanup_target EXIT

	log "=== Sock tests: run-target (background) ==="
	rm -f "$FUZZ_SOCK" "$VALIDATE_FUZZ_TARGET_LOG"
	log "building fuzz target -> ${FUZZ_TARGET_BIN}"
	go build -o "$FUZZ_TARGET_BIN" ./cmd/fuzz/
	log "target log: ${VALIDATE_FUZZ_TARGET_LOG} (panics / [fuzz-server] errors)"
	setsid env JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny \
		JAM_FUZZ_DATA_PATH="${JAM_FUZZ_HOST_DIR}/" \
		JAM_FUZZ_SOCK_PATH="$FUZZ_SOCK" \
		JAM_FUZZ_LOG_LEVEL="${JAM_FUZZ_LOG_LEVEL:-ERROR}" \
		"$FUZZ_TARGET_BIN" >>"$VALIDATE_FUZZ_TARGET_LOG" 2>&1 &
	FUZZ_TARGET_PID=$!
	wait_for_socket "$FUZZ_SOCK"

	if step_enabled 3; then
		log "=== Step 3: conformance fuzz-reports traces ==="
		local trace_path="$FUZZ_TRACES"
		if [[ -n "$FUZZ_SMOKE_TRACE_DIR" ]]; then
			trace_path="${FUZZ_TRACES%/}/${FUZZ_SMOKE_TRACE_DIR}"
		fi
		run_test_folder_to_file "$trace_path" "$VALIDATE_FUZZ_OUTPUT"
		assert_sock_output "$VALIDATE_FUZZ_OUTPUT" "$trace_path" "step 3"
	fi

	if step_enabled fuzzy; then
		log "=== Step fuzzy: jam-test-vectors/traces/fuzzy ==="
		run_test_folder_to_file "$FUZZ_FUZZY_TRACES" "$VALIDATE_FUZZ_OUTPUT_FUZZY"
		assert_sock_output "$VALIDATE_FUZZ_OUTPUT_FUZZY" "$FUZZ_FUZZY_TRACES" "step fuzzy"
	fi
}

step4_jam_testing_hint() {
	log "=== Step 4: FluffyLabs jam-testing (minifuzz / picofuzz) ==="
	if [[ "${VALIDATE_FUZZ_RUN_JAM_TESTING:-0}" != "1" ]]; then
		log "skipped (see READMERef/VALIDATE_FUZZ.md § jam-testing)"
		return 0
	fi
	if [[ -x "$REPO_ROOT/scripts/run_jam_testing_local.sh" ]]; then
		"$REPO_ROOT/scripts/run_jam_testing_local.sh"
		return 0
	fi
	die "VALIDATE_FUZZ_RUN_JAM_TESTING=1 but scripts/run_jam_testing_local.sh missing"
}

main() {
	log "repo: $REPO_ROOT"
	log "steps: $VALIDATE_FUZZ_STEPS"
	step_enabled 1 && step1_jam_test_vectors
	step_enabled 2 && step2_jam_test_vectors_trace
	step_sock_traces
	step_enabled 4 && step4_jam_testing_hint
	log "=== validate_fuzz: all requested steps passed ==="
}

main "$@"
