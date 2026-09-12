# Task Manager

A small task-management backend microservice built with Go, Gin, and PostgreSQL.

The service provides a RESTful API for creating, reading, updating, deleting, filtering, and paginating tasks.

---

## Architecture

The project is designed as an independently deployable **Task Manager microservice**.

It follows a clean architecture that separates HTTP handling, business logic, persistence, and infrastructure concerns.

```text
                         ┌─────────────────────┐
                         │      Frontend       │
                         │   Web / Mobile UI   │
                         └──────────┬──────────┘
                                    │
                                  REST
                                    │
                                    ▼
                    ┌──────────────────────────────┐
                    │      Task Manager Service    │
                    │                              │
                    │  ┌────────────────────────┐  │
                    │  │      HTTP Handler       │  │
                    │  │        Gin             │  │
                    │  └───────────┬────────────┘  │
                    │              │               │
                    │              ▼               │
                    │  ┌────────────────────────┐  │
                    │  │    Service / Use Case   │  │
                    │  │    Business Logic       │  │
                    │  └───────────┬────────────┘  │
                    │              │               │
                    │              ▼               │
                    │  ┌────────────────────────┐  │
                    │  │  Repository Interface  │  │
                    │  └───────────┬────────────┘  │
                    └──────────────┼───────────────┘
                                   │
                                   ▼
                    ┌──────────────────────────────┐
                    │          PostgreSQL           │
                    │        Task Persistence       │
                    └──────────────────────────────┘
```

### Request flow

A typical request follows this flow:

```text
HTTP Request
     │
     ▼
Request ID Middleware
     │
     ▼
Tracing Middleware
     │
     ▼
Logging / Metrics Middleware
     │
     ▼
Gin Handler
     │
     ▼
Task Service
     │
     ▼
Task Repository Interface
     │
     ▼
PostgreSQL
     │
     ▼
HTTP Response
```

### Architectural responsibilities

#### HTTP Handler

Responsible for:

* Parsing HTTP requests
* Validating request parameters
* Converting HTTP input into service inputs
* Calling the service layer
* Mapping domain/service errors to HTTP responses
* Returning consistent JSON responses

#### Service / Use Case

Responsible for:

* Business rules
* Task validation
* Default task status
* Pagination calculations
* Coordinating repository operations
* Translating persistence failures into service-level errors

#### Repository

Responsible for:

* PostgreSQL queries
* Persistence operations
* Database error handling
* Mapping database records to domain objects

The service depends on the repository **interface**, rather than directly depending on PostgreSQL.

This makes the business logic independently testable with mocks.

#### Infrastructure

Infrastructure components include:

* PostgreSQL connection pooling
* Database migrations
* Prometheus metrics
* OpenTelemetry tracing
* Structured logging
* Docker
* pprof runtime profiling

---

## Microservice Design

The Task Manager is a standalone microservice rather than a monolithic application containing unrelated business domains.

It has:

* Its own HTTP API
* Its own domain model
* Its own persistence layer
* Its own configuration
* Its own container image
* Its own health endpoint
* Its own metrics
* Its own tracing/logging
* Independent deployment and scaling capability

The service can therefore exist as one component in a larger distributed system:

```text
                         ┌──────────────┐
                         │   Frontend   │
                         └───────┬──────┘
                                 │
                 ┌───────────────┼───────────────┐
                 │               │               │
                 ▼               ▼               ▼
        ┌────────────────┐ ┌──────────────┐ ┌──────────────────┐
        │ Task Manager   │ │ User Service │ │ Other Services   │
        │ Microservice   │ │              │ │                  │
        └───────┬────────┘ └──────────────┘ └──────────────────┘
                │
                ▼
        ┌────────────────┐
        │   PostgreSQL   │
        └────────────────┘
```

The current project implements the **Task Manager service** only.

---

## Technology Stack

* **Language:** Go
* **HTTP framework:** Gin
* **Database:** PostgreSQL
* **Database driver:** pgx
* **Connection pooling:** pgxpool
* **Migrations:** golang-migrate
* **Metrics:** Prometheus
* **Tracing:** OpenTelemetry
* **Logging:** Go `slog`
* **Containerization:** Docker
* **Orchestration for local development:** Docker Compose
* **Runtime profiling:** Go pprof
* **API specification:** OpenAPI

---

## Task Model

Each task contains:

| Field         | Type      | Description               |
| ------------- | --------- | ------------------------- |
| `id`          | UUID      | Unique task identifier    |
| `title`       | string    | Task title                |
| `description` | string    | Optional task description |
| `status`      | string    | Task status               |
| `assignee`    | string    | Optional assignee         |
| `created_at`  | timestamp | Creation timestamp        |
| `updated_at`  | timestamp | Last update timestamp     |

