.PHONY: run test test-integration coverage vet fmt tidy build \
        docker-up docker-down \
        migration-up migration-down migration-version \
        migration-test-up migration-test-version

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
	go build -o bin/task-manager ./cmd/server

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migration-up:
	migrate \
		-path migrations \
		-database "$$DATABASE_URL" \
		up

migration-down:
	migrate \
		-path migrations \
		-database "$$DATABASE_URL" \
		down 1

migration-version:
	migrate \
		-path migrations \
		-database "$$DATABASE_URL" \
		version

test-integration:
	@TEST_DATABASE_URL="$${TEST_DATABASE_URL:-postgres://task_manager:task_manager@localhost:5432/task_manager_test?sslmode=disable}" \
	go test ./internal/repository -run 'TestPostgresTaskRepository' -v

migration-test-up:
	migrate \
		-path migrations \
		-database "$$TEST_DATABASE_URL" \
		up

migration-test-version:
	migrate \
		-path migrations \
		-database "$$TEST_DATABASE_URL" \
		version