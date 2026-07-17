# The version variables are read from the VERSION_GP and VERSION_TARGET files
VERSION_GP := $(shell cat VERSION_GP)
VERSION_TARGET := $(shell cat VERSION_TARGET)

.PHONY: run
run:
	go run ./cmd/node

.PHONY: build
build:
	go build -o ./build/JAM-Protocol ./cmd/node

.PHONY: test
test:
	go test $(args) ./...

# Test defaults (matching cmd/node/test.go)
size ?= tiny
type ?= jam-test-vectors
format ?= binary

.PHONY: test-jam-test-vectors
test-jam-test-vectors:
	@if [ -n "$(mode)" ]; then \
	    echo "Testing $(mode) (size=$(size), type=$(type), format=$(format))..."; \
	    go run ./cmd/node test --mode $(mode) --size $(size) --type $(type) --format $(format); \
	else \
		MODES="safrole assurances preimages disputes history accumulate authorizations statistics reports"; \
		for m in $$MODES; do \
			echo "Testing $$m (size=$(size))..."; \
			go run ./cmd/node test --mode "$$m" --size "$(size)" --type "$(type)" --format "$(format)"; \
			echo ""; \
		done; \
	fi

.PHONY: test-jam-test-vectors-trace
test-jam-test-vectors-trace:
	@if [ -n "$(mode)" ]; then \
		echo "Testing trace $(mode)..."; \
		go run ./cmd/node test --type "trace" --mode "$(mode)"; \
	else \
		MODES="fallback safrole preimages_light preimages storage_light storage fuzzy_light"; \
		for mode in $$MODES; do \
			echo "Testing trace $$mode..."; \
			go run ./cmd/node test --type "trace" --mode "$$mode"; \
			echo ""; \
		done; \
	fi

# Fuzz conformance timing test (target server + test_folder client).
# scripts/run_fuzz_timing_test.sh is self-contained: it prepares the trace symlink,
# builds the target, replays the folder, and prints the STF + Psi_A timing summary on
# shutdown. PVM_BACKEND defaults to the script's default (interpreter); override via env,
# e.g. `PVM_BACKEND=recompiler make test-timing-fuzz-trace`.
# Usage: make test-timing-fuzz-trace                          # all folders
#        make test-timing-fuzz-trace folder=1766241814        # one folder
FUZZ_TRACES ?= target/fuzz

.PHONY: test-timing-fuzz-trace
test-timing-fuzz-trace:
	bash scripts/run_fuzz_timing_test.sh $(if $(folder),$(FUZZ_TRACES)/$(folder),)

# Test with detailed timing breakdown for trace tests
# Usage: make test-timing-jam-test-vectors-trace mode=safrole
#        make test-timing-jam-test-vectors-trace (runs all trace modes)
.PHONY: test-timing-jam-test-vectors-trace
test-timing-jam-test-vectors-trace:
	@if [ -n "$(mode)" ]; then \
		echo "Testing trace $(mode) with timing..."; \
		TIMING=1 go run ./cmd/node test --type "trace" --mode "$(mode)"; \
	else \
		MODES="fallback safrole preimages_light preimages storage_light storage fuzzy_light"; \
		for mode in $$MODES; do \
			echo ""; \
			echo "========================================"; \
			echo "Testing trace $$mode with timing..."; \
			echo "========================================"; \
			TIMING=1 go run ./cmd/node test --type "trace" --mode "$$mode"; \
		done; \
	fi

# Benchmark trace tests with 5 runs (only reports statistics if all runs pass)
# Usage: make test-benchmark-trace mode=safrole
.PHONY: test-benchmark-trace
test-benchmark-trace:
	@if [ -z "$(mode)" ]; then \
		echo "Error: mode is required for benchmark. Usage: make test-benchmark-trace mode=safrole"; \
		exit 1; \
	fi
	@echo "Running benchmark for trace $(mode) (5 runs)..."
	TIMING=1 go run ./cmd/node test --type "trace" --mode "$(mode)" --benchmark 5

.PHONY: lint
lint:
	golangci-lint run $(args) ./...

.PHONY: lint-fix
lint-fix:
	@make lint args='--fix -v'

.PHONY: fmt
fmt:
	go fmt ./...

# Fuzz host dir (matches scripts/run_fuzz_target_docker.sh default bind-mount path on host).
JAM_FUZZ_HOST_DIR ?= .jam_fuzz_docker_run
# Recompiler is linux/amd64-only (cmd/fuzz errors if requested elsewhere), so
# pick the default per platform. Override: make run-target PVM_BACKEND=interpreter
ifeq ($(shell go env GOOS GOARCH),linux amd64)
PVM_BACKEND ?= recompiler
else
PVM_BACKEND ?= interpreter
endif