Supported statuses:

```text
pending
in_progress
completed
```

### Validation

The service validates:

* Required task title
* Maximum title length
* Maximum description length
* Valid task status
* Maximum assignee length

---

## Requirements

For local development:

* Go 1.25+
* PostgreSQL 17+
* Docker
* Docker Compose
* `migrate` CLI for manually running migrations

Node.js/npm is only required if you want to validate the OpenAPI document using Swagger CLI.

---

## Configuration

Copy the example environment file:

```bash
cp .env.example .env
```

The application reads configuration from environment variables.

Example configuration:

```dotenv
APP_ENV=development
HTTP_PORT=8080
LOG_LEVEL=info

DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=task_manager
DATABASE_PASSWORD=task_manager
DATABASE_NAME=task_manager
DATABASE_SSLMODE=disable
DATABASE_MAX_CONNS=10
DATABASE_MIN_CONNS=2
DATABASE_MAX_CONN_LIFETIME_MINUTES=30
DATABASE_MAX_CONN_IDLE_TIME_MINUTES=5
```

For Docker Compose, PostgreSQL is available to the application as:

```dotenv
DATABASE_HOST=postgres
DATABASE_PORT=5432
```

---

## Running with Docker Compose

Start the complete environment:

```bash
docker compose up -d --build
```

The Compose environment contains:

```text
PostgreSQL
     │
     ▼
Migration Runner
     │
     ▼
Task Manager Application
```

PostgreSQL has a health check, and the application waits for PostgreSQL and successful migrations before starting.

Check the containers:

```bash
docker compose ps
```

Check the service:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "ok"
}
```

Stop the environment:

```bash
docker compose down
```

PostgreSQL data is stored in a Docker volume.

---

## Database Migrations

The Docker Compose setup runs migrations automatically before starting the application.

For manual migrations, configure `DATABASE_URL` and run:

```bash
make migration-up
```

Check the migration version:

```bash
make migration-version
```

Roll back the latest migration:

```bash
make migration-down
```

The test database can be migrated with:

```bash
make migration-test-up
```

---

## Running Locally

Start PostgreSQL and configure the database connection through environment variables.

Then run:

```bash
make migration-up
make run
```

The HTTP server listens on:

```text
http://localhost:8080
```

---

# API

Base API path:

```text
/api/v1
```

## Health Check

```http
GET /health
```

Example:

```bash
curl http://localhost:8080/health
```

Expected response:

```json
{
  "status": "ok"
}
```

---

## Create a Task

```http
POST /api/v1/tasks
```

Example:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Implement authentication",
    "description": "Add authentication to the API",
    "status": "pending",
    "assignee": "alice"
  }'
```

The `status` field is optional when creating a task.

When omitted, the task starts with:

```text
pending
```

---

## List Tasks

```http
GET /api/v1/tasks
```

Example:

```bash
curl http://localhost:8080/api/v1/tasks
```

### Pagination

```bash
curl 'http://localhost:8080/api/v1/tasks?page=1&page_size=20'
```

The default page is `1`.

The default page size is `20`.

The maximum page size is `100`.

### Filter by status

```bash
curl 'http://localhost:8080/api/v1/tasks?status=in_progress'
```

### Filter by assignee

```bash
curl 'http://localhost:8080/api/v1/tasks?assignee=alice'
```

### Combine filters and pagination

```bash
curl 'http://localhost:8080/api/v1/tasks?status=pending&assignee=alice&page=1&page_size=20'
```

---

## Get a Task

```http
GET /api/v1/tasks/{id}
```

Example:

```bash
curl http://localhost:8080/api/v1/tasks/00000000-0000-0000-0000-000000000000
```

Replace the UUID with an existing task ID.

---

## Update a Task

```http
PUT /api/v1/tasks/{id}
```

Example:

```bash
curl -X PUT http://localhost:8080/api/v1/tasks/00000000-0000-0000-0000-000000000000 \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "Implement authentication",
    "description": "Authentication implementation completed",
    "status": "completed",
    "assignee": "alice"
  }'
```

Replace the UUID with an existing task ID.

---

## Delete a Task

```http
DELETE /api/v1/tasks/{id}
```

