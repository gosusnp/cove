COVE_API_KEY ?= dev-key

run:
	COVE_API_KEY=$(COVE_API_KEY) go run main.go

build-mcp:
	go build -o bin/cove-mcp ./cmd/mcp

test:
	go test ./...

test-verbose:
	go test -v -cover ./...

lint:
	go vet ./...
	golangci-lint run ./...

fmt:
	gofmt -w .

check: fmt lint test
