# Task API

REST API for task management built in Go, PostgreSQL and Docker.

Built as a learning project to practice Go backend fundamentals: clean package structure, interface-based dependency inversion, idiomatic error handling and containerized local development.

---

## Architecture

```
taskapi/
├── cmd/api/          # Entrypoint — wires dependencies and starts the server
├── internal/
│   ├── domain/       # Core entity (Task) + repository interface
│   ├── service/      # Business logic — depends only on the domain interface
│   ├── repository/   # PostgreSQL implementation of domain.TaskRepository
│   └── handler/      # HTTP layer — decodes requests, calls service, encodes responses
├── db/migrations/    # SQL migration files
├── Dockerfile
└── docker-compose.yml
```

### Key decisions

**Domain layer owns the interface, not the repository.**
`domain.TaskRepository` is defined alongside the `Task` entity. This means the business logic (`service`) depends on an abstraction, not on PostgreSQL. Swapping the storage backend requires no changes to the domain or service layers.

**No global state.**
Dependencies are injected through constructors (`NewTaskService`, `NewTaskHandler`). There are no `init()` functions or package-level variables holding connections.

**Errors are wrapped, not swallowed.**
Every function wraps errors with context using `fmt.Errorf("…: %w", err)` so the call stack is visible when something fails.

**HTTP layer is thin.**
Handlers decode input, delegate to the service, and encode output. No business logic lives in handlers.

---

## Running locally

**Requirements:** Docker and Docker Compose.

```bash
# Clone the repo
git clone https://github.com/pedrovsilva/taskapi
cd taskapi

# Start API + PostgreSQL
docker compose up --build
```

The API will be available at `http://localhost:8080`.

---

## API endpoints

| Method | Path              | Description        |
|--------|-------------------|--------------------|
| POST   | /api/v1/tasks     | Create a task      |
| GET    | /api/v1/tasks     | List all tasks     |
| GET    | /api/v1/tasks/:id | Get task by ID     |
| PATCH  | /api/v1/tasks/:id | Update a task      |
| DELETE | /api/v1/tasks/:id | Delete a task      |
| GET    | /health           | Health check       |

### Example requests

```bash
# Create
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Learn Go","description":"Study structs, interfaces and error handling"}'

# List
curl http://localhost:8080/api/v1/tasks

# Update status to done
curl -X PATCH http://localhost:8080/api/v1/tasks/<id> \
  -H "Content-Type: application/json" \
  -d '{"status":"done"}'

# Delete
curl -X DELETE http://localhost:8080/api/v1/tasks/<id>
```

### Task status flow

```
pending → in_progress → done
```

Once a task reaches `done`, it cannot be transitioned back. The `done_at` timestamp is set automatically.

---

## Running tests

Tests use an in-memory repository — no database required.

```bash
go test ./...
```

Expected output:
```
ok  github.com/pedrovsilva/taskapi/internal/domain
ok  github.com/pedrovsilva/taskapi/internal/service
ok  github.com/pedrovsilva/taskapi/internal/handler
```

### Test strategy

- **Unit tests** (`domain/`, `service/`): test business logic in isolation using an in-memory repository. No I/O, no database, fast.
- **Integration tests** (`handler/`): test the full HTTP layer using `net/http/httptest`. The router, handler, service and in-memory repository all run together — no mocking frameworks needed.
- **Not tested here**: the PostgreSQL repository. That would require a real database and is better covered by integration/e2e tests in CI.

---

## Swagger UI

After starting the API (`docker compose up`), open:

```
http://localhost:8080/swagger/index.html
```

The raw OpenAPI spec is available at:

```
http://localhost:8080/swagger/doc.json
```

---

## Running without Docker

```bash
# Copy and fill in environment variables
cp .env.example .env

# Run the migration against your local PostgreSQL
psql $DATABASE_URL -f db/migrations/001_create_tasks.sql

# Start the server
go run ./cmd/api
```