Example:

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/00000000-0000-0000-0000-000000000000
```

Replace the UUID with an existing task ID.

---

# Observability

The service includes structured logging, request IDs, metrics, tracing, and runtime profiling.

## Request IDs

Every request receives an `X-Request-ID`.

If the client provides one, the service preserves it.

Otherwise, the service generates a UUID.

The request ID is:

* Returned in the response header
* Included in structured logs

Example:

```bash
curl -i http://localhost:8080/health
```

---

## Structured Logging

The service uses Go's structured `slog` logger.

HTTP request logs include information such as:

* Request ID
* HTTP method
* Request path
* Matched route
* Response status
* Request duration

---

## Prometheus Metrics

Prometheus metrics are exposed at:

```text
/metrics
```

Example:

```bash
curl http://localhost:8080/metrics
```

The service exposes:

* HTTP request counter
* HTTP request duration histogram
* Current task count gauge

The task count metric is:

```text
task_manager_tasks_count
```

HTTP metrics include:

```text
task_manager_http_requests_total
task_manager_http_request_duration_seconds
```

The task count is initialized from PostgreSQL and refreshed after task mutations.

---

# Load Testing and Benchmarking

The project includes Go benchmarks for measuring service-layer performance without requiring a running PostgreSQL instance.

Run the benchmarks with:

```bash
go test ./internal/service -bench=. -benchmem
```

The current benchmarks cover:

* `BenchmarkTaskService_Create` — measures task creation through the service layer using a mock repository.
* `BenchmarkTaskService_GetByID` — measures retrieving a task through the service layer using a mock repository.

## Actual benchmark result

The following result was measured during development:

```text
BenchmarkTaskService_Create-16       7003384   154.9 ns/op   16 B/op  1 allocs/op
BenchmarkTaskService_GetByID-16     39989529    29.96 ns/op   0 B/op  0 allocs/op
PASS
```

Interpretation:

* `BenchmarkTaskService_Create`: approximately `154.9 ns/op` in the measured environment.
* `BenchmarkTaskService_GetByID`: approximately `29.96 ns/op` in the measured environment.
* `B/op` represents allocated bytes per operation.
* `allocs/op` represents allocations per operation.

These measurements are **service-layer benchmarks using mocked persistence**. They are not representative of complete HTTP or PostgreSQL throughput.

Results are environment-dependent and should not be interpreted as production capacity.

## HTTP load testing

For HTTP-level load testing, the running service can be exercised with tools such as `hey`, `wrk`, or `k6`.

Example using `hey`:

```bash
docker compose up -d --build

hey -n 1000 -c 10 http://localhost:8080/health
```

Where:

* `-n 1000` sends 1,000 requests.
* `-c 10` uses 10 concurrent workers.

For API load testing, create test data first and then target the relevant task endpoints.

During a load test, `/metrics` and pprof can be monitored to observe:

* Request rate
* Request latency
* HTTP status distribution
* Current task count
* Memory behavior
* Goroutine behavior
* Runtime characteristics

HTTP load-test results should be recorded together with:

* Endpoint
* Request count
* Concurrency
* Payload
* Test tool and version
* Application version/commit
* Database configuration
* Docker/host environment

This avoids presenting environment-specific measurements as universal performance guarantees.

---

# pprof Runtime Profiling Report

The service exposes Go's built-in pprof runtime profiling endpoints under:

```text
/debug/pprof/
```

The pprof endpoints were verified against the running Docker Compose deployment.

## pprof index verification

The following endpoint was tested:

```bash
curl http://localhost:8080/debug/pprof/
```

The endpoint returned HTTP `200` and exposed the standard Go runtime profiles, including:

```text
allocs
block
goroutine
heap
mutex
threadcreate
```

The pprof index and profiles were successfully verified against the running Docker deployment.

## Real heap profile collection

A real heap profile was collected from the running service with:

```bash
curl http://localhost:8080/debug/pprof/heap -o /tmp/heap.pprof
```

The request completed successfully and produced:

```text
3256 bytes
```

for the collected heap profile.

This confirms that the application was able to serve an actual runtime heap profile from the running container.

## Inspecting the collected profile

The profile can be inspected using Go's pprof tooling:

```bash
go tool pprof /tmp/heap.pprof
```

For a top-level textual report:

```bash
go tool pprof -top /tmp/heap.pprof
```

For an interactive web visualization:

```bash
go tool pprof -http=:8081 /tmp/heap.pprof
```

Then open:

```text
http://localhost:8081
```

## Available pprof endpoints

The service exposes:

```text
/debug/pprof/
/debug/pprof/cmdline
/debug/pprof/profile
/debug/pprof/symbol
/debug/pprof/trace
/debug/pprof/goroutine
/debug/pprof/heap
/debug/pprof/threadcreate
/debug/pprof/block
/debug/pprof/mutex
/debug/pprof/allocs
```

## Profiling report interpretation

The collected heap profile is a runtime snapshot.

Therefore:

* Profile size depends on the runtime state.
* Memory allocations depend on the current workload.
* Goroutine counts depend on the current runtime state.
* Profiling results can change between executions.
* The collected profile should be treated as a diagnostic artifact rather than a fixed performance benchmark.

No specific CPU, memory, goroutine, or allocation conclusions are claimed here beyond the endpoints and heap-profile collection that were actually verified.

---

# Testing

Run the complete test suite:

```bash
go test ./...
```

Run tests with coverage:

```bash
make coverage
```

Run static analysis:

```bash
make vet
```

Format the project:

```bash
make fmt
```

Run repository integration tests:

```bash
make test-integration
```

The integration tests use:

```text
TEST_DATABASE_URL
```

If it is not provided, the Makefile supplies the project's local test database connection as the default.

## Coverage

The project was tested with:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

The measured overall statement coverage reached:

```text
70.6%
```

This satisfies the project's target of at least 70% statement coverage.

---

# Integration Testing

Repository integration tests use a real PostgreSQL database.

Set:

```bash
export TEST_DATABASE_URL="postgres://task_manager:task_manager@localhost:5432/task_manager_test?sslmode=disable"
```

Apply the migrations:

```bash
make migration-test-up
```

Run the integration tests:

```bash
make test-integration
```

The repository integration tests cover:

* Create and retrieve
* Not-found behavior
* Listing
* Filtering
* Pagination
* Update
* Delete
* Database behavior

Unit tests use mocked repository/database dependencies where appropriate.

---

# OpenAPI

The complete API specification is available in:

```text
openapi.yaml
```

It documents:

* Health endpoint
* Metrics endpoint
* Task creation
* Task listing
* Task retrieval
* Task updates
* Task deletion
* Request schemas
* Response schemas
* Error responses
* Pagination
* Filtering
* Task statuses

Validate the specification with:

```bash
npx @apidevtools/swagger-cli validate openapi.yaml
```

Expected result:

```text
openapi.yaml is valid
```

---

# Docker

The application uses a multi-stage Docker build.

The build process:

```text
Go source
    │
    ▼
