.PHONY: run test coverage vet fmt tidy build

run:
	go run ./cmd/server

test:
	go test ./...

coverage:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

build:
	go build -o bin/manager-task ./cmd/server