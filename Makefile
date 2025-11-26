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
	@MODES="safrole assurances preimages disputes history accumulate authorizations statistics reports"; \
	SIZES="tiny full"; \
	for mode in $$MODES; do \
	    for size in $$SIZES; do \
	        echo "Testing $$mode ($$size)..."; \
	        go run ./cmd/node test --mode "$$mode" --size "$$size"; \
	        echo ""; \
	    done; \
	done

.PHONY: lint
lint:
	golangci-lint run $(args) ./...

.PHONY: lint-fix
lint-fix:
	@make lint args='--fix -v'

.PHONY: fmt
fmt:
	go fmt ./...