Go builder image
    │
    │ compile
    ▼
Standalone binary
    │
    ▼
Minimal Alpine runtime image
```

The runtime container runs as a non-root user.

Docker Compose provides:

* PostgreSQL
* Migration runner
* Task Manager application
* PostgreSQL health checks
* Application health checks
* Persistent PostgreSQL storage

Build and start:

```bash
docker compose up -d --build
```

Stop:

```bash
docker compose down
```

---

# Project Structure

```text
task-manager/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── config/
│   │   ├── config.go
│   │   └── config_test.go
│   ├── database/
│   │   └── postgres.go
│   ├── domain/
│   │   ├── task.go
│   │   └── task_test.go
│   ├── handler/
│   │   ├── task_handler.go
│   │   ├── mock_task_service_test.go
│   │   └── task_handler_test.go
│   ├── metrics/
│   │   ├── metrics.go
│   │   └── metrics_test.go
│   ├── middleware/
│   │   ├── request_id.go
│   │   ├── request_logger.go
│   │   ├── request_logger_test.go
│   │   ├── metrics.go
│   │   ├── metrics_test.go
│   │   └── tracing.go
│   ├── repository/
│   │   ├── task_repository.go
│   │   ├── postgres_task_repository.go
│   │   ├── mock_task_repository.go
│   │   ├── postgres_task_repository_test.go
│   │   └── postgres_task_repository_integration_test.go
│   ├── server/
│   │   ├── server.go
│   │   └── http_integration_test.go
│   ├── service/
│   │   ├── errors.go
│   │   ├── task_service.go
│   │   ├── task_service_impl.go
│   │   ├── task_service_test.go
│   │   └── task_service_benchmark_test.go
│   └── tracing/
│       └── tracing.go
├── migrations/
│   ├── 000001_create_tasks.up.sql
│   └── 000001_create_tasks.down.sql
├── docs/
├── tests/
├── .env.example
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── openapi.yaml
└── README.md
```

---

# Useful Commands

Run the application:

```bash
make run
```

Run all tests:

```bash
make test
```

Run integration tests:

```bash
make test-integration
```

Generate coverage:

```bash
make coverage
```

Run static analysis:

```bash
make vet
```

Format the project:

```bash
make fmt
```

Update dependencies:

```bash
make tidy
```

Build the application:

```bash
make build
```

Start Docker environment:

```bash
make docker-up
```

Stop Docker environment:

```bash
make docker-down
```

Run service benchmarks:

```bash
go test ./internal/service -bench=. -benchmem
```

---

# Verification Checklist

Before considering the service ready, run:

```bash
gofmt -l .
go test ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
go vet ./...
npx @apidevtools/swagger-cli validate openapi.yaml
docker compose up -d --build
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl http://localhost:8080/debug/pprof/
```

The project should be verified through the complete development workflow:

```text
Inspect
  ↓
Design
  ↓
Implement
  ↓
Unit Test
  ↓
Integration Test
  ↓
Build
  ↓
Run
  ↓
Debug
  ↓
Verify
  ↓
Document
  ↓
Review
```