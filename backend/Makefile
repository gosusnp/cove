COVE_API_KEY ?= dev-key

run:
	COVE_API_KEY=$(COVE_API_KEY) go run main.go

build-mcp:
	go build -o bin/cove-mcp ./cmd/mcp

test:
	go tool gotestsum --format pkgname-and-test-fails -- -cover ./...

lint:
	go vet ./...
	go tool golangci-lint run ./...

fmt:
	gofmt -w .

license-fix:
	go tool addlicense -f ../.license_header.txt .

check: lint test

fix: license-fix fmt

.PHONY: lint fmt license-fix check fix
