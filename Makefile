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

.PHONY: run-target
run-target:
	mkdir -p $(JAM_FUZZ_HOST_DIR)
	JAM_FUZZ=1 JAM_FUZZ_SPEC=tiny JAM_FUZZ_DATA_PATH=$(JAM_FUZZ_HOST_DIR)/ JAM_FUZZ_SOCK_PATH=$(JAM_FUZZ_HOST_DIR)/fuzz.sock go run ./cmd/fuzz/

JAM_FUZZ_IMAGE ?= new-jamneration-target:latest

.PHONY: fuzz-docker-build
fuzz-docker-build:
	docker build \
		--build-arg GP_VERSION=$(VERSION_GP) \
		--build-arg TARGET_VERSION=$(VERSION_TARGET) \
		--build-arg OUTPUT=new-jamneration-target \
		-t $(JAM_FUZZ_IMAGE) \
		-f docker/Dockerfile .

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

# --- Fuzz validation (spec: READMERef/VALIDATE_FUZZ.md) ---
VALIDATE_FUZZ_SCRIPT := ./scripts/validate_fuzz.sh
FUZZ_SMOKE_TRACE_DIR ?= 1766241814

.PHONY: validate-fuzz validate-fuzz-ci validate-fuzz-vectors validate-fuzz-trace validate-fuzz-sock validate-fuzz-sock-smoke validate-fuzz-fuzzy validate-fuzz-jam-testing-local
validate-fuzz:
	$(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-ci:
	VALIDATE_FUZZ_STEPS=1,2,3,fuzzy $(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-vectors:
	VALIDATE_FUZZ_STEPS=1 $(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-trace:
	VALIDATE_FUZZ_STEPS=2 $(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-sock:
	VALIDATE_FUZZ_STEPS=3 $(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-sock-smoke:
	VALIDATE_FUZZ_STEPS=3 FUZZ_SMOKE_TRACE_DIR=$(FUZZ_SMOKE_TRACE_DIR) $(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-fuzzy:
	VALIDATE_FUZZ_STEPS=fuzzy $(VALIDATE_FUZZ_SCRIPT)

validate-fuzz-jam-testing-local:
	VALIDATE_FUZZ_RUN_JAM_TESTING=1 VALIDATE_FUZZ_STEPS=4 $(VALIDATE_FUZZ_SCRIPT)
