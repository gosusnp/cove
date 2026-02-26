run:
	go run main.go

test:
	go test ./...

test-verbose:
	go test -v -cover ./...

lint:
	go vet ./...
	golangci-lint run ./...

fmt:
	gofmt -w .
