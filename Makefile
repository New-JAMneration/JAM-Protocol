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
