all: lint test build

.PHONY: lint
lint: custom_lint
	golangci-lint run ./...

.PHONY: custom_lint
custom_lint:
	go run ./tools/lint-grouped-imports

.PHONY: test
test:
	go test -coverprofile=cover.out ./...

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: build
build:
	go build -o distress-agent ./cmd/agent

