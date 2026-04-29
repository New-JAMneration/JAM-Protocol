#!/usr/bin/env bash
# host bind-mount is arbitrary. Default: <repo>/.jam_fuzz_docker_run (script path, not cwd). Override: JAM_FUZZ_HOST_DATA=...
set -euo pipefail

IMAGE="${JAM_FUZZ_IMAGE:-new-jamneration-target:latest}"
HOST_DATA="${JAM_FUZZ_HOST_DATA:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/.jam_fuzz_docker_run}"

mkdir -p "${HOST_DATA}"
chmod a+rwx "${HOST_DATA}" 2>/dev/null || true

exec docker run --rm \
	--init \
	-u "$(id -u):$(id -g)" \
	--cap-add IPC_LOCK \
	-e JAM_FUZZ=1 \
	-e "JAM_FUZZ_SPEC=${JAM_FUZZ_SPEC:-tiny}" \
	-e JAM_FUZZ_DATA_PATH=/tmp/jam_fuzz \
	-e JAM_FUZZ_SOCK_PATH=/tmp/jam_fuzz/fuzz.sock \
	-e "JAM_FUZZ_LOG_LEVEL=${JAM_FUZZ_LOG_LEVEL:-info}" \
	-v "${HOST_DATA}:/tmp/jam_fuzz" \
	"${IMAGE}" "$@"
