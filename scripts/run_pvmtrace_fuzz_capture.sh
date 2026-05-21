#!/usr/bin/env bash
# Capture PVM per-instruction traces for a fuzz trace folder (interpreter + recompiler),
# deblob static metadata, and run offline pvm-diff.
#
# Usage:
#   scripts/run_pvmtrace_fuzz_capture.sh [trace-folder] [json-for-deblob]
#
# Defaults:
#   trace-folder  = pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/1766241814
#   json-for-deblob = <trace-folder>/00000179.json
#
# Env:
#   JAM_FUZZ_IMAGE     docker image (default: new-jamneration-target:trace)
#   PVMTRACE_OUT       host output root (default: ./pvmtrace-out)
#   PVM_DEBLOB_DIR     deblob output root (default: ./pvm-deblob)
#   SKIP_DOCKER_BUILD  set to 1 to skip image build
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

TRACE_FOLDER="${1:-pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/1766241814}"
DEBLOB_JSON="${2:-${TRACE_FOLDER}/00000179.json}"
IMAGE="${JAM_FUZZ_IMAGE:-new-jamneration-target:trace}"
OUT_ROOT="${PVMTRACE_OUT:-./pvmtrace-out}"
DEBLOB_ROOT="${PVM_DEBLOB_DIR:-./pvm-deblob}"
SOCK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/pvmtrace-sock.XXXXXX")"
SERVER_CID=""

cleanup() {
	if [[ -n "${SERVER_CID}" ]]; then
		docker rm -f "${SERVER_CID}" >/dev/null 2>&1 || true
	fi
	rm -rf "${SOCK_DIR}"
}
trap cleanup EXIT

if [[ ! -d "${TRACE_FOLDER}" ]]; then
	echo "trace folder not found: ${TRACE_FOLDER}" >&2
	exit 1
fi
if [[ ! -f "${DEBLOB_JSON}" ]]; then
	echo "deblob json not found: ${DEBLOB_JSON}" >&2
	exit 1
fi

mkdir -p "${OUT_ROOT}" "${DEBLOB_ROOT}" "${SOCK_DIR}"
chmod a+rwx "${SOCK_DIR}" 2>/dev/null || true

if [[ "${SKIP_DOCKER_BUILD:-0}" != "1" ]]; then
	echo "==> Building trace-enabled fuzz image (${IMAGE})..."
	docker buildx build --platform=linux/amd64 \
		--build-arg GP_VERSION="$(cat VERSION_GP)" \
		--build-arg TARGET_VERSION="$(cat VERSION_TARGET)" \
		--build-arg OUTPUT=new-jamneration-target \
		--build-arg BUILD_TAGS=trace \
		-t "${IMAGE}" \
		-f docker/Dockerfile --load .
fi

echo "==> Deblob program metadata from ${DEBLOB_JSON}..."
if ! docker run --rm \
	--entrypoint /build/pvmtrace \
	-v "${ROOT}:${ROOT}" \
	-w "${ROOT}" \
	-e PVM_DEBLOB_DIR="${ROOT}/${DEBLOB_ROOT}" \
	"${IMAGE}" \
	deblob program "${DEBLOB_JSON}"; then
	echo "warning: deblob failed (trace diff can still run without --meta)" >&2
fi

find_trace_dir() {
	local run_id="$1"
	local info
	info="$(find "${OUT_ROOT}/${run_id}" -type f -path '*/meta/info.json' 2>/dev/null | head -1)"
	if [[ -z "${info}" ]]; then
		return
	fi
	dirname "$(dirname "${info}")"
}

