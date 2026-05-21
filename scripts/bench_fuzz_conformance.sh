#!/usr/bin/env bash
# Benchmark fuzz conformance: macOS interpreter vs Docker interpreter vs Docker recompiler.
# Does NOT enable PVMtrace (uses production image / unset JAM_PVM_TRACE_DIR).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TRACES="${TRACES:-$ROOT/pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/}"
IMAGE="${IMAGE:-new-jamneration-target:latest}"
MAC_BIN="${MAC_BIN:-$ROOT/build/fuzz-bench-mac}"
SOCK_VOL="${SOCK_VOL:-jam-bench-sock}"
RESULT_DIR="${RESULT_DIR:-/tmp/jam-fuzz-bench}"
MAC_SOCK_DIR="$ROOT/.jam_fuzz_bench"

mkdir -p "$RESULT_DIR" "$MAC_SOCK_DIR"

unset JAM_PVM_TRACE_DIR JAM_PVM_TRACE_RUN_ID JAM_PVM_TRACE_STREAMS

json_count="$(find "$TRACES" -name '*.json' | wc -l | tr -d ' ')"
echo "=== Fuzz conformance benchmark ==="
echo "traces: $TRACES"
echo "json files: $json_count"
echo "docker image: $IMAGE"
echo "results: $RESULT_DIR"
echo ""

if [[ ! -f "$MAC_BIN" ]]; then
	echo "Building macOS fuzz binary..."
	go build -o "$MAC_BIN" ./cmd/fuzz
fi

docker image inspect "$IMAGE" --format 'image_arch={{.Architecture}}' 2>/dev/null || {
	echo "Docker image $IMAGE not found. Build with:"
	echo "  docker buildx build --platform=linux/amd64 ... -t $IMAGE -f docker/Dockerfile --load ."
	exit 1
}

cleanup_mac() {
	pkill -f "$MAC_SOCK_DIR/fuzz.sock" 2>/dev/null || true
	rm -f "$MAC_SOCK_DIR/fuzz.sock"
}
cleanup_docker() {
	docker rm -f jam-bench-server 2>/dev/null || true
}

run_mac_interpreter() {
	local log="$RESULT_DIR/1-macos-interpreter.log"
	local timing="$RESULT_DIR/1-macos-interpreter.time"
	cleanup_mac
	echo ">>> [1/3] macOS native interpreter"
	JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny \
		JAM_FUZZ_DATA_PATH="$MAC_SOCK_DIR/" \
		JAM_FUZZ_SOCK_PATH="$MAC_SOCK_DIR/fuzz.sock" \
		"$MAC_BIN" "$MAC_SOCK_DIR/fuzz.sock" -pvm-backend interpreter \
		> "$RESULT_DIR/1-macos-interpreter-server.log" 2>&1 &
	local pid=$!
	for _ in $(seq 1 30); do
		[[ -S "$MAC_SOCK_DIR/fuzz.sock" ]] && break
		sleep 1
	done
	if [[ ! -S "$MAC_SOCK_DIR/fuzz.sock" ]]; then
		echo "macOS server failed to start"; tail -20 "$RESULT_DIR/1-macos-interpreter-server.log"; exit 1
	fi
	/usr/bin/time -p "$MAC_BIN" test_folder "$MAC_SOCK_DIR/fuzz.sock" "$TRACES" \
		> "$log" 2> "$timing"
	kill "$pid" 2>/dev/null || true
	wait "$pid" 2>/dev/null || true
	cleanup_mac
}

run_docker_backend() {
	local backend="$1"
	local num="$2"
	local log="$RESULT_DIR/${num}-docker-${backend}.log"
	local timing="$RESULT_DIR/${num}-docker-${backend}.time"
	cleanup_docker
	docker volume create "$SOCK_VOL" >/dev/null 2>&1 || true
	echo ">>> [${num}/3] Docker linux/amd64 ${backend}"
	docker run -d --rm --platform=linux/amd64 --name jam-bench-server \
		-e JAM_PVM_TRACE_DIR= \
		-v "${SOCK_VOL}:/sockdir" \
		"$IMAGE" \
		/sockdir/fuzz.sock -pvm-backend "$backend" \
		> /dev/null
	for _ in $(seq 1 30); do
		docker run --rm --platform=linux/amd64 -v "${SOCK_VOL}:/sockdir" alpine \
			test -S /sockdir/fuzz.sock 2>/dev/null && break
		sleep 1
	done
	/usr/bin/time -p docker run --rm --platform=linux/amd64 \
		-e JAM_PVM_TRACE_DIR= \
		-v "${SOCK_VOL}:/sockdir" \
		-v "$ROOT/pkg/test_data:/data:ro" \
		"$IMAGE" \
		test_folder /sockdir/fuzz.sock /data/jam-conformance/fuzz-reports/0.7.2/traces/ \
		> "$log" 2> "$timing"
	cleanup_docker
}

summarize() {
	local label="$1"
	local log="$2"
	local timing="$3"
	local passed failed real user sys
	passed="$(grep -c 'PASSED:' "$log" 2>/dev/null || echo 0)"
	failed="$(grep -c 'FAILED:' "$log" 2>/dev/null || echo 0)"
	real="$(awk '/^real/{print $2}' "$timing")"
	user="$(awk '/^user/{print $2}' "$timing")"
	sys="$(awk '/^sys/{print $2}' "$timing")"
	printf "%-28s | real %8ss | user %8ss | sys %6ss | pass %4s | fail %s\n" \
		"$label" "$real" "$user" "$sys" "$passed" "$failed"
}

run_mac_interpreter
run_docker_backend interpreter 2
run_docker_backend recompiler 3

echo ""
echo "=== Summary (client test_folder wall time, warm server) ==="
printf "%-28s | %-18s | %-18s | %-10s | %-10s\n" "Scenario" "real (s)" "user (s)" "passed" "failed"
printf "%-28s-+-%-18s-+-%-18s-+-%-10s-+-%-10s\n" "----------------------------" "------------------" "------------------" "----------" "----------"
summarize "macOS interpreter" "$RESULT_DIR/1-macos-interpreter.log" "$RESULT_DIR/1-macos-interpreter.time"
summarize "Docker interpreter" "$RESULT_DIR/2-docker-interpreter.log" "$RESULT_DIR/2-docker-interpreter.time"
summarize "Docker recompiler" "$RESULT_DIR/3-docker-recompiler.log" "$RESULT_DIR/3-docker-recompiler.time"
echo ""
echo "Logs: $RESULT_DIR/"
