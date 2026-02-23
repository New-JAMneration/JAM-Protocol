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
	    export USE_MINI_REDIS=true; \
	    go run ./cmd/node test --mode $(mode) --size $(size) --type $(type) --format $(format); \
	else \
		MODES="safrole assurances preimages disputes history accumulate authorizations statistics reports"; \
		for m in $$MODES; do \
			echo "Testing $$m (size=$(size))..."; \
			export USE_MINI_REDIS=true; \
			go run ./cmd/node test --mode "$$m" --size "$(size)" --type "$(type)" --format "$(format)"; \
			echo ""; \
		done; \
	fi

.PHONY: test-jam-test-vectors-trace
test-jam-test-vectors-trace:
	@if [ -n "$(mode)" ]; then \
		echo "Testing trace $(mode)..."; \
		export USE_MINI_REDIS=true; go run ./cmd/node test --type "trace" --mode "$(mode)"; \
	else \
		MODES="fallback safrole preimages_light preimages storage_light storage fuzzy_light"; \
		for mode in $$MODES; do \
			echo "Testing trace $$mode..."; \
			export USE_MINI_REDIS=true; go run ./cmd/node test --type "trace" --mode "$$mode"; \
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
		TIMING=1 USE_MINI_REDIS=true go run ./cmd/node test --type "trace" --mode "$(mode)"; \
	else \
		MODES="fallback safrole preimages_light preimages storage_light storage fuzzy_light"; \
		for mode in $$MODES; do \
			echo ""; \
			echo "========================================"; \
			echo "Testing trace $$mode with timing..."; \
			echo "========================================"; \
			TIMING=1 USE_MINI_REDIS=true go run ./cmd/node test --type "trace" --mode "$$mode"; \
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
	TIMING=1 USE_MINI_REDIS=true go run ./cmd/node test --type "trace" --mode "$(mode)" --benchmark 5

.PHONY: lint
lint:
	golangci-lint run $(args) ./...

.PHONY: lint-fix
lint-fix:
	@make lint args='--fix -v'

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: run-target
run-target:
	export USE_MINI_REDIS=true; \
	go run ./cmd/fuzz/ /tmp/jam_target.sock

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
