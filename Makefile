.PHONY: run
run:
	go run ./cmd/server/main.go

.PHONY: build
build:
	go build ./...

.PHONY: test
test:
	go test $(args) ./...

.PHONY: lint
lint:
	golangci-lint run $(args) ./...

.PHONY: lint-fix
lint-fix:
	@make lint args='--fix -v'

.PHONY: fmt
fmt:
	go fmt ./...
