-include .env
export

COVE_API_KEY ?= dev-key
DATABASE_URL ?= postgres://cove:covedev@localhost:5432/cove?sslmode=disable

run:
	go run main.go

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

devenv.up:
	docker compose -f ../docker-compose.yml up -d

devenv.down:
	docker compose -f ../docker-compose.yml down

devenv.reset:
	docker compose -f ../docker-compose.yml down -v

check: lint test

fix: license-fix fmt

.PHONY: lint fmt license-fix check fix devenv.up devenv.down devenv.reset
