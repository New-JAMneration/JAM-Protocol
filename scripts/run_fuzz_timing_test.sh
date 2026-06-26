#!/usr/bin/env bash
# Run fuzz conformance with STF timing (server + test_folder client).
# Usage: bash scripts/run_fuzz_timing_test.sh [folder-path]
#        bash scripts/run_fuzz_timing_test.sh --prepare-only
set -euo pipefail

cd "$(dirname "$0")/.."

PREPARE_ONLY=0
if [[ "${1:-}" == "--prepare-only" ]]; then
	PREPARE_ONLY=1
	shift
fi

JAM_FUZZ_HOST_DIR="${JAM_FUZZ_HOST_DIR:-.jam_fuzz_docker_run}"
SOCK="${JAM_FUZZ_HOST_DIR}/fuzz.sock"
PVM_BACKEND="${PVM_BACKEND:-interpreter}"
FUZZ_TRACES_SRC="${FUZZ_TRACES_SRC:-pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces}"
FUZZ_TRACES="${FUZZ_TRACES:-target/fuzz}"

resolve_path() {
	if command -v readlink >/dev/null 2>&1; then
		readlink -f "$1" 2>/dev/null || echo "$1"
	else
		echo "$1"
	fi
}

has_json_files() {
	local dir="$1"
	[[ -n "$(find -L "$dir" -name '*.json' -print -quit 2>/dev/null)" ]]
}

prepare_fuzz_traces() {
	if [[ ! -d "${FUZZ_TRACES_SRC}" ]]; then
		echo "ERROR: fuzz trace source not found: ${FUZZ_TRACES_SRC}" >&2
		echo "Initialize test data with:" >&2
		echo "  git submodule update --init pkg/test_data/jam-conformance" >&2
		exit 1
	fi

	mkdir -p target

	if [[ -L "${FUZZ_TRACES}" ]]; then
		:
	elif [[ -e "${FUZZ_TRACES}" ]]; then
		if has_json_files "${FUZZ_TRACES}"; then
			:
		else
			echo "Replacing empty ${FUZZ_TRACES} with symlink to ${FUZZ_TRACES_SRC}..."
			rm -rf "${FUZZ_TRACES}"
			ln -sfn "../${FUZZ_TRACES_SRC}" "${FUZZ_TRACES}"
		fi
	else
		echo "Linking ${FUZZ_TRACES} -> ${FUZZ_TRACES_SRC}..."
		ln -sfn "../${FUZZ_TRACES_SRC}" "${FUZZ_TRACES}"
	fi

	if ! has_json_files "${FUZZ_TRACES}"; then
		echo "ERROR: no JSON files under ${FUZZ_TRACES}" >&2
		echo "Expected traces in ${FUZZ_TRACES_SRC}" >&2
		exit 1
	fi
}

FUZZ_FOLDER="${1:-${FUZZ_FOLDER:-${FUZZ_TRACES}}}"
prepare_fuzz_traces

if [[ "${FUZZ_FOLDER}" != /* ]]; then
	FUZZ_FOLDER="$(resolve_path "${FUZZ_FOLDER}")"
fi

if [[ ! -d "${FUZZ_FOLDER}" ]]; then
	echo "ERROR: fuzz folder not found: ${FUZZ_FOLDER}" >&2
	exit 1
fi

if ! has_json_files "${FUZZ_FOLDER}"; then
	echo "ERROR: no JSON files found in ${FUZZ_FOLDER}" >&2
	echo "Pick a trace subfolder, e.g.:" >&2
	echo "  bash scripts/run_fuzz_timing_test.sh target/fuzz/1766241814" >&2
	exit 1
fi

if [[ "${PREPARE_ONLY}" -eq 1 ]]; then
	echo "Fuzz traces ready at ${FUZZ_FOLDER}"
	exit 0
fi

mkdir -p "${JAM_FUZZ_HOST_DIR}"
rm -f "${SOCK}"

SERVER_BIN="${JAM_FUZZ_HOST_DIR}/fuzz-target"

start_server() {
	rm -f "${SOCK}"
	echo "Building fuzz target..."
	go build -o "${SERVER_BIN}" ./cmd/fuzz/
	echo "Starting fuzz target server (TIMING=1, PVM_BACKEND=${PVM_BACKEND})..."
	# Run the compiled binary, NOT `go run`: `go run` does not forward SIGTERM to
	# the child, so `kill` would terminate the wrapper without the server ever
	# running its graceful-shutdown defers — and the TIMING summary would never print.
	TIMING=1 JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny JAM_PVM_BACKEND="${PVM_BACKEND}" \
		JAM_FUZZ_DATA_PATH="${JAM_FUZZ_HOST_DIR}/" JAM_FUZZ_SOCK_PATH="${SOCK}" \
		"${SERVER_BIN}" 2>&1 &
	SERVER_PID=$!

	for _ in $(seq 1 120); do
		if [[ -S "${SOCK}" ]]; then
			sleep 1
			return 0
		fi
		if ! kill -0 "${SERVER_PID}" 2>/dev/null; then
			wait "${SERVER_PID}" || true
			echo "ERROR: fuzz server exited before socket ready"
			exit 1
		fi
		sleep 1
	done

	echo "ERROR: timeout waiting for ${SOCK}"
	kill "${SERVER_PID}" 2>/dev/null || true
	exit 1
}

stop_server() {
	if [[ -n "${SERVER_PID:-}" ]] && kill -0 "${SERVER_PID}" 2>/dev/null; then
		kill "${SERVER_PID}" 2>/dev/null || true
		wait "${SERVER_PID}" 2>/dev/null || true
	fi
	rm -f "${SOCK}"
}

trap stop_server EXIT INT TERM

start_server
echo "Running test_folder on ${FUZZ_FOLDER}..."
# Reuse the already-built binary for the client too (faster, same code).
"${SERVER_BIN}" test_folder "${SOCK}" "${FUZZ_FOLDER}"