.PHONY: run-target
run-target:
	mkdir -p $(JAM_FUZZ_HOST_DIR)
	JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny JAM_PVM_BACKEND=$(PVM_BACKEND) JAM_FUZZ_DATA_PATH=$(JAM_FUZZ_HOST_DIR)/ JAM_FUZZ_SOCK_PATH=$(JAM_FUZZ_HOST_DIR)/fuzz.sock go run ./cmd/fuzz/

JAM_FUZZ_IMAGE ?= new-jamneration-target:latest
# Matches CI release (linux/amd64). Required on arm64/aarch64 hosts (Apple Silicon, Linux ARM).
DOCKER_PLATFORM ?= --platform linux/amd64
JAM_FUZZ_TRACE_IMAGE ?= new-jamneration-target:trace

# Debug image (:trace): BUILD_TAGS=trace compiles the PVMtrace recorder in, for
# pvmtrace-fuzz-capture / pvm-diff. Not for conformance runs — tracing adds overhead.
.PHONY: fuzz-docker-build-trace
fuzz-docker-build-trace:
	docker buildx build $(DOCKER_PLATFORM) \
		--build-arg GP_VERSION=$(VERSION_GP) \
		--build-arg TARGET_VERSION=$(VERSION_TARGET) \
		--build-arg OUTPUT=new-jamneration-target \
		--build-arg BUILD_TAGS=trace \
		-t $(JAM_FUZZ_TRACE_IMAGE) \
		-f docker/Dockerfile \
		--load .

# Capture interpreter + recompiler PVM traces for a fuzz folder and run pvm-diff.
# The script rebuilds the :trace image each run; SKIP_DOCKER_BUILD=1 reuses an existing one.
# Usage: make pvmtrace-fuzz-capture TRACE_FOLDER=pkg/test_data/.../1766241814
TRACE_FOLDER ?= pkg/test_data/jam-conformance/fuzz-reports/0.7.2/traces/1766241814
DEBLOB_JSON ?= $(TRACE_FOLDER)/00000179.json

.PHONY: pvmtrace-fuzz-capture
pvmtrace-fuzz-capture:
	JAM_FUZZ_IMAGE=$(JAM_FUZZ_TRACE_IMAGE) \
		bash scripts/run_pvmtrace_fuzz_capture.sh "$(TRACE_FOLDER)" "$(DEBLOB_JSON)"

# Production image (:latest): no build tags, trace hooks compile to no-ops.
# Used by CI releases and fuzz-docker-run. buildx: the Dockerfile needs BuildKit.
.PHONY: fuzz-docker-build
fuzz-docker-build:
	docker buildx build $(DOCKER_PLATFORM) \
		--build-arg GP_VERSION=$(VERSION_GP) \
		--build-arg TARGET_VERSION=$(VERSION_TARGET) \
		--build-arg OUTPUT=new-jamneration-target \
		-t $(JAM_FUZZ_IMAGE) \
		-f docker/Dockerfile \
		--load .

.PHONY: fuzz-docker-run
fuzz-docker-run:
	bash scripts/run_fuzz_target_docker.sh

# The command build the target binary locally
# For release builds, use `make release-target` instead
.PHONY: build-target
build-target:
	go build -ldflags "-X 'main.GP_VERSION=$(VERSION_GP)' -X 'main.TARGET_VERSION=$(VERSION_TARGET)'" -o ./build/new-jamneration-target ./cmd/fuzz

# The command use docker to build the release target binary
# It will copy the built binary to ./build/new-jamneration-target
.PHONY: release-target
release-target:
	git submodule update --init pkg/Rust-VRF
	bash ./scripts/release.sh $(VERSION_GP) $(VERSION_TARGET)

# The command run the release target binary in a docker container
.PHONY: run-release-target
run-release-target:
	bash ./scripts/run_release.sh

.PHONY: validate-fuzz validate-fuzz-ci validate-fuzz-vectors validate-fuzz-trace validate-fuzz-sock validate-fuzz-sock-smoke validate-fuzz-fuzzy validate-fuzz-jam-testing-local
validate-fuzz:
	$(VALIDATE_FUZZ_SCRIPT)

# Build the recompiler unit-test container (linux/amd64)
.PHONY: build-recompiler-test-env
build-recompiler-test-env:
	docker build $(DOCKER_PLATFORM) -t go-jit-test -f PVM/Dockerfile .

# The command run the ASM test in a docker container
.PHONY: run-asm-test
run-asm-test:
	docker run --rm -it \
		$(DOCKER_PLATFORM) \
		--security-opt seccomp=unconfined \
		--privileged \
		-v "$(shell pwd)":/app \
		go-jit-test \
		go test -v ./PVM/recompiler/asm/.

# Run recompiler compiler_test + signal_handler_test in docker
.PHONY: run-recompiler-test
run-recompiler-test:
	docker run --rm -it \
		$(DOCKER_PLATFORM) \
		--security-opt seccomp=unconfined \
		--privileged \
		-v "$(shell pwd)":/app \
		go-jit-test \
		go test -v ./PVM/recompiler/...
