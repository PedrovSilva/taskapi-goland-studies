# Task API

REST API for task management built in Go, PostgreSQL and Docker.

The project was built to practice backend engineering with Go, focusing on dependency inversion, explicit business rules, PostgreSQL persistence, HTTP testing, graceful shutdown and integration testing.

## Stack

* Go 1.25
* Chi
* PostgreSQL 16
* pgx
* Docker / Docker Compose
* Swagger / OpenAPI
* Testify
* GitHub Actions

## Architecture

```text
.
├── cmd/api/          # Application entrypoint and dependency wiring
├── internal/
│   ├── domain/       # Task entity, status rules and repository contract
│   ├── service/      # Business rules and use cases
│   ├── repository/   # PostgreSQL and in-memory implementations
│   └── handler/      # HTTP transport layer
├── db/migrations/    # PostgreSQL schema
├── docs/             # Generated Swagger documentation
├── Dockerfile
└── docker-compose.yml
```

The domain defines `TaskRepository`, so the service layer does not depend directly on PostgreSQL. Repository implementations can therefore be replaced without changing the business layer.

The HTTP handlers are intentionally thin: they decode requests, call the service and translate application errors into HTTP responses.

## Running locally

Requirements:

* Docker
* Docker Compose

Clone the repository:

```bash
git clone https://github.com/PedrovSilva/taskapi-goland-studies.git
cd taskapi-goland-studies
```

Start the API and PostgreSQL:

```bash
docker compose up --build
```

The API will be available at:

```text
http://localhost:8080
```

## API

| Method | Endpoint             | Description                     |
| ------ | -------------------- | ------------------------------- |
| POST   | `/api/v1/tasks`      | Create a task                   |
| GET    | `/api/v1/tasks`      | List tasks                      |
| GET    | `/api/v1/tasks/{id}` | Get a task                      |
| PATCH  | `/api/v1/tasks/{id}` | Update a task                   |
| DELETE | `/api/v1/tasks/{id}` | Delete a task                   |
| GET    | `/health`            | Liveness check                  |
| GET    | `/ready`             | Readiness check with PostgreSQL |

### Health checks

`/health` only verifies that the HTTP process is alive.

`/ready` verifies that the process is alive and that PostgreSQL is reachable.

When PostgreSQL is unavailable, `/ready` returns:

```text
503 Service Unavailable
```

This separates **liveness** from **readiness**, which is useful when the application is deployed under a container orchestrator.

### Example

Create a task:

```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Learn Go",
    "description": "Study interfaces and error handling"
  }'
```

List tasks:

```bash
curl http://localhost:8080/api/v1/tasks
```

Update a task:

```bash
curl -X PATCH http://localhost:8080/api/v1/tasks/<id> \
  -H "Content-Type: application/json" \
  -d '{
    "status": "done"
  }'
```

Delete a task:

```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/<id>
```

## Task status

Supported statuses:

* `pending`
* `in_progress`
* `done`

A task can move between `pending` and `in_progress`, and can be marked as `done`.

Once a task reaches `done`, the service prevents changing it to another status.

When a task is marked as done, the `done_at` timestamp is recorded automatically.

This is intentionally a small business rule rather than a full workflow/state-machine implementation.

## Context propagation

Request contexts are propagated from the HTTP layer through the service and repository layers to PostgreSQL.

This allows database operations to respect request cancellation and deadlines.

The repository does not create independent background contexts for database operations.

Example:

```go
func (r *PostgresTaskRepository) GetByID(
    ctx context.Context,
    id uuid.UUID,
) (*domain.Task, error) {
    row := r.pool.QueryRow(ctx, query, id)

    // ...
}
```

## Graceful shutdown

The application uses `http.Server` and handles:

* `SIGINT`
* `SIGTERM`

When a shutdown signal is received, the server allows in-flight requests to finish before terminating.

A 10-second timeout is used for graceful shutdown.

If graceful shutdown fails, the server falls back to forced closing.

This makes the application behave correctly when running inside containers or other managed environments.

## Testing

The project uses multiple levels of testing.

### Domain tests

Validate core domain behavior such as:

* required title;
* maximum title length;
* valid statuses;
* task completion;
* invalid state transitions.

### Service tests

Validate business rules independently from PostgreSQL using the in-memory repository.

### HTTP tests

Use `net/http/httptest` to exercise the HTTP layer and router without requiring a real database.

### PostgreSQL integration tests

The PostgreSQL repository is also tested against a real PostgreSQL instance.

The integration tests cover:

* Create;
* GetByID;
* Update;
* List;
* Delete;
* UUID persistence;
* PostgreSQL enum mapping;
* timestamps;
* `done_at`;
* not-found behavior.

Run the complete test suite with PostgreSQL:

```bash
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/taskapi?sslmode=disable"

psql "$TEST_DATABASE_URL" \
  -f db/migrations/001_create_tasks.sql

go test ./... -count=1
```

If `TEST_DATABASE_URL` is not configured, PostgreSQL integration tests are skipped.

## CI

GitHub Actions runs automatically on:

* pushes to `main`;
* pull requests.

The CI pipeline:

1. Starts PostgreSQL 16.
2. Applies the database migration.
3. Checks Go formatting.
4. Runs all tests.
5. Runs `go vet`.

```text
GitHub Actions
      │
      ├── PostgreSQL 16
      │
      ├── Database migration
      │
      ├── gofmt
      │
      ├── go test ./...
      │
      └── go vet ./...
```

The CI therefore exercises the real PostgreSQL repository rather than relying exclusively on an in-memory implementation.

## Swagger

With the application running, Swagger UI is available at:

```text
http://localhost:8080/swagger/index.html
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/doc.json
```

## Configuration

The application uses the following environment variables:

```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/taskapi?sslmode=disable
```

Docker Compose provides these variables automatically.

For local development without Docker, configure the variables manually or use `.env.example` as a reference.

## Project goals

This project focuses on demonstrating practical backend engineering with Go:

* idiomatic Go package organization;
* dependency inversion through interfaces;
* explicit domain rules;
* PostgreSQL persistence;
* context-aware database operations;
* HTTP testing;
* PostgreSQL integration testing;
* Dockerized development;
* liveness and readiness endpoints;
* graceful shutdown;
* automated CI.
