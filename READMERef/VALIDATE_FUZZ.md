# Fuzz validation specification (Agent / CI)

> **Convention:** For any change touching fuzz, the target, `cmd/fuzz`, or jam-conformance, read this document first, then run `make validate-fuzz` (or the GitHub Actions **Fuzz Validate** workflow).

## Glossary

| Term | Meaning |
|------|---------|
| **target** | Our fuzz server (`go run ./cmd/fuzz/` or Docker image `new-jamneration-target`) |
| **client** | Our fuzz client (`go run ./cmd/fuzz/ test_folder …`) or external **polkajamfuzz** (may not run on all platforms) |
| **jam-conformance** | Shared Polkadot config and scripts: `./pkg/test_data/jam-conformance` |

## Validation checklist (pass criteria)

### 1. `make test-jam-test-vectors`

Implemented as `make validate-fuzz-vectors` / `scripts/validate_fuzz.sh` step 1.

- **statistics** (tiny): expect **1/3 pass** → `Passed=1`, `Failed=2` (current correct behavior, not a bug).
- **All other modes** (safrole, assurances, preimages, disputes, history, accumulate, authorizations, reports): **Failed must be 0**.

> Note: `cmd/node test` does not exit non-zero when `failed > 0`; the script parses `Total: …, Passed: …, Failed: …` from output.

### 2. `make test-jam-test-vectors-trace`

Implemented as `make validate-fuzz-trace` / step 2.

- All trace modes (fallback, safrole, preimages_light, preimages, storage_light, storage, fuzzy_light): **Failed = 0**.

### 3. Target + `test_folder` (conformance traces)

Implemented as `make validate-fuzz-sock` / step 3.

1. Terminal A: `make run-target` (or `make fuzz-docker-run`)
2. Terminal B:

```bash
go run ./cmd/fuzz/ test_folder \
  ./.jam_fuzz_docker_run/fuzz.sock \
  ./pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/ \
  > output.txt
```

- **Pass:** `output.txt` must not contain `FAILED:` or `FAILED!!:`.
- Manual steps: [PR_MANUAL_TEST_FUZZ_TARGET.md](./PR_MANUAL_TEST_FUZZ_TARGET.md)

### 3b. Target + `test_folder` (fuzzy traces)

Implemented as `make validate-fuzz-fuzzy` / step `fuzzy` (same target process as step 3).

Start the target (same as step 3), then:

```bash
go run ./cmd/fuzz/ test_folder \
  ./.jam_fuzz_docker_run/fuzz.sock \
  ./pkg/test_data/jam-test-vectors/traces/fuzzy/ \
  > output_fuzzy.txt
```

- **Pass:** `output_fuzzy.txt` (or `VALIDATE_FUZZ_OUTPUT_FUZZY`) has no `FAILED` / `FAILED!!:`.

**Common reasons the client sees `EOF`** (when setup is correct):

1. Target process killed by OOM (long trace runs, unbounded memory).
2. Target closes the connection: `SetState` failure, STF runtime error, request decode failure (protocol `ImportBlock` errors usually return `ErrorMessage`, not EOF).
3. Target panic (recovered per connection; should not take down the whole process).

With `JAM_FUZZ=1`, the target should use **in-memory persistence only** and **≤ 24 unfinalized blocks** so Pebble does not grow and memory stays bounded.

### 4. FluffyLabs jam-testing (minifuzz / picofuzz)

Step 4 (skipped by default; optional local smoke).

