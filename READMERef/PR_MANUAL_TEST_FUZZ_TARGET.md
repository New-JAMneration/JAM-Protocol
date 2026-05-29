# Manual fuzz target verification (for PR descriptions)

Run from the **repository root** in two terminals: **A = target**, **B = client**.

Default host paths match `Makefile` and `scripts/run_fuzz_target_docker.sh` (`.jam_fuzz_docker_run/`; socket file **`fuzz.sock`**).

Full validation spec: [VALIDATE_FUZZ.md](./VALIDATE_FUZZ.md).

---

## Unix socket (pick one target mode)

| Setup | Socket path (from repo root) |
|-------|------------------------------|
| `make fuzz-docker-run` or `./scripts/run_fuzz_target_docker.sh` | `./.jam_fuzz_docker_run/fuzz.sock` |
| `make run-target` (local Go) | Same (default `JAM_FUZZ_HOST_DIR=.jam_fuzz_docker_run`) |

If you use another host directory: set `JAM_FUZZ_HOST_DATA=…` for Docker, or `make run-target JAM_FUZZ_HOST_DIR=…` for local Go, and use `fuzz.sock` under that directory as `--target-sock`.

---

## A. Target — Docker

```bash
make fuzz-docker-build
./scripts/run_fuzz_target_docker.sh
# or: make fuzz-docker-run
```

Default image: `new-jamneration-target:latest` (`Makefile` `JAM_FUZZ_IMAGE`).

---

## B. Target — local `go run`

```bash
make run-target
```

Equivalent manual environment:

```bash
mkdir -p .jam_fuzz_docker_run
JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny \
  JAM_FUZZ_DATA_PATH=.jam_fuzz_docker_run/ \
  JAM_FUZZ_SOCK_PATH=.jam_fuzz_docker_run/fuzz.sock \
  go run ./cmd/fuzz/
```

---

## Host: `test_folder` (client)

**Start B only after A is listening.**

Full trace tree (slow / large; use a single subdirectory for a quick check):

```bash
go run ./cmd/fuzz/ test_folder \
  ./.jam_fuzz_docker_run/fuzz.sock \
  ./pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/
```

Short smoke (replace `YOUR_TRACE_DIR` with one subdirectory name):

```bash
go run ./cmd/fuzz/ test_folder \
  ./.jam_fuzz_docker_run/fuzz.sock \
  ./pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/YOUR_TRACE_DIR/
```

---

## Expected results

- **Target:** listens successfully first (no `bind: permission denied`; `dial … no such file` on the client usually means the server is not up).
- **Client:** socket path must match the target exactly.

---

## Troubleshooting

| Symptom | Action |
|---------|--------|
| `dial unix … no such file` | Confirm target is running and the socket path matches the table above. |
| Docker `bind … permission denied` | Fix permissions on the host bind mount, or use another `JAM_FUZZ_HOST_DATA` and `mkdir -p`. |
| Client `EOF` on ImportBlock | Check target logs for panic/OOM; see [VALIDATE_FUZZ.md](./VALIDATE_FUZZ.md) § EOF. With `make validate-fuzz-sock`, see `.jam_fuzz_docker_run/fuzz_target.log`. |
| Fuzz needs Redis | On the **target** side, set `USE_MINI_REDIS=true` if required by your local setup. |