run_backend_trace() {
	local backend="$1"
	local run_id="$2"

	rm -rf "${OUT_ROOT:?}/${run_id}"
	mkdir -p "${OUT_ROOT}/${run_id}"

	echo "==> Starting fuzz server (backend=${backend}, run_id=${run_id})..."
	SERVER_CID="$(docker run -d --rm \
		--platform linux/amd64 \
		--init \
		-u "$(id -u):$(id -g)" \
		--cap-add IPC_LOCK \
		-v "${SOCK_DIR}:/tmp/jam_fuzz" \
		-v "${ROOT}/${OUT_ROOT}:/trace" \
		-v "${ROOT}/${TRACE_FOLDER}:/case:ro" \
		-e JAM_FUZZ=1 \
		-e JAM_FUZZ_SPEC=tiny \
		-e JAM_FUZZ_DATA_PATH=/tmp/jam_fuzz \
		-e JAM_FUZZ_SOCK_PATH=/tmp/jam_fuzz/fuzz.sock \
		-e JAM_FUZZ_LOG_LEVEL=info \
		-e JAM_PVM_TRACE_DIR=/trace \
		-e JAM_PVM_TRACE_RUN_ID="${run_id}" \
		-e USE_MINI_REDIS=true \
		"${IMAGE}" \
		--pvm-backend "${backend}" \
		/tmp/jam_fuzz/fuzz.sock)"

	# Wait for unix socket
	for _ in $(seq 1 60); do
		if [[ -S "${SOCK_DIR}/fuzz.sock" ]]; then
			break
		fi
		sleep 0.5
	done
	if [[ ! -S "${SOCK_DIR}/fuzz.sock" ]]; then
		echo "server failed to create socket; logs:" >&2
		docker logs "${SERVER_CID}" >&2 || true
		exit 1
	fi

	echo "==> Replaying fuzz folder via test_folder..."
	if ! docker run --rm \
		--platform linux/amd64 \
		--init \
		-u "$(id -u):$(id -g)" \
		-v "${SOCK_DIR}:/tmp/jam_fuzz" \
		-v "${ROOT}/${TRACE_FOLDER}:/case:ro" \
		-e JAM_FUZZ=1 \
		-e JAM_FUZZ_SPEC=tiny \
		-e JAM_FUZZ_DATA_PATH=/tmp/jam_fuzz \
		-e JAM_FUZZ_SOCK_PATH=/tmp/jam_fuzz/fuzz.sock \
		-e JAM_FUZZ_LOG_LEVEL=info \
		-e USE_MINI_REDIS=true \
		"${IMAGE}" \
		test_folder /tmp/jam_fuzz/fuzz.sock /case; then
		echo "warning: test_folder returned non-zero for backend=${backend} (trace may still be partial)" >&2
	fi

	docker logs "${SERVER_CID}" 2>&1 | tail -20 || true
	docker rm -f "${SERVER_CID}" >/dev/null
	SERVER_CID=""
}

run_backend_trace interpreter interp
run_backend_trace recompiler recomp

INTERP_DIR="$(find_trace_dir interp)"
RECOMP_DIR="$(find_trace_dir recomp)"

if [[ -z "${INTERP_DIR}" || -z "${RECOMP_DIR}" ]]; then
	echo "failed to locate trace output under ${OUT_ROOT}" >&2
	find "${OUT_ROOT}" -maxdepth 4 -type f 2>/dev/null || true
	exit 1
fi

DEBLOB_DIR="$(find "${DEBLOB_ROOT}" -name instr_meta.json.gz -print 2>/dev/null | head -1 | xargs dirname)"
if [[ -z "${DEBLOB_DIR}" ]]; then
	echo "warning: deblob metadata not found under ${DEBLOB_ROOT}" >&2
fi

echo ""
echo "Interpreter trace: ${INTERP_DIR}"
echo "Recompiler trace:  ${RECOMP_DIR}"
[[ -n "${DEBLOB_DIR}" ]] && echo "Deblob meta:       ${DEBLOB_DIR}"
echo ""

echo "==> pvm-diff find-diff"
docker run --rm \
	--platform linux/amd64 \
	--entrypoint /build/pvmtrace \
	-v "${ROOT}:${ROOT}" \
	-w "${ROOT}" \
	"${IMAGE}" \
	pvm-diff find-diff \
	--left "${INTERP_DIR}" \
	--right "${RECOMP_DIR}"

echo ""
echo "Done. Trace roots under ${OUT_ROOT}/"