- Repo: <https://github.com/FluffyLabs/jam-testing>
- Official CI runs your Docker image on self-hosted runners (e.g. `ghcr.io/new-jamneration/new-jamneration-target:latest`) for minifuzz → picofuzz.
- **Local run** requires Docker (e.g. `make fuzz-docker-build`), Node.js 20+, and a clone of jam-testing. On arm64/aarch64 hosts, build/run scripts default to `DOCKER_PLATFORM=--platform linux/amd64`; see [RELEASE_AND_PUBLISH.md § Docker platform](./RELEASE_AND_PUBLISH.md#docker-platform-docker_platform).

```bash
# Optional one-shot smoke (minifuzz fallback only)
make validate-fuzz-jam-testing-local
# Or:
./scripts/run_jam_testing_local.sh
```

Env: `TARGET_IMAGE`, `JAM_TESTING_DIR`, `JAM_TESTING_SUITE` (default `fallback`).

**Limits:** Full picofuzz / dashboard registration is on FluffyLabs GitHub Actions; local scripts are dev smoke only.

**Common error:** `git submodule update --recursive` hanging on `Enter passphrase for key ... git@github.com:davxy/jam-test-vectors` — nested submodule in picofuzz-data uses SSH. Use `./scripts/ci_jam_testing_full.sh` (HTTPS rewrite, minimal submodules) or remove `.ci/jam-testing` and retry.

## Automation

### Makefile

| Target | Description |
|--------|-------------|
| `make validate-fuzz` | Steps 1–3 + fuzzy (full local run) |
| `make validate-fuzz-ci` | Steps 1 + 2 + 3 (full conformance tree) + fuzzy; aligns with CI `jam-test-vectors` / `jam-test-vectors-traces` / sock jobs |
| `make validate-fuzz-vectors` | Step 1 only |
| `make validate-fuzz-trace` | Step 2 only |
| `make validate-fuzz-sock` | Step 3 only (full conformance traces) |
| `make validate-fuzz-sock-smoke` | Step 3 smoke (default `1766241814`) |
| `make validate-fuzz-fuzzy` | Fuzzy `test_folder` only |
| `make validate-fuzz-jam-testing-local` | Optional jam-testing minifuzz smoke |

### Environment variables (`scripts/validate_fuzz.sh`)

| Variable | Default | Description |
|----------|---------|-------------|
| `VALIDATE_FUZZ_STEPS` | `1,2,3,fuzzy` | Steps: `1`, `2`, `3`, `fuzzy`, `4` |
| `FUZZ_SMOKE_TRACE_DIR` | (empty) | If set, step 3 runs only that subdirectory |
| `VALIDATE_FUZZ_OUTPUT` | `output.txt` | Step 3 client output |
| `VALIDATE_FUZZ_OUTPUT_FUZZY` | `output_fuzzy.txt` | Fuzzy client output |
| `VALIDATE_FUZZ_TARGET_LOG` | `$JAM_FUZZ_HOST_DIR/fuzz_target.log` | Background target log (step 3 / fuzzy) |
| `FUZZ_FUZZY_TRACES` | `pkg/test_data/jam-test-vectors/traces/fuzzy` | Fuzzy trace path |
| `JAM_FUZZ_HOST_DIR` | `.jam_fuzz_docker_run` | Same as `make run-target` |
| `STATISTICS_EXPECT_PASSED` / `FAILED` | `1` / `2` | Expected statistics 1/3 |
| `VALIDATE_FUZZ_RUN_JAM_TESTING` | `0` | Set `1` to run `run_jam_testing_local.sh` |

### jam-conformance traces (pinned submodule — manual bump)

- **CI and local full runs** use the commit pinned in **`pkg/test_data/jam-conformance`** (`checkout` with `submodules: recursive`).
- **CI does not track conformance `main`** (avoids unrelated PR failures when upstream traces change).
- To update manually:

```bash
git submodule update --init pkg/test_data/jam-conformance
cd pkg/test_data/jam-conformance
git fetch origin
git checkout origin/main   # or a specific tag/commit
cd ../../..
git add pkg/test_data/jam-conformance
git commit -m "chore: bump jam-conformance submodule"
```

- **jam-testing** (minifuzz / picofuzz) is cloned from `FluffyLabs/jam-testing` at a **pinned commit** in `scripts/ci_jam_testing_full.sh` (`JAM_TESTING_REF`, default `028072cc…`); bump intentionally — not `main`.
- **Fuzz target memory:** `JAM_FUZZ=1` uses in-memory persistence; successful imports run `PruneOldData` + `TrimUnfinalizedBlocksForFuzz` (keep ≤24). Protocol-invalid blocks are **not** reverted (revert broke 18/1179 conformance traces locally).

### CI (full gate per PR)

Workflow: `.github/workflows/fuzz-validate.yml`

| Job | Content |
|-----|---------|
| `build-fuzz-target-image` | Build `new-jamneration-target:ci` once; artifact for downstream jobs |
| `jam-test-vectors` | Step 1 (jam-test-vectors shell tests) |
| `jam-test-vectors-traces` | Step 2 (trace modes) |
| `jam-conformance-sock` | Docker target + conformance `fuzz-reports/0.7.2/traces` (1179) |
| `jam-test-vectors-traces-fuzzy-sock` | Docker target + `jam-test-vectors/traces/fuzzy` (201) |
| `jam-testing-minifuzz-picofuzz` | Minifuzz 6 + picofuzz 4 suites (`scripts/ci_jam_testing_full.sh`) |
| `fuzz-validate-all-green` | Summary; all jobs must succeed |

- Triggers: `pull_request` / `push` to `main` (path filters) / `workflow_dispatch`
- Local equivalent: `make validate-fuzz` + `make fuzz-docker-build` + `./scripts/ci_fuzz_docker_sock.sh` + `./scripts/ci_jam_testing_full.sh`
- Conformance smoke only: `make validate-fuzz-sock-smoke` (`FUZZ_SMOKE_TRACE_DIR=1766241814`)

## Pre-push local smoke (before opening PR)

Requires submodules initialized locally (`git submodule update --init pkg/Rust-VRF pkg/test_data/jam-test-vectors pkg/test_data/jam-conformance`).

```bash
./scripts/ci_fuzz_pre_push.sh
```

Covers: VRF build (if needed), Docker image, `make validate-fuzz-ci`, and `ci_fuzz_docker_sock.sh`. Does **not** run jam-testing (add `./scripts/ci_jam_testing_full.sh` manually).

**CI note:** `Fuzz Validate` uses `secrets.TOKEN` (PAT, `repo` scope) for private `pkg/Rust-VRF`, via `x-access-token` URL rewrite (same intent as `release.yml`). If the secret is missing or the PR is from a fork without secret access, jobs fail with “repository not found” for `Rust-VRF.git`.

## Agent checklist

1. Read this file and [PR_MANUAL_TEST_FUZZ_TARGET.md](./PR_MANUAL_TEST_FUZZ_TARGET.md) when changing the target.
2. Run `make validate-fuzz` or at least `make validate-fuzz-ci`.
3. If the change affects the published image or performance registration, run jam-testing picofuzz (step 4).
4. In the PR description, include `make validate-fuzz-ci` results (statistics 1/3; no FAILED in fuzzy/conformance output). On sock failures, check `$JAM_FUZZ_HOST_DIR/fuzz_target.log` for panics.

## Reference (aligned with implementation)

```
FUZZ
target: our
client: our client (run target) or polkjamfuzz
config: ./pkg/test_data/jam-conformance
```
