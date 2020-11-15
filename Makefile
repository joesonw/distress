all: lint test build

.PHONY: lint
lint:
	go run ./tools/lint-grouped-imports
	golangci-lint run ./...

.PHONY: test
test:
	go test -coverprofile=cover.out ./...

.PHONY: cover
cover:
	go tool cover -html cover.out

.PHONY: build
build:
	go build -o distress-agent ./cmd/agent

