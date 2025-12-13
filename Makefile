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

.PHONY: test-jam-test-vectors
test-jam-test-vectors:
	@if [ -n "$(mode)" ] && [ -n "$(size)" ]; then \
	    echo "Testing $(mode) ($(size))..."; \
	    export USE_MINI_REDIS=true; go run ./cmd/node test --mode "$(mode)" --size "$(size)"; \
	else \
		MODES="safrole assurances preimages disputes history accumulate authorizations statistics reports"; \
		SIZES="tiny full"; \
		for mode in $$MODES; do \
			for size in $$SIZES; do \
				echo "Testing $$mode ($$size)..."; \
				export USE_MINI_REDIS=true; go run ./cmd/node test --mode "$$mode" --size "$$size"; \
				echo ""; \
			done; \
		done; \
	fi


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
	export USE_MINI_REDIS=true; go run ./cmd/fuzz/ /tmp/jam_target.sock

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
